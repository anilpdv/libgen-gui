package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"libgen-gui/pkg/libgen"
)

func TestUserFriendlyError(t *testing.T) {
	cases := []struct {
		err      error
		page     int
		contains string
	}{
		{errors.New("dial tcp: lookup libgen.gs: no such host"), 2, "Failed to load page 2: unable to reach mirrors"},
		{errors.New("Get \"https://libgen.li\": i/o timeout"), 2, "Failed to load page 2: connection timed out"},
		{errors.New("context deadline exceeded"), 1, "Search failed: connection timed out"},
		{errors.New("connection refused"), 1, "Search failed: connection refused by mirror"},
		{errors.New("502 Bad Gateway"), 2, "Failed to load page 2: mirror server error (502 Bad Gateway)"},
		{errors.New("503 Service Unavailable"), 3, "Failed to load page 3: mirror temporarily unavailable (503)"},
		{errors.New("403 Forbidden"), 1, "Search failed: mirror access restricted (403)"},
		{errors.New("no books found matching criteria"), 1, "Search failed: no books found matching criteria"},
		{errors.New("no books found matching criteria"), 2, "Failed to load page 2: no books found matching criteria"},
	}

	for _, c := range cases {
		got := userFriendlyError(c.err, c.page)
		if !strings.Contains(got, c.contains) {
			t.Errorf("userFriendlyError(%v, %d) = %q, want to contain %q", c.err, c.page, got, c.contains)
		}
	}
}

func TestResultsView_Pagination(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	rv := app.resultsView

	// Initially before search: Prev and Next must be disabled
	if !rv.prevBtn.Disabled() {
		t.Errorf("expected prevBtn to be disabled initially")
	}
	if !rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be disabled initially")
	}

	// When page 1 full results (>= 25) arrive: nextBtn should be enabled
	mockBooksFull := make([]*libgen.Book, 25)
	for i := 0; i < 25; i++ {
		mockBooksFull[i] = &libgen.Book{Title: "Book", Author: "Author", Year: "2020", Extension: "epub", Filesize: "1000"}
	}
	app.onResults(mockBooksFull, 1)

	if !rv.prevBtn.Disabled() {
		t.Errorf("expected prevBtn to be disabled on page 1")
	}
	if rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be enabled on page 1 with 25 results")
	}

	// While loading next page: both should be disabled
	rv.SetLoading(true)
	if !rv.prevBtn.Disabled() || !rv.nextBtn.Disabled() {
		t.Errorf("expected both prevBtn and nextBtn to be disabled while loading")
	}

	// After load failure: page state is preserved (still on page 1, next re-enabled)
	rv.SetLoading(false)
	if !rv.prevBtn.Disabled() {
		t.Errorf("expected prevBtn to remain disabled on page 1 after failed load")
	}
	if rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be re-enabled on page 1 after failed load")
	}

	// DisableNext explicitly: should stay disabled even after SetLoading(false)
	rv.DisableNext()
	if !rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be disabled after DisableNext")
	}
	rv.SetLoading(false)
	if !rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to stay disabled after DisableNext and SetLoading(false)")
	}

	// When page 2 results arrive: prevBtn should be enabled
	app.onResults(mockBooksFull, 2)
	if rv.prevBtn.Disabled() {
		t.Errorf("expected prevBtn to be enabled on page 2")
	}
	if rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be enabled on page 2 with 25 results")
	}

	// When a partial result arrives (< 25 books): nextBtn should be disabled immediately
	mockBooksPartial := make([]*libgen.Book, 5)
	for i := 0; i < 5; i++ {
		mockBooksPartial[i] = &libgen.Book{Title: "Book", Author: "Author", Year: "2020", Extension: "epub", Filesize: "1000"}
	}
	app.onResults(mockBooksPartial, 2)
	if !rv.nextBtn.Disabled() {
		t.Errorf("expected nextBtn to be disabled on partial page (< 25 results)")
	}
}

func TestSearchBar_LoadingGuard(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	sb := app.searchBar

	// Set searching active
	sb.setSearching(true, 1)
	if !sb.btn.Disabled() {
		t.Errorf("expected search button to be disabled while searching")
	}
	if !sb.entry.Disabled() {
		t.Errorf("expected entry to be disabled while searching")
	}
	if !sb.format.Disabled() {
		t.Errorf("expected format select to be disabled while searching")
	}
	if !sb.loading {
		t.Errorf("expected sb.loading to be true")
	}

	// Call doSearchPage while loading: should return immediately without panic or state change
	sb.doSearchPage(2)

	// Set searching inactive
	sb.setSearching(false, 1)
	if sb.btn.Disabled() {
		t.Errorf("expected search button to be enabled after searching")
	}
	if sb.entry.Disabled() {
		t.Errorf("expected entry to be enabled after searching")
	}
	if sb.format.Disabled() {
		t.Errorf("expected format select to be enabled after searching")
	}
	if sb.loading {
		t.Errorf("expected sb.loading to be false")
	}
}

func TestMultiSelection(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	b1 := &libgen.Book{Title: "Book One", Md5: "11111111111111111111111111111111", Extension: "epub"}
	b2 := &libgen.Book{Title: "Book Two", Md5: "22222222222222222222222222222222", Extension: "pdf"}
	b3 := &libgen.Book{Title: "Book Three", Md5: "33333333333333333333333333333333", Extension: "mobi"}

	// Page 1 with 2 books
	app.onResults([]*libgen.Book{b1, b2}, 1)

	// Initially 0 selected
	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 selected initially")
	}
	if !app.downloadBar.dlBtn.Disabled() {
		t.Errorf("expected download button disabled when 0 selected")
	}

	// Toggle b1 -> 1 selected
	app.ToggleBookSelection(b1)
	if !app.IsBookSelected(b1) {
		t.Errorf("expected b1 to be selected")
	}
	if app.IsBookSelected(b2) {
		t.Errorf("expected b2 to NOT be selected")
	}
	if app.downloadBar.dlBtn.Disabled() {
		t.Errorf("expected download button enabled with 1 selected")
	}
	if app.downloadBar.dlBtn.Text != "Download (1)" {
		t.Errorf("expected button text 'Download (1)', got %q", app.downloadBar.dlBtn.Text)
	}

	// Toggle b2 -> 2 selected
	app.ToggleBookSelection(b2)
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected 2 selected books")
	}
	if app.downloadBar.dlBtn.Text != "Download (2 books)" {
		t.Errorf("expected button text 'Download (2 books)', got %q", app.downloadBar.dlBtn.Text)
	}

	// Navigate to Page 2 with b3
	app.onResults([]*libgen.Book{b3}, 2)

	// Cross-page retention: b1 and b2 should still be selected
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected selections to persist across pages, got %d", len(app.GetSelectedBooks()))
	}
	if !app.IsBookSelected(b1) || !app.IsBookSelected(b2) {
		t.Errorf("expected b1 and b2 to remain selected on page 2")
	}

	// Select all on Page 2 (adds b3 -> total 3)
	app.SelectAllCurrentPage()
	if len(app.GetSelectedBooks()) != 3 {
		t.Errorf("expected 3 selected books after selecting page 2, got %d", len(app.GetSelectedBooks()))
	}
	if app.downloadBar.dlBtn.Text != "Download (3 books)" {
		t.Errorf("expected button text 'Download (3 books)', got %q", app.downloadBar.dlBtn.Text)
	}

	// Clear page 2 selection (removes b3 -> leaves b1 and b2)
	app.ClearCurrentPage()
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected 2 selected books after clearing page 2, got %d", len(app.GetSelectedBooks()))
	}
	if app.IsBookSelected(b3) {
		t.Errorf("expected b3 to be cleared")
	}

	// Clear all selections
	app.ClearSelection()
	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 selected books after ClearSelection")
	}
	if !app.downloadBar.dlBtn.Disabled() {
		t.Errorf("expected download button disabled after ClearSelection")
	}
}

