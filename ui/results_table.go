package ui

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ciehanski/libgen-cli/libgen"
)

// ResultsView displays the search results as a scrollable, selectable list
// with checkboxes for multi-book selection and sortable columns.
type ResultsView struct {
	app   *App
	books []*libgen.Book
	list  *widget.List
	empty *widget.Label
	wrap  *fyne.Container

	headerCheck *widget.Check
	header      *fyne.Container
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	pageLbl     *widget.Label
	main        *fyne.Container

	sortCol    string
	sortAsc    bool
	hasMore    bool
	sortSelect *widget.Select
}

var mobileSortOptions = []string{
	"Title (A-Z)",
	"Title (Z-A)",
	"Year (Newest)",
	"Year (Oldest)",
	"Size (Largest)",
	"Size (Smallest)",
	"Author (A-Z)",
	"Format (A-Z)",
}

func (rv *ResultsView) isMobile() bool {
	if rv.app != nil {
		return rv.app.IsMobile()
	}
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return true
	}
	return fyne.CurrentDevice().IsMobile()
}

func NewResultsView(a *App) *ResultsView {
	rv := &ResultsView{
		app:     a,
		sortCol: "Title",
		sortAsc: true,
	}

	rv.empty = widget.NewLabelWithStyle(
		"Search for a book above to see results.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	rv.list = widget.NewList(
		func() int { return len(rv.books) },
		func() fyne.CanvasObject {
			if rv.isMobile() {
				return newMobileResultRow()
			}
			return newResultRow()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(rv.books) {
				return
			}
			if rv.isMobile() {
				rv.updateMobileResultRow(obj, rv.books[id])
			} else {
				rv.updateResultRow(obj, rv.books[id])
			}
		},
	)

	// Clicking anywhere on a row toggles its selection
	rv.list.OnSelected = func(id widget.ListItemID) {
		if rv.app == nil || id < 0 || int(id) >= len(rv.books) {
			return
		}
		b := rv.books[id]
		rv.app.ToggleBookSelection(b)
		rv.list.UnselectAll()
		rv.list.Refresh()
	}

	// Header select-all checkbox
	rv.headerCheck = widget.NewCheck("", func(checked bool) {
		if rv.app == nil {
			return
		}
		if checked {
			rv.app.SelectAllCurrentPage()
		} else {
			rv.app.ClearCurrentPage()
		}
	})

	if rv.isMobile() {
		rv.sortSelect = widget.NewSelect(mobileSortOptions, func(selected string) {
			switch selected {
			case "Title (A-Z)":
				rv.sortCol = "Title"
				rv.sortAsc = true
			case "Title (Z-A)":
				rv.sortCol = "Title"
				rv.sortAsc = false
			case "Year (Newest)":
				rv.sortCol = "Year"
				rv.sortAsc = false
			case "Year (Oldest)":
				rv.sortCol = "Year"
				rv.sortAsc = true
			case "Size (Largest)":
				rv.sortCol = "Size"
				rv.sortAsc = false
			case "Size (Smallest)":
				rv.sortCol = "Size"
				rv.sortAsc = true
			case "Author (A-Z)":
				rv.sortCol = "Author"
				rv.sortAsc = true
			case "Format (A-Z)":
				rv.sortCol = "Fmt"
				rv.sortAsc = true
			}
			rv.doSort()
		})
		rv.sortSelect.SetSelected("Title (A-Z)")

		selectAllLbl := widget.NewLabelWithStyle("Select All", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		selectAllBox := container.NewHBox(rv.headerCheck, selectAllLbl)
		rv.header = container.NewBorder(nil, nil, selectAllBox, rv.sortSelect)
	} else {
		titleBtn := widget.NewButton("Title ▲", func() { rv.sortBy("Title") })
		titleBtn.Importance = widget.LowImportance
		titleBtn.Alignment = widget.ButtonAlignLeading

		authorBtn := widget.NewButton("Author", func() { rv.sortBy("Author") })
		authorBtn.Importance = widget.LowImportance
		authorBtn.Alignment = widget.ButtonAlignLeading

		yearBtn := widget.NewButton("Year", func() { rv.sortBy("Year") })
		yearBtn.Importance = widget.LowImportance

		fmtBtn := widget.NewButton("Fmt", func() { rv.sortBy("Fmt") })
		fmtBtn.Importance = widget.LowImportance

		sizeBtn := widget.NewButton("Size", func() { rv.sortBy("Size") })
		sizeBtn.Importance = widget.LowImportance
		sizeBtn.Alignment = widget.ButtonAlignTrailing

		headerGrid := container.New(&tableRowLayout{},
			titleBtn, authorBtn, yearBtn, fmtBtn, sizeBtn, widget.NewLabel(""),
		)

		rv.header = container.NewBorder(nil, nil, rv.headerCheck, nil, headerGrid)
	}

	rv.prevBtn = widget.NewButtonWithIcon("Prev", theme.NavigateBackIcon(), func() {
		if a.currentPage > 1 {
			a.searchBar.doSearchPage(a.currentPage - 1)
		}
	})
	rv.prevBtn.Importance = widget.LowImportance
	rv.prevBtn.Disable()

	rv.nextBtn = widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		a.searchBar.doSearchPage(a.currentPage + 1)
	})
	rv.nextBtn.Importance = widget.LowImportance
	rv.nextBtn.Disable()

	rv.pageLbl = widget.NewLabelWithStyle("Page 1", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	footer := container.NewCenter(
		container.NewHBox(
			rv.prevBtn,
			container.NewPadded(rv.pageLbl),
			rv.nextBtn,
		),
	)

	rv.wrap = container.NewStack(rv.empty)
	rv.main = container.NewBorder(rv.header, footer, nil, nil, rv.wrap)
	return rv
}

