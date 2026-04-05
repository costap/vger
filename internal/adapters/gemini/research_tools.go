package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/costap/vger/internal/adapters/cncf"
	"github.com/costap/vger/internal/domain"
)

var askVideoDecl = &genai.FunctionDeclaration{
	Name: "ask_video",
	Description: "Ask a specific question about a cached conference talk. " +
		"Use when you need deeper detail from a particular video than the summary provides. " +
		"Returns the answer based on the full cached analysis of that talk.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"video_id": {
				Type:        genai.TypeString,
				Description: "The YouTube video ID of the cached talk to query",
			},
			"question": {
				Type:        genai.TypeString,
				Description: "The specific question to answer about this talk",
			},
		},
		Required: []string{"video_id", "question"},
	},
}

var searchCacheDecl = &genai.FunctionDeclaration{
	Name: "search_cache",
	Description: "Search the local cache for conference talks related to a query term. " +
		"Use to discover additional relevant videos not already in your evidence set. " +
		"Returns a list of matching video IDs and titles.",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"query": {
				Type:        genai.TypeString,
				Description: "Search term to find relevant cached talks, e.g. 'eBPF dataplane' or 'multi-cluster federation'",
			},
		},
		Required: []string{"query"},
	},
}

// researchToolSet holds Go implementations for the research investigation tools.
// It extends the base analysis tools (lookup_cncf_project, validate_url) with
// research-specific tools that let Gemini query cached videos and the local cache.
type researchToolSet struct {
	cncfClient    *cncf.Client
	hitsMap       map[string]*domain.CachedAnalysis // video_id → cached analysis (pre-loaded)
	cacheSearcher domain.CacheSearcher
	geminiClient  *Client
	declarations  []*genai.FunctionDeclaration
}

func newResearchToolSet(
	cncfClient *cncf.Client,
	hits []*domain.CachedAnalysis,
	cacheSearcher domain.CacheSearcher,
	geminiClient *Client,
) *researchToolSet {
	hitsMap := make(map[string]*domain.CachedAnalysis, len(hits))
	for _, h := range hits {
		hitsMap[h.VideoID] = h
	}
	return &researchToolSet{
		cncfClient:    cncfClient,
		hitsMap:       hitsMap,
		cacheSearcher: cacheSearcher,
		geminiClient:  geminiClient,
		declarations: []*genai.FunctionDeclaration{
			askVideoDecl,
			searchCacheDecl,
			lookupCNCFDecl,
			validateURLDecl,
		},
	}
}

// execute dispatches a Gemini FunctionCall to the appropriate Go implementation.
func (t *researchToolSet) execute(ctx context.Context, fc *genai.FunctionCall) map[string]any {
	switch fc.Name {
	case "ask_video":
		return t.executeAskVideo(ctx, fc)

	case "search_cache":
		return t.executeSearchCache(ctx, fc)

	case "lookup_cncf_project":
		name, _ := fc.Args["name"].(string)
		if t.cncfClient == nil {
			return map[string]any{"found": false, "stage": ""}
		}
		stage, found := t.cncfClient.LookupProject(ctx, name)
		if !found {
			return map[string]any{"found": false, "stage": ""}
		}
		return map[string]any{"found": true, "stage": stage}

	case "validate_url":
		url, _ := fc.Args["url"].(string)
		if t.cncfClient == nil {
			return map[string]any{"reachable": false}
		}
		reachable := t.cncfClient.ValidateURL(ctx, url)
		return map[string]any{"reachable": reachable}

	default:
		return map[string]any{"error": "unknown function: " + fc.Name}
	}
}

func (t *researchToolSet) executeAskVideo(ctx context.Context, fc *genai.FunctionCall) map[string]any {
	videoID, _ := fc.Args["video_id"].(string)
	question, _ := fc.Args["question"].(string)

	if videoID == "" || question == "" {
		return map[string]any{"error": "video_id and question are required"}
	}

	cached, ok := t.hitsMap[videoID]
	if !ok {
		return map[string]any{"error": fmt.Sprintf("video %q not found in cache", videoID)}
	}

	answer, err := t.geminiClient.Ask(ctx, question, cached, "")
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("ask failed: %s", err.Error())}
	}

	return map[string]any{
		"video_title": cached.Report.VideoTitle,
		"answer":      answer,
	}
}

func (t *researchToolSet) executeSearchCache(ctx context.Context, fc *genai.FunctionCall) map[string]any {
	query, _ := fc.Args["query"].(string)
	if query == "" {
		return map[string]any{"error": "query is required"}
	}

	if t.cacheSearcher == nil {
		return map[string]any{"results": []any{}, "note": "cache search unavailable"}
	}

	results, err := t.cacheSearcher.Search(ctx, query, 5)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("search failed: %s", err.Error())}
	}

	items := make([]map[string]string, 0, len(results))
	for _, r := range results {
		items = append(items, map[string]string{
			"video_id": r.VideoID,
			"title":    r.Report.VideoTitle,
		})
		// Register newly discovered videos so ask_video can query them.
		if _, known := t.hitsMap[r.VideoID]; !known {
			t.hitsMap[r.VideoID] = r
		}
	}

	// Convert to []any for genai serialisation.
	out := make([]any, len(items))
	for i, item := range items {
		out[i] = item
	}

	if len(out) == 0 {
		return map[string]any{"results": out, "note": "no additional results found"}
	}

	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = fmt.Sprintf("%s (%s)", item["title"], item["video_id"])
	}
	return map[string]any{
		"results": out,
		"summary": fmt.Sprintf("Found %d talks: %s", len(items), strings.Join(titles, "; ")),
	}
}