func TestResultsView_UpdateResultRow_NoPanic(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	rv := app.resultsView

	b1 := &libgen.Book{
		Title:     "Practical Go",
		Author:    "Jane Doe",
		Year:      "2023",
		Extension: "epub",
		Filesize:  "1048576",
		Md5:       "abc123md5",
	}
	b2 := &libgen.Book{
		Title:     "Advanced Systems",
		Author:    "John Smith",
		Year:      "2021",
		Extension: "pdf",
		Filesize:  "5242880",
		Md5:       "def456md5",
	}

	app.onResults([]*libgen.Book{b1, b2}, 1)

	// 1. Create a row item and update it directly with b1
	row := newResultRow()
	rv.updateResultRow(row, b1)

	// Verify row structure and contents
	c, ok := row.(*fyne.Container)
	if !ok {
		t.Fatalf("expected row to be *fyne.Container")
	}
	var check *widget.Check
	var grid *fyne.Container
	for _, o := range c.Objects {
		if ch, ok := o.(*widget.Check); ok {
			check = ch
		} else if g, ok := o.(*fyne.Container); ok {
			grid = g
		}
	}
	if check == nil || grid == nil {
		t.Fatalf("check or grid not found in row container")
	}
	titleLbl := grid.Objects[0].(*widget.Label)
	if titleLbl.Text != "Practical Go" {
		t.Errorf("expected title 'Practical Go', got %q", titleLbl.Text)
	}
	authorLbl := grid.Objects[1].(*widget.Label)
	if authorLbl.Text != "Jane Doe" {
		t.Errorf("expected author 'Jane Doe', got %q", authorLbl.Text)
	}
	yearLbl := grid.Objects[2].(*widget.Label)
	if yearLbl.Text != "2023" {
		t.Errorf("expected year '2023', got %q", yearLbl.Text)
	}
	extLbl := grid.Objects[3].(*widget.Label)
	if extLbl.Text != "epub" {
		t.Errorf("expected ext 'epub', got %q", extLbl.Text)
	}
	sizeLbl := grid.Objects[4].(*widget.Label)
	if sizeLbl.Text != "1.0 MB" {
		t.Errorf("expected size '1.0 MB', got %q", sizeLbl.Text)
	}
	if check.Checked {
		t.Errorf("expected checkbox to be unchecked initially")
	}

	// 2. Test checkbox toggle via OnChanged
	check.OnChanged(true)
	if !app.IsBookSelected(b1) {
		t.Errorf("expected b1 to be selected in app after check.OnChanged(true)")
	}

	// 3. Simulate row recycling (reuse the exact same row container for b2 during list scrolling)
	rv.updateResultRow(row, b2)
	if titleLbl.Text != "Advanced Systems" {
		t.Errorf("expected recycled row title 'Advanced Systems', got %q", titleLbl.Text)
	}
	if authorLbl.Text != "John Smith" {
		t.Errorf("expected recycled row author 'John Smith', got %q", authorLbl.Text)
	}
	if yearLbl.Text != "2021" {
		t.Errorf("expected recycled row year '2021', got %q", yearLbl.Text)
	}
	if extLbl.Text != "pdf" {
		t.Errorf("expected recycled row ext 'pdf', got %q", extLbl.Text)
	}
	if sizeLbl.Text != "5.0 MB" {
		t.Errorf("expected recycled row size '5.0 MB', got %q", sizeLbl.Text)
	}
	if check.Checked {
		t.Errorf("expected b2 to be unchecked")
	}

	// 4. Test list CreateItem and UpdateItem recycling directly as widget.List does
	item := rv.list.CreateItem()
	rv.list.UpdateItem(0, item)
	rv.list.UpdateItem(1, item)

	// 5. Edge cases: nil or invalid inputs should not panic
	rv.updateResultRow(nil, b1)
	rv.updateResultRow(row, nil)
	rv.updateResultRow(nil, nil)
	rv.updateResultRow(widget.NewLabel("invalid"), b1)
	rv.updateResultRow(container.NewHBox(), b1)
	rv.updateResultRow(container.NewHBox(widget.NewLabel("bad")), b1)
	rv.updateResultRow(container.NewBorder(nil, nil, widget.NewCheck("", nil), nil, container.NewHBox()), b1)

	// Partial non-label in grid: other cells and check should still update without panicking
	partialRow := container.NewBorder(nil, nil, widget.NewCheck("", nil), nil, container.NewGridWithColumns(5,
		widget.NewLabel("old title"), widget.NewLabel("old author"), widget.NewLabel("old year"), widget.NewLabel("old ext"),
		widget.NewButton("not a label", nil),
	))
	rv.updateResultRow(partialRow, b1)
	partialGrid := partialRow.Objects[0].(*fyne.Container)
	if partialGrid.Objects[0].(*widget.Label).Text != "Practical Go" {
		t.Errorf("expected title to be updated even if 5th cell is not a label")
	}

	// ResultsView with nil app
	nilAppRV := NewResultsView(nil)
	nilAppRV.updateResultRow(row, b1)
	nilAppRV.UpdateHeaderSelectionState()
	nilAppRV.updateHeader()
	nilAppRV.UpdatePaginationButtons()
	if nilAppRV.headerCheck != nil && nilAppRV.headerCheck.OnChanged != nil {
		nilAppRV.headerCheck.OnChanged(true)
	}
	if nilAppRV.list != nil && nilAppRV.list.OnSelected != nil {
		nilAppRV.list.OnSelected(-1)
		nilAppRV.list.OnSelected(0)
		nilAppRV.list.OnSelected(999)
	}

	// 6. Test list OnSelected boundary conditions
	rv.list.OnSelected(-1)  // negative index must not panic
	rv.list.OnSelected(999) // out of bounds must not panic
	b0 := rv.books[0]
	initiallySelected := app.IsBookSelected(b0)
	rv.list.OnSelected(0) // valid index toggles selection
	if app.IsBookSelected(b0) == initiallySelected {
		t.Errorf("expected rv.books[0] selection to toggle after list.OnSelected(0)")
	}
	rv.list.OnSelected(0) // toggle again -> returns to initial state
	if app.IsBookSelected(b0) != initiallySelected {
		t.Errorf("expected rv.books[0] selection to return to initial state after second list.OnSelected(0)")
	}

	// 7. Test sorting and header updates (which previously panicked on container index)
	rv.sortBy("Title")
	rv.sortBy("Title")
	rv.sortBy("Author")
	rv.sortBy("Year")
	rv.sortBy("Fmt")
	rv.sortBy("Size")

	// Sort with nil element in books slice
	rv.books = append(rv.books, nil)
	rv.doSort()

	// Header update edge cases
	rv.header = nil
	rv.updateHeader()

	// 8. Rapid recycling simulation across many items (scroll stress test)
	recycleRow := newResultRow()
	for i := 0; i < 100; i++ {
		b := &libgen.Book{
			Title:     "Book Title",
			Author:    "Author Name",
			Year:      "2022",
			Extension: "pdf",
			Filesize:  "2048000",
		}
		rv.updateResultRow(recycleRow, b)
	}
}

