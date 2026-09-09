package discovery

import (
	"sync"
	"time"
)

// Peer represents a device discovered on the local network.
type Peer struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	OS       string    `json:"os"`
	LastSeen time.Time `json:"lastSeen"`
}

// PeerRegistry holds a thread-safe set of known network peers.
type PeerRegistry struct {
	mu    sync.RWMutex
	peers map[string]Peer
}

// NewPeerRegistry initializes an empty registry.
func NewPeerRegistry() *PeerRegistry {
	return &PeerRegistry{
		peers: make(map[string]Peer),
	}
}

// AddOrUpdate registers or updates a peer's last seen time.
func (r *PeerRegistry) AddOrUpdate(p Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.LastSeen = time.Now()
	r.peers[p.ID] = p
}

// GetActivePeers returns all peers seen within the timeout period.
func (r *PeerRegistry) GetActivePeers(timeout time.Duration) []Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var active []Peer
	for id, peer := range r.peers {
		if now.Sub(peer.LastSeen) > timeout {
			delete(r.peers, id)
			continue
		}
		active = append(active, peer)
	}
	return active
}
