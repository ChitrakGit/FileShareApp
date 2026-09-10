package desktop

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"
)

// setupSystray initializes the taskbar icon.
func setupSystray(a *App) {
	// systray.Run must be called, but wait, Wails blocks the main thread.
	// On Windows, systray can run in a goroutine, but on Mac it must run on the main thread.
	// Wails v2 has an experimental systray feature, but we are using getlantern/systray.
	// Since getlantern/systray on Windows works in a goroutine, we'll try it.

	systray.Run(func() {
		onReady(a)
	}, onExit)
}

func onReady(a *App) {
	// Provide a blank/default icon for now; in production we'd load bytes from an embed.
	// For testing, just set title.
	systray.SetTitle("FileShare")
	systray.SetTooltip("FileShare - Running")

	mShow := systray.AddMenuItem("Show FileShare", "Open the main dashboard")
	mDrawer := systray.AddMenuItem("Open Quick Drawer", "Float the device drawer")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit the application completely")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				a.EnableFullUI()
			case <-mDrawer.ClickedCh:
				a.EnableDrawerMode()
			case <-mQuit.ClickedCh:
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func onExit() {
	fmt.Println("Exiting systray...")
}
