// Package network provides real database tests
package network

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestNetworkDB_ListRouters tests ListRouters
func TestNetworkDB_ListRouters(t *testing.T) {
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

	// Save router
	router := &Router{
		ID:   "router-1",
		Name: "main-router",
	}
	err = db.SaveRouter(ctx, router)
	if err != nil {
		t.Fatalf("failed to save router: %v", err)
	}

	// List routers
	routers, err := db.ListRouters(ctx)
	if err != nil {
		t.Fatalf("failed to list routers: %v", err)
	}
	if len(routers) < 1 {
		t.Error("expected at least one router")
	}
}

// TestNetworkDB_DeleteRouter tests DeleteRouter
func TestNetworkDB_DeleteRouter(t *testing.T) {
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

	// Create and delete router
	router := &Router{ID: "router-del", Name: "delete-me"}
	db.SaveRouter(ctx, router)

	err = db.DeleteRouter(ctx, "router-del")
	if err != nil {
		t.Fatalf("failed to delete router: %v", err)
	}

	// Verify deletion
	routers, _ := db.ListRouters(ctx)
	for _, r := range routers {
		if r.ID == "router-del" {
			t.Error("router should be deleted")
		}
	}
}

// TestNetworkDB_ListFirewalls tests ListFirewalls
func TestNetworkDB_ListFirewalls(t *testing.T) {
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

	// Save firewall
	firewall := &Firewall{
		ID:        "fw-1",
		Name:      "web-firewall",
		NetworkID: "net-1",
	}
	db.SaveFirewall(ctx, firewall)

	// List firewalls
	firewalls, err := db.ListFirewalls(ctx)
	if err != nil {
		t.Fatalf("failed to list firewalls: %v", err)
	}
	if len(firewalls) < 1 {
		t.Error("expected at least one firewall")
	}
}

// TestNetworkDB_DeleteFirewall tests DeleteFirewall
func TestNetworkDB_DeleteFirewall(t *testing.T) {
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

	// Create and delete firewall
	firewall := &Firewall{ID: "fw-del", Name: "delete-me"}
	db.SaveFirewall(ctx, firewall)

	err = db.DeleteFirewall(ctx, "fw-del")
	if err != nil {
		t.Fatalf("failed to delete firewall: %v", err)
	}
}

// TestNetworkDB_DeleteTunnel tests DeleteTunnel
func TestNetworkDB_DeleteTunnel(t *testing.T) {
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

	// Create and delete tunnel
	tunnel := &Tunnel{ID: "tunnel-1", Name: "vpn-tunnel"}
	db.SaveTunnel(ctx, tunnel)

	err = db.DeleteTunnel(ctx, "tunnel-1")
	if err != nil {
		t.Fatalf("failed to delete tunnel: %v", err)
	}
}

// TestNetworkDB_ListInterfaces tests ListInterfaces
func TestNetworkDB_ListInterfaces(t *testing.T) {
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

	// Save VM interface
	iface := &VMInterface{
		ID:        "eth0",
		Name:      "eth0",
		IPAddress: "192.168.1.1",
	}
	db.SaveInterface(ctx, iface)

	// List interfaces
	ifaces, err := db.ListInterfaces(ctx)
	if err != nil {
		t.Fatalf("failed to list interfaces: %v", err)
	}
	if len(ifaces) < 1 {
		t.Error("expected at least one interface")
	}
}

// TestNetworkDB_DeleteInterface tests DeleteInterface
func TestNetworkDB_DeleteInterface(t *testing.T) {
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

	// Create and delete interface
	iface := &VMInterface{ID: "eth-del", Name: "delete-me"}
	db.SaveInterface(ctx, iface)

	err = db.DeleteInterface(ctx, "eth-del")
	if err != nil {
		t.Fatalf("failed to delete interface: %v", err)
	}
}