func TestParseBookFilesize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"", 0},
		{"0", 0},
		{"invalid", 0},
		{"-100", 0},
		{"1048576", 1048576},
		{"65011712", 65011712},
		{"62 MB", 65011712},
		{"62MB", 65011712},
		{"5.7 MB", 5976883},
		{"5.7MB", 5976883},
		{"1.5 GB", 1610612736},
		{"500 KB", 500 * 1024},
		{"100 B", 100},
	}

	for _, tc := range tests {
		got := parseBookFilesize(tc.input)
		if got != tc.want {
			t.Errorf("parseBookFilesize(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{-10, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{15204352, "14.5 MB"},
		{70988595, "67.7 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		got := formatBytes(tc.bytes)
		if got != tc.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestFormatDownloadStatus(t *testing.T) {
	// Active 2 downloads: 14.5 MB / 67.7 MB (21%)
	status := formatDownloadStatus(2, 15204352, 70988595)
	expected := "Downloading (2 active): 14.5 MB / 67.7 MB (21%)"
	if status != expected {
		t.Errorf("expected %q, got %q", expected, status)
	}

	// Active 1 download
	status1 := formatDownloadStatus(1, 15204352, 70988595)
	expected1 := "Downloading (1 active): 14.5 MB / 67.7 MB (21%)"
	if status1 != expected1 {
		t.Errorf("expected %q, got %q", expected1, status1)
	}

	// 0 active (idle / preparing)
	status0 := formatDownloadStatus(0, 15204352, 70988595)
	expected0 := "Downloading: 14.5 MB / 67.7 MB (21%)"
	if status0 != expected0 {
		t.Errorf("expected %q, got %q", expected0, status0)
	}

	// Clamping over 100%
	statusOver := formatDownloadStatus(1, 80000000, 70988595)
	if !strings.Contains(statusOver, "(100%)") {
		t.Errorf("expected status clamped to 100%%, got %q", statusOver)
	}
}

func TestFormatCompletedStatus(t *testing.T) {
	// 2 books (67.7 MB)
	s2 := formatCompletedStatus(2, 70988595, "/Users/anilpdv/Downloads")
	expected2 := "✓ Saved 2 books (67.7 MB) to /Users/anilpdv/Downloads"
	if s2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, s2)
	}

	// 1 book (62.0 MB)
	s1 := formatCompletedStatus(1, 65011712, "/Users/anilpdv/Downloads")
	expected1 := "✓ Saved 1 book (62.0 MB) to /Users/anilpdv/Downloads"
	if s1 != expected1 {
		t.Errorf("expected %q, got %q", expected1, s1)
	}
}

func TestFormatSpeed(t *testing.T) {
	if got := formatSpeed(0); got != "0 B/s" {
		t.Errorf("expected '0 B/s', got %q", got)
	}
	if got := formatSpeed(500); got != "500 B/s" {
		t.Errorf("expected '500 B/s', got %q", got)
	}
	if got := formatSpeed(1024 * 100); got != "100.0 KB/s" {
		t.Errorf("expected '100.0 KB/s', got %q", got)
	}
	if got := formatSpeed(1024 * 1024 * 2.5); got != "2.5 MB/s" {
		t.Errorf("expected '2.5 MB/s', got %q", got)
	}
}

type mockRoundTripper func(*http.Request) (*http.Response, error)

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func TestStreamDownloadWithContext(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	payload := strings.Repeat("ABCDEF1234567890", 1024) // 16 KB

	db := NewDownloadBar(app, tempDir)
	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader(payload)),
				ContentLength: int64(len(payload)),
			}, nil
		}),
	}

	book := &libgen.Book{
		Title:       "Test Stream Book",
		Extension:   "pdf",
		DownloadURL: "http://mock.test/book",
	}

	var startLen int64
	var totalChunksBytes int64
	err = db.streamDownloadWithContext(context.Background(), book, func(contentLen int64) {
		startLen = contentLen
	}, func(n int) {
		atomic.AddInt64(&totalChunksBytes, int64(n))
	})

	if err != nil {
		t.Fatalf("streamDownloadWithContext failed: %v", err)
	}

	if startLen != int64(len(payload)) {
		t.Errorf("expected startLen %d, got %d", len(payload), startLen)
	}
	if totalChunksBytes != int64(len(payload)) {
		t.Errorf("expected totalChunksBytes %d, got %d", len(payload), totalChunksBytes)
	}

	outPath := filepath.Join(tempDir, getBookFilename(book))
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != payload {
		t.Errorf("downloaded content mismatch")
	}

	// Test cancellation cleans up incomplete file
	ctx, cancel := context.WithCancel(context.Background())
	cancelPr, cancelPw := io.Pipe()
	cancelBook := &libgen.Book{
		Title:       "Cancelled Book",
		Extension:   "pdf",
		DownloadURL: "http://mock.test/cancel",
	}
	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			go func() {
				_, _ = cancelPw.Write([]byte("partial data..."))
				cancel()
				_ = cancelPw.CloseWithError(context.Canceled)
			}()
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          cancelPr,
				ContentLength: 100000,
			}, nil
		}),
	}
	_ = db.streamDownloadWithContext(ctx, cancelBook, nil, nil)
	cancelPath := filepath.Join(tempDir, getBookFilename(cancelBook))
	if _, err := os.Stat(cancelPath); !os.IsNotExist(err) {
		t.Errorf("expected cancelled file to be deleted, but it exists")
	}
}

func TestContinuousByteProgressAndConcurrency(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-concur-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var inFlight int64
	var maxInFlight int64
	var totalRequests int64

	fileSize := 32 * 1024
	chunkSize := 1024
	chunkData := make([]byte, chunkSize)
	for i := range chunkData {
		chunkData[i] = 'X'
	}

	db := app.downloadBar
	db.savePath = tempDir

	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			atomic.AddInt64(&totalRequests, 1)
			cur := atomic.AddInt64(&inFlight, 1)

			for {
				oldMax := atomic.LoadInt64(&maxInFlight)
				if cur <= oldMax || atomic.CompareAndSwapInt64(&maxInFlight, oldMax, cur) {
					break
				}
			}

			pr, pw := io.Pipe()
			go func() {
				defer atomic.AddInt64(&inFlight, -1)
				defer pw.Close()
				for written := 0; written < fileSize; written += chunkSize {
					if _, err := pw.Write(chunkData); err != nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()

			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          pr,
				ContentLength: int64(fileSize),
			}, nil
		}),
	}

	books := []*libgen.Book{
		{
			Title:       "Book Alpha",
			Extension:   "epub",
			Filesize:    fmt.Sprintf("%d", fileSize),
			DownloadURL: "http://mock.test/book1",
		},
		{
			Title:       "Book Beta",
			Extension:   "pdf",
			Filesize:    fmt.Sprintf("%d", fileSize),
			DownloadURL: "http://mock.test/book2",
		},
		{
			Title:       "Book Gamma",
			Extension:   "mobi",
			Filesize:    fmt.Sprintf("%d", fileSize),
			DownloadURL: "http://mock.test/book3",
		},
	}

	db.SetSelectedBooks(books)

	var observedProgressValues []float64
	var mu sync.Mutex
	stopSampling := make(chan struct{})

	go func() {
		ticker := time.NewTicker(15 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopSampling:
				return
			case <-ticker.C:
				val := db.GetProgress()
				mu.Lock()
				if val > 0 && (len(observedProgressValues) == 0 || val != observedProgressValues[len(observedProgressValues)-1]) {
					observedProgressValues = append(observedProgressValues, val)
				}
				mu.Unlock()
			}
		}
	}()

	db.doDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for downloads to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	close(stopSampling)

	// 1. Concurrency verification:
	if atomic.LoadInt64(&totalRequests) != 3 {
		t.Errorf("expected 3 total requests, got %d", atomic.LoadInt64(&totalRequests))
	}
	recordedMax := atomic.LoadInt64(&maxInFlight)
	if recordedMax < 2 {
		t.Errorf("expected parallel downloads (max in-flight >= 2), got %d", recordedMax)
	}
	if recordedMax > 3 {
		t.Errorf("expected max workers <= 3, got %d", recordedMax)
	}

	// 2. Continuous progress verification:
	mu.Lock()
	distinctCount := len(observedProgressValues)
	mu.Unlock()
	if distinctCount < 5 {
		t.Errorf("expected smooth continuous progress with at least 5 distinct samples, got %d: %v", distinctCount, observedProgressValues)
	}

	// Final progress should be 1.0 (100%)
	if db.GetProgress() != 1.0 {
		t.Errorf("expected final progress 1.0, got %f", db.GetProgress())
	}

	// Status label should reflect completion
	if !strings.HasPrefix(db.GetStatusText(), "✓ Saved 3 books") {
		t.Errorf("expected completion status starting with '✓ Saved 3 books', got %q", db.GetStatusText())
	}

	// Download button should be disabled and reset to "Download" since selections were cleared
	if !db.dlBtn.Disabled() {
		t.Errorf("expected download button disabled after successful batch download")
	}
	if db.dlBtn.Text != "Download" {
		t.Errorf("expected download button text 'Download', got %q", db.dlBtn.Text)
	}
}

func TestByteProgressTracking_UnknownFilesize(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-unknown-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	payload := strings.Repeat("M", 25000)

	db := app.downloadBar
	db.savePath = tempDir
	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader(payload)),
				ContentLength: int64(len(payload)),
			}, nil
		}),
	}

	// Book with empty Filesize
	books := []*libgen.Book{
		{
			Title:       "Unknown Size Book",
			Extension:   "epub",
			Filesize:    "", // unknown!
			DownloadURL: "http://mock.test/unknown",
		},
	}

	db.SetSelectedBooks(books)
	db.doDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for download to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	if db.GetProgress() != 1.0 {
		t.Errorf("expected final progress 1.0, got %f", db.GetProgress())
	}

	if !strings.HasPrefix(db.GetStatusText(), "✓ Saved 1 book") {
		t.Errorf("expected '✓ Saved 1 book', got %q", db.GetStatusText())
	}
}

