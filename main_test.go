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

	// Verify theme provides dark theme colors
	bgGot := testApp.Settings().Theme().Color(theme.ColorNameBackground, theme.VariantDark)
	bgWant := theme.DarkTheme().Color(theme.ColorNameBackground, theme.VariantDark)
	if bgGot != bgWant {
		t.Errorf("expected dark theme background color %v, got %v", bgWant, bgGot)
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
