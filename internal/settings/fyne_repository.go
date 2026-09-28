package settings

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const (
	keyDownloadLocation   = "settings.download.location"
	keyStorageURI         = "settings.download.storage_uri"
	keyStorageDisplayName = "settings.download.storage_display_name"
	keyOpenAfterDownload  = "settings.download.open_after"
	keyExistingFileAction = "settings.download.existing_file_action"

	keyRequestTimeoutSeconds = "settings.network.timeout_seconds"
	keyRetryCount            = "settings.network.retry_count"
	keyEnableIPFS            = "settings.network.enable_ipfs"

	keyMirrorMode      = "settings.mirrors.mode"
	keyPreferredMirror = "settings.mirrors.preferred"

	keyTheme         = "settings.appearance.theme"
	keyDefaultFormat = "settings.search.default_format"
	keyAutoQueue     = "settings.download.auto_queue"

	// Legacy preference keys for backwards compatibility / migration
	legacyDownloadDirectoryKey = "download_directory"
	legacyDownloadFolderKey    = "download_folder"
	legacyPreferredMirrorKey   = "preferred_mirror"
	legacyNetworkTimeoutSecKey = "network_timeout_sec"
	legacyMaxRetriesKey        = "max_retries"
	legacyAutoQueueKey         = "auto_queue_enabled"
	legacyAutoOpenBookKey      = "auto_open_book"
	legacyDefaultFormatKey     = "default_format_filter"
	legacyEnableIPFSKey        = "enable_ipfs_fallback"
)

// FyneRepository implements Repository backed by Fyne Preferences with automatic legacy migration.
type FyneRepository struct {
	preferences fyne.Preferences
	defaults    Settings
}

// NewFyneRepository creates a new Fyne-backed settings repository.
func NewFyneRepository(preferences fyne.Preferences, defaults Settings) *FyneRepository {
	return &FyneRepository{
		preferences: preferences,
		defaults:    defaults,
	}
}

// Load retrieves settings from preferences, migrating legacy keys if necessary.
func (r *FyneRepository) Load() (Settings, error) {
	if r.preferences == nil {
		return r.defaults, nil
	}

	// 1. Resolve Download Location / Storage URI with migration fallbacks
	downloadLocation := r.preferences.String(keyDownloadLocation)
	if downloadLocation == "" {
		downloadLocation = r.preferences.String(legacyDownloadDirectoryKey)
	}
	if downloadLocation == "" {
		downloadLocation = r.preferences.String(legacyDownloadFolderKey)
	}
	if downloadLocation == "" {
		downloadLocation = r.defaults.DownloadLocation
	}

	storageURI := r.preferences.StringWithFallback(keyStorageURI, r.defaults.StorageURI)
	storageDisplayName := r.preferences.StringWithFallback(keyStorageDisplayName, r.defaults.StorageDisplayName)
	if storageDisplayName == "" {
		if storageURI != "" {
			storageDisplayName = storageURI
		} else {
			storageDisplayName = downloadLocation
		}
	}

	// 2. Open after download
	openAfter := r.preferences.BoolWithFallback(keyOpenAfterDownload, r.defaults.OpenAfterDownload)
	if !r.preferences.BoolWithFallback(keyOpenAfterDownload, false) && r.preferences.BoolWithFallback(legacyAutoOpenBookKey, false) {
		openAfter = true
	}

	// 3. Existing file action
	existingFileActionStr := r.preferences.StringWithFallback(keyExistingFileAction, string(r.defaults.ExistingFileAction))
	existingFileAction := ExistingFileAction(existingFileActionStr)
	if existingFileAction == "" {
		existingFileAction = ExistingFileRename
	}

	// 4. Request timeout
	timeoutSec := r.preferences.IntWithFallback(keyRequestTimeoutSeconds, 0)
	if timeoutSec <= 0 {
		timeoutSec = r.preferences.IntWithFallback(legacyNetworkTimeoutSecKey, int(r.defaults.RequestTimeout.Seconds()))
	}
	if timeoutSec <= 0 {
		timeoutSec = 15
	}
	requestTimeout := time.Duration(timeoutSec) * time.Second

	// 5. Retry count
	retryCount := r.preferences.IntWithFallback(keyRetryCount, -1)
	if retryCount < 0 {
		retryCount = r.preferences.IntWithFallback(legacyMaxRetriesKey, r.defaults.RetryCount)
	}

	// 6. Enable IPFS
	enableIPFS := r.preferences.BoolWithFallback(keyEnableIPFS, r.defaults.EnableIPFS)
	if legacyIPFS := r.preferences.BoolWithFallback(legacyEnableIPFSKey, r.defaults.EnableIPFS); legacyIPFS != enableIPFS {
		enableIPFS = legacyIPFS
	}

	// 7. Mirrors
	mirrorModeStr := r.preferences.StringWithFallback(keyMirrorMode, string(r.defaults.MirrorMode))
	preferredMirror := r.preferences.StringWithFallback(keyPreferredMirror, r.defaults.PreferredMirror)
	if preferredMirror == "" {
		preferredMirror = r.preferences.StringWithFallback(legacyPreferredMirrorKey, "")
	}
	if preferredMirror != "" && preferredMirror != "auto" && mirrorModeStr == "" {
		mirrorModeStr = string(MirrorModePreferred)
	}
	mirrorMode := MirrorMode(mirrorModeStr)
	if mirrorMode == "" {
		mirrorMode = MirrorModeAutomatic
	}

	// 8. Theme
	themeStr := r.preferences.StringWithFallback(keyTheme, string(r.defaults.Theme))
	theme := Theme(themeStr)
	if theme == "" {
		theme = ThemeSystem
	}

	// 9. Format & AutoQueue
	defaultFormat := r.preferences.StringWithFallback(keyDefaultFormat, r.defaults.DefaultFormat)
	if legacyFmt := r.preferences.StringWithFallback(legacyDefaultFormatKey, ""); legacyFmt != "" {
		defaultFormat = legacyFmt
	}
	autoQueue := r.preferences.BoolWithFallback(keyAutoQueue, r.defaults.AutoQueue)
	if legacyAQ := r.preferences.BoolWithFallback(legacyAutoQueueKey, r.defaults.AutoQueue); legacyAQ != autoQueue {
		autoQueue = legacyAQ
	}

	value := Settings{
		DownloadLocation:   strings.TrimSpace(downloadLocation),
		StorageURI:         strings.TrimSpace(storageURI),
		StorageDisplayName: strings.TrimSpace(storageDisplayName),
		OpenAfterDownload:  openAfter,
		ExistingFileAction: existingFileAction,
		RequestTimeout:     requestTimeout,
		RetryCount:         retryCount,
		EnableIPFS:         enableIPFS,
		MirrorMode:         mirrorMode,
		PreferredMirror:    preferredMirror,
		Theme:              theme,
		DefaultFormat:      defaultFormat,
		AutoQueue:          autoQueue,
	}

	if err := value.Validate(); err != nil {
		return Settings{}, fmt.Errorf("invalid stored settings: %w", err)
	}

	return value, nil
}

