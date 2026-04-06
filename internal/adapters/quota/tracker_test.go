package quota

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTracker_IncrementAndToday(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quota.json")

	tr, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := tr.Today(); got != 0 {
		t.Errorf("Today() before increment = %d, want 0", got)
	}

	tr.Increment(1)
	tr.Increment(1)
	tr.Increment(3)

	if got := tr.Today(); got != 5 {
		t.Errorf("Today() = %d, want 5", got)
	}
}

func TestTracker_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quota.json")

	tr1, _ := New(path)
	tr1.Increment(7)

	// Reload from disk.
	tr2, err := New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := tr2.Today(); got != 7 {
		t.Errorf("after reload Today() = %d, want 7", got)
	}
}

func TestTracker_ResetOnDateChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quota.json")

	tr, _ := New(path)
	tr.Increment(10)

	// Manually write a stale date to the file.
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	data := `{"date":"` + yesterday + `","requests":10}`
	_ = os.WriteFile(path, []byte(data), 0600)

	tr2, _ := New(path)
	if got := tr2.Today(); got != 0 {
		t.Errorf("Today() after date change = %d, want 0 (should have reset)", got)
	}

	tr2.Increment(2)
	if got := tr2.Today(); got != 2 {
		t.Errorf("Today() after increment post-reset = %d, want 2", got)
	}
}

func TestTracker_ConcurrentIncrements(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quota.json")
	tr, _ := New(path)

	const goroutines = 20
	done := make(chan struct{}, goroutines)
	for range goroutines {
		go func() {
			tr.Increment(1)
			done <- struct{}{}
		}()
	}
	for range goroutines {
		<-done
	}
	if got := tr.Today(); got != goroutines {
		t.Errorf("Today() after concurrent increments = %d, want %d", got, goroutines)
	}
}
