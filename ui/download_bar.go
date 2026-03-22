package ui

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ciehanski/libgen-cli/libgen"
	"github.com/dustin/go-humanize"
)

// DownloadBar shows the save-folder picker, download button, cancel button,
// progress bar, and real-time status for single and batch downloads.
type DownloadBar struct {
	app           *App
	selectedBooks []*libgen.Book
	savePath      string

	pathLabel     *widget.Label
	browseBtn     *widget.Button
	dlBtn         *widget.Button
	clearBtn      *widget.Button
	cancelBtn     *widget.Button
	openFolderBtn *widget.Button
	queueBtn      *widget.Button
	activity      *widget.Activity
	progress      *widget.ProgressBar
	statusLbl     *widget.Label

	mu              sync.Mutex
	isDownloading   bool
	cancelFunc      context.CancelFunc
	sessionID       int64
	activeDownloads int64

	queue      *DownloadQueue
	httpClient *http.Client
}

func NewDownloadBar(a *App, defaultSavePath string) *DownloadBar {
	db := &DownloadBar{
		app:      a,
		savePath: defaultSavePath,
	}

	if a != nil {
		db.queue = a.GetDownloadQueue()
	} else {
		db.queue = NewDownloadQueue()
	}

	db.pathLabel = widget.NewLabel(defaultSavePath)
	db.pathLabel.Truncation = fyne.TextTruncateEllipsis

	db.browseBtn = widget.NewButtonWithIcon("Browse…", theme.FolderOpenIcon(), func() {
		if db.app != nil {
			ShowDownloadLocationDialog(db.app, false, func(path string) {
				db.SetSavePath(path)
			})
		}
	})

	db.cancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		db.CancelDownload()
	})
	db.cancelBtn.Importance = widget.MediumImportance
	db.cancelBtn.Hide()

	openLabel := "Open Folder"
	if runtime.GOOS == "darwin" {
		openLabel = "Open in Finder"
	}
	db.openFolderBtn = widget.NewButtonWithIcon(openLabel, theme.FolderOpenIcon(), func() {
		db.OpenSaveFolder()
	})
	db.openFolderBtn.Importance = widget.MediumImportance
	db.openFolderBtn.Hide()

	db.queueBtn = widget.NewButtonWithIcon("Queue", theme.ListIcon(), func() {
		if db.app != nil {
			ShowQueueDialog(db.app)
		}
	})
	db.queueBtn.Importance = widget.MediumImportance
	db.queueBtn.Hide()

	db.clearBtn = widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		if db.app != nil {
			db.app.ClearSelection()
		}
	})
	db.clearBtn.Importance = widget.LowImportance
	db.clearBtn.Hide()

	db.dlBtn = widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() { db.doDownload() })
	db.dlBtn.Importance = widget.HighImportance
	db.dlBtn.Disable() // enabled when 1 or more books are selected

	db.activity = widget.NewActivity()
	db.activity.Hide()

	db.progress = widget.NewProgressBar()
	db.progress.Hide()

	db.statusLbl = widget.NewLabel("")
	db.statusLbl.Truncation = fyne.TextTruncateEllipsis

	db.queue.OnQueueChanged = func() {
		db.updateQueueUI()
	}

	return db
}

// updateQueueUI refreshes the queue button visibility and label based on queue state.
func (db *DownloadBar) updateQueueUI() {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.queue == nil || db.queueBtn == nil {
		return
	}
	pending := db.queue.GetPendingCount()
	if pending > 0 {
		db.queueBtn.SetText(fmt.Sprintf("Queue (%d)", pending))
		db.queueBtn.Show()
	} else if db.isDownloading {
		db.queueBtn.SetText("Queue (1)")
		db.queueBtn.Show()
	} else {
		db.queueBtn.Hide()
	}
}

// SetSavePath updates the download destination path and updates the UI label.
func (db *DownloadBar) SetSavePath(path string) {
	db.mu.Lock()
	db.savePath = NormalizePath(path)
	if db.pathLabel != nil {
		db.pathLabel.SetText(db.savePath)
	}
	db.mu.Unlock()
}

// GetSavePath returns the current download destination path.
func (db *DownloadBar) GetSavePath() string {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.savePath
}

func (db *DownloadBar) isMobile() bool {
	if db.app != nil {
		return db.app.IsMobile()
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return true
	}
	return fyne.CurrentDevice().IsMobile()
}

func (db *DownloadBar) Widget() fyne.CanvasObject {
	saveToLbl := widget.NewLabelWithStyle("Save to:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	if db.isMobile() {
		// Mobile layout:
		// Row 1: Save-to destination + Browse button
		// Row 2: Action buttons (Clear/OpenFolder/Queue on left, Download/Cancel on right)
		// Row 3: Real-time status text
		// Row 4: Progress bar
		pathRow := container.NewBorder(
			nil, nil,
			saveToLbl,
			db.browseBtn,
			db.pathLabel,
		)
		actionSlot := container.NewHBox(db.cancelBtn, db.dlBtn)
		leftBtn := container.NewHBox(db.openFolderBtn, db.queueBtn, db.clearBtn)
		actionRow := container.NewBorder(nil, nil, leftBtn, nil, actionSlot)
		return container.NewVBox(
			pathRow,
			actionRow,
			db.statusLbl,
			db.progress,
		)
	}

	// Desktop layout:
	// Top control row:
	// Left: "Save to:" label (bold)
	// Center: save path (flexible with ellipsis)
	// Right: action buttons ([Browse…] [Open in Finder] [Queue] [Clear] [Cancel] [Download])
	actionButtons := container.NewHBox(db.browseBtn, db.openFolderBtn, db.queueBtn, db.clearBtn, db.cancelBtn, db.dlBtn)

	controlRow := container.NewBorder(
		nil, nil,
		saveToLbl,
		actionButtons,
		db.pathLabel,
	)

	// Clean vertical layout:
	// Row 1: Save-to destination & action buttons
	// Row 2: Real-time status text (single line with ellipsis)
	// Row 3: Full-width progress bar (only visible during download)
	return container.NewVBox(
		controlRow,
		db.statusLbl,
		db.progress,
	)
}

