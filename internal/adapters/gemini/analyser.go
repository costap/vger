package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/domain"
)

const defaultModel = "gemini-2.5-flash"

const systemPrompt = `You are an expert in cloud-native technology, Kubernetes, and the CNCF ecosystem.
You will be given a conference talk video. Analyse it and return a JSON object with exactly this schema:

{
  "summary": "<concise technical summary of the talk, 3-5 sentences>",
  "notes": "<detailed freeform narrative covering EVERYTHING mentioned in the video: all technologies and projects (even brief mentions), speaker names and affiliations, demo highlights, code or architecture details shown, audience Q&A moments, and any quotes worth preserving. Write this as a thorough paragraph or set of paragraphs — this will be used to answer follow-up questions so err on the side of completeness>",
  "technologies": [
    {
      "name": "<technology or project name>",
      "description": "<one sentence describing what it is>",
      "why_relevant": "<why an engineer should pay attention to this>",
      "learn_more": "<URL to official docs or project site>",
      "cncf_stage": "<one of: graduated, incubating, sandbox, or empty string if not a CNCF project>"
    }
  ]
}

Return only the JSON object. Do not wrap it in markdown code fences. Do not add commentary outside the JSON.
Focus the technologies list on projects that are novel or worth learning more about. The notes field should be exhaustive.`

// analysisResponse mirrors the JSON schema requested from the model.
type analysisResponse struct {
	Summary      string               `json:"summary"`
	Notes        string               `json:"notes"`
	Technologies []technologyResponse `json:"technologies"`
}

type technologyResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WhyRelevant string `json:"why_relevant"`
	LearnMore   string `json:"learn_more"`
	CNCFStage   string `json:"cncf_stage"`
}

// Client implements domain.VideoAnalyser using the Gemini API.
type Client struct {
	APIKey string
	Model  string
}

func New(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{APIKey: apiKey, Model: model}
}

// AnalyseVideo passes the video URL directly to the Gemini multimodal API
// and returns a structured report. No video download is performed.
func (c *Client) AnalyseVideo(ctx context.Context, url string, meta *domain.VideoMetadata) (*domain.Report, error) {
	if url == "" {
		return nil, fmt.Errorf("url must not be empty")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: systemPrompt},
				{
					FileData: &genai.FileData{
						FileURI:  url,
						MIMEType: "video/mp4",
					},
				},
				{
					Text: fmt.Sprintf("Video title: %s\nChannel: %s\nPublished: %s",
						meta.Title, meta.ChannelName, meta.PublishedAt),
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := client.Models.GenerateContent(ctx, c.Model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini generate content: %w", err)
	}

	raw := resp.Text()
	if raw == "" {
		return nil, fmt.Errorf("empty response from gemini")
	}

	raw = stripCodeFences(raw)

	var ar analysisResponse
	if err := json.Unmarshal([]byte(raw), &ar); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w (raw: %.200s)", err, raw)
	}

	report := &domain.Report{
		VideoTitle:       meta.Title,
		VideoURL:         url,
		VideoDurationSec: meta.DurationSec,
		Summary:          ar.Summary,
		Notes:            ar.Notes,
	}
	for _, t := range ar.Technologies {
		report.Technologies = append(report.Technologies, domain.Technology{
			Name:        t.Name,
			Description: t.Description,
			WhyRelevant: t.WhyRelevant,
			LearnMore:   t.LearnMore,
			CNCFStage:   t.CNCFStage,
		})
	}

	return report, nil
}

// stripCodeFences removes markdown ```json ... ``` or ``` ... ``` wrappers
// that the model may produce despite being instructed not to.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	for _, fence := range []string{"```json", "```"} {
		if strings.HasPrefix(s, fence) {
			s = strings.TrimPrefix(s, fence)
			s = strings.TrimSuffix(s, "```")
			s = strings.TrimSpace(s)
			break
		}
	}
	return s
}