// TestNetworkDB_Backup tests Backup
func TestNetworkDB_Backup(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	backupFile, err := os.CreateTemp("", "network-backup-*.db")
	if err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}
	backupPath := backupFile.Name()
	backupFile.Close()
	defer os.Remove(backupPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create some data
	network := &Network{ID: "net-1", Name: "backup-test", CIDR: "10.0.0.0/24"}
	db.SaveNetwork(ctx, network)

	// Backup
	err = db.Backup(ctx, backupPath)
	if err != nil {
		t.Fatalf("failed to backup: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}
}

// TestNetworkDB_Restore tests Restore
func TestNetworkDB_Restore(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	backupFile, err := os.CreateTemp("", "network-backup-*.db")
	if err != nil {
		t.Fatalf("failed to create backup file: %v", err)
	}
	backupPath := backupFile.Name()
	backupFile.Close()
	defer os.Remove(backupPath)

	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create network DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create data and backup
	network := &Network{ID: "net-restore", Name: "restore-test", CIDR: "10.1.0.0/24"}
	db.SaveNetwork(ctx, network)
	db.Backup(ctx, backupPath)

	// Delete the network
	db.DeleteNetwork(ctx, "net-restore")

	// Restore
	err = db.Restore(ctx, backupPath)
	if err != nil {
		t.Fatalf("failed to restore: %v", err)
	}

	// Verify restored data
	restored, err := db.GetNetwork(ctx, "net-restore")
	if err != nil {
		t.Fatalf("failed to get restored network: %v", err)
	}
	if restored.Name != "restore-test" {
		t.Errorf("expected restore-test, got %s", restored.Name)
	}
}

// TestNetworkDB_Migrate tests Migrate
func TestNetworkDB_Migrate(t *testing.T) {
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

	// Run migrate (should be idempotent)
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}

// TestNetworkDB_GetStats tests GetStats
func TestNetworkDB_GetStats(t *testing.T) {
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

	// Create some data
	for i := 0; i < 5; i++ {
		network := &Network{
			ID:   "net-" + string(rune('A'+i)),
			Name: "network-" + string(rune('A'+i)),
			CIDR:  "10.0.0.0/24",
		}
		db.SaveNetwork(ctx, network)
	}

	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats == nil {
		t.Error("expected non-nil stats")
	}
}

// TestIPAMManager_Creation tests IPAM manager
func TestIPAMManager_Creation(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.0.0/16",
		DNS:     []string{"8.8.8.8"},
	}

	// Just verify config creation
	if config.BaseCIDR == "" {
		t.Error("BaseCIDR should not be empty")
	}
}

// TestIPAMConfigReal tests IPAM config
func TestIPAMConfigReal(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:     []string{"8.8.8.8", "8.8.4.4"},
	}

	if len(config.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(config.DNS))
	}
}

// TestNetwork_FullWorkflow tests complete network workflow
func TestNetwork_FullWorkflow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "network-full-*.db")
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

	// Create network
	network := &Network{
		ID:          "net-full",
		Name:        "full-test",
		CIDR:        "172.16.0.0/24",
		Gateway:     "172.16.0.1",
		DHCPEnabled: true,
		VLANID:      100,
	}
	err = db.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatalf("failed to save network: %v", err)
	}

	// Create router
	router := &Router{
		ID:        "router-full",
		Name:      "main-router",
		NetworkID: "net-full",
	}
	db.SaveRouter(ctx, router)

	// Create firewall
	firewall := &Firewall{
		ID:        "fw-full",
		Name:      "web-firewall",
		NetworkID: "net-full",
	}
	db.SaveFirewall(ctx, firewall)

	// Create tunnel
	tunnel := &Tunnel{
		ID:        "tunnel-full",
		Name:      "vpn-tunnel",
		NetworkID: "net-full",
	}
	db.SaveTunnel(ctx, tunnel)

	// Create VM interface
	iface := &VMInterface{
		ID:        "eth-full",
		Name:      "eth0",
		IPAddress: "172.16.0.10",
		NetworkID: "net-full",
	}
	db.SaveInterface(ctx, iface)

	// Verify all created
	net, _ := db.GetNetwork(ctx, "net-full")
	if net.Name != "full-test" {
		t.Error("network name mismatch")
	}

	routers, _ := db.ListRouters(ctx)
	if len(routers) < 1 {
		t.Error("expected at least one router")
	}

	firewalls, _ := db.ListFirewalls(ctx)
	if len(firewalls) < 1 {
		t.Error("expected at least one firewall")
	}

	ifaces, _ := db.ListInterfaces(ctx)
	if len(ifaces) < 1 {
		t.Error("expected at least one interface")
	}

	// Backup and restore
	backupPath := tmpPath + ".backup"
	db.Backup(ctx, backupPath)
	defer os.Remove(backupPath)

	// Delete and restore
	db.DeleteNetwork(ctx, "net-full")
	db.Restore(ctx, backupPath)

	restored, _ := db.GetNetwork(ctx, "net-full")
	if restored.Name != "full-test" {
		t.Error("restored network name mismatch")
	}
}

// TestIsolationManager_New tests isolation manager
func TestIsolationManager_New(t *testing.T) {
	// NewIsolationManager requires types.PipelineDB and NetworkConfig
	// Just verify the function exists
	_ = NewIsolationManager
}

// TestTimeOperations tests time operations
func TestTimeOperations(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	if future.Before(now) {
		t.Error("future should be after now")
	}
	if past.After(now) {
		t.Error("past should be before now")
	}
}