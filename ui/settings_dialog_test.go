package ui

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSettingsDialog_RenderTabsAndActions(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := app.NewWindow("TestSettings")
	defer w.Close()

	guiApp := NewApp(w)
	_ = guiApp.Build()

	// 1. Test opening each tab of the settings dialog
	for tabIdx := 0; tabIdx < 4; tabIdx++ {
		ShowFullSettingsDialog(guiApp, tabIdx)
	}

	// 2. Test ShowSettingsDialog and ShowMirrorsDialog wrappers
	guiApp.ShowSettingsDialog()
	guiApp.ShowMirrorsDialog()

	// 3. Test SearchBar mirror badge updating
	sb := guiApp.searchBar
	if sb == nil {
		t.Fatalf("searchBar is nil")
	}

	// Test full availability
	sb.updateMirrorStatusBadge(2, 2)
	if sb.settingsBtn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance for 2/2 mirrors, got %v", sb.settingsBtn.Importance)
	}

	// Test partial availability
	sb.updateMirrorStatusBadge(1, 2)
	if sb.settingsBtn.Importance != widget.WarningImportance {
		t.Errorf("expected WarningImportance for 1/2 mirrors, got %v", sb.settingsBtn.Importance)
	}

	// Test zero availability
	sb.updateMirrorStatusBadge(0, 2)
	if sb.settingsBtn.Importance != widget.DangerImportance {
		t.Errorf("expected DangerImportance for 0/2 mirrors, got %v", sb.settingsBtn.Importance)
	}

	// Mobile format
	mobile := true
	guiApp.mobileOverride = &mobile
	sb.updateMirrorStatusBadge(2, 2)
	if text := sb.settingsBtn.Text; text != "2/2" {
		t.Errorf("expected mobile text '2/2', got %q", text)
	}

	desktop := false
	guiApp.mobileOverride = &desktop
	sb.updateMirrorStatusBadge(2, 2)
	if text := sb.settingsBtn.Text; text != "Mirrors: 2/2" {
		t.Errorf("expected desktop text 'Mirrors: 2/2', got %q", text)
	}
}

func TestMirrorHealthManager_LiveProbe(t *testing.T) {
	mgr := GetMirrorHealthManager()
	if mgr == nil {
		t.Fatalf("default health manager is nil")
	}

	mirrors := mgr.GetMirrors()
	if len(mirrors) == 0 {
		t.Fatalf("expected default mirrors configured, got 0")
	}

	// Fast probe with short timeout to ensure no hang
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mgr.ProbeAll(ctx)
	active, total := mgr.GetActiveSearchCount()
	if total == 0 {
		t.Errorf("expected total search mirrors > 0, got %d", total)
	}
	t.Logf("Live probe completed: %d / %d active search mirrors", active, total)
}
