package settings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"libgen-gui/internal/storage"
)

// Controller coordinates settings persistence, location selection, validation, and default target provider updates.
type Controller struct {
	repository       Repository
	locationSelector storage.LocationSelector
	locationOpener   storage.LocationOpener
	targetProvider   storage.TargetProvider
}

// NewController creates a new settings Controller.
func NewController(
	repository Repository,
	locationSelector storage.LocationSelector,
	locationOpener storage.LocationOpener,
	targetProvider storage.TargetProvider,
) *Controller {
	return &Controller{
		repository:       repository,
		locationSelector: locationSelector,
		locationOpener:   locationOpener,
		targetProvider:   targetProvider,
	}
}

// Load returns currently saved settings.
func (c *Controller) Load() (Settings, error) {
	return c.repository.Load()
}

// ChangeDownloadLocation opens the platform folder picker, saves settings, and updates the target provider.
func (c *Controller) ChangeDownloadLocation(ctx context.Context) (storage.Location, error) {
	if c.locationSelector == nil {
		return storage.Location{}, errors.New("location selection is not supported on this platform")
	}

	location, err := c.locationSelector.SelectLocation(ctx)
	if err != nil {
		return storage.Location{}, err
	}

	if err := location.Validate(); err != nil {
		return storage.Location{}, err
	}

	current, err := c.repository.Load()
	if err != nil {
		return storage.Location{}, fmt.Errorf("load settings: %w", err)
	}

	previous := current

	switch location.Kind {
	case storage.LocationDesktopPath:
		current.DownloadLocation = location.Path
		current.StorageURI = ""
	case storage.LocationAndroidSAF:
		current.DownloadLocation = ""
		current.StorageURI = location.URI
	}

	current.StorageDisplayName = location.DisplayName

	if err := current.Validate(); err != nil {
		return storage.Location{}, err
	}

	// Save first, then update in-memory default. Roll back if activation fails.
	if err := c.repository.Save(current); err != nil {
		return storage.Location{}, fmt.Errorf("save download location: %w", err)
	}

	if c.targetProvider != nil {
		if err := c.targetProvider.Set(location); err != nil {
			_ = c.repository.Save(previous)
			return storage.Location{}, fmt.Errorf("activate download location: %w", err)
		}
	}

	return location, nil
}

// OpenDownloadLocation opens the active download directory in the system file manager.
func (c *Controller) OpenDownloadLocation(ctx context.Context) error {
	if c.targetProvider == nil {
		return errors.New("download location target provider not configured")
	}

	location, err := c.targetProvider.Current()
	if err != nil {
		return err
	}

	if c.locationOpener == nil {
		return errors.New("opening the download location is unsupported")
	}

	return c.locationOpener.OpenLocation(ctx, location)
}

// ResetToDefault restores default settings and updates the target provider.
func (c *Controller) ResetToDefault(defaultLocation storage.Location) error {
	defaults := Default()
	switch defaultLocation.Kind {
	case storage.LocationDesktopPath:
		defaults.DownloadLocation = defaultLocation.Path
	case storage.LocationAndroidSAF:
		defaults.StorageURI = defaultLocation.URI
	}
	defaults.StorageDisplayName = defaultLocation.DisplayName

	if err := c.repository.Save(defaults); err != nil {
		return err
	}

	if c.targetProvider != nil {
		_ = c.targetProvider.Set(defaultLocation)
	}
	return nil
}

// SetOpenAfterDownload toggles automatic book opening upon download completion.
func (c *Controller) SetOpenAfterDownload(enabled bool) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.OpenAfterDownload = enabled
	return c.repository.Save(current)
}

// SetExistingFileAction configures behavior when a file already exists.
func (c *Controller) SetExistingFileAction(action ExistingFileAction) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.ExistingFileAction = action
	return c.repository.Save(current)
}

// SetNetworkTimeout sets the HTTP request timeout.
func (c *Controller) SetNetworkTimeout(timeout time.Duration) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.RequestTimeout = timeout
	return c.repository.Save(current)
}

// SetRetryCount sets the max retry attempts.
func (c *Controller) SetRetryCount(retries int) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.RetryCount = retries
	return c.repository.Save(current)
}

// SetMirrorMode configures automatic vs preferred mirror selection.
func (c *Controller) SetMirrorMode(mode MirrorMode, preferredMirror string) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.MirrorMode = mode
	current.PreferredMirror = preferredMirror
	return c.repository.Save(current)
}

// SetTheme configures the UI theme.
func (c *Controller) SetTheme(theme Theme) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.Theme = theme
	return c.repository.Save(current)
}

// SetEnableIPFS toggles IPFS fallback download.
func (c *Controller) SetEnableIPFS(enabled bool) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.EnableIPFS = enabled
	return c.repository.Save(current)
}

// SetAutoQueue toggles auto-queueing.
func (c *Controller) SetAutoQueue(enabled bool) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.AutoQueue = enabled
	return c.repository.Save(current)
}

// SetDefaultFormat sets default format filter.
func (c *Controller) SetDefaultFormat(fmt string) error {
	current, err := c.repository.Load()
	if err != nil {
		return err
	}
	current.DefaultFormat = fmt
	return c.repository.Save(current)
}
