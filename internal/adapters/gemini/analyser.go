package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/costap/vger/internal/adapters/cncf"
	"github.com/costap/vger/internal/domain"
)

const defaultModel = "gemini-2.5-flash"

// fallbackModel is used when a connection reset occurs on the primary model.
// Both 2.5-series models share the same 1M token limit, but Pro uses a different
// server-side processing pipeline that may handle large YouTube videos better.
const fallbackModel = "gemini-2.5-pro"

// ErrQuotaExhausted is returned by AnalyseVideo when the Gemini quota cannot be
// recovered after a 65s retry. The original 429 (no RetryInfo) could be either
// TPM throttle or daily RPD exhaustion — if a 65s wait didn't help, treat it as
// RPD. The quota resets at midnight Pacific Time.
var ErrQuotaExhausted = fmt.Errorf("Gemini quota limit persists after retry — likely daily RPD exhausted (resets at midnight PT)\n  run 'vger quota' to see today's usage; check https://aistudio.google.com for details")

const systemPromptBase = `You are an expert in cloud-native technology, Kubernetes, and the CNCF ecosystem.
You will be given a conference talk video. Analyse it and return a JSON object with exactly this schema:

{
  "speakers": ["<Full Name (Affiliation if known) — e.g. Jane Smith (Red Hat)>"],
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
Focus the technologies list on projects that are novel or worth learning more about. The notes field should be exhaustive.
For speakers: extract names from the video intro, title cards, and self-introductions. Include affiliation in parentheses where mentioned, e.g. "Jane Smith (Red Hat)". Each entry must be a person's name — do not add affiliations as separate entries.`

const systemPromptToolSuffix = `

You have two tools available:
- lookup_cncf_project(name): returns the current CNCF graduation stage from the live CNCF landscape
- validate_url(url): checks whether a URL is reachable

For each technology you identify:
1. Call lookup_cncf_project to get its current CNCF stage (do not rely on training data for this).
2. Call validate_url to verify the learn_more URL before including it; omit the URL if unreachable.
Once you have verified all technologies, produce the final JSON response.`

// maxToolRounds is the maximum number of function-calling iterations allowed per analysis.
const maxToolRounds = 10

// analysisResponse mirrors the JSON schema requested from the model.
type analysisResponse struct {
	Speakers     flexibleStringSlice  `json:"speakers"`
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
	APIKey     string
	Model      string
	cncfClient *cncf.Client // nil = no tools (single-shot mode)
	LogFn      func(string) // optional; called with status messages (e.g. model fallback warnings)
	QuotaFn    func()       // optional; called once per GenerateContent dispatch (successful or not)
	TokenFn    func(int64)  // optional; called with TotalTokenCount after each successful response
}

// New creates a Client without tool support (single-shot analysis).
func New(apiKey, model string) *Client {
	if model == "" {
		model = defaultModel
	}
	return &Client{APIKey: apiKey, Model: model}
}

// NewWithTools creates a Client that registers CNCF lookup and URL validation as
// Gemini function-calling tools, enabling a multi-turn ReAct analysis loop.
func NewWithTools(apiKey, model string, cncfClient *cncf.Client) *Client {
	c := New(apiKey, model)
	c.cncfClient = cncfClient
	return c
}

// overloadRetryFn returns an onRetry callback for withRetry that logs a
// user-visible warning when Gemini returns a 503 UNAVAILABLE (high demand).
// Returns nil when LogFn is not set (withRetry is nil-safe for onRetry).
func (c *Client) overloadRetryFn(maxAttempts int) func(attempt int, sleep time.Duration, err error) {
	if c.LogFn == nil {
		return nil
	}
	return func(attempt int, sleep time.Duration, err error) {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unavailable {
			c.LogFn(fmt.Sprintf("⚠ Gemini overloaded (attempt %d/%d, retrying in %s)", attempt, maxAttempts, sleep.Round(time.Second)))
		}
	}
}

