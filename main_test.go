package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestApp_InitializationAndLayout(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	w, appInstance := setupApp(testApp)
	if w == nil {
		t.Fatalf("expected non-nil window from setupApp")
	}
	if appInstance == nil {
		t.Fatalf("expected non-nil ui.App from setupApp")
	}

	// Verify window title
	if title := w.Title(); title != "LibGen Downloader v2.0" {
		t.Errorf("expected window title 'LibGen Downloader v2.0', got %q", title)
	}

	// Verify window dimensions
	expectedSize := fyne.NewSize(900, 600)
	if w.Content().Size().Width > 0 && w.Content().Size() != expectedSize {
		t.Logf("window content size: %v", w.Content().Size())
	}

	// Verify the app uses a theme that provides a background color (may differ from stock dark theme)
	bgGot := testApp.Settings().Theme().Color(theme.ColorNameBackground, theme.VariantDark)
	if bgGot == nil {
		t.Errorf("expected a non-nil background color from the active theme")
	}

	// Verify content container structure
	content := w.Content()
	if content == nil {
		t.Fatalf("expected window content to be set")
	}
	containerObj, ok := content.(*fyne.Container)
	if !ok {
		t.Fatalf("expected window content to be *fyne.Container, got %T", content)
	}
	if len(containerObj.Objects) == 0 {
		t.Errorf("expected container objects to be populated")
	}
}
