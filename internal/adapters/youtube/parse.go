package youtube

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ExtractVideoID is the exported entry point used by callers outside this
// package (e.g. the CLI) to obtain a video ID from a YouTube URL.
func (c *Client) ExtractVideoID(rawURL string) (string, error) {
	return extractVideoID(rawURL)
}

// ExtractPlaylistID is the exported entry point for obtaining a playlist ID
// from a YouTube playlist URL or a raw playlist ID.
func (c *Client) ExtractPlaylistID(rawInput string) (string, error) {
	return extractPlaylistID(rawInput)
}

// extractVideoID parses a YouTube video ID from any common URL format:
//
//	https://www.youtube.com/watch?v=ID
//	https://youtu.be/ID
//	https://www.youtube.com/embed/ID
//	https://www.youtube.com/live/ID
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
		if v := u.Query().Get("v"); v != "" {
			return v, nil
		}
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

// extractPlaylistID parses a YouTube playlist ID from a URL or returns the
// input unchanged if it looks like a raw playlist ID (starts with "PL", "FL", etc.).
//
//	https://www.youtube.com/playlist?list=PLj6h78yzYM2P...
//	https://youtube.com/watch?v=ID&list=PLj6h78yzYM2P...
func extractPlaylistID(rawInput string) (string, error) {
	if rawInput == "" {
		return "", fmt.Errorf("playlist id or url must not be empty")
	}

	u, err := url.Parse(rawInput)
	if err != nil || u.Scheme == "" {
		// Not a URL — treat as a raw playlist ID.
		return rawInput, nil
	}

	if list := u.Query().Get("list"); list != "" {
		return list, nil
	}

	return "", fmt.Errorf("no playlist id found in url: %s", rawInput)
}