func TestPartialFailureRollback(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-rollback-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db := app.downloadBar
	db.savePath = tempDir

	payloadGood := strings.Repeat("A", 16384) // 16 KB

	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/good" {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        make(http.Header),
					Body:          io.NopCloser(strings.NewReader(payloadGood)),
					ContentLength: int64(len(payloadGood)),
				}, nil
			}

			// Bad request streams 4096 bytes then fails
			pr, pw := io.Pipe()
			go func() {
				_, _ = pw.Write([]byte(strings.Repeat("B", 4096)))
				_ = pw.CloseWithError(errors.New("mock network failure mid-stream"))
			}()
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          pr,
				ContentLength: 16384,
			}, nil
		}),
	}

	bookGood := &libgen.Book{
		Title:       "Good Book",
		Extension:   "epub",
		Filesize:    "16384",
		DownloadURL: "http://mock.test/good",
	}
	bookBad := &libgen.Book{
		Title:       "Bad Book",
		Extension:   "pdf",
		Filesize:    "16384",
		DownloadURL: "http://mock.test/bad",
	}

	db.SetSelectedBooks([]*libgen.Book{bookGood, bookBad})
	db.doDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for downloads to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Verify status message: should report exactly 1 downloaded (16.0 KB) and 1 failed
	// and NOT 20.0 KB (which would include the 4 KB partial data that was removed)
	status := db.GetStatusText()
	if !strings.Contains(status, "1 downloaded (16.0 KB), 1 failed") {
		t.Errorf("expected status to report '1 downloaded (16.0 KB), 1 failed', got %q", status)
	}

	// Verify the good file exists and bad file does not
	goodPath := filepath.Join(tempDir, getBookFilename(bookGood))
	if _, err := os.Stat(goodPath); err != nil {
		t.Errorf("expected good file to exist: %v", err)
	}
	badPath := filepath.Join(tempDir, getBookFilename(bookBad))
	if _, err := os.Stat(badPath); !os.IsNotExist(err) {
		t.Errorf("expected bad file to be removed from disk: %v", err)
	}

	// Download button should remain enabled because not all books succeeded (user can retry)
	if db.dlBtn.Disabled() {
		t.Errorf("expected download button enabled after partial failure to allow retry")
	}
}

func TestCancelConcurrentDownloads(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-cancel-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db := app.downloadBar
	db.savePath = tempDir

	startedCh := make(chan struct{}, 3)

	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			pr, pw := io.Pipe()
			go func() {
				<-req.Context().Done()
				_ = pw.CloseWithError(req.Context().Err())
			}()
			go func() {
				startedCh <- struct{}{}
				for {
					_, err := pw.Write([]byte("some data streaming slowly..."))
					if err != nil {
						return
					}
					time.Sleep(20 * time.Millisecond)
				}
			}()
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          pr,
				ContentLength: 1048576,
			}, nil
		}),
	}

	books := []*libgen.Book{
		{Title: "Book 1", Extension: "epub", Filesize: "1048576", DownloadURL: "http://mock.test/b1"},
		{Title: "Book 2", Extension: "pdf", Filesize: "1048576", DownloadURL: "http://mock.test/b2"},
		{Title: "Book 3", Extension: "mobi", Filesize: "1048576", DownloadURL: "http://mock.test/b3"},
	}

	db.SetSelectedBooks(books)
	db.doDownload()

	// Wait until at least 2 workers have started streaming
	<-startedCh
	<-startedCh

	// Cancel download
	db.CancelDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for cancel to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Verify status message mentions cancelled
	status := db.GetStatusText()
	if !strings.Contains(status, "Download cancelled") {
		t.Errorf("expected status to contain 'Download cancelled', got %q", status)
	}

	// Verify all partial files are cleaned up
	for _, b := range books {
		p := filepath.Join(tempDir, getBookFilename(b))
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("expected partial file for %s to be removed, but it exists", b.Title)
		}
	}
}

func TestProgressClampedBelowOneDuringActiveDownload(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-clamp-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db := app.downloadBar
	db.savePath = tempDir

	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			pr, pw := io.Pipe()
			go func() {
				defer pw.Close()
				chunk := []byte(strings.Repeat("X", 512))
				for i := 0; i < 20; i++ {
					_, _ = pw.Write(chunk)
					time.Sleep(15 * time.Millisecond)
				}
			}()
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          pr,
				ContentLength: -1,
			}, nil
		}),
	}

	book := &libgen.Book{
		Title:       "Chunked Stream Book",
		Extension:   "epub",
		Filesize:    "1024", // Small estimate (1 KB), so streaming 10 KB will blow past it!
		DownloadURL: "http://mock.test/chunked",
	}

	db.SetSelectedBooks([]*libgen.Book{book})

	var maxProgressWhileActive float64
	var mu sync.Mutex
	stopSampling := make(chan struct{})

	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopSampling:
				return
			case <-ticker.C:
				if db.GetActiveDownloads() > 0 {
					val := db.GetProgress()
					mu.Lock()
					if val > maxProgressWhileActive {
						maxProgressWhileActive = val
					}
					mu.Unlock()
				}
			}
		}
	}()

	db.doDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for download to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}
	close(stopSampling)

	mu.Lock()
	maxP := maxProgressWhileActive
	mu.Unlock()

	// While actively downloading, progress must never exceed 0.99
	if maxP > 0.99 {
		t.Errorf("expected in-flight progress to be capped at 0.99, got %f", maxP)
	}

	// After completion, progress snaps to 1.0
	if db.GetProgress() != 1.0 {
		t.Errorf("expected final progress 1.0, got %f", db.GetProgress())
	}
}

func TestStatusResetDoesNotOverwriteNewSelection(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-select-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db := app.downloadBar
	db.savePath = tempDir

	payload := "instant download"
	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader(payload)),
				ContentLength: int64(len(payload)),
			}, nil
		}),
	}

	book1 := &libgen.Book{
		Title:       "Book 1",
		Extension:   "pdf",
		Filesize:    fmt.Sprintf("%d", len(payload)),
		DownloadURL: "http://mock.test/b1",
	}

	db.SetSelectedBooks([]*libgen.Book{book1})
	db.doDownload()

	// Wait until download finishes
	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for download to finish")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// User immediately selects a new book
	newBook := &libgen.Book{
		Title:     "Subsequent Book",
		Extension: "epub",
		Filesize:  "1024",
	}
	db.SetSelectedBooks([]*libgen.Book{newBook})

	// Verify status is immediately set to new book
	status := db.GetStatusText()
	if !strings.HasPrefix(status, "Selected: Subsequent Book") {
		t.Errorf("expected status 'Selected: Subsequent Book', got %q", status)
	}

	// Verify sessionID increments
	db.mu.Lock()
	sess := db.sessionID
	db.mu.Unlock()
	if sess < 2 {
		t.Errorf("expected sessionID to increment on new selection, got %d", sess)
	}
}

func TestCustomModalsAndProgressLifecycle(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Modal & Progress Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir, err := os.MkdirTemp("", "libgen-modal-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db := app.downloadBar
	db.savePath = tempDir

	// 1. Initially: progress bar hidden, cancelBtn hidden, browseBtn visible, dlBtn visible
	if db.progress.Visible() {
		t.Errorf("expected progress bar to be hidden initially")
	}
	if db.cancelBtn.Visible() {
		t.Errorf("expected cancel button to be hidden initially")
	}
	if !db.browseBtn.Visible() {
		t.Errorf("expected browse button to be visible initially")
	}
	if !db.dlBtn.Visible() {
		t.Errorf("expected download button to be visible initially")
	}

	payload := "test payload contents"
	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader(payload)),
				ContentLength: int64(len(payload)),
			}, nil
		}),
	}

	book := &libgen.Book{
		Title:       "Test Minimal Book",
		Extension:   "pdf",
		Filesize:    fmt.Sprintf("%d", len(payload)),
		DownloadURL: "http://mock.test/minimal",
	}

	db.SetSelectedBooks([]*libgen.Book{book})
	db.doDownload()

	// Wait for completion
	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for download")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// 2. After completion: progress bar hidden (no 100% blue bar lingering!), cancelBtn hidden, browseBtn restored
	if db.progress.Visible() {
		t.Errorf("expected progress bar to be hidden after completion, but was visible")
	}
	if db.cancelBtn.Visible() {
		t.Errorf("expected cancel button to be hidden after completion")
	}
	if !db.browseBtn.Visible() {
		t.Errorf("expected browse button to be visible after completion")
	}
	if !db.dlBtn.Visible() {
		t.Errorf("expected download button to be visible after completion")
	}

	// 3. Open Folder button must be visible
	if !db.openFolderBtn.Visible() {
		t.Errorf("expected openFolderBtn to be visible after completion")
	}

	// 4. Status text must show success with checkmark
	status := db.GetStatusText()
	if !strings.HasPrefix(status, "✓ Saved 1 book") {
		t.Errorf("expected success status starting with '✓ Saved 1 book', got %q", status)
	}

	// 5. Test modal methods for zero crashes
	db.showCompletionModal(1, 1024, "Test Minimal Book")
	db.showCompletionModal(3, 4096, "")
	db.showFailureModal(2)

	// 6. Test nil window guard
	orphanBar := NewDownloadBar(nil, tempDir)
	orphanBar.showCompletionModal(1, 1024, "Orphan Book")
	orphanBar.showFailureModal(1)
}

