package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/adapters/cncf"
	"github.com/costap/vger/internal/domain"
)

const researchSystemPrompt = `You are V'Ger, a cloud-native technology intelligence system.

You have been given evidence gathered from multiple sources about a specific technology topic.
Your task is to synthesise this evidence into a structured research brief.

Return a JSON object with EXACTLY this schema:

{
  "brief": "<2-3 sentence what-and-why: what the topic is, why it matters now>",
  "landscape_map": [
    {
      "name": "<project name>",
      "cncf_stage": "<graduated|incubating|sandbox|>",
      "category": "<category>",
      "homepage": "<url>",
      "relevance": "<1 sentence: why relevant to this topic>"
    }
  ],
  "evidence_videos": [
    {
      "video_title": "<title>",
      "video_url": "<url>",
      "relevance": "<1-2 sentences: what this talk contributes to understanding the topic>"
    }
  ],
  "investigation_paths": [
    {
      "title": "<path title>",
      "description": "<1-2 sentence description>",
      "actions": ["<concrete step>", ...]
    }
  ],
  "competing_approaches": ["<approach or technology>", ...],
  "verdict": "<2-3 sentence bottom-line recommendation: should the reader invest time in this topic? what should they do first?>"
}

Rules:
- landscape_map: include only projects genuinely relevant to the topic; add projects from the evidence even if not in CNCF landscape data
- evidence_videos: only include videos from the EVIDENCE section; do not invent titles or URLs
- investigation_paths: provide 2-4 distinct paths (e.g. "Start with fundamentals", "Production evaluation", "Security angle")
- competing_approaches: list 2-5 alternatives or complementary technologies
- Return only the JSON object. Do not wrap it in markdown code fences.`

// researchReportJSON is the JSON shape returned by Gemini for research synthesis.
type researchReportJSON struct {
	Brief               string               `json:"brief"`
	LandscapeMap        []relatedProjectJSON `json:"landscape_map"`
	EvidenceVideos      []evidenceEntryJSON  `json:"evidence_videos"`
	InvestigationPaths  []investPathJSON     `json:"investigation_paths"`
	CompetingApproaches []string             `json:"competing_approaches"`
	Verdict             string               `json:"verdict"`
}

type relatedProjectJSON struct {
	Name      string `json:"name"`
	CNCFStage string `json:"cncf_stage"`
	Category  string `json:"category"`
	Homepage  string `json:"homepage"`
	Relevance string `json:"relevance"`
}

type evidenceEntryJSON struct {
	VideoTitle string `json:"video_title"`
	VideoURL   string `json:"video_url"`
	Relevance  string `json:"relevance"`
}

type investPathJSON struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}

// ResearchSynthesize sends all gathered research context to Gemini and returns a
// structured ResearchReport.
//
// When maxDepth is 0, a single GenerateContent call is made (Phase 1 behaviour).
// When maxDepth > 0, an investigation phase runs first: Gemini uses ask_video,
// search_cache, and lookup_cncf_project tools to deepen its understanding, producing
// a text transcript that is appended to the synthesis prompt (Phase 2 / Option B).
//
// cacheSearcher is required when maxDepth > 0 and may be nil otherwise.
func (c *Client) ResearchSynthesize(
	ctx context.Context,
	topic string,
	hits []*domain.CachedAnalysis,
	projects []cncf.ProjectInfo,
	signals []*domain.Signal,
	talks []domain.VideoListing,
	lens *LensContext,
	maxDepth int,
	cacheSearcher domain.CacheSearcher,
	userContext string,
) (*domain.ResearchReport, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	systemPrompt := researchSystemPrompt
	if lens != nil {
		systemPrompt = lens.RoleContext + "\n\n" + researchSystemPrompt
	}

	prompt := buildResearchPrompt(topic, hits, projects, signals, talks, userContext)

	// Phase 2: run investigation loop before synthesis.
	if maxDepth > 0 {
		transcript, err := c.runInvestigation(ctx, gc, topic, hits, cacheSearcher, maxDepth)
		if err != nil {
			// Investigation failure is non-fatal; fall back to Phase 1 synthesis.
			transcript = fmt.Sprintf("(investigation failed: %s)", err.Error())
		}
		if transcript != "" {
			prompt += "\n\nINVESTIGATION TRANSCRIPT (Gemini deep-dive findings):\n" + transcript
		}
	}

	resp, err := gc.Models.GenerateContent(ctx, c.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
			ResponseMIMEType:  "application/json",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini research synthesize: %w", err)
	}

	raw := stripCodeFences(strings.TrimSpace(resp.Text()))
	raw = extractJSONObject(raw)

	var r researchReportJSON
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("unmarshal research response: %w (raw: %.200s)", err, raw)
	}

	return toResearchReport(topic, r, hits, talks), nil
}

