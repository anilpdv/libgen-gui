package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SAFStorage implements Storage for Android Storage Access Framework (SAF) and Scoped Storage.
type SAFStorage struct {
	rootURI  string
	fallback *FileSystemStorage
	mu       sync.RWMutex
}

// NewSAFStorage constructs an Android storage adapter.
func NewSAFStorage(rootURI string) (*SAFStorage, error) {
	if rootURI == "" {
		rootURI = "/storage/emulated/0/Download"
	}

	cleanDir := rootURI
	if strings.HasPrefix(rootURI, "file://") {
		cleanDir = strings.TrimPrefix(rootURI, "file://")
	}

	fs, err := NewFileSystemStorage(cleanDir)
	if err != nil {
		// Fallback to temp dir if external storage not writable yet
		fs, _ = NewFileSystemStorage(os.TempDir())
	}

	return &SAFStorage{
		rootURI:  rootURI,
		fallback: fs,
	}, nil
}

func (s *SAFStorage) Create(ctx context.Context, filename string) (io.WriteCloser, error) {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.Create(ctx, filename)
	}
	return nil, fmt.Errorf("SAF storage backend unavailable for %s", filename)
}

func (s *SAFStorage) OpenAppend(ctx context.Context, filename string) (io.WriteCloser, int64, error) {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.OpenAppend(ctx, filename)
	}
	return nil, 0, fmt.Errorf("SAF storage backend unavailable for %s", filename)
}

func (s *SAFStorage) Exists(ctx context.Context, filename string) (bool, error) {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.Exists(ctx, filename)
	}
	return false, nil
}

func (s *SAFStorage) Size(ctx context.Context, filename string) (int64, error) {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.Size(ctx, filename)
	}
	return 0, nil
}

func (s *SAFStorage) Delete(ctx context.Context, filename string) error {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.Delete(ctx, filename)
	}
	return nil
}

func (s *SAFStorage) CommitPart(ctx context.Context, partFilename, finalFilename string) error {
	s.mu.RLock()
	fb := s.fallback
	s.mu.RUnlock()

	if fb != nil {
		return fb.CommitPart(ctx, partFilename, finalFilename)
	}
	return fmt.Errorf("SAF commit failed for %s", finalFilename)
}

func (s *SAFStorage) BaseLocation() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rootURI
}

// NewStorage creates the appropriate Storage provider based on platform and path.
func NewStorage(pathOrURI string, isMobile bool) (Storage, error) {
	if isMobile || strings.HasPrefix(pathOrURI, "content://") {
		return NewSAFStorage(pathOrURI)
	}
	return NewFileSystemStorage(filepath.Clean(pathOrURI))
}
