package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"libgen-gui/internal/network"
	"libgen-gui/internal/storage"
	"libgen-gui/pkg/libgen"
)

// DownloadRequest represents a client request to download a specific book.
type DownloadRequest struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	DownloadURL string   `json:"download_url"`
	PageURL     string   `json:"page_url"`
	Extension   string   `json:"extension"`
	Md5         string   `json:"md5"`
	Size        int64    `json:"size"`
}

// DownloadSnapshot represents an immutable point-in-time snapshot of a task for views.
type DownloadSnapshot struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Authors     []string         `json:"authors"`
	Filename    string           `json:"filename"`
	Extension   string           `json:"extension"`
	Md5         string           `json:"md5"`
	Destination storage.Location `json:"destination"`
	State       State            `json:"state"`
	Downloaded  int64            `json:"downloaded"`
	Total       int64            `json:"total"`
	Progress    float64          `json:"progress"`
	Speed       int64            `json:"speed"`
	Error       error            `json:"-"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

// DownloadService interface decouples the UI from backend download mechanics.
type DownloadService interface {
	Enqueue(req DownloadRequest) (string, error)
	Pause(id string) error
	Resume(id string) error
	Cancel(id string) error
	List() []DownloadSnapshot
	Subscribe(fn Listener)
}

// Dependencies contains dependencies for constructing a Manager.
type Dependencies struct {
	HTTPClient     *http.Client
	StorageFactory storage.Factory
	TargetProvider storage.TargetProvider
	MirrorManager  *network.MirrorManager
	QueueStore     QueueStore
}

// Manager orchestrates download execution, queueing, resumable ranges, .part files, and storage.
type Manager struct {
	httpClient     *http.Client
	storageFactory storage.Factory
	targetProvider storage.TargetProvider
	mirrorManager  *network.MirrorManager
	queue          *Queue

	mu          sync.RWMutex
	listeners   []Listener
	activeTasks map[string]*Task
	ctx         context.Context
	cancel      context.CancelFunc
	workerOnce  sync.Once
}

// NewManager creates an initialized DownloadManager with backwards-compatible fallbacks.
func NewManager(httpClient *http.Client, storageService storage.Storage, mirrorManager *network.MirrorManager, store QueueStore) *Manager {
	var factory storage.Factory
	var targetProvider storage.TargetProvider

	if storageService != nil {
		factory = &singleStorageFactory{storage: storageService}
		initialLoc := storage.Location{
			Kind:        storage.LocationDesktopPath,
			Path:        storageService.BaseLocation(),
			DisplayName: storageService.BaseLocation(),
		}
		targetProvider, _ = storage.NewMutableTargetProvider(initialLoc)
	} else {
		factory = storage.NewStorageFactory()
	}

	return NewManagerWithDependencies(Dependencies{
		HTTPClient:     httpClient,
		StorageFactory: factory,
		TargetProvider: targetProvider,
		MirrorManager:  mirrorManager,
		QueueStore:     store,
	})
}

// NewManagerWithDependencies creates a Manager using explicit dependency injection.
func NewManagerWithDependencies(deps Dependencies) *Manager {
	if deps.HTTPClient == nil {
		deps.HTTPClient = network.NewClient()
	}
	if deps.StorageFactory == nil {
		deps.StorageFactory = storage.NewStorageFactory()
	}
	if deps.TargetProvider == nil {
		defaultLoc, _ := storage.ResolveDefaultLocation(false)
		deps.TargetProvider, _ = storage.NewMutableTargetProvider(defaultLoc)
	}
	if deps.MirrorManager == nil {
		deps.MirrorManager = network.NewMirrorManager(deps.HTTPClient, network.DefaultDownloadMirrors)
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		httpClient:     deps.HTTPClient,
		storageFactory: deps.StorageFactory,
		targetProvider: deps.TargetProvider,
		mirrorManager:  deps.MirrorManager,
		queue:          NewQueue(deps.QueueStore),
		activeTasks:    make(map[string]*Task),
		ctx:            ctx,
		cancel:         cancel,
	}

	m.startWorkerLoop()
	return m
}

type singleStorageFactory struct {
	storage storage.Storage
}

func (s *singleStorageFactory) For(loc storage.Location) (storage.Storage, error) {
	return s.storage, nil
}

// SetStorage updates the active storage backend (backwards-compatibility helper).
func (m *Manager) SetStorage(s storage.Storage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storageFactory = &singleStorageFactory{storage: s}
	if m.targetProvider != nil {
		_ = m.targetProvider.Set(storage.Location{
			Kind:        storage.LocationDesktopPath,
			Path:        s.BaseLocation(),
			DisplayName: s.BaseLocation(),
		})
	}
}

// SetTargetProvider sets the TargetProvider for resolving destination locations.
func (m *Manager) SetTargetProvider(p storage.TargetProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.targetProvider = p
}

// SetStorageFactory sets the Storage Factory.
func (m *Manager) SetStorageFactory(f storage.Factory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storageFactory = f
}

// SetHTTPClient updates the http client (e.g. for testing).
func (m *Manager) SetHTTPClient(client *http.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpClient = client
}

// Subscribe registers a progress/state event listener.
func (m *Manager) Subscribe(fn Listener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

func (m *Manager) publish(event DownloadEvent) {
	m.mu.RLock()
	listeners := append([]Listener(nil), m.listeners...)
	m.mu.RUnlock()

	for _, listener := range listeners {
		listener(event)
	}
}

// Enqueue captures the destination AT ENQUEUE TIME and adds the task to the queue.
func (m *Manager) Enqueue(req DownloadRequest) (string, error) {
	var destination storage.Location
	if m.targetProvider != nil {
		var err error
		destination, err = m.targetProvider.Current()
		if err != nil {
			return "", fmt.Errorf("resolve download location: %w", err)
		}
	}

	task := NewTaskWithDestination(
		req.ID,
		req.Title,
		req.Authors,
		req.DownloadURL,
		req.PageURL,
		req.Md5,
		req.Extension,
		req.Size,
		destination,
	)

	added := m.queue.Enqueue(task)
	if len(added) == 0 {
		return "", errors.New("failed to enqueue download")
	}

	enqueued := added[0]
	m.publish(DownloadEvent{
		TaskID:     enqueued.ID,
		Title:      enqueued.Title,
		Filename:   enqueued.Filename,
		State:      enqueued.State(),
		Downloaded: enqueued.Downloaded(),
		Total:      enqueued.Total(),
		Progress:   enqueued.Progress(),
	})

	return enqueued.ID, nil
}

// Pause pauses an active or queued download.
func (m *Manager) Pause(id string) error {
	task := m.queue.Get(id)
	if task == nil {
		return errors.New("task not found")
	}

	m.mu.Lock()
	if active, ok := m.activeTasks[id]; ok {
		active.Cancel()
		delete(m.activeTasks, id)
	}
	m.mu.Unlock()

	if err := task.Transition(StatePaused); err != nil {
		return err
	}

	m.queue.SaveState()
	m.publish(DownloadEvent{
		TaskID:     task.ID,
		Title:      task.Title,
		Filename:   task.Filename,
		State:      StatePaused,
		Downloaded: task.Downloaded(),
		Total:      task.Total(),
		Progress:   task.Progress(),
	})
	return nil
}

// Resume re-enqueues a paused or failed task.
func (m *Manager) Resume(id string) error {
	task := m.queue.Get(id)
	if task == nil {
		return errors.New("task not found")
	}

	if err := task.Transition(StateQueued); err != nil {
		return err
	}

	m.queue.SaveState()
	m.publish(DownloadEvent{
		TaskID:     task.ID,
		Title:      task.Title,
		Filename:   task.Filename,
		State:      StateQueued,
		Downloaded: task.Downloaded(),
		Total:      task.Total(),
		Progress:   task.Progress(),
	})
	return nil
}

// Cancel cancels a task and removes any partial downloaded .part file.
func (m *Manager) Cancel(id string) error {
	task := m.queue.Get(id)
	if task == nil {
		return errors.New("task not found")
	}

	m.mu.Lock()
	if active, ok := m.activeTasks[id]; ok {
		active.Cancel()
		delete(m.activeTasks, id)
	}
	factory := m.storageFactory
	m.mu.Unlock()

	task.Cancel()
	m.queue.SaveState()

	// Clean up temporary .part file from task's specific storage destination
	if factory != nil && task.Destination.Kind != "" {
		if store, err := factory.For(task.Destination); err == nil {
			_ = store.Delete(context.Background(), task.PartFilename())
		}
	}

	m.publish(DownloadEvent{
		TaskID:     task.ID,
		Title:      task.Title,
		Filename:   task.Filename,
		State:      StateCancelled,
		Downloaded: task.Downloaded(),
		Total:      task.Total(),
		Progress:   task.Progress(),
	})
	return nil
}

// List returns snapshots of all tasks in the queue.
func (m *Manager) List() []DownloadSnapshot {
	tasks := m.queue.List()
	snapshots := make([]DownloadSnapshot, len(tasks))
	for i, t := range tasks {
		snapshots[i] = t.Snapshot()
		snapshots[i].Destination = t.Destination
	}
	return snapshots
}

// Task returns a task by ID (for inspection/testing).
func (m *Manager) Task(id string) *Task {
	return m.queue.Get(id)
}

func (m *Manager) startWorkerLoop() {
	m.workerOnce.Do(func() {
		go func() {
			for {
				if m.ctx.Err() != nil {
					return
				}

				task := m.queue.NextPending()
				if task == nil {
					time.Sleep(200 * time.Millisecond)
					continue
				}

				m.executeDownload(task)
			}
		}()
	})
}

func (m *Manager) executeDownload(task *Task) {
	if err := task.Transition(StateDownloading); err != nil {
		return
	}

	taskCtx, taskCancel := task.SetContext(m.ctx)
	defer taskCancel()

	m.mu.Lock()
	m.activeTasks[task.ID] = task
	factory := m.storageFactory
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.activeTasks, task.ID)
		m.mu.Unlock()
		m.queue.SaveState()
	}()

	m.publish(DownloadEvent{
		TaskID:     task.ID,
		Title:      task.Title,
		Filename:   task.Filename,
		State:      StateDownloading,
		Downloaded: task.Downloaded(),
		Total:      task.Total(),
		Progress:   task.Progress(),
	})

	// Resolve storage backend for the task's captured destination
	storageService, err := factory.For(task.Destination)
	if err != nil {
		task.SetError(err)
		_ = task.Transition(StateFailed)
		m.publish(DownloadEvent{
			TaskID:     task.ID,
			Title:      task.Title,
			Filename:   task.Filename,
			State:      StateFailed,
			Downloaded: task.Downloaded(),
			Total:      task.Total(),
			Progress:   task.Progress(),
			Error:      err,
		})
		return
	}

	err = m.downloadTaskWithRetry(taskCtx, task, storageService)
	if err != nil {
		if taskCtx.Err() != nil || errors.Is(err, context.Canceled) {
			_ = task.Transition(StateCancelled)
			_ = storageService.Delete(context.Background(), task.PartFilename())
			m.publish(DownloadEvent{
				TaskID:     task.ID,
				Title:      task.Title,
				Filename:   task.Filename,
				State:      StateCancelled,
				Downloaded: task.Downloaded(),
				Total:      task.Total(),
				Progress:   task.Progress(),
			})
			return
		}

		task.SetError(err)
		_ = task.Transition(StateFailed)
		slog.Error("download failed", "task_id", task.ID, "title", task.Title, "error", err)
		m.publish(DownloadEvent{
			TaskID:     task.ID,
			Title:      task.Title,
			Filename:   task.Filename,
			State:      StateFailed,
			Downloaded: task.Downloaded(),
			Total:      task.Total(),
			Progress:   task.Progress(),
			Error:      err,
		})
		return
	}

	// Commit .part file to final filename atomically in task's destination
	commitErr := storageService.CommitPart(context.Background(), task.PartFilename(), task.Filename)
	if commitErr != nil {
		task.SetError(commitErr)
		_ = task.Transition(StateFailed)
		m.publish(DownloadEvent{
			TaskID:     task.ID,
			Title:      task.Title,
			Filename:   task.Filename,
			State:      StateFailed,
			Downloaded: task.Downloaded(),
			Total:      task.Total(),
			Progress:   task.Progress(),
			Error:      commitErr,
		})
		return
	}

	_ = task.Transition(StateCompleted)
	task.SetDownloaded(task.Total())
	m.publish(DownloadEvent{
		TaskID:     task.ID,
		Title:      task.Title,
		Filename:   task.Filename,
		State:      StateCompleted,
		Downloaded: task.Total(),
		Total:      task.Total(),
		Progress:   1.0,
	})
}

func (m *Manager) downloadTaskWithRetry(ctx context.Context, task *Task, store storage.Storage) error {
	// Resolve download URL if missing
	if task.DownloadURL == "" {
		book := &libgen.Book{
			Title:     task.Title,
			Extension: task.Extension,
			Md5:       task.Md5,
			PageURL:   task.PageURL,
		}
		if err := libgen.GetDownloadURL(book, false); err != nil || book.DownloadURL == "" {
			return fmt.Errorf("unable to resolve download mirror link: %w", err)
		}
		task.DownloadURL = book.DownloadURL
		task.PageURL = book.PageURL
	}

	return network.Retry(ctx, 3, func(attemptCtx context.Context, attempt int) error {
		return m.streamTask(attemptCtx, task, store)
	})
}

func (m *Manager) streamTask(ctx context.Context, task *Task, store storage.Storage) error {
	partName := task.PartFilename()

	// 1. Check existing partial download size on storage
	var existingBytes int64
	if exists, _ := store.Exists(ctx, partName); exists {
		existingBytes, _ = store.Size(ctx, partName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.DownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", network.DefaultUserAgent)
	req.Header.Set("Accept", "*/*")
	if task.PageURL != "" {
		req.Header.Set("Referer", task.PageURL)
	}

	// Safe Resumable Range Request
	if existingBytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingBytes))
	}

	m.mu.RLock()
	client := m.httpClient
	m.mu.RUnlock()

	resp, err := client.Do(req)
	if err != nil {
		return network.ClassifyError(err, task.DownloadURL, 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return network.ClassifyError(errors.New(resp.Status), task.DownloadURL, resp.StatusCode)
	}

	var writer io.WriteCloser
	var currentDownloaded int64

	switch resp.StatusCode {
	case http.StatusPartialContent:
		// Server supports Range resume! Append to existing .part file
		w, curSize, openErr := store.OpenAppend(ctx, partName)
		if openErr != nil {
			return openErr
		}
		writer = w
		currentDownloaded = curSize
		if task.Total() <= 0 && resp.ContentLength > 0 {
			task.SetTotal(curSize + resp.ContentLength)
		}
	case http.StatusOK:
		// Server ignored Range header or fresh download; start from beginning cleanly
		w, createErr := store.Create(ctx, partName)
		if createErr != nil {
			return createErr
		}
		writer = w
		currentDownloaded = 0
		if resp.ContentLength > 0 {
			task.SetTotal(resp.ContentLength)
		}
	default:
		return fmt.Errorf("unexpected HTTP status %d from mirror", resp.StatusCode)
	}
	defer writer.Close()

	task.SetDownloaded(currentDownloaded)

	// Throttled UI progress streaming (every 150ms)
	lastNotifyTime := time.Now()
	var bytesSinceLastNotify int64
	var speedSampleTime = time.Now()
	var bytesSample atomic.Int64

	buf := make([]byte, 32*1024)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			newTotal := task.AddDownloaded(int64(n))
			bytesSinceLastNotify += int64(n)
			bytesSample.Add(int64(n))

			now := time.Now()
			if now.Sub(lastNotifyTime) >= 150*time.Millisecond {
				elapsedSec := now.Sub(speedSampleTime).Seconds()
				if elapsedSec >= 0.5 {
					sampled := bytesSample.Swap(0)
					bps := int64(float64(sampled) / elapsedSec)
					task.SetSpeed(bps)
					speedSampleTime = now
				}

				m.publish(DownloadEvent{
					TaskID:     task.ID,
					Title:      task.Title,
					Filename:   task.Filename,
					State:      StateDownloading,
					Downloaded: newTotal,
					Total:      task.Total(),
					Progress:   task.Progress(),
					Speed:      task.Speed(),
				})
				lastNotifyTime = now
				bytesSinceLastNotify = 0
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
	}

	return nil
}