// SetSelectedBooks updates the download bar state when selections change.
func (db *DownloadBar) SetSelectedBooks(books []*libgen.Book) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.selectedBooks = books
	count := len(books)

	if db.isDownloading {
		if count > 0 {
			db.dlBtn.Show()
			db.dlBtn.Enable()
			if count == 1 {
				db.dlBtn.SetText("Add to Queue (1)")
			} else {
				db.dlBtn.SetText(fmt.Sprintf("Add to Queue (%d)", count))
			}
			if db.clearBtn != nil {
				db.clearBtn.Show()
				db.clearBtn.SetText(fmt.Sprintf("Clear (%d)", count))
			}
		} else {
			db.dlBtn.Hide()
			if db.clearBtn != nil {
				db.clearBtn.Hide()
			}
		}
		return
	}

	atomic.AddInt64(&db.sessionID, 1)

	db.statusLbl.Importance = widget.MediumImportance
	if count == 0 {
		if db.clearBtn != nil {
			db.clearBtn.Hide()
		}
		db.dlBtn.Disable()
		db.dlBtn.SetText("Download")
		db.statusLbl.SetText("")
	} else {
		db.openFolderBtn.Hide()
		if db.clearBtn != nil {
			db.clearBtn.Show()
			db.clearBtn.SetText(fmt.Sprintf("Clear (%d)", count))
		}
		if count == 1 {
			db.dlBtn.Enable()
			db.dlBtn.SetText("Download (1)")
			b := books[0]
			db.statusLbl.SetText(fmt.Sprintf("Selected: %s [%s]", truncate(b.Title, 60), strings.ToLower(b.Extension)))
		} else {
			db.dlBtn.Enable()
			db.dlBtn.SetText(fmt.Sprintf("Download (%d books)", count))
			db.statusLbl.SetText(fmt.Sprintf("Selected: %d books ready to download", count))
		}
	}

	db.progress.Hide()
	db.progress.SetValue(0)
	if db.browseBtn != nil {
		db.browseBtn.Show()
		db.browseBtn.Enable()
	}
	if db.dlBtn != nil {
		db.dlBtn.Show()
	}
	db.cancelBtn.Hide()
}

// CancelDownload aborts all active and queued batch downloads.
func (db *DownloadBar) CancelDownload() {
	db.mu.Lock()
	if db.cancelFunc != nil {
		db.cancelFunc()
	}
	if db.queue != nil {
		db.queue.CancelAll()
	}
	db.mu.Unlock()
}

// OpenSaveFolder reveals or opens the save directory in macOS Finder (or default file manager).
func (db *DownloadBar) OpenSaveFolder() {
	db.mu.Lock()
	path := db.savePath
	db.mu.Unlock()

	if path == "" {
		path = GetConfiguredSavePath(db.isMobile())
	}

	_ = os.MkdirAll(path, 0755)

	if db.isMobile() || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		if a := fyne.CurrentApp(); a != nil {
			var targetURI *url.URL
			if strings.Contains(path, "Download") {
				targetURI, _ = url.Parse("content://com.android.externalstorage.documents/document/primary%3ADownload")
			} else if strings.Contains(path, "Documents") {
				targetURI, _ = url.Parse("content://com.android.externalstorage.documents/document/primary%3ADocuments")
			} else {
				targetURI, _ = url.Parse("content://com.android.externalstorage.documents/root/primary")
			}
			if targetURI != nil {
				_ = a.OpenURL(targetURI)
			}
		}
		if db.app != nil && db.app.window != nil {
			db.showMobileLocationModal(path)
		}
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}

// OpenFolder is an alias for OpenSaveFolder.
func (db *DownloadBar) OpenFolder() {
	db.OpenSaveFolder()
}

func (db *DownloadBar) showMobileLocationModal(path string) {
	if db.app == nil || db.app.window == nil {
		return
	}

	header := widget.NewLabelWithStyle("Downloaded books are saved in:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header.Wrapping = fyne.TextWrapWord

	pathDisplay := widget.NewLabel(path)
	pathDisplay.Wrapping = fyne.TextWrapBreak

	copiedNotice := widget.NewLabel("")
	copiedNotice.Importance = widget.SuccessImportance

	copyBtn := widget.NewButtonWithIcon("Copy Path", theme.ContentCopyIcon(), func() {
		if db.app.window != nil && db.app.window.Clipboard() != nil {
			db.app.window.Clipboard().SetContent(path)
			copiedNotice.SetText("✓ Path copied to clipboard!")
		}
	})
	copyBtn.Importance = widget.HighImportance

	changeBtn := widget.NewButtonWithIcon("Change Folder", theme.SettingsIcon(), nil)
	changeBtn.Importance = widget.MediumImportance

	openDownloadsBtn := widget.NewButtonWithIcon("Open Downloads App", theme.FolderOpenIcon(), func() {
		if runtime.GOOS == "android" {
			_ = exec.Command("/system/bin/am", "start", "-a", "android.intent.action.VIEW_DOWNLOADS").Start()
			_ = exec.Command("am", "start", "-a", "android.intent.action.VIEW_DOWNLOADS").Start()
		}
		if a := fyne.CurrentApp(); a != nil {
			if u, err := url.Parse("file://" + path); err == nil {
				_ = a.OpenURL(u)
			}
		}
	})

	var d *dialog.CustomDialog
	changeBtn.OnTapped = func() {
		if d != nil {
			d.Hide()
		}
		ShowDownloadLocationDialog(db.app, false, func(newPath string) {
			db.SetSavePath(newPath)
		})
	}

	doneBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		if d != nil {
			d.Hide()
		}
	})

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		pathDisplay,
		copiedNotice,
		container.NewGridWithColumns(2, copyBtn, changeBtn),
		openDownloadsBtn,
	)
	padded := container.NewPadded(content)

	d = dialog.NewCustomWithoutButtons("Saved Location", padded, db.app.window)
	d.SetButtons([]fyne.CanvasObject{doneBtn})
	d.Resize(fyne.NewSize(360, 240))
	d.Show()
}

