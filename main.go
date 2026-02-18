package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"libgen-gui/ui"
)

//go:embed icon.png
var appIconBytes []byte

// setupApp initializes window, theme, icons, and UI content hierarchy without blocking.
func setupApp(a fyne.App) (fyne.Window, *ui.App) {
	if len(appIconBytes) > 0 {
		a.SetIcon(fyne.NewStaticResource("icon.png", appIconBytes))
	}
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("LibGen Downloader v2.0")
	if len(appIconBytes) > 0 {
		w.SetIcon(fyne.NewStaticResource("icon.png", appIconBytes))
	}
	w.Resize(fyne.NewSize(900, 600))
	w.SetFixedSize(false)

	libgenApp := ui.NewApp(w)
	w.SetContent(libgenApp.Build())
	return w, libgenApp
}

func main() {
	a := app.NewWithID("com.libgen.downloader")
	w, _ := setupApp(a)
	w.ShowAndRun()
}

