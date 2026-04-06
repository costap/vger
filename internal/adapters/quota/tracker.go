package quota

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// record is the on-disk representation of today's usage.
type record struct {
	Date     string `json:"date"`
	Requests int64  `json:"requests"`
}

// Tracker counts Gemini API requests dispatched by vger today.
// The count is persisted to a JSON file and resets automatically when the
// Pacific Time date changes, matching the Gemini API RPD reset schedule.
// All methods are goroutine-safe.
type Tracker struct {
	mu   sync.Mutex
	path string
	rec  record
}

// New creates a Tracker backed by the given file path.
// If the file does not exist it is created on first Increment call.
func New(path string) (*Tracker, error) {
	t := &Tracker{path: path}
	if err := t.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return t, nil
}

// DefaultPath returns the default tracker file path (~/.vger/quota.json).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".vger", "quota.json"), nil
}

// Increment adds n to today's request count and persists the new value.
// A date change resets the count to n before saving.
func (t *Tracker) Increment(n int64) {
	today := pacificDate()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rec.Date != today {
		t.rec = record{Date: today, Requests: n}
	} else {
		t.rec.Requests += n
	}
	_ = t.save() // best-effort; tracking failure must not break scans
}

// Today returns the number of API requests dispatched by vger today (Pacific Time).
func (t *Tracker) Today() int64 {
	today := pacificDate()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rec.Date != today {
		return 0
	}
	return t.rec.Requests
}

// load reads the tracker file. Returns os.ErrNotExist if the file is absent.
func (t *Tracker) load() error {
	data, err := os.ReadFile(t.path)
	if err != nil {
		return err
	}
	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return nil // corrupt file → treat as empty; will be overwritten on next save
	}
	if r.Date == pacificDate() {
		t.rec = r
	}
	return nil
}

// save writes the current record to disk. The parent directory is created if needed.
func (t *Tracker) save() error {
	if err := os.MkdirAll(filepath.Dir(t.path), 0750); err != nil {
		return err
	}
	data, err := json.Marshal(t.rec)
	if err != nil {
		return err
	}
	return os.WriteFile(t.path, data, 0600)
}

// pacificDate returns the current date in Pacific Time (America/Los_Angeles),
// matching the Gemini API RPD reset schedule. Falls back to UTC on error.
func pacificDate() string {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.Now().UTC().Format("2006-01-02")
	}
	return time.Now().In(loc).Format("2006-01-02")
}