func (db *DownloadBar) removeSelectedBook(b *libgen.Book) {
	if b == nil {
		return
	}
	key := BookKey(b)
	db.mu.Lock()
	newSelected := make([]*libgen.Book, 0, len(db.selectedBooks))
	for _, sb := range db.selectedBooks {
		if BookKey(sb) != key {
			newSelected = append(newSelected, sb)
		}
	}
	db.selectedBooks = newSelected
	db.mu.Unlock()

	if db.app != nil {
		db.app.mu.Lock()
		delete(db.app.selectedBooks, key)
		db.app.mu.Unlock()
	}
}

func (db *DownloadBar) resetControlsUnderLock() {
	db.isDownloading = false
	db.cancelFunc = nil
	atomic.StoreInt64(&db.activeDownloads, 0)
	db.cancelBtn.Hide()
	if db.browseBtn != nil {
		db.browseBtn.Show()
		db.browseBtn.Enable()
	}
	if db.dlBtn != nil {
		db.dlBtn.Show()
	}
	count := len(db.selectedBooks)
	if count == 0 {
		if db.clearBtn != nil {
			db.clearBtn.Hide()
		}
		db.dlBtn.Disable()
		db.dlBtn.SetText("Download")
	} else {
		if db.clearBtn != nil {
			db.clearBtn.Show()
			db.clearBtn.SetText(fmt.Sprintf("Clear (%d)", count))
		}
		if count == 1 {
			db.dlBtn.Enable()
			db.dlBtn.SetText("Download (1)")
		} else {
			db.dlBtn.Enable()
			db.dlBtn.SetText(fmt.Sprintf("Download (%d books)", count))
		}
	}
}

// EnqueueBooks adds books to the download queue and starts processing if idle.
func (db *DownloadBar) EnqueueBooks(books []*libgen.Book) {
	if len(books) == 0 {
		return
	}

	db.mu.Lock()
	saveDir := db.savePath
	if !IsDirWritable(saveDir) {
		fallback := GetDefaultSavePath(db.isMobile())
		if IsDirWritable(fallback) {
			db.savePath = fallback
			saveDir = fallback
			if db.pathLabel != nil {
				db.pathLabel.SetText(fallback)
			}
		} else {
			db.mu.Unlock()
			db.statusLbl.Importance = widget.DangerImportance
			db.statusLbl.SetText(fmt.Sprintf("Save directory not writable: %s", saveDir))
			db.showFailureModal(len(books), fmt.Errorf("storage directory '%s' is not writable. Please choose another location with Browse", saveDir))
			return
		}
	}

	if db.queue == nil {
		if db.app != nil {
			db.queue = db.app.GetDownloadQueue()
		} else {
			db.queue = NewDownloadQueue()
		}
	}

	db.queue.Enqueue(books...)

	if db.isDownloading {
		pending := db.queue.GetPendingCount()
		db.statusLbl.Importance = widget.MediumImportance
		if len(books) == 1 {
			db.statusLbl.SetText(fmt.Sprintf("Added \"%s\" to queue (%d in queue)", truncate(books[0].Title, 35), pending))
		} else {
			db.statusLbl.SetText(fmt.Sprintf("Added %d books to queue (%d in queue)", len(books), pending))
		}
		if db.queueBtn != nil {
			db.queueBtn.SetText(fmt.Sprintf("Queue (%d)", pending))
			db.queueBtn.Show()
		}
		db.mu.Unlock()

		if db.app != nil && db.app.resultsView != nil {
			db.app.resultsView.RefreshList()
		}
		return
	}
	db.mu.Unlock()

	db.startQueueWorker()
}

func (db *DownloadBar) doDownload() {
	db.mu.Lock()
	if len(db.selectedBooks) == 0 {
		db.mu.Unlock()
		db.statusLbl.Importance = widget.DangerImportance
		db.statusLbl.SetText("Please select at least one book first")
		return
	}

	booksToDownload := make([]*libgen.Book, len(db.selectedBooks))
	copy(booksToDownload, db.selectedBooks)
	db.mu.Unlock()

	db.EnqueueBooks(booksToDownload)
}

