package download

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"libgen-gui/internal/storage"
)

// Task represents an individual download work unit.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Authors     []string  `json:"authors"`
	DownloadURL string    `json:"download_url"`
	PageURL     string    `json:"page_url"`
	Filename    string    `json:"filename"`
	Extension   string    `json:"extension"`
	Md5         string    `json:"md5"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	mu         sync.RWMutex
	state      State
	downloaded atomic.Int64
	total      atomic.Int64
	speed      atomic.Int64 // bytes per second
	lastErr    error

	ctx    context.Context
	cancel context.CancelFunc
}

// NewTask initializes a new Task with a sanitized filename and initial queued state.
func NewTask(id, title string, authors []string, downloadURL, pageURL, md5, ext string, totalBytes int64) *Task {
	if ext == "" {
		ext = "pdf"
	}
	cleanTitle := storage.SanitizeFilename(title)
	filename := cleanTitle
	if len(authors) > 0 {
		cleanAuthor := storage.SanitizeFilename(authors[0])
		filename = fmt.Sprintf("%s by %s.%s", cleanTitle, cleanAuthor, strings.ToLower(ext))
	} else {
		filename = fmt.Sprintf("%s.%s", cleanTitle, strings.ToLower(ext))
	}
	filename = storage.SanitizeFilename(filename)

	t := &Task{
		ID:          id,
		Title:       title,
		Authors:     authors,
		DownloadURL: downloadURL,
		PageURL:     pageURL,
		Filename:    filename,
		Extension:   strings.ToLower(ext),
		Md5:         strings.ToLower(md5),
		CreatedAt:   time.Now(),
		state:       StateQueued,
	}
	t.total.Store(totalBytes)
	return t
}

// PartFilename returns the temporary .part file name on disk while downloading.
func (t *Task) PartFilename() string {
	return t.Filename + ".part"
}

// State returns current task lifecycle state thread-safely.
func (t *Task) State() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// Transition performs state transition validation under lock.
func (t *Task) Transition(next State) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := ValidateTransition(t.state, next); err != nil {
		return err
	}

	t.state = next
	if next == StateCompleted || next == StateFailed || next == StateCancelled {
		t.CompletedAt = time.Now()
	} else if next == StateDownloading && t.StartedAt.IsZero() {
		t.StartedAt = time.Now()
	}

	return nil
}

// SetContext sets active download cancellation context.
func (t *Task) SetContext(parent context.Context) (context.Context, context.CancelFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if parent == nil {
		parent = context.Background()
	}

	t.ctx, t.cancel = context.WithCancel(parent)
	return t.ctx, t.cancel
}

// Cancel cancels active context and transitions task to StateCancelled.
func (t *Task) Cancel() {
	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.state = StateCancelled
	t.CompletedAt = time.Now()
	t.mu.Unlock()
}

// Downloaded returns current downloaded bytes.
func (t *Task) Downloaded() int64 {
	return t.downloaded.Load()
}

// SetDownloaded updates downloaded bytes.
func (t *Task) SetDownloaded(bytes int64) {
	t.downloaded.Store(bytes)
}

// AddDownloaded adds delta bytes atomically.
func (t *Task) AddDownloaded(delta int64) int64 {
	return t.downloaded.Add(delta)
}

// Total returns total expected bytes.
func (t *Task) Total() int64 {
	return t.total.Load()
}

// SetTotal updates total expected bytes.
func (t *Task) SetTotal(total int64) {
	t.total.Store(total)
}

// Progress returns normalized progress float between 0.0 and 1.0.
func (t *Task) Progress() float64 {
	tot := t.total.Load()
	if tot <= 0 {
		return 0.0
	}
	dl := t.downloaded.Load()
	p := float64(dl) / float64(tot)
	if p > 1.0 {
		return 1.0
	}
	if p < 0.0 {
		return 0.0
	}
	return p
}

// Speed returns download speed in bytes per second.
func (t *Task) Speed() int64 {
	return t.speed.Load()
}

// SetSpeed updates current download speed.
func (t *Task) SetSpeed(bps int64) {
	t.speed.Store(bps)
}

// Error returns the last recorded error.
func (t *Task) Error() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastErr
}

// SetError records an error.
func (t *Task) SetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastErr = err
}

// Snapshot returns an immutable point-in-time snapshot under read lock.
func (t *Task) Snapshot() DownloadSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return DownloadSnapshot{
		ID:          t.ID,
		Title:       t.Title,
		Authors:     t.Authors,
		Filename:    t.Filename,
		Extension:   t.Extension,
		Md5:         t.Md5,
		State:       t.state,
		Downloaded:  t.downloaded.Load(),
		Total:       t.total.Load(),
		Progress:    t.Progress(),
		Speed:       t.speed.Load(),
		Error:       t.lastErr,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
	}
}
