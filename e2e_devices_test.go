package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/ciehanski/libgen-cli/libgen"
	"libgen-gui/ui"
)

// DeviceProfile defines the screen dimensions and platform behavior for multi-device testing.
type DeviceProfile struct {
	Name        string
	Width       float32
	Height      float32
	IsMobile    bool
	IsTablet    bool
	Description string
}

var allDeviceProfiles = []DeviceProfile{
	{
		Name:        "Desktop_Standard",
		Width:       900,
		Height:      600,
		IsMobile:    false,
		Description: "Standard desktop window (macOS / Windows / Linux)",
	},
	{
		Name:        "Desktop_HiDPI",
		Width:       1440,
		Height:      900,
		IsMobile:    false,
		Description: "High-resolution desktop monitor display",
	},
	{
		Name:        "Mobile_Phone_Portrait",
		Width:       390,
		Height:      844,
		IsMobile:    true,
		Description: "Modern smartphone in portrait orientation (iPhone / Android)",
	},
	{
		Name:        "Mobile_Phone_Landscape",
		Width:       844,
		Height:      390,
		IsMobile:    true,
		Description: "Smartphone in landscape orientation",
	},
	{
		Name:        "Tablet_Portrait",
		Width:       800,
		Height:      1280,
		IsMobile:    true,
		IsTablet:    true,
		Description: "Tablet in portrait orientation (iPad / Android Tablet)",
	},
	{
		Name:        "Tablet_Landscape",
		Width:       1024,
		Height:      768,
		IsMobile:    true,
		IsTablet:    true,
		Description: "Tablet in landscape orientation (iPad / Android Tablet)",
	},
}

type e2eRoundTripper func(*http.Request) (*http.Response, error)

func (rt e2eRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// TestE2E_AppLaunch_AllDevices verifies initial GUI launch and layout across every device form factor.
func TestE2E_AppLaunch_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			if title := w.Title(); title != "LibGen Downloader v2.0" {
				t.Fatalf("[%s] expected title 'LibGen Downloader v2.0', got %q", dev.Name, title)
			}

			// Verify SearchBar widgets
			sb := appInstance.GetSearchBar()
			if sb == nil {
				t.Fatalf("[%s] searchBar is nil", dev.Name)
			}
			if sb.GetEntry() == nil || sb.GetSearchButton() == nil {
				t.Fatalf("[%s] search entry or button not initialized", dev.Name)
			}
			if sb.GetSearchButton().Disabled() {
				t.Errorf("[%s] search button should be enabled initially", dev.Name)
			}

			// Verify ResultsView widgets
			rv := appInstance.GetResultsView()
			if rv == nil {
				t.Fatalf("[%s] resultsView is nil", dev.Name)
			}
			if len(rv.GetBooks()) != 0 {
				t.Errorf("[%s] expected 0 books initially, got %d", dev.Name, len(rv.GetBooks()))
			}
			if !rv.GetPrevButton().Disabled() {
				t.Errorf("[%s] prev button should be disabled initially", dev.Name)
			}
			if !rv.GetNextButton().Disabled() {
				t.Errorf("[%s] next button should be disabled initially", dev.Name)
			}

			// Verify DownloadBar widgets
			db := appInstance.GetDownloadBar()
			if db == nil {
				t.Fatalf("[%s] downloadBar is nil", dev.Name)
			}
			if !db.GetDownloadButton().Disabled() {
				t.Errorf("[%s] download button should be disabled initially", dev.Name)
			}
			if db.GetProgressBar().Visible() {
				t.Errorf("[%s] progress bar should be hidden initially", dev.Name)
			}
			if db.GetCancelButton().Visible() {
				t.Errorf("[%s] cancel button should be hidden initially", dev.Name)
			}
		})
	}
}

