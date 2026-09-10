package test

import (
	"path/filepath"
	"runtime"
	"testing"
	"os"

	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

func TestEndToEndV3(t *testing.T) {
	// Test 1: Cross-Platform Default Download Directory Resolution
	t.Run("CrossPlatform_DefaultDownloadDir", func(t *testing.T) {
		reg := discovery.NewPeerRegistry()
		disc, err := discovery.NewService(9095, reg)
		if err != nil {
			t.Fatalf("Failed to create discovery service: %v", err)
		}

		// Pass empty download directory to trigger default resolution
		srv := server.NewServer(9095, "", disc, true, "")

		expectedDir := ""
		if runtime.GOOS == "windows" {
			expectedDir = filepath.Clean("C:/Users/Public/Documents/FileShare")
		} else {
			home, _ := os.UserHomeDir()
			expectedDir = filepath.Join(home, "Downloads", "FileShare")
		}

		if srv.DownloadDir != expectedDir {
			t.Errorf("Expected default download dir to be '%s', got '%s'", expectedDir, srv.DownloadDir)
		} else {
			t.Logf("Successfully resolved OS-specific path: %s", srv.DownloadDir)
		}
	})
}
