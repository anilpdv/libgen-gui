package ui

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"libgen-gui/internal/settings"
	"libgen-gui/internal/storage"
	"libgen-gui/pkg/libgen"
)

// ShowFullSettingsDialog displays a comprehensive settings dialog covering Downloads, Network, Mirrors, and Appearance.
func ShowFullSettingsDialog(a *App, initialTab int) {
	if a == nil || a.window == nil {
		return
	}

	isMobile := a.IsMobile()
	healthMgr := GetMirrorHealthManager()
	var d *dialog.CustomDialog

	// Initialize Settings repository & controller
	defaultLoc, _ := storage.ResolveDefaultLocation(isMobile)
	defaultSettings := settings.Default()
	if defaultLoc.Kind == storage.LocationAndroidSAF {
		defaultSettings.StorageURI = defaultLoc.URI
		defaultSettings.StorageDisplayName = defaultLoc.DisplayName
	} else {
		defaultSettings.DownloadLocation = defaultLoc.Path
		defaultSettings.StorageDisplayName = defaultLoc.DisplayName
	}

	app := fyne.CurrentApp()
	var repo settings.Repository
	if app != nil && app.Preferences() != nil {
		repo = settings.NewFyneRepository(app.Preferences(), defaultSettings)
	} else {
		repo = settings.NewMemoryRepository(defaultSettings)
	}

	currentSettings, _ := repo.Load()
	initialLoc := storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        currentSettings.DownloadLocation,
		URI:         currentSettings.StorageURI,
		DisplayName: currentSettings.StorageDisplayName,
	}
	if currentSettings.StorageURI != "" {
		initialLoc.Kind = storage.LocationAndroidSAF
	}
	if initialLoc.DisplayName == "" {
		if initialLoc.Path != "" {
			initialLoc.DisplayName = initialLoc.Path
		} else if initialLoc.URI != "" {
			initialLoc.DisplayName = initialLoc.URI
		} else {
			initialLoc = defaultLoc
		}
	}

	targetProv, _ := storage.NewMutableTargetProvider(initialLoc)
	locSelector := storage.NewDesktopLocationSelector(a.window)
	locOpener := storage.NewDesktopLocationOpener()
	ctrl := settings.NewController(repo, locSelector, locOpener, targetProv)

	// ─── TAB 1: Downloads ───────────────────────────────────────
	downloadsBox := container.NewVBox()

	curDisplay := currentSettings.StorageDisplayName
	if curDisplay == "" {
		if currentSettings.DownloadLocation != "" {
			curDisplay = currentSettings.DownloadLocation
		} else if currentSettings.StorageURI != "" {
			curDisplay = currentSettings.StorageURI
		} else {
			curDisplay = defaultLoc.DisplayName
		}
	}

	pathHeader := widget.NewLabelWithStyle("Current Download Folder", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	downloadsBox.Add(pathHeader)

	pathDisplay := widget.NewLabel(curDisplay)
	pathDisplay.Wrapping = fyne.TextWrapBreak
	pathDisplay.TextStyle = fyne.TextStyle{Monospace: true}
	downloadsBox.Add(pathDisplay)

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	updatePathStatus := func(display string) {
		normalized := NormalizePath(display)
		if normalized != "" && IsDirWritable(normalized) {
			statusLabel.SetText("🟢 Folder is writable and ready")
			statusLabel.Importance = widget.SuccessImportance
		} else if currentSettings.StorageURI != "" {
			statusLabel.SetText("🟢 Android Storage Access Framework folder configured")
			statusLabel.Importance = widget.SuccessImportance
		} else {
			statusLabel.SetText("⚠️ Please verify folder write permissions")
			statusLabel.Importance = widget.WarningImportance
		}
	}
	updatePathStatus(curDisplay)
	downloadsBox.Add(statusLabel)

	// Change Folder Button
	changeFolderBtn := widget.NewButtonWithIcon("Change Folder", theme.FolderOpenIcon(), nil)
	changeFolderBtn.Importance = widget.HighImportance
	changeFolderBtn.OnTapped = func() {
		changeFolderBtn.Disable()
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			location, err := ctrl.ChangeDownloadLocation(ctx)
			changeFolderBtn.Enable()

			if errors.Is(err, storage.ErrSelectionCancelled) {
				return
			}

			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}

			pathDisplay.SetText(location.DisplayName)
			updatePathStatus(location.DisplayName)
			if a.downloadBar != nil {
				if location.Path != "" {
					a.downloadBar.SetSavePath(location.Path)
				} else {
					a.downloadBar.SetSavePath(location.DisplayName)
				}
			}
		}()
	}

	// Open Folder Button
	openFolderBtn := widget.NewButtonWithIcon("Open Folder", theme.NavigateNextIcon(), func() {
		go func() {
			err := ctrl.OpenDownloadLocation(context.Background())
			if err != nil {
				// Fallback to direct system open
				if currentSettings.DownloadLocation != "" {
					openFileInSystem(currentSettings.DownloadLocation)
				} else {
					dialog.ShowError(err, a.window)
				}
			}
		}()
	})
	openFolderBtn.Importance = widget.LowImportance

	// Reset to Default Button
	resetFolderBtn := widget.NewButtonWithIcon("Reset to Default", theme.ViewRefreshIcon(), func() {
		if err := ctrl.ResetToDefault(defaultLoc); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		pathDisplay.SetText(defaultLoc.DisplayName)
		updatePathStatus(defaultLoc.DisplayName)
		if a.downloadBar != nil {
			a.downloadBar.SetSavePath(defaultLoc.Path)
		}
	})
	resetFolderBtn.Importance = widget.MediumImportance

	var folderActionRow fyne.CanvasObject
	if isMobile {
		folderActionRow = container.NewVBox(changeFolderBtn, resetFolderBtn)
	} else {
		folderActionRow = container.NewHBox(changeFolderBtn, openFolderBtn, resetFolderBtn)
	}
	downloadsBox.Add(folderActionRow)

	downloadsBox.Add(widget.NewSeparator())

	// Open after download check
	openAfterCheck := widget.NewCheck("Open completed files automatically", func(enabled bool) {
		_ = ctrl.SetOpenAfterDownload(enabled)
		SaveAutoOpenBook(enabled)
	})
	openAfterCheck.SetChecked(currentSettings.OpenAfterDownload)
	downloadsBox.Add(openAfterCheck)

	// Auto-queue check
	autoQueueCheck := widget.NewCheck("Automatically enqueue downloads when active", func(enabled bool) {
		_ = ctrl.SetAutoQueue(enabled)
		SaveAutoQueue(enabled)
	})
	autoQueueCheck.SetChecked(currentSettings.AutoQueue)
	downloadsBox.Add(autoQueueCheck)

	// Existing file behavior
	existingLabel := widget.NewLabelWithStyle("When a file already exists:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	downloadsBox.Add(existingLabel)

	existingFileSelect := widget.NewSelect(
		[]string{"Rename automatically", "Overwrite", "Skip"},
		func(selected string) {
			var action settings.ExistingFileAction
			switch selected {
			case "Rename automatically":
				action = settings.ExistingFileRename
			case "Overwrite":
				action = settings.ExistingFileOverwrite
			case "Skip":
				action = settings.ExistingFileSkip
			default:
				action = settings.ExistingFileRename
			}
			_ = ctrl.SetExistingFileAction(action)
		},
	)
	switch currentSettings.ExistingFileAction {
	case settings.ExistingFileOverwrite:
		existingFileSelect.SetSelected("Overwrite")
	case settings.ExistingFileSkip:
		existingFileSelect.SetSelected("Skip")
	default:
		existingFileSelect.SetSelected("Rename automatically")
	}
	downloadsBox.Add(existingFileSelect)

	// Quick presets for convenience
	presets := GetLocationPresets(isMobile)
	if len(presets) > 0 {
		downloadsBox.Add(widget.NewSeparator())
		presetLabel := widget.NewLabelWithStyle("Quick Location Presets:", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
		downloadsBox.Add(presetLabel)

		var presetLabels []string
		for _, p := range presets {
			presetLabels = append(presetLabels, p.Label)
		}
		presetRadio := widget.NewRadioGroup(presetLabels, func(choice string) {
			for _, p := range presets {
				if p.Label == choice {
					loc := storage.Location{
						Kind:        storage.LocationDesktopPath,
						Path:        p.Path,
						DisplayName: p.Path,
					}
					if isMobile || strings.HasPrefix(p.Path, "content://") {
						loc.Kind = storage.LocationAndroidSAF
						loc.URI = p.Path
					}
					_ = targetProv.Set(loc)
					currentSettings.DownloadLocation = loc.Path
					currentSettings.StorageURI = loc.URI
					currentSettings.StorageDisplayName = loc.DisplayName
					_ = repo.Save(currentSettings)

					pathDisplay.SetText(p.Path)
					updatePathStatus(p.Path)
					SaveConfiguredSavePath(p.Path)
					if a.downloadBar != nil {
						a.downloadBar.SetSavePath(p.Path)
					}
					break
				}
			}
		})
		for _, p := range presets {
			if NormalizePath(p.Path) == NormalizePath(curDisplay) {
				presetRadio.SetSelected(p.Label)
				break
			}
		}
		downloadsBox.Add(presetRadio)
	}

	// ─── TAB 2: Network ─────────────────────────────────────────
	networkBox := container.NewVBox()

	timeoutHeader := widget.NewLabelWithStyle("Network Timeout & Retry Limit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	networkBox.Add(timeoutHeader)

	timeoutLabel := widget.NewLabel("Request Timeout:")
	timeoutSelect := widget.NewSelect([]string{"10 seconds", "15 seconds (Recommended)", "30 seconds", "45 seconds", "60 seconds"}, func(s string) {
		sec := 15
		switch {
		case strings.HasPrefix(s, "10"):
			sec = 10
		case strings.HasPrefix(s, "15"):
			sec = 15
		case strings.HasPrefix(s, "30"):
			sec = 30
		case strings.HasPrefix(s, "45"):
			sec = 45
		case strings.HasPrefix(s, "60"):
			sec = 60
		}
		_ = ctrl.SetNetworkTimeout(time.Duration(sec) * time.Second)
		SaveNetworkTimeoutSec(sec)
	})
	switch int(currentSettings.RequestTimeout.Seconds()) {
	case 10:
		timeoutSelect.SetSelected("10 seconds")
	case 30:
		timeoutSelect.SetSelected("30 seconds")
	case 45:
		timeoutSelect.SetSelected("45 seconds")
	case 60:
		timeoutSelect.SetSelected("60 seconds")
	default:
		timeoutSelect.SetSelected("15 seconds (Recommended)")
	}
	networkBox.Add(container.NewHBox(timeoutLabel, timeoutSelect))

	retryLabel := widget.NewLabel("Retry Limit:")
	retrySelect := widget.NewSelect([]string{"1 retry", "2 retries (Default)", "3 retries", "5 retries"}, func(s string) {
		r := 2
		switch {
		case strings.HasPrefix(s, "1"):
			r = 1
		case strings.HasPrefix(s, "2"):
			r = 2
		case strings.HasPrefix(s, "3"):
			r = 3
		case strings.HasPrefix(s, "5"):
			r = 5
		}
		_ = ctrl.SetRetryCount(r)
		SaveMaxRetries(r)
	})
	switch currentSettings.RetryCount {
	case 1:
		retrySelect.SetSelected("1 retry")
	case 3:
		retrySelect.SetSelected("3 retries")
	case 5:
		retrySelect.SetSelected("5 retries")
	default:
		retrySelect.SetSelected("2 retries (Default)")
	}
	networkBox.Add(container.NewHBox(retryLabel, retrySelect))

	ipfsCheck := widget.NewCheck("Enable IPFS Gateway fallback downloads if primary mirror fails", func(val bool) {
		_ = ctrl.SetEnableIPFS(val)
		SaveEnableIPFS(val)
	})
	ipfsCheck.SetChecked(currentSettings.EnableIPFS)
	networkBox.Add(ipfsCheck)

	// ─── TAB 3: Mirrors ─────────────────────────────────────────
	mirrorsBox := container.NewVBox()

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

	recheckAllBtn := widget.NewButtonWithIcon("Recheck Mirrors", theme.ViewRefreshIcon(), func() {
		mirrorSummaryLabel.SetText("⏳ Probing all mirrors concurrently…")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			healthMgr.ProbeAll(ctx)
		}()
	})
	recheckAllBtn.Importance = widget.HighImportance

	mirrorsBox.Add(container.NewBorder(nil, nil, mirrorSummaryLabel, recheckAllBtn, widget.NewLabel("")))
	mirrorsBox.Add(widget.NewSeparator())
	mirrorsBox.Add(mirrorListContainer)

	mirrorsBox.Add(widget.NewSeparator())
	prefMirrorLabel := widget.NewLabelWithStyle("Preferred Search Mirror:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	mirrorsBox.Add(prefMirrorLabel)

	mirrorOptions := []string{"Automatic (Fastest / Recommended)"}
	for _, m := range libgen.SearchMirrors {
		mirrorOptions = append(mirrorOptions, m.Host)
	}
	prefMirrorSelect := widget.NewSelect(mirrorOptions, func(selected string) {
		if strings.HasPrefix(selected, "Automatic") {
			_ = ctrl.SetMirrorMode(settings.MirrorModeAutomatic, "auto")
			SavePreferredMirror("auto")
		} else {
			_ = ctrl.SetMirrorMode(settings.MirrorModePreferred, selected)
			SavePreferredMirror(selected)
		}
	})
	if currentSettings.MirrorMode == settings.MirrorModePreferred && currentSettings.PreferredMirror != "" && currentSettings.PreferredMirror != "auto" {
		prefMirrorSelect.SetSelected(currentSettings.PreferredMirror)
	} else {
		prefMirrorSelect.SetSelected("Automatic (Fastest / Recommended)")
	}
	mirrorsBox.Add(prefMirrorSelect)

	renderMirrors()

	// ─── TAB 4: Appearance ──────────────────────────────────────
	appearanceBox := container.NewVBox()

	themeHeader := widget.NewLabelWithStyle("Theme Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	appearanceBox.Add(themeHeader)

	themeSelect := widget.NewSelect([]string{"System Default", "Dark Theme", "Light Theme"}, func(choice string) {
		var t settings.Theme
		switch choice {
		case "Dark Theme":
			t = settings.ThemeDark
		case "Light Theme":
			t = settings.ThemeLight
		default:
			t = settings.ThemeSystem
		}
		_ = ctrl.SetTheme(t)
	})
	switch currentSettings.Theme {
	case settings.ThemeDark:
		themeSelect.SetSelected("Dark Theme")
	case settings.ThemeLight:
		themeSelect.SetSelected("Light Theme")
	default:
		themeSelect.SetSelected("System Default")
	}
	appearanceBox.Add(themeSelect)

	appearanceBox.Add(widget.NewSeparator())
	filterHeader := widget.NewLabelWithStyle("Default Format Filter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	appearanceBox.Add(filterHeader)

	formatSelect := widget.NewSelect(formatOptions, func(val string) {
		_ = ctrl.SetDefaultFormat(val)
		SaveDefaultFormatFilter(val)
		if a.searchBar != nil && a.searchBar.format != nil {
			a.searchBar.format.SetSelected(val)
		}
	})
	formatSelect.SetSelected(GetDefaultFormatFilter())
	appearanceBox.Add(formatSelect)

	appearanceBox.Add(widget.NewSeparator())
	appNameLbl := widget.NewLabelWithStyle("LibGen Downloader v2.0.0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	appearanceBox.Add(appNameLbl)
	platformInfo := fmt.Sprintf("Platform: %s/%s • Go %s • Build 48", runtime.GOOS, runtime.GOARCH, runtime.Version())
	appearanceBox.Add(widget.NewLabel(platformInfo))

	// ─── Tabs Construction ──────────────────────────────────────
	var tabs *container.AppTabs
	if isMobile {
		tabs = container.NewAppTabs(
			container.NewTabItem("Downloads", container.NewVScroll(container.NewPadded(downloadsBox))),
			container.NewTabItem("Network", container.NewVScroll(container.NewPadded(networkBox))),
			container.NewTabItem("Mirrors", container.NewVScroll(container.NewPadded(mirrorsBox))),
			container.NewTabItem("Appearance", container.NewVScroll(container.NewPadded(appearanceBox))),
		)
	} else {
		tabs = container.NewAppTabs(
			container.NewTabItemWithIcon("Downloads", theme.DownloadIcon(), container.NewVScroll(container.NewPadded(downloadsBox))),
			container.NewTabItemWithIcon("Network", theme.HelpIcon(), container.NewVScroll(container.NewPadded(networkBox))),
			container.NewTabItemWithIcon("Mirrors", theme.ViewRefreshIcon(), container.NewVScroll(container.NewPadded(mirrorsBox))),
			container.NewTabItemWithIcon("Appearance", theme.ColorPaletteIcon(), container.NewVScroll(container.NewPadded(appearanceBox))),
		)
	}

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

	dialogSize := fyne.NewSize(560, 500)
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
