//go:build linux

package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const autostartFile = "fileshare.desktop"

func getAutostartDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "autostart")
}

// EnableStartup adds the application to Linux startup via standard autostart directories
func EnableStartup() error {
	dir := getAutostartDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=FileShare
Comment=Fast LAN File Transfer
Exec="%s" desktop
Terminal=false
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`, filepath.Clean(exePath))

	return os.WriteFile(filepath.Join(dir, autostartFile), []byte(content), 0644)
}

// DisableStartup removes the application from Linux startup
func DisableStartup() error {
	path := filepath.Join(getAutostartDir(), autostartFile)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsStartupEnabled checks if the app is configured to run at startup
func IsStartupEnabled() bool {
	path := filepath.Join(getAutostartDir(), autostartFile)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		content, err := os.ReadFile(path)
		if err == nil && strings.Contains(string(content), "X-GNOME-Autostart-enabled=true") {
			return true
		}
	}
	return false
}
