package storage

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Factory creates storage backends for specific Locations.
type Factory interface {
	For(Location) (Storage, error)
}

// StorageFactory implements Factory caching or instantiating Storage providers.
type StorageFactory struct {
	mu    sync.RWMutex
	cache map[string]Storage
}

// NewStorageFactory creates a new StorageFactory.
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{
		cache: make(map[string]Storage),
	}
}

// For returns or constructs a Storage implementation appropriate for the given Location.
func (f *StorageFactory) For(loc Location) (Storage, error) {
	if err := loc.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage location: %w", err)
	}

	key := string(loc.Kind) + ":" + loc.Path + ":" + loc.URI

	f.mu.RLock()
	existing, ok := f.cache[key]
	f.mu.RUnlock()
	if ok {
		return existing, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if existing, ok := f.cache[key]; ok {
		return existing, nil
	}

	var s Storage
	var err error

	switch loc.Kind {
	case LocationDesktopPath:
		s, err = NewFileSystemStorage(filepath.Clean(loc.Path))
	case LocationAndroidSAF:
		s, err = NewSAFStorage(loc.URI)
	default:
		if strings.HasPrefix(loc.URI, "content://") || loc.URI != "" {
			s, err = NewSAFStorage(loc.URI)
		} else {
			s, err = NewFileSystemStorage(filepath.Clean(loc.Path))
		}
	}

	if err != nil {
		return nil, err
	}

	f.cache[key] = s
	return s, nil
}
