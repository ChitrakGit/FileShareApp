package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LanInfo holds local network and connection configuration for a device.
type LanInfo struct {
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"` // e.g. "IPv4"
	Subnet      string `json:"subnet"`   // e.g. "192.168.1.0/24"
	EndpointURL string `json:"endpointUrl"`
}

// SystemInfo holds platform and OS metadata.
type SystemInfo struct {
	OS       string `json:"os"`
	Hostname string `json:"hostname,omitempty"`
}

// ConnectionInfo tracks discovery source and online timestamps.
type ConnectionInfo struct {
	DiscoveryMethod string    `json:"discoveryMethod"` // "UDP_BROADCAST", "MANUAL_ADDED", "SUBNET_SCAN", "INCOMING_TRANSFER"
	Status          string    `json:"status"`          // "ONLINE", "OFFLINE"
	FirstSeen       time.Time `json:"firstSeen"`
	LastSeen        time.Time `json:"lastSeen"`
}

// DeviceRecord represents complete stored metadata for a network device.
type DeviceRecord struct {
	DeviceID       string         `json:"deviceId"`
	DeviceName     string         `json:"deviceName"`
	LanInfo        LanInfo        `json:"lanInfo"`
	SystemInfo     SystemInfo     `json:"systemInfo"`
	Connection     ConnectionInfo `json:"connection"`
	TotalTransfers int            `json:"totalTransfers"`
	LastTransferAt *time.Time     `json:"lastTransferAt,omitempty"`
}

// DeviceStore handles thread-safe persistence of connected devices into store/devices.json.
type DeviceStore struct {
	mu       sync.RWMutex
	filePath string
	devices  map[string]*DeviceRecord
}

var (
	defaultStore *DeviceStore
	once         sync.Once
)

// Default returns the global DeviceStore singleton pointing to ./store/devices.json.
func Default() *DeviceStore {
	once.Do(func() {
		defaultStore = NewDeviceStore("store")
	})
	return defaultStore
}

// NewDeviceStore creates a store in the specified directory (e.g. "store").
func NewDeviceStore(dir string) *DeviceStore {
	_ = os.MkdirAll(dir, 0755)
	filePath := filepath.Join(dir, "devices.json")

	ds := &DeviceStore{
		filePath: filePath,
		devices:  make(map[string]*DeviceRecord),
	}

	ds.load()
	return ds
}

func (s *DeviceStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}

	var list []DeviceRecord
	if err := json.Unmarshal(data, &list); err == nil {
		for i := range list {
			item := list[i]
			s.devices[item.DeviceID] = &item
		}
	}
}

func (s *DeviceStore) persistLocked() error {
	var list []DeviceRecord
	for _, d := range s.devices {
		list = append(list, *d)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// RecordDevice inserts or updates a device record with LAN connection information.
func (s *DeviceStore) RecordDevice(id, name, ip string, port int, osName, method string) *DeviceRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == "" {
		id = fmt.Sprintf("%s:%d", ip, port)
	}
	if name == "" {
		name = fmt.Sprintf("Device-%s", ip)
	}

	now := time.Now()
	subnet := calculateSubnet(ip)
	endpointURL := fmt.Sprintf("http://%s:%d", ip, port)

	existing, exists := s.devices[id]
	if !exists {
		// Also check by IP and port in case ID differed slightly
		for _, dev := range s.devices {
			if dev.LanInfo.IP == ip && dev.LanInfo.Port == port {
				existing = dev
				exists = true
				break
			}
		}
	}

	if exists {
		existing.DeviceName = name
		existing.LanInfo.IP = ip
		existing.LanInfo.Port = port
		existing.LanInfo.Subnet = subnet
		existing.LanInfo.EndpointURL = endpointURL
		if osName != "" {
			existing.SystemInfo.OS = osName
		}
		existing.Connection.Status = "ONLINE"
		existing.Connection.LastSeen = now
		if method != "" {
			existing.Connection.DiscoveryMethod = method
		}
		_ = s.persistLocked()
		return existing
	}

	record := &DeviceRecord{
		DeviceID:   id,
		DeviceName: name,
		LanInfo: LanInfo{
			IP:          ip,
			Port:        port,
			Protocol:    "IPv4",
			Subnet:      subnet,
			EndpointURL: endpointURL,
		},
		SystemInfo: SystemInfo{
			OS:       osName,
			Hostname: name,
		},
		Connection: ConnectionInfo{
			DiscoveryMethod: method,
			Status:          "ONLINE",
			FirstSeen:       now,
			LastSeen:        now,
		},
		TotalTransfers: 0,
	}

	s.devices[id] = record
	_ = s.persistLocked()
	return record
}

// RecordTransfer increments transfer counter for a device and updates last active timestamp.
func (s *DeviceStore) RecordTransfer(idOrIP string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, dev := range s.devices {
		if dev.DeviceID == idOrIP || dev.LanInfo.IP == idOrIP || fmt.Sprintf("%s:%d", dev.LanInfo.IP, dev.LanInfo.Port) == idOrIP || dev.DeviceName == idOrIP {
			dev.TotalTransfers++
			dev.LastTransferAt = &now
			dev.Connection.LastSeen = now
			dev.Connection.Status = "ONLINE"
			_ = s.persistLocked()
			return
		}
	}
}

// GetAllDevices returns all recorded devices from the store.
func (s *DeviceStore) GetAllDevices() []DeviceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []DeviceRecord
	for _, d := range s.devices {
		list = append(list, *d)
	}
	return list
}

// RemoveDevice removes a device from the store.
func (s *DeviceStore) RemoveDevice(idOrIP string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetKey := ""
	for k, dev := range s.devices {
		if k == idOrIP || dev.DeviceID == idOrIP || dev.LanInfo.IP == idOrIP {
			targetKey = k
			break
		}
	}

	if targetKey == "" {
		return false, nil
	}

	delete(s.devices, targetKey)
	err := s.persistLocked()
	return true, err
}

func calculateSubnet(ip string) string {
	lastDot := strings.LastIndex(ip, ".")
	if lastDot != -1 {
		return ip[:lastDot] + ".0/24"
	}
	return "255.255.255.0"
}
