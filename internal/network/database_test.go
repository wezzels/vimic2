// Package network provides database tests
package network

import (
	"context"
	"os"
	"testing"
)

// TestNetworkDB_RealCreation tests real database creation
func TestNetworkDB_RealCreation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil network DB")
	}
}

// TestNetworkDB_SaveAndGetNetwork tests save/get network operations
func TestNetworkDB_SaveAndGetNetwork(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	// Create and save network
	network := &Network{
		ID:          "net-1",
		Name:        "test-network",
		CIDR:        "192.168.1.0/24",
		Gateway:     "192.168.1.1",
		DHCPEnabled: true,
		VLANID:       100,
	}

	ctx := context.Background()
	err = db.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatalf("failed to save network: %v", err)
	}

	// Retrieve it
	retrieved, err := db.GetNetwork(ctx, "net-1")
	if err != nil {
		t.Fatalf("failed to get network: %v", err)
	}

	if retrieved.Name != "test-network" {
		t.Errorf("expected test-network, got %s", retrieved.Name)
	}
}

// TestNetworkDB_ListNetworks tests listing networks
func TestNetworkDB_ListNetworks(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Save multiple networks
	for i := 1; i <= 3; i++ {
		network := &Network{
			ID:   "net-" + string(rune('0'+i)),
			Name: "network-" + string(rune('0'+i)),
			CIDR: "192.168." + string(rune('0'+i)) + ".0/24",
		}
		db.SaveNetwork(ctx, network)
	}

	networks, err := db.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("failed to list networks: %v", err)
	}

	if len(networks) < 3 {
		t.Errorf("expected at least 3 networks, got %d", len(networks))
	}
}

// TestNetworkDB_DeleteNetwork tests deleting networks
func TestNetworkDB_DeleteNetwork(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create and delete
	network := &Network{ID: "net-to-delete", Name: "delete-me"}
	db.SaveNetwork(ctx, network)

	err = db.DeleteNetwork(ctx, "net-to-delete")
	if err != nil {
		t.Fatalf("failed to delete network: %v", err)
	}

	// Should not exist
	_, err = db.GetNetwork(ctx, "net-to-delete")
	if err == nil {
		t.Error("expected error for deleted network")
	}
}

// TestNetworkDB_SaveAndGetRouter tests router operations
func TestNetworkDB_SaveAndGetRouter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	router := &Router{
		ID:   "router-1",
		Name: "main-router",
	}

	err = db.SaveRouter(ctx, router)
	if err != nil {
		t.Fatalf("failed to save router: %v", err)
	}

	retrieved, err := db.GetRouter(ctx, "router-1")
	if err != nil {
		t.Fatalf("failed to get router: %v", err)
	}

	if retrieved.Name != "main-router" {
		t.Errorf("expected main-router, got %s", retrieved.Name)
	}
}

// TestNetworkDB_SaveAndGetFirewall tests firewall operations
func TestNetworkDB_SaveAndGetFirewall(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	firewall := &Firewall{
		ID:        "fw-1",
		Name:      "web-firewall",
		NetworkID: "net-1",
	}

	err = db.SaveFirewall(ctx, firewall)
	if err != nil {
		t.Fatalf("failed to save firewall: %v", err)
	}

	retrieved, err := db.GetFirewall(ctx, "fw-1")
	if err != nil {
		t.Fatalf("failed to get firewall: %v", err)
	}

	if retrieved.Name != "web-firewall" {
		t.Errorf("expected web-firewall, got %s", retrieved.Name)
	}
}

// TestNetworkDB_Stats tests database stats
func TestNetworkDB_Stats(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Get initial stats
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats == nil {
		t.Error("expected non-nil stats")
	}
}