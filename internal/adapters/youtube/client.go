package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

// Client implements domain.MetadataFetcher and domain.ChannelLister
// using the YouTube Data API v3.
type Client struct {
	APIKey string
}

// New constructs a Client with the given API key.
func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// newService creates an authenticated YouTube API service for a single request.
func (c *Client) newService(ctx context.Context) (*yt.Service, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.APIKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}
	return svc, nil
}
