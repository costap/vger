package youtube

import (
	"context"
	"fmt"
	"strings"

	"github.com/costap/vger/internal/domain"
)

// ResolveChannel accepts a channel ID (UCxx...) or handle (@name or name)
// and returns the canonical channel ID and display name.
func (c *Client) ResolveChannel(ctx context.Context, channelRef string) (string, string, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return "", "", err
	}

	call := svc.Channels.List([]string{"id", "snippet"})

	// Channel IDs start with "UC" and are 24 chars; everything else is a handle.
	if strings.HasPrefix(channelRef, "UC") && len(channelRef) == 24 {
		call = call.Id(channelRef)
	} else {
		handle := strings.TrimPrefix(channelRef, "@")
		call = call.ForHandle(handle)
	}

	resp, err := call.Do()
	if err != nil {
		return "", "", fmt.Errorf("youtube channels.list: %w", err)
	}
	if len(resp.Items) == 0 {
		return "", "", fmt.Errorf("channel not found: %s", channelRef)
	}

	ch := resp.Items[0]
	return ch.Id, ch.Snippet.Title, nil
}

// getUploadsPlaylistID returns the uploads playlist ID for a channel.
// The uploads playlist contains every video the channel has ever published,
// in reverse-chronological order.
func (c *Client) getUploadsPlaylistID(ctx context.Context, channelID string) (string, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return "", err
	}

	resp, err := svc.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil {
		return "", fmt.Errorf("youtube channels.list (contentDetails): %w", err)
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("channel not found: %s", channelID)
	}

	playlistID := resp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	if playlistID == "" {
		return "", fmt.Errorf("no uploads playlist found for channel: %s", channelID)
	}
	return playlistID, nil
}

// ListVideos walks the channel's uploads playlist and returns up to maxResults
// videos whose title or description contains query (case-insensitive).
// When query is empty all videos are returned up to maxResults.
//
// Unlike search.list, this approach is comprehensive — it will find every
// matching video regardless of age, popularity, or search index coverage.
// It implements domain.ChannelLister.
func (c *Client) ListVideos(ctx context.Context, channelID, query string, maxResults int64) ([]domain.VideoListing, error) {
	listings, _, err := c.listVideosWithScanCount(ctx, channelID, query, maxResults)
	return listings, err
}

// ListVideosDetailed is the same as ListVideos but also returns the total number
// of playlist items scanned. Used by the CLI to display progress context.
func (c *Client) ListVideosDetailed(ctx context.Context, channelID, query string, maxResults int64) (listings []domain.VideoListing, scanned int, err error) {
	return c.listVideosWithScanCount(ctx, channelID, query, maxResults)
}

