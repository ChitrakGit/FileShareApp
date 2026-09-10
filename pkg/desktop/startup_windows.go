package desktop

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const appRegistryName = "FileShareLAN"

// EnableStartup adds the application to Windows startup
func EnableStartup() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Add "desktop" argument so it starts in desktop mode
	startCmd := fmt.Sprintf(`"%s" desktop`, filepath.Clean(exePath))
	return k.SetStringValue(appRegistryName, startCmd)
}

// DisableStartup removes the application from Windows startup
func DisableStartup() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.DeleteValue(appRegistryName)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsStartupEnabled checks if the app is configured to run at startup
func IsStartupEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(appRegistryName)
	return err == nil && val != ""
}
