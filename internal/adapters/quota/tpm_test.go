package quota

import (
	"context"
	"testing"
	"time"
)

func TestTPMWindow_AddAndTokensLastMinute(t *testing.T) {
	w := &TPMWindow{}
	w.Add(100_000)
	w.Add(200_000)
	got := w.TokensLastMinute()
	if got != 300_000 {
		t.Fatalf("expected 300000, got %d", got)
	}
}

func TestTPMWindow_ExpiredEntriesExcluded(t *testing.T) {
	w := &TPMWindow{}
	// Inject an old entry directly (older than 60s).
	w.mu.Lock()
	w.entries = append(w.entries, tpmEntry{
		ts:     time.Now().Add(-61 * time.Second),
		tokens: 999_999,
	})
	w.mu.Unlock()

	w.Add(50_000)
	got := w.TokensLastMinute()
	if got != 50_000 {
		t.Fatalf("expected 50000 (old entry expired), got %d", got)
	}
}

func TestTPMWindow_WaitForHeadroom_ImmediateUnderLimit(t *testing.T) {
	w := &TPMWindow{}
	w.Add(200_000)
	ctx := context.Background()
	// limit=1M, already used 200K, estimating 500K → total 700K ≤ 1M → should not block
	err := w.WaitForHeadroom(ctx, 500_000, 1_000_000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTPMWindow_WaitForHeadroom_NoLimitSkips(t *testing.T) {
	w := &TPMWindow{}
	w.Add(10_000_000) // massive usage, but limit=0 means skip
	ctx := context.Background()
	err := w.WaitForHeadroom(ctx, 5_000_000, 0)
	if err != nil {
		t.Fatalf("expected no error with limit=0, got %v", err)
	}
}

func TestTPMWindow_WaitForHeadroom_ContextCancel(t *testing.T) {
	w := &TPMWindow{}
	// Fill the window to the brim.
	w.Add(950_000)
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately — WaitForHeadroom should return ctx.Err().
	cancel()
	err := w.WaitForHeadroom(ctx, 200_000, 1_000_000)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestTPMWindow_HeadroomSleep_ReturnsZeroWhenEnoughRoom(t *testing.T) {
	w := &TPMWindow{}
	w.Add(100_000)
	sleep := w.headroomSleep(200_000, 1_000_000)
	if sleep > 0 {
		t.Fatalf("expected no sleep needed, got %s", sleep)
	}
}
