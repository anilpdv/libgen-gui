package download

import (
	"sync"
)

// Queue coordinates thread-safe FIFO ordering and deduplication for download tasks.
type Queue struct {
	mu    sync.RWMutex
	tasks []*Task
	store QueueStore
}

// NewQueue creates an initialized Queue.
func NewQueue(store QueueStore) *Queue {
	q := &Queue{
		tasks: make([]*Task, 0),
		store: store,
	}

	if store != nil {
		if loaded, err := store.Load(); err == nil && len(loaded) > 0 {
			q.tasks = loaded
		}
	}

	return q
}

// Enqueue adds new tasks to the queue if not already present.
func (q *Queue) Enqueue(tasks ...*Task) []*Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	var added []*Task
	for _, t := range tasks {
		if t == nil {
			continue
		}

		// Check for existing pending or active task by MD5 or ID
		var existing *Task
		for _, ex := range q.tasks {
			if (t.Md5 != "" && ex.Md5 == t.Md5) || (t.ID != "" && ex.ID == t.ID) {
				if ex.State() == StateQueued || ex.State() == StateDownloading || ex.State() == StatePaused {
					existing = ex
					break
				}
			}
		}

		if existing != nil {
			added = append(added, existing)
			continue
		}

		q.tasks = append(q.tasks, t)
		added = append(added, t)
	}

	if q.store != nil {
		_ = q.store.Save(q.tasks)
	}

	return added
}

// Get finds a task by ID or MD5.
func (q *Queue) Get(idOrMd5 string) *Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, t := range q.tasks {
		if t.ID == idOrMd5 || t.Md5 == idOrMd5 {
			return t
		}
	}
	return nil
}

// List returns an immutable copy of all tasks.
func (q *Queue) List() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	res := make([]*Task, len(q.tasks))
	copy(res, q.tasks)
	return res
}

// NextPending returns the first task in StateQueued.
func (q *Queue) NextPending() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, t := range q.tasks {
		if t.State() == StateQueued {
			return t
		}
	}
	return nil
}

// Remove dequeues a task if it is queued or failed.
func (q *Queue) Remove(idOrMd5 string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, t := range q.tasks {
		if t.ID == idOrMd5 || t.Md5 == idOrMd5 {
			if t.State() == StateQueued || t.State() == StatePaused || t.State() == StateFailed || t.State() == StateCancelled {
				q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
				if q.store != nil {
					_ = q.store.Save(q.tasks)
				}
				return true
			}
		}
	}
	return false
}

// ClearCompleted removes completed, failed, or cancelled tasks from queue history.
func (q *Queue) ClearCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	remaining := make([]*Task, 0, len(q.tasks))
	for _, t := range q.tasks {
		st := t.State()
		if st == StateQueued || st == StateDownloading || st == StatePaused {
			remaining = append(remaining, t)
		}
	}
	q.tasks = remaining

	if q.store != nil {
		_ = q.store.Save(q.tasks)
	}
}

// SaveState persists current queue items to disk.
func (q *Queue) SaveState() {
	q.mu.RLock()
	tasks := append([]*Task(nil), q.tasks...)
	q.mu.RUnlock()

	if q.store != nil {
		_ = q.store.Save(tasks)
	}
}
