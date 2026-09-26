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

func TestE2EV6_DirectoryTransfer(t *testing.T) {
	port := 8996
	// 1. Setup directories
	downloadDir, err := os.MkdirTemp("", "fileshare-v6-download")
	if err != nil {
		t.Fatalf("Failed to create download dir: %v", err)
	}
	defer os.RemoveAll(downloadDir)

	sourceDir, err := os.MkdirTemp("", "fileshare-v6-source")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// Create a complex directory structure
	testDir := filepath.Join(sourceDir, "test_dir")
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test_dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("File 1 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1.txt: %v", err)
	}

	subDir := filepath.Join(testDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("File 2 content inside subdir"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2.txt: %v", err)
	}

	// 2. Start Receiver Server
	reg := discovery.NewPeerRegistry()
	disc, _ := discovery.NewService(port, reg)
	disc.Start()
	defer disc.Stop()

	srv := server.NewServer(port, downloadDir, disc, true, "123456")
	go srv.Start()
	defer srv.Stop()

	time.Sleep(1 * time.Second) // Wait for server to boot

	// 4. Send via QUIC Client
	targetAddr := fmt.Sprintf("127.0.0.1:%d", port)
	
	// SendPaths blocks until done
	startTime := time.Now()
	err = client.SendPaths([]string{targetAddr}, []string{testDir}, "123456", "TestSender")
	if err != nil {
		t.Fatalf("SendPaths failed: %v", err)
	}
	duration := time.Since(startTime)
	t.Logf("Directory transferred in %v", duration)

	// Wait for transfer to finalize on disk
	time.Sleep(500 * time.Millisecond)

	// Verify file received
	receivedFile1 := filepath.Join(downloadDir, "test_dir", "file1.txt")
	content1, err := os.ReadFile(receivedFile1)
	if err != nil {
		t.Fatalf("Failed to read received file1: %v", err)
	}

	if string(content1) != "File 1 content" {
		t.Errorf("Content mismatch, got: %s", string(content1))
	}

	receivedFile2 := filepath.Join(downloadDir, "test_dir", "subdir", "file2.txt")
	content2, err := os.ReadFile(receivedFile2)
	if err != nil {
		t.Fatalf("Failed to read received file2: %v", err)
	}

	if string(content2) != "File 2 content inside subdir" {
		t.Errorf("Content mismatch, got: %s", string(content2))
	}
}
