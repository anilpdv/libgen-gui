package settings

import (
	"errors"
	"time"
)

type MirrorMode string

const (
	MirrorModeAutomatic MirrorMode = "automatic"
	MirrorModePreferred MirrorMode = "preferred"
)

type ExistingFileAction string

const (
	ExistingFileRename    ExistingFileAction = "rename"
	ExistingFileOverwrite ExistingFileAction = "overwrite"
	ExistingFileSkip      ExistingFileAction = "skip"
)

type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

// Settings encapsulates typed, validated user preferences across platforms.
type Settings struct {
	// Downloads
	DownloadLocation   string             `json:"downloadLocation,omitempty"`
	StorageURI         string             `json:"storageURI,omitempty"`
	StorageDisplayName string             `json:"storageDisplayName,omitempty"`
	OpenAfterDownload  bool               `json:"openAfterDownload"`
	ExistingFileAction ExistingFileAction `json:"existingFileAction"`

	// Network
	RequestTimeout time.Duration `json:"requestTimeout"`
	RetryCount     int           `json:"retryCount"`
	EnableIPFS     bool          `json:"enableIPFS"`

	// Mirrors
	MirrorMode      MirrorMode `json:"mirrorMode"`
	PreferredMirror string     `json:"preferredMirror,omitempty"`

	// Appearance
	Theme Theme `json:"theme"`

	// Search / Format
	DefaultFormat string `json:"defaultFormat,omitempty"`
	AutoQueue     bool   `json:"autoQueue"`
}

// Default returns safe default configuration.
func Default() Settings {
	return Settings{
		OpenAfterDownload:  false,
		ExistingFileAction: ExistingFileRename,

		RequestTimeout: 15 * time.Second,
		RetryCount:     2,
		EnableIPFS:     true,

		MirrorMode: MirrorModeAutomatic,
		Theme:      ThemeSystem,

		DefaultFormat: "Any",
		AutoQueue:     true,
	}
}

// DefaultSettings is a convenience constructor for backwards compatibility.
func DefaultSettings(downloadLocation string) Settings {
	s := Default()
	s.DownloadLocation = downloadLocation
	s.StorageDisplayName = downloadLocation
	return s
}

// Validate ensures settings fall within acceptable application limits.
func (s Settings) Validate() error {
	if s.RequestTimeout < time.Second {
		return errors.New("request timeout must be at least 1 second")
	}

	if s.RequestTimeout > 10*time.Minute {
		return errors.New("request timeout cannot exceed 10 minutes")
	}

	if s.RetryCount < 0 || s.RetryCount > 10 {
		return errors.New("retry count must be between 0 and 10")
	}

	switch s.ExistingFileAction {
	case ExistingFileRename,
		ExistingFileOverwrite,
		ExistingFileSkip,
		"":
		// empty action is normalized to ExistingFileRename
	default:
		return errors.New("invalid existing-file action")
	}

	switch s.MirrorMode {
	case MirrorModeAutomatic, "":
	case MirrorModePreferred:
		if s.PreferredMirror == "" {
			return errors.New("preferred mirror is required in preferred mode")
		}
	default:
		return errors.New("invalid mirror mode")
	}

	switch s.Theme {
	case ThemeSystem, ThemeLight, ThemeDark, "":
	default:
		return errors.New("invalid theme")
	}

	return nil
}
