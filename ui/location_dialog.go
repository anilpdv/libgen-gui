package ui

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	PrefKeyDownloadFolder   = "download_folder"
	PrefKeyFolderConfigured = "download_folder_configured"
)

// NormalizePath converts raw filesystem paths, file:// URIs, or Android SAF (Storage Access Framework)
// document tree URIs into standard, absolute POSIX directory paths.
func NormalizePath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// 1. Strip file:// scheme
	if strings.HasPrefix(raw, "file://") {
		raw = strings.TrimPrefix(raw, "file://")
	}

	// 2. Decode URL encoding (e.g. %3A -> :, %20 -> space, %2F -> /)
	if decoded, err := url.QueryUnescape(raw); err == nil {
		raw = decoded
	}
	if decoded, err := url.PathUnescape(raw); err == nil {
		raw = decoded
	}
	raw = strings.ReplaceAll(raw, "%3A", ":")
	raw = strings.ReplaceAll(raw, "%3a", ":")
	raw = strings.ReplaceAll(raw, "%2F", "/")
	raw = strings.ReplaceAll(raw, "%2f", "/")
	raw = strings.ReplaceAll(raw, "%20", " ")

	// 3. Resolve Android SAF tree and document URIs
	// Examples:
	//   content://com.android.externalstorage.documents/tree/primary:Books
	//   /tree/primary:Books
	//   tree/primary:Books
	//   primary:Books
	//   raw:/storage/emulated/0/Download
	lower := strings.ToLower(raw)

	// Direct raw path in document provider (e.g. raw:/storage/emulated/0/Download)
	if rawIdx := strings.Index(lower, "raw:"); rawIdx != -1 {
		p := raw[rawIdx+len("raw:"):]
		if strings.HasPrefix(p, "/") {
			return filepath.Clean(p)
		}
	}

	if idx := strings.Index(lower, "primary:"); idx != -1 {
		subPath := raw[idx+len("primary:"):]
		subPath = strings.TrimPrefix(subPath, "/")
		base := "/storage/emulated/0"
		if ext := os.Getenv("EXTERNAL_STORAGE"); ext != "" {
			base = ext
		}
		if subPath == "" {
			return base
		}
		return filepath.Clean(filepath.Join(base, subPath))
	}

	// Secondary SD card storage (e.g. /tree/1234-5678:Books)
	if treeIdx := strings.Index(lower, "/tree/"); treeIdx != -1 {
		remainder := raw[treeIdx+len("/tree/"):]
		if colonIdx := strings.Index(remainder, ":"); colonIdx != -1 {
			volumeID := remainder[:colonIdx]
			subPath := strings.TrimPrefix(remainder[colonIdx+1:], "/")
			if !strings.Contains(volumeID, "/") && volumeID != "" {
				return filepath.Clean(filepath.Join("/storage", volumeID, subPath))
			}
		}
	}

	// Document provider (e.g. /document/1234-5678:Books)
	if docIdx := strings.Index(lower, "/document/"); docIdx != -1 {
		remainder := raw[docIdx+len("/document/"):]
		if colonIdx := strings.Index(remainder, ":"); colonIdx != -1 {
			volumeID := remainder[:colonIdx]
			subPath := strings.TrimPrefix(remainder[colonIdx+1:], "/")
			if strings.EqualFold(volumeID, "primary") {
				base := "/storage/emulated/0"
				if ext := os.Getenv("EXTERNAL_STORAGE"); ext != "" {
					base = ext
				}
				return filepath.Clean(filepath.Join(base, subPath))
			}
			if !strings.Contains(volumeID, "/") && volumeID != "" {
				return filepath.Clean(filepath.Join("/storage", volumeID, subPath))
			}
		}
	}

	return filepath.Clean(raw)
}

// LocationPreset represents a suggested download destination.
type LocationPreset struct {
	Label       string
	Path        string
	Description string
}

