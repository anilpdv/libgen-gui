package ui

import (
	"context"
	"sync"
	"time"

	"libgen-gui/pkg/libgen"
)

// QueueItemStatus represents the lifecycle state of an item in the download queue.
type QueueItemStatus int

const (
	QueueStatusPending QueueItemStatus = iota
	QueueStatusDownloading
	QueueStatusCompleted
	QueueStatusFailed
	QueueStatusCancelled
)

func (s QueueItemStatus) String() string {
	switch s {
	case QueueStatusPending:
		return "Pending"
	case QueueStatusDownloading:
		return "Downloading"
	case QueueStatusCompleted:
		return "Completed"
	case QueueStatusFailed:
		return "Failed"
	case QueueStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// DownloadItem represents a book in the download queue with its current status and metrics.
type DownloadItem struct {
	Book            *libgen.Book
	Status          QueueItemStatus
	Progress        float64 // 0.0 to 1.0
	BytesDownloaded int64
	TotalBytes      int64
	Speed           float64 // bytes per second
	Error           error
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time

	cancelFunc context.CancelFunc
}

// QueueSummary contains statistics when a queue batch completes.
type QueueSummary struct {
	Total      int
	Successful int
	Failed     int
	Cancelled  int
	TotalBytes int64
}

// DownloadQueue manages a thread-safe FIFO download queue with controlled concurrency.
type DownloadQueue struct {
	mu           sync.RWMutex
	items        []*DownloadItem
	activeItem   *DownloadItem
	isRunning    bool
	queueCtx     context.Context
	queueCancel  context.CancelFunc
	activeCancel context.CancelFunc

	// Callbacks
	OnQueueChanged  func()
	OnItemStarted   func(item *DownloadItem)
	OnItemProgress  func(item *DownloadItem)
	OnItemCompleted func(item *DownloadItem)
	OnItemFailed    func(item *DownloadItem, err error)
	OnQueueFinished func(summary QueueSummary)
}

// NewDownloadQueue creates a new initialized DownloadQueue.
func NewDownloadQueue() *DownloadQueue {
	return &DownloadQueue{
		items: make([]*DownloadItem, 0),
	}
}

// Enqueue adds one or more books to the end of the queue.
// If a book is already pending or downloading, it returns the existing item to prevent duplicates.
func (q *DownloadQueue) Enqueue(books ...*libgen.Book) []*DownloadItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	var added []*DownloadItem
	now := time.Now()

	for _, b := range books {
		if b == nil {
			continue
		}
		key := BookKey(b)

		// Check if already in queue and active or pending
		var existing *DownloadItem
		for _, it := range q.items {
			if BookKey(it.Book) == key && (it.Status == QueueStatusPending || it.Status == QueueStatusDownloading) {
				existing = it
				break
			}
		}

		if existing != nil {
			added = append(added, existing)
			continue
		}

		item := &DownloadItem{
			Book:       b,
			Status:     QueueStatusPending,
			TotalBytes: parseBookFilesize(b.Filesize),
			CreatedAt:  now,
		}
		q.items = append(q.items, item)
		added = append(added, item)
	}

	if len(added) > 0 && q.OnQueueChanged != nil {
		go q.OnQueueChanged()
	}

	return added
}

// Remove dequeues an item that is still in QueueStatusPending.
func (q *DownloadQueue) Remove(bookKey string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, it := range q.items {
		if BookKey(it.Book) == bookKey && it.Status == QueueStatusPending {
			q.items = append(q.items[:i], q.items[i+1:]...)
			if q.OnQueueChanged != nil {
				go q.OnQueueChanged()
			}
			return true
		}
	}
	return false
}

// CancelCurrent aborts the currently downloading book and advances to the next pending item.
func (q *DownloadQueue) CancelCurrent() {
	q.mu.Lock()
	if q.activeCancel != nil {
		q.activeCancel()
	}
	q.mu.Unlock()
}

// CancelAll aborts the current download and marks all pending items as cancelled.
func (q *DownloadQueue) CancelAll() {
	q.mu.Lock()
	if q.queueCancel != nil {
		q.queueCancel()
	}
	if q.activeCancel != nil {
		q.activeCancel()
	}
	for _, it := range q.items {
		if it.Status == QueueStatusPending {
			it.Status = QueueStatusCancelled
			it.CompletedAt = time.Now()
		}
	}
	q.mu.Unlock()

	if q.OnQueueChanged != nil {
		go q.OnQueueChanged()
	}
}

// ClearCompleted removes completed, failed, and cancelled items from the queue history.
func (q *DownloadQueue) ClearCompleted() {
	q.mu.Lock()
	defer q.mu.Unlock()

	remaining := make([]*DownloadItem, 0, len(q.items))
	for _, it := range q.items {
		if it.Status == QueueStatusPending || it.Status == QueueStatusDownloading {
			remaining = append(remaining, it)
		}
	}
	q.items = remaining

	if q.OnQueueChanged != nil {
		go q.OnQueueChanged()
	}
}

// GetItems returns a copy of all items currently in the queue.
func (q *DownloadQueue) GetItems() []*DownloadItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	res := make([]*DownloadItem, len(q.items))
	copy(res, q.items)
	return res
}

// GetActiveItem returns the currently downloading item, or nil if idle.
func (q *DownloadQueue) GetActiveItem() *DownloadItem {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.activeItem
}

