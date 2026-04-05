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

const defaultModel = "gemini-2.5-flash"

const systemPromptBase = `You are an expert in cloud-native technology, Kubernetes, and the CNCF ecosystem.
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

// AnalyseVideo passes the video URL directly to the Gemini multimodal API
// and returns a structured report. No video download is performed.
//
// When a cncfClient was provided via NewWithTools, the analysis runs as a
// multi-turn function-calling loop where Gemini invokes tools to verify CNCF
// stages and validate URLs before producing its final answer.
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
		return c.analyseWithTools(ctx, client, contents, meta, url)
	}
	return c.analyseSingleShot(ctx, client, contents, meta, url)
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
) (*domain.Report, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
	}
	var report *domain.Report
	err := withRetry(ctx, 5, func() error {
		resp, callErr := client.Models.GenerateContent(ctx, c.Model, contents, config)
		if callErr != nil {
			return callErr
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

		err := withRetry(ctx, 5, func() error {
			var callErr error
			resp, callErr = client.Models.GenerateContent(ctx, c.Model, contents, config)
			if callErr != nil {
				return callErr
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
