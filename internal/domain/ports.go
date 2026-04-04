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

// ChannelLister retrieves a list of videos or playlists from a YouTube channel,
// optionally filtered by a search query.
type ChannelLister interface {
	// ResolveChannel accepts a channel ID (UCxx...) or a handle (@name / name)
	// and returns the canonical channel ID and display name.
	ResolveChannel(ctx context.Context, channelRef string) (id, name string, err error)

	// ListVideos returns up to maxResults videos from the given channel,
	// newest first. If query is non-empty it filters by title and description.
	ListVideos(ctx context.Context, channelID, query string, maxResults int64) ([]VideoListing, error)

	// ListPlaylists returns up to maxResults playlists from the given channel,
	// newest first. If query is non-empty it filters by title and description.
	ListPlaylists(ctx context.Context, channelID, query string, maxResults int64) ([]PlaylistListing, error)
}

// AnalysisCache persists and retrieves completed scan results keyed by video ID.
type AnalysisCache interface {
	// Save writes the cached analysis to the backing store.
	Save(ctx context.Context, entry *CachedAnalysis) error

	// Load retrieves a previously cached analysis by video ID.
	// Returns (nil, nil) if no entry exists for the given ID.
	Load(ctx context.Context, videoID string) (*CachedAnalysis, error)
}

// CacheSearcher performs relevance-scored full-text search over cached analyses.
type CacheSearcher interface {
	// Search returns up to maxResults cached analyses ranked by relevance to query.
	// Relevance is scored by matching against technology names, summary, and notes.
	Search(ctx context.Context, query string, maxResults int) ([]*CachedAnalysis, error)
}

// SignalSearcher performs a keyword search over the signal store.
type SignalSearcher interface {
	// Search returns all signals whose title, note, category, or enrichment context
	// contains the query string (case-insensitive).
	Search(ctx context.Context, query string) ([]*Signal, error)
}

// VideoQA answers follow-up questions about a video using a previously cached analysis
// as context. No video re-upload is performed.
type VideoQA interface {
	Ask(ctx context.Context, question string, cached *CachedAnalysis) (string, error)
}

// SignalEnricher uses an AI model to enrich or parse tech signals.
type SignalEnricher interface {
	// EnrichSignal generates AI context for an existing signal — WhatItIs,
	// Maturity, Alternatives, StackFit, and NextSteps.
	EnrichSignal(ctx context.Context, sig *Signal) (*SignalEnrichment, error)

	// ParseSignalFromPrompt extracts a Signal from a free-text description.
	// The caller is responsible for assigning ID, Status, and timestamps.
	ParseSignalFromPrompt(ctx context.Context, prompt string) (*Signal, error)
}

type SignalStore interface {
	// Save writes a signal to the backing store. Creates or overwrites by ID.
	Save(ctx context.Context, signal *Signal) error

	// Load retrieves a signal by ID. Returns (nil, nil) if not found.
	Load(ctx context.Context, id string) (*Signal, error)

	// LoadAll returns all signals, sorted by ID ascending.
	LoadAll(ctx context.Context) ([]*Signal, error)

	// LoadByStatus returns all signals matching the given status.
	LoadByStatus(ctx context.Context, status string) ([]*Signal, error)

	// LoadByCategory returns all signals matching the given category.
	LoadByCategory(ctx context.Context, category string) ([]*Signal, error)

	// NextID returns the next available zero-padded 4-digit ID string (e.g. "0042").
	NextID(ctx context.Context) (string, error)

	// Delete removes a signal by ID. Returns (nil) if not found (idempotent).
	Delete(ctx context.Context, id string) error
}

