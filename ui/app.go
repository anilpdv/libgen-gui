package ui

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"libgen-gui/pkg/libgen"
)

// App holds shared state across all UI panels.
type App struct {
	window        fyne.Window
	books         []*libgen.Book
	selectedBooks map[string]*libgen.Book
	mu            sync.Mutex

	currentPage int
	statsLbl    *widget.Label

	queue          *DownloadQueue
	mobileOverride *bool

	// cross-panel references (set during Build)
	downloadBar *DownloadBar
	resultsView *ResultsView
	searchBar   *SearchBar
}

func NewApp(w fyne.Window) *App {
	a := &App{
		window:        w,
		selectedBooks: make(map[string]*libgen.Book),
		currentPage:   1,
		queue:         NewDownloadQueue(),
	}
	a.statsLbl = widget.NewLabel("")
	a.statsLbl.Alignment = fyne.TextAlignCenter
	a.statsLbl.Hide()
	return a
}

// IsMobile returns true if running on a mobile device (Android, iOS) or if mobile mode was explicitly forced.
func (a *App) IsMobile() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.mobileOverride != nil {
		return *a.mobileOverride
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return true
	}
	return fyne.CurrentDevice().IsMobile()
}

// SetMobileMode overrides the mobile detection for testing or responsive previews.
func (a *App) SetMobileMode(mobile bool) {
	a.mu.Lock()
	a.mobileOverride = &mobile
	a.mu.Unlock()
}

// IsDirWritable checks if a directory exists and can be written to by creating and removing a test file.
func IsDirWritable(dir string) bool {
	if dir == "" {
		return false
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}
	testFile := filepath.Join(dir, fmt.Sprintf(".test_write_%d", time.Now().UnixNano()))
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return true
}

// GetDefaultSavePath determines a valid, writable download directory.
// On Android and iOS, it checks application-specific external and internal storage
// before falling back to temporary directories.
// On desktop systems, it checks ~/Downloads, ~/Desktop, and user home.
func GetDefaultSavePath(isMobile bool) string {
	if isMobile || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		candidates := []string{}

		// 1. Android public external app files directory
		// Standard path: /storage/emulated/0/Android/data/com.libgen.downloader/files/Download
		// This directory is always writable without permissions on Android 11-15 (API 30-35).
		if ext := os.Getenv("EXTERNAL_STORAGE"); ext != "" {
			candidates = append(candidates, filepath.Join(ext, "Android", "data", "com.libgen.downloader", "files", "Download"))
		}
		candidates = append(candidates,
			"/storage/emulated/0/Android/data/com.libgen.downloader/files/Download",
			"/sdcard/Android/data/com.libgen.downloader/files/Download",
		)

		// 2. Android public Downloads directory (works if MANAGE_EXTERNAL_STORAGE is active or permissions exist)
		candidates = append(candidates,
			"/storage/emulated/0/Download",
			"/sdcard/Download",
		)

		// 3. Fyne app storage root
		if fyne.CurrentApp() != nil && fyne.CurrentApp().Storage() != nil {
			if root := fyne.CurrentApp().Storage().RootURI(); root != nil && root.Path() != "" {
				candidates = append(candidates, filepath.Join(root.Path(), "Download"))
			}
		}

		// 4. Android internal files directories
		if filesDir := os.Getenv("FILESDIR"); filesDir != "" {
			candidates = append(candidates, filepath.Join(filesDir, "Download"))
		}
		candidates = append(candidates,
			"/data/user/0/com.libgen.downloader/files/Download",
			"/data/data/com.libgen.downloader/files/Download",
		)

		// Test mobile candidates in order
		for _, cand := range candidates {
			if IsDirWritable(cand) {
				return cand
			}
		}
	}

	// Desktop candidates (macOS, Windows, Linux)
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		dl := filepath.Join(home, "Downloads")
		if IsDirWritable(dl) {
			return dl
		}
		desktop := filepath.Join(home, "Desktop")
		if IsDirWritable(desktop) {
			return desktop
		}
		if IsDirWritable(home) {
			return home
		}
	}

	// System temp dir as guaranteed writable fallback
	tmp := filepath.Join(os.TempDir(), "LibGenDownloads")
	if IsDirWritable(tmp) {
		return tmp
	}
	return os.TempDir()
}

// ShowSettingsDialog displays the comprehensive application settings modal.
func (a *App) ShowSettingsDialog() {
	ShowFullSettingsDialog(a, 0)
}

// ShowMirrorsDialog displays the settings modal directly focused on the Mirrors tab.
func (a *App) ShowMirrorsDialog() {
	ShowFullSettingsDialog(a, 1)
}

func (a *App) Build() fyne.CanvasObject {
	// configured or default save path
	savePath := GetConfiguredSavePath(a.IsMobile())

	a.downloadBar = NewDownloadBar(a, savePath)
	a.resultsView = NewResultsView(a)
	a.searchBar = NewSearchBar(a)

	// Probe mirrors in background on startup (skip in test runner to prevent network delays)
	if flag.Lookup("test.v") == nil && !strings.HasSuffix(os.Args[0], ".test") {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			GetMirrorHealthManager().ProbeAll(ctx)
		}()
	}

	// Prompt user on first run to choose their download directory
	if a.window != nil && !IsDownloadFolderConfigured() {
		go func() {
			time.Sleep(250 * time.Millisecond)
			ShowDownloadLocationDialog(a, true, nil)
		}()
	}

	// layout: search on top, results in middle, download bar at bottom
	top := container.NewVBox(
		a.searchBar.Widget(),
		widget.NewSeparator(),
	)
	bottom := container.NewVBox(
		widget.NewSeparator(),
		a.downloadBar.Widget(),
	)

	return container.NewBorder(top, bottom, nil, nil, a.resultsView.Widget())
}

