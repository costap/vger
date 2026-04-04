package signals

import (
	"context"
	"strings"

	"github.com/costap/vger/internal/domain"
)

// Search returns all signals whose Title, Note, Category, or enrichment context
// contains the query string (case-insensitive). Both JSONStore and MarkdownStore
// expose this method for use by the research command.

// Search filters the JSONStore's signals by relevance to query.
func (s *JSONStore) Search(ctx context.Context, query string) ([]*domain.Signal, error) {
	all, err := s.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	return filterSignals(all, query), nil
}

// Search filters the MarkdownStore's signals by relevance to query.
func (s *MarkdownStore) Search(ctx context.Context, query string) ([]*domain.Signal, error) {
	all, err := s.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	return filterSignals(all, query), nil
}

func filterSignals(signals []*domain.Signal, query string) []*domain.Signal {
	q := strings.ToLower(query)
	var out []*domain.Signal
	for _, sig := range signals {
		if matchesSignal(sig, q) {
			out = append(out, sig)
		}
	}
	return out
}

func matchesSignal(sig *domain.Signal, q string) bool {
	if strings.Contains(strings.ToLower(sig.Title), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sig.Note), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sig.Category), q) {
		return true
	}
	if sig.Enrichment != nil {
		if strings.Contains(strings.ToLower(sig.Enrichment.WhatItIs), q) {
			return true
		}
	}
	return false
}
