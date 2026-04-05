package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/domain"
)

const enrichSignalPrompt = `You are an expert solutions architect and technology evaluator.

You have been given a technology signal — something spotted as potentially interesting or relevant.
Analyse the signal and return a JSON object with exactly this schema:

{
  "what_it_is": "<2-3 sentence explanation: what this technology is, what problem it solves, and how it fits in the ecosystem>",
  "maturity": "<assessment of production-readiness: community size, stability, adoption, risks — 2-3 sentences>",
  "alternatives": ["<alternative 1>", "<alternative 2>", "<alternative 3>"],
  "stack_fit": "<how this technology fits with a modern cloud-native stack (Kubernetes, Go, GitOps, observability) — 2-3 sentences>",
  "next_steps": ["<concrete action 1>", "<concrete action 2>", "<concrete action 3>"]
}

Rules:
- alternatives: list 2-5 realistic alternatives or related technologies
- next_steps: 2-5 concrete, actionable steps to evaluate this technology (e.g. "Read the docs at...", "Try the quickstart", "Compare benchmark with X")
- Return only the JSON object. Do not wrap it in markdown code fences.`

type enrichSignalResponse struct {
	WhatItIs     string   `json:"what_it_is"`
	Maturity     string   `json:"maturity"`
	Alternatives []string `json:"alternatives"`
	StackFit     string   `json:"stack_fit"`
	NextSteps    []string `json:"next_steps"`
}

// EnrichSignal calls Gemini to generate AI context for an existing signal.
// It fills WhatItIs, Maturity, Alternatives, StackFit, and NextSteps.
// userContext is injected into the prompt when non-empty so that StackFit
// and NextSteps are tailored to the user's actual environment.
func (c *Client) EnrichSignal(ctx context.Context, sig *domain.Signal, userContext string) (*domain.SignalEnrichment, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("SIGNAL TO ENRICH:\n\n")
	sb.WriteString(fmt.Sprintf("Title: %s\n", sig.Title))
	if sig.URL != "" {
		sb.WriteString(fmt.Sprintf("URL: %s\n", sig.URL))
	}
	if sig.Category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s\n", sig.Category))
	}
	if len(sig.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(sig.Tags, ", ")))
	}
	if sig.Note != "" {
		sb.WriteString(fmt.Sprintf("Why captured: %s\n", sig.Note))
	}
	if strings.TrimSpace(userContext) != "" {
		sb.WriteString(fmt.Sprintf("\nUSER CONTEXT (tailor stack_fit and next_steps to this environment):\n%s\n", strings.TrimSpace(userContext)))
	}

	resp, err := gc.Models.GenerateContent(ctx, c.Model,
		genai.Text(sb.String()),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(enrichSignalPrompt, genai.RoleUser),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini enrich signal: %w", err)
	}

	raw := strings.TrimSpace(resp.Text())
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var er enrichSignalResponse
	if err := json.Unmarshal([]byte(raw), &er); err != nil {
		return nil, fmt.Errorf("parse enrich response: %w\nraw: %s", err, raw)
	}

	return &domain.SignalEnrichment{
		EnrichedAt:   time.Now().UTC(),
		WhatItIs:     er.WhatItIs,
		Maturity:     er.Maturity,
		Alternatives: er.Alternatives,
		StackFit:     er.StackFit,
		NextSteps:    er.NextSteps,
	}, nil
}
