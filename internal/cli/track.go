package cli

import (
	"fmt"
	"os"

	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

// resolveSignalStore returns the appropriate SignalStore implementation.
//
// When TECHDR_DIR is set (or the default tech-signals path exists) a
// MarkdownStore backed by that repo is returned; otherwise the JSON store
// at ~/.vger/signals/ is used.
func resolveSignalStore() (domain.SignalStore, error) {
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
			return nil, fmt.Errorf("TECHDR_DIR %q does not exist: %w", dir, err)
		}
		return signals.NewMarkdownStore(dir), nil
	}

	// Fall back to JSON store.
	jsonDir, err := signals.DefaultDir()
	if err != nil {
		return nil, fmt.Errorf("resolve signals dir: %w", err)
	}
	return signals.New(jsonDir), nil
}

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Track and manage technology signals",
	Long: `Capture, manage, and review technologies and ideas worth investigating.

Storage adapts automatically:
  • TECHDR_DIR set / tech-signals repo detected → Markdown files + git auto-commit
  • default                                     → JSON files at ~/.vger/signals/

Examples:
  vger track add                           # interactive capture
  vger track add --ai "tweet about eBPF…" # AI-assisted capture
  vger track list --status spotted         # browse your backlog
  vger track show 0001                     # view a signal in detail
  vger track enrich 0001                   # AI-enrich a signal
  vger track status 0001 evaluating        # update investigation status
  vger track link 0001 --video <url>       # link to a conference talk scan
  vger track digest                        # AI-powered backlog synthesis
  vger track digest --status spotted --enrich --output ~/review.md`,
}

func init() {
	trackCmd.AddCommand(trackAddCmd)
	trackCmd.AddCommand(trackListCmd)
	trackCmd.AddCommand(trackShowCmd)
	trackCmd.AddCommand(trackStatusCmd)
	trackCmd.AddCommand(trackLinkCmd)
	trackCmd.AddCommand(trackEnrichCmd)
	trackCmd.AddCommand(trackDigestCmd)
}
