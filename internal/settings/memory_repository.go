package settings

import "sync"

// MemoryRepository provides an in-memory implementation of Repository for testing and headless execution.
type MemoryRepository struct {
	mu       sync.RWMutex
	settings Settings
	defaults Settings
}

// NewMemoryRepository creates a new in-memory repository with initial defaults.
func NewMemoryRepository(initial Settings) *MemoryRepository {
	return &MemoryRepository{
		settings: initial,
		defaults: initial,
	}
}

func (m *MemoryRepository) Load() (Settings, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings, nil
}

func (m *MemoryRepository) Save(s Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	return nil
}

func (m *MemoryRepository) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = m.defaults
	return nil
}
