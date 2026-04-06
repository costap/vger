package quota

import (
	"context"
	"sync"
	"time"
)

// tpmEntry is a single token-usage observation.
type tpmEntry struct {
	ts     time.Time
	tokens int64
}

// TPMWindow is a sliding 60-second window that tracks token usage for Gemini
// API calls. It is used to throttle dispatch when approaching the tokens-per-minute
// (TPM) quota limit. All methods are goroutine-safe.
type TPMWindow struct {
	mu      sync.Mutex
	entries []tpmEntry
}

// Add records that tokens were consumed right now.
func (w *TPMWindow) Add(tokens int64) {
	if tokens <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = append(w.entries, tpmEntry{ts: time.Now(), tokens: tokens})
	w.prune()
}

// TokensLastMinute returns the total tokens consumed in the last 60 seconds.
func (w *TPMWindow) TokensLastMinute() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.prune()
	var total int64
	for _, e := range w.entries {
		total += e.tokens
	}
	return total
}

// WaitForHeadroom blocks until adding estimated tokens to the current window
// would not exceed limit. Returns immediately if limit ≤ 0 (no throttle) or
// if there is already enough headroom. Respects context cancellation.
func (w *TPMWindow) WaitForHeadroom(ctx context.Context, estimated, limit int64) error {
	if limit <= 0 {
		return nil
	}
	for {
		sleep := w.headroomSleep(estimated, limit)
		if sleep <= 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
}

// headroomSleep returns the duration to sleep before estimated tokens can be
// added without exceeding limit, or ≤ 0 if there is already enough headroom.
// Must be called without holding the mutex.
func (w *TPMWindow) headroomSleep(estimated, limit int64) time.Duration {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.prune()

	var total int64
	for _, e := range w.entries {
		total += e.tokens
	}
	if total+estimated <= limit {
		return 0
	}

	// Find the oldest entry whose expiry frees enough space.
	cumFree := int64(0)
	for _, e := range w.entries {
		cumFree += e.tokens
		if total-cumFree+estimated <= limit {
			// Sleep until this entry ages out of the 60s window.
			expiresAt := e.ts.Add(time.Minute)
			remaining := time.Until(expiresAt) + 100*time.Millisecond
			if remaining < 100*time.Millisecond {
				return 100 * time.Millisecond
			}
			return remaining
		}
	}
	// All entries together are not enough — wait a full minute.
	return time.Minute
}

// prune removes entries older than 60 seconds. Must be called with the mutex held.
func (w *TPMWindow) prune() {
	cutoff := time.Now().Add(-time.Minute)
	i := 0
	for i < len(w.entries) && w.entries[i].ts.Before(cutoff) {
		i++
	}
	if i > 0 {
		w.entries = w.entries[i:]
	}
}