// TestE2E_SearchAndPagination_AllDevices executes the full search and pagination user journey across all devices.
func TestE2E_SearchAndPagination_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			sb := appInstance.GetSearchBar()
			rv := appInstance.GetResultsView()

			// Mock books dataset
			mockBooksPage1 := make([]*libgen.Book, 25)
			for i := 0; i < 25; i++ {
				mockBooksPage1[i] = &libgen.Book{
					Title:     fmt.Sprintf("Book Volume %d", i+1),
					Author:    "Alan Turing",
					Year:      "2023",
					Extension: "pdf",
					Filesize:  "1048576",
					Md5:       fmt.Sprintf("md5_p1_%03d", i+1),
				}
			}
			mockBooksPage2 := make([]*libgen.Book, 10)
			for i := 0; i < 10; i++ {
				mockBooksPage2[i] = &libgen.Book{
					Title:     fmt.Sprintf("Advanced Volume %d", i+26),
					Author:    "Ada Lovelace",
					Year:      "2024",
					Extension: "pdf",
					Filesize:  "2097152",
					Md5:       fmt.Sprintf("md5_p2_%03d", i+26),
				}
			}

			// Configure search function hook
			var searchCalls int32
			sb.SetSearchFunc(func(opts *libgen.SearchOptions) ([]*libgen.Book, error) {
				atomic.AddInt32(&searchCalls, 1)
				if opts.Page == 2 {
					return mockBooksPage2, nil
				}
				return mockBooksPage1, nil
			})

			// 1. User types query into search entry using Fyne test.Type
			entry := sb.GetEntry()
			test.Type(entry, "Computer Architecture")
			if entry.Text != "Computer Architecture" {
				t.Fatalf("[%s] expected search entry text 'Computer Architecture', got %q", dev.Name, entry.Text)
			}

			// 2. User selects format filter
			sb.GetFormatSelect().SetSelected("pdf")
			if sb.GetFormatSelect().Selected != "pdf" {
				t.Fatalf("[%s] expected format 'pdf', got %q", dev.Name, sb.GetFormatSelect().Selected)
			}

			// 3. User taps Search button
			test.Tap(sb.GetSearchButton())

			// Wait for search result processing
			deadline := time.Now().Add(3 * time.Second)
			for {
				if len(rv.GetBooks()) == 25 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for search results", dev.Name)
				}
				time.Sleep(10 * time.Millisecond)
			}

			// Verify results state
			if len(rv.GetBooks()) != 25 {
				t.Errorf("[%s] expected 25 results, got %d", dev.Name, len(rv.GetBooks()))
			}
			if !rv.GetPrevButton().Disabled() {
				t.Errorf("[%s] prev button should be disabled on page 1", dev.Name)
			}
			if rv.GetNextButton().Disabled() {
				t.Errorf("[%s] next button should be enabled on page 1 with 25 results", dev.Name)
			}

			// 4. User taps Next Page button
			test.Tap(rv.GetNextButton())

			deadline = time.Now().Add(3 * time.Second)
			for {
				if len(rv.GetBooks()) == 10 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for page 2 results", dev.Name)
				}
				time.Sleep(10 * time.Millisecond)
			}

			if len(rv.GetBooks()) != 10 {
				t.Errorf("[%s] expected 10 books on page 2, got %d", dev.Name, len(rv.GetBooks()))
			}
			if rv.GetPrevButton().Disabled() {
				t.Errorf("[%s] prev button should be enabled on page 2", dev.Name)
			}
			if !rv.GetNextButton().Disabled() {
				t.Errorf("[%s] next button should be disabled on partial page (< 25 results)", dev.Name)
			}

			// 5. User taps Prev Page button to return to page 1
			test.Tap(rv.GetPrevButton())

			deadline = time.Now().Add(3 * time.Second)
			for {
				if len(rv.GetBooks()) == 25 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out navigating back to page 1", dev.Name)
				}
				time.Sleep(10 * time.Millisecond)
			}

			if len(rv.GetBooks()) != 25 {
				t.Errorf("[%s] expected 25 books after returning to page 1, got %d", dev.Name, len(rv.GetBooks()))
			}
		})
	}
}