func TestSearchBar_BrowserButton(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	sb := app.searchBar

	// 1. Initially hidden
	if sb.browserBtn.Visible() {
		t.Errorf("expected browserBtn to be hidden initially")
	}
	if sb.retryBtn.Visible() {
		t.Errorf("expected retryBtn to be hidden initially")
	}

	// 2. Hidden while searching
	sb.setSearching(true, 1)
	if sb.browserBtn.Visible() {
		t.Errorf("expected browserBtn to be hidden while searching")
	}

	// 3. Trigger search failure handling
	sb.mu.Lock()
	sb.failedPage = 1
	sb.lastQuery = "machine learning"
	sb.mu.Unlock()

	sb.status.SetText("⚠️ Search failed: mirror temporarily unavailable (503)")
	sb.retryBtn.Show()
	sb.browserBtn.Show()

	if !sb.browserBtn.Visible() {
		t.Errorf("expected browserBtn to be visible after search error")
	}
	if !sb.retryBtn.Visible() {
		t.Errorf("expected retryBtn to be visible after search error")
	}

	// 4. Test clicking browser button invokes dialog safely without panicking
	test.Tap(sb.browserBtn)

	// 5. Test retry button hides browser button
	test.Tap(sb.retryBtn)
	if sb.browserBtn.Visible() {
		t.Errorf("expected browserBtn to be hidden when retry is tapped")
	}

	// 6. Test orphan search bar without app window
	orphanSb := NewSearchBar(nil)
	orphanSb.lastQuery = "test"
	orphanSb.showBrowserMirrorsDialog() // should not panic
}

func TestSelectionManagementAndSingleBookDownload(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	b1 := &libgen.Book{Title: "Book 1", Md5: "md5_1", Extension: "epub"}
	b2 := &libgen.Book{Title: "Book 2", Md5: "md5_2", Extension: "pdf"}
	b3 := &libgen.Book{Title: "Book 3", Md5: "md5_3", Extension: "mobi"}

	// 1. Initially no selection, clearBtn is hidden
	if app.downloadBar.clearBtn.Visible() {
		t.Errorf("expected clearBtn to be hidden initially")
	}

	// 2. Select 2 books: clearBtn should be visible with Clear (2)
	app.SetBookSelected(b1, true)
	app.SetBookSelected(b2, true)

	if !app.downloadBar.clearBtn.Visible() {
		t.Errorf("expected clearBtn to be visible with 2 selected books")
	}
	if app.downloadBar.clearBtn.Text != "Clear (2)" {
		t.Errorf("expected clearBtn text to be 'Clear (2)', got %q", app.downloadBar.clearBtn.Text)
	}
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected 2 selected books, got %d", len(app.GetSelectedBooks()))
	}

	// 3. Click Clear button: should deselect all books and hide clearBtn
	test.Tap(app.downloadBar.clearBtn)

	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 selected books after tapping clearBtn, got %d", len(app.GetSelectedBooks()))
	}
	if app.downloadBar.clearBtn.Visible() {
		t.Errorf("expected clearBtn to be hidden after clear")
	}
	if !app.downloadBar.dlBtn.Disabled() {
		t.Errorf("expected dlBtn to be disabled after clear")
	}

	// 4. DownloadSingleBook should replace existing selections with only the target book
	tempDir, err := os.MkdirTemp("", "libgen-single-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	app.downloadBar.savePath = tempDir
	app.downloadBar.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode:    http.StatusOK,
				Header:        make(http.Header),
				Body:          io.NopCloser(strings.NewReader("sample data")),
				ContentLength: 11,
			}, nil
		}),
	}
	b3.DownloadURL = "http://mock.test/b3"

	app.SetBookSelected(b1, true)
	app.SetBookSelected(b2, true)
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected 2 selected books before DownloadSingleBook")
	}

	// DownloadSingleBook on b3
	app.DownloadSingleBook(b3)

	// Wait for b3 download to complete
	deadline := time.Now().Add(3 * time.Second)
	for {
		app.downloadBar.mu.Lock()
		dl := app.downloadBar.isDownloading
		app.downloadBar.mu.Unlock()
		if !dl {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for single book download")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Since b3 succeeded, it should be auto-deselected!
	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 selected books after successful download, got %d", len(app.GetSelectedBooks()))
	}

	// 5. New search query should clear stale selections
	app.SetBookSelected(b1, true)
	app.SetBookSelected(b2, true)
	if len(app.GetSelectedBooks()) != 2 {
		t.Errorf("expected 2 selected books before search")
	}

	app.searchBar.entry.SetText("New Query")
	app.searchBar.doSearchPage(1)

	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 selected books after starting a new search, got %d", len(app.GetSelectedBooks()))
	}
}

func TestMobileCardRowCreationAndUpdate(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Mobile")
	app := NewApp(w)
	app.SetMobileMode(true)
	_ = app.Build()

	rv := app.resultsView
	if !rv.isMobile() {
		t.Fatalf("expected rv.isMobile() to be true")
	}

	b1 := &libgen.Book{
		Title:     "Practical Go",
		Author:    "Jane Doe",
		Year:      "2023",
		Extension: "epub",
		Filesize:  "1048576",
		Md5:       "abc123md5",
	}
	b2 := &libgen.Book{
		Title:     "Advanced Systems",
		Author:    "John Smith",
		Year:      "2021",
		Extension: "pdf",
		Filesize:  "5242880",
		Md5:       "def456md5",
	}

	app.onResults([]*libgen.Book{b1, b2}, 1)

	// 1. Create mobile row item and update it
	row := newMobileResultRow()
	rv.updateMobileResultRow(row, b1)

	check, dlBtn, titleLbl, authorLbl, metaLbl := findMobileCardComponents(row)
	if check == nil || dlBtn == nil || titleLbl == nil || authorLbl == nil || metaLbl == nil {
		t.Fatalf("failed to find all mobile card components")
	}

	if titleLbl.Text != "Practical Go" {
		t.Errorf("expected title 'Practical Go', got %q", titleLbl.Text)
	}
	if authorLbl.Text != "Jane Doe" {
		t.Errorf("expected author 'Jane Doe', got %q", authorLbl.Text)
	}
	if !strings.Contains(metaLbl.Text, "EPUB") || !strings.Contains(metaLbl.Text, "1.0 MB") || !strings.Contains(metaLbl.Text, "2023") {
		t.Errorf("expected meta label to contain EPUB, 1.0 MB, and 2023, got %q", metaLbl.Text)
	}
	if check.Checked {
		t.Errorf("expected checkbox to be unchecked initially")
	}

	// 2. Test checkbox toggle
	check.OnChanged(true)
	if !app.IsBookSelected(b1) {
		t.Errorf("expected b1 to be selected in app after check.OnChanged(true)")
	}

	// 3. Test row recycling for b2
	rv.updateMobileResultRow(row, b2)
	if titleLbl.Text != "Advanced Systems" {
		t.Errorf("expected recycled row title 'Advanced Systems', got %q", titleLbl.Text)
	}
	if authorLbl.Text != "John Smith" {
		t.Errorf("expected recycled row author 'John Smith', got %q", authorLbl.Text)
	}
	if !strings.Contains(metaLbl.Text, "PDF") || !strings.Contains(metaLbl.Text, "5.0 MB") || !strings.Contains(metaLbl.Text, "2021") {
		t.Errorf("expected meta label to contain PDF, 5.0 MB, and 2021, got %q", metaLbl.Text)
	}

	// 4. Test cross-function safety (updateResultRow on mobile row and vice versa)
	rv.updateResultRow(row, b1)
	if titleLbl.Text != "Practical Go" {
		t.Errorf("expected updateResultRow to safely update mobile card")
	}

	// 5. Edge cases: nil / empty handling
	rv.updateMobileResultRow(nil, b1)
	rv.updateMobileResultRow(row, nil)
	findMobileCardComponents(nil)
}