// Save writes current settings to preferences.
func (r *FyneRepository) Save(value Settings) error {
	if err := value.Validate(); err != nil {
		return err
	}

	if r.preferences == nil {
		return nil
	}

	r.preferences.SetString(keyDownloadLocation, value.DownloadLocation)
	r.preferences.SetString(keyStorageURI, value.StorageURI)
	r.preferences.SetString(keyStorageDisplayName, value.StorageDisplayName)
	r.preferences.SetBool(keyOpenAfterDownload, value.OpenAfterDownload)
	r.preferences.SetString(keyExistingFileAction, string(value.ExistingFileAction))

	r.preferences.SetInt(keyRequestTimeoutSeconds, int(value.RequestTimeout.Seconds()))
	r.preferences.SetInt(keyRetryCount, value.RetryCount)
	r.preferences.SetBool(keyEnableIPFS, value.EnableIPFS)

	r.preferences.SetString(keyMirrorMode, string(value.MirrorMode))
	r.preferences.SetString(keyPreferredMirror, value.PreferredMirror)

	r.preferences.SetString(keyTheme, string(value.Theme))
	r.preferences.SetString(keyDefaultFormat, value.DefaultFormat)
	r.preferences.SetBool(keyAutoQueue, value.AutoQueue)

	// Keep legacy keys updated for seamless interop with any legacy UI components
	if value.DownloadLocation != "" {
		r.preferences.SetString(legacyDownloadFolderKey, value.DownloadLocation)
		r.preferences.SetString(legacyDownloadDirectoryKey, value.DownloadLocation)
	}
	if value.PreferredMirror != "" {
		r.preferences.SetString(legacyPreferredMirrorKey, value.PreferredMirror)
	}
	r.preferences.SetInt(legacyNetworkTimeoutSecKey, int(value.RequestTimeout.Seconds()))
	r.preferences.SetInt(legacyMaxRetriesKey, value.RetryCount)
	r.preferences.SetBool(legacyAutoOpenBookKey, value.OpenAfterDownload)
	r.preferences.SetBool(legacyAutoQueueKey, value.AutoQueue)
	r.preferences.SetBool(legacyEnableIPFSKey, value.EnableIPFS)
	if value.DefaultFormat != "" {
		r.preferences.SetString(legacyDefaultFormatKey, value.DefaultFormat)
	}

	return nil
}

// Reset restores settings to default configuration and saves.
func (r *FyneRepository) Reset() error {
	return r.Save(r.defaults)
}
