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
	Search             search.SearchService
	Downloads          download.DownloadService
	Settings           settings.SettingsService
	SettingsController *settings.Controller
	Mirrors            *network.MirrorManager
	Storage            storage.Storage
	TargetProvider     storage.TargetProvider
	StorageFactory     storage.Factory
}

// BuildDefaultDependencies initializes all core infrastructure and services.
func BuildDefaultDependencies(isMobile bool) Dependencies {
	httpClient := network.NewClient()
	mirrorMgr := network.NewMirrorManager(httpClient, network.DefaultSearchMirrors)

	defaultLoc, _ := storage.ResolveDefaultLocation(isMobile)
	defaultSettings := settings.Default()
	if defaultLoc.Kind == storage.LocationAndroidSAF {
		defaultSettings.StorageURI = defaultLoc.URI
		defaultSettings.StorageDisplayName = defaultLoc.DisplayName
	} else {
		defaultSettings.DownloadLocation = defaultLoc.Path
		defaultSettings.StorageDisplayName = defaultLoc.DisplayName
	}

	appInstance := fyne.CurrentApp()
	var settingsRepo settings.Repository
	if appInstance != nil && appInstance.Preferences() != nil {
		settingsRepo = settings.NewFyneRepository(appInstance.Preferences(), defaultSettings)
	} else {
		settingsRepo = settings.NewMemoryRepository(defaultSettings)
	}

	loadedSettings, _ := settingsRepo.Load()
	initialLoc := storage.Location{
		Kind:        storage.LocationDesktopPath,
		Path:        loadedSettings.DownloadLocation,
		URI:         loadedSettings.StorageURI,
		DisplayName: loadedSettings.StorageDisplayName,
	}
	if loadedSettings.StorageURI != "" {
		initialLoc.Kind = storage.LocationAndroidSAF
	}
	if initialLoc.DisplayName == "" {
		if initialLoc.Path != "" {
			initialLoc.DisplayName = initialLoc.Path
		} else {
			initialLoc = defaultLoc
		}
	}

	targetProvider, _ := storage.NewMutableTargetProvider(initialLoc)
	storageFactory := storage.NewStorageFactory()

	storageService, _ := storageFactory.For(initialLoc)

	queueStorePath := filepath.Join(defaultLoc.Path, ".queue.json")
	if defaultLoc.Path == "" {
		queueStorePath = ""
	}
	queueStore := download.NewJSONQueueStoreWithTarget(queueStorePath, initialLoc)

	downloadMgr := download.NewManagerWithDependencies(download.Dependencies{
		HTTPClient:     httpClient,
		StorageFactory: storageFactory,
		TargetProvider: targetProvider,
		MirrorManager:  mirrorMgr,
		QueueStore:     queueStore,
	})

	searchSvc := search.NewService(mirrorMgr, httpClient)
	settingsController := settings.NewController(settingsRepo, nil, storage.NewDesktopLocationOpener(), targetProvider)
	settingsSvc := settings.NewServiceAdapter(settingsRepo)

	return Dependencies{
		Search:             searchSvc,
		Downloads:          downloadMgr,
		Settings:           settingsSvc,
		SettingsController: settingsController,
		Mirrors:            mirrorMgr,
		Storage:            storageService,
		TargetProvider:     targetProvider,
		StorageFactory:     storageFactory,
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
