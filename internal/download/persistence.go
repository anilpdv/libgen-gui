package download

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"libgen-gui/internal/storage"
)

// PersistedTask holds JSON serialized state of a download item including its captured destination.
type PersistedTask struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Authors     []string         `json:"authors"`
	DownloadURL string           `json:"download_url"`
	PageURL     string           `json:"page_url"`
	Filename    string           `json:"filename"`
	Extension   string           `json:"extension"`
	Md5         string           `json:"md5"`
	Destination storage.Location `json:"destination"`
	State       State            `json:"state"`
	Downloaded  int64            `json:"downloaded"`
	Total       int64            `json:"total"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

// QueueStore defines persistence interface for download queues.
type QueueStore interface {
	Save(tasks []*Task) error
	Load() ([]*Task, error)
}

// JSONQueueStore persists queue state atomically to disk.
type JSONQueueStore struct {
	filePath      string
	defaultTarget storage.Location
	mu            sync.Mutex
}

// NewJSONQueueStore creates a new JSON queue store.
func NewJSONQueueStore(filePath string) *JSONQueueStore {
	return NewJSONQueueStoreWithTarget(filePath, storage.Location{})
}

// NewJSONQueueStoreWithTarget creates a new JSON queue store with a fallback target for legacy migration.
func NewJSONQueueStoreWithTarget(filePath string, defaultTarget storage.Location) *JSONQueueStore {
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			filePath = filepath.Join(os.TempDir(), "libgen_queue.json")
		} else {
			filePath = filepath.Join(home, ".libgen_downloader", "queue.json")
		}
	}
	return &JSONQueueStore{
		filePath:      filePath,
		defaultTarget: defaultTarget,
	}
}

// SetDefaultTarget updates the fallback target used for legacy task migration.
func (s *JSONQueueStore) SetDefaultTarget(target storage.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultTarget = target
}

// MigrateRecoveredTask normalizes an unmarshaled legacy task and ensures it has a valid destination.
func MigrateRecoveredTask(task *Task, defaultTarget storage.Location) error {
	if task.Destination.Kind == "" {
		task.Destination = defaultTarget
	}

	if task.Destination.Kind != "" {
		if err := task.Destination.Validate(); err != nil {
			return fmt.Errorf("task %s has invalid destination: %w", task.ID, err)
		}
	}

	if task.State() == StateDownloading {
		_ = task.Transition(StatePaused)
	}

	return nil
}

// Save atomically writes tasks to disk via a .tmp file and rename.
func (s *JSONQueueStore) Save(tasks []*Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	persisted := make([]PersistedTask, 0, len(tasks))
	for _, t := range tasks {
		if t == nil {
			continue
		}
		persisted = append(persisted, PersistedTask{
			ID:          t.ID,
			Title:       t.Title,
			Authors:     t.Authors,
			DownloadURL: t.DownloadURL,
			PageURL:     t.PageURL,
			Filename:    t.Filename,
			Extension:   t.Extension,
			Md5:         t.Md5,
			Destination: t.Destination,
			State:       t.State(),
			Downloaded:  t.Downloaded(),
			Total:       t.Total(),
			CreatedAt:   t.CreatedAt,
			CompletedAt: t.CompletedAt,
		})
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filePath)
}

// Load reads persisted tasks from disk and normalizes interrupted downloading states to paused.
func (s *JSONQueueStore) Load() ([]*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Task{}, nil
		}
		return nil, err
	}

	var persisted []PersistedTask
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, errors.New("corrupted queue file")
	}

	tasks := make([]*Task, 0, len(persisted))
	for _, p := range persisted {
		task := &Task{
			ID:          p.ID,
			Title:       p.Title,
			Authors:     p.Authors,
			DownloadURL: p.DownloadURL,
			PageURL:     p.PageURL,
			Filename:    p.Filename,
			Extension:   p.Extension,
			Md5:         p.Md5,
			Destination: p.Destination,
			CreatedAt:   p.CreatedAt,
			CompletedAt: p.CompletedAt,
			state:       p.State,
		}
		task.downloaded.Store(p.Downloaded)
		task.total.Store(p.Total)

		// State Recovery on Startup & Destination Migration
		if err := MigrateRecoveredTask(task, s.defaultTarget); err != nil {
			// If destination is invalid but defaultTarget is available, fallback
			if s.defaultTarget.Kind != "" {
				task.Destination = s.defaultTarget
			}
		}

		if task.state == StateDownloading {
			task.state = StatePaused
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
