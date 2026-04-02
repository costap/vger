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
