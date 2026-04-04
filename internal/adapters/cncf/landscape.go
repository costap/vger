package cncf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/costap/vger/internal/domain"
)

const (
	landscapeURL    = "https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml"
	cacheFileName   = "cncf_landscape.json"
	cacheTTL        = 24 * time.Hour
	urlCheckTimeout = 10 * time.Second
)

// ProjectInfo holds rich metadata for a CNCF landscape project.
type ProjectInfo struct {
	Name     string `json:"name"`
	Stage    string `json:"stage"`    // graduated | incubating | sandbox
	Category string `json:"category"` // top-level landscape category
	Homepage string `json:"homepage"`
}

// landscapeCache is the on-disk format of the cached CNCF project lookup.
type landscapeCache struct {
	CachedAt     time.Time         `json:"cached_at"`
	Projects     map[string]string `json:"projects"`      // normalised name → stage (kept for Enrich)
	FullProjects []ProjectInfo     `json:"full_projects"` // rich records for research / topic lookup
}

// Client fetches and caches the CNCF landscape data, and enriches reports with
// accurate stage data and validated learn_more URLs.
type Client struct {
	cacheDir string
}

// New creates a Client that stores its landscape cache in cacheDir.
func New(cacheDir string) *Client {
	return &Client{cacheDir: cacheDir}
}

// Enrich updates the technologies in report with accurate CNCF stage data and
// validates each learn_more URL, clearing ones that are unreachable.
func (c *Client) Enrich(ctx context.Context, report *domain.Report) error {
	projects, err := c.loadOrFetch(ctx)
	if err != nil {
		// Enrichment is best-effort; do not fail the scan.
		return nil
	}

	httpClient := &http.Client{Timeout: urlCheckTimeout}

	for i := range report.Technologies {
		t := &report.Technologies[i]

		// Correct CNCF stage from live landscape data.
		if stage, found := lookupStage(projects, t.Name); found {
			t.CNCFStage = stage
		} else if t.CNCFStage != "" {
			// Gemini assigned a stage but the project is not in the landscape —
			// trust the landscape and clear the stale classification.
			t.CNCFStage = ""
		}

		// Validate the learn_more URL.
		if t.LearnMore != "" && !urlReachable(ctx, httpClient, t.LearnMore) {
			t.LearnMore = ""
		}
	}
	return nil
}

// loadOrFetch returns the simple name→stage map (used by Enrich and tool-calling).
func (c *Client) loadOrFetch(ctx context.Context) (map[string]string, error) {
	lc, err := c.loadOrFetchFull(ctx)
	if err != nil {
		return nil, err
	}
	return lc.Projects, nil
}

// loadOrFetchFull returns the complete landscapeCache, fetching from GitHub if stale.
func (c *Client) loadOrFetchFull(ctx context.Context) (*landscapeCache, error) {
	cachePath := filepath.Join(c.cacheDir, cacheFileName)

	// Try loading from disk.
	if data, err := os.ReadFile(cachePath); err == nil {
		var lc landscapeCache
		if json.Unmarshal(data, &lc) == nil && time.Since(lc.CachedAt) < cacheTTL {
			return &lc, nil
		}
	}

	// Fetch fresh from GitHub.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, landscapeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch landscape: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read landscape: %w", err)
	}

	projects, full, err := parseYAML(body)
	if err != nil {
		return nil, fmt.Errorf("parse landscape: %w", err)
	}

	lc := &landscapeCache{CachedAt: time.Now(), Projects: projects, FullProjects: full}

	// Persist to disk.
	if err := os.MkdirAll(c.cacheDir, 0o750); err == nil {
		if data, err := json.MarshalIndent(lc, "", "  "); err == nil {
			_ = os.WriteFile(cachePath, data, 0o640)
		}
	}

	return lc, nil
}

// parseYAML recursively walks the landscape YAML and collects all nodes that
// have both a "name" and a "project" key (the latter indicates CNCF membership).
// Returns both the simple name→stage map (for Enrich) and the full ProjectInfo slice.
func parseYAML(data []byte) (map[string]string, []ProjectInfo, error) {
	var root interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, nil, err
	}
	simple := make(map[string]string)
	var full []ProjectInfo
	walkNode(root, "", simple, &full)
	return simple, full, nil
}

func walkNode(v interface{}, category string, out map[string]string, full *[]ProjectInfo) {
	switch val := v.(type) {
	case map[string]interface{}:
		// Track category context for items below this node.
		if cat, ok := val["category"].(string); ok && cat != "" {
			category = cat
		}
		name, hasName := val["name"].(string)
		project, hasProject := val["project"].(string)
		if hasName && hasProject && project != "" {
			out[normalise(name)] = project
			homepage, _ := val["homepage_url"].(string)
			*full = append(*full, ProjectInfo{
				Name:     name,
				Stage:    project,
				Category: category,
				Homepage: homepage,
			})
		}
		for _, child := range val {
			walkNode(child, category, out, full)
		}
	case []interface{}:
		for _, item := range val {
			walkNode(item, category, out, full)
		}
	}
}

// lookupStage finds the CNCF stage for a technology name using normalised matching.
func lookupStage(projects map[string]string, name string) (string, bool) {
	key := normalise(name)
	if stage, ok := projects[key]; ok {
		return stage, true
	}
	// Try prefix matching for compound names (e.g. "OpenTelemetry Collector" → "opentelemetry")
	for k, stage := range projects {
		if strings.HasPrefix(key, k) || strings.HasPrefix(k, key) {
			return stage, true
		}
	}
	return "", false
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]`)

// normalise lowercases and strips all non-alphanumeric characters for fuzzy matching.
func normalise(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "")
	// Strip common suffixes that appear inconsistently.
	for _, suffix := range []string{"io", "cncf", "project"} {
		if strings.HasSuffix(s, suffix) && len(s) > len(suffix) {
			s = strings.TrimSuffix(s, suffix)
		}
	}
	return s
}

// LookupProject returns the CNCF graduation stage for the given project name.
// Returns ("", false) if the project is not found in the landscape.
// The landscape data is loaded from the local cache or fetched from GitHub if stale.
func (c *Client) LookupProject(ctx context.Context, name string) (stage string, found bool) {
	projects, err := c.loadOrFetch(ctx)
	if err != nil {
		return "", false
	}
	return lookupStage(projects, name)
}

// ValidateURL returns true if the URL responds with a 2xx or 3xx status code within
// urlCheckTimeout.
func (c *Client) ValidateURL(ctx context.Context, url string) bool {
	hc := &http.Client{Timeout: urlCheckTimeout}
	return urlReachable(ctx, hc, url)
}

// urlReachable performs a HEAD request and returns true if the URL responds with
// a 2xx or 3xx status code within urlCheckTimeout.
func urlReachable(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 400
}

// LookupByTopic returns up to 15 CNCF projects whose name, category, or homepage
// contains the given topic string (case-insensitive). Used by vger research.
func (c *Client) LookupByTopic(ctx context.Context, topic string) []ProjectInfo {
	lc, err := c.loadOrFetchFull(ctx)
	if err != nil {
		return nil
	}

	q := strings.ToLower(topic)
	const maxResults = 15

	var matches []ProjectInfo
	for _, p := range lc.FullProjects {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Category), q) ||
			strings.Contains(strings.ToLower(p.Homepage), q) {
			matches = append(matches, p)
			if len(matches) >= maxResults {
				break
			}
		}
	}
	return matches
}
