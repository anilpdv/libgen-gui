package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"libgen-gui/internal/network"
	"libgen-gui/internal/storage"
)

func TestStateTransitions(t *testing.T) {
	task := NewTask("1", "Title", []string{"Author"}, "", "", "md5_1", "pdf", 1000)

	if task.State() != StateQueued {
		t.Errorf("expected initial state Queued, got %s", task.State())
	}

	// Valid transition: Queued -> Downloading
	if err := task.Transition(StateDownloading); err != nil {
		t.Fatalf("expected valid transition Queued -> Downloading, got %v", err)
	}

	// Valid transition: Downloading -> Paused
	if err := task.Transition(StatePaused); err != nil {
		t.Fatalf("expected valid transition Downloading -> Paused, got %v", err)
	}

	// Valid transition: Paused -> Downloading
	if err := task.Transition(StateDownloading); err != nil {
		t.Fatalf("expected valid transition Paused -> Downloading, got %v", err)
	}

	// Valid transition: Downloading -> Completed
	if err := task.Transition(StateCompleted); err != nil {
		t.Fatalf("expected valid transition Downloading -> Completed, got %v", err)
	}

	// Illegal transition: Completed -> Downloading
	if err := task.Transition(StateDownloading); err == nil {
		t.Errorf("expected error for illegal transition Completed -> Downloading, got nil")
	}
}

func TestQueue_FIFOAndDeduplication(t *testing.T) {
	q := NewQueue(nil)

	t1 := NewTask("1", "Book 1", nil, "", "", "md5_1", "epub", 1000)
	t2 := NewTask("2", "Book 2", nil, "", "", "md5_2", "pdf", 2000)

	q.Enqueue(t1)
	q.Enqueue(t2)

	// Try enqueuing t1 duplicate
	q.Enqueue(t1)

	list := q.List()
	if len(list) != 2 {
		t.Errorf("expected 2 tasks in queue after deduplication, got %d", len(list))
	}

	next := q.NextPending()
	if next.ID != "1" {
		t.Errorf("expected FIFO next task to be 1, got %s", next.ID)
	}
}

func TestPersistence_AtomicSaveAndStartupRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "queue.json")
	store := NewJSONQueueStore(storeFile)

	t1 := NewTask("1", "Interrupted Book", nil, "", "", "md5_int", "pdf", 5000)
	_ = t1.Transition(StateDownloading) // Simulating app crash while downloading

	t2 := NewTask("2", "Queued Book", nil, "", "", "md5_q", "pdf", 3000)

	if err := store.Save([]*Task{t1, t2}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(storeFile); err != nil {
		t.Fatalf("expected queue.json to exist: %v", err)
	}

	// Recover on startup
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 loaded tasks, got %d", len(loaded))
	}

	// Verify StateDownloading was recovered and normalized to StatePaused
	if loaded[0].State() != StatePaused {
		t.Errorf("expected crashed downloading task to recover as StatePaused, got %s", loaded[0].State())
	}
	if loaded[1].State() != StateQueued {
		t.Errorf("expected queued task to remain StateQueued, got %s", loaded[1].State())
	}
}

func TestRecoveredLegacyTaskGetsDefaultDestination(t *testing.T) {
	task := &Task{
		ID:    "legacy-task",
		state: StateQueued,
	}

	defaultLocation := storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        "/downloads/default",
		DisplayName: "Default",
	}

	if err := MigrateRecoveredTask(task, defaultLocation); err != nil {
		t.Fatal(err)
	}

	if task.Destination.Path != defaultLocation.Path {
		t.Fatalf("legacy task did not receive default destination: got %q, expected %q", task.Destination.Path, defaultLocation.Path)
	}
}

