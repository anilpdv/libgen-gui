package ui

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ciehanski/libgen-cli/libgen"
)

// SearchBar contains the query input, format picker, search button,
// subtle activity spinner, and user-friendly status/error display with retry.
type SearchBar struct {
	app       *App
	entry     *widget.Entry
	format    *widget.Select
	btn       *widget.Button
	activity  *widget.Activity
	status     *widget.Label
	retryBtn    *widget.Button
	browserBtn  *widget.Button
	settingsBtn *widget.Button
	statusRow   *fyne.Container

	mu            sync.Mutex
	loading       bool
	failedPage    int
	lastQuery     string
	lastExt       string
	currentMirror url.URL
	searchFn      func(opts *libgen.SearchOptions) ([]*libgen.Book, error)
}

var formatOptions = []string{"Any", "epub", "pdf", "mobi", "djvu", "azw3", "fb2"}

func NewSearchBar(a *App) *SearchBar {
	sb := &SearchBar{app: a}

	sb.entry = widget.NewEntry()
	if sb.isMobile() {
		sb.entry.SetPlaceHolder("Search books, authors, ISBNs… (v2.0)")
	} else {
		sb.entry.SetPlaceHolder("Search for a book, author, or ISBN…")
	}

	sb.format = widget.NewSelect(formatOptions, nil)
	sb.format.SetSelected("Any")

	sb.activity = widget.NewActivity()
	sb.activity.Hide()

	sb.status = widget.NewLabel("")
	sb.status.Wrapping = fyne.TextWrapWord

	sb.retryBtn = widget.NewButtonWithIcon("Retry", theme.ViewRefreshIcon(), func() {
		sb.retryBtn.Hide()
		sb.browserBtn.Hide()
		sb.mu.Lock()
		targetPage := sb.failedPage
		sb.mu.Unlock()
		if targetPage <= 0 {
			targetPage = 1
		}
		sb.doSearchPage(targetPage)
	})
	sb.retryBtn.Importance = widget.WarningImportance
	sb.retryBtn.Hide()

	sb.browserBtn = widget.NewButtonWithIcon("Open in Browser", theme.NavigateNextIcon(), func() {
		sb.showBrowserMirrorsDialog()
	})
	sb.browserBtn.Importance = widget.LowImportance
	sb.browserBtn.Hide()

	sb.btn = widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() { sb.doSearchPage(1) })
	sb.btn.Importance = widget.HighImportance

	sb.settingsBtn = widget.NewButtonWithIcon("…", theme.SettingsIcon(), func() {
		if sb.app != nil {
			sb.app.ShowSettingsDialog()
		}
	})
	sb.settingsBtn.Importance = widget.LowImportance

	// Subscribe to live mirror health updates
	healthMgr := GetMirrorHealthManager()
	healthMgr.Subscribe(func(activeSearch, totalSearch, activeAll, totalAll int) {
		sb.updateMirrorStatusBadge(activeSearch, totalSearch)
	})

	// also trigger search on Enter key
	sb.entry.OnSubmitted = func(_ string) { sb.doSearchPage(1) }

	return sb
}

func (sb *SearchBar) updateMirrorStatusBadge(active, total int) {
	if sb.settingsBtn == nil {
		return
	}
	if sb.isMobile() {
		if total == 0 {
			sb.settingsBtn.SetText("…")
		} else {
			sb.settingsBtn.SetText(fmt.Sprintf("%d/%d", active, total))
		}
	} else {
		if total == 0 {
			sb.settingsBtn.SetText("Mirrors: …")
		} else {
			sb.settingsBtn.SetText(fmt.Sprintf("Mirrors: %d/%d", active, total))
		}
	}

	if active == 0 && total > 0 {
		sb.settingsBtn.Importance = widget.DangerImportance
	} else if active < total {
		sb.settingsBtn.Importance = widget.WarningImportance
	} else {
		sb.settingsBtn.Importance = widget.LowImportance
	}
	sb.settingsBtn.Refresh()
}

