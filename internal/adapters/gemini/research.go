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
// structured ResearchReport. It uses a single GenerateContent call (no tool loop).
func (c *Client) ResearchSynthesize(
	ctx context.Context,
	topic string,
	hits []*domain.CachedAnalysis,
	projects []cncf.ProjectInfo,
	signals []*domain.Signal,
	talks []domain.VideoListing,
	lens *LensContext,
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

	prompt := buildResearchPrompt(topic, hits, projects, signals, talks)

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

	return toResearchReport(topic, r, talks), nil
}

func buildResearchPrompt(
	topic string,
	hits []*domain.CachedAnalysis,
	projects []cncf.ProjectInfo,
	signals []*domain.Signal,
	talks []domain.VideoListing,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TOPIC: %q\n\n", topic))

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

func toResearchReport(topic string, r researchReportJSON, talks []domain.VideoListing) *domain.ResearchReport {
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
