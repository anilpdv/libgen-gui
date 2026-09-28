package settings

import (
	"context"
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"libgen-gui/internal/storage"
)

func TestSettings_Validation(t *testing.T) {
	valid := DefaultSettings("/tmp/downloads")
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected default settings to be valid, got %v", err)
	}

	invalidTimeout := valid
	invalidTimeout.RequestTimeout = 500 * time.Millisecond
	if err := invalidTimeout.Validate(); err == nil {
		t.Errorf("expected error for timeout < 1s, got nil")
	}

	invalidRetries := valid
	invalidRetries.RetryCount = 20
	if err := invalidRetries.Validate(); err == nil {
		t.Errorf("expected error for retries > 10, got nil")
	}
}

func TestRepositoryPersistsDownloadLocation(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	defaults := Default()
	defaults.DownloadLocation = "/default"
	defaults.StorageDisplayName = "/default"

	repository := NewFyneRepository(app.Preferences(), defaults)

	value := defaults
	value.DownloadLocation = "/tmp/books"
	value.StorageDisplayName = "/tmp/books"

	if err := repository.Save(value); err != nil {
		t.Fatal(err)
	}

	loaded, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.DownloadLocation != "/tmp/books" {
		t.Fatalf("expected /tmp/books, got %q", loaded.DownloadLocation)
	}
}

func TestRepositoryPersistsStorageURI(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	defaults := Default()
	repository := NewFyneRepository(app.Preferences(), defaults)

	value := defaults
	value.DownloadLocation = ""
	value.StorageURI = "content://example/tree/books"
	value.StorageDisplayName = "Books"

	if err := repository.Save(value); err != nil {
		t.Fatal(err)
	}

	loaded, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.StorageURI != value.StorageURI {
		t.Fatalf("expected URI %q, got %q", value.StorageURI, loaded.StorageURI)
	}
}

type fakeSelector struct {
	location storage.Location
	err      error
}

func (f *fakeSelector) SelectLocation(ctx context.Context) (storage.Location, error) {
	return f.location, f.err
}

type recordingRepository struct {
	settings  Settings
	saveCount int
}

func (r *recordingRepository) Load() (Settings, error) {
	return r.settings, nil
}

func (r *recordingRepository) Save(s Settings) error {
	r.saveCount++
	r.settings = s
	return nil
}

func (r *recordingRepository) Reset() error {
	return nil
}

func TestChangeLocationDoesNotSaveWhenCancelled(t *testing.T) {
	repo := &recordingRepository{
		settings: Default(),
	}
	selector := &fakeSelector{
		err: storage.ErrSelectionCancelled,
	}

	provider, _ := storage.NewMutableTargetProvider(storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        "/initial",
		DisplayName: "/initial",
	})

	controller := NewController(repo, selector, nil, provider)

	_, err := controller.ChangeDownloadLocation(context.Background())
	if !errors.Is(err, storage.ErrSelectionCancelled) {
		t.Fatalf("expected cancellation error, got %v", err)
	}

	if repo.saveCount != 0 {
		t.Fatal("settings were saved after selection cancellation")
	}
}

func TestController_ChangeLocationSuccess(t *testing.T) {
	repo := &recordingRepository{
		settings: Default(),
	}
	selector := &fakeSelector{
		location: storage.Location{
			Kind:        storage.LocationDesktopPath,
			Path:        "/new/folder",
			DisplayName: "/new/folder",
		},
	}
	provider, _ := storage.NewMutableTargetProvider(storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        "/initial",
		DisplayName: "/initial",
	})

	controller := NewController(repo, selector, nil, provider)
	loc, err := controller.ChangeDownloadLocation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loc.Path != "/new/folder" {
		t.Errorf("expected location /new/folder, got %s", loc.Path)
	}

	cur, _ := provider.Current()
	if cur.Path != "/new/folder" {
		t.Errorf("expected target provider to have /new/folder, got %s", cur.Path)
	}
}