func (sb *SearchBar) isMobile() bool {
	if sb.app != nil {
		return sb.app.IsMobile()
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return true
	}
	return fyne.CurrentDevice().IsMobile()
}

func (sb *SearchBar) Widget() fyne.CanvasObject {
	// Clean, subtle status bar:
	// [Activity spinner] [Status message] ... [Browser button] [Retry button]
	actionsRow := container.NewHBox(sb.browserBtn, sb.retryBtn)
	sb.statusRow = container.NewBorder(
		nil, nil,
		sb.activity,
		actionsRow,
		sb.status,
	)

	if sb.isMobile() {
		// Mobile layout:
		// Row 1: Search entry (expanding) + Search button + Settings button
		// Row 2: Format dropdown + Status / spinner / action buttons
		searchRow := container.NewBorder(
			nil, nil, nil,
			container.NewHBox(sb.btn, sb.settingsBtn),
			sb.entry,
		)
		formatRow := container.NewBorder(
			nil, nil,
			sb.format,
			nil,
			sb.statusRow,
		)
		return container.NewVBox(searchRow, formatRow)
	}

	// Desktop layout:
	searchRow := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(sb.format, sb.btn, sb.settingsBtn),
		sb.entry,
	)

	return container.NewVBox(searchRow, sb.statusRow)
}

// SetStatus updates the search status message.
func (sb *SearchBar) SetStatus(msg string) {
	if sb.status != nil {
		sb.status.SetText(msg)
	}
}

func (sb *SearchBar) doSearchPage(page int) {
	sb.mu.Lock()
	if sb.loading {
		sb.mu.Unlock()
		return
	}
	sb.loading = true
	sb.mu.Unlock()

	var query, ext string
	if page <= 1 {
		page = 1
		query = strings.TrimSpace(sb.entry.Text)
		if query == "" {
			sb.mu.Lock()
			sb.loading = false
			sb.mu.Unlock()
			return
		}
		ext = sb.format.Selected
		if ext == "Any" {
			ext = ""
		}
		sb.mu.Lock()
		sb.lastQuery = query
		sb.lastExt = ext
		sb.mu.Unlock()

		if sb.app != nil {
			sb.app.ClearSelection()
		}
	} else {
		sb.mu.Lock()
		query = sb.lastQuery
		ext = sb.lastExt
		sb.mu.Unlock()
		if query == "" {
			query = strings.TrimSpace(sb.entry.Text)
			if query == "" {
				sb.mu.Lock()
				sb.loading = false
				sb.mu.Unlock()
				return
			}
			sb.mu.Lock()
			sb.lastQuery = query
			sb.mu.Unlock()
		}
	}

	sb.setSearching(true, page)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				sb.setSearching(false, page)
				sb.status.SetText(fmt.Sprintf("⚠️ Unexpected error: %v", r))
			}
		}()

		type searchResult struct {
			books []*libgen.Book
			err   error
		}

		resCh := make(chan searchResult, 1)

		timeout := GetNetworkTimeout()
		if timeout <= 0 {
			timeout = 15 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		go func() {
			sb.mu.Lock()
			mirror := sb.currentMirror
			sb.mu.Unlock()

			preferred := GetPreferredMirror()
			if preferred != "" && preferred != "auto" && mirror.Host != preferred {
				for _, m := range libgen.SearchMirrors {
					if m.Host == preferred {
						mirror = m
						break
					}
				}
			}

			if mirror.Host == "" {
				mirror = libgen.GetWorkingMirror(libgen.SearchMirrors)
			}

			working := mirror
			opts := &libgen.SearchOptions{
				Query:         query,
				SearchMirror:  mirror,
				Results:       25,
				AutoPaging:    ext != "",
				MaxPages:      5,
				Page:          page,
				Context:       ctx,
				WorkingMirror: &working,
			}
			if ext != "" {
				opts.Extension = []string{ext}
			}

			searchCall := libgen.Search
			sb.mu.Lock()
			if sb.searchFn != nil {
				searchCall = sb.searchFn
			}
			sb.mu.Unlock()

			books, err := searchCall(opts)
			if err == nil {
				sb.mu.Lock()
				if working.Host != "" {
					sb.currentMirror = working
				} else {
					sb.currentMirror = mirror
				}
				sb.mu.Unlock()
			} else if !strings.Contains(strings.ToLower(err.Error()), "no books found") {
				// Clear cached mirror on network/DNS error so future requests re-probe mirrors
				sb.mu.Lock()
				sb.currentMirror = url.URL{}
				sb.mu.Unlock()
			}
			resCh <- searchResult{books: books, err: err}
		}()

		// Clean timeout of 15 seconds so search/pagination never hangs indefinitely
		var res searchResult
		select {
		case res = <-resCh:
		case <-ctx.Done():
			res = searchResult{
				books: nil,
				err:   errors.New("connection timed out"),
			}
		}

		sb.setSearching(false, page)

		isNoResults := (res.err != nil && strings.Contains(strings.ToLower(res.err.Error()), "no books found")) || (res.err == nil && len(res.books) == 0)

		if isNoResults {
			sb.retryBtn.Hide()
			if page > 1 {
				sb.browserBtn.Hide()
				sb.status.SetText(fmt.Sprintf("No more results on page %d.", page))
				if sb.app.resultsView != nil {
					sb.app.resultsView.DisableNext()
				}
			} else {
				sb.status.SetText("No results found — try a different query or format.")
				sb.browserBtn.Show()
				sb.app.onResults(nil, 1)
			}
			return
		}

		if res.err != nil {
			sb.mu.Lock()
			sb.failedPage = page
			sb.mu.Unlock()
			friendly := userFriendlyError(res.err, page)
			sb.status.SetText("⚠️  " + friendly)
			sb.retryBtn.Show()
			sb.browserBtn.Show()
			return
		}

		sb.status.SetText("")
		sb.retryBtn.Hide()
		sb.browserBtn.Hide()
		sb.app.onResults(res.books, page)
	}()
}