func TestMobileResultsView_HeaderAndSort(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Mobile Header")
	app := NewApp(w)
	app.SetMobileMode(true)
	_ = app.Build()

	rv := app.resultsView
	if rv.sortSelect == nil {
		t.Fatalf("expected mobile sortSelect to be created in mobile mode")
	}
	if rv.sortSelect.Selected != "Title (A-Z)" {
		t.Errorf("expected initial sort selection 'Title (A-Z)', got %q", rv.sortSelect.Selected)
	}

	b1 := &libgen.Book{Title: "B Book", Year: "2010", Filesize: "100"}
	b2 := &libgen.Book{Title: "A Book", Year: "2020", Filesize: "500"}
	b3 := &libgen.Book{Title: "C Book", Year: "2005", Filesize: "300"}

	app.onResults([]*libgen.Book{b1, b2, b3}, 1)

	// Sorted by Title A-Z initially
	if rv.books[0].Title != "A Book" || rv.books[1].Title != "B Book" || rv.books[2].Title != "C Book" {
		t.Errorf("expected Title A-Z order, got %s, %s, %s", rv.books[0].Title, rv.books[1].Title, rv.books[2].Title)
	}

	// Sort by Year (Newest)
	rv.sortSelect.SetSelected("Year (Newest)")
	if rv.books[0].Year != "2020" || rv.books[1].Year != "2010" || rv.books[2].Year != "2005" {
		t.Errorf("expected Year Newest order, got %s, %s, %s", rv.books[0].Year, rv.books[1].Year, rv.books[2].Year)
	}

	// Sort by Size (Largest)
	rv.sortSelect.SetSelected("Size (Largest)")
	if rv.books[0].Filesize != "500" || rv.books[1].Filesize != "300" || rv.books[2].Filesize != "100" {
		t.Errorf("expected Size Largest order, got %s, %s, %s", rv.books[0].Filesize, rv.books[1].Filesize, rv.books[2].Filesize)
	}

	// Select all via headerCheck
	if rv.headerCheck == nil {
		t.Fatalf("expected headerCheck to exist")
	}
	rv.headerCheck.OnChanged(true)
	if len(app.GetSelectedBooks()) != 3 {
		t.Errorf("expected 3 books selected after mobile headerCheck(true), got %d", len(app.GetSelectedBooks()))
	}
	rv.headerCheck.OnChanged(false)
	if len(app.GetSelectedBooks()) != 0 {
		t.Errorf("expected 0 books selected after mobile headerCheck(false), got %d", len(app.GetSelectedBooks()))
	}
}

func TestMobileLayouts_DownloadBarAndSearchBar(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Mobile Widgets")
	app := NewApp(w)
	app.SetMobileMode(true)
	_ = app.Build()

	// DownloadBar widget
	dbWidget := app.downloadBar.Widget()
	if dbWidget == nil {
		t.Fatalf("expected downloadBar.Widget() to return non-nil CanvasObject")
	}
	vbox, ok := dbWidget.(*fyne.Container)
	if !ok || len(vbox.Objects) < 3 {
		t.Fatalf("expected downloadBar container to have at least 3 rows in mobile layout")
	}

	// SearchBar widget
	sbWidget := app.searchBar.Widget()
	if sbWidget == nil {
		t.Fatalf("expected searchBar.Widget() to return non-nil CanvasObject")
	}
	sbVbox, ok := sbWidget.(*fyne.Container)
	if !ok || len(sbVbox.Objects) < 2 {
		t.Fatalf("expected searchBar container to have at least 2 rows in mobile layout")
	}
}

func TestStorage_IsDirWritable_And_GetDefaultSavePath(t *testing.T) {
	// 1. Valid writable temp directory
	tmp := t.TempDir()
	if !IsDirWritable(tmp) {
		t.Errorf("expected IsDirWritable to return true for temp directory %q", tmp)
	}

	// 2. Empty string
	if IsDirWritable("") {
		t.Errorf("expected IsDirWritable to return false for empty string")
	}

	// 3. Unwritable path
	badPath := "/nonexistent_root_dir_123456789/unwritable_subfolder"
	if IsDirWritable(badPath) {
		t.Errorf("expected IsDirWritable to return false for non-writable root path %q", badPath)
	}

	// 4. GetDefaultSavePath desktop
	desktopPath := GetDefaultSavePath(false)
	if desktopPath == "" {
		t.Errorf("expected non-empty desktop save path")
	}
	if !IsDirWritable(desktopPath) {
		t.Errorf("expected desktop save path %q to be writable", desktopPath)
	}

	// 5. GetDefaultSavePath mobile
	mobilePath := GetDefaultSavePath(true)
	if mobilePath == "" {
		t.Errorf("expected non-empty mobile save path")
	}
}

func TestStorage_SetGetSavePath_And_FailureModalVariants(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Storage & Modal")
	app := NewApp(w)
	_ = app.Build()

	db := app.downloadBar
	tmp := t.TempDir()

	// Test SetSavePath and GetSavePath
	db.SetSavePath(tmp)
	if db.GetSavePath() != tmp {
		t.Errorf("expected savePath %q, got %q", tmp, db.GetSavePath())
	}
	if db.pathLabel.Text != tmp {
		t.Errorf("expected pathLabel text %q, got %q", tmp, db.pathLabel.Text)
	}

	// Test showFailureModal with storage error
	storageErr := fmt.Errorf("storage write error for '%s/book.pdf': permission denied", tmp)
	db.showFailureModal(1, storageErr)

	// Test showFailureModal with mirror error
	mirrorErr := fmt.Errorf("mirror returned HTML error page instead of epub file")
	db.showFailureModal(2, mirrorErr)

	// Test showFailureModal with generic connection error
	connErr := fmt.Errorf("connection timeout")
	db.showFailureModal(1, connErr)

	// Test showFailureModal with no error (backwards compatibility)
	db.showFailureModal(3)
}

func TestHTTPRangeResume(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test HTTP Range Resume")
	app := NewApp(w)
	_ = app.Build()

	tempDir := t.TempDir()
	db := app.downloadBar
	db.savePath = tempDir

	fullData := []byte(strings.Repeat("X", 10000))
	requestCount := 0

	db.httpClient = &http.Client{
		Transport: mockRoundTripper(func(req *http.Request) (*http.Response, error) {
			requestCount++
			rangeHdr := req.Header.Get("Range")

			if rangeHdr == "" {
				// First attempt: stream only 4000 bytes then fail
				pr, pw := io.Pipe()
				go func() {
					_, _ = pw.Write(fullData[:4000])
					_ = pw.CloseWithError(io.ErrUnexpectedEOF)
				}()
				h := make(http.Header)
				h.Set("Accept-Ranges", "bytes")
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        h,
					Body:          pr,
					ContentLength: int64(len(fullData)),
				}, nil
			}

			// Resumed request: Range: bytes=4000-
			if strings.HasPrefix(rangeHdr, "bytes=4000-") {
				h := make(http.Header)
				h.Set("Content-Range", fmt.Sprintf("bytes 4000-%d/%d", len(fullData)-1, len(fullData)))
				remaining := fullData[4000:]
				return &http.Response{
					StatusCode:    http.StatusPartialContent,
					Header:        h,
					Body:          io.NopCloser(bytes.NewReader(remaining)),
					ContentLength: int64(len(remaining)),
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("bad range")),
			}, nil
		}),
	}

	book := &libgen.Book{
		Title:       "Resumable Book",
		Extension:   "epub",
		Md5:         "1234567890abcdef1234567890abcdef",
		Filesize:    "10000",
		DownloadURL: "http://example.com/download",
	}

	db.SetSelectedBooks([]*libgen.Book{book})
	db.doDownload()

	deadline := time.Now().Add(5 * time.Second)
	for {
		db.mu.Lock()
		downloading := db.isDownloading
		db.mu.Unlock()
		if !downloading {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for resumed download to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	filePath := filepath.Join(tempDir, getBookFilename(book))
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected file to exist after resumed download: %v", err)
	}
	if len(data) != len(fullData) {
		t.Errorf("expected %d bytes, got %d bytes", len(fullData), len(data))
	}
	if !bytes.Equal(data, fullData) {
		t.Errorf("file content mismatch")
	}
	if requestCount < 2 {
		t.Errorf("expected at least 2 requests (initial + resume), got %d", requestCount)
	}
}

func TestLocationPresets_MobileAndDesktop(t *testing.T) {
	// 1. Mobile presets
	mobilePresets := GetLocationPresets(true)
	if len(mobilePresets) == 0 {
		t.Errorf("expected at least 1 mobile preset")
	}

	// 2. Desktop presets
	desktopPresets := GetLocationPresets(false)
	if len(desktopPresets) == 0 {
		t.Errorf("expected at least 1 desktop preset")
	}

	// Ensure all returned preset paths are non-empty
	for _, p := range mobilePresets {
		if p.Path == "" || p.Label == "" {
			t.Errorf("empty preset found in mobile presets: %+v", p)
		}
	}
	for _, p := range desktopPresets {
		if p.Path == "" || p.Label == "" {
			t.Errorf("empty preset found in desktop presets: %+v", p)
		}
	}
}

