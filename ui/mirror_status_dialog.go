package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MirrorStatusDialog is the focused live-health view opened from the mirror
// status button. It intentionally keeps mirror preferences in Settings and
// presents only current availability and diagnostics here.
type MirrorStatusDialog struct {
	app     *App
	manager *MirrorHealthManager

	Dialog        dialog.Dialog
	SummaryLabel  *widget.Label
	LastUpdated   *widget.Label
	MirrorList    *fyne.Container
	RefreshButton *widget.Button

	mu          sync.Mutex
	unsubscribe func()
	closed      bool
}

// NewMirrorStatusDialog builds a dedicated mirror status dialog. Accepting the
// manager as an argument keeps the view testable while the application uses the
// shared manager through ShowMirrorStatusDialog.
func NewMirrorStatusDialog(a *App, manager *MirrorHealthManager) *MirrorStatusDialog {
	if a == nil || a.window == nil {
		return nil
	}
	if manager == nil {
		manager = GetMirrorHealthManager()
	}

	view := &MirrorStatusDialog{
		app:          a,
		manager:      manager,
		SummaryLabel: widget.NewLabelWithStyle("Checking mirror health…", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		LastUpdated:  widget.NewLabel(""),
		MirrorList:   container.NewVBox(),
	}
	view.LastUpdated.Importance = widget.LowImportance

	view.RefreshButton = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		view.refresh()
	})
	view.RefreshButton.Importance = widget.HighImportance

	header := container.NewBorder(
		nil,
		view.LastUpdated,
		nil,
		view.RefreshButton,
		view.SummaryLabel,
	)
	help := widget.NewLabel("Live reachability and response latency for every configured search mirror and IPFS gateway.")
	help.Wrapping = fyne.TextWrapWord
	help.Importance = widget.LowImportance

	listScroll := container.NewVScroll(container.NewPadded(view.MirrorList))
	content := container.NewBorder(
		container.NewVBox(header, help, widget.NewSeparator()),
		nil, nil, nil,
		listScroll,
	)

	var custom *dialog.CustomDialog
	closeButton := widget.NewButtonWithIcon("Close", theme.ConfirmIcon(), func() {
		if custom != nil {
			custom.Hide()
		}
	})
	custom = dialog.NewCustomWithoutButtons("Mirror Status", content, a.window)
	custom.SetButtons([]fyne.CanvasObject{closeButton})
	custom.SetOnClosed(view.close)
	view.Dialog = custom

	size := fyne.NewSize(620, 460)
	if a.IsMobile() {
		windowSize := a.window.Canvas().Size()
		if windowSize.Width > 0 && windowSize.Height > 0 {
			size = fyne.NewSize(windowSize.Width-16, windowSize.Height-60)
		} else {
			size = fyne.NewSize(380, 520)
		}
	}
	custom.Resize(size)

	view.render()
	view.unsubscribe = manager.Subscribe(func(_, _, _, _ int) {
		view.render()
	})
	return view
}

// Show displays the mirror status dialog.
func (d *MirrorStatusDialog) Show() {
	if d != nil && d.Dialog != nil {
		d.Dialog.Show()
	}
}

func (d *MirrorStatusDialog) close() {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return
	}
	d.closed = true
	unsubscribe := d.unsubscribe
	d.unsubscribe = nil
	d.mu.Unlock()

	if unsubscribe != nil {
		unsubscribe()
	}
}

func (d *MirrorStatusDialog) refresh() {
	if d == nil || d.manager == nil {
		return
	}

	d.RefreshButton.Disable()
	d.SummaryLabel.SetText("⏳ Refreshing mirror status…")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		d.manager.ProbeAll(ctx)
		d.RefreshButton.Enable()
		d.render()
	}()
}

