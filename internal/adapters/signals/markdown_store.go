package signals

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/costap/vger/internal/domain"
)

// MarkdownStore implements domain.SignalStore by persisting each signal as a
// TechDR Markdown file under $TECHDR_DIR/signals/YYYY/, one file per signal.
// After each mutation it runs git add + git commit in the repo root so the
// tech-signals git history stays current.
type MarkdownStore struct {
	repoDir    string // absolute path to the tech-signals repo root
	signalsDir string // <repoDir>/signals
	nextIDFile string // <repoDir>/.next-id
}

// NewMarkdownStore creates a MarkdownStore rooted at repoDir.
func NewMarkdownStore(repoDir string) *MarkdownStore {
	return &MarkdownStore{
		repoDir:    repoDir,
		signalsDir: filepath.Join(repoDir, "signals"),
		nextIDFile: filepath.Join(repoDir, ".next-id"),
	}
}

// ── SignalStore interface ─────────────────────────────────────────────────────

// Save writes the signal as a TechDR Markdown file and commits it to git.
// The file path is <signalsDir>/<YYYY>/<id>-<date>-<slug>.md.
// If the file already exists (update), it overwrites it.
func (m *MarkdownStore) Save(_ context.Context, sig *domain.Signal) error {
	year := sig.Date[:4]
	if year == "" {
		year = time.Now().Format("2006")
	}

	yearDir := filepath.Join(m.signalsDir, year)
	if err := os.MkdirAll(yearDir, 0o750); err != nil {
		return fmt.Errorf("create signals year dir: %w", err)
	}

	// Find existing file for this ID (may have different slug).
	existing := m.findFile(sig.ID)
	var filePath string
	if existing != "" {
		filePath = existing
	} else {
		slug := slugify(sig.Title)
		if slug == "" {
			slug = "signal"
		}
		filePath = filepath.Join(yearDir, fmt.Sprintf("%s-%s-%s.md", sig.ID, sig.Date, slug))
	}

	content := renderSignalToMarkdown(sig)
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write signal file: %w", err)
	}

	commitMsg := m.commitMessage(sig, existing == "")
	if err := m.gitCommit(filePath, commitMsg); err != nil {
		// Non-fatal: file is written, git failure is a warning.
		fmt.Fprintf(os.Stderr, "warning: git commit failed: %v\n", err)
	}

	return nil
}

// Load retrieves a signal by ID. Returns (nil, nil) if not found.
func (m *MarkdownStore) Load(_ context.Context, id string) (*domain.Signal, error) {
	path := m.findFile(id)
	if path == "" {
		return nil, nil
	}
	return parseSignalFromMarkdown(path)
}

// LoadAll returns all signals sorted by ID ascending.
func (m *MarkdownStore) LoadAll(ctx context.Context) ([]*domain.Signal, error) {
	return m.loadWhere(ctx, nil)
}

// LoadByStatus returns all signals matching the given status.
func (m *MarkdownStore) LoadByStatus(ctx context.Context, status string) ([]*domain.Signal, error) {
	return m.loadWhere(ctx, func(s *domain.Signal) bool { return s.Status == status })
}

// LoadByCategory returns all signals matching the given category.
func (m *MarkdownStore) LoadByCategory(ctx context.Context, category string) ([]*domain.Signal, error) {
	return m.loadWhere(ctx, func(s *domain.Signal) bool { return s.Category == category })
}

// NextID returns the next zero-padded 4-digit ID from the .next-id file.
// If the file doesn't exist, starts at "0001".
func (m *MarkdownStore) NextID(_ context.Context) (string, error) {
	raw, err := os.ReadFile(m.nextIDFile)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read .next-id: %w", err)
	}

	id := "0001"
	if len(raw) > 0 {
		id = fmt.Sprintf("%04d", parseIDInt(strings.TrimSpace(string(raw))))
	}
	return id, nil
}

// bumpNextID increments the .next-id counter.
func (m *MarkdownStore) bumpNextID() error {
	raw, _ := os.ReadFile(m.nextIDFile)
	current := 1
	if len(raw) > 0 {
		current = parseIDInt(strings.TrimSpace(string(raw)))
	}
	next := fmt.Sprintf("%04d\n", current+1)
	return os.WriteFile(m.nextIDFile, []byte(next), 0o644)
}

