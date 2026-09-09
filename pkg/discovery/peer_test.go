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