func TestQueuedTaskRetainsCapturedDestination(t *testing.T) {
	first := storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        "/downloads/first",
		DisplayName: "First",
	}

	second := storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        "/downloads/second",
		DisplayName: "Second",
	}

	provider, err := storage.NewMutableTargetProvider(first)
	if err != nil {
		t.Fatal(err)
	}

	factory := storage.NewStorageFactory()
	mgr := NewManagerWithDependencies(Dependencies{
		StorageFactory: factory,
		TargetProvider: provider,
	})

	firstID, err := mgr.Enqueue(DownloadRequest{
		ID:          "first_task",
		Title:       "First",
		Extension:   "pdf",
		DownloadURL: "https://example.invalid/first",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := provider.Set(second); err != nil {
		t.Fatal(err)
	}

	secondID, err := mgr.Enqueue(DownloadRequest{
		ID:          "second_task",
		Title:       "Second",
		Extension:   "pdf",
		DownloadURL: "https://example.invalid/second",
	})
	if err != nil {
		t.Fatal(err)
	}

	firstTask := mgr.Task(firstID)
	secondTask := mgr.Task(secondID)

	if firstTask.Destination.Path != first.Path {
		t.Fatalf("first task destination changed unexpectedly: got %s, expected %s", firstTask.Destination.Path, first.Path)
	}

	if secondTask.Destination.Path != second.Path {
		t.Fatalf("second task did not use new destination: got %s, expected %s", secondTask.Destination.Path, second.Path)
	}
}

func TestManager_ResumableDownloadAndPartCommit(t *testing.T) {
	payload := strings.Repeat("ChunkData12345678", 500) // 8000 bytes
	var rangeRequests atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHdr := r.Header.Get("Range")
		if rangeHdr != "" {
			rangeRequests.Add(1)
			// Parse "bytes=4000-"
			var start int
			_, _ = fmt.Sscanf(rangeHdr, "bytes=%d-", &start)
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(payload)-1, len(payload)))
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)-start))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, payload[start:])
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, payload)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	fsStore, _ := storage.NewFileSystemStorage(tmpDir)
	client := network.NewClient()
	mirrorMgr := network.NewMirrorManager(client, []string{ts.URL})
	queueStore := NewJSONQueueStore(filepath.Join(tmpDir, "queue.json"))

	mgr := NewManager(client, fsStore, mirrorMgr, queueStore)

	var mu sync.Mutex
	var events []DownloadEvent
	mgr.Subscribe(func(ev DownloadEvent) {
		mu.Lock()
		events = append(events, ev)
		mu.Unlock()
	})

	// Pre-create .part file with 4000 bytes to simulate partial download resume
	partFilename := "Go Programming.pdf.part"
	w, err := fsStore.Create(context.Background(), partFilename)
	if err != nil {
		t.Fatalf("failed creating part file: %v", err)
	}
	_, _ = io.WriteString(w, payload[:4000])
	_ = w.Close()

	// Enqueue task matching the filename
	taskID, err := mgr.Enqueue(DownloadRequest{
		ID:          "task_resume",
		Title:       "Go Programming",
		DownloadURL: ts.URL,
		Extension:   "pdf",
		Md5:         "md5_res",
		Size:        int64(len(payload)),
	})
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	// Wait for download completion
	deadline := time.Now().Add(5 * time.Second)
	for {
		list := mgr.List()
		if len(list) > 0 && list[0].State == StateCompleted {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for download to complete")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Verify range request was used
	if rangeRequests.Load() != 1 {
		t.Errorf("expected 1 Range HTTP request, got %d", rangeRequests.Load())
	}

	// Verify .part file was committed to final filename
	finalFilename := "Go Programming.pdf"
	exists, _ := fsStore.Exists(context.Background(), finalFilename)
	if !exists {
		t.Errorf("expected final file %q to exist after download completion", finalFilename)
	}

	partExists, _ := fsStore.Exists(context.Background(), partFilename)
	if partExists {
		t.Errorf("expected temporary .part file to be removed after commit")
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, finalFilename))
	if string(data) != payload {
		t.Errorf("downloaded content mismatch")
	}

	if taskID == "" {
		t.Errorf("expected non-empty task ID")
	}
}