// Delete removes a signal file via git rm and commits.
func (m *MarkdownStore) Delete(_ context.Context, id string) error {
	path := m.findFile(id)
	if path == "" {
		return nil // idempotent
	}

	cmd := exec.Command("git", "rm", "-f", path)
	cmd.Dir = m.repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fall back to plain delete if git rm fails.
		if rmErr := os.Remove(path); rmErr != nil {
			return fmt.Errorf("git rm failed (%s), plain remove also failed: %w", string(out), rmErr)
		}
	}

	_ = m.gitCommitFiles([]string{path}, fmt.Sprintf("delete: %s", id))
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// loadWhere loads all signals and applies an optional predicate filter.
func (m *MarkdownStore) loadWhere(_ context.Context, predicate func(*domain.Signal) bool) ([]*domain.Signal, error) {
	var paths []string

	err := filepath.WalkDir(m.signalsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".gitkeep") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("walk signals dir: %w", err)
	}

	sort.Strings(paths)

	var signals []*domain.Signal
	for _, p := range paths {
		sig, err := parseSignalFromMarkdown(p)
		if err != nil || sig == nil {
			continue
		}
		if predicate == nil || predicate(sig) {
			signals = append(signals, sig)
		}
	}
	return signals, nil
}

// findFile scans for the first Markdown file whose base name starts with id.
func (m *MarkdownStore) findFile(id string) string {
	var found string
	_ = filepath.WalkDir(m.signalsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || found != "" {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, id) && strings.HasSuffix(base, ".md") {
			found = path
		}
		return nil
	})
	return found
}

// commitMessage returns a git commit message matching techdr.sh conventions.
func (m *MarkdownStore) commitMessage(sig *domain.Signal, isNew bool) string {
	if isNew {
		return fmt.Sprintf("signal: %s %s", sig.ID, sig.Title)
	}
	if sig.Enrichment != nil && sig.Enrichment.WhatItIs != "" {
		return fmt.Sprintf("enrich: %s %s", sig.ID, sig.Title)
	}
	return fmt.Sprintf("update: %s %s", sig.ID, sig.Title)
}

// gitCommit stages a single file and commits with the given message.
func (m *MarkdownStore) gitCommit(filePath, message string) error {
	return m.gitCommitFiles([]string{filePath, m.nextIDFile}, message)
}

// gitCommitFiles stages the given paths and commits.
func (m *MarkdownStore) gitCommitFiles(paths []string, message string) error {
	addArgs := append([]string{"add"}, paths...)
	add := exec.Command("git", addArgs...)
	add.Dir = m.repoDir
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %s: %w", strings.TrimSpace(string(out)), err)
	}

	commit := exec.Command("git", "commit", "-m", message, "--allow-empty")
	commit.Dir = m.repoDir
	if out, err := commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SaveStatusChange commits with a "status: old → new" message.
func (m *MarkdownStore) SaveStatusChange(ctx context.Context, sig *domain.Signal, oldStatus string) error {
	if err := m.writeFile(sig); err != nil {
		return err
	}
	filePath := m.findFile(sig.ID)
	msg := fmt.Sprintf("status: %s %s → %s", sig.ID, oldStatus, sig.Status)
	return m.gitCommitFiles([]string{filePath}, msg)
}

// writeFile writes the Markdown without committing.
func (m *MarkdownStore) writeFile(sig *domain.Signal) error {
	existing := m.findFile(sig.ID)
	if existing == "" {
		return fmt.Errorf("signal %s not found", sig.ID)
	}
	return os.WriteFile(existing, []byte(renderSignalToMarkdown(sig)), 0o644)
}

// BumpID increments .next-id after a successful add.
func (m *MarkdownStore) BumpID() error {
	return m.bumpNextID()
}

// IsMarkdownStore returns true — used by CLI to enable editor review.
func (m *MarkdownStore) IsMarkdownStore() bool { return true }

// FilePath returns the Markdown file path for a signal ID (for editor open).
func (m *MarkdownStore) FilePath(id string) string {
	return m.findFile(id)
}

// parseIDInt converts "0042" to 42.
func parseIDInt(s string) int {
	n, err := strconv.Atoi(strings.TrimLeft(s, "0"))
	if err != nil || n == 0 {
		return 1
	}
	return n
}
