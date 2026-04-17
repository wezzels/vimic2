//go:build integration

package network

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ==================== NetworkDB ListNetworks ====================

func TestNetworkDB_ListNetworks_Full(t *testing.T) {
	db := setupNetDBFull(t)
	defer db.Close()

	ctx := context.Background()
	network := &Network{
		ID:         "net-list-1",
		Name:       "test-network-1",
		Type:       NetworkTypeBridge,
		BridgeName: "br-test-1",
		CIDR:       "10.0.1.0/24",
		Gateway:    "10.0.1.1",
		DNS:        []string{"8.8.8.8"},
	}
	err := db.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatal(err)
	}

	network2 := &Network{
		ID:         "net-list-2",
		Name:       "test-network-2",
		Type:       NetworkTypeVLAN,
		BridgeName: "br-test-2",
		CIDR:       "10.0.2.0/24",
		Gateway:    "10.0.2.1",
		VLANID:     200,
	}
	err = db.SaveNetwork(ctx, network2)
	if err != nil {
		t.Fatal(err)
	}

	networks, err := db.ListNetworks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networks) < 2 {
		t.Errorf("Expected at least 2 networks, got %d", len(networks))
	}
	t.Logf("Listed %d networks", len(networks))
}

// ==================== NetworkDB GetStats ====================

func TestNetworkDB_GetStats_Full(t *testing.T) {
	db := setupNetDBFull(t)
	defer db.Close()

	stats, err := db.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	t.Logf("Stats: %v", stats)
}

// ==================== IPAM AllocateIP / ReleaseIP / GetMAC ====================

func TestIPAM_AllocateIP_ReleaseIP_GetMAC(t *testing.T) {
	ipam, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Skipf("NewIPAMManager failed: %v", err)
	}

	cidr, gateway, err := ipam.Allocate()
	if err != nil {
		t.Skipf("Allocate failed: %v", err)
	}
	t.Logf("Allocated subnet: %s, gateway: %s", cidr, gateway)

	ip, err := ipam.AllocateIP(cidr, "aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("AllocateIP failed: %v", err)
	}
	t.Logf("Allocated IP: %s", ip)

	foundIP, err := ipam.GetIP("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("GetIP failed: %v", err)
	}
	if foundIP != ip {
		t.Errorf("GetIP returned %s, expected %s", foundIP, ip)
	}

	foundMAC, err := ipam.GetMAC(cidr, ip)
	if err != nil {
		t.Fatalf("GetMAC failed: %v", err)
	}
	if foundMAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("GetMAC returned %s, expected aa:bb:cc:dd:ee:ff", foundMAC)
	}

	err = ipam.ReleaseIP(cidr, ip)
	if err != nil {
		t.Fatalf("ReleaseIP failed: %v", err)
	}

	err = ipam.Release(cidr)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}
}

// ==================== VLAN SetStateFile ====================

func TestVLANAllocator_SetStateFile_Full(t *testing.T) {
	va, err := NewVLANAllocator(300, 400)
	if err != nil {
		t.Fatal(err)
	}

	stateFile := filepath.Join(t.TempDir(), "vlan-state.json")
	va.SetStateFile(stateFile)

	_, err = va.Allocate()
	if err != nil {
		t.Fatal(err)
	}

	err = va.saveState()
	if err != nil {
		t.Fatalf("saveState failed: %v", err)
	}

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file should exist after saveState")
	}
}

// ==================== Helper ====================

func setupNetDBFull(t *testing.T) *NetworkDB {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-netdb-full-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewNetworkDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	return db
}