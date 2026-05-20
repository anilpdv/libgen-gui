package ui

import (
	"time"

	"fyne.io/fyne/v2"
)

const (
	PrefKeyPreferredMirror      = "preferred_mirror"
	PrefKeyNetworkTimeout       = "network_timeout_sec"
	PrefKeyMaxRetries           = "max_retries"
	PrefKeyAutoQueue            = "auto_queue_enabled"
	PrefKeyAutoOpenBook         = "auto_open_book"
	PrefKeyShowCompletionModal  = "show_completion_modal"
	PrefKeyDefaultFormat        = "default_format_filter"
	PrefKeyEnableIPFS           = "enable_ipfs_fallback"

	DefaultPreferredMirror     = "auto"
	DefaultNetworkTimeoutSec   = 15
	DefaultMaxRetries          = 2
	DefaultAutoQueue           = true
	DefaultAutoOpenBook        = false
	DefaultShowCompletionModal = true
	DefaultFormatFilter        = "Any"
	DefaultEnableIPFS          = true
)

// AppSettings holds all customizable application settings.
type AppSettings struct {
	DownloadFolder      string
	PreferredMirror     string
	NetworkTimeout      time.Duration
	MaxRetries          int
	AutoQueue           bool
	AutoOpenBook        bool
	ShowCompletionModal bool
	DefaultFormat       string
	EnableIPFS          bool
}

// LoadAppSettings loads all preferences with safe defaults.
func LoadAppSettings(isMobile bool) AppSettings {
	app := fyne.CurrentApp()
	prefs := func() fyne.Preferences {
		if app != nil {
			return app.Preferences()
		}
		return nil
	}()

	s := AppSettings{
		DownloadFolder:      GetConfiguredSavePath(isMobile),
		PreferredMirror:     DefaultPreferredMirror,
		NetworkTimeout:      time.Duration(DefaultNetworkTimeoutSec) * time.Second,
		MaxRetries:          DefaultMaxRetries,
		AutoQueue:           DefaultAutoQueue,
		AutoOpenBook:        DefaultAutoOpenBook,
		ShowCompletionModal: DefaultShowCompletionModal,
		DefaultFormat:       DefaultFormatFilter,
		EnableIPFS:          DefaultEnableIPFS,
	}

	if prefs == nil {
		return s
	}

	if val := prefs.StringWithFallback(PrefKeyPreferredMirror, DefaultPreferredMirror); val != "" {
		s.PreferredMirror = val
	}
	if val := prefs.IntWithFallback(PrefKeyNetworkTimeout, DefaultNetworkTimeoutSec); val > 0 {
		s.NetworkTimeout = time.Duration(val) * time.Second
	}
	if val := prefs.IntWithFallback(PrefKeyMaxRetries, DefaultMaxRetries); val >= 0 {
		s.MaxRetries = val
	}
	s.AutoQueue = prefs.BoolWithFallback(PrefKeyAutoQueue, DefaultAutoQueue)
	s.AutoOpenBook = prefs.BoolWithFallback(PrefKeyAutoOpenBook, DefaultAutoOpenBook)
	s.ShowCompletionModal = prefs.BoolWithFallback(PrefKeyShowCompletionModal, DefaultShowCompletionModal)
	if val := prefs.StringWithFallback(PrefKeyDefaultFormat, DefaultFormatFilter); val != "" {
		s.DefaultFormat = val
	}
	s.EnableIPFS = prefs.BoolWithFallback(PrefKeyEnableIPFS, DefaultEnableIPFS)

	return s
}

// SavePreferredMirror persists the preferred mirror ("auto", "libgen.li", "libgen.vg", etc.)
func SavePreferredMirror(m string) {
	if m == "" {
		m = DefaultPreferredMirror
	}
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetString(PrefKeyPreferredMirror, m)
	}
}

// GetPreferredMirror returns the user's preferred mirror choice.
func GetPreferredMirror() string {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().StringWithFallback(PrefKeyPreferredMirror, DefaultPreferredMirror)
	}
	return DefaultPreferredMirror
}

