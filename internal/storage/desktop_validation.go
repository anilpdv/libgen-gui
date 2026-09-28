package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ValidateDesktopDirectory checks that path exists, is a directory, and is writable.
func ValidateDesktopDirectory(path string) error {
	if path == "" {
		return errors.New("folder path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("inspect selected folder: %w", err)
	}

	if !info.IsDir() {
		return errors.New("selected location is not a folder")
	}

	probe, err := os.CreateTemp(path, ".libgen-write-test-*")
	if err != nil {
		return fmt.Errorf("folder is not writable: %w", err)
	}

	probeName := probe.Name()
	closeErr := probe.Close()
	removeErr := os.Remove(probeName)

	if closeErr != nil {
		return fmt.Errorf("close write test: %w", closeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("remove write test: %w", removeErr)
	}

	return nil
}

// ResolveDefaultLocation determines a valid, writable default Location for the platform.
func ResolveDefaultLocation(isMobile bool) (Location, error) {
	if isMobile || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		candidates := []string{
			"/storage/emulated/0/Download",
			"/sdcard/Download",
			"/storage/emulated/0/Android/data/com.libgen.downloader/files/Download",
		}

		for _, cand := range candidates {
			if ValidateDesktopDirectory(cand) == nil {
				return Location{
					Kind:        LocationAndroidSAF,
					URI:         cand,
					DisplayName: "Public Downloads",
				}, nil
			}
		}

		return Location{
			Kind:        LocationAndroidSAF,
			URI:         "/storage/emulated/0/Download",
			DisplayName: "Downloads",
		}, nil
	}

	// Desktop candidates (macOS, Windows, Linux)
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		dl := filepath.Join(home, "Downloads")
		if ValidateDesktopDirectory(dl) == nil {
			return Location{
				Kind:        LocationDesktopPath,
				Path:        dl,
				DisplayName: filepath.Clean(dl),
			}, nil
		}
		desktop := filepath.Join(home, "Desktop")
		if ValidateDesktopDirectory(desktop) == nil {
			return Location{
				Kind:        LocationDesktopPath,
				Path:        desktop,
				DisplayName: filepath.Clean(desktop),
			}, nil
		}
		if ValidateDesktopDirectory(home) == nil {
			return Location{
				Kind:        LocationDesktopPath,
				Path:        home,
				DisplayName: filepath.Clean(home),
			}, nil
		}
	}

	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("LibGenDownloads_%d", time.Now().UnixNano()))
	_ = os.MkdirAll(tmp, 0755)
	if ValidateDesktopDirectory(tmp) == nil {
		return Location{
			Kind:        LocationDesktopPath,
			Path:        tmp,
			DisplayName: filepath.Clean(tmp),
		}, nil
	}

	return Location{
		Kind:        LocationDesktopPath,
		Path:        os.TempDir(),
		DisplayName: os.TempDir(),
	}, nil
}