// GetPendingCount returns the number of items waiting in queue (excluding the active one).
func (q *DownloadQueue) GetPendingCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	count := 0
	for _, it := range q.items {
		if it.Status == QueueStatusPending {
			count++
		}
	}
	return count
}

// GetTotalRemainingCount returns the active item (if any) plus all pending items.
func (q *DownloadQueue) GetTotalRemainingCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	count := 0
	for _, it := range q.items {
		if it.Status == QueueStatusPending || it.Status == QueueStatusDownloading {
			count++
		}
	}
	return count
}

// IsRunning returns true if a worker is currently processing items.
func (q *DownloadQueue) IsRunning() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.isRunning
}

// GetBookStatus returns the status of a specific book in the queue, or QueueStatusPending if not found.
func (q *DownloadQueue) GetBookStatus(b *libgen.Book) (QueueItemStatus, bool) {
	if b == nil {
		return QueueStatusPending, false
	}
	key := BookKey(b)

	q.mu.RLock()
	defer q.mu.RUnlock()

	// Check backwards so latest state is returned
	for i := len(q.items) - 1; i >= 0; i-- {
		if BookKey(q.items[i].Book) == key {
			return q.items[i].Status, true
		}
	}
	return QueueStatusPending, false
}

// StartWorker begins processing the queue if not already running.
// downloadFn is the function responsible for downloading a single book item.
func (q *DownloadQueue) StartWorker(downloadFn func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead int64, totalBytes int64, speed float64)) error) {
	q.mu.Lock()
	if q.isRunning {
		q.mu.Unlock()
		return
	}
	q.isRunning = true
	q.queueCtx, q.queueCancel = context.WithCancel(context.Background())
	q.mu.Unlock()

	go q.workerLoop(downloadFn)
}

func (q *DownloadQueue) workerLoop(downloadFn func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead int64, totalBytes int64, speed float64)) error) {
	defer func() {
		q.mu.Lock()
		q.isRunning = false
		q.activeItem = nil
		q.activeCancel = nil
		q.mu.Unlock()

		if q.OnQueueChanged != nil {
			go q.OnQueueChanged()
		}
	}()

	for {
		q.mu.Lock()
		if q.queueCtx.Err() != nil {
			q.mu.Unlock()
			break
		}

		// Find the next pending item
		var nextItem *DownloadItem
		for _, it := range q.items {
			if it.Status == QueueStatusPending {
				nextItem = it
				break
			}
		}

		if nextItem == nil {
			// No more pending items; compute summary and exit loop
			summary := q.computeSummaryUnderLock()
			q.mu.Unlock()
			if q.OnQueueFinished != nil {
				q.OnQueueFinished(summary)
			}
			return
		}

		// Prepare active item with its own cancel context child of queueCtx
		actCtx, actCancel := context.WithCancel(q.queueCtx)
		nextItem.Status = QueueStatusDownloading
		nextItem.StartedAt = time.Now()
		nextItem.cancelFunc = actCancel
		q.activeItem = nextItem
		q.activeCancel = actCancel
		q.mu.Unlock()

		if q.OnQueueChanged != nil {
			go q.OnQueueChanged()
		}
		if q.OnItemStarted != nil {
			q.OnItemStarted(nextItem)
		}

		// Execute the download
		err := downloadFn(actCtx, nextItem, func(bytesRead int64, totalBytes int64, speed float64) {
			q.mu.Lock()
			nextItem.BytesDownloaded = bytesRead
			if totalBytes > 0 {
				nextItem.TotalBytes = totalBytes
				nextItem.Progress = float64(bytesRead) / float64(totalBytes)
				if nextItem.Progress > 1.0 {
					nextItem.Progress = 1.0
				}
			}
			nextItem.Speed = speed
			q.mu.Unlock()

			if q.OnItemProgress != nil {
				q.OnItemProgress(nextItem)
			}
		})

		q.mu.Lock()
		nextItem.CompletedAt = time.Now()
		q.activeItem = nil
		q.activeCancel = nil

		if err != nil {
			if actCtx.Err() != nil || q.queueCtx.Err() != nil {
				nextItem.Status = QueueStatusCancelled
				if q.OnQueueChanged != nil {
					go q.OnQueueChanged()
				}
			} else {
				nextItem.Status = QueueStatusFailed
				nextItem.Error = err
				if q.OnItemFailed != nil {
					go q.OnItemFailed(nextItem, err)
				}
				if q.OnQueueChanged != nil {
					go q.OnQueueChanged()
				}
			}
		} else {
			nextItem.Status = QueueStatusCompleted
			nextItem.Progress = 1.0
			if q.OnItemCompleted != nil {
				go q.OnItemCompleted(nextItem)
			}
			if q.OnQueueChanged != nil {
				go q.OnQueueChanged()
			}
		}
		q.mu.Unlock()

		// Brief throttle between queue items to prevent mirror connection flooding
		time.Sleep(100 * time.Millisecond)
	}

	q.mu.Lock()
	summary := q.computeSummaryUnderLock()
	q.mu.Unlock()
	if q.OnQueueFinished != nil {
		q.OnQueueFinished(summary)
	}
}

func (q *DownloadQueue) computeSummaryUnderLock() QueueSummary {
	var s QueueSummary
	s.Total = len(q.items)
	for _, it := range q.items {
		switch it.Status {
		case QueueStatusCompleted:
			s.Successful++
			s.TotalBytes += it.BytesDownloaded
		case QueueStatusFailed:
			s.Failed++
		case QueueStatusCancelled:
			s.Cancelled++
		}
	}
	return s
}
