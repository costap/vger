package cli

import (
	"fmt"

	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var listChannel string
var listSearch string
var listMax int64

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List videos from a YouTube channel",
	Long: `List videos from a YouTube channel ordered by publish date, newest first.
The channel can be specified as a channel ID (UCxx...) or a handle (@cncf / cncf).

Examples:
  vger list --channel @cncf
  vger list --channel @cncf --search kubecon
  vger list --channel UCvqbFHwN-nwalWPjPUKpvTA --search "kubecon 2024" --max 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listChannel == "" {
			return fmt.Errorf("--channel is required")
		}

		ytClient := youtube.New(youtubeAPIKey)

		ui.Status(fmt.Sprintf("Resolving channel: %s", listChannel))

		channelID, channelName, err := ytClient.ResolveChannel(cmd.Context(), listChannel)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Status(fmt.Sprintf("Channel: %s  [%s]", channelName, channelID))

		if listSearch != "" {
			ui.Status(fmt.Sprintf("Search filter: %q", listSearch))
		}

		ui.Status(fmt.Sprintf("Retrieving up to %d videos...", listMax))

		listings, err := ytClient.ListVideos(cmd.Context(), channelID, listSearch, listMax)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		if len(listings) == 0 {
			ui.Complete("No videos found matching the criteria.")
			return nil
		}

		ui.Complete(fmt.Sprintf("%d videos retrieved.", len(listings)))
		ui.SectionHeader(fmt.Sprintf("Videos — %s", channelName))
		fmt.Println()

		for i, v := range listings {
			ui.ListingRow(i+1, v.PublishedAt, v.Title, v.URL)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listChannel, "channel", "", "Channel ID (UCxx...) or handle (@name)")
	listCmd.Flags().StringVar(&listSearch, "search", "", "Filter videos by title/description keyword")
	listCmd.Flags().Int64Var(&listMax, "max", 25, "Maximum number of results to return (1-50)")
}
