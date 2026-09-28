package app

import (
	_ "embed"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"libgen-gui/internal/download"
	"libgen-gui/internal/network"
	"libgen-gui/internal/search"
	"libgen-gui/internal/settings"
	"libgen-gui/internal/storage"
	"libgen-gui/ui"
)

// Dependencies bundles all domain services for dependency injection.
type Dependencies struct {
	Search    search.SearchService
	Downloads download.DownloadService
	Settings  settings.SettingsService
	Mirrors   *network.MirrorManager
	Storage   storage.Storage
}

// BuildDefaultDependencies initializes all core infrastructure and services.
func BuildDefaultDependencies(isMobile bool) Dependencies {
	httpClient := network.NewClient()
	mirrorMgr := network.NewMirrorManager(httpClient, network.DefaultSearchMirrors)

	savePath := ui.GetConfiguredSavePath(isMobile)
	storageService, _ := storage.NewStorage(savePath, isMobile)

	queueStorePath := filepath.Join(savePath, ".queue.json")
	queueStore := download.NewJSONQueueStore(queueStorePath)

	downloadMgr := download.NewManager(httpClient, storageService, mirrorMgr, queueStore)
	searchSvc := search.NewService(mirrorMgr, httpClient)
	settingsSvc := settings.NewMemorySettingsService(settings.DefaultSettings(savePath))

	return Dependencies{
		Search:    searchSvc,
		Downloads: downloadMgr,
		Settings:  settingsSvc,
		Mirrors:   mirrorMgr,
		Storage:   storageService,
	}
}

// Setup initializes the window, theme, icons, and UI content hierarchy without blocking.
func Setup(a fyne.App, iconBytes []byte) (fyne.Window, *ui.App) {
	if len(iconBytes) > 0 {
		a.SetIcon(fyne.NewStaticResource("icon.png", iconBytes))
	}
	a.Settings().SetTheme(&ui.ModernTheme{})

	w := a.NewWindow("LibGen Downloader v2.0")
	if len(iconBytes) > 0 {
		w.SetIcon(fyne.NewStaticResource("icon.png", iconBytes))
	}
	w.Resize(fyne.NewSize(900, 600))
	w.SetFixedSize(false)

	libgenApp := ui.NewApp(w)
	w.SetContent(libgenApp.Build())
	return w, libgenApp
}

// Run is the main application entrypoint constructing services and running the GUI.
func Run(iconBytes []byte) error {
	a := app.NewWithID("com.libgen.downloader")
	w, _ := Setup(a, iconBytes)
	w.ShowAndRun()
	return nil
}
