package test

import (
	"fileshare/pkg/discovery"
	"os"
	"testing"
)

func TestE2EV5_AddDeviceOnTheFly(t *testing.T) {
	// 1. Setup local registry
	// Set custom home dir to avoid polluting user's ~/.fileshare
	tmpDir, _ := os.MkdirTemp("", "fileshare-v5-test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir) // For Windows

	reg := discovery.NewPeerRegistry()

	// 2. Add device manually
	err := reg.AddSavedPeer(discovery.Peer{
		ID:   "192.168.1.99:8997",
		Name: "My Linux Device",
		IP:   "192.168.1.99",
		Port: 8997,
		OS:   "linux",
	})

	if err != nil {
		t.Fatalf("Failed to add saved peer: %v", err)
	}

	// 3. Verify the registry
	peers := reg.GetAllPeers(1 * 60 * 1000 * 1000 * 1000) // 1 minute
	found := false
	for _, p := range peers {
		if p.Name == "My Linux Device" {
			found = true
			if !p.IsSaved {
				t.Errorf("Expected peer to be saved")
			}
			if p.OS != "linux" {
				t.Errorf("Expected OS to be linux, got %s", p.OS)
			}
		}
	}

	if !found {
		t.Errorf("Newly added device not found in registry")
	}
}
