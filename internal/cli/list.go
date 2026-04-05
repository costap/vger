package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

// sortListingsByDate sorts VideoListings by PublishedAt descending (newest first).
func sortListingsByDate(listings []domain.VideoListing) {
	sort.Slice(listings, func(i, j int) bool {
		return listings[i].PublishedAt > listings[j].PublishedAt
	})
}

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
var listCached bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List videos or playlists from a YouTube channel",
	Long: `List videos or playlists from a YouTube channel ordered by publish date, newest first.
The channel can be specified as a channel ID (UCxx...) or a handle (@cncf / cncf).

By default, videos are listed. Use --playlists to list playlists instead.
Use --playlist to list the videos inside a specific playlist.

Use --cached to browse all locally cached videos without fetching from YouTube.
This is the fastest way to search across everything you have ever scanned.

Video results are retrieved by walking the channel's complete uploads history and
filtering client-side — so results are comprehensive, not limited by YouTube's
search index. All matching videos will be found regardless of age or popularity.

Cached videos (★) show technology tags extracted by Gemini from prior scans.
Use --tags to filter the listing to only videos whose cached analysis contains
a matching technology or playlist name (case-insensitive substring).
Composable with --search.

Examples:
  vger list --channel @cncf
  vger list --channel @cncf --playlists
  vger list --channel @cncf --playlists --search kubecon
  vger list --channel @cncf --search argocon
  vger list --channel @cncf --tags ebpf
  vger list --channel @cncf --tags kubernetes --search 2024
  vger list --playlist "https://www.youtube.com/playlist?list=PLj6h78yzYM2P-3T82bF...a"
  vger list --playlist PLj6h78yzYM2P...a --search "service mesh"
  vger list --channel UCvqbFHwN-nwalWPjPUKpvTA --search "kubecon 2024" --max 100
  vger list --cached
  vger list --cached --tags ebpf
  vger list --cached --tags "kubecon" --tags "ebpf"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --cached mode: browse all locally cached analyses, no YouTube call needed.
		if listCached {
			return runListCachedVideos(cmd)
		}

		ytClient := youtube.New(youtubeAPIKey)

		// --playlist takes precedence — no channel required
		if listPlaylist != "" {
			return runListPlaylistVideos(cmd, ytClient)
		}

		if listChannel == "" {
			return fmt.Errorf("--channel, --playlist, or --cached is required")
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
	listCmd.Flags().StringVar(&listTags, "tags", "", "Filter by technology tag or playlist name (e.g. --tags ebpf, --tags kubecon)")
	listCmd.Flags().Int64Var(&listMax, "max", 50, "Maximum number of results to return")
	listCmd.Flags().BoolVar(&listPlaylists, "playlists", false, "List playlists instead of videos")
	listCmd.Flags().BoolVar(&listCached, "cached", false, "Browse all locally cached videos (no YouTube API call)")

	_ = listCmd.RegisterFlagCompletionFunc("channel", channelCompletionFunc)
}

// runListCachedVideos lists all videos that have a local cache entry.
// No YouTube API calls are made. Supports --search and --tags filtering.
func runListCachedVideos(cmd *cobra.Command) error {
	dir, err := cache.DefaultDir()
	if err != nil {
		ui.RedAlert(err)
		return err
	}
	c := cache.New(dir)

	ui.Status("Loading cached analyses...")

	index, err := c.LoadIndex()
	if err != nil {
		ui.RedAlert(err)
		return err
	}
	if len(index) == 0 {
		ui.Complete("No cached videos found. Run 'vger scan <url>' to analyse videos first.")
		return nil
	}

	ids := make([]string, 0, len(index))
	for id := range index {
		ids = append(ids, id)
	}

	entries, err := c.LoadByVideoIDs(cmd.Context(), ids)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	// Build VideoListings from cache entries.
	listings := make([]domain.VideoListing, 0, len(entries))
	entryMap := make(map[string]*domain.CachedAnalysis, len(entries))
	for _, e := range entries {
		listings = append(listings, domain.VideoListing{
			VideoID:      e.VideoID,
			Title:        e.Metadata.Title,
			PublishedAt:  e.Metadata.PublishedAt,
			URL:          e.Metadata.URL,
			ChannelTitle: e.Metadata.ChannelName,
			// Duration and ViewCount are available via VideoMetadata.DurationSec /
			// ViewCount but VideoListing stores ISO 8601 duration; skip for now.
		})
		entryMap[e.VideoID] = e
	}

	// Sort by published date descending (newest first).
	sortListingsByDate(listings)

	// Apply --search filter (title match).
	if listSearch != "" {
		q := strings.ToLower(listSearch)
		filtered := listings[:0]
		for _, v := range listings {
			if strings.Contains(strings.ToLower(v.Title), q) {
				filtered = append(filtered, v)
			}
		}
		listings = filtered
	}

	// Apply --tags filter (tech tags + playlist titles).
	if listTags != "" {
		listings = filterByTag(listings, entryMap, listTags)
	}

	if len(listings) == 0 {
		ui.Complete("No cached videos match the given filters.")
		return nil
	}

	ui.Complete(fmt.Sprintf("%d cached video(s).", len(listings)))
	ui.SectionHeader("Cached Videos")
	fmt.Println()
	for i, v := range listings {
		ui.ListingRow(i+1, v, entryMap[v.VideoID])
	}
	fmt.Println()
	return nil
}

// filterByTag returns the subset of listings that are cached and have at least one
// technology tag OR playlist title containing tagQuery (case-insensitive substring match).
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
				goto next
			}
		}
		for _, e := range entry.Events() {
			if strings.Contains(strings.ToLower(e), q) {
				out = append(out, v)
				goto next
			}
		}
		for _, s := range entry.Speakers() {
			if strings.Contains(strings.ToLower(s), q) {
				out = append(out, v)
				goto next
			}
		}
	next:
	}
	return out
}
