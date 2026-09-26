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

func TestE2EMany_Transfer(t *testing.T) {
	port1 := 8997
	port2 := 8998
	
	// Setup download directories
	downloadDir1, err := os.MkdirTemp("", "fileshare-many-down1")
	if err != nil {
		t.Fatalf("Failed to create download dir 1: %v", err)
	}
	defer os.RemoveAll(downloadDir1)

	downloadDir2, err := os.MkdirTemp("", "fileshare-many-down2")
	if err != nil {
		t.Fatalf("Failed to create download dir 2: %v", err)
	}
	defer os.RemoveAll(downloadDir2)

	sourceDir, err := os.MkdirTemp("", "fileshare-many-source")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// Start Receiver Server 1
	reg1 := discovery.NewPeerRegistry()
	disc1, _ := discovery.NewService(port1, reg1)
	disc1.Start()
	defer disc1.Stop()

	srv1 := server.NewServer(port1, downloadDir1, disc1, true, "123456")
	go srv1.Start()
	defer srv1.Stop()

	// Start Receiver Server 2
	reg2 := discovery.NewPeerRegistry()
	disc2, _ := discovery.NewService(port2, reg2)
	disc2.Start()
	defer disc2.Stop()

	srv2 := server.NewServer(port2, downloadDir2, disc2, true, "123456")
	go srv2.Start()
	defer srv2.Stop()

	time.Sleep(1 * time.Second) // Wait for servers to boot

	// Create dummy file to send
	testFile := filepath.Join(sourceDir, "many_test_file.txt")
	err = os.WriteFile(testFile, []byte("Hello FileShare Many over QUIC!"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	targetAddrs := []string{
		fmt.Sprintf("127.0.0.1:%d", port1),
		fmt.Sprintf("127.0.0.1:%d", port2),
	}
	
	// SendPaths blocks until done
	err = client.SendPaths(targetAddrs, []string{testFile}, "123456", "TestSender")
	if err != nil {
		t.Fatalf("SendPaths failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Verify file received on server 1
	receivedFile1 := filepath.Join(downloadDir1, "many_test_file.txt")
	content1, err := os.ReadFile(receivedFile1)
	if err != nil {
		t.Fatalf("Failed to read received file 1: %v", err)
	}
	if string(content1) != "Hello FileShare Many over QUIC!" {
		t.Errorf("Content mismatch on server 1, got: %s", string(content1))
	}

	// Verify file received on server 2
	receivedFile2 := filepath.Join(downloadDir2, "many_test_file.txt")
	content2, err := os.ReadFile(receivedFile2)
	if err != nil {
		t.Fatalf("Failed to read received file 2: %v", err)
	}
	if string(content2) != "Hello FileShare Many over QUIC!" {
		t.Errorf("Content mismatch on server 2, got: %s", string(content2))
	}
}

func TestE2EMany_Transfer_SlowReceiverDropped(t *testing.T) {
	port1 := 8981
	port2 := 8982
	
	downloadDir1, _ := os.MkdirTemp("", "fileshare-many-down1")
	defer os.RemoveAll(downloadDir1)

	downloadDir2, _ := os.MkdirTemp("", "fileshare-many-down2")
	defer os.RemoveAll(downloadDir2)

	sourceDir, _ := os.MkdirTemp("", "fileshare-many-source")
	defer os.RemoveAll(sourceDir)

	// Start Receiver Server 1 (Healthy)
	reg1 := discovery.NewPeerRegistry()
	disc1, _ := discovery.NewService(port1, reg1)
	disc1.Start()
	defer disc1.Stop()

	srv1 := server.NewServer(port1, downloadDir1, disc1, true, "123456")
	go srv1.Start()
	defer srv1.Stop()

	// Start Receiver Server 2 (Slow/Failing)
	// We'll simulate a failing receiver by spinning up a server, but quickly shutting it down
	reg2 := discovery.NewPeerRegistry()
	disc2, _ := discovery.NewService(port2, reg2)
	disc2.Start()

	srv2 := server.NewServer(port2, downloadDir2, disc2, true, "123456")
	go srv2.Start()

	time.Sleep(1 * time.Second) // Wait for servers to boot

	// Create dummy file to send
	testFile := filepath.Join(sourceDir, "many_test_file.txt")
	os.WriteFile(testFile, []byte("Hello FileShare Many over QUIC!"), 0644)

	targetAddrs := []string{
		fmt.Sprintf("127.0.0.1:%d", port1),
		fmt.Sprintf("127.0.0.1:%d", port2),
	}
	
	// Shut down server 2 *right before* sending to simulate a connection error or broken stream
	srv2.Stop()
	disc2.Stop()
	
	// SendPaths should not fail the entire transfer if at least one receiver succeeds
	err := client.SendPaths(targetAddrs, []string{testFile}, "123456", "TestSender")
	if err != nil {
		t.Fatalf("SendPaths failed when one receiver was offline: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Verify file received on healthy server 1
	receivedFile1 := filepath.Join(downloadDir1, "many_test_file.txt")
	content1, err := os.ReadFile(receivedFile1)
	if err != nil {
		t.Fatalf("Failed to read received file 1: %v", err)
	}
	if string(content1) != "Hello FileShare Many over QUIC!" {
		t.Errorf("Content mismatch on server 1, got: %s", string(content1))
	}
}
