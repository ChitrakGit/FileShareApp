//go:build darwin

package desktop

import (
	"fmt"
	"os"
	"path/filepath"
)

const plistFile = "com.fileshare.app.plist"

func getLaunchAgentsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents")
}

// EnableStartup adds the application to macOS startup via LaunchAgents
func EnableStartup() error {
	dir := getLaunchAgentsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.fileshare.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>desktop</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, filepath.Clean(exePath))

	return os.WriteFile(filepath.Join(dir, plistFile), []byte(content), 0644)
}

// DisableStartup removes the application from macOS startup
func DisableStartup() error {
	path := filepath.Join(getLaunchAgentsDir(), plistFile)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsStartupEnabled checks if the app is configured to run at startup
func IsStartupEnabled() bool {
	path := filepath.Join(getLaunchAgentsDir(), plistFile)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
