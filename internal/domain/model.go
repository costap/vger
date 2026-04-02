package domain

// VideoMetadata holds data retrieved from the video source prior to analysis.
type VideoMetadata struct {
	URL         string
	Title       string
	Description string
	ChannelName string
	PublishedAt string
	DurationSec int
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
	VideoTitle   string
	VideoURL     string
	Stardate     string
	Summary      string
	Technologies []Technology
}