func (sb *SearchBar) setSearching(active bool, page int) {
	sb.mu.Lock()
	sb.loading = active
	sb.mu.Unlock()

	if active {
		sb.btn.Disable()
		sb.entry.Disable()
		sb.format.Disable()
		if sb.app.resultsView != nil {
			sb.app.resultsView.SetLoading(true)
		}
		sb.retryBtn.Hide()
		sb.browserBtn.Hide()
		sb.activity.Show()
		sb.activity.Start()
		if page > 1 {
			sb.status.SetText(fmt.Sprintf("Loading page %d…", page))
		} else {
			sb.status.SetText("Searching…")
		}
	} else {
		sb.btn.Enable()
		sb.entry.Enable()
		sb.format.Enable()
		if sb.app.resultsView != nil {
			sb.app.resultsView.SetLoading(false)
		}
		sb.activity.Stop()
		sb.activity.Hide()
	}
}

func userFriendlyError(err error, page int) string {
	if err == nil {
		return ""
	}
	low := strings.ToLower(err.Error())

	var reason string
	switch {
	case strings.Contains(low, "timeout") || strings.Contains(low, "deadline exceeded"):
		reason = "connection timed out"
	case strings.Contains(low, "no such host") || strings.Contains(low, "lookup") || strings.Contains(low, "unreachable"):
		reason = "unable to reach mirrors (network error)"
	case strings.Contains(low, "connection refused") || strings.Contains(low, "reset by peer"):
		reason = "connection refused by mirror"
	case strings.Contains(low, "502") || strings.Contains(low, "bad gateway"):
		reason = "mirror server error (502 Bad Gateway)"
	case strings.Contains(low, "503") || strings.Contains(low, "service unavailable"):
		reason = "mirror temporarily unavailable (503)"
	case strings.Contains(low, "500") || strings.Contains(low, "internal server"):
		reason = "mirror internal server error (500)"
	case strings.Contains(low, "403") || strings.Contains(low, "forbidden"):
		reason = "mirror access restricted (403)"
	case strings.Contains(low, "no books found"):
		reason = "no books found matching criteria"
	default:
		msg := err.Error()
		if idx := strings.LastIndex(msg, ": "); idx != -1 && idx < len(msg)-2 {
			msg = msg[idx+2:]
		}
		reason = msg
	}

	if page > 1 {
		return fmt.Sprintf("Failed to load page %d: %s", page, reason)
	}
	return fmt.Sprintf("Search failed: %s", reason)
}