// AnalyseVideo passes the video URL directly to the Gemini multimodal API
// and returns a structured report. No video download is performed.
//
// When a cncfClient was provided via NewWithTools, the analysis runs as a
// multi-turn function-calling loop where Gemini invokes tools to verify CNCF
// stages and validate URLs before producing its final answer.
//
// Error handling for oversized videos and quota:
//   - Token overflow (400 INVALID_ARGUMENT): all gemini-2.5 models share the same
//     1M token limit, so no fallback helps. Returns a clear user-facing error.
//   - Connection reset: Gemini dropped the TCP connection server-side. The Pro model
//     uses a different processing pipeline that may handle this better; retried once.
//   - Daily quota exhausted (429 with no RetryInfo): returns ErrQuotaExhausted so
//     callers can stop dispatching further requests immediately.
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

	report, err := c.analyseWithModel(ctx, client, meta, url, c.Model)
	if err != nil {
		switch {
		case isModelDeprecated(err):
			return nil, fmt.Errorf(
				"model %q is deprecated or unavailable — update with: vger scan --model <name>\n"+
					"  current recommended models: gemini-2.5-flash (default), gemini-2.5-pro",
				c.Model)
		case isQuotaExhausted(err):
			// 429 without RetryInfo is ambiguous: it may be TPM throttle (clears in
			// ~60s) or daily RPD exhaustion (doesn't clear until midnight Pacific).
			// Retry once after 65s — enough to clear the TPM rolling window.
			// If it fails again, treat as RPD exhaustion and stop dispatching.
			if c.LogFn != nil {
				c.LogFn("⏳ quota limit hit (TPM or RPD) — waiting 65s before retry")
			}
			select {
			case <-ctx.Done():
				return nil, ErrQuotaExhausted
			case <-time.After(65 * time.Second):
			}
			report, err = c.analyseWithModel(ctx, client, meta, url, c.Model)
			if err != nil {
				if isQuotaExhausted(err) {
					return nil, ErrQuotaExhausted
				}
				if isModelDeprecated(err) {
					return nil, fmt.Errorf("model %q is deprecated or unavailable — use gemini-2.5-flash or gemini-2.5-pro", c.Model)
				}
			}
		case isTokenOverflow(err):
			// Same 1M limit on all 2.5 models — no fallback can help.
			return nil, fmt.Errorf("video exceeds Gemini's 1M token limit and is too long to analyse (try a shorter clip)")
		case isConnectionReset(err) && c.Model != fallbackModel:
			// Server-side timeout; Pro pipeline may handle it better.
			if c.LogFn != nil {
				c.LogFn(fmt.Sprintf("connection reset on %s — retrying with %s", c.Model, fallbackModel))
			}
			report, err = c.analyseWithModel(ctx, client, meta, url, fallbackModel)
			if err != nil {
				if isQuotaExhausted(err) {
					return nil, ErrQuotaExhausted
				}
				if isTokenOverflow(err) {
					return nil, fmt.Errorf("video exceeds Gemini's 1M token limit and is too long to analyse (try a shorter clip)")
				}
			}
		}
	}
	return report, err
}

// analyseWithModel runs analysis for url using the specified model name,
// building the request contents fresh so that each attempt is independent.
func (c *Client) analyseWithModel(
	ctx context.Context,
	client *genai.Client,
	meta *domain.VideoMetadata,
	url string,
	model string,
) (*domain.Report, error) {
	prompt := systemPromptBase
	if c.cncfClient != nil {
		prompt += systemPromptToolSuffix
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
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

	if c.cncfClient != nil {
		return c.analyseWithTools(ctx, client, contents, meta, url, model)
	}
	return c.analyseSingleShot(ctx, client, contents, meta, url, model)
}

// analyseSingleShot performs a single GenerateContent call with JSON MIME type enforcement.
// Both the API call and JSON parsing are wrapped in withRetry so that truncated or
// malformed responses (common when the API is under load) are retried automatically.
func (c *Client) analyseSingleShot(
	ctx context.Context,
	client *genai.Client,
	contents []*genai.Content,
	meta *domain.VideoMetadata,
	url string,
	model string,
) (*domain.Report, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}
	const maxAttempts = 7
	var report *domain.Report
	err := withRetry(ctx, maxAttempts, c.overloadRetryFn(maxAttempts), func() error {
		if c.QuotaFn != nil {
			c.QuotaFn()
		}
		resp, callErr := client.Models.GenerateContent(ctx, model, contents, config)
		if callErr != nil {
			return callErr
		}
		if c.TokenFn != nil && resp.UsageMetadata != nil {
			c.TokenFn(int64(resp.UsageMetadata.TotalTokenCount))
		}
		var parseErr error
		report, parseErr = parseReport(resp.Text(), meta, url)
		return parseErr
	})
	if err != nil {
		return nil, fmt.Errorf("gemini generate content: %w", err)
	}
	return report, nil
}

