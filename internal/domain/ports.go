package domain

import "context"

// MetadataFetcher retrieves metadata for a video from its source URL.
type MetadataFetcher interface {
	FetchMetadata(ctx context.Context, url string) (*VideoMetadata, error)
}

// VideoAnalyser submits a video URL to a multimodal model and returns structured analysis.
type VideoAnalyser interface {
	AnalyseVideo(ctx context.Context, url string, metadata *VideoMetadata) (*Report, error)
}

// ChannelLister retrieves a list of videos from a YouTube channel,
// optionally filtered by a search query, ordered by publish date descending.
type ChannelLister interface {
	// ResolveChannel accepts a channel ID (UCxx...) or a handle (@name / name)
	// and returns the canonical channel ID and display name.
	ResolveChannel(ctx context.Context, channelRef string) (id, name string, err error)

	// ListVideos returns up to maxResults videos from the given channel,
	// newest first. If query is non-empty it filters by title and description.
	ListVideos(ctx context.Context, channelID, query string, maxResults int64) ([]VideoListing, error)
}

// AnalysisCache persists and retrieves completed scan results keyed by video ID.
type AnalysisCache interface {
	// Save writes the cached analysis to the backing store.
	Save(ctx context.Context, entry *CachedAnalysis) error

	// Load retrieves a previously cached analysis by video ID.
	// Returns (nil, nil) if no entry exists for the given ID.
	Load(ctx context.Context, videoID string) (*CachedAnalysis, error)
}

// VideoQA answers follow-up questions about a video using a previously cached analysis
// as context. No video re-upload is performed.
type VideoQA interface {
	Ask(ctx context.Context, question string, cached *CachedAnalysis) (string, error)
}

