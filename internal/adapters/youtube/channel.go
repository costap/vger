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

// ListVideos returns up to maxResults videos from the given channel,
// ordered by publish date descending. If query is non-empty, results are
// filtered by the search term against title and description.
func (c *Client) ListVideos(ctx context.Context, channelID, query string, maxResults int64) ([]domain.VideoListing, error) {
	svc, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}

	call := svc.Search.List([]string{"id", "snippet"}).
		ChannelId(channelID).
		Type("video").
		Order("date").
		MaxResults(maxResults)

	if query != "" {
		call = call.Q(query)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("youtube search.list: %w", err)
	}

	listings := make([]domain.VideoListing, 0, len(resp.Items))
	for _, item := range resp.Items {
		id := item.Id.VideoId
		listings = append(listings, domain.VideoListing{
			VideoID:     id,
			Title:       item.Snippet.Title,
			PublishedAt: item.Snippet.PublishedAt,
			URL:         "https://www.youtube.com/watch?v=" + id,
			Description: item.Snippet.Description,
		})
	}
	return listings, nil
}