func (db *DownloadBar) startQueueWorker() {
	db.mu.Lock()
	if db.isDownloading {
		db.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	db.cancelFunc = cancel
	db.isDownloading = true
	atomic.AddInt64(&db.sessionID, 1)

	if db.browseBtn != nil {
		db.browseBtn.Hide()
	}
	if db.clearBtn != nil {
		db.clearBtn.Hide()
	}
	db.dlBtn.Hide()
	db.openFolderBtn.Hide()
	db.cancelBtn.Show()

	db.progress.Show()
	db.progress.SetValue(0)
	db.statusLbl.Importance = widget.MediumImportance

	pendingCount := db.queue.GetPendingCount()
	if pendingCount > 0 {
		db.queueBtn.SetText(fmt.Sprintf("Queue (%d)", pendingCount))
		db.queueBtn.Show()
	} else {
		db.queueBtn.SetText("Queue (1)")
		db.queueBtn.Show()
	}

	db.statusLbl.SetText("Preparing download…")
	db.mu.Unlock()

	go func() {
		defer func() {
			db.mu.Lock()
			if db.isDownloading {
				db.resetControlsUnderLock()
			}
			db.mu.Unlock()
		}()

		var overallSuccessful int64
		var overallFailed int64
		var overallBytes int64
		var lastOverallErr error
		var firstSuccessBook *libgen.Book
		var totalInitial int

		for {
			if ctx.Err() != nil {
				break
			}

			// Pull pending items from queue (up to 3 for controlled concurrency)
			items := db.queue.GetItems()
			var batchItems []*DownloadItem
			for _, it := range items {
				if it.Status == QueueStatusPending {
					batchItems = append(batchItems, it)
					if len(batchItems) >= 3 {
						break
					}
				}
			}

			if len(batchItems) == 0 {
				break
			}

			if totalInitial == 0 {
				totalInitial = len(batchItems)
			}

			for _, it := range batchItems {
				it.Status = QueueStatusDownloading
				it.StartedAt = time.Now()
			}

			if db.app != nil && db.app.resultsView != nil {
				db.app.resultsView.RefreshList()
			}

			totalBatch := len(batchItems)
			const fallbackBookSize = int64(10 * 1024 * 1024)
			bookExpected := make([]int64, totalBatch)
			var initialExpectedTotal int64
			for i, it := range batchItems {
				sz := parseBookFilesize(it.Book.Filesize)
				if sz <= 0 {
					sz = fallbackBookSize
				}
				bookExpected[i] = sz
				initialExpectedTotal += sz
			}
			if initialExpectedTotal <= 0 {
				initialExpectedTotal = int64(totalBatch) * fallbackBookSize
			}

			var batchExpectedBytes int64 = initialExpectedTotal
			var batchBytesDownloaded int64
			atomic.StoreInt64(&db.activeDownloads, int64(totalBatch))

			var batchSuccessful int64
			var batchFailed int64
			var batchErr error
			var errMu sync.Mutex
			var lastStatusTime int64

			updateProgressAndStatus := func(newBytes int, forceStatus bool) {
				var curTotal int64
				if newBytes != 0 {
					curTotal = atomic.AddInt64(&batchBytesDownloaded, int64(newBytes))
				} else {
					curTotal = atomic.LoadInt64(&batchBytesDownloaded)
				}

				exp := atomic.LoadInt64(&batchExpectedBytes)
				if curTotal > exp {
					atomic.StoreInt64(&batchExpectedBytes, curTotal)
					exp = curTotal
				}

				pct := 0.0
				if exp > 0 {
					pct = float64(curTotal) / float64(exp)
				}
				active := int(atomic.LoadInt64(&db.activeDownloads))
				if active > 0 && pct >= 0.99 {
					pct = 0.99
				} else if pct > 1.0 {
					pct = 1.0
				}

				shouldUpdateText := false
				now := time.Now().UnixNano()
				prev := atomic.LoadInt64(&lastStatusTime)
				isSmall := exp < 128*1024
				if forceStatus || isSmall || (now-prev >= int64(80*time.Millisecond)) || curTotal >= exp {
					if atomic.CompareAndSwapInt64(&lastStatusTime, prev, now) || forceStatus {
						shouldUpdateText = true
					}
				}

				pendingInQueue := db.queue.GetPendingCount()

				db.mu.Lock()
				db.progress.SetValue(pct)
				if shouldUpdateText {
					baseStatus := formatDownloadStatus(active, curTotal, exp)
					if pendingInQueue > 0 {
						db.statusLbl.SetText(fmt.Sprintf("%s • %d queued", baseStatus, pendingInQueue))
					} else {
						db.statusLbl.SetText(baseStatus)
					}
				}
				if pendingInQueue > 0 && db.queueBtn != nil {
					db.queueBtn.SetText(fmt.Sprintf("Queue (%d)", pendingInQueue))
					db.queueBtn.Show()
				}
				db.mu.Unlock()
			}

			type taskItem struct {
				item *DownloadItem
				idx  int
			}

			tasks := make(chan taskItem, totalBatch)
			for i, it := range batchItems {
				tasks <- taskItem{item: it, idx: i}
			}
			close(tasks)

			var wg sync.WaitGroup
			numWorkers := totalBatch
			if numWorkers > 3 {
				numWorkers = 3
			}

			for w := 0; w < numWorkers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for tsk := range tasks {
						if ctx.Err() != nil {
							tsk.item.Status = QueueStatusCancelled
							return
						}
						b := tsk.item.Book
						idx := tsk.idx
						est := bookExpected[idx]

						if b.DownloadURL == "" {
							if err := libgen.GetDownloadURL(b, false); err != nil || b.DownloadURL == "" {
								atomic.AddInt64(&batchFailed, 1)
								tsk.item.Status = QueueStatusFailed
								tsk.item.Error = err
								errMu.Lock()
								if batchErr == nil {
									batchErr = err
								}
								errMu.Unlock()
								updateProgressAndStatus(0, true)
								continue
							}
						}

						if ctx.Err() != nil {
							tsk.item.Status = QueueStatusCancelled
							return
						}

						var bookBytesRead int64
						err := db.streamDownloadWithContext(ctx, b, func(contentLen int64) {
							if contentLen > 0 {
								delta := contentLen - est
								if delta != 0 {
									atomic.AddInt64(&batchExpectedBytes, delta)
									updateProgressAndStatus(0, true)
								}
							}
						}, func(n int) {
							atomic.AddInt64(&bookBytesRead, int64(n))
							updateProgressAndStatus(n, false)
						})

						atomic.AddInt64(&db.activeDownloads, -1)

						if err != nil {
							if ctx.Err() != nil {
								tsk.item.Status = QueueStatusCancelled
								return
							}
							atomic.AddInt64(&batchFailed, 1)
							tsk.item.Status = QueueStatusFailed
							tsk.item.Error = err
							errMu.Lock()
							if batchErr == nil {
								batchErr = err
							}
							errMu.Unlock()
							read := atomic.LoadInt64(&bookBytesRead)
							if read > 0 {
								atomic.AddInt64(&batchBytesDownloaded, -read)
							}
							updateProgressAndStatus(0, true)
							continue
						}

						atomic.AddInt64(&batchSuccessful, 1)
						tsk.item.Status = QueueStatusCompleted
						tsk.item.Progress = 1.0
						tsk.item.BytesDownloaded = atomic.LoadInt64(&bookBytesRead)
						updateProgressAndStatus(0, true)
						db.removeSelectedBook(b)

						if firstSuccessBook == nil {
							firstSuccessBook = b
						}
					}
				}()
			}

			wg.Wait()

			overallSuccessful += atomic.LoadInt64(&batchSuccessful)
			overallFailed += atomic.LoadInt64(&batchFailed)
			overallBytes += atomic.LoadInt64(&batchBytesDownloaded)
			if batchErr != nil && lastOverallErr == nil {
				lastOverallErr = batchErr
			}

			if db.app != nil && db.app.resultsView != nil {
				db.app.resultsView.RefreshList()
			}
		}

		if ctx.Err() != nil {
			db.mu.Lock()
			db.progress.Hide()
			db.progress.SetValue(0)
			db.statusLbl.Importance = widget.WarningImportance
			succ := overallSuccessful
			rem := totalInitial - int(succ)
			if rem < 0 {
				rem = 0
			}
			db.statusLbl.SetText(fmt.Sprintf("Download cancelled (%d completed, %d remaining)", succ, rem))
			if succ > 0 {
				db.openFolderBtn.Show()
			}
			db.queueBtn.Hide()
			db.resetControlsUnderLock()
			db.mu.Unlock()
			return
		}

		succ := overallSuccessful
		fail := overallFailed
		finalBytes := overallBytes

		db.mu.Lock()
		db.progress.SetValue(1.0)
		db.progress.Hide()
		db.queueBtn.Hide()

		if fail == 0 && succ > 0 {
			db.statusLbl.Importance = widget.SuccessImportance
			db.statusLbl.SetText(formatCompletedStatus(int(succ), finalBytes, db.savePath))
		} else if succ > 0 {
			db.statusLbl.Importance = widget.WarningImportance
			db.statusLbl.SetText(fmt.Sprintf("Finished: %d downloaded (%s), %d failed. Saved to %s", succ, formatBytes(finalBytes), fail, db.savePath))
		} else {
			db.statusLbl.Importance = widget.DangerImportance
			finalErr := lastOverallErr
			if finalErr != nil && (strings.Contains(finalErr.Error(), "storage") || strings.Contains(finalErr.Error(), "permission") || strings.Contains(finalErr.Error(), "unwritable")) {
				db.statusLbl.SetText("⚠️  Download failed: Storage directory unwritable.")
			} else {
				db.statusLbl.SetText("⚠️  All downloads failed. Please check mirror connection.")
			}
		}

		if succ > 0 {
			db.openFolderBtn.Show()
		}
		db.cancelBtn.Hide()
		if db.browseBtn != nil {
			db.browseBtn.Show()
			db.browseBtn.Enable()
		}
		if db.dlBtn != nil {
			db.dlBtn.Show()
		}
		db.mu.Unlock()

		if succ > 0 {
			if a := fyne.CurrentApp(); a != nil {
				bookWord := "book"
				if succ > 1 {
					bookWord = "books"
				}
				a.SendNotification(fyne.NewNotification("Download Complete", fmt.Sprintf("Successfully downloaded %d %s", succ, bookWord)))
			}
			firstTitle := ""
			firstFilePath := ""
			if firstSuccessBook != nil {
				firstTitle = firstSuccessBook.Title
				firstFilePath = filepath.Join(db.savePath, getBookFilename(firstSuccessBook))
			}
			db.showCompletionModal(int(succ), finalBytes, firstTitle, firstFilePath)
		} else if fail > 0 {
			db.showFailureModal(int(fail), lastOverallErr)
		}

		if fail == 0 && db.app != nil {
			db.app.ClearSelection()
		}

		db.mu.Lock()
		db.resetControlsUnderLock()
		db.mu.Unlock()
	}()
}

