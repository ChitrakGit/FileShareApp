package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeviceStore(t *testing.T) {
	tempDir := t.TempDir()
	ds := NewDeviceStore(tempDir)

	// 1. Record device
	rec := ds.RecordDevice("dev-laptop-1", "MacBook-Pro", "192.168.1.100", 8990, "darwin", "UDP_BROADCAST", "")
	if rec == nil {
		t.Fatal("expected non-nil record")
	}

	if rec.LanInfo.Subnet != "192.168.1.0/24" {
		t.Fatalf("expected subnet 192.168.1.0/24, got %s", rec.LanInfo.Subnet)
	}

	// 2. Verify file was created in store dir
	expectedFile := filepath.Join(tempDir, "devices.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("devices.json was not created at %s", expectedFile)
	}

	// 3. Record transfer
	ds.RecordTransfer("192.168.1.100")
	devices := ds.GetAllDevices()
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].TotalTransfers != 1 {
		t.Fatalf("expected 1 transfer, got %d", devices[0].TotalTransfers)
	}

	// 4. Remove device
	removed, err := ds.RemoveDevice("dev-laptop-1")
	if err != nil || !removed {
		t.Fatalf("expected device removed, got %v, err=%v", removed, err)
	}

	if len(ds.GetAllDevices()) != 0 {
		t.Fatalf("expected 0 devices after removal")
	}
}
