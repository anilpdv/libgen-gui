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

// FileSystemStorage implements Storage for desktop platforms (macOS, Linux, Windows).
type FileSystemStorage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileSystemStorage creates a FileSystemStorage instance for baseDir.
func NewFileSystemStorage(baseDir string) (*FileSystemStorage, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			baseDir = os.TempDir()
		} else {
			baseDir = filepath.Join(home, "Downloads")
		}
	}

	cleanDir := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %q: %w", cleanDir, err)
	}

	return &FileSystemStorage{
		baseDir: cleanDir,
	}, nil
}

func (fs *FileSystemStorage) resolvePath(filename string) (string, error) {
	fs.mu.RLock()
	base := fs.baseDir
	fs.mu.RUnlock()

	sanitized := SanitizeFilename(filename)
	fullPath := filepath.Join(base, sanitized)

	// Guard against directory traversal attacks
	cleanFull := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanFull, base) {
		return "", fmt.Errorf("path traversal violation for filename %q", filename)
	}

	return cleanFull, nil
}

func (fs *FileSystemStorage) Create(ctx context.Context, filename string) (io.WriteCloser, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	target, err := fs.resolvePath(filename)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

func (fs *FileSystemStorage) OpenAppend(ctx context.Context, filename string) (io.WriteCloser, int64, error) {
	if ctx.Err() != nil {
		return nil, 0, ctx.Err()
	}

	target, err := fs.resolvePath(filename)
	if err != nil {
		return nil, 0, err
	}

	var existingSize int64
	if fi, err := os.Stat(target); err == nil {
		existingSize = fi.Size()
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}

	return f, existingSize, nil
}

func (fs *FileSystemStorage) Exists(ctx context.Context, filename string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	target, err := fs.resolvePath(filename)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(target)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (fs *FileSystemStorage) Size(ctx context.Context, filename string) (int64, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	target, err := fs.resolvePath(filename)
	if err != nil {
		return 0, err
	}

	fi, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	return fi.Size(), nil
}

func (fs *FileSystemStorage) Delete(ctx context.Context, filename string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	target, err := fs.resolvePath(filename)
	if err != nil {
		return err
	}

	err = os.Remove(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (fs *FileSystemStorage) CommitPart(ctx context.Context, partFilename, finalFilename string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	partPath, err := fs.resolvePath(partFilename)
	if err != nil {
		return err
	}

	finalPath, err := fs.resolvePath(finalFilename)
	if err != nil {
		return err
	}

	// Atomic rename of .part file to final destination
	return os.Rename(partPath, finalPath)
}

func (fs *FileSystemStorage) BaseLocation() string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.baseDir
}
