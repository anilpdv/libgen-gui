package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

func TestSettingsPreferences_LifecycleAndDefaults(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	// 1. Check default settings loading
	s := LoadAppSettings(false)
	if s.PreferredMirror != DefaultPreferredMirror {
		t.Errorf("expected default mirror %q, got %q", DefaultPreferredMirror, s.PreferredMirror)
	}
	if s.NetworkTimeout != time.Duration(DefaultNetworkTimeoutSec)*time.Second {
		t.Errorf("expected default timeout %vs, got %v", DefaultNetworkTimeoutSec, s.NetworkTimeout)
	}
	if s.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected max retries %d, got %d", DefaultMaxRetries, s.MaxRetries)
	}
	if !s.AutoQueue {
		t.Errorf("expected auto-queue true by default")
	}
	if s.AutoOpenBook {
		t.Errorf("expected auto-open false by default")
	}
	if !s.EnableIPFS {
		t.Errorf("expected enable IPFS true by default")
	}

	// 2. Modify preferences
	SavePreferredMirror("libgen.li")
	if got := GetPreferredMirror(); got != "libgen.li" {
		t.Errorf("expected preferred mirror libgen.li, got %q", got)
	}

	SaveNetworkTimeoutSec(30)
	if got := GetNetworkTimeout(); got != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", got)
	}

	SaveMaxRetries(5)
	if got := GetMaxRetries(); got != 5 {
		t.Errorf("expected retries 5, got %d", got)
	}

	SaveAutoQueue(false)
	if got := GetAutoQueue(); got != false {
		t.Errorf("expected auto-queue false, got true")
	}

	SaveAutoOpenBook(true)
	if got := GetAutoOpenBook(); got != true {
		t.Errorf("expected auto-open true, got false")
	}

	SaveDefaultFormatFilter("epub")
	if got := GetDefaultFormatFilter(); got != "epub" {
		t.Errorf("expected format filter epub, got %q", got)
	}

	SaveEnableIPFS(false)
	if got := GetEnableIPFS(); got != false {
		t.Errorf("expected enable IPFS false, got true")
	}

	// Verify LoadAppSettings reflects changes
	updated := LoadAppSettings(false)
	if updated.PreferredMirror != "libgen.li" || updated.MaxRetries != 5 || !updated.AutoOpenBook {
		t.Errorf("updated settings mismatch: %+v", updated)
	}

	// 3. Reset All Settings
	ResetAllSettings(false)
	resetSettings := LoadAppSettings(false)
	if resetSettings.PreferredMirror != DefaultPreferredMirror {
		t.Errorf("after reset: expected preferred mirror %q, got %q", DefaultPreferredMirror, resetSettings.PreferredMirror)
	}
	if resetSettings.NetworkTimeout != time.Duration(DefaultNetworkTimeoutSec)*time.Second {
		t.Errorf("after reset: expected timeout %vs, got %v", DefaultNetworkTimeoutSec, resetSettings.NetworkTimeout)
	}
	if resetSettings.AutoOpenBook != DefaultAutoOpenBook {
		t.Errorf("after reset: expected auto-open false, got %v", resetSettings.AutoOpenBook)
	}
}