// TestE2E_SelectionAndMultiPageRetention_AllDevices validates selection across pages and device layouts.
func TestE2E_SelectionAndMultiPageRetention_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			db := appInstance.GetDownloadBar()

			b1 := &libgen.Book{Title: "Book 1", Md5: "md5_b1", Filesize: "1000", Extension: "epub"}
			b2 := &libgen.Book{Title: "Book 2", Md5: "md5_b2", Filesize: "2000", Extension: "pdf"}
			b3 := &libgen.Book{Title: "Book 3", Md5: "md5_b3", Filesize: "3000", Extension: "mobi"}

			// Load page 1
			appInstance.SetResults([]*libgen.Book{b1, b2}, 1)

			// Initially 0 selected, button disabled
			if len(appInstance.GetSelectedBooks()) != 0 {
				t.Fatalf("[%s] expected 0 selected initially", dev.Name)
			}
			if !db.GetDownloadButton().Disabled() {
				t.Fatalf("[%s] download button should be disabled with 0 selected", dev.Name)
			}

			// Select b1
			appInstance.ToggleBookSelection(b1)
			if !appInstance.IsBookSelected(b1) {
				t.Errorf("[%s] b1 should be selected", dev.Name)
			}
			if db.GetDownloadButton().Disabled() {
				t.Errorf("[%s] download button should be enabled with 1 selected", dev.Name)
			}
			if db.GetDownloadButton().Text != "Download (1)" {
				t.Errorf("[%s] expected 'Download (1)', got %q", dev.Name, db.GetDownloadButton().Text)
			}

			// Select b2
			appInstance.ToggleBookSelection(b2)
			if len(appInstance.GetSelectedBooks()) != 2 {
				t.Errorf("[%s] expected 2 selected books", dev.Name)
			}
			if db.GetDownloadButton().Text != "Download (2 books)" {
				t.Errorf("[%s] expected 'Download (2 books)', got %q", dev.Name, db.GetDownloadButton().Text)
			}

			// Load page 2 with b3
			appInstance.SetResults([]*libgen.Book{b3}, 2)

			// Verify cross-page selection persistence
			if len(appInstance.GetSelectedBooks()) != 2 {
				t.Errorf("[%s] expected 2 selected books across page navigation, got %d", dev.Name, len(appInstance.GetSelectedBooks()))
			}
			if !appInstance.IsBookSelected(b1) || !appInstance.IsBookSelected(b2) {
				t.Errorf("[%s] b1 and b2 should remain selected", dev.Name)
			}

			// Select all on page 2 (adds b3)
			appInstance.SelectAllCurrentPage()
			if len(appInstance.GetSelectedBooks()) != 3 {
				t.Errorf("[%s] expected 3 selected books, got %d", dev.Name, len(appInstance.GetSelectedBooks()))
			}
			if db.GetDownloadButton().Text != "Download (3 books)" {
				t.Errorf("[%s] expected 'Download (3 books)', got %q", dev.Name, db.GetDownloadButton().Text)
			}

			// Clear all selections
			appInstance.ClearSelection()
			if len(appInstance.GetSelectedBooks()) != 0 {
				t.Errorf("[%s] expected 0 selected after ClearSelection", dev.Name)
			}
			if !db.GetDownloadButton().Disabled() {
				t.Errorf("[%s] download button should be disabled after ClearSelection", dev.Name)
			}
		})
	}
}

// TestE2E_FullDownloadLifecycle_AllDevices executes the end-to-end download flow with progress and completion.
func TestE2E_FullDownloadLifecycle_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			tempDir := t.TempDir()
			db := appInstance.GetDownloadBar()
			db.SetSavePath(tempDir)

			payload := strings.Repeat("FyneGraphicAppsE2ETestData", 512) // ~13 KB
			db.SetHTTPClient(&http.Client{
				Transport: e2eRoundTripper(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode:    http.StatusOK,
						Header:        make(http.Header),
						Body:          io.NopCloser(bytes.NewReader([]byte(payload))),
						ContentLength: int64(len(payload)),
					}, nil
				}),
			})

			book := &libgen.Book{
				Title:       fmt.Sprintf("Test Book for %s", dev.Name),
				Extension:   "epub",
				Filesize:    fmt.Sprintf("%d", len(payload)),
				DownloadURL: "http://mock.test/e2e-book",
				Md5:         fmt.Sprintf("md5_e2e_%s", dev.Name),
			}

			appInstance.ToggleBookSelection(book)
			dlBtn := db.GetDownloadButton()
			if dlBtn.Disabled() {
				t.Fatalf("[%s] download button should be enabled after book selection", dev.Name)
			}

			// Trigger download via Fyne test.Tap on the Download button
			test.Tap(dlBtn)

			// Wait for completion
			deadline := time.Now().Add(5 * time.Second)
			for {
				if db.GetProgress() == 1.0 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for download to complete", dev.Name)
				}
				time.Sleep(20 * time.Millisecond)
			}

			// Verify file was written to disk
			expectedFilename := ui.GetBookFilename(book)
			downloadedPath := filepath.Join(tempDir, expectedFilename)
			data, err := os.ReadFile(downloadedPath)
			if err != nil {
				t.Fatalf("[%s] expected downloaded file to exist at %s: %v", dev.Name, downloadedPath, err)
			}
			if string(data) != payload {
				t.Errorf("[%s] downloaded content mismatch", dev.Name)
			}

			// Verify UI post-download state
			if !db.GetOpenFolderButton().Visible() {
				t.Errorf("[%s] open folder button should be visible after download completion", dev.Name)
			}
			if db.GetProgressBar().Visible() {
				t.Errorf("[%s] progress bar should be hidden after completion", dev.Name)
			}
			if !strings.HasPrefix(db.GetStatusText(), "✓ Saved 1 book") {
				t.Errorf("[%s] expected status starting with '✓ Saved 1 book', got %q", dev.Name, db.GetStatusText())
			}
		})
	}
}

