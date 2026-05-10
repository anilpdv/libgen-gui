package ui

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"libgen-gui/pkg/libgen"
)

// ShowFullSettingsDialog displays a comprehensive, tabbed settings dialog
// covering Storage, Mirrors & Network Health, Downloads & Queue, and About.
func ShowFullSettingsDialog(a *App, initialTab int) {
	if a == nil || a.window == nil {
		return
	}

	isMobile := a.IsMobile()
	healthMgr := GetMirrorHealthManager()
	var d *dialog.CustomDialog

	// ─── TAB 1: Storage & Download Location ─────────────────────
	storageBox := container.NewVBox()

	selectedPath := ""
	if a.downloadBar != nil {
		selectedPath = NormalizePath(a.downloadBar.GetSavePath())
	}
	if selectedPath == "" {
		selectedPath = GetConfiguredSavePath(isMobile)
	}

	pathHeader := widget.NewLabelWithStyle("Download Directory", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	storageBox.Add(pathHeader)

	pathDisplay := widget.NewLabel(selectedPath)
	pathDisplay.Wrapping = fyne.TextWrapBreak
	pathDisplay.TextStyle = fyne.TextStyle{Monospace: true}
	storageBox.Add(pathDisplay)

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	updatePathStatus := func(p string) {
		normalized := NormalizePath(p)
		if IsDirWritable(normalized) {
			statusLabel.SetText("🟢 Location is writable and ready:\n" + normalized)
			statusLabel.Importance = widget.SuccessImportance
		} else {
			statusLabel.SetText("🔴 Directory is not writable. Please choose another location.")
			statusLabel.Importance = widget.DangerImportance
		}
	}
	updatePathStatus(selectedPath)
	storageBox.Add(statusLabel)

	presets := GetLocationPresets(isMobile)
	var presetLabels []string
	for _, p := range presets {
		presetLabels = append(presetLabels, p.Label)
	}
	presetRadio := widget.NewRadioGroup(presetLabels, func(choice string) {
		for _, p := range presets {
			if p.Label == choice {
				selectedPath = p.Path
				pathDisplay.SetText(selectedPath)
				updatePathStatus(selectedPath)
				SaveConfiguredSavePath(selectedPath)
				if a.downloadBar != nil {
					a.downloadBar.SetSavePath(selectedPath)
				}
				break
			}
		}
	})

	for _, p := range presets {
		if NormalizePath(p.Path) == NormalizePath(selectedPath) {
			presetRadio.SetSelected(p.Label)
			break
		}
	}

	presetLabel := widget.NewLabelWithStyle("Quick Presets:", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	storageBox.Add(presetLabel)
	storageBox.Add(presetRadio)

	browseBtn := widget.NewButtonWithIcon("Browse Other Folder...", theme.FolderOpenIcon(), func() {
		folderDlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			rawPath := uri.Path()
			if rawPath == "" {
				rawPath = uri.String()
			}
			resolved := NormalizePath(rawPath)
			if IsDirWritable(resolved) {
				selectedPath = resolved
				pathDisplay.SetText(selectedPath)
				updatePathStatus(selectedPath)
				SaveConfiguredSavePath(selectedPath)
				if a.downloadBar != nil {
					a.downloadBar.SetSavePath(selectedPath)
				}
				presetRadio.SetSelected("")
			} else {
				dialog.ShowError(fmt.Errorf("Selected folder is not writable: %s", resolved), a.window)
			}
		}, a.window)
		folderDlg.Resize(fyne.NewSize(500, 400))
		folderDlg.Show()
	})
	browseBtn.Importance = widget.MediumImportance

	openFolderBtn := widget.NewButtonWithIcon("Open Download Folder", theme.NavigateNextIcon(), func() {
		if selectedPath != "" {
			openFileInSystem(selectedPath)
		}
	})
	openFolderBtn.Importance = widget.LowImportance

	storageBox.Add(container.NewHBox(browseBtn, openFolderBtn))

	// ─── TAB 2: Mirrors & Network Health ────────────────────────
	networkBox := container.NewVBox()

	mirrorSummaryLabel := widget.NewLabelWithStyle("Checking mirror health…", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	mirrorListContainer := container.NewVBox()

	renderMirrors := func() {
		mirrorListContainer.Objects = nil
		mirrors := healthMgr.GetMirrors()
		activeSearch, totalSearch := healthMgr.GetActiveSearchCount()

		if activeSearch == totalSearch && totalSearch > 0 {
			mirrorSummaryLabel.SetText(fmt.Sprintf("🟢 All Search Mirrors Active (%d / %d Online)", activeSearch, totalSearch))
			mirrorSummaryLabel.Importance = widget.SuccessImportance
		} else if activeSearch > 0 {
			mirrorSummaryLabel.SetText(fmt.Sprintf("⚠️ Partial Mirror Availability (%d / %d Active)", activeSearch, totalSearch))
			mirrorSummaryLabel.Importance = widget.WarningImportance
		} else {
			mirrorSummaryLabel.SetText(fmt.Sprintf("🔴 All Search Mirrors Offline (%d / %d Active)", activeSearch, totalSearch))
			mirrorSummaryLabel.Importance = widget.DangerImportance
		}

		for _, m := range mirrors {
			info := m
			hostLbl := widget.NewLabelWithStyle(info.Host, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			roleLbl := widget.NewLabel(fmt.Sprintf("[%s]", info.Role))
			roleLbl.Importance = widget.LowImportance

			statusText := ""
			switch info.State {
			case StateOnline:
				statusText = fmt.Sprintf("🟢 Online (%dms)", info.Latency.Milliseconds())
				if info.StatusCode > 0 {
					statusText += fmt.Sprintf(" • HTTP %d", info.StatusCode)
				}
			case StateSlow:
				statusText = fmt.Sprintf("🟡 Slow (%dms)", info.Latency.Milliseconds())
			case StateOffline:
				if info.StatusCode > 0 {
					statusText = fmt.Sprintf("🔴 Offline (HTTP %d)", info.StatusCode)
				} else if info.ErrorMsg != "" {
					statusText = "🔴 Offline (" + info.ErrorMsg + ")"
				} else {
					statusText = "🔴 Offline"
				}
			case StateChecking:
				statusText = "⏳ Probing…"
			}

			statusLbl := widget.NewLabel(statusText)
			if info.State == StateOnline {
				statusLbl.Importance = widget.SuccessImportance
			} else if info.State == StateSlow {
				statusLbl.Importance = widget.WarningImportance
			} else if info.State == StateOffline {
				statusLbl.Importance = widget.DangerImportance
			}

			testSingleBtn := widget.NewButtonWithIcon("Test", theme.ViewRefreshIcon(), func() {
				statusLbl.SetText("⏳ Testing…")
				go func(targetHost string) {
					ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
					defer cancel()
					healthMgr.ProbeSingle(ctx, targetHost)
				}(info.Host)
			})
			testSingleBtn.Importance = widget.LowImportance

			row := container.NewBorder(
				nil, nil,
				container.NewHBox(hostLbl, roleLbl),
				testSingleBtn,
				statusLbl,
			)
			mirrorListContainer.Add(row)
		}
		mirrorListContainer.Refresh()
	}

	testAllBtn := widget.NewButtonWithIcon("Test All Mirrors", theme.ViewRefreshIcon(), func() {
		mirrorSummaryLabel.SetText("⏳ Probing all mirrors concurrently…")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			healthMgr.ProbeAll(ctx)
		}()
	})
	testAllBtn.Importance = widget.HighImportance

	networkBox.Add(container.NewBorder(nil, nil, mirrorSummaryLabel, testAllBtn, widget.NewLabel("")))
	networkBox.Add(widget.NewSeparator())
	networkBox.Add(mirrorListContainer)

	// Preferred Mirror option
	networkBox.Add(widget.NewSeparator())
	prefMirrorLabel := widget.NewLabelWithStyle("Preferred Search Mirror:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	networkBox.Add(prefMirrorLabel)

	mirrorOptions := []string{"Automatic (Fastest / Recommended)"}
	for _, m := range libgen.SearchMirrors {
		mirrorOptions = append(mirrorOptions, m.Host)
	}
	prefMirrorSelect := widget.NewSelect(mirrorOptions, func(selected string) {
		if strings.HasPrefix(selected, "Automatic") {
			SavePreferredMirror("auto")
		} else {
			SavePreferredMirror(selected)
		}
	})
	currentPref := GetPreferredMirror()
	if currentPref == "" || currentPref == "auto" {
		prefMirrorSelect.SetSelected("Automatic (Fastest / Recommended)")
	} else {
		prefMirrorSelect.SetSelected(currentPref)
	}
	networkBox.Add(prefMirrorSelect)

	// Network Timeout & Retries
	timeoutLabel := widget.NewLabelWithStyle("Network Request Timeout:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	timeoutSelect := widget.NewSelect([]string{"10 seconds", "15 seconds (Recommended)", "30 seconds", "45 seconds"}, func(s string) {
		switch {
		case strings.HasPrefix(s, "10"):
			SaveNetworkTimeoutSec(10)
		case strings.HasPrefix(s, "15"):
			SaveNetworkTimeoutSec(15)
		case strings.HasPrefix(s, "30"):
			SaveNetworkTimeoutSec(30)
		case strings.HasPrefix(s, "45"):
			SaveNetworkTimeoutSec(45)
		}
	})
	curTimeout := GetNetworkTimeout().Seconds()
	switch int(curTimeout) {
	case 10:
		timeoutSelect.SetSelected("10 seconds")
	case 30:
		timeoutSelect.SetSelected("30 seconds")
	case 45:
		timeoutSelect.SetSelected("45 seconds")
	default:
		timeoutSelect.SetSelected("15 seconds (Recommended)")
	}
	networkBox.Add(container.NewHBox(timeoutLabel, timeoutSelect))

	// IPFS Fallback Checkbox
	ipfsCheck := widget.NewCheck("Enable IPFS Gateway fallback downloads if primary mirror fails", func(val bool) {
		SaveEnableIPFS(val)
	})
	ipfsCheck.SetChecked(GetEnableIPFS())
	networkBox.Add(ipfsCheck)

	renderMirrors()

	// ─── TAB 3: Downloads & Queue ───────────────────────────────
	downloadsBox := container.NewVBox()

	queueSectionHeader := widget.NewLabelWithStyle("Queue & Concurrency", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	downloadsBox.Add(queueSectionHeader)

	autoQueueCheck := widget.NewCheck("Automatically enqueue books when a download is in progress", func(val bool) {
		SaveAutoQueue(val)
	})
	autoQueueCheck.SetChecked(GetAutoQueue())
	downloadsBox.Add(autoQueueCheck)

	autoOpenCheck := widget.NewCheck("Automatically open book after download completes", func(val bool) {
		SaveAutoOpenBook(val)
	})
	autoOpenCheck.SetChecked(GetAutoOpenBook())
	downloadsBox.Add(autoOpenCheck)

	downloadsBox.Add(widget.NewSeparator())
	filterHeader := widget.NewLabelWithStyle("Default Format Filter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	downloadsBox.Add(filterHeader)

	formatSelect := widget.NewSelect(formatOptions, func(val string) {
		SaveDefaultFormatFilter(val)
		if a.searchBar != nil && a.searchBar.format != nil {
			a.searchBar.format.SetSelected(val)
		}
	})
	formatSelect.SetSelected(GetDefaultFormatFilter())
	downloadsBox.Add(formatSelect)

	// ─── TAB 4: About & System ──────────────────────────────────
	aboutBox := container.NewVBox()

	appNameLbl := widget.NewLabelWithStyle("LibGen Downloader v2.0.0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	aboutBox.Add(appNameLbl)

	platformInfo := fmt.Sprintf("Platform: %s/%s • Go %s • Build 47", runtime.GOOS, runtime.GOARCH, runtime.Version())
	aboutBox.Add(widget.NewLabel(platformInfo))
	aboutBox.Add(widget.NewLabel("Ultra-fast, multi-mirror book search and resilient download manager."))

	aboutBox.Add(widget.NewSeparator())
	resetBtn := widget.NewButtonWithIcon("Reset All Settings to Defaults", theme.DeleteIcon(), func() {
		confirmDlg := dialog.NewConfirm("Reset Settings", "Are you sure you want to reset all settings to defaults?", func(confirmed bool) {
			if confirmed {
				ResetAllSettings(isMobile)
				selectedPath = GetDefaultSavePath(isMobile)
				pathDisplay.SetText(selectedPath)
				updatePathStatus(selectedPath)
				if a.downloadBar != nil {
					a.downloadBar.SetSavePath(selectedPath)
				}
				autoQueueCheck.SetChecked(DefaultAutoQueue)
				autoOpenCheck.SetChecked(DefaultAutoOpenBook)
				formatSelect.SetSelected(DefaultFormatFilter)
				ipfsCheck.SetChecked(DefaultEnableIPFS)
				prefMirrorSelect.SetSelected("Automatic (Fastest / Recommended)")
				timeoutSelect.SetSelected("15 seconds (Recommended)")
			}
		}, a.window)
		confirmDlg.Show()
	})
	resetBtn.Importance = widget.DangerImportance
	aboutBox.Add(resetBtn)

	var tabs *container.AppTabs
	if isMobile {
		tabs = container.NewAppTabs(
			container.NewTabItem("Storage", container.NewVScroll(container.NewPadded(storageBox))),
			container.NewTabItem("Mirrors", container.NewVScroll(container.NewPadded(networkBox))),
			container.NewTabItem("Queue", container.NewVScroll(container.NewPadded(downloadsBox))),
			container.NewTabItem("About", container.NewVScroll(container.NewPadded(aboutBox))),
		)
	} else {
		tabs = container.NewAppTabs(
			container.NewTabItemWithIcon("Storage", theme.FolderIcon(), container.NewVScroll(container.NewPadded(storageBox))),
			container.NewTabItemWithIcon("Mirrors", theme.ViewRefreshIcon(), container.NewVScroll(container.NewPadded(networkBox))),
			container.NewTabItemWithIcon("Downloads", theme.DownloadIcon(), container.NewVScroll(container.NewPadded(downloadsBox))),
			container.NewTabItemWithIcon("About", theme.InfoIcon(), container.NewVScroll(container.NewPadded(aboutBox))),
		)
	}

	// Subscribe to live mirror updates while the dialog is open
	unsub := healthMgr.Subscribe(func(activeS, totalS, activeA, totalA int) {
		renderMirrors()
	})

	if initialTab >= 0 && initialTab < len(tabs.Items) {
		tabs.SelectIndex(initialTab)
	}

	closeBtn := widget.NewButtonWithIcon("Close", theme.ConfirmIcon(), func() {
		unsub()
		if d != nil {
			d.Hide()
		}
	})
	closeBtn.Importance = widget.HighImportance

	dialogSize := fyne.NewSize(540, 480)
	if isMobile && a.window != nil {
		wSize := a.window.Canvas().Size()
		if wSize.Width > 0 && wSize.Height > 0 {
			dialogSize = fyne.NewSize(wSize.Width-16, wSize.Height-60)
		} else {
			dialogSize = fyne.NewSize(380, 520)
		}
	}

	d = dialog.NewCustomWithoutButtons("Settings & Diagnostics", tabs, a.window)
	d.SetButtons([]fyne.CanvasObject{closeBtn})
	d.Resize(dialogSize)
	d.Show()
}
