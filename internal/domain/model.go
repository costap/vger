package domain

import "time"

// VideoMetadata holds data retrieved from the video source prior to analysis.
type VideoMetadata struct {
	URL         string
	Title       string
	Description string
	ChannelName string
	PublishedAt string
	DurationSec int
}

// VideoListing is a lightweight entry returned when browsing a channel's video catalogue.
type VideoListing struct {
	VideoID     string
	Title       string
	PublishedAt string
	URL         string
	Description string
}

// PlaylistListing is a lightweight entry returned when browsing a channel's playlists.
type PlaylistListing struct {
	PlaylistID  string
	Title       string
	Description string
	PublishedAt string
	VideoCount  int64
	URL         string
}

// Technology represents a technology or project identified in the video.
type Technology struct {
	Name        string
	Description string
	WhyRelevant string
	LearnMore   string
	CNCFStage   string // "graduated", "incubating", "sandbox", or empty
}

// Report is the final structured output produced by the agent for a single video.
type Report struct {
	VideoTitle       string
	VideoURL         string
	VideoDurationSec int
	Stardate         string
	Summary          string
	Notes            string       // detailed freeform narrative from Gemini covering everything mentioned
	Technologies     []Technology
}

// CachedAnalysis is the persisted form of a completed scan.
// It stores both the raw metadata and the final report so that follow-up
// questions can be answered without re-uploading the video.
type CachedAnalysis struct {
	VideoID  string        `json:"video_id"`
	CachedAt time.Time     `json:"cached_at"`
	Metadata VideoMetadata `json:"metadata"`
	Report   Report        `json:"report"`
}

// TechCount records how many talks in a playlist mentioned a given technology.
type TechCount struct {
	Name  string
	Count int
}

// PriorityTalk is a recommended talk from the AI synthesis layer.
type PriorityTalk struct {
	Title  string
	URL    string
	Reason string
}

// DigestReport is the output of the AI synthesis layer for a playlist.
type DigestReport struct {
	OverarchingTheme string
	LearningPath     []string       // technology names in recommended study order
	PriorityTalks    []PriorityTalk // top talks to watch first
	KeyInsights      string         // freeform narrative of the most important takeaways
}