func (rv *ResultsView) Widget() fyne.CanvasObject {
	return rv.main
}

func (rv *ResultsView) SetLoading(loading bool) {
	if loading {
		rv.prevBtn.Disable()
		rv.nextBtn.Disable()
	} else {
		rv.UpdatePaginationButtons()
	}
}

func (rv *ResultsView) UpdatePaginationButtons() {
	if rv.app == nil {
		return
	}
	if rv.app.currentPage <= 1 {
		rv.prevBtn.Disable()
	} else {
		rv.prevBtn.Enable()
	}
	if len(rv.books) == 0 || !rv.hasMore {
		rv.nextBtn.Disable()
	} else {
		rv.nextBtn.Enable()
	}
	rv.pageLbl.SetText(fmt.Sprintf("Page %d", rv.app.currentPage))
}

func (rv *ResultsView) DisableNext() {
	rv.hasMore = false
	rv.nextBtn.Disable()
}

func (rv *ResultsView) SetBooks(books []*libgen.Book) {
	rv.books = books
	rv.hasMore = len(books) >= 25
	rv.UpdatePaginationButtons()
	rv.doSort()
	rv.UpdateHeaderSelectionState()
}

func (rv *ResultsView) RefreshList() {
	if rv.list != nil {
		rv.list.Refresh()
	}
	rv.UpdateHeaderSelectionState()
}

func (rv *ResultsView) UpdateHeaderSelectionState() {
	if rv.headerCheck == nil || rv.app == nil {
		return
	}
	if len(rv.books) == 0 {
		rv.headerCheck.OnChanged = nil
		rv.headerCheck.SetChecked(false)
		rv.headerCheck.OnChanged = func(checked bool) {
			if checked {
				rv.app.SelectAllCurrentPage()
			} else {
				rv.app.ClearCurrentPage()
			}
		}
		return
	}

	allSelected := true
	for _, b := range rv.books {
		if !rv.app.IsBookSelected(b) {
			allSelected = false
			break
		}
	}

	rv.headerCheck.OnChanged = nil
	rv.headerCheck.SetChecked(allSelected)
	rv.headerCheck.OnChanged = func(checked bool) {
		if checked {
			rv.app.SelectAllCurrentPage()
		} else {
			rv.app.ClearCurrentPage()
		}
	}
}

func (rv *ResultsView) sortBy(col string) {
	if rv.sortCol == col {
		rv.sortAsc = !rv.sortAsc
	} else {
		rv.sortCol = col
		rv.sortAsc = true
	}
	rv.updateHeader()
	rv.doSort()
}

func (rv *ResultsView) syncSortSelect() {
	if rv.sortSelect == nil {
		return
	}
	var text string
	switch rv.sortCol {
	case "Title":
		if rv.sortAsc {
			text = "Title (A-Z)"
		} else {
			text = "Title (Z-A)"
		}
	case "Author":
		if rv.sortAsc {
			text = "Author (A-Z)"
		} else {
			text = "Author (Z-A)"
		}
	case "Year":
		if rv.sortAsc {
			text = "Year (Oldest)"
		} else {
			text = "Year (Newest)"
		}
	case "Fmt":
		if rv.sortAsc {
			text = "Format (A-Z)"
		} else {
			text = "Format (Z-A)"
		}
	case "Size":
		if rv.sortAsc {
			text = "Size (Smallest)"
		} else {
			text = "Size (Largest)"
		}
	}
	if text != "" && rv.sortSelect.Selected != text {
		rv.sortSelect.Selected = text
		rv.sortSelect.Refresh()
	}
}