// BookKey returns a unique key for selection tracking
func BookKey(b *libgen.Book) string {
	if b == nil {
		return ""
	}
	if b.Md5 != "" {
		return strings.ToLower(b.Md5)
	}
	if b.ID != "" {
		return b.ID
	}
	return b.Title + "_" + b.Author + "_" + b.Year + "_" + b.Extension
}

// IsBookSelected checks if a book is currently in the selection map
func (a *App) IsBookSelected(b *libgen.Book) bool {
	if b == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	key := BookKey(b)
	_, exists := a.selectedBooks[key]
	return exists
}

// ToggleBookSelection toggles the selection state of a book
func (a *App) ToggleBookSelection(b *libgen.Book) {
	if b == nil {
		return
	}
	a.mu.Lock()
	key := BookKey(b)
	if _, exists := a.selectedBooks[key]; exists {
		delete(a.selectedBooks, key)
	} else {
		a.selectedBooks[key] = b
	}
	a.mu.Unlock()

	a.notifySelectionChanged()
}

// SetBookSelected sets the selection state of a book
func (a *App) SetBookSelected(b *libgen.Book, selected bool) {
	if b == nil {
		return
	}
	a.mu.Lock()
	key := BookKey(b)
	if selected {
		a.selectedBooks[key] = b
	} else {
		delete(a.selectedBooks, key)
	}
	a.mu.Unlock()

	a.notifySelectionChanged()
}

// SelectAllCurrentPage selects all books on the current page
func (a *App) SelectAllCurrentPage() {
	a.mu.Lock()
	for _, b := range a.books {
		key := BookKey(b)
		a.selectedBooks[key] = b
	}
	a.mu.Unlock()

	a.notifySelectionChanged()
	if a.resultsView != nil {
		a.resultsView.RefreshList()
	}
}

// ClearCurrentPage unselects all books on the current page
func (a *App) ClearCurrentPage() {
	a.mu.Lock()
	for _, b := range a.books {
		key := BookKey(b)
		delete(a.selectedBooks, key)
	}
	a.mu.Unlock()

	a.notifySelectionChanged()
	if a.resultsView != nil {
		a.resultsView.RefreshList()
	}
}

// ClearSelection removes all selections across all pages
func (a *App) ClearSelection() {
	a.mu.Lock()
	a.selectedBooks = make(map[string]*libgen.Book)
	a.mu.Unlock()

	a.notifySelectionChanged()
	if a.resultsView != nil {
		a.resultsView.RefreshList()
	}
}

// DeselectBook removes a specific book from the selection
func (a *App) DeselectBook(b *libgen.Book) {
	if b == nil {
		return
	}
	a.mu.Lock()
	key := BookKey(b)
	delete(a.selectedBooks, key)
	a.mu.Unlock()

	a.notifySelectionChanged()
	if a.resultsView != nil {
		a.resultsView.RefreshList()
	}
}

// DownloadSingleBook enqueues the specified book immediately. If downloading, it appends to queue.
func (a *App) DownloadSingleBook(b *libgen.Book) {
	if b == nil || a.downloadBar == nil {
		return
	}
	a.downloadBar.EnqueueBooks([]*libgen.Book{b})
}

// GetDownloadQueue returns the shared DownloadQueue instance.
func (a *App) GetDownloadQueue() *DownloadQueue {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.queue == nil {
		a.queue = NewDownloadQueue()
	}
	return a.queue
}

// GetBookQueueStatus returns the current queue status of a book, if any.
func (a *App) GetBookQueueStatus(b *libgen.Book) (QueueItemStatus, bool) {
	if a == nil || b == nil {
		return QueueStatusPending, false
	}
	a.mu.Lock()
	q := a.queue
	a.mu.Unlock()
	if q == nil {
		return QueueStatusPending, false
	}
	return q.GetBookStatus(b)
}

// GetSelectedBooks returns a slice of all currently selected books
func (a *App) GetSelectedBooks() []*libgen.Book {
	a.mu.Lock()
	defer a.mu.Unlock()
	res := make([]*libgen.Book, 0, len(a.selectedBooks))
	for _, b := range a.selectedBooks {
		res = append(res, b)
	}
	return res
}

func (a *App) notifySelectionChanged() {
	selected := a.GetSelectedBooks()
	if a.downloadBar != nil {
		a.downloadBar.SetSelectedBooks(selected)
	}
	if a.resultsView != nil {
		a.resultsView.UpdateHeaderSelectionState()
	}
}

// onResults is called by SearchBar after a successful search.
func (a *App) onResults(books []*libgen.Book, page int) {
	a.books = books
	a.currentPage = page
	var msg string
	if len(books) == 0 {
		if page > 1 {
			msg = fmt.Sprintf("0 results — Page %d", page)
		} else {
			msg = "0 results"
		}
	} else {
		msg = fmt.Sprintf("Found %d results — Page %d", len(books), page)
	}
	if a.searchBar != nil {
		a.searchBar.SetStatus(msg)
	}
	if a.statsLbl != nil {
		a.statsLbl.SetText(msg)
	}
	a.resultsView.SetBooks(books)
	a.notifySelectionChanged()
}

// SetResults updates the app's books and triggers results view update.
func (a *App) SetResults(books []*libgen.Book, page int) {
	a.onResults(books, page)
}

// GetSearchBar returns the search bar instance.
func (a *App) GetSearchBar() *SearchBar {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.searchBar
}

// GetResultsView returns the results view instance.
func (a *App) GetResultsView() *ResultsView {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.resultsView
}

// GetDownloadBar returns the download bar instance.
func (a *App) GetDownloadBar() *DownloadBar {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.downloadBar
}
