package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fileshare/pkg/store"
)

// Peer represents a device discovered on the local network or saved manually.
type Peer struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	Port     int       `json:"port"`
	OS       string    `json:"os"`
	LastSeen time.Time `json:"lastSeen"`
	IsSaved  bool      `json:"isSaved"`
	Online   bool      `json:"online"`
}

// PeerRegistry holds thread-safe sets of dynamically discovered and persistently saved peers.
type PeerRegistry struct {
	mu          sync.RWMutex
	activePeers map[string]Peer
	savedPeers  map[string]Peer
	storagePath string
}

// NewPeerRegistry initializes registry and loads saved peers from disk.
func NewPeerRegistry() *PeerRegistry {
	storagePath := getStorageFilePath()

	reg := &PeerRegistry{
		activePeers: make(map[string]Peer),
		savedPeers:  make(map[string]Peer),
		storagePath: storagePath,
	}

	reg.loadSavedPeers()
	return reg
}

func getStorageFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".fileshare")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "saved_peers.json")
}

func (r *PeerRegistry) loadSavedPeers() {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		return
	}

	var list []Peer
	if err := json.Unmarshal(data, &list); err == nil {
		for _, p := range list {
			p.IsSaved = true
			r.savedPeers[p.ID] = p
		}
	}
}

func (r *PeerRegistry) persistSavedPeersLocked() error {
	var list []Peer
	for _, p := range r.savedPeers {
		list = append(list, p)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.storagePath, data, 0644)
}

// AddOrUpdate registers or updates an active peer seen via broadcast.
func (r *PeerRegistry) AddOrUpdate(p Peer) {
	r.mu.Lock()
	p.LastSeen = time.Now()
	p.Online = true
	r.activePeers[p.ID] = p
	r.mu.Unlock()

	// Persist to store/devices.json
	store.Default().RecordDevice(p.ID, p.Name, p.IP, p.Port, p.OS, "UDP_BROADCAST", "")
}

// AddSavedPeer saves a peer permanently to disk.
func (r *PeerRegistry) AddSavedPeer(p Peer) error {
	r.mu.Lock()

	if p.ID == "" {
		p.ID = fmt.Sprintf("%s:%d", p.IP, p.Port)
	}
	p.IsSaved = true
	p.LastSeen = time.Now()
	p.Online = true

	r.savedPeers[p.ID] = p
	err := r.persistSavedPeersLocked()
	r.mu.Unlock()

	// Persist to store/devices.json
	store.Default().RecordDevice(p.ID, p.Name, p.IP, p.Port, p.OS, "MANUAL_ADDED", "")
	return err
}

// RemoveSavedPeer deletes a peer from permanent storage.
func (r *PeerRegistry) RemoveSavedPeer(idOrIP string) (bool, error) {
	r.mu.Lock()

	targetKey := ""
	for k, p := range r.savedPeers {
		if k == idOrIP || p.ID == idOrIP || p.IP == idOrIP || fmt.Sprintf("%s:%d", p.IP, p.Port) == idOrIP || p.Name == idOrIP {
			targetKey = k
			break
		}
	}

	if targetKey == "" {
		r.mu.Unlock()
		return false, nil
	}

	delete(r.savedPeers, targetKey)
	err := r.persistSavedPeersLocked()
	r.mu.Unlock()

	// Remove from store/devices.json
	_, _ = store.Default().RemoveDevice(targetKey)
	return true, err
}

// GetActivePeers returns all peers seen via broadcast within the timeout period.
func (r *PeerRegistry) GetActivePeers(timeout time.Duration) []Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var active []Peer
	for id, peer := range r.activePeers {
		if now.Sub(peer.LastSeen) > timeout {
			delete(r.activePeers, id)
			continue
		}
		active = append(active, peer)
	}
	return active
}

// GetAllPeers combines active broadcast peers and saved peers into a unified list.
func (r *PeerRegistry) GetAllPeers(timeout time.Duration) []Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	merged := make(map[string]Peer)

	// Add saved peers first
	for id, p := range r.savedPeers {
		p.IsSaved = true
		merged[id] = p
	}

	// Add or overlay active broadcast peers
	for id, p := range r.activePeers {
		if now.Sub(p.LastSeen) <= timeout {
			if existing, ok := merged[id]; ok {
				existing.Online = true
				existing.LastSeen = p.LastSeen
				merged[id] = existing
			} else {
				p.Online = true
				merged[id] = p
			}
		}
	}

	var result []Peer
	for _, p := range merged {
		result = append(result, p)
	}
	return result
}

// ProbePeer checks if a FileShare instance is reachable at the given address.
func ProbePeer(addr string, defaultPort int, timeout time.Duration) (*Peer, error) {
	host := addr
	port := defaultPort

	if strings.Contains(addr, ":") {
		parts := strings.Split(addr, ":")
		host = parts[0]
		_, _ = fmt.Sscanf(parts[1], "%d", &port)
	}

	url := fmt.Sprintf("http://%s:%d/api/info", host, port)
	client := &http.Client{Timeout: timeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("device not reachable at %s:%d: %w", host, port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	var info struct {
		DeviceName string `json:"deviceName"`
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		OS         string `json:"os"`
		MACAddress string `json:"macAddress"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid response from %s: %w", url, err)
	}

	p := &Peer{
		ID:       fmt.Sprintf("%s:%d", host, port),
		Name:     info.DeviceName,
		IP:       host,
		Port:     port,
		OS:       info.OS,
		LastSeen: time.Now(),
		Online:   true,
	}

	// Persist to store/devices.json
	store.Default().RecordDevice(p.ID, p.Name, p.IP, p.Port, p.OS, "ACTIVE_PROBE", info.MACAddress)

	return p, nil
}

// ScanSubnet actively probes an IPv4 /24 subnet (e.g., "192.168.1") for FileShare instances.
func ScanSubnet(subnetPrefix string, port int, timeout time.Duration) []Peer {
	subnetPrefix = strings.TrimSuffix(subnetPrefix, ".")
	var discovered []Peer
	var mu sync.Mutex

	// Worker pool with 64 concurrent workers
	jobs := make(chan int, 254)
	var wg sync.WaitGroup

	workers := 64
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for hostNum := range jobs {
				ip := fmt.Sprintf("%s.%d", subnetPrefix, hostNum)
				p, err := ProbePeer(ip, port, timeout)
				if err == nil && p != nil {
					mu.Lock()
					discovered = append(discovered, *p)
					mu.Unlock()
				}
			}
		}()
	}

	for i := 1; i <= 254; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	return discovered
}

// ProbeTCPPort quickly tests if a TCP port is open (useful for rapid filtering).
func ProbeTCPPort(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