func (rv *ResultsView) updateHeader() {
	if rv.isMobile() {
		rv.syncSortSelect()
		return
	}
	if rv.header == nil {
		return
	}
	var grid *fyne.Container
	for _, o := range rv.header.Objects {
		if g, ok := o.(*fyne.Container); ok {
			if grid == nil || len(g.Objects) >= 5 {
				grid = g
			}
		}
	}
	if grid == nil {
		return
	}
	labels := []string{"Title", "Author", "Year", "Fmt", "Size"}
	for i, lbl := range labels {
		if i >= len(grid.Objects) {
			break
		}
		btn, ok := grid.Objects[i].(*widget.Button)
		if !ok {
			continue
		}
		if rv.sortCol == lbl {
			if rv.sortAsc {
				btn.SetText(lbl + " ▲")
			} else {
				btn.SetText(lbl + " ▼")
			}
		} else {
			btn.SetText(lbl)
		}
	}
}

func (rv *ResultsView) doSort() {
	if len(rv.books) == 0 {
		if rv.wrap != nil {
			rv.wrap.Objects = []fyne.CanvasObject{rv.empty}
			rv.wrap.Refresh()
		}
		return
	}

	sort.SliceStable(rv.books, func(i, j int) bool {
		a, b := rv.books[i], rv.books[j]
		if a == nil || b == nil {
			return a != nil
		}
		var less bool
		switch rv.sortCol {
		case "Title":
			less = strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "Author":
			less = strings.ToLower(a.Author) < strings.ToLower(b.Author)
		case "Year":
			less = a.Year < b.Year
		case "Fmt":
			less = strings.ToLower(a.Extension) < strings.ToLower(b.Extension)
		case "Size":
			var as, bs int64
			fmt.Sscan(a.Filesize, &as)
			fmt.Sscan(b.Filesize, &bs)
			less = as < bs
		}
		if !rv.sortAsc {
			return !less
		}
		return less
	})

	if rv.wrap != nil && rv.list != nil {
		rv.wrap.Objects = []fyne.CanvasObject{rv.list}
		rv.list.UnselectAll()
		rv.list.Refresh()
		rv.wrap.Refresh()
	}
}

// ─── Row template & custom column layout ─────────────────────────────────────

// tableRowLayout positions the 5 columns (Title, Author, Year, Fmt, Size):
// Title (65% flex) and Author (35% flex) receive maximum space for readability,
// while Year (70px), Fmt (50px), and Size (80px) remain compact and fixed.
type tableRowLayout struct{}

func (l *tableRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 5 {
		return
	}
	title := objects[0]
	author := objects[1]
	year := objects[2]
	ext := objects[3]
	sizeObj := objects[4]

	pad := theme.Padding()
	yearWidth := float32(60)
	extWidth := float32(48)
	sizeWidth := float32(72)
	dlWidth := float32(36)

	hasDl := len(objects) >= 6
	rightColsWidth := yearWidth + extWidth + sizeWidth + pad*2
	if hasDl {
		rightColsWidth += dlWidth + pad
	}

	h := size.Height

	// Right columns: Year, Fmt, Size, [Download button]
	rightStart := size.Width - rightColsWidth
	if rightStart < 140 {
		rightStart = 140
	}

	year.Move(fyne.NewPos(rightStart, 0))
	year.Resize(fyne.NewSize(yearWidth, h))

	ext.Move(fyne.NewPos(rightStart+yearWidth+pad, 0))
	ext.Resize(fyne.NewSize(extWidth, h))

	sizeObj.Move(fyne.NewPos(rightStart+yearWidth+pad+extWidth+pad, 0))
	sizeObj.Resize(fyne.NewSize(sizeWidth, h))

	if hasDl {
		dlObj := objects[5]
		dlObj.Move(fyne.NewPos(rightStart+yearWidth+pad+extWidth+pad+sizeWidth+pad, 0))
		dlObj.Resize(fyne.NewSize(dlWidth, h))
	}

	// Middle space for Title & Author
	midWidth := rightStart - pad
	if midWidth < 80 {
		midWidth = 80
	}

	titleWidth := midWidth * 0.65
	authorWidth := midWidth - titleWidth - pad

	title.Move(fyne.NewPos(0, 0))
	title.Resize(fyne.NewSize(titleWidth, h))

	author.Move(fyne.NewPos(titleWidth+pad, 0))
	author.Resize(fyne.NewSize(authorWidth, h))
}

func (l *tableRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(320, 32)
}