// listVideosWithScanCount is the shared implementation for ListVideos and
// ListVideosDetailed.
func (c *Client) listVideosWithScanCount(ctx context.Context, channelID, query string, maxResults int64) ([]domain.VideoListing, int, error) {
	playlistID, err := c.getUploadsPlaylistID(ctx, channelID)
	if err != nil {
		return nil, 0, err
	}

	svc, err := c.newService(ctx)
	if err != nil {
		return nil, 0, err
	}

	queryLower := strings.ToLower(query)
	var listings []domain.VideoListing
	scanned := 0
	pageToken := ""

	for {
		call := svc.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50) // API max per page

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, scanned, fmt.Errorf("youtube playlistItems.list: %w", err)
		}

		for _, item := range resp.Items {
			scanned++
			sn := item.Snippet

			// Skip deleted/private videos (resourceId.videoId will be empty or title will be "Deleted video")
			if sn.ResourceId == nil || sn.ResourceId.VideoId == "" {
				continue
			}

			if queryLower != "" {
				titleMatch := strings.Contains(strings.ToLower(sn.Title), queryLower)
				descMatch := strings.Contains(strings.ToLower(sn.Description), queryLower)
				if !titleMatch && !descMatch {
					continue
				}
			}

			id := sn.ResourceId.VideoId
			listings = append(listings, domain.VideoListing{
				VideoID:      id,
				Title:        sn.Title,
				PublishedAt:  sn.PublishedAt,
				URL:          "https://www.youtube.com/watch?v=" + id,
				Description:  sn.Description,
				ChannelTitle: sn.VideoOwnerChannelTitle,
			})

			if int64(len(listings)) >= maxResults {
				return listings, scanned, nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return listings, scanned, nil
}

// ListPlaylists returns up to maxResults playlists from the given channel ordered
// by publication date descending. If query is non-empty, only playlists whose
// title or description contain the query (case-insensitive) are returned.
func (c *Client) ListPlaylists(ctx context.Context, channelID, query string, maxResults int64) ([]domain.PlaylistListing, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []domain.PlaylistListing
	pageToken := ""

	for {
		call := svc.Playlists.List([]string{"snippet", "contentDetails"}).
			ChannelId(channelID).
			MaxResults(50)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("youtube playlists.list: %w", err)
		}

		for _, item := range resp.Items {
			if queryLower != "" {
				titleMatch := strings.Contains(strings.ToLower(item.Snippet.Title), queryLower)
				descMatch := strings.Contains(strings.ToLower(item.Snippet.Description), queryLower)
				if !titleMatch && !descMatch {
					continue
				}
			}

			results = append(results, domain.PlaylistListing{
				PlaylistID:  item.Id,
				Title:       item.Snippet.Title,
				Description: item.Snippet.Description,
				PublishedAt: item.Snippet.PublishedAt,
				VideoCount:  item.ContentDetails.ItemCount,
				URL:         "https://www.youtube.com/playlist?list=" + item.Id,
			})

			if int64(len(results)) >= maxResults {
				return results, nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return results, nil
}

// ListPlaylistVideos returns up to maxResults videos from the given playlist ID,
// in playlist order (newest-first for most YouTube playlists). If query is
// non-empty, only videos whose title or description contain it are returned.
// Accepts a raw playlist ID or a full YouTube playlist URL.
func (c *Client) ListPlaylistVideos(ctx context.Context, playlistID, query string, maxResults int64) ([]domain.VideoListing, int, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return nil, 0, err
	}

	queryLower := strings.ToLower(query)
	var listings []domain.VideoListing
	scanned := 0
	pageToken := ""

	for {
		call := svc.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, scanned, fmt.Errorf("youtube playlistItems.list: %w", err)
		}

		for _, item := range resp.Items {
			scanned++
			sn := item.Snippet

			if sn.ResourceId == nil || sn.ResourceId.VideoId == "" {
				continue
			}

			if queryLower != "" {
				titleMatch := strings.Contains(strings.ToLower(sn.Title), queryLower)
				descMatch := strings.Contains(strings.ToLower(sn.Description), queryLower)
				if !titleMatch && !descMatch {
					continue
				}
			}

			id := sn.ResourceId.VideoId
			listings = append(listings, domain.VideoListing{
				VideoID:      id,
				Title:        sn.Title,
				PublishedAt:  sn.PublishedAt,
				URL:          "https://www.youtube.com/watch?v=" + id,
				Description:  sn.Description,
				ChannelTitle: sn.VideoOwnerChannelTitle,
			})

			if int64(len(listings)) >= maxResults {
				return listings, scanned, nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return listings, scanned, nil
}

// GetPlaylistTitle fetches the display title of a playlist by its ID.
// Returns the playlist ID itself if the title cannot be resolved.
// Quota cost: 1 unit.
func (c *Client) GetPlaylistTitle(ctx context.Context, playlistID string) (string, error) {
svc, err := c.newService(ctx)
if err != nil {
return playlistID, err
}

resp, err := svc.Playlists.List([]string{"snippet"}).
Id(playlistID).
MaxResults(1).
Do()
if err != nil {
return playlistID, fmt.Errorf("youtube playlists.list: %w", err)
}
if len(resp.Items) == 0 {
return playlistID, nil
}
return resp.Items[0].Snippet.Title, nil
}