func (d *MirrorStatusDialog) render() {
	if d == nil || d.manager == nil || d.MirrorList == nil {
		return
	}

	mirrors := d.manager.GetMirrors()
	activeSearch, totalSearch := d.manager.GetActiveSearchCount()
	activeAll, totalAll := d.manager.GetActiveAllCount()

	switch {
	case totalAll == 0:
		d.SummaryLabel.SetText("No mirrors configured")
		d.SummaryLabel.Importance = widget.WarningImportance
	case activeAll == totalAll:
		d.SummaryLabel.SetText(fmt.Sprintf("🟢 All mirrors available (%d/%d) • Search %d/%d", activeAll, totalAll, activeSearch, totalSearch))
		d.SummaryLabel.Importance = widget.SuccessImportance
	case activeAll > 0:
		d.SummaryLabel.SetText(fmt.Sprintf("🟡 Partial availability (%d/%d) • Search %d/%d", activeAll, totalAll, activeSearch, totalSearch))
		d.SummaryLabel.Importance = widget.WarningImportance
	default:
		d.SummaryLabel.SetText(fmt.Sprintf("🔴 No mirrors available (0/%d) • Search 0/%d", totalAll, totalSearch))
		d.SummaryLabel.Importance = widget.DangerImportance
	}
	d.SummaryLabel.Refresh()

	objects := make([]fyne.CanvasObject, 0, len(mirrors)*2)
	var newest time.Time
	for _, mirror := range mirrors {
		if mirror.LastChecked.After(newest) {
			newest = mirror.LastChecked
		}

		host := widget.NewLabelWithStyle(mirror.Host, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		role := widget.NewLabel(string(mirror.Role))
		role.Importance = widget.LowImportance

		status := widget.NewLabel(formatMirrorStatus(mirror))
		status.Wrapping = fyne.TextWrapWord
		switch mirror.State {
		case StateOnline:
			status.Importance = widget.SuccessImportance
		case StateSlow:
			status.Importance = widget.WarningImportance
		case StateOffline:
			status.Importance = widget.DangerImportance
		}

		latency := widget.NewLabel(formatMirrorLatency(mirror))
		latency.Alignment = fyne.TextAlignTrailing
		latency.TextStyle = fyne.TextStyle{Monospace: true}
		latency.Importance = widget.LowImportance

		details := container.NewVBox(host, role)
		row := container.NewBorder(nil, nil, details, latency, status)
		objects = append(objects, row, widget.NewSeparator())
	}
	if len(objects) > 0 {
		objects = objects[:len(objects)-1]
	}
	d.MirrorList.Objects = objects
	d.MirrorList.Refresh()

	if newest.IsZero() {
		d.LastUpdated.SetText("Not checked yet")
	} else {
		d.LastUpdated.SetText("Last checked " + newest.Local().Format("15:04:05"))
	}
	d.LastUpdated.Refresh()
}

func formatMirrorLatency(mirror MirrorInfo) string {
	if mirror.State == StateChecking || mirror.Latency <= 0 {
		return "—"
	}
	if mirror.Latency < time.Millisecond {
		return "<1 ms"
	}
	return fmt.Sprintf("%d ms", mirror.Latency.Milliseconds())
}

func formatMirrorStatus(mirror MirrorInfo) string {
	switch mirror.State {
	case StateOnline:
		if mirror.StatusCode > 0 {
			return fmt.Sprintf("● Online • HTTP %d", mirror.StatusCode)
		}
		return "● Online"
	case StateSlow:
		if mirror.StatusCode > 0 {
			return fmt.Sprintf("● Slow • HTTP %d", mirror.StatusCode)
		}
		return "● Slow"
	case StateOffline:
		if mirror.StatusCode > 0 {
			return fmt.Sprintf("● Offline • HTTP %d", mirror.StatusCode)
		}
		if message := strings.TrimSpace(mirror.ErrorMsg); message != "" {
			if len(message) > 70 {
				message = message[:67] + "…"
			}
			return "● Offline • " + message
		}
		return "● Offline"
	default:
		return "● Checking…"
	}
}

// ShowMirrorStatusDialog opens the dedicated live mirror diagnostics view.
func ShowMirrorStatusDialog(a *App) {
	view := NewMirrorStatusDialog(a, GetMirrorHealthManager())
	if view != nil {
		view.Show()
	}
}
