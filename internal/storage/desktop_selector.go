package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type selectionResult struct {
	location Location
	err      error
}

// DesktopLocationSelector implements LocationSelector using Fyne's folder open dialog.
type DesktopLocationSelector struct {
	parent fyne.Window
}

// NewDesktopLocationSelector creates a new DesktopLocationSelector.
func NewDesktopLocationSelector(parent fyne.Window) *DesktopLocationSelector {
	return &DesktopLocationSelector{
		parent: parent,
	}
}

func (s *DesktopLocationSelector) SelectLocation(ctx context.Context) (Location, error) {
	if s.parent == nil {
		return Location{}, ErrSelectionCancelled
	}

	resultChannel := make(chan selectionResult, 1)

	picker := dialog.NewFolderOpen(
		func(uri fyne.ListableURI, err error) {
			if err != nil {
				resultChannel <- selectionResult{err: err}
				return
			}

			if uri == nil {
				resultChannel <- selectionResult{
					err: ErrSelectionCancelled,
				}
				return
			}

			path := uri.Path()
			if path == "" {
				path = uri.String()
			}

			location := Location{
				Kind:        LocationDesktopPath,
				Path:        path,
				DisplayName: filepath.Clean(path),
			}

			if err := location.Validate(); err != nil {
				resultChannel <- selectionResult{err: err}
				return
			}

			resultChannel <- selectionResult{
				location: location,
			}
		},
		s.parent,
	)

	picker.Show()

	select {
	case <-ctx.Done():
		return Location{}, ctx.Err()

	case result := <-resultChannel:
		if result.err != nil {
			return Location{}, result.err
		}

		if err := ValidateDesktopDirectory(result.location.Path); err != nil {
			return Location{}, fmt.Errorf("selected folder is not writable: %w", err)
		}

		return result.location, nil
	}
}