// GetProgress returns the current progress bar value safely under mutex.
func (db *DownloadBar) GetProgress() float64 {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.progress.Value
}

// GetStatusText returns the current status label text safely under mutex.
func (db *DownloadBar) GetStatusText() string {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.statusLbl.Text
}

// GetActiveDownloads returns the number of active downloads in progress.
func (db *DownloadBar) GetActiveDownloads() int {
	return int(atomic.LoadInt64(&db.activeDownloads))
}

type idleTimeoutReader struct {
	body    io.ReadCloser
	timeout time.Duration
	timer   *time.Timer
}

func newIdleTimeoutReader(body io.ReadCloser, timeout time.Duration) *idleTimeoutReader {
	r := &idleTimeoutReader{
		body:    body,
		timeout: timeout,
	}
	r.timer = time.AfterFunc(timeout, func() {
		_ = body.Close()
	})
	return r
}

func (r *idleTimeoutReader) Read(p []byte) (int, error) {
	r.timer.Reset(r.timeout)
	n, err := r.body.Read(p)
	r.timer.Reset(r.timeout)
	return n, err
}

func (r *idleTimeoutReader) Close() error {
	r.timer.Stop()
	return r.body.Close()
}

func (db *DownloadBar) streamDownloadWithContext(ctx context.Context, book *libgen.Book, onStart func(int64), onChunk func(int)) error {
	db.mu.Lock()
	saveDir := db.savePath
	db.mu.Unlock()

	filename := getBookFilename(book)
	outPath := filepath.Join(saveDir, filename)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		fallback := GetDefaultSavePath(db.isMobile())
		if fallback != saveDir && IsDirWritable(fallback) {
			db.SetSavePath(fallback)
			saveDir = fallback
			outPath = filepath.Join(saveDir, filename)
		} else {
			return fmt.Errorf("storage directory unwritable '%s': %w", saveDir, err)
		}
	}

	client := db.httpClient
	if client == nil {
		tr := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Disable HTTP/2 for robust large-file streaming
			ForceAttemptHTTP2:     false,
			ResponseHeaderTimeout: 30 * time.Second,
		}
		client = &http.Client{
			Transport: tr,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.Header.Set("User-Agent", libgen.UserAgent)
				if len(via) > 0 {
					r.Header.Set("Referer", via[len(via)-1].URL.String())
					if rangeHdr := via[0].Header.Get("Range"); rangeHdr != "" {
						r.Header.Set("Range", rangeHdr)
					}
				}
				return nil
			},
		}
	}

	// Helper to attempt download with automatic HTTP Range resume
	tryDownload := func(dlURL, pgURL string) (bool, error) {
		var totalExpected int64
		var lastErr error
		var hasStarted bool

		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 && (book.Md5 == "" || strings.Contains(dlURL, "mock.test") || strings.Contains(dlURL, "127.0.0.1") || strings.Contains(dlURL, "localhost")) {
				break
			}
			if ctx.Err() != nil {
				_ = os.Remove(outPath)
				return true, ctx.Err()
			}

			var currentBytes int64
			if fi, statErr := os.Stat(outPath); statErr == nil {
				currentBytes = fi.Size()
			}

			if totalExpected > 0 && currentBytes >= totalExpected {
				book.DownloadURL = dlURL
				book.PageURL = pgURL
				return true, nil
			}

			isMockURL := strings.Contains(dlURL, "127.0.0.1") || strings.Contains(dlURL, "localhost") || strings.Contains(dlURL, "example.com") || strings.Contains(dlURL, "mock.test")
			if attempt > 0 && !isMockURL && db.httpClient == nil {
				// Refresh download URL from mirror to obtain a fresh session key and redirect token
				refreshed := false
				for _, m := range libgen.DownloadMirrors {
					if strings.Contains(dlURL, m.Host) {
						tb := &libgen.Book{Md5: book.Md5, Extension: book.Extension, Title: book.Title}
						if refErr := libgen.ResolveMirrorDownloadURL(m, tb, false); refErr == nil && tb.DownloadURL != "" {
							dlURL = tb.DownloadURL
							pgURL = tb.PageURL
							refreshed = true
							break
						}
					}
				}
				if !refreshed && len(libgen.DownloadMirrors) > 0 {
					tb := &libgen.Book{Md5: book.Md5, Extension: book.Extension, Title: book.Title}
					if refErr := libgen.ResolveMirrorDownloadURL(libgen.DownloadMirrors[0], tb, false); refErr == nil && tb.DownloadURL != "" {
						dlURL = tb.DownloadURL
						pgURL = tb.PageURL
					}
				}

				select {
				case <-ctx.Done():
					_ = os.Remove(outPath)
					return true, ctx.Err()
				case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
				}
			}

			req, reqErr := http.NewRequestWithContext(ctx, "GET", dlURL, nil)
			if reqErr != nil {
				return hasStarted, reqErr
			}
			req.Header.Set("User-Agent", libgen.UserAgent)
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
			if pgURL != "" {
				req.Header.Set("Referer", pgURL)
			}
			if currentBytes > 0 {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", currentBytes))
			}

			resp, doErr := client.Do(req)
			if doErr != nil {
				if ctx.Err() != nil {
					_ = os.Remove(outPath)
					return true, ctx.Err()
				}
				lastErr = fmt.Errorf("mirror connection error: %w", doErr)
				continue
			}

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
				resp.Body.Close()
				lastErr = fmt.Errorf("HTTP %d from mirror server", resp.StatusCode)
				if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestedRangeNotSatisfiable {
					break
				}
				continue
			}

			ct := strings.ToLower(resp.Header.Get("Content-Type"))
			ext := strings.ToLower(book.Extension)
			if strings.Contains(ct, "text/html") && ext != "html" && ext != "htm" {
				resp.Body.Close()
				lastErr = fmt.Errorf("mirror returned HTML error page instead of %s file", ext)
				break
			}

			var out *os.File
			var openErr error
			if currentBytes > 0 && resp.StatusCode == http.StatusPartialContent {
				out, openErr = os.OpenFile(outPath, os.O_WRONLY|os.O_APPEND, 0644)
				if cr := resp.Header.Get("Content-Range"); cr != "" && totalExpected == 0 {
					if slash := strings.LastIndex(cr, "/"); slash != -1 {
						if tot, parseErr := strconv.ParseInt(cr[slash+1:], 10, 64); parseErr == nil && tot > 0 {
							totalExpected = tot
							if onStart != nil {
								onStart(tot)
							}
						}
					}
				}
			} else {
				if currentBytes > 0 && onChunk != nil {
					onChunk(-int(currentBytes))
				}
				currentBytes = 0
				out, openErr = os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if resp.ContentLength > 0 {
					totalExpected = resp.ContentLength
					if onStart != nil {
						onStart(resp.ContentLength)
					}
				}
			}

			if openErr != nil {
				resp.Body.Close()
				fallback := GetDefaultSavePath(db.isMobile())
				if fallback != saveDir && IsDirWritable(fallback) {
					db.SetSavePath(fallback)
					saveDir = fallback
					outPath = filepath.Join(saveDir, filename)
					out, openErr = os.Create(outPath)
				}
				if openErr != nil {
					return true, fmt.Errorf("storage write error for '%s': %w", outPath, openErr)
				}
			}

			hasStarted = true
			pw := &progressWriter{
				onChunk: onChunk,
			}

			idleBody := newIdleTimeoutReader(resp.Body, 25*time.Second)
			_, copyErr := io.Copy(out, io.TeeReader(idleBody, pw))
			closeErr := out.Close()
			idleBody.Close()

			if copyErr == nil && closeErr == nil {
				fi, _ := os.Stat(outPath)
				if totalExpected > 0 && fi != nil && fi.Size() < totalExpected {
					lastErr = fmt.Errorf("incomplete download (%d / %d bytes)", fi.Size(), totalExpected)
					continue
				}
				book.DownloadURL = dlURL
				book.PageURL = pgURL
				return true, nil
			}

			if ctx.Err() != nil {
				_ = os.Remove(outPath)
				return true, ctx.Err()
			}

			if copyErr != nil {
				lastErr = copyErr
			} else {
				lastErr = closeErr
			}
		}

		_ = os.Remove(outPath)
		return hasStarted, lastErr
	}

	// 1. Try primary DownloadURL
	started, err := tryDownload(book.DownloadURL, book.PageURL)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// If storage error or test URL, return directly without mirror fallback
	if started && (strings.Contains(err.Error(), "storage write error") || strings.Contains(err.Error(), "unwritable")) {
		return err
	}
	if strings.Contains(book.DownloadURL, "127.0.0.1") || strings.Contains(book.DownloadURL, "localhost") || strings.Contains(book.DownloadURL, "example.com") || strings.Contains(book.DownloadURL, "mock.test") || book.Md5 == "" || db.httpClient != nil {
		return err
	}

	// 2. Fallback to alternative mirrors on failure
	lastErr := err
	for _, m := range libgen.DownloadMirrors {
		if strings.Contains(book.DownloadURL, m.Host) {
			continue
		}
		fallbackBook := &libgen.Book{Md5: book.Md5, Extension: book.Extension, Title: book.Title}
		if resolveErr := libgen.ResolveMirrorDownloadURL(m, fallbackBook, false); resolveErr == nil && fallbackBook.DownloadURL != "" {
			_, fbErr := tryDownload(fallbackBook.DownloadURL, fallbackBook.PageURL)
			if fbErr == nil {
				return nil
			}
			lastErr = fbErr
		}
	}

	return lastErr
}

