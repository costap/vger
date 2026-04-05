package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/domain"
)

// channelEntry is persisted in ~/.vger/channels.json.
type channelEntry struct {
	Handle string `json:"handle"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

func channelHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".vger", "channels.json"), nil
}

// loadChannelHistory reads previously resolved channel handles from the history file.
func loadChannelHistory() []channelEntry {
	p, err := channelHistoryPath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var entries []channelEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// saveChannelHistory adds or updates an entry in the channel history file.
// Errors are silently ignored — this is a best-effort quality-of-life feature.
func saveChannelHistory(handle, id, name string) {
	if handle == "" || id == "" {
		return
	}
	p, err := channelHistoryPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return
	}

	existing := loadChannelHistory()
	updated := false
	for i, e := range existing {
		if e.Handle == handle || e.ID == id {
			existing[i] = channelEntry{Handle: handle, ID: id, Name: name}
			updated = true
			break
		}
	}
	if !updated {
		existing = append(existing, channelEntry{Handle: handle, ID: id, Name: name})
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0o640)
}

// channelCompletionFunc is a Cobra completion function for --channel flags.
// It merges the curated well-known list with the user's personal channel history.
func channelCompletionFunc(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	seen := make(map[string]bool)
	var results []string

	// History entries first (personal channels are more likely to be what the user wants).
	for _, e := range loadChannelHistory() {
		handle := e.Handle
		if handle == "" {
			handle = e.ID
		}
		if seen[handle] {
			continue
		}
		seen[handle] = true
		entry := handle
		if e.Name != "" {
			entry = handle + "\t" + e.Name
		}
		if toComplete == "" || strings.HasPrefix(strings.ToLower(handle), strings.ToLower(toComplete)) {
			results = append(results, entry)
		}
	}

	// Well-known channels.
	for _, ch := range wellKnownChannels {
		handle := ch
		if idx := strings.IndexByte(ch, '\t'); idx >= 0 {
			handle = ch[:idx]
		}
		if seen[handle] {
			continue
		}
		seen[handle] = true
		if toComplete == "" || strings.HasPrefix(strings.ToLower(handle), strings.ToLower(toComplete)) {
			results = append(results, ch)
		}
	}

	return results, cobra.ShellCompDirectiveNoFileComp
}

// minimalCacheEntry is the subset of a cache JSON file needed for completion.
type minimalCacheEntry struct {
	Metadata struct {
		URL   string `json:"URL"`
		Title string `json:"Title"`
	} `json:"metadata"`
}

// cachedVideoCompletionFunc is a Cobra ValidArgsFunction for commands that accept
// a YouTube video URL as a positional argument. It scans the local cache directory
// and returns one completion entry per cached video, with the title as a hint.
func cachedVideoCompletionFunc(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first positional argument (the URL).
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cacheDir := filepath.Join(home, ".vger", "cache")

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var results []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || e.Name() == "cncf_landscape.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(cacheDir, e.Name()))
		if err != nil {
			continue
		}
		var entry minimalCacheEntry
		if err := json.Unmarshal(data, &entry); err != nil || entry.Metadata.URL == "" {
			continue
		}
		completion := entry.Metadata.URL
		if entry.Metadata.Title != "" {
			completion = fmt.Sprintf("%s\t%s", entry.Metadata.URL, entry.Metadata.Title)
		}
		results = append(results, completion)
	}

	return results, cobra.ShellCompDirectiveNoFileComp
}

// signalIDCompletionFunc is a Cobra ValidArgsFunction for track sub-commands that
// accept a signal ID as the first positional argument. It reads the active signal
// store and returns "id\ttitle [status]" entries.
func signalIDCompletionFunc(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first positional argument (the signal ID).
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store, err := loadSignalStoreForCompletion()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	if cmd != nil {
		ctx = cmd.Context()
	}

	sigs, err := store.LoadAll(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	results := make([]string, 0, len(sigs))
	for _, s := range sigs {
		hint := fmt.Sprintf("%s\t%s [%s]", s.ID, s.Title, s.Status)
		results = append(results, hint)
	}
	return results, cobra.ShellCompDirectiveNoFileComp
}

// trackStatusCompletionFunc is a Cobra ValidArgsFunction for `vger track status`.
// For the first arg (signal ID) it delegates to signalIDCompletionFunc.
// For the second arg it returns static status values.
func trackStatusCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return signalIDCompletionFunc(cmd, args, toComplete)
	}
	if len(args) == 1 {
		return []string{
			"evaluating\tCurrently being evaluated",
			"adopted\tDecided to adopt",
			"rejected\tDecided not to adopt",
			"monitoring\tKeeping an eye on it",
			"on-hold\tPaused evaluation",
			"spotted\tJust noticed, not yet evaluated",
		}, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// loadSignalStoreForCompletion returns a domain.SignalStore suitable for
// reading during tab completion. Mirrors the logic in track/cmd.go.
func loadSignalStoreForCompletion() (domain.SignalStore, error) {
	dir := os.Getenv("TECHDR_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		candidate := home + "/code/github.com/costap/tech-signals"
		if _, err := os.Stat(candidate + "/.next-id"); err == nil {
			dir = candidate
		}
	}
	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			return nil, fmt.Errorf("TECHDR_DIR %q does not exist", dir)
		}
		return signals.NewMarkdownStore(dir), nil
	}

	jsonDir, err := signals.DefaultDir()
	if err != nil {
		return nil, err
	}
	return signals.New(jsonDir), nil
}