// GetLocationPresets returns recommended download directories based on platform.
func GetLocationPresets(isMobile bool) []LocationPreset {
	var presets []LocationPreset

	if isMobile || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		// 1. Android App-Specific External Storage (guaranteed writable without permissions)
		appExt := "/storage/emulated/0/Android/data/com.libgen.downloader/files/Download"
		presets = append(presets, LocationPreset{
			Label:       "App Storage (Always Accessible)",
			Path:        appExt,
			Description: "Private app folder. Always writable without permissions.",
		})

		// 2. Android Public Downloads (visible in Files / Downloads app)
		publicDl := "/storage/emulated/0/Download"
		presets = append(presets, LocationPreset{
			Label:       "Public Downloads (/sdcard/Download)",
			Path:        publicDl,
			Description: "Standard Downloads folder. Directly visible in Files and e-reader apps.",
		})

		// 3. Android Public Documents
		publicDocs := "/storage/emulated/0/Documents"
		presets = append(presets, LocationPreset{
			Label:       "Public Documents",
			Path:        publicDocs,
			Description: "Standard Documents folder on device storage.",
		})
	} else {
		// Desktop Presets (macOS, Windows, Linux)
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			dl := filepath.Join(home, "Downloads")
			if IsDirWritable(dl) {
				presets = append(presets, LocationPreset{
					Label:       "Downloads (Recommended)",
					Path:        dl,
					Description: "Standard user Downloads folder.",
				})
			}
			docs := filepath.Join(home, "Documents")
			if IsDirWritable(docs) {
				presets = append(presets, LocationPreset{
					Label:       "Documents",
					Path:        docs,
					Description: "User Documents directory.",
				})
			}
			desktop := filepath.Join(home, "Desktop")
			if IsDirWritable(desktop) {
				presets = append(presets, LocationPreset{
					Label:       "Desktop",
					Path:        desktop,
					Description: "User Desktop folder.",
				})
			}
		}
	}

	return presets
}

// GetConfiguredSavePath retrieves the user-configured download folder from preferences,
// or returns the system default if not yet set or no longer writable.
func GetConfiguredSavePath(isMobile bool) string {
	defaultPath := GetDefaultSavePath(isMobile)
	app := fyne.CurrentApp()
	if app == nil || app.Preferences() == nil {
		return defaultPath
	}

	saved := app.Preferences().StringWithFallback(PrefKeyDownloadFolder, "")
	if saved != "" {
		normalized := NormalizePath(saved)
		if IsDirWritable(normalized) {
			return normalized
		}
	}
	return defaultPath
}

// SaveConfiguredSavePath stores the chosen path in Fyne preferences.
func SaveConfiguredSavePath(path string) {
	path = NormalizePath(path)
	app := fyne.CurrentApp()
	if app != nil && app.Preferences() != nil {
		app.Preferences().SetString(PrefKeyDownloadFolder, path)
		app.Preferences().SetBool(PrefKeyFolderConfigured, true)
	}
}

// IsDownloadFolderConfigured returns true if the user has completed location setup.
func IsDownloadFolderConfigured() bool {
	app := fyne.CurrentApp()
	if app == nil || app.Preferences() == nil {
		return false
	}
	return app.Preferences().BoolWithFallback(PrefKeyFolderConfigured, false)
}

