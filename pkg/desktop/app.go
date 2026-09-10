package desktop

import (
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"fileshare/web"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Setup systray
	go setupSystray(a)

	// Bind events for JS
	runtime.EventsOn(ctx, "minimize_to_drawer", func(optionalData ...interface{}) {
		a.EnableDrawerMode()
	})
	runtime.EventsOn(ctx, "restore_ui", func(optionalData ...interface{}) {
		a.EnableFullUI()
	})
}

// EnableDrawerMode shrinks the window to the top-right corner
func (a *App) EnableDrawerMode() {
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetSize(a.ctx, 350, 600)
	
	// Position top right (simplified, usually we query screen size)
	// For now we position it near the top right
	runtime.WindowSetPosition(a.ctx, 1500, 50)
	runtime.EventsEmit(a.ctx, "mode_changed", "drawer")
}

// EnableFullUI restores the dashboard
func (a *App) EnableFullUI() {
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	runtime.WindowSetSize(a.ctx, 1024, 768)
	runtime.WindowCenter(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "mode_changed", "full")
}

// ToggleStartup toggles the run-on-startup behavior in the OS
func (a *App) ToggleStartup(enable bool) error {
	if enable {
		return EnableStartup()
	}
	return DisableStartup()
}

// GetStartupState returns true if startup is enabled
func (a *App) GetStartupState() bool {
	return IsStartupEnabled()
}

func RunDesktopApp() error {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "FileShare",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: web.GetFS(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: true, // Frameless for glassmorphism
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})

	return err
}
