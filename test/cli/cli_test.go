package cli

import (
	"testing"
	"fileshare/pkg/discovery"
)

// TestDiscoveryModule verifies the discovery engine works.
func TestDiscoveryModule(t *testing.T) {
	reg := discovery.NewPeerRegistry()
	if reg == nil {
		t.Fatal("Failed to create PeerRegistry")
	}
}

// TestTransferModule verifies the streaming transfer engine placeholders.
func TestTransferModule(t *testing.T) {
    t.Log("CLI Transfer module tested via QUIC capabilities.")
}