func (sb *SearchBar) showBrowserMirrorsDialog() {
	sb.mu.Lock()
	query := sb.lastQuery
	sb.mu.Unlock()
	if query == "" && sb.entry != nil {
		query = strings.TrimSpace(sb.entry.Text)
	}
	if query == "" {
		return
	}

	escaped := url.QueryEscape(query)

	if sb.app == nil || sb.app.window == nil {
		if u, err := url.Parse(fmt.Sprintf("https://annas-archive.org/search?q=%s", escaped)); err == nil {
			if a := fyne.CurrentApp(); a != nil {
				_ = a.OpenURL(u)
			}
		}
		return
	}

	var d *dialog.CustomDialog

	openLink := func(targetURL string) {
		if d != nil {
			d.Hide()
		}
		if u, err := url.Parse(targetURL); err == nil {
			if a := fyne.CurrentApp(); a != nil {
				_ = a.OpenURL(u)
			}
		}
	}

	header := widget.NewLabel(fmt.Sprintf("Search for %q on an alternative web mirror:", query))
	header.Wrapping = fyne.TextWrapWord

	btnAnnas := widget.NewButtonWithIcon("Anna's Archive (Universal Shadow Library)", theme.SearchIcon(), func() {
		openLink(fmt.Sprintf("https://annas-archive.org/search?q=%s", escaped))
	})
	btnAnnas.Importance = widget.HighImportance

	btnLibgenLi := widget.NewButtonWithIcon("LibGen.li (Standard Web Mirror)", theme.SearchIcon(), func() {
		openLink(fmt.Sprintf("https://libgen.li/index.php?req=%s", escaped))
	})

	btnLibgenIs := widget.NewButtonWithIcon("LibGen.is (Classic Web Mirror)", theme.SearchIcon(), func() {
		openLink(fmt.Sprintf("https://libgen.is/search.php?req=%s", escaped))
	})

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		btnAnnas,
		btnLibgenLi,
		btnLibgenIs,
	)

	d = dialog.NewCustom("Alternative Web Mirrors", "Close", content, sb.app.window)
	d.Show()
	d.Resize(fyne.NewSize(420, 220))
}

// SetSearchFunc allows overriding the search backend for testing.
func (sb *SearchBar) SetSearchFunc(fn func(opts *libgen.SearchOptions) ([]*libgen.Book, error)) {
	sb.mu.Lock()
	sb.searchFn = fn
	sb.mu.Unlock()
}

// GetEntry returns the search query input entry widget.
func (sb *SearchBar) GetEntry() *widget.Entry {
	return sb.entry
}

// GetFormatSelect returns the format dropdown widget.
func (sb *SearchBar) GetFormatSelect() *widget.Select {
	return sb.format
}

// GetSearchButton returns the primary search button.
func (sb *SearchBar) GetSearchButton() *widget.Button {
	return sb.btn
}

// GetRetryButton returns the retry button.
func (sb *SearchBar) GetRetryButton() *widget.Button {
	return sb.retryBtn
}

// GetBrowserButton returns the open in browser button.
func (sb *SearchBar) GetBrowserButton() *widget.Button {
	return sb.browserBtn
}

// GetSettingsButton returns the settings modal trigger button.
func (sb *SearchBar) GetSettingsButton() *widget.Button {
	return sb.settingsBtn
}

// GetStatusLabel returns the status display label.
func (sb *SearchBar) GetStatusLabel() *widget.Label {
	return sb.status
}