// analyseWithTools runs a multi-turn function-calling loop, executing Gemini tool
// calls until the model produces a final text response, then parses it as JSON.
func (c *Client) analyseWithTools(
	ctx context.Context,
	client *genai.Client,
	contents []*genai.Content,
	meta *domain.VideoMetadata,
	url string,
	model string,
) (*domain.Report, error) {
	tools := newToolSet(c.cncfClient)
	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{FunctionDeclarations: tools.declarations},
		},
	}

	for round := 0; round < maxToolRounds; round++ {
		var resp *genai.GenerateContentResponse
		var fcs []*genai.FunctionCall
		var report *domain.Report

		const maxAttempts = 7
		err := withRetry(ctx, maxAttempts, c.overloadRetryFn(maxAttempts), func() error {
			if c.QuotaFn != nil {
				c.QuotaFn()
			}
			var callErr error
			resp, callErr = client.Models.GenerateContent(ctx, model, contents, config)
			if callErr != nil {
				return callErr
			}
			if c.TokenFn != nil && resp.UsageMetadata != nil {
				c.TokenFn(int64(resp.UsageMetadata.TotalTokenCount))
			}
			fcs = resp.FunctionCalls()
			if len(fcs) == 0 {
				// Final answer round — parse inside the retry so truncated JSON is retried.
				var parseErr error
				report, parseErr = parseReport(resp.Text(), meta, url)
				return parseErr
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("gemini generate content (round %d): %w", round+1, err)
		}

		if len(fcs) == 0 {
			return report, nil
		}

		// Append the model's message to the conversation.
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			contents = append(contents, resp.Candidates[0].Content)
		}

		// Execute each tool call and bundle all responses into one user turn.
		toolContent := &genai.Content{Role: genai.RoleUser}
		for _, fc := range fcs {
			result := tools.execute(ctx, fc)
			toolContent.Parts = append(toolContent.Parts,
				genai.NewPartFromFunctionResponse(fc.Name, result))
		}
		contents = append(contents, toolContent)
	}

	return nil, fmt.Errorf("tool call loop exceeded %d rounds without a final answer", maxToolRounds)
}

// parseReport extracts the domain.Report from the raw JSON text returned by the model.
func parseReport(raw string, meta *domain.VideoMetadata, url string) (*domain.Report, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty response from gemini")
	}

	raw = stripCodeFences(raw)
	raw = extractJSONObject(raw)
	raw = sanitizeJSONControls(raw)

	var ar analysisResponse
	if err := json.Unmarshal([]byte(raw), &ar); err != nil {
		return nil, fmt.Errorf("unmarshal gemini response: %w (raw: %.200s)", err, raw)
	}

	report := &domain.Report{
		VideoTitle:       meta.Title,
		VideoURL:         url,
		VideoDurationSec: meta.DurationSec,
		Speakers:         deduplicateSpeakers([]string(ar.Speakers)),
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

// extractJSONObject returns the substring from the first '{' to the last '}',
// discarding any preamble or postamble text the model may add around the JSON
// object when ResponseMIMEType is not enforced (e.g. in tool-calling mode).
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end <= start {
		return s // not a JSON object shape; return as-is and let unmarshal report the error
	}
	return s[start : end+1]
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

// sanitizeJSONControls replaces bare control characters inside JSON string
// literals with their valid JSON escape sequences. Gemini occasionally emits
// literal newlines or carriage returns inside strings (e.g. in the notes field),
// making the JSON technically invalid. This pass fixes them before unmarshal.
//
// It only rewrites characters found inside string values — control characters
// outside strings (whitespace between tokens) are left untouched.
func sanitizeJSONControls(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	inString := false
	escaped := false
	for _, ch := range s {
		if escaped {
			buf.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			buf.WriteRune(ch)
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			buf.WriteRune(ch)
			continue
		}
		if inString {
			switch ch {
			case '\n':
				buf.WriteString(`\n`)
			case '\r':
				buf.WriteString(`\r`)
			case '\t':
				buf.WriteString(`\t`)
			default:
				buf.WriteRune(ch)
			}
		} else {
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}

// deduplicateSpeakers removes duplicate and blank entries from the speakers list,
// using case-insensitive comparison. Preserves original casing of first occurrence.
func deduplicateSpeakers(speakers []string) []string {
	seen := make(map[string]struct{}, len(speakers))
	out := make([]string, 0, len(speakers))
	for _, s := range speakers {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}
