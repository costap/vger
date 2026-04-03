package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/domain"
)

const synthesisePrompt = `You are an expert in cloud-native technology, Kubernetes, and the CNCF ecosystem.

You have been given summaries and technology lists from multiple conference talks in a playlist.
Synthesise this information and return a JSON object with exactly this schema:

{
  "overarching_theme": "<2-3 sentence description of the major theme and direction of this playlist>",
  "learning_path": ["<tech name>", "<tech name>", ...],
  "priority_talks": [
    {
      "title": "<talk title>",
      "url": "<youtube url>",
      "reason": "<why this talk should be watched first — 1-2 sentences>"
    }
  ],
  "key_insights": "<freeform paragraph: the 3-5 most important insights an engineer should take away from this playlist>"
}

Rules:
- learning_path: list 5-10 technology names in the order an engineer should learn them, based on their prevalence and importance across the talks
- priority_talks: pick 3-5 talks that provide the most value; prioritise foundational or high-impact topics
- Return only the JSON object. Do not wrap it in markdown code fences.`

type synthesiseResponse struct {
	OverarchingTheme string              `json:"overarching_theme"`
	LearningPath     []string            `json:"learning_path"`
	PriorityTalks    []priorityTalkResp  `json:"priority_talks"`
	KeyInsights      string              `json:"key_insights"`
}

type priorityTalkResp struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Reason string `json:"reason"`
}

// Synthesise takes a slice of cached analyses and asks Gemini to produce a
// cross-playlist digest: overarching theme, recommended learning path, and
// priority talks to watch first.
func (c *Client) Synthesise(ctx context.Context, entries []*domain.CachedAnalysis) (*domain.DigestReport, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("PLAYLIST TALKS:\n\n")
	for i, e := range entries {
		sb.WriteString(fmt.Sprintf("--- Talk %d ---\n", i+1))
		sb.WriteString(fmt.Sprintf("Title: %s\n", e.Report.VideoTitle))
		sb.WriteString(fmt.Sprintf("URL: %s\n", e.Report.VideoURL))
		sb.WriteString(fmt.Sprintf("Summary: %s\n", e.Report.Summary))
		sb.WriteString("Technologies: ")
		names := make([]string, len(e.Report.Technologies))
		for j, t := range e.Report.Technologies {
			names[j] = t.Name
		}
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n\n")
	}

	resp, err := gc.Models.GenerateContent(ctx, c.Model,
		genai.Text(sb.String()),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(synthesisePrompt, genai.RoleUser),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini synthesise: %w", err)
	}

	raw := strings.TrimSpace(resp.Text())
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var sr synthesiseResponse
	if err := json.Unmarshal([]byte(raw), &sr); err != nil {
		return nil, fmt.Errorf("parse synthesise response: %w\nraw: %s", err, raw)
	}

	pt := make([]domain.PriorityTalk, len(sr.PriorityTalks))
	for i, p := range sr.PriorityTalks {
		pt[i] = domain.PriorityTalk{Title: p.Title, URL: p.URL, Reason: p.Reason}
	}

	return &domain.DigestReport{
		OverarchingTheme: sr.OverarchingTheme,
		LearningPath:     sr.LearningPath,
		PriorityTalks:    pt,
		KeyInsights:      sr.KeyInsights,
	}, nil
}
