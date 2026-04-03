package youtube

import (
	"context"
	"fmt"

	"github.com/costap/vger/internal/domain"
)

// FetchMetadata retrieves title, description, channel, publish date, and duration
// for the video identified by the given URL. Supported URL formats:
//
// https://www.youtube.com/watch?v=ID
// https://youtu.be/ID
// https://www.youtube.com/embed/ID
// https://www.youtube.com/live/ID
func (c *Client) FetchMetadata(ctx context.Context, rawURL string) (*domain.VideoMetadata, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("url must not be empty")
	}

	videoID, err := extractVideoID(rawURL)
	if err != nil {
		return nil, fmt.Errorf("extract video id: %w", err)
	}

	svc, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := svc.Videos.
		List([]string{"snippet", "contentDetails"}).
		Id(videoID).
		Do()
	if err != nil {
		return nil, fmt.Errorf("youtube videos.list: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	item := resp.Items[0]
	return &domain.VideoMetadata{
		URL:         rawURL,
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		ChannelName: item.Snippet.ChannelTitle,
		PublishedAt: item.Snippet.PublishedAt,
		DurationSec: parseISO8601Duration(item.ContentDetails.Duration),
	}, nil
}
