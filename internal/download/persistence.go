package download

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistedTask holds JSON serialized state of a download item.
type PersistedTask struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Authors     []string  `json:"authors"`
	DownloadURL string    `json:"download_url"`
	PageURL     string    `json:"page_url"`
	Filename    string    `json:"filename"`
	Extension   string    `json:"extension"`
	Md5         string    `json:"md5"`
	State       State     `json:"state"`
	Downloaded  int64     `json:"downloaded"`
	Total       int64     `json:"total"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// QueueStore defines persistence interface for download queues.
type QueueStore interface {
	Save(tasks []*Task) error
	Load() ([]*Task, error)
}

// JSONQueueStore persists queue state atomically to disk.
type JSONQueueStore struct {
	filePath string
	mu       sync.Mutex
}

// NewJSONQueueStore creates a new JSON queue store.
func NewJSONQueueStore(filePath string) *JSONQueueStore {
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			filePath = filepath.Join(os.TempDir(), "libgen_queue.json")
		} else {
			filePath = filepath.Join(home, ".libgen_downloader", "queue.json")
		}
	}
	return &JSONQueueStore{
		filePath: filePath,
	}
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
			CreatedAt:   p.CreatedAt,
			CompletedAt: p.CompletedAt,
			state:       p.State,
		}
		task.downloaded.Store(p.Downloaded)
		task.total.Store(p.Total)

		// State Recovery on Startup:
		// If application crashed or stopped during an active download, normalize to Paused.
		if task.state == StateDownloading {
			task.state = StatePaused
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