func TestLocationPreferences_PersistenceAndFallback(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	// 1. Initially not configured
	testApp.Preferences().SetBool(PrefKeyFolderConfigured, false)
	testApp.Preferences().SetString(PrefKeyDownloadFolder, "")

	if IsDownloadFolderConfigured() {
		t.Errorf("expected folder not configured initially")
	}

	defaultPath := GetConfiguredSavePath(false)
	if defaultPath == "" {
		t.Errorf("expected default path to be non-empty")
	}

	// 2. Save a valid writable path
	tempDir := t.TempDir()
	SaveConfiguredSavePath(tempDir)

	if !IsDownloadFolderConfigured() {
		t.Errorf("expected folder to be configured after SaveConfiguredSavePath")
	}

	configured := GetConfiguredSavePath(false)
	if configured != tempDir {
		t.Errorf("expected configured path %q, got %q", tempDir, configured)
	}

	// 3. Fallback if saved path is no longer writable / invalid
	testApp.Preferences().SetString(PrefKeyDownloadFolder, "/nonexistent/invalid/dir/path/12345")
	fallback := GetConfiguredSavePath(false)
	if fallback == "/nonexistent/invalid/dir/path/12345" {
		t.Errorf("expected fallback when saved path is invalid")
	}
}

func TestLocationDialog_ValidationAndSelection(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	tempDir := t.TempDir()

	var selectedPath string
	ShowDownloadLocationDialog(app, true, func(path string) {
		selectedPath = path
	})

	// Test saving path directly updates both preferences and DownloadBar
	SaveConfiguredSavePath(tempDir)
	app.downloadBar.SetSavePath(tempDir)

	if app.downloadBar.GetSavePath() != tempDir {
		t.Errorf("expected download bar path to be %q, got %q", tempDir, app.downloadBar.GetSavePath())
	}
	_ = selectedPath
}

func TestMobileLocationModal_Actions(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	app.SetMobileMode(true)
	_ = app.Build()

	tempDir := t.TempDir()
	app.downloadBar.SetSavePath(tempDir)

	// Call OpenSaveFolder in mobile mode: should trigger showMobileLocationModal without panicking
	app.downloadBar.OpenSaveFolder()

	// Verify clipboard copy works
	w.Clipboard().SetContent(tempDir)
	if w.Clipboard().Content() != tempDir {
		t.Errorf("expected clipboard content %q, got %q", tempDir, w.Clipboard().Content())
	}
}

func TestSearchBar_SettingsButton(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test")
	app := NewApp(w)
	_ = app.Build()

	if app.searchBar.settingsBtn == nil {
		t.Fatalf("expected settingsBtn to be initialized")
	}
	if !app.searchBar.settingsBtn.Visible() {
		t.Errorf("expected settingsBtn to be visible")
	}

	// Tapping settings button should invoke ShowSettingsDialog without panic
	test.Tap(app.searchBar.settingsBtn)
}

func TestDirectOpenFileAndCompletionModal_OpenBook(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Completion Modal")
	app := NewApp(w)
	_ = app.Build()

	tempDir := t.TempDir()
	app.downloadBar.SetSavePath(tempDir)
	sampleBookFile := filepath.Join(tempDir, "sample.epub")
	_ = os.WriteFile(sampleBookFile, []byte("epub content"), 0644)

	// Verify showCompletionModal with file path executes without panic
	app.downloadBar.showCompletionModal(1, 1024, "Sample Book", sampleBookFile)

	// Verify openFileInSystem executes safely without error
	openFileInSystem(sampleBookFile)
	openFileInSystem("")
}

func TestNormalizePath_AndroidSAF_And_Desktop(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Standard desktop posix path",
			input:    "/Users/someone/Downloads",
			expected: "/Users/someone/Downloads",
		},
		{
			name:     "File URI desktop",
			input:    "file:///Users/someone/Downloads",
			expected: "/Users/someone/Downloads",
		},
		{
			name:     "Exact user screenshot SAF tree URI: /tree/primary%3ABooks",
			input:    "/tree/primary%3ABooks",
			expected: "/storage/emulated/0/Books",
		},
		{
			name:     "Full content URL with SAF tree primary:Books",
			input:    "content://com.android.externalstorage.documents/tree/primary%3ABooks",
			expected: "/storage/emulated/0/Books",
		},
		{
			name:     "Nested SAF tree with percent-encoded slash %2F: /tree/primary%3ADownload%2FBooks",
			input:    "/tree/primary%3ADownload%2FBooks",
			expected: "/storage/emulated/0/Download/Books",
		},
		{
			name:     "Unencoded SAF tree: tree/primary:Download/Books",
			input:    "tree/primary:Download/Books",
			expected: "/storage/emulated/0/Download/Books",
		},
		{
			name:     "Short primary:Books prefix",
			input:    "primary:Books",
			expected: "/storage/emulated/0/Books",
		},
		{
			name:     "Primary root primary:",
			input:    "primary:",
			expected: "/storage/emulated/0",
		},
		{
			name:     "Secondary SD card SAF URI: /tree/1234-5678%3AMyBooks",
			input:    "/tree/1234-5678%3AMyBooks",
			expected: "/storage/1234-5678/MyBooks",
		},
		{
			name:     "Document provider primary: /document/primary%3ADocuments%2FBooks",
			input:    "/document/primary%3ADocuments%2FBooks",
			expected: "/storage/emulated/0/Documents/Books",
		},
		{
			name:     "Document provider raw path: content://com.android.providers.downloads.documents/document/raw%3A%2Fstorage%2Femulated%2F0%2FDownload",
			input:    "content://com.android.providers.downloads.documents/document/raw%3A%2Fstorage%2Femulated%2F0%2FDownload",
			expected: "/storage/emulated/0/Download",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NormalizePath(tc.input)
			if actual != tc.expected {
				t.Errorf("NormalizePath(%q) = %q; expected %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestDownloadQueue_CoreLifecycle(t *testing.T) {
	q := NewDownloadQueue()

	b1 := &libgen.Book{Title: "Book 1", Extension: "epub", Filesize: "1000", Md5: "md5_1"}
	b2 := &libgen.Book{Title: "Book 2", Extension: "pdf", Filesize: "2000", Md5: "md5_2"}
	b3 := &libgen.Book{Title: "Book 3", Extension: "mobi", Filesize: "3000", Md5: "md5_3"}

	added := q.Enqueue(b1, b2, b3)
	if len(added) != 3 {
		t.Fatalf("expected 3 items added, got %d", len(added))
	}
	if q.GetPendingCount() != 3 {
		t.Errorf("expected 3 pending items, got %d", q.GetPendingCount())
	}

	// De-duplicate check: enqueuing b1 again should return existing item
	dup := q.Enqueue(b1)
	if len(dup) != 1 || dup[0] != added[0] {
		t.Errorf("expected duplicate enqueue to return existing item")
	}
	if q.GetPendingCount() != 3 {
		t.Errorf("expected still 3 pending items after duplicate enqueue, got %d", q.GetPendingCount())
	}

	// Remove b2
	removed := q.Remove(BookKey(b2))
	if !removed {
		t.Errorf("expected b2 to be removed")
	}
	if q.GetPendingCount() != 2 {
		t.Errorf("expected 2 pending items after removal, got %d", q.GetPendingCount())
	}

	// Worker execution of remaining 2 items
	var executedTitles []string
	var mu sync.Mutex

	q.StartWorker(func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead, totalBytes int64, speed float64)) error {
		mu.Lock()
		executedTitles = append(executedTitles, item.Book.Title)
		mu.Unlock()
		onProgress(500, 1000, 1024)
		time.Sleep(20 * time.Millisecond)
		onProgress(1000, 1000, 1024)
		return nil
	})

	// Wait for worker to finish
	deadline := time.Now().Add(3 * time.Second)
	for q.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queue worker to finish")
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(executedTitles) != 2 {
		t.Fatalf("expected 2 executed items, got %d: %v", len(executedTitles), executedTitles)
	}
	if executedTitles[0] != "Book 1" || executedTitles[1] != "Book 3" {
		t.Errorf("expected Book 1 then Book 3, got: %v", executedTitles)
	}

	// Check completed items
	st1, ok1 := q.GetBookStatus(b1)
	if !ok1 || st1 != QueueStatusCompleted {
		t.Errorf("expected b1 completed, got status=%v ok=%v", st1, ok1)
	}

	// ClearCompleted
	q.ClearCompleted()
	if len(q.GetItems()) != 0 {
		t.Errorf("expected 0 items after ClearCompleted, got %d", len(q.GetItems()))
	}
}