// investigationSystemPrompt instructs Gemini on how to use research tools.
const investigationSystemPrompt = `You are V'Ger, a cloud-native technology intelligence system in investigation mode.
Your goal is to build a comprehensive understanding of a technology topic using the tools available to you.

You have access to:
- ask_video(video_id, question): ask a specific question about a cached conference talk
- search_cache(query): discover additional cached talks related to a search term
- lookup_cncf_project(name): get the current CNCF graduation stage of a project

Strategy:
1. Use search_cache to find talks you may not have been given
2. Use ask_video to get deeper detail from the most relevant talks
3. Use lookup_cncf_project to verify project maturity claims

When you have gathered sufficient evidence, write a comprehensive investigation summary as plain text (not JSON). Cover:
- What the technology is, why it matters now, and key adoption signals from the talks
- Real-world patterns, trade-offs, and implementation challenges mentioned
- Related CNCF projects and their current maturity
- Open questions or gaps that the synthesis should address`

// runInvestigation runs a tool-enabled multi-turn investigation loop and returns
// a plain-text transcript of Gemini's findings. The transcript is intended to be
// appended to the synthesis prompt to enrich the final JSON output.
func (c *Client) runInvestigation(
	ctx context.Context,
	gc *genai.Client,
	topic string,
	hits []*domain.CachedAnalysis,
	cacheSearcher domain.CacheSearcher,
	maxDepth int,
) (string, error) {
	tools := newResearchToolSet(c.cncfClient, hits, cacheSearcher, c)

	// Build the investigation prompt: topic + inventory of queryable videos.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TOPIC: %q\n\n", topic))
	if len(hits) > 0 {
		sb.WriteString(fmt.Sprintf("AVAILABLE CACHED TALKS (%d) — use ask_video with these IDs:\n", len(hits)))
		for _, h := range hits {
			sb.WriteString(fmt.Sprintf("  video_id=%q  title=%q\n", h.VideoID, h.Report.VideoTitle))
		}
	} else {
		sb.WriteString("No cached talks available yet — use search_cache to discover relevant videos.\n")
	}
	sb.WriteString("\nInvestigate the topic using the tools above, then write your investigation summary.")

	contents := []*genai.Content{
		{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{{Text: sb.String()}},
		},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(investigationSystemPrompt, genai.RoleUser),
		Tools: []*genai.Tool{
			{FunctionDeclarations: tools.declarations},
		},
	}

	for round := 0; round < maxDepth; round++ {
		resp, err := gc.Models.GenerateContent(ctx, c.Model, contents, config)
		if err != nil {
			return "", fmt.Errorf("investigation round %d: %w", round+1, err)
		}

		fcs := resp.FunctionCalls()
		if len(fcs) == 0 {
			// No more tool calls — model has written its investigation summary.
			return strings.TrimSpace(resp.Text()), nil
		}

		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			contents = append(contents, resp.Candidates[0].Content)
		}

		toolContent := &genai.Content{Role: genai.RoleUser}
		for _, fc := range fcs {
			result := tools.execute(ctx, fc)
			toolContent.Parts = append(toolContent.Parts,
				genai.NewPartFromFunctionResponse(fc.Name, result))
		}
		contents = append(contents, toolContent)
	}

	// Rounds exhausted — prompt for a final summary with remaining context.
	contents = append(contents, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: "You have reached the maximum investigation depth. Please write your investigation summary now based on what you have gathered so far."}},
	})
	resp, err := gc.Models.GenerateContent(ctx, c.Model, contents, &genai.GenerateContentConfig{
		SystemInstruction: config.SystemInstruction,
	})
	if err != nil {
		return "", fmt.Errorf("investigation final summary: %w", err)
	}
	return strings.TrimSpace(resp.Text()), nil
}

