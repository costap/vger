package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/costap/vger/internal/domain"
)

// JSONCache implements domain.AnalysisCache by storing each analysis as a
// JSON file under a configurable directory, one file per video ID.
type JSONCache struct {
	dir string
}

// New creates a JSONCache that stores files in dir.
// The directory is created on first use if it does not exist.
func New(dir string) *JSONCache {
	return &JSONCache{dir: dir}
}

// DefaultDir returns the default cache directory: ~/.vger/cache.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".vger", "cache"), nil
}

// Save serialises the CachedAnalysis to <dir>/<video-id>.json.
func (c *JSONCache) Save(_ context.Context, entry *domain.CachedAnalysis) error {
	if err := os.MkdirAll(c.dir, 0o750); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	path := c.path(entry.VideoID)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}
	return nil
}

// Load reads <dir>/<video-id>.json and deserialises it.
// Returns (nil, nil) when no cached entry exists for the given ID.
func (c *JSONCache) Load(_ context.Context, videoID string) (*domain.CachedAnalysis, error) {
	data, err := os.ReadFile(c.path(videoID))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read cache file: %w", err)
	}

	var entry domain.CachedAnalysis
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cache entry: %w", err)
	}
	return &entry, nil
}

// LoadByVideoIDs loads multiple cached entries by video ID.
// Entries that are not found in the cache are silently skipped.
func (c *JSONCache) LoadByVideoIDs(ctx context.Context, videoIDs []string) ([]*domain.CachedAnalysis, error) {
	var results []*domain.CachedAnalysis
	for _, id := range videoIDs {
		entry, err := c.Load(ctx, id)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			results = append(results, entry)
		}
	}
	return results, nil
}

func (c *JSONCache) path(videoID string) string {
	return filepath.Join(c.dir, videoID+".json")
}
