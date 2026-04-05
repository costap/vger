package youtube

import (
	"context"
	"fmt"

	"github.com/costap/vger/internal/domain"
)

// SearchVideos searches all of YouTube for videos matching query using the
// search.list API. Unlike ListVideos (which walks a channel's uploads playlist),
// this searches across every channel on YouTube.
//
// Note: search.list costs 100 quota units per API call (vs 1 for playlistItems.list).
// Use sparingly and prefer ListVideos when you only need results from a known channel.
func (c *Client) SearchVideos(ctx context.Context, query string, maxResults int64) ([]domain.VideoListing, error) {
	if maxResults <= 0 {
		maxResults = 20
	}
	// API cap per page is 50; we issue a single page since research discovery
	// uses at most 20 results and we filter out cached videos afterwards.
	perPage := maxResults
	if perPage > 50 {
		perPage = 50
	}

	svc, err := c.newService(ctx)
	if err != nil {
		return nil, err
	}

	var listings []domain.VideoListing
	pageToken := ""

	for int64(len(listings)) < maxResults {
		call := svc.Search.List([]string{"snippet"}).
			Q(query).
			Type("video").
			MaxResults(perPage)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("youtube search.list: %w", err)
		}

		for _, item := range resp.Items {
			if item.Id == nil || item.Id.VideoId == "" {
				continue
			}
			id := item.Id.VideoId
			sn := item.Snippet
			listings = append(listings, domain.VideoListing{
				VideoID:     id,
				Title:       sn.Title,
				PublishedAt: sn.PublishedAt,
				URL:         "https://www.youtube.com/watch?v=" + id,
				Description: sn.Description,
			})
			if int64(len(listings)) >= maxResults {
				return listings, nil
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return listings, nil
}
