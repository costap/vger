package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/domain"
)

// Ask answers a follow-up question about a previously analysed video.
// It uses the cached report as text context — no video re-upload is performed.
func (c *Client) Ask(ctx context.Context, question string, cached *domain.CachedAnalysis) (string, error) {
	if question == "" {
		return "", fmt.Errorf("question must not be empty")
	}
	if cached == nil {
		return "", fmt.Errorf("no cached analysis provided")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	prompt := buildQAPrompt(question, cached)

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: prompt}},
		},
	}

	resp, err := client.Models.GenerateContent(ctx, c.Model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("gemini generate content: %w", err)
	}

	answer := strings.TrimSpace(resp.Text())
	if answer == "" {
		return "", fmt.Errorf("empty response from gemini")
	}
	return answer, nil
}

// buildQAPrompt constructs the text-only context prompt for a follow-up question.
func buildQAPrompt(question string, cached *domain.CachedAnalysis) string {
	r := &cached.Report
	m := &cached.Metadata

	var techLines strings.Builder
	for i, t := range r.Technologies {
		techLines.WriteString(fmt.Sprintf("  %d. %s — %s", i+1, t.Name, t.Description))
		if t.CNCFStage != "" {
			techLines.WriteString(fmt.Sprintf(" (CNCF %s)", t.CNCFStage))
		}
		techLines.WriteString("\n")
	}

	notesSection := ""
	if r.Notes != "" {
		notesSection = fmt.Sprintf("\nDetailed notes (everything mentioned in the video):\n%s\n", r.Notes)
	}

	return fmt.Sprintf(`You are an expert in cloud-native technology. You previously analysed a conference talk video. Use only the context below to answer the user's question concisely and accurately.

--- VIDEO CONTEXT ---
Title:    %s
Channel:  %s
Duration: %s
Published: %s

Summary:
%s
%s
Technologies identified:
%s
--- END CONTEXT ---

Question: %s`,
		m.Title,
		m.ChannelName,
		formatDurationQA(m.DurationSec),
		m.PublishedAt,
		r.Summary,
		notesSection,
		techLines.String(),
		question,
	)
}

// formatDurationQA formats seconds as a human-readable duration string.
func formatDurationQA(secs int) string {
	if secs <= 0 {
		return "unknown"
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}

// AskDeep answers a question about a video by re-submitting the YouTube URL
// to Gemini as a FileData part, giving the model direct access to the full
// video content. This is more expensive than Ask but can answer anything.
func (c *Client) AskDeep(ctx context.Context, question string, videoURL string, cached *domain.CachedAnalysis) (string, error) {
	if question == "" {
		return "", fmt.Errorf("question must not be empty")
	}
	if videoURL == "" {
		return "", fmt.Errorf("video url must not be empty")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	titleHint := ""
	if cached != nil {
		titleHint = fmt.Sprintf("Video title: %s\n", cached.Report.VideoTitle)
	}

	prompt := fmt.Sprintf(`You are an expert in cloud-native technology. Watch the video and answer the following question directly and concisely. Base your answer on what is actually said or shown in the video.

%sQuestion: %s`, titleHint, question)

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
				{
					FileData: &genai.FileData{
						FileURI:  videoURL,
						MIMEType: "video/mp4",
					},
				},
			},
		},
	}

	resp, err := client.Models.GenerateContent(ctx, c.Model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("gemini generate content: %w", err)
	}

	answer := strings.TrimSpace(resp.Text())
	if answer == "" {
		return "", fmt.Errorf("empty response from gemini")
	}
	return answer, nil
}
