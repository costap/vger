package signals

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/costap/vger/internal/domain"
)

// JSONStore implements domain.SignalStore by persisting each signal as a JSON
// file under a configurable directory, one file per signal ID.
type JSONStore struct {
	dir string
}

// New creates a JSONStore that stores files in dir.
func New(dir string) *JSONStore {
	return &JSONStore{dir: dir}
}

// DefaultDir returns the default signals directory: ~/.vger/signals.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".vger", "signals"), nil
}

// Save serialises the signal to <dir>/<id>.json.
func (s *JSONStore) Save(_ context.Context, signal *domain.Signal) error {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return fmt.Errorf("create signals dir: %w", err)
	}
	data, err := json.MarshalIndent(signal, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal signal: %w", err)
	}
	if err := os.WriteFile(s.path(signal.ID), data, 0o640); err != nil {
		return fmt.Errorf("write signal file: %w", err)
	}
	return nil
}

// Load reads <dir>/<id>.json and deserialises it.
// Returns (nil, nil) when no signal exists for the given ID.
func (s *JSONStore) Load(_ context.Context, id string) (*domain.Signal, error) {
	data, err := os.ReadFile(s.path(id))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read signal file: %w", err)
	}
	var sig domain.Signal
	if err := json.Unmarshal(data, &sig); err != nil {
		return nil, fmt.Errorf("unmarshal signal: %w", err)
	}
	return &sig, nil
}

// LoadAll returns all signals sorted by ID ascending.
func (s *JSONStore) LoadAll(ctx context.Context) ([]*domain.Signal, error) {
	return s.loadWhere(ctx, nil)
}

// LoadByStatus returns all signals matching the given status.
func (s *JSONStore) LoadByStatus(ctx context.Context, status string) ([]*domain.Signal, error) {
	return s.loadWhere(ctx, func(sig *domain.Signal) bool {
		return sig.Status == status
	})
}

// LoadByCategory returns all signals matching the given category.
func (s *JSONStore) LoadByCategory(ctx context.Context, category string) ([]*domain.Signal, error) {
	return s.loadWhere(ctx, func(sig *domain.Signal) bool {
		return sig.Category == category
	})
}

// NextID returns the next available zero-padded 4-digit ID (e.g. "0042").
// It scans existing files and returns max+1, or "0001" if the store is empty.
func (s *JSONStore) NextID(_ context.Context) (string, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return "0001", nil
	}
	if err != nil {
		return "", fmt.Errorf("read signals dir: %w", err)
	}

	max := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".json")
		n, err := strconv.Atoi(base)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return fmt.Sprintf("%04d", max+1), nil
}

// Delete removes the signal file for the given ID. No-ops if the file does not exist.
func (s *JSONStore) Delete(_ context.Context, id string) error {
	err := os.Remove(s.path(id))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// loadWhere reads all JSON files and applies an optional filter predicate.
func (s *JSONStore) loadWhere(ctx context.Context, predicate func(*domain.Signal) bool) ([]*domain.Signal, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read signals dir: %w", err)
	}

	var signals []*domain.Signal
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		sig, err := s.Load(ctx, id)
		if err != nil {
			return nil, err
		}
		if sig == nil {
			continue
		}
		if predicate == nil || predicate(sig) {
			signals = append(signals, sig)
		}
	}

	sort.Slice(signals, func(i, j int) bool {
		return signals[i].ID < signals[j].ID
	})
	return signals, nil
}

func (s *JSONStore) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}
