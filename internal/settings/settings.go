package settings

import (
	"errors"
	"strings"
	"sync"
	"time"
)

// Default settings values.
const (
	DefaultPreferredMirror     = "auto"
	DefaultNetworkTimeoutSec   = 15
	DefaultMaxRetries          = 2
	DefaultAutoQueue           = true
	DefaultAutoOpenBook        = false
	DefaultShowCompletionModal = true
	DefaultFormatFilter        = "Any"
	DefaultEnableIPFS          = true
	DefaultTheme               = "dark"
)

// Settings encapsulates typed, validated user preferences.
type Settings struct {
	DownloadLocation    string        `json:"download_location"`
	PreferredMirror     string        `json:"preferred_mirror"`
	Theme               string        `json:"theme"`
	RequestTimeout      time.Duration `json:"request_timeout"`
	MaxRetries          int           `json:"max_retries"`
	AutoQueue           bool          `json:"auto_queue"`
	AutoOpenBook        bool          `json:"auto_open_book"`
	ShowCompletionModal bool          `json:"show_completion_modal"`
	DefaultFormat       string        `json:"default_format"`
	EnableIPFS          bool          `json:"enable_ipfs"`
}

// DefaultSettings returns safe default configuration.
func DefaultSettings(downloadLocation string) Settings {
	return Settings{
		DownloadLocation:    downloadLocation,
		PreferredMirror:     DefaultPreferredMirror,
		Theme:               DefaultTheme,
		RequestTimeout:      time.Duration(DefaultNetworkTimeoutSec) * time.Second,
		MaxRetries:          DefaultMaxRetries,
		AutoQueue:           DefaultAutoQueue,
		AutoOpenBook:        DefaultAutoOpenBook,
		ShowCompletionModal: DefaultShowCompletionModal,
		DefaultFormat:       DefaultFormatFilter,
		EnableIPFS:          DefaultEnableIPFS,
	}
}

// Validate ensures settings fall within acceptable application limits.
func (s Settings) Validate() error {
	if s.RequestTimeout < 1*time.Second {
		return errors.New("request timeout must be at least 1 second")
	}
	if s.RequestTimeout > 5*time.Minute {
		return errors.New("request timeout cannot exceed 5 minutes")
	}
	if s.MaxRetries < 0 || s.MaxRetries > 10 {
		return errors.New("max retries must be between 0 and 10")
	}
	if strings.TrimSpace(s.DownloadLocation) == "" {
		return errors.New("download location cannot be empty")
	}
	return nil
}

// SettingsService interface decouples business logic from specific persistence backends (e.g. Fyne preferences).
type SettingsService interface {
	GetSettings() Settings
	UpdateSettings(s Settings) error
	ResetSettings() error
}

// MemorySettingsService provides an in-memory implementation for headless testing.
type MemorySettingsService struct {
	mu       sync.RWMutex
	settings Settings
}

// NewMemorySettingsService creates an in-memory settings service.
func NewMemorySettingsService(initial Settings) *MemorySettingsService {
	return &MemorySettingsService{
		settings: initial,
	}
}

func (m *MemorySettingsService) GetSettings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

func (m *MemorySettingsService) UpdateSettings(s Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	return nil
}

func (m *MemorySettingsService) ResetSettings() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = DefaultSettings(m.settings.DownloadLocation)
	return nil
}