func newResultRow() fyne.CanvasObject {
	check := widget.NewCheck("", nil)

	title := widget.NewLabel("Title")
	title.Truncation = fyne.TextTruncateEllipsis

	author := widget.NewLabel("Author")
	author.Truncation = fyne.TextTruncateEllipsis
	author.TextStyle = fyne.TextStyle{Italic: true}

	year := widget.NewLabel("2000")
	year.Alignment = fyne.TextAlignCenter

	ext := widget.NewLabel("epub")
	ext.Alignment = fyne.TextAlignCenter
	ext.TextStyle = fyne.TextStyle{Bold: true}

	size := widget.NewLabel("1.0 MB")
	size.Alignment = fyne.TextAlignTrailing

	dlBtn := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	dlBtn.Importance = widget.LowImportance

	grid := container.New(&tableRowLayout{},
		title, author, year, ext, size, dlBtn,
	)
	row := container.NewBorder(nil, nil, check, nil, grid)
	return row
}

func newMobileResultRow() fyne.CanvasObject {
	check := widget.NewCheck("", nil)

	title := widget.NewLabel("Title")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Truncation = fyne.TextTruncateEllipsis

	author := widget.NewLabel("Author")
	author.TextStyle = fyne.TextStyle{Italic: true}
	author.Truncation = fyne.TextTruncateEllipsis

	meta := widget.NewLabel("PDF • 1.0 MB • 2020")
	meta.Importance = widget.MediumImportance
	meta.Truncation = fyne.TextTruncateEllipsis

	dlBtn := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	dlBtn.Importance = widget.LowImportance

	textStack := container.NewVBox(title, author, meta)

	centerCheck := container.NewCenter(check)
	centerDl := container.NewCenter(dlBtn)

	content := container.NewBorder(nil, nil, centerCheck, centerDl, textStack)
	return container.NewPadded(content)
}

func findMobileCardComponents(obj fyne.CanvasObject) (*widget.Check, *widget.Button, *widget.Label, *widget.Label, *widget.Label) {
	if obj == nil {
		return nil, nil, nil, nil, nil
	}

	var check *widget.Check
	var dlBtn *widget.Button
	var labels []*widget.Label

	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		switch val := o.(type) {
		case *widget.Check:
			if check == nil {
				check = val
			}
		case *widget.Button:
			if dlBtn == nil {
				dlBtn = val
			}
		case *widget.Label:
			labels = append(labels, val)
		case *fyne.Container:
			for _, child := range val.Objects {
				walk(child)
			}
		}
	}

	walk(obj)

	var titleLbl, authorLbl, metaLbl *widget.Label
	if len(labels) >= 1 {
		titleLbl = labels[0]
	}
	if len(labels) >= 2 {
		authorLbl = labels[1]
	}
	if len(labels) >= 3 {
		metaLbl = labels[2]
	}

	return check, dlBtn, titleLbl, authorLbl, metaLbl
}

func applyRowDownloadStatus(dlBtn *widget.Button, book *libgen.Book, app *App) {
	if dlBtn == nil || book == nil {
		return
	}
	b := book
	dlBtn.OnTapped = func() {
		if app != nil {
			app.DownloadSingleBook(b)
		}
	}

	if app != nil {
		status, inQueue := app.GetBookQueueStatus(book)
		if inQueue {
			switch status {
			case QueueStatusDownloading:
				dlBtn.SetIcon(theme.ViewRefreshIcon())
				dlBtn.Importance = widget.HighImportance
				dlBtn.Disable()
				return
			case QueueStatusPending:
				dlBtn.SetIcon(theme.HistoryIcon())
				dlBtn.Importance = widget.MediumImportance
				dlBtn.Disable()
				return
			case QueueStatusCompleted:
				dlBtn.SetIcon(theme.ConfirmIcon())
				dlBtn.Importance = widget.LowImportance
				dlBtn.Disable()
				return
			case QueueStatusFailed:
				dlBtn.SetIcon(theme.WarningIcon())
				dlBtn.Importance = widget.DangerImportance
				dlBtn.Enable()
				return
			}
		}
	}

	dlBtn.SetIcon(theme.DownloadIcon())
	dlBtn.Importance = widget.LowImportance
	dlBtn.Enable()
}

