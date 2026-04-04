package cli

import (
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Track and manage technology signals",
	Long: `Capture, manage, and review technologies and ideas worth investigating.

Signals are stored in ~/.vger/signals/ as JSON files and can be enriched
with AI context and linked to vger video scans.

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