func buildResearchPrompt(
	topic string,
	hits []*domain.CachedAnalysis,
	projects []cncf.ProjectInfo,
	signals []*domain.Signal,
	talks []domain.VideoListing,
	userContext string,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TOPIC: %q\n\n", topic))

	if strings.TrimSpace(userContext) != "" {
		sb.WriteString(fmt.Sprintf("USER CONTEXT (tailor your answer to this environment):\n%s\n\n", strings.TrimSpace(userContext)))
	}

	// Evidence from scanned videos.
	if len(hits) > 0 {
		sb.WriteString(fmt.Sprintf("EVIDENCE FROM SCANNED VIDEOS (%d talks):\n", len(hits)))
		for i, h := range hits {
			sb.WriteString(fmt.Sprintf("[%d] Title: %s\n", i+1, h.Report.VideoTitle))
			sb.WriteString(fmt.Sprintf("    URL: %s\n", h.Report.VideoURL))
			summary := h.Report.Summary
			if len(summary) > 400 {
				summary = summary[:400] + "…"
			}
			sb.WriteString(fmt.Sprintf("    Summary: %s\n", summary))
			names := make([]string, len(h.Report.Technologies))
			for j, t := range h.Report.Technologies {
				names[j] = t.Name
			}
			sb.WriteString(fmt.Sprintf("    Technologies: %s\n\n", strings.Join(names, ", ")))
		}
	} else {
		sb.WriteString("EVIDENCE FROM SCANNED VIDEOS: none found in local cache.\n\n")
	}

	// CNCF landscape projects.
	if len(projects) > 0 {
		sb.WriteString(fmt.Sprintf("CNCF LANDSCAPE (%d related projects):\n", len(projects)))
		for _, p := range projects {
			sb.WriteString(fmt.Sprintf("- %s [%s, %s] — %s\n", p.Name, p.Stage, p.Category, p.Homepage))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("CNCF LANDSCAPE: no matching projects found.\n\n")
	}

	// Tracked signals.
	if len(signals) > 0 {
		sb.WriteString(fmt.Sprintf("TRACKED SIGNALS (%d):\n", len(signals)))
		for _, s := range signals {
			sb.WriteString(fmt.Sprintf("- %s: %q [%s, %s]\n", s.ID, s.Title, s.Status, s.Category))
			if s.Note != "" {
				note := s.Note
				if len(note) > 200 {
					note = note[:200] + "…"
				}
				sb.WriteString(fmt.Sprintf("  Note: %s\n", note))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("TRACKED SIGNALS: none.\n\n")
	}

	// Discovered unscanned talks.
	if len(talks) > 0 {
		sb.WriteString(fmt.Sprintf("UNDISCOVERED TALKS (%d found via YouTube):\n", len(talks)))
		for _, t := range talks {
			sb.WriteString(fmt.Sprintf("- %q — %s\n", t.Title, t.URL))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func toResearchReport(topic string, r researchReportJSON, hits []*domain.CachedAnalysis, talks []domain.VideoListing) *domain.ResearchReport {
	// Build a URL→speakers lookup from cached hits for evidence enrichment.
	speakersByURL := make(map[string][]string, len(hits))
	for _, h := range hits {
		if len(h.Report.Speakers) > 0 {
			speakersByURL[h.Metadata.URL] = h.Report.Speakers
		}
	}

	projects := make([]domain.RelatedProject, len(r.LandscapeMap))
	for i, p := range r.LandscapeMap {
		projects[i] = domain.RelatedProject{
			Name:      p.Name,
			CNCFStage: p.CNCFStage,
			Category:  p.Category,
			Homepage:  p.Homepage,
			Relevance: p.Relevance,
		}
	}

	evidence := make([]domain.EvidenceEntry, len(r.EvidenceVideos))
	for i, e := range r.EvidenceVideos {
		evidence[i] = domain.EvidenceEntry{
			VideoTitle: e.VideoTitle,
			VideoURL:   e.VideoURL,
			Speakers:   speakersByURL[e.VideoURL],
			Relevance:  e.Relevance,
		}
	}

	paths := make([]domain.InvestPath, len(r.InvestigationPaths))
	for i, p := range r.InvestigationPaths {
		paths[i] = domain.InvestPath{
			Title:       p.Title,
			Description: p.Description,
			Actions:     p.Actions,
		}
	}

	return &domain.ResearchReport{
		Topic:               topic,
		Brief:               r.Brief,
		LandscapeMap:        projects,
		EvidenceVideos:      evidence,
		DiscoveredTalks:     talks,
		InvestigationPaths:  paths,
		CompetingApproaches: r.CompetingApproaches,
		Verdict:             r.Verdict,
	}
}

// lensContext carries the role context from an analytical lens for research synthesis.
// Exported so the CLI can pass it in without importing lenses.go internals.
type LensContext struct {
	RoleContext string
}
