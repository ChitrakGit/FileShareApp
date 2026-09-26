package test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fileshare/pkg/client"
	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

func TestEndToEnd(t *testing.T) {
	// Setup temporary directory for server downloads
	downloadDir := t.TempDir()

	// Create a Peer Registry and Discovery Service
	reg := discovery.NewPeerRegistry()
	disc, err := discovery.NewService(9099, reg)
	if err != nil {
		t.Fatalf("Failed to create discovery service: %v", err)
	}

	// Start server on port 9099
	srv := server.NewServer(9099, downloadDir, disc, true, "")

	go func() {
		_ = srv.Start()
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Test 1: Info endpoint to verify MAC Address feature (V2)
	t.Run("API Info Endpoint", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:9099/api/info")
		if err != nil {
			t.Fatalf("Failed to hit /api/info: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var info map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			t.Fatalf("Failed to decode info JSON: %v", err)
		}

		if mac, ok := info["macAddress"].(string); !ok || mac == "" {
			t.Errorf("Expected macAddress to be present and non-empty, got: %v", info["macAddress"])
		} else {
			t.Logf("Successfully retrieved MAC Address: %s", mac)
		}
	})

	// Test 2: File Transfer streaming over HTTP (V2)
	t.Run("File Transfer Streaming", func(t *testing.T) {
		testFile := filepath.Join(t.TempDir(), "hello.txt")
		err = os.WriteFile(testFile, []byte("Hello, FileShare!"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = client.SendPaths([]string{"127.0.0.1:9099"}, []string{testFile}, "", "IntegrationTestClient")
		if err != nil {
			t.Fatalf("Failed to send file: %v", err)
		}

		// Verify file was received correctly in the download directory
		receivedFile := filepath.Join(downloadDir, "hello.txt")
		content, err := os.ReadFile(receivedFile)
		if err != nil {
			t.Fatalf("Failed to read received file: %v", err)
		}

		if string(content) != "Hello, FileShare!" {
			t.Errorf("Expected 'Hello, FileShare!', got '%s'", string(content))
		} else {
			t.Logf("Successfully sent and verified received file content!")
		}
	})
}
