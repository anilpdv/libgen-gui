package storage

import (
	"errors"
	"strings"
)

type LocationKind string

const (
	LocationDesktopPath LocationKind = "desktop-path"
	LocationAndroidSAF  LocationKind = "android-saf"
)

type Location struct {
	Kind        LocationKind `json:"kind"`
	Path        string       `json:"path,omitempty"`
	URI         string       `json:"uri,omitempty"`
	DisplayName string       `json:"displayName"`
}

func (l Location) Validate() error {
	switch l.Kind {
	case LocationDesktopPath:
		if strings.TrimSpace(l.Path) == "" {
			return errors.New("desktop storage path is empty")
		}

	case LocationAndroidSAF:
		if strings.TrimSpace(l.URI) == "" {
			return errors.New("Android SAF URI is empty")
		}

	default:
		return errors.New("unsupported storage location kind")
	}

	if strings.TrimSpace(l.DisplayName) == "" {
		return errors.New("storage display name is empty")
	}

	return nil
}