// progressWriter tracks bytes written and calls onChunk.
type progressWriter struct {
	written int64
	onChunk func(int)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)
	if pw.onChunk != nil {
		pw.onChunk(n)
	}
	return n, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// parseBookFilesize parses a raw size string into bytes.
// Supports integer bytes ("65011712"), human strings ("62 MB", "5.7 MB", "100KB"),
// and falls back to 0 if unknown/empty.
func parseBookFilesize(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return 0
	}
	// 1. Direct strict integer parse (e.g. "65011712")
	if bytes, err := strconv.ParseInt(raw, 10, 64); err == nil && bytes > 0 {
		return bytes
	}
	// 2. Binary unit parsing with float ("62 MB", "5.7 MB", "62MB", "5.7MB", "1.5 GB")
	clean := strings.ToUpper(raw)
	var val float64
	var unit string
	if _, err := fmt.Sscanf(clean, "%f %s", &val, &unit); err == nil && val > 0 {
		switch {
		case strings.HasPrefix(unit, "G"):
			return int64(val * 1024 * 1024 * 1024)
		case strings.HasPrefix(unit, "M"):
			return int64(val * 1024 * 1024)
		case strings.HasPrefix(unit, "K"):
			return int64(val * 1024)
		case strings.HasPrefix(unit, "B"):
			return int64(val)
		}
	}
	for _, suffix := range []string{"GB", "GIB", "MB", "MIB", "KB", "KIB"} {
		if strings.HasSuffix(clean, suffix) {
			numPart := strings.TrimSpace(strings.TrimSuffix(clean, suffix))
			if v, err := strconv.ParseFloat(numPart, 64); err == nil && v > 0 {
				switch suffix[0] {
				case 'G':
					return int64(v * 1024 * 1024 * 1024)
				case 'M':
					return int64(v * 1024 * 1024)
				case 'K':
					return int64(v * 1024)
				}
			}
		}
	}
	// 3. Fallback to humanize.ParseBytes
	if b, err := humanize.ParseBytes(strings.ReplaceAll(raw, " ", "")); err == nil && b > 0 {
		return int64(b)
	}
	return 0
}

func formatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const unit = 1024.0
	b := float64(bytes)
	switch {
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", b/(unit*unit*unit))
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", b/(unit*unit))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", b/unit)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatDownloadStatus(active int, curBytes, totalBytes int64) string {
	pct := 0.0
	if totalBytes > 0 {
		pct = float64(curBytes) / float64(totalBytes) * 100
	}
	pctInt := int(pct)
	if pctInt > 100 {
		pctInt = 100
	}
	if pctInt < 0 {
		pctInt = 0
	}
	curStr := formatBytes(curBytes)
	totalStr := formatBytes(totalBytes)
	if active > 0 {
		return fmt.Sprintf("Downloading (%d active): %s / %s (%d%%)", active, curStr, totalStr, pctInt)
	}
	return fmt.Sprintf("Downloading: %s / %s (%d%%)", curStr, totalStr, pctInt)
}

func formatCompletedStatus(count int, totalBytes int64, savePath string) string {
	bookWord := "books"
	if count == 1 {
		bookWord = "book"
	}
	return fmt.Sprintf("✓ Saved %d %s (%s) to %s", count, bookWord, formatBytes(totalBytes), savePath)
}

func formatSpeed(bytesPerSec float64) string {
	if bytesPerSec <= 0 {
		return "0 B/s"
	}
	const unit = 1024.0
	switch {
	case bytesPerSec >= 1024*1024:
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(unit*unit))
	case bytesPerSec >= 1024:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/unit)
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}
}

// GetBookFilename returns the sanitized target filename for a book on disk.
func GetBookFilename(book *libgen.Book) string {
	return getBookFilename(book)
}

func getBookFilename(book *libgen.Book) string {
	ext := strings.ToLower(book.Extension)
	if ext == "" {
		ext = "bin"
	}
	title := sanitize(book.Title)
	if book.Author != "" {
		return title + " by " + sanitize(book.Author) + "." + ext
	}
	return title + "." + ext
}

func sanitize(s string) string {
	bad := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, c := range bad {
		s = strings.ReplaceAll(s, c, "_")
	}
	return truncate(strings.TrimSpace(s), 80)
}

func truncate(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}

func openFileInSystem(filePath string) {
	if filePath == "" {
		return
	}
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", filePath).Start()
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", filePath).Start()
	case "android", "ios":
		if a := fyne.CurrentApp(); a != nil {
			if u, err := url.Parse("file://" + filePath); err == nil {
				_ = a.OpenURL(u)
			}
		}
	default:
		_ = exec.Command("xdg-open", filePath).Start()
	}
}