// TestE2E_InFlightCancel_AllDevices tests user cancelling active downloads mid-stream.
func TestE2E_InFlightCancel_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			tempDir := t.TempDir()
			db := appInstance.GetDownloadBar()
			db.SetSavePath(tempDir)

			startedStreaming := make(chan struct{})
			db.SetHTTPClient(&http.Client{
				Transport: e2eRoundTripper(func(req *http.Request) (*http.Response, error) {
					pr, pw := io.Pipe()
					go func() {
						<-req.Context().Done()
						_ = pw.CloseWithError(req.Context().Err())
					}()
					go func() {
						close(startedStreaming)
						for {
							if _, err := pw.Write([]byte("streaming block...")); err != nil {
								return
							}
							time.Sleep(15 * time.Millisecond)
						}
					}()
					return &http.Response{
						StatusCode:    http.StatusOK,
						Header:        make(http.Header),
						Body:          pr,
						ContentLength: 1048576,
					}, nil
				}),
			})

			book := &libgen.Book{
				Title:       fmt.Sprintf("Cancel Book %s", dev.Name),
				Extension:   "pdf",
				Filesize:    "1048576",
				DownloadURL: "http://mock.test/cancel-book",
				Md5:         fmt.Sprintf("md5_cancel_%s", dev.Name),
			}

			appInstance.ToggleBookSelection(book)
			test.Tap(db.GetDownloadButton())

			// Wait for worker to begin streaming
			select {
			case <-startedStreaming:
			case <-time.After(2 * time.Second):
				t.Fatalf("[%s] timed out waiting for worker to start streaming", dev.Name)
			}

			// Cancel button should be visible during active download
			cancelBtn := db.GetCancelButton()
			if !cancelBtn.Visible() {
				t.Fatalf("[%s] cancel button should be visible during active download", dev.Name)
			}

			// User taps Cancel button using Fyne test.Tap
			test.Tap(cancelBtn)

			// Wait for cancellation
			deadline := time.Now().Add(3 * time.Second)
			for {
				if strings.Contains(db.GetStatusText(), "cancelled") {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for cancellation to complete", dev.Name)
				}
				time.Sleep(20 * time.Millisecond)
			}

			// Verify partial file cleaned up from disk
			filePath := filepath.Join(tempDir, ui.GetBookFilename(book))
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				t.Errorf("[%s] expected partial file to be removed from disk upon cancel", dev.Name)
			}
		})
	}
}

