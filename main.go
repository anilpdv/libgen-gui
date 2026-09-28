package main

import (
	_ "embed"
	"log"

	"fyne.io/fyne/v2"
	internalApp "libgen-gui/internal/app"
	"libgen-gui/ui"
)

//go:embed icon.png
var appIconBytes []byte

// setupApp initializes window, theme, icons, and UI content hierarchy without blocking.
func setupApp(a fyne.App) (fyne.Window, *ui.App) {
	return internalApp.Setup(a, appIconBytes)
}

func main() {
	if err := internalApp.Run(appIconBytes); err != nil {
		log.Fatal(err)
	}
}