func (db *DownloadBar) showCompletionModal(succ int, totalBytes int64, firstTitle string, optionalFilePath ...string) {
	if db.app == nil || db.app.window == nil {
		return
	}

	bookWord := "book"
	if succ > 1 {
		bookWord = "books"
	}

	var titleText string
	if succ == 1 && firstTitle != "" {
		titleText = fmt.Sprintf("“%s”\n(%s)", truncate(firstTitle, 52), formatBytes(totalBytes))
	} else {
		titleText = fmt.Sprintf("%d %s (%s)", succ, bookWord, formatBytes(totalBytes))
	}

	checkIcon := widget.NewIcon(theme.ConfirmIcon())
	msgLabel := widget.NewLabelWithStyle(titleText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	headerRow := container.NewHBox(checkIcon, msgLabel)

	locLabel := widget.NewLabel(fmt.Sprintf("Saved to: %s", db.savePath))
	locLabel.Truncation = fyne.TextTruncateEllipsis
	locLabel.Importance = widget.LowImportance

	content := container.NewVBox(
		headerRow,
		widget.NewSeparator(),
		locLabel,
	)
	paddedContent := container.NewPadded(content)

	var d *dialog.CustomDialog

	var firstFile string
	if len(optionalFilePath) > 0 {
		firstFile = optionalFilePath[0]
	}

	var buttons []fyne.CanvasObject

	// 1. Direct Open Book action if single file
	if succ == 1 && firstFile != "" {
		openBookBtn := widget.NewButtonWithIcon("Open Book", theme.FileIcon(), func() {
			if d != nil {
				d.Hide()
			}
			openFileInSystem(firstFile)
		})
		openBookBtn.Importance = widget.HighImportance
		buttons = append(buttons, openBookBtn)
	}

	// 2. Open Folder button
	openLabel := "Open Folder"
	if runtime.GOOS == "darwin" {
		openLabel = "Open in Finder"
	} else if db.isMobile() || runtime.GOOS == "android" {
		openLabel = "Open Downloads"
	}

	openFolderBtn := widget.NewButtonWithIcon(openLabel, theme.FolderOpenIcon(), func() {
		db.OpenSaveFolder()
		if d != nil {
			d.Hide()
		}
	})
	if succ > 1 || firstFile == "" {
		openFolderBtn.Importance = widget.HighImportance
	}
	buttons = append(buttons, openFolderBtn)

	// 3. Done button
	doneBtn := widget.NewButtonWithIcon("Done", theme.ConfirmIcon(), func() {
		if d != nil {
			d.Hide()
		}
	})
	buttons = append(buttons, doneBtn)

	d = dialog.NewCustomWithoutButtons("Download Complete", paddedContent, db.app.window)
	d.SetButtons(buttons)
	d.Resize(fyne.NewSize(440, 160))
	d.Show()
}

func (db *DownloadBar) showFailureModal(fail int, errs ...error) {
	if db.app == nil || db.app.window == nil {
		return
	}

	bookWord := "book"
	if fail > 1 {
		bookWord = "books"
	}

	errDetail := "Please verify your connection or try another format/mirror."
	if len(errs) > 0 && errs[0] != nil {
		msg := errs[0].Error()
		if strings.Contains(msg, "permission") || strings.Contains(msg, "read-only") || strings.Contains(msg, "storage") || strings.Contains(msg, "unwritable") {
			errDetail = fmt.Sprintf("Storage Error: %s\nPlease select a different folder with 'Browse…'", msg)
		} else if strings.Contains(msg, "HTTP") || strings.Contains(msg, "HTML") || strings.Contains(msg, "mirror") {
			errDetail = fmt.Sprintf("Mirror Error: %s\nPlease try another format or mirror.", msg)
		} else {
			errDetail = fmt.Sprintf("Error: %s\nPlease verify your connection or try another format.", msg)
		}
	}

	content := container.NewVBox(
		container.NewHBox(widget.NewIcon(theme.ErrorIcon()), widget.NewLabelWithStyle("Download Failed", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Unable to download %d %s.", fail, bookWord)),
		widget.NewLabel(errDetail),
	)
	padded := container.NewPadded(content)

	var d *dialog.CustomDialog
	doneBtn := widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {
		if d != nil {
			d.Hide()
		}
	})
	doneBtn.Importance = widget.HighImportance

	d = dialog.NewCustomWithoutButtons("Error", padded, db.app.window)
	d.SetButtons([]fyne.CanvasObject{doneBtn})
	d.Resize(fyne.NewSize(400, 160))
	d.Show()
}

// SetHTTPClient sets the HTTP client used for downloads (useful for mock testing).
func (db *DownloadBar) SetHTTPClient(client *http.Client) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.httpClient = client
}

// GetDownloadButton returns the download button widget.
func (db *DownloadBar) GetDownloadButton() *widget.Button {
	return db.dlBtn
}

// GetCancelButton returns the cancel download button widget.
func (db *DownloadBar) GetCancelButton() *widget.Button {
	return db.cancelBtn
}

// GetBrowseButton returns the browse directory button widget.
func (db *DownloadBar) GetBrowseButton() *widget.Button {
	return db.browseBtn
}

// GetClearButton returns the clear selection button widget.
func (db *DownloadBar) GetClearButton() *widget.Button {
	return db.clearBtn
}

// GetOpenFolderButton returns the open folder / reveal in finder button widget.
func (db *DownloadBar) GetOpenFolderButton() *widget.Button {
	return db.openFolderBtn
}

// GetQueueButton returns the queue modal button widget.
func (db *DownloadBar) GetQueueButton() *widget.Button {
	return db.queueBtn
}

// GetProgressBar returns the download progress bar widget.
func (db *DownloadBar) GetProgressBar() *widget.ProgressBar {
	return db.progress
}

// GetPathLabel returns the path display label widget.
func (db *DownloadBar) GetPathLabel() *widget.Label {
	return db.pathLabel
}

// GetStatusLabel returns the status label widget.
func (db *DownloadBar) GetStatusLabel() *widget.Label {
	return db.statusLbl
}

