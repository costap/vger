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
var listPlaylists bool
var listPlaylist string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List videos or playlists from a YouTube channel",
	Long: `List videos or playlists from a YouTube channel ordered by publish date, newest first.
The channel can be specified as a channel ID (UCxx...) or a handle (@cncf / cncf).

By default, videos are listed. Use --playlists to list playlists instead.
Use --playlist to list the videos inside a specific playlist.

Video results are retrieved by walking the channel's complete uploads history and
filtering client-side — so results are comprehensive, not limited by YouTube's
search index. All matching videos will be found regardless of age or popularity.

Examples:
  vger list --channel @cncf
  vger list --channel @cncf --playlists
  vger list --channel @cncf --playlists --search kubecon
  vger list --channel @cncf --search argocon
  vger list --playlist "https://www.youtube.com/playlist?list=PLj6h78yzYM2P-3T82bF...a"
  vger list --playlist PLj6h78yzYM2P...a --search "service mesh"
  vger list --channel UCvqbFHwN-nwalWPjPUKpvTA --search "kubecon 2024" --max 100`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ytClient := youtube.New(youtubeAPIKey)

		// --playlist takes precedence — no channel required
		if listPlaylist != "" {
			return runListPlaylistVideos(cmd, ytClient)
		}

		if listChannel == "" {
			return fmt.Errorf("--channel or --playlist is required")
		}

		ui.Status(fmt.Sprintf("Resolving channel: %s", listChannel))

		channelID, channelName, err := ytClient.ResolveChannel(cmd.Context(), listChannel)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Status(fmt.Sprintf("Channel: %s  [%s]", channelName, channelID))

		if listPlaylists {
			return runListPlaylists(cmd, ytClient, channelID, channelName)
		}
		return runListVideos(cmd, ytClient, channelID, channelName)
	},
}

func runListVideos(cmd *cobra.Command, ytClient *youtube.Client, channelID, channelName string) error {
	if listSearch != "" {
		ui.Status(fmt.Sprintf("Search filter: %q (scanning full upload history...)", listSearch))
	} else {
		ui.Status(fmt.Sprintf("Retrieving up to %d most recent videos...", listMax))
	}

	listings, scanned, err := ytClient.ListVideosDetailed(cmd.Context(), channelID, listSearch, listMax)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if len(listings) == 0 {
		if listSearch != "" {
			ui.Complete(fmt.Sprintf("No videos found matching %q (scanned %d videos).", listSearch, scanned))
		} else {
			ui.Complete("No videos found.")
		}
		return nil
	}

	if listSearch != "" {
		ui.Complete(fmt.Sprintf("Found %d matching videos (scanned %d total).", len(listings), scanned))
	} else {
		ui.Complete(fmt.Sprintf("%d videos retrieved.", len(listings)))
	}

	ui.SectionHeader(fmt.Sprintf("Videos — %s", channelName))
	fmt.Println()
	for i, v := range listings {
		ui.ListingRow(i+1, v.PublishedAt, v.Title, v.URL)
	}
	fmt.Println()
	return nil
}

func runListPlaylists(cmd *cobra.Command, ytClient *youtube.Client, channelID, channelName string) error {
	if listSearch != "" {
		ui.Status(fmt.Sprintf("Search filter: %q", listSearch))
	}
	ui.Status(fmt.Sprintf("Retrieving up to %d playlists...", listMax))

	playlists, err := ytClient.ListPlaylists(cmd.Context(), channelID, listSearch, listMax)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if len(playlists) == 0 {
		if listSearch != "" {
			ui.Complete(fmt.Sprintf("No playlists found matching %q.", listSearch))
		} else {
			ui.Complete("No playlists found.")
		}
		return nil
	}

	ui.Complete(fmt.Sprintf("%d playlists retrieved.", len(playlists)))
	ui.SectionHeader(fmt.Sprintf("Playlists — %s", channelName))
	fmt.Println()
	for i, p := range playlists {
		ui.PlaylistRow(i+1, p.PublishedAt, p.Title, p.URL, p.VideoCount)
	}
	fmt.Println()
	return nil
}

func runListPlaylistVideos(cmd *cobra.Command, ytClient *youtube.Client) error {
	playlistID, err := ytClient.ExtractPlaylistID(listPlaylist)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if listSearch != "" {
		ui.Status(fmt.Sprintf("Playlist: %s  |  filter: %q", playlistID, listSearch))
	} else {
		ui.Status(fmt.Sprintf("Playlist: %s", playlistID))
	}
	ui.Status(fmt.Sprintf("Retrieving up to %d videos...", listMax))

	listings, scanned, err := ytClient.ListPlaylistVideos(cmd.Context(), playlistID, listSearch, listMax)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if len(listings) == 0 {
		if listSearch != "" {
			ui.Complete(fmt.Sprintf("No videos found matching %q (scanned %d).", listSearch, scanned))
		} else {
			ui.Complete("No videos found in playlist.")
		}
		return nil
	}

	if listSearch != "" {
		ui.Complete(fmt.Sprintf("Found %d matching videos (scanned %d total).", len(listings), scanned))
	} else {
		ui.Complete(fmt.Sprintf("%d videos retrieved.", len(listings)))
	}

	ui.SectionHeader(fmt.Sprintf("Playlist — %s", playlistID))
	fmt.Println()
	for i, v := range listings {
		ui.ListingRow(i+1, v.PublishedAt, v.Title, v.URL)
	}
	fmt.Println()
	return nil
}

func init() {
	listCmd.Flags().StringVar(&listChannel, "channel", "", "Channel ID (UCxx...) or handle (@name)")
	listCmd.Flags().StringVar(&listPlaylist, "playlist", "", "Playlist ID or URL to list videos from")
	listCmd.Flags().StringVar(&listSearch, "search", "", "Filter by title/description keyword")
	listCmd.Flags().Int64Var(&listMax, "max", 50, "Maximum number of results to return")
	listCmd.Flags().BoolVar(&listPlaylists, "playlists", false, "List playlists instead of videos")
}

