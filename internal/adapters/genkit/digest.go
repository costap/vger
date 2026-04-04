package genkitadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/costap/vger/internal/domain"
)

const digestSystemPrompt = `You are an experienced solutions architect reviewing your personal technology backlog.

You will be given a list of technology signals — things you have flagged as worth investigating.
Your job is to help the architect get the most value from their limited investigation time.

Return a JSON object with exactly this schema:
{
  "weekly_focus": [
    {
      "signal_id": "<id>",
      "title": "<signal title>",
      "url": "<signal url or empty string>",
      "reason": "<why investigate this one now — 1-2 sentences>"
    }
  ],
  "clusters": [
    {
      "theme": "<technology theme name>",
      "signal_ids": ["<id1>", "<id2>"],
      "summary": "<1-2 sentences on the common thread>"
    }
  ],
  "learning_path": ["<tech name or title 1>", "<tech name or title 2>"],
  "key_insights": "<narrative paragraph: patterns, trends, and connections across the backlog>"
}

Rules:
- weekly_focus: pick exactly 3 signals (or fewer if fewer than 3 exist) to investigate this week.
  Prioritise based on: strategic fit for a cloud-native platform team, maturity, and age (older spotted signals first).
- clusters: group related signals by technology theme (e.g. "eBPF-based networking", "AI-native CI/CD").
  A signal can appear in at most one cluster. Omit clusters if fewer than 2 signals share a theme.
- learning_path: 3-8 technology names in the order an engineer should investigate them.
- key_insights: 2-4 sentence narrative highlighting patterns and trends.
- Return only the JSON object. Do not wrap it in markdown code fences.`

// DigestInput is the input to the digest flow.
type DigestInput struct {
	Signals []domain.Signal   `json:"signals"`
	Pulse   domain.SignalPulse `json:"pulse"`
}

// DigestSignals runs the Genkit digest flow: takes a list of signals and returns
// a structured SignalDigestReport with weekly focus, clusters, and key insights.
func DigestSignals(ctx context.Context, apiKey, model string, input DigestInput) (*domain.SignalDigestReport, error) {
	g := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: apiKey}))

	digestFlow := genkit.DefineFlow(g, "digest-signals",
		func(ctx context.Context, in DigestInput) (*domain.SignalDigestReport, error) {
			prompt := buildDigestPrompt(in)

			result, _, err := genkit.GenerateData[domain.SignalDigestReport](ctx, g,
				ai.WithModelName("googleai/"+model),
				ai.WithSystem(digestSystemPrompt),
				ai.WithPrompt(prompt),
			)
			if err != nil {
				return nil, fmt.Errorf("gemini digest: %w", err)
			}
			return result, nil
		},
	)

	return digestFlow.Run(ctx, input)
}

func buildDigestPrompt(input DigestInput) string {
	var sb strings.Builder

	sb.WriteString("SIGNAL BACKLOG OVERVIEW:\n")
	sb.WriteString(fmt.Sprintf("Total signals: %d\n", len(input.Signals)))
	sb.WriteString("By status:")
	for status, count := range input.Pulse.ByStatus {
		sb.WriteString(fmt.Sprintf(" %s=%d", status, count))
	}
	sb.WriteString("\nBy category:")
	for cat, count := range input.Pulse.ByCategory {
		sb.WriteString(fmt.Sprintf(" %s=%d", cat, count))
	}
	sb.WriteString("\n\n")

	sb.WriteString("SIGNALS:\n\n")
	for _, sig := range input.Signals {
		sb.WriteString(fmt.Sprintf("--- Signal %s ---\n", sig.ID))
		sb.WriteString(fmt.Sprintf("Title: %s\n", sig.Title))
		sb.WriteString(fmt.Sprintf("Date captured: %s\n", sig.Date))
		sb.WriteString(fmt.Sprintf("Status: %s\n", sig.Status))
		sb.WriteString(fmt.Sprintf("Category: %s\n", sig.Category))
		if sig.Source != "" {
			sb.WriteString(fmt.Sprintf("Source: %s\n", sig.Source))
		}
		if sig.URL != "" {
			sb.WriteString(fmt.Sprintf("URL: %s\n", sig.URL))
		}
		if len(sig.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(sig.Tags, ", ")))
		}
		if sig.Note != "" {
			sb.WriteString(fmt.Sprintf("Why captured: %s\n", sig.Note))
		}
		if sig.Enrichment != nil {
			e := sig.Enrichment
			if e.WhatItIs != "" {
				sb.WriteString(fmt.Sprintf("What it is: %s\n", e.WhatItIs))
			}
			if e.Maturity != "" {
				sb.WriteString(fmt.Sprintf("Maturity: %s\n", e.Maturity))
			}
			if e.StackFit != "" {
				sb.WriteString(fmt.Sprintf("Stack fit: %s\n", e.StackFit))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
