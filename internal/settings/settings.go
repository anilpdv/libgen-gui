package settings

import (
	"strings"
	"sync"
)

// SettingsService interface decouples business logic from specific persistence backends.
type SettingsService interface {
	GetSettings() Settings
	UpdateSettings(s Settings) error
	ResetSettings() error
}

// MemorySettingsService provides an in-memory implementation for headless testing and backwards compatibility.
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
	loc := m.settings.DownloadLocation
	if loc == "" {
		loc = m.settings.StorageDisplayName
	}
	m.settings = DefaultSettings(loc)
	return nil
}

// ServiceAdapter adapts Repository to SettingsService.
type ServiceAdapter struct {
	repo Repository
}

// NewServiceAdapter wraps a Repository into a SettingsService.
func NewServiceAdapter(repo Repository) *ServiceAdapter {
	return &ServiceAdapter{repo: repo}
}

func (a *ServiceAdapter) GetSettings() Settings {
	s, err := a.repo.Load()
	if err != nil {
		return Default()
	}
	return s
}

func (a *ServiceAdapter) UpdateSettings(s Settings) error {
	if strings.TrimSpace(s.DownloadLocation) == "" && strings.TrimSpace(s.StorageURI) == "" {
		s.DownloadLocation = s.StorageDisplayName
	}
	return a.repo.Save(s)
}

func (a *ServiceAdapter) ResetSettings() error {
	return a.repo.Reset()
}