// SaveNetworkTimeoutSec persists network timeout in seconds.
func SaveNetworkTimeoutSec(sec int) {
	if sec <= 0 {
		sec = DefaultNetworkTimeoutSec
	}
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetInt(PrefKeyNetworkTimeout, sec)
	}
}

// GetNetworkTimeout returns the configured network timeout duration.
func GetNetworkTimeout() time.Duration {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		sec := app.Preferences().IntWithFallback(PrefKeyNetworkTimeout, DefaultNetworkTimeoutSec)
		if sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return time.Duration(DefaultNetworkTimeoutSec) * time.Second
}

// SaveMaxRetries persists max download retry attempts.
func SaveMaxRetries(retries int) {
	if retries < 0 {
		retries = DefaultMaxRetries
	}
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetInt(PrefKeyMaxRetries, retries)
	}
}

// GetMaxRetries returns configured retry attempts.
func GetMaxRetries() int {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().IntWithFallback(PrefKeyMaxRetries, DefaultMaxRetries)
	}
	return DefaultMaxRetries
}

// SaveAutoQueue persists the auto-queue toggle.
func SaveAutoQueue(enabled bool) {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetBool(PrefKeyAutoQueue, enabled)
	}
}

// GetAutoQueue returns whether books should be automatically queued if a download is active.
func GetAutoQueue() bool {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().BoolWithFallback(PrefKeyAutoQueue, DefaultAutoQueue)
	}
	return DefaultAutoQueue
}

// SaveAutoOpenBook persists the auto-open toggle.
func SaveAutoOpenBook(enabled bool) {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetBool(PrefKeyAutoOpenBook, enabled)
	}
}

// GetAutoOpenBook returns whether downloaded books should be opened automatically.
func GetAutoOpenBook() bool {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().BoolWithFallback(PrefKeyAutoOpenBook, DefaultAutoOpenBook)
	}
	return DefaultAutoOpenBook
}

// SaveDefaultFormatFilter persists the default format filter.
func SaveDefaultFormatFilter(fmt string) {
	if fmt == "" {
		fmt = DefaultFormatFilter
	}
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetString(PrefKeyDefaultFormat, fmt)
	}
}

// GetDefaultFormatFilter returns the default format filter.
func GetDefaultFormatFilter() string {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().StringWithFallback(PrefKeyDefaultFormat, DefaultFormatFilter)
	}
	return DefaultFormatFilter
}

// SaveEnableIPFS persists whether IPFS fallback downloads are enabled.
func SaveEnableIPFS(enabled bool) {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		app.Preferences().SetBool(PrefKeyEnableIPFS, enabled)
	}
}

// GetEnableIPFS returns whether IPFS fallback downloads are enabled.
func GetEnableIPFS() bool {
	if app := fyne.CurrentApp(); app != nil && app.Preferences() != nil {
		return app.Preferences().BoolWithFallback(PrefKeyEnableIPFS, DefaultEnableIPFS)
	}
	return DefaultEnableIPFS
}

// ResetAllSettings restores all user preferences to factory defaults.
func ResetAllSettings(isMobile bool) {
	app := fyne.CurrentApp()
	if app == nil || app.Preferences() == nil {
		return
	}
	prefs := app.Preferences()
	prefs.RemoveValue(PrefKeyPreferredMirror)
	prefs.RemoveValue(PrefKeyNetworkTimeout)
	prefs.RemoveValue(PrefKeyMaxRetries)
	prefs.RemoveValue(PrefKeyAutoQueue)
	prefs.RemoveValue(PrefKeyAutoOpenBook)
	prefs.RemoveValue(PrefKeyShowCompletionModal)
	prefs.RemoveValue(PrefKeyDefaultFormat)
	prefs.RemoveValue(PrefKeyEnableIPFS)
	// reset download folder to platform default
	prefs.SetString(PrefKeyDownloadFolder, GetDefaultSavePath(isMobile))
	prefs.SetBool(PrefKeyFolderConfigured, true)
}
