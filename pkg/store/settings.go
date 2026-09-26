package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// AppSettings represents user configurable settings.
type AppSettings struct {
	SaveDirectory   string   `json:"saveDirectory"`
	DeviceName      string   `json:"deviceName"`
	MulticastAddr   string   `json:"multicastAddr"`
	MulticastPort   int      `json:"multicastPort"`
	BlockedPortsIn  []int    `json:"blockedPortsIn"`
	BlockedPortsOut []int    `json:"blockedPortsOut"`
	BlockedDevices  []string `json:"blockedDevices"`
}

// SettingsStore handles persistence of user settings.
type SettingsStore struct {
	mu       sync.RWMutex
	filePath string
	settings AppSettings
}

var (
	defaultSettingsStore *SettingsStore
	settingsOnce         sync.Once
)

// DefaultSettings returns the global SettingsStore singleton pointing to ./store/settings.json.
func DefaultSettings() *SettingsStore {
	settingsOnce.Do(func() {
		defaultSettingsStore = NewSettingsStore("store")
	})
	return defaultSettingsStore
}

// NewSettingsStore initializes settings persistence.
func NewSettingsStore(dir string) *SettingsStore {
	_ = os.MkdirAll(dir, 0755)
	filePath := filepath.Join(dir, "settings.json")

	ss := &SettingsStore{
		filePath: filePath,
		settings: AppSettings{
			// Default values
			MulticastAddr: "224.0.0.1",
			MulticastPort: 53535,
			BlockedPortsIn: []int{},
			BlockedPortsOut: []int{},
			BlockedDevices: []string{},
		},
	}
	ss.load()
	return ss
}

func (s *SettingsStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &s.settings)
}

func (s *SettingsStore) persistLocked() error {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// Get returns a copy of current settings.
func (s *SettingsStore) Get() AppSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// Update saves new settings.
func (s *SettingsStore) Update(newSettings AppSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = newSettings
	return s.persistLocked()
}