func (rv *ResultsView) updateMobileResultRow(obj fyne.CanvasObject, book *libgen.Book) {
	if obj == nil || book == nil {
		return
	}
	check, dlBtn, titleLbl, authorLbl, metaLbl := findMobileCardComponents(obj)
	if titleLbl == nil || check == nil {
		return
	}

	titleLbl.SetText(book.Title)
	if authorLbl != nil {
		authorText := book.Author
		if authorText == "" {
			authorText = "Unknown Author"
		}
		authorLbl.SetText(authorText)
	}
	if metaLbl != nil {
		ext := strings.ToUpper(strings.TrimSpace(book.Extension))
		if ext == "" {
			ext = "FILE"
		}
		sz := formatSize(book.Filesize)
		yr := strings.TrimSpace(book.Year)
		if yr != "" {
			metaLbl.SetText(fmt.Sprintf("%s • %s • %s", ext, sz, yr))
		} else {
			metaLbl.SetText(fmt.Sprintf("%s • %s", ext, sz))
		}
	}

	if dlBtn != nil {
		applyRowDownloadStatus(dlBtn, book, rv.app)
	}

	check.OnChanged = nil
	if rv.app != nil {
		check.SetChecked(rv.app.IsBookSelected(book))
		check.OnChanged = func(checked bool) {
			rv.app.SetBookSelected(book, checked)
			rv.UpdateHeaderSelectionState()
		}
	} else {
		check.SetChecked(false)
	}
}

func (rv *ResultsView) updateResultRow(obj fyne.CanvasObject, book *libgen.Book) {
	if obj == nil || book == nil {
		return
	}
	row, ok := obj.(*fyne.Container)
	if !ok {
		return
	}

	var check *widget.Check
	var grid *fyne.Container
	for _, o := range row.Objects {
		if c, ok := o.(*widget.Check); ok {
			check = c
		} else if g, ok := o.(*fyne.Container); ok {
			if grid == nil || len(g.Objects) >= 5 {
				grid = g
			}
		}
	}
	if check == nil || grid == nil || len(grid.Objects) < 5 {
		rv.updateMobileResultRow(obj, book)
		return
	}

	cells := grid.Objects
	if titleLbl, ok := cells[0].(*widget.Label); ok {
		titleLbl.SetText(book.Title)
	}
	if authorLbl, ok := cells[1].(*widget.Label); ok {
		authorLbl.SetText(book.Author)
	}
	if yearLbl, ok := cells[2].(*widget.Label); ok {
		yearLbl.SetText(book.Year)
	}
	if extLbl, ok := cells[3].(*widget.Label); ok {
		ext := strings.ToLower(book.Extension)
		extLbl.SetText(ext)
		switch ext {
		case "epub", "pdf", "mobi":
			extLbl.TextStyle = fyne.TextStyle{Bold: true}
		default:
			extLbl.TextStyle = fyne.TextStyle{}
		}
		extLbl.Refresh()
	}
	if sizeLbl, ok := cells[4].(*widget.Label); ok {
		sizeLbl.SetText(formatSize(book.Filesize))
	}

	if len(cells) >= 6 {
		if dlBtn, ok := cells[5].(*widget.Button); ok {
			applyRowDownloadStatus(dlBtn, book, rv.app)
		}
	}

	// Update check without triggering event
	check.OnChanged = nil
	if rv.app != nil {
		check.SetChecked(rv.app.IsBookSelected(book))
		check.OnChanged = func(checked bool) {
			rv.app.SetBookSelected(book, checked)
			rv.UpdateHeaderSelectionState()
		}
	} else {
		check.SetChecked(false)
	}
}

func formatSize(raw string) string {
	if raw == "" || raw == "0" {
		return "—"
	}
	var bytes int64
	if _, err := fmt.Sscan(raw, &bytes); err != nil || bytes <= 0 {
		return raw
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := "KMGTPE"
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), units[exp])
}

// GetList returns the results list widget.
func (rv *ResultsView) GetList() *widget.List {
	return rv.list
}

// GetEmptyLabel returns the empty state label.
func (rv *ResultsView) GetEmptyLabel() *widget.Label {
	return rv.empty
}

// GetHeaderCheck returns the header select-all check widget.
func (rv *ResultsView) GetHeaderCheck() *widget.Check {
	return rv.headerCheck
}

// GetPrevButton returns the previous page button.
func (rv *ResultsView) GetPrevButton() *widget.Button {
	return rv.prevBtn
}

// GetNextButton returns the next page button.
func (rv *ResultsView) GetNextButton() *widget.Button {
	return rv.nextBtn
}

// GetPageLabel returns the pagination label.
func (rv *ResultsView) GetPageLabel() *widget.Label {
	return rv.pageLbl
}

// GetSortSelect returns the mobile sort dropdown widget.
func (rv *ResultsView) GetSortSelect() *widget.Select {
	return rv.sortSelect
}

// GetBooks returns the currently loaded book slice.
func (rv *ResultsView) GetBooks() []*libgen.Book {
	return rv.books
}
