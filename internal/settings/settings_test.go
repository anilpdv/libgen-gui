package settings

import (
	"testing"
	"time"
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
	invalidRetries.MaxRetries = 20
	if err := invalidRetries.Validate(); err == nil {
		t.Errorf("expected error for retries > 10, got nil")
	}

	emptyLocation := valid
	emptyLocation.DownloadLocation = "   "
	if err := emptyLocation.Validate(); err == nil {
		t.Errorf("expected error for empty download location, got nil")
	}
}

func TestMemorySettingsService_CRUD(t *testing.T) {
	svc := NewMemorySettingsService(DefaultSettings("/tmp/dl"))

	current := svc.GetSettings()
	if current.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected default max retries %d, got %d", DefaultMaxRetries, current.MaxRetries)
	}

	current.MaxRetries = 5
	if err := svc.UpdateSettings(current); err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	updated := svc.GetSettings()
	if updated.MaxRetries != 5 {
		t.Errorf("expected updated retries to be 5, got %d", updated.MaxRetries)
	}

	_ = svc.ResetSettings()
	reset := svc.GetSettings()
	if reset.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected reset retries to be %d, got %d", DefaultMaxRetries, reset.MaxRetries)
	}
}
