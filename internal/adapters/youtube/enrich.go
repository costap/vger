package youtube

import (
	"context"
	"fmt"

	"github.com/costap/vger/internal/domain"
)

// EnrichWithDetails fetches duration and view count for a slice of VideoListings
// using a batched videos.list API call (up to 50 IDs per request).
// Results are written in-place; the function is best-effort — a single batch
// failure returns an error but already-enriched entries retain their values.
//
// Quota cost: 1 unit per batch of ≤50 videos (compared with 0 for the listing).
func (c *Client) EnrichWithDetails(ctx context.Context, listings []domain.VideoListing) error {
	if len(listings) == 0 {
		return nil
	}

	// Build an index so we can locate each listing after the batch call.
	index := make(map[string]*domain.VideoListing, len(listings))
	for i := range listings {
		index[listings[i].VideoID] = &listings[i]
	}

	svc, err := c.newService(ctx)
	if err != nil {
		return err
	}

	// Batch in groups of 50 (YouTube API limit per request).
	ids := make([]string, 0, len(listings))
	for _, l := range listings {
		ids = append(ids, l.VideoID)
	}

	for start := 0; start < len(ids); start += 50 {
		end := start + 50
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[start:end]

		resp, err := svc.Videos.
			List([]string{"contentDetails", "statistics"}).
			Id(batch...).
			MaxResults(50).
			Do()
		if err != nil {
			return fmt.Errorf("youtube videos.list (enrich batch %d-%d): %w", start, end, err)
		}

		for _, item := range resp.Items {
			entry, ok := index[item.Id]
			if !ok {
				continue
			}
			if item.ContentDetails != nil {
				entry.Duration = item.ContentDetails.Duration
			}
			if item.Statistics != nil {
				entry.ViewCount = int64(item.Statistics.ViewCount)
			}
		}
	}

	return nil
}
