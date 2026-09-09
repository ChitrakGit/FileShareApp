package discovery

import (
	"testing"
	"time"
)

func TestPeerRegistry(t *testing.T) {
	reg := NewPeerRegistry()

	peer1 := Peer{
		ID:   "device-1",
		Name: "Alice-PC",
		IP:   "192.168.1.100",
		Port: 8990,
		OS:   "windows",
	}

	peer2 := Peer{
		ID:   "device-2",
		Name: "Bob-Mac",
		IP:   "192.168.1.101",
		Port: 8990,
		OS:   "darwin",
	}

	reg.AddOrUpdate(peer1)
	reg.AddOrUpdate(peer2)

	active := reg.GetActivePeers(5 * time.Second)
	if len(active) != 2 {
		t.Fatalf("expected 2 active peers, got %d", len(active))
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	activeAfterTimeout := reg.GetActivePeers(50 * time.Millisecond)
	if len(activeAfterTimeout) != 0 {
		t.Fatalf("expected 0 active peers after timeout, got %d", len(activeAfterTimeout))
	}
}

func TestSavedPeers(t *testing.T) {
	reg := NewPeerRegistry()
	// Use isolated temp file for test
	reg.storagePath = t.TempDir() + "/saved_peers.json"

	p := Peer{
		ID:   "office-pc",
		Name: "Office-Workstation",
		IP:   "192.168.1.55",
		Port: 8990,
		OS:   "windows",
	}

	// 1. Add saved peer
	if err := reg.AddSavedPeer(p); err != nil {
		t.Fatalf("AddSavedPeer failed: %v", err)
	}

	// 2. Check GetAllPeers includes it
	all := reg.GetAllPeers(1 * time.Second)
	found := false
	for _, item := range all {
		if item.Name == "Office-Workstation" && item.IsSaved {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected saved peer to be present in GetAllPeers")
	}

	// 3. Remove saved peer
	removed, err := reg.RemoveSavedPeer("office-pc")
	if err != nil || !removed {
		t.Fatalf("RemoveSavedPeer failed: %v, removed=%v", err, removed)
	}

	allAfterRemove := reg.GetAllPeers(1 * time.Second)
	for _, item := range allAfterRemove {
		if item.ID == "office-pc" {
			t.Fatalf("expected peer to be removed, but still found")
		}
	}
}
