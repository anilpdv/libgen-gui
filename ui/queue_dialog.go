package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowQueueDialog displays a responsive, interactive dialog showing the active download,
// upcoming queued items, and recent session downloads.
func ShowQueueDialog(app *App) {
	if app == nil || app.window == nil {
		return
	}
	q := app.GetDownloadQueue()
	if q == nil {
		return
	}

	var d *dialog.CustomDialog

	contentBox := container.NewVBox()

	refreshContent := func() {
		contentBox.Objects = nil

		items := q.GetItems()
		active := q.GetActiveItem()

		// ─── 1. Active Download Section ─────────────────────────────
		activeHeader := widget.NewLabelWithStyle("Currently Downloading", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		contentBox.Add(activeHeader)

		if active != nil {
			titleLbl := widget.NewLabel(active.Book.Title)
			titleLbl.Truncation = fyne.TextTruncateEllipsis
			titleLbl.TextStyle = fyne.TextStyle{Bold: true}

			metaText := fmt.Sprintf("%s • %s", strings.ToUpper(active.Book.Extension), formatSize(active.Book.Filesize))
			if active.Speed > 0 {
				metaText += fmt.Sprintf(" • %s/s", formatBytes(int64(active.Speed)))
			}
			metaLbl := widget.NewLabel(metaText)
			metaLbl.Importance = widget.MediumImportance

			progBar := widget.NewProgressBar()
			progBar.SetValue(active.Progress)

			cancelActiveBtn := widget.NewButtonWithIcon("Skip/Cancel", theme.CancelIcon(), func() {
				q.CancelCurrent()
			})
			cancelActiveBtn.Importance = widget.MediumImportance

			activeRow := container.NewBorder(
				nil, nil,
				nil, cancelActiveBtn,
				container.NewVBox(titleLbl, metaLbl, progBar),
			)
			contentBox.Add(container.NewPadded(activeRow))
		} else {
			idleLbl := widget.NewLabelWithStyle("No active download in progress.", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
			idleLbl.Importance = widget.LowImportance
			contentBox.Add(container.NewPadded(idleLbl))
		}

		contentBox.Add(widget.NewSeparator())

		// ─── 2. Pending Queue Section ──────────────────────────────
		var pendingItems []*DownloadItem
		for _, it := range items {
			if it.Status == QueueStatusPending {
				pendingItems = append(pendingItems, it)
			}
		}

		queueHeader := widget.NewLabelWithStyle(fmt.Sprintf("Up Next (%d in queue)", len(pendingItems)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		contentBox.Add(queueHeader)

		if len(pendingItems) > 0 {
			queueList := container.NewVBox()
			for idx, item := range pendingItems {
				it := item
				pos := idx + 1
				lbl := widget.NewLabel(fmt.Sprintf("%d. %s [%s • %s]", pos, truncate(it.Book.Title, 45), strings.ToUpper(it.Book.Extension), formatSize(it.Book.Filesize)))
				lbl.Truncation = fyne.TextTruncateEllipsis

				removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
					q.Remove(BookKey(it.Book))
				})
				removeBtn.Importance = widget.LowImportance

				row := container.NewBorder(nil, nil, nil, removeBtn, lbl)
				queueList.Add(row)
			}
			contentBox.Add(queueList)
		} else {
			emptyQueueLbl := widget.NewLabelWithStyle("Queue is empty. Select books or tap download to add more.", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
			emptyQueueLbl.Importance = widget.LowImportance
			contentBox.Add(container.NewPadded(emptyQueueLbl))
		}

		// ─── 3. Completed / History Section ────────────────────────
		var completedItems []*DownloadItem
		for _, it := range items {
			if it.Status == QueueStatusCompleted || it.Status == QueueStatusFailed || it.Status == QueueStatusCancelled {
				completedItems = append(completedItems, it)
			}
		}

		if len(completedItems) > 0 {
			contentBox.Add(widget.NewSeparator())
			historyHeader := container.NewBorder(
				nil, nil,
				widget.NewLabelWithStyle("Session History", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewButtonWithIcon("Clear History", theme.DeleteIcon(), func() {
					q.ClearCompleted()
				}),
			)
			contentBox.Add(historyHeader)

			historyList := container.NewVBox()
			for _, item := range completedItems {
				it := item
				var statusIcon fyne.Resource
				var statusImportance widget.Importance
				statusText := it.Status.String()

				switch it.Status {
				case QueueStatusCompleted:
					statusIcon = theme.ConfirmIcon()
					statusImportance = widget.SuccessImportance
				case QueueStatusFailed:
					statusIcon = theme.ErrorIcon()
					statusImportance = widget.DangerImportance
				case QueueStatusCancelled:
					statusIcon = theme.CancelIcon()
					statusImportance = widget.WarningImportance
				}

				rowLbl := widget.NewLabel(fmt.Sprintf("%s — %s", statusText, truncate(it.Book.Title, 40)))
				rowLbl.Importance = statusImportance
				rowLbl.Truncation = fyne.TextTruncateEllipsis

				var actionBtn fyne.CanvasObject
				if it.Status == QueueStatusCompleted && app.downloadBar != nil {
					savePath := app.downloadBar.GetSavePath()
					filePath := filepath.Join(savePath, getBookFilename(it.Book))
					openBtn := widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), func() {
						openFileInSystem(filePath)
					})
					openBtn.Importance = widget.LowImportance
					actionBtn = openBtn
				} else {
					actionBtn = widget.NewIcon(statusIcon)
				}

				row := container.NewBorder(nil, nil, nil, actionBtn, rowLbl)
				historyList.Add(row)
			}
			contentBox.Add(historyList)
		}

		contentBox.Refresh()
	}

	refreshContent()

	// Listen for queue updates to live-refresh the modal content
	origChanged := q.OnQueueChanged
	q.OnQueueChanged = func() {
		if origChanged != nil {
			origChanged()
		}
		refreshContent()
	}

	cancelAllBtn := widget.NewButtonWithIcon("Cancel All", theme.CancelIcon(), func() {
		q.CancelAll()
	})
	cancelAllBtn.Importance = widget.DangerImportance

	closeBtn := widget.NewButtonWithIcon("Close", theme.ConfirmIcon(), func() {
		q.OnQueueChanged = origChanged
		if d != nil {
			d.Hide()
		}
	})
	closeBtn.Importance = widget.HighImportance

	scrollable := container.NewVScroll(container.NewPadded(contentBox))
	scrollable.SetMinSize(fyne.NewSize(420, 360))

	d = dialog.NewCustomWithoutButtons("Download Queue", scrollable, app.window)
	d.SetButtons([]fyne.CanvasObject{cancelAllBtn, closeBtn})
	d.Resize(fyne.NewSize(460, 420))
	d.Show()
}
