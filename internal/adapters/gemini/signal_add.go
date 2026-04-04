package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/domain"
)

const parseSignalPrompt = `You are a technology intelligence assistant.

You will be given a free-text description of a technology or idea that someone found interesting.
Extract the key fields and return a JSON object with exactly this schema:

{
  "title": "<short, clear name for this technology or idea — 3-8 words>",
  "url": "<primary URL if mentioned, otherwise empty string>",
  "source": "<where it was discovered — infer from context: Blog post, Twitter/X, Conference talk, Colleague, Podcast, GitHub, Reddit, Newsletter, YouTube, or Other>",
  "category": "<one of: networking, security, platform, data, ai, observability, developer-experience, process, other>",
  "tags": ["<tag1>", "<tag2>"],
  "note": "<why this is interesting — 1-2 sentences capturing the original intent>"
}

Rules:
- title: concise and specific (e.g. "eBPF-based sidecarless service mesh" not "new thing")
- source: infer from context clues (e.g. "tweet" → Twitter/X, "video" → YouTube, "article" → Blog post)
- category: pick the single best fit from the allowed list
- tags: 2-5 lowercase tags describing key concepts (e.g. ["ebpf", "service-mesh", "kubernetes"])
- note: capture the "why" from the original description; don't repeat the title
- Return only the JSON object. Do not wrap it in markdown code fences.`

type parseSignalResponse struct {
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Source   string   `json:"source"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Note     string   `json:"note"`
}

// ParseSignalFromPrompt extracts a Signal from a free-text description using Gemini.
// The caller is responsible for assigning ID, Status, Date, CreatedAt, and UpdatedAt.
func (c *Client) ParseSignalFromPrompt(ctx context.Context, prompt string) (*domain.Signal, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	resp, err := gc.Models.GenerateContent(ctx, c.Model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(parseSignalPrompt, genai.RoleUser),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini parse signal: %w", err)
	}

	raw := strings.TrimSpace(resp.Text())
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var pr parseSignalResponse
	if err := json.Unmarshal([]byte(raw), &pr); err != nil {
		return nil, fmt.Errorf("parse signal response: %w\nraw: %s", err, raw)
	}

	return &domain.Signal{
		Title:    pr.Title,
		URL:      pr.URL,
		Source:   pr.Source,
		Category: pr.Category,
		Tags:     pr.Tags,
		Note:     pr.Note,
	}, nil
}
