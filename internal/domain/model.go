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

// ── Track ─────────────────────────────────────────────────────────────────────

// Valid signal status values.
const (
	SignalStatusSpotted    = "spotted"
	SignalStatusEvaluating = "evaluating"
	SignalStatusAdopted    = "adopted"
	SignalStatusRejected   = "rejected"
	SignalStatusParked     = "parked"
)

// ValidSignalStatuses is the ordered list of allowed status values.
var ValidSignalStatuses = []string{
	SignalStatusSpotted,
	SignalStatusEvaluating,
	SignalStatusAdopted,
	SignalStatusRejected,
	SignalStatusParked,
}

// ValidSignalCategories is the ordered list of allowed category values.
var ValidSignalCategories = []string{
	"networking", "security", "platform", "data", "ai",
	"observability", "developer-experience", "process", "other",
}

// Signal is a technology or idea captured by the architect for later investigation.
type Signal struct {
	ID             string           `json:"id"`
	Title          string           `json:"title"`
	Date           string           `json:"date"`            // YYYY-MM-DD
	Source         string           `json:"source"`          // "Blog post", "Twitter/X", "Colleague", …
	URL            string           `json:"url"`
	Category       string           `json:"category"`
	Status         string           `json:"status"`
	Note           string           `json:"note"`            // why captured
	Tags           []string         `json:"tags,omitempty"`
	LinkedVideoIDs []string         `json:"linked_video_ids,omitempty"`
	Enrichment     *SignalEnrichment `json:"enrichment,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// SignalEnrichment is the AI-generated context added after initial capture.
type SignalEnrichment struct {
	EnrichedAt   time.Time `json:"enriched_at"`
	WhatItIs     string    `json:"what_it_is"`
	Maturity     string    `json:"maturity"`
	Alternatives []string  `json:"alternatives"`
	StackFit     string    `json:"stack_fit"`
	NextSteps    []string  `json:"next_steps"`
}

