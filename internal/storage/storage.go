package storage

import (
	"context"
	"io"
	"path/filepath"
	"strings"
)

// Storage abstracts underlying file and storage providers (Desktop filesystem, Android SAF, Scoped Storage).
type Storage interface {
	// Create creates a new file for writing (truncating if exists).
	Create(ctx context.Context, filename string) (io.WriteCloser, error)

	// OpenAppend opens an existing file or creates a new one for resumable downloads. Returns writer and current byte size.
	OpenAppend(ctx context.Context, filename string) (io.WriteCloser, int64, error)

	// Exists checks if a file exists and is accessible.
	Exists(ctx context.Context, filename string) (bool, error)

	// Size returns the file size in bytes, or 0 if not found.
	Size(ctx context.Context, filename string) (int64, error)

	// Delete removes a file (e.g. on cancellation or cleanup).
	Delete(ctx context.Context, filename string) error

	// CommitPart atomically renames a temporary .part file to its final destination name.
	CommitPart(ctx context.Context, partFilename, finalFilename string) error

	// BaseLocation returns the human-readable root directory or URI.
	BaseLocation() string
}

// SanitizeFilename hardens titles/filenames against directory traversal and invalid OS characters.
func SanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.ReplaceAll(name, "\r", " ")
	name = strings.ReplaceAll(name, "\n", " ")
	name = strings.ReplaceAll(name, "\t", " ")

	// Strip directory traversal prefixes
	for strings.HasPrefix(name, "../") || strings.HasPrefix(name, "..\\") {
		name = strings.TrimPrefix(strings.TrimPrefix(name, "../"), "..\\")
	}
	for strings.HasPrefix(name, "./") || strings.HasPrefix(name, ".\\") {
		name = strings.TrimPrefix(strings.TrimPrefix(name, "./"), ".\\")
	}

	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", " -",
		"*", "_",
		"?", "_",
		"\"", "'",
		"<", "_",
		">", "_",
		"|", "_",
	)
	name = replacer.Replace(name)

	// Clean double spaces
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}

	name = filepath.Base(name)
	name = strings.TrimLeft(name, "._- ")

	if name == "" {
		return "download.bin"
	}

	// Maximum length constraint for filesystem compatibility (e.g. eCryptfs, FAT32)
	const maxLen = 180
	runes := []rune(name)
	if len(runes) > maxLen {
		ext := filepath.Ext(name)
		stemLen := maxLen - len([]rune(ext))
		if stemLen > 0 {
			name = string(runes[:stemLen]) + ext
		} else {
			name = string(runes[:maxLen])
		}
	}

	return strings.TrimSpace(name)
}
