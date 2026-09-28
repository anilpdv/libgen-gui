package storage

import (
	"context"
	"errors"
)

// ErrSelectionCancelled is returned when the user cancels folder/location selection.
var ErrSelectionCancelled = errors.New("storage location selection cancelled")

// ErrStoragePermissionRevoked is returned when Android SAF permission has been revoked.
var ErrStoragePermissionRevoked = errors.New("storage permission has been revoked")

// LocationSelector abstracts GUI folder selection across desktop and mobile.
type LocationSelector interface {
	SelectLocation(context.Context) (Location, error)
}

// LocationOpener abstracts opening the download folder in the OS file manager.
type LocationOpener interface {
	OpenLocation(context.Context, Location) error
}

// LocationValidator checks whether a storage location is currently accessible.
type LocationValidator interface {
	ValidateLocation(context.Context, Location) error
}