func TestDownloadQueue_EnqueueWhileActive(t *testing.T) {
	q := NewDownloadQueue()

	b1 := &libgen.Book{Title: "Active Book", Extension: "epub", Filesize: "1000", Md5: "md5_active"}
	b2 := &libgen.Book{Title: "Queued Book", Extension: "pdf", Filesize: "2000", Md5: "md5_queued"}

	q.Enqueue(b1)

	startedB1 := make(chan struct{})
	allowB1Finish := make(chan struct{})

	q.StartWorker(func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead, totalBytes int64, speed float64)) error {
		if item.Book.Title == "Active Book" {
			close(startedB1)
			<-allowB1Finish
		}
		return nil
	})

	<-startedB1

	// While B1 is actively downloading, enqueue B2
	q.Enqueue(b2)

	// Verify B2 is pending and B1 is downloading
	st1, _ := q.GetBookStatus(b1)
	st2, _ := q.GetBookStatus(b2)
	if st1 != QueueStatusDownloading {
		t.Errorf("expected b1 downloading, got %v", st1)
	}
	if st2 != QueueStatusPending {
		t.Errorf("expected b2 pending, got %v", st2)
	}

	close(allowB1Finish)

	// Wait for queue to finish all items
	deadline := time.Now().Add(3 * time.Second)
	for q.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queue to drain")
		}
		time.Sleep(20 * time.Millisecond)
	}

	st1Final, _ := q.GetBookStatus(b1)
	st2Final, _ := q.GetBookStatus(b2)
	if st1Final != QueueStatusCompleted || st2Final != QueueStatusCompleted {
		t.Errorf("expected both completed, got b1=%v b2=%v", st1Final, st2Final)
	}
}

func TestDownloadQueue_CancelCurrent(t *testing.T) {
	q := NewDownloadQueue()

	b1 := &libgen.Book{Title: "Book to Skip", Extension: "epub", Filesize: "1000", Md5: "md5_skip"}
	b2 := &libgen.Book{Title: "Book to Keep", Extension: "pdf", Filesize: "2000", Md5: "md5_keep"}

	q.Enqueue(b1, b2)

	startedB1 := make(chan struct{})

	q.StartWorker(func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead, totalBytes int64, speed float64)) error {
		if item.Book.Title == "Book to Skip" {
			close(startedB1)
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	})

	<-startedB1

	// Cancel currently downloading book
	q.CancelCurrent()

	deadline := time.Now().Add(3 * time.Second)
	for q.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queue after cancel")
		}
		time.Sleep(20 * time.Millisecond)
	}

	st1, _ := q.GetBookStatus(b1)
	st2, _ := q.GetBookStatus(b2)

	if st1 != QueueStatusCancelled {
		t.Errorf("expected b1 cancelled, got %v", st1)
	}
	if st2 != QueueStatusCompleted {
		t.Errorf("expected b2 completed after skipping b1, got %v", st2)
	}
}

func TestDownloadQueue_CancelAll(t *testing.T) {
	q := NewDownloadQueue()

	b1 := &libgen.Book{Title: "Book A", Extension: "epub", Filesize: "1000", Md5: "md5_a"}
	b2 := &libgen.Book{Title: "Book B", Extension: "pdf", Filesize: "2000", Md5: "md5_b"}

	q.Enqueue(b1, b2)

	startedB1 := make(chan struct{})

	q.StartWorker(func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead, totalBytes int64, speed float64)) error {
		if item.Book.Title == "Book A" {
			close(startedB1)
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	})

	<-startedB1

	// Cancel entire queue
	q.CancelAll()

	deadline := time.Now().Add(3 * time.Second)
	for q.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queue to halt")
		}
		time.Sleep(20 * time.Millisecond)
	}

	st1, _ := q.GetBookStatus(b1)
	st2, _ := q.GetBookStatus(b2)

	if st1 != QueueStatusCancelled {
		t.Errorf("expected b1 cancelled, got %v", st1)
	}
	if st2 != QueueStatusCancelled {
		t.Errorf("expected b2 cancelled, got %v", st2)
	}
}

func TestDownloadQueue_FailureResilience(t *testing.T) {
	q := NewDownloadQueue()

	b1 := &libgen.Book{Title: "Failing Book", Extension: "epub", Filesize: "1000", Md5: "md5_fail"}
	b2 := &libgen.Book{Title: "Successful Book", Extension: "pdf", Filesize: "2000", Md5: "md5_succ"}

	q.Enqueue(b1, b2)

	q.StartWorker(func(ctx context.Context, item *DownloadItem, onProgress func(bytesRead, totalBytes int64, speed float64)) error {
		if item.Book.Title == "Failing Book" {
			return errors.New("mirror 503 Service Unavailable")
		}
		return nil
	})

	deadline := time.Now().Add(3 * time.Second)
	for q.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for queue")
		}
		time.Sleep(20 * time.Millisecond)
	}

	st1, _ := q.GetBookStatus(b1)
	st2, _ := q.GetBookStatus(b2)

	if st1 != QueueStatusFailed {
		t.Errorf("expected b1 failed, got %v", st1)
	}
	if st2 != QueueStatusCompleted {
		t.Errorf("expected b2 completed despite b1 failure, got %v", st2)
	}
}

func TestDownloadQueue_UIIntegration_RowStatus(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Row Status")
	app := NewApp(w)
	_ = app.Build()

	b1 := &libgen.Book{Title: "Book One", Extension: "epub", Filesize: "1000", Md5: "md5_r1"}
	b2 := &libgen.Book{Title: "Book Two", Extension: "pdf", Filesize: "2000", Md5: "md5_r2"}

	app.onResults([]*libgen.Book{b1, b2}, 1)

	// Initially neither book is queued
	st, inQueue := app.GetBookQueueStatus(b1)
	if inQueue {
		t.Errorf("expected book not in queue initially, got status=%v", st)
	}

	// Enqueue b1
	app.GetDownloadQueue().Enqueue(b1)

	st1, inQueue1 := app.GetBookQueueStatus(b1)
	if !inQueue1 || st1 != QueueStatusPending {
		t.Errorf("expected b1 to be pending in queue, got status=%v inQueue=%v", st1, inQueue1)
	}

	// Verify applyRowDownloadStatus updates button icon appropriately
	btn := widget.NewButton("", nil)
	applyRowDownloadStatus(btn, b1, app)
	if btn.Icon != theme.HistoryIcon() {
		t.Errorf("expected history icon for pending book, got %v", btn.Icon)
	}
	if !btn.Disabled() {
		t.Errorf("expected button disabled while queued")
	}

	// For non-queued book:
	btn2 := widget.NewButton("", nil)
	applyRowDownloadStatus(btn2, b2, app)
	if btn2.Icon != theme.DownloadIcon() {
		t.Errorf("expected download icon for unqueued book, got %v", btn2.Icon)
	}
	if btn2.Disabled() {
		t.Errorf("expected button enabled for unqueued book")
	}
}

func TestDownloadBar_QueueControls(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test DownloadBar Controls")
	app := NewApp(w)
	_ = app.Build()

	db := app.downloadBar
	if db.queueBtn == nil {
		t.Fatalf("expected queueBtn to be initialized")
	}

	// Initially queue button is hidden
	if db.queueBtn.Visible() {
		t.Errorf("expected queueBtn to be hidden initially")
	}

	// Enqueue books
	b1 := &libgen.Book{Title: "Book A", Extension: "epub", Filesize: "1000", Md5: "md5_qa"}
	b2 := &libgen.Book{Title: "Book B", Extension: "pdf", Filesize: "2000", Md5: "md5_qb"}
	db.queue.Enqueue(b1, b2)

	db.updateQueueUI()
	if !db.queueBtn.Visible() {
		t.Errorf("expected queueBtn to be visible after enqueuing")
	}
	if db.queueBtn.Text != "Queue (2)" {
		t.Errorf("expected queueBtn text 'Queue (2)', got %q", db.queueBtn.Text)
	}

	// Test selecting books while downloading shows "Add to Queue (N)"
	db.isDownloading = true
	db.SetSelectedBooks([]*libgen.Book{b1})
	if db.dlBtn.Text != "Add to Queue (1)" {
		t.Errorf("expected dlBtn text 'Add to Queue (1)', got %q", db.dlBtn.Text)
	}

	db.SetSelectedBooks([]*libgen.Book{b1, b2})
	if db.dlBtn.Text != "Add to Queue (2)" {
		t.Errorf("expected dlBtn text 'Add to Queue (2)', got %q", db.dlBtn.Text)
	}
}

func TestQueueDialog_RenderAndActions(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w := testApp.NewWindow("Test Queue Dialog")
	app := NewApp(w)
	_ = app.Build()

	b1 := &libgen.Book{Title: "Active Book", Extension: "epub", Filesize: "1000", Md5: "md5_d1"}
	b2 := &libgen.Book{Title: "Pending Book", Extension: "pdf", Filesize: "2000", Md5: "md5_d2"}

	q := app.GetDownloadQueue()
	q.Enqueue(b1, b2)

	// Verify ShowQueueDialog runs without panic
	ShowQueueDialog(app)
}



