package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fileshare/pkg/client"
	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

func TestE2EV4_QUIC_Transfer(t *testing.T) {
	port := 8995
	// 1. Setup directories
	downloadDir, err := os.MkdirTemp("", "fileshare-v4-download")
	if err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}
	defer os.RemoveAll(downloadDir)

	sourceDir, err := os.MkdirTemp("", "fileshare-v4-source")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// 2. Start Receiver Server
	reg := discovery.NewPeerRegistry()
	disc, _ := discovery.NewService(port, reg)
	disc.Start()
	defer disc.Stop()

	srv := server.NewServer(port, downloadDir, disc, true, "123456")
	go srv.Start()
	defer srv.Stop()

	time.Sleep(1 * time.Second) // Wait for server to boot

	// 3. Create dummy file to send
	testFile := filepath.Join(sourceDir, "v4_test_file.txt")
	err = os.WriteFile(testFile, []byte("Hello FileShare V4 over QUIC!"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// 4. Send via QUIC Client
	targetAddr := fmt.Sprintf("127.0.0.1:%d", port)
	
	// SendPaths blocks until done
	err = client.SendPaths([]string{targetAddr}, []string{testFile}, "123456", "TestSender")
	if err != nil {
		t.Fatalf("SendPaths failed: %v", err)
	}

	// Wait for transfer to finalize on disk
	time.Sleep(500 * time.Millisecond)

	// Verify file received
	receivedFile := filepath.Join(downloadDir, "v4_test_file.txt")
	// The file gets saved inside the root of download dir directly if it was a single file payload,
	// wait, tar stream might put it in a subdirectory or preserve base name.
	content, err := os.ReadFile(receivedFile)
	if err != nil {
		t.Fatalf("Failed to read received file: %v", err)
	}

	if string(content) != "Hello FileShare V4 over QUIC!" {
		t.Errorf("Content mismatch, got: %s", string(content))
	}
}