// ShowDownloadLocationDialog displays a 100% graphical modal to configure the download directory without typing.
func ShowDownloadLocationDialog(a *App, isFirstRun bool, onDone func(path string)) {
	if a == nil || a.window == nil {
		return
	}

	isMobile := a.IsMobile()
	presets := GetLocationPresets(isMobile)
	selectedPath := ""
	if a.downloadBar != nil {
		selectedPath = NormalizePath(a.downloadBar.GetSavePath())
	}
	if selectedPath == "" {
		selectedPath = GetConfiguredSavePath(isMobile)
	}

	title := "Download Location Settings"
	if isFirstRun {
		title = "Welcome! Set Download Location"
	}

	headerMsg := "Select where downloaded books will be saved:"
	if isFirstRun {
		headerMsg = "Before downloading books, please choose your preferred save location:"
	}
	headerLabel := widget.NewLabelWithStyle(headerMsg, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	headerLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	curLabel := widget.NewLabel(fmt.Sprintf("Selected: %s", selectedPath))
	curLabel.Wrapping = fyne.TextWrapBreak
	curLabel.Importance = widget.LowImportance

	updateValidation := func(p string) bool {
		p = NormalizePath(p)
		if p == "" {
			statusLabel.SetText("Please select a directory.")
			statusLabel.Importance = widget.WarningImportance
			return false
		}
		if IsDirWritable(p) {
			statusLabel.SetText(fmt.Sprintf("Location is writable and ready: %s", p))
			statusLabel.Importance = widget.SuccessImportance
			curLabel.SetText(fmt.Sprintf("Selected: %s", p))
			return true
		}
		statusLabel.SetText(fmt.Sprintf("Path is not writable or requires permissions: %s\nTip: Select 'Public Downloads' (/sdcard/Download) or 'App Storage' above.", p))
		statusLabel.Importance = widget.DangerImportance
		curLabel.SetText(fmt.Sprintf("Selected: %s", p))
		return false
	}

	// Preset options
	options := make([]string, 0, len(presets))
	pathToOption := make(map[string]string)
	optionToPath := make(map[string]string)

	for _, p := range presets {
		opt := p.Label
		options = append(options, opt)
		pathToOption[p.Path] = opt
		optionToPath[opt] = p.Path
	}

	var radio *widget.RadioGroup
	radio = widget.NewRadioGroup(options, func(selected string) {
		if path, exists := optionToPath[selected]; exists {
			selectedPath = NormalizePath(path)
			updateValidation(selectedPath)
		}
	})

	// Set initial selection
	if opt, ok := pathToOption[selectedPath]; ok {
		radio.SetSelected(opt)
	}
	updateValidation(selectedPath)

	var d *dialog.CustomDialog

	// 100% Graphical Folder Browser (opens native / Fyne folder picker on ALL platforms)
	browseBtn := widget.NewButtonWithIcon("Browse Other Folder…", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				raw := uri.Path()
				if raw == "" {
					raw = uri.String()
				}
				normalized := NormalizePath(raw)
				if normalized != "" {
					selectedPath = normalized
					radio.SetSelected("") // clear preset selection
					updateValidation(selectedPath)
				}
			}
		}, a.window)
	})

	saveBtn := widget.NewButtonWithIcon("Save & Continue", theme.ConfirmIcon(), func() {
		targetPath := NormalizePath(selectedPath)
		if !IsDirWritable(targetPath) {
			statusLabel.SetText(fmt.Sprintf("Cannot save to '%s'. Please choose a writable folder or preset above.", targetPath))
			statusLabel.Importance = widget.DangerImportance
			return
		}

		SaveConfiguredSavePath(targetPath)
		if a.downloadBar != nil {
			a.downloadBar.SetSavePath(targetPath)
		}
		if onDone != nil {
			onDone(targetPath)
		}
		if d != nil {
			d.Hide()
		}
	})
	saveBtn.Importance = widget.HighImportance

	var cancelBtn *widget.Button
	if !isFirstRun {
		cancelBtn = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
			if d != nil {
				d.Hide()
			}
		})
	}

	contentItems := []fyne.CanvasObject{
		headerLabel,
		widget.NewSeparator(),
		radio,
		browseBtn,
		widget.NewSeparator(),
		curLabel,
		statusLabel,
	}

	content := container.NewVBox(contentItems...)
	padded := container.NewPadded(content)

	buttons := []fyne.CanvasObject{saveBtn}
	if cancelBtn != nil {
		buttons = append([]fyne.CanvasObject{cancelBtn}, saveBtn)
	}

	d = dialog.NewCustomWithoutButtons(title, padded, a.window)
	d.SetButtons(buttons)
	if isMobile {
		d.Resize(fyne.NewSize(360, 320))
	} else {
		d.Resize(fyne.NewSize(480, 340))
	}
	d.Show()
}
