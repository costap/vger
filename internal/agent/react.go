package agent

import (
	"context"
	"fmt"

	"github.com/costap/vger/internal/domain"
)

// ScanAgent orchestrates the analysis pipeline for a single video URL.
// It wires together a MetadataFetcher and a VideoAnalyser, sequencing calls in
// the order: fetch metadata → analyse video → return report.
//
// The outer pipeline is intentionally fixed in Go: both steps are always required
// in the same order, and keeping them deterministic allows cache checks before any
// LLM cost is incurred. The VideoAnalyser implementation (gemini.Client) runs its
// own internal function-calling loop for non-deterministic enrichment steps such as
// CNCF landscape lookups and URL validation.
type ScanAgent struct {
	fetcher  domain.MetadataFetcher
	analyser domain.VideoAnalyser
}

// New constructs a ScanAgent from the provided port implementations.
func New(fetcher domain.MetadataFetcher, analyser domain.VideoAnalyser) *ScanAgent {
	return &ScanAgent{
		fetcher:  fetcher,
		analyser: analyser,
	}
}

// Run executes the full analysis pipeline for the given URL and returns a Report
// and the VideoMetadata that was fetched during analysis.
func (a *ScanAgent) Run(ctx context.Context, url string) (*domain.Report, *domain.VideoMetadata, error) {
	meta, err := a.fetcher.FetchMetadata(ctx, url)
	if err != nil {
		return nil, nil, fmt.Errorf("metadata fetch: %w", err)
	}

	report, err := a.analyser.AnalyseVideo(ctx, url, meta)
	if err != nil {
		return nil, nil, fmt.Errorf("video analysis: %w", err)
	}

	return report, meta, nil
}
