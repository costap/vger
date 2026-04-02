package youtube

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"

	"github.com/costap/vger/internal/domain"
)

// Client implements domain.MetadataFetcher using the YouTube Data API v3.
type Client struct {
	APIKey string
}

func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// FetchMetadata retrieves title, description, channel, publish date, and duration
// for the video identified by the given URL. Supported URL formats:
//
//	https://www.youtube.com/watch?v=ID
//	https://youtu.be/ID
//	https://www.youtube.com/embed/ID
//	https://www.youtube.com/live/ID
func (c *Client) FetchMetadata(ctx context.Context, rawURL string) (*domain.VideoMetadata, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("url must not be empty")
	}

	videoID, err := extractVideoID(rawURL)
	if err != nil {
		return nil, fmt.Errorf("extract video id: %w", err)
	}

	svc, err := yt.NewService(ctx, option.WithAPIKey(c.APIKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
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

// extractVideoID parses a YouTube video ID from any common URL format.
func extractVideoID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	switch u.Host {
	case "youtu.be":
		id := strings.TrimPrefix(u.Path, "/")
		if id == "" {
			return "", fmt.Errorf("no video id in youtu.be url")
		}
		return id, nil

	case "www.youtube.com", "youtube.com", "m.youtube.com":
		// /watch?v=ID
		if v := u.Query().Get("v"); v != "" {
			return v, nil
		}
		// /embed/ID or /live/ID
		for _, prefix := range []string{"/embed/", "/live/"} {
			if strings.HasPrefix(u.Path, prefix) {
				id := strings.TrimPrefix(u.Path, prefix)
				id = strings.SplitN(id, "/", 2)[0]
				if id != "" {
					return id, nil
				}
			}
		}
		return "", fmt.Errorf("unrecognised youtube url path: %s", u.Path)
	}

	return "", fmt.Errorf("not a recognised youtube url: %s", rawURL)
}

// iso8601DurationRe matches the PT#H#M#S format returned by the YouTube API.
var iso8601DurationRe = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)

// parseISO8601Duration converts an ISO 8601 duration string (e.g. "PT1H15M30S") to seconds.
// Returns 0 if the string does not match the expected pattern.
func parseISO8601Duration(d string) int {
	m := iso8601DurationRe.FindStringSubmatch(d)
	if m == nil {
		return 0
	}
	toInt := func(s string) int {
		if s == "" {
			return 0
		}
		n, _ := strconv.Atoi(s)
		return n
	}
	return toInt(m[1])*3600 + toInt(m[2])*60 + toInt(m[3])
}
