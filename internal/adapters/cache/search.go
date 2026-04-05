package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/costap/vger/internal/domain"
)

// Search returns up to maxResults cached analyses ranked by relevance to query.
// Scoring weights (case-insensitive substring matching):
//   - Technology.Name exact match   → 3 pts each
//   - Technology.Name contains query → 2 pts each
//   - Report.Summary contains query  → 2 pts
//   - Report.Notes contains query    → 1 pt
func (c *JSONCache) Search(_ context.Context, query string, maxResults int) ([]*domain.CachedAnalysis, error) {
	entries, err := os.ReadDir(c.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)

	type scored struct {
		entry *domain.CachedAnalysis
		score int
	}

	var results []scored

	for _, de := range entries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".json") {
			continue
		}
		// Skip the CNCF landscape cache file.
		if de.Name() == "cncf_landscape.json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.dir, de.Name()))
		if err != nil {
			continue
		}

		var entry domain.CachedAnalysis
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		score := scoreEntry(&entry, q)
		if score > 0 {
			results = append(results, scored{entry: &entry, score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	out := make([]*domain.CachedAnalysis, len(results))
	for i, r := range results {
		out[i] = r.entry
	}
	return out, nil
}

func scoreEntry(entry *domain.CachedAnalysis, q string) int {
	score := 0
	for _, tech := range entry.Report.Technologies {
		name := strings.ToLower(tech.Name)
		if name == q {
			score += 3
		} else if strings.Contains(name, q) {
			score += 2
		}
	}
	if strings.Contains(strings.ToLower(entry.Metadata.Title), q) {
		score += 2
	}
	for _, p := range entry.Playlists {
		if strings.Contains(strings.ToLower(p.PlaylistTitle), q) {
			score += 2
			break // count once even if multiple playlists match
		}
	}
	if strings.Contains(strings.ToLower(entry.Report.Summary), q) {
		score += 2
	}
	if strings.Contains(strings.ToLower(entry.Report.Notes), q) {
		score += 1
	}
	return score
}
