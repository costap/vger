package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

// loadCacheEntries loads full CachedAnalysis entries for the given video IDs.
// Best-effort: returns an empty map on any error so listings still render.
func loadCacheEntries(videoIDs []string) map[string]*domain.CachedAnalysis {
	dir, err := cache.DefaultDir()
	if err != nil {
		return map[string]*domain.CachedAnalysis{}
	}
	c := cache.New(dir)
	entries, err := c.LoadByVideoIDs(context.Background(), videoIDs)
	if err != nil {
		return map[string]*domain.CachedAnalysis{}
	}
	m := make(map[string]*domain.CachedAnalysis, len(entries))
	for _, e := range entries {
		m[e.VideoID] = e
	}
	return m
}

var listChannel string
var listSearch string
var listTags string
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

Cached videos (★) show technology tags extracted by Gemini from prior scans.
Use --tags to filter the listing to only videos whose cached analysis contains
a matching technology (case-insensitive substring). Composable with --search.

Examples:
  vger list --channel @cncf
  vger list --channel @cncf --playlists
  vger list --channel @cncf --playlists --search kubecon
  vger list --channel @cncf --search argocon
  vger list --channel @cncf --tags ebpf
  vger list --channel @cncf --tags kubernetes --search 2024
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
		saveChannelHistory(listChannel, channelID, channelName)

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

	if err := ytClient.EnrichWithDetails(cmd.Context(), listings); err != nil {
		ui.Status(fmt.Sprintf("(metadata enrichment unavailable: %v)", err))
	}

	// Load cache entries for all retrieved videos (best-effort).
	ids := make([]string, len(listings))
	for i, v := range listings {
		ids[i] = v.VideoID
	}
	cacheEntries := loadCacheEntries(ids)

	// Apply --tags filter if set: keep only cached videos whose technology tags match.
	if listTags != "" {
		listings = filterByTag(listings, cacheEntries, listTags)
		if len(listings) == 0 {
			ui.Complete(fmt.Sprintf("No cached videos found matching tag %q.", listTags))
			return nil
		}
		ui.Status(fmt.Sprintf("Tag filter %q matched %d cached video(s).", listTags, len(listings)))
	}

	ui.SectionHeader(fmt.Sprintf("Videos — %s", channelName))
	fmt.Println()
	for i, v := range listings {
		ui.ListingRow(i+1, v, cacheEntries[v.VideoID])
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

	if err := ytClient.EnrichWithDetails(cmd.Context(), listings); err != nil {
		ui.Status(fmt.Sprintf("(metadata enrichment unavailable: %v)", err))
	}

	// Load cache entries for all retrieved videos (best-effort).
	ids := make([]string, len(listings))
	for i, v := range listings {
		ids[i] = v.VideoID
	}
	cacheEntries := loadCacheEntries(ids)

	// Apply --tags filter if set: keep only cached videos whose technology tags match.
	if listTags != "" {
		listings = filterByTag(listings, cacheEntries, listTags)
		if len(listings) == 0 {
			ui.Complete(fmt.Sprintf("No cached videos found matching tag %q.", listTags))
			return nil
		}
		ui.Status(fmt.Sprintf("Tag filter %q matched %d cached video(s).", listTags, len(listings)))
	}

	ui.SectionHeader(fmt.Sprintf("Playlist — %s", playlistID))
	fmt.Println()
	for i, v := range listings {
		ui.ListingRow(i+1, v, cacheEntries[v.VideoID])
	}
	fmt.Println()
	return nil
}

func init() {
	listCmd.Flags().StringVar(&listChannel, "channel", "", "Channel ID (UCxx...) or handle (@name)")
	listCmd.Flags().StringVar(&listPlaylist, "playlist", "", "Playlist ID or URL to list videos from")
	listCmd.Flags().StringVar(&listSearch, "search", "", "Filter by title/description keyword")
	listCmd.Flags().StringVar(&listTags, "tags", "", "Filter by technology tag from cached analyses (e.g. --tags ebpf)")
	listCmd.Flags().Int64Var(&listMax, "max", 50, "Maximum number of results to return")
	listCmd.Flags().BoolVar(&listPlaylists, "playlists", false, "List playlists instead of videos")

	_ = listCmd.RegisterFlagCompletionFunc("channel", channelCompletionFunc)
}

// filterByTag returns the subset of listings that are cached and have at least one
// technology tag containing tagQuery (case-insensitive substring match).
func filterByTag(listings []domain.VideoListing, entries map[string]*domain.CachedAnalysis, tagQuery string) []domain.VideoListing {
	q := strings.ToLower(tagQuery)
	var out []domain.VideoListing
	for _, v := range listings {
		entry, ok := entries[v.VideoID]
		if !ok || entry == nil {
			continue
		}
		for _, t := range entry.Tags() {
			if strings.Contains(strings.ToLower(t), q) {
				out = append(out, v)
				break
			}
		}
	}
	return out
}