// TestE2E_MirrorFailureAndRetry_AllDevices verifies error UI and retry handling across all devices.
func TestE2E_MirrorFailureAndRetry_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			sb := appInstance.GetSearchBar()

			var attempt int32
			sb.SetSearchFunc(func(opts *libgen.SearchOptions) ([]*libgen.Book, error) {
				curr := atomic.AddInt32(&attempt, 1)
				if curr == 1 {
					return nil, errors.New("503 Service Unavailable")
				}
				return []*libgen.Book{
					{Title: "Recovered Book", Extension: "epub", Filesize: "1000", Md5: "recovered_md5"},
				}, nil
			})

			test.Type(sb.GetEntry(), "Distributed Systems")
			test.Tap(sb.GetSearchButton())

			// Wait for error state
			deadline := time.Now().Add(3 * time.Second)
			for {
				if sb.GetRetryButton().Visible() && sb.GetBrowserButton().Visible() {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for error state (retry: %v, browser: %v)",
						dev.Name, sb.GetRetryButton().Visible(), sb.GetBrowserButton().Visible())
				}
				time.Sleep(10 * time.Millisecond)
			}

			if !strings.Contains(sb.GetStatusLabel().Text, "503") {
				t.Errorf("[%s] status should mention 503 error, got %q", dev.Name, sb.GetStatusLabel().Text)
			}

			// User taps Retry button
			test.Tap(sb.GetRetryButton())

			// Wait for recovery
			deadline = time.Now().Add(3 * time.Second)
			rv := appInstance.GetResultsView()
			for {
				if len(rv.GetBooks()) == 1 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("[%s] timed out waiting for recovery after retry", dev.Name)
				}
				time.Sleep(10 * time.Millisecond)
			}

			if len(rv.GetBooks()) != 1 {
				t.Errorf("[%s] expected 1 book after retry, got %d", dev.Name, len(rv.GetBooks()))
			}
			if sb.GetRetryButton().Visible() {
				t.Errorf("[%s] retry button should be hidden after successful recovery", dev.Name)
			}
		})
	}
}

// TestE2E_SettingsModal_AllDevices verifies settings dialog rendering and path validation on each device.
func TestE2E_SettingsModal_AllDevices(t *testing.T) {
	for _, dev := range allDeviceProfiles {
		t.Run(dev.Name, func(t *testing.T) {
			testApp := test.NewApp()
			defer testApp.Quit()

			w, appInstance := setupApp(testApp)
			w.Resize(fyne.NewSize(dev.Width, dev.Height))
			appInstance.SetMobileMode(dev.IsMobile)

			sb := appInstance.GetSearchBar()
			settingsBtn := sb.GetSettingsButton()
			if settingsBtn == nil {
				t.Fatalf("[%s] settings button is nil", dev.Name)
			}

			// Tap settings button (should open dialog without crashing or freezing)
			test.Tap(settingsBtn)

			// Verify location presets for this device type
			presets := ui.GetLocationPresets(dev.IsMobile)
			if len(presets) == 0 {
				t.Fatalf("[%s] expected location presets for device", dev.Name)
			}

			for _, p := range presets {
				if p.Label == "" || p.Path == "" {
					t.Errorf("[%s] invalid preset: %+v", dev.Name, p)
				}
				normalized := ui.NormalizePath(p.Path)
				if normalized == "" {
					t.Errorf("[%s] failed to normalize preset path: %s", dev.Name, p.Path)
				}
			}
		})
	}
}

// TestE2E_DeviceOrientationChange verifies dynamic window resizing and orientation switching.
func TestE2E_DeviceOrientationChange(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w, appInstance := setupApp(testApp)

	// 1. Start in Phone Portrait
	appInstance.SetMobileMode(true)
	w.Resize(fyne.NewSize(390, 844))

	b1 := &libgen.Book{Title: "Orientation Book 1", Extension: "epub", Filesize: "1000", Md5: "rot_1"}
	b2 := &libgen.Book{Title: "Orientation Book 2", Extension: "pdf", Filesize: "2000", Md5: "rot_2"}
	appInstance.GetResultsView().SetBooks([]*libgen.Book{b1, b2})

	// 2. Rotate to Phone Landscape
	w.Resize(fyne.NewSize(844, 390))

	// Verify selection still functions after rotation
	appInstance.ToggleBookSelection(b1)
	if !appInstance.IsBookSelected(b1) {
		t.Errorf("book 1 should be selected after landscape rotation")
	}

	// 3. Switch to Desktop window size and desktop mode
	appInstance.SetMobileMode(false)
	w.Resize(fyne.NewSize(900, 600))

	// Verify book selection retained across mode switch
	if !appInstance.IsBookSelected(b1) {
		t.Errorf("book 1 should remain selected after switching to desktop mode")
	}

	// 4. Switch to Tablet Landscape
	appInstance.SetMobileMode(true)
	w.Resize(fyne.NewSize(1024, 768))
	if len(appInstance.GetResultsView().GetBooks()) != 2 {
		t.Errorf("expected 2 books on tablet layout, got %d", len(appInstance.GetResultsView().GetBooks()))
	}
}
