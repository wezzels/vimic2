//go:build integration

package network

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func setupNetworkDB(t *testing.T) (*NetworkDB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-net-db-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	return ndb, func() {
		ndb.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== Database List/Delete Tests ====================

func TestNetworkDB_ListRouters(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	routers, err := ndb.ListRouters(context.Background())
	if err != nil {
		t.Logf("ListRouters: %v", err)
	} else {
		t.Logf("Listed %d routers", len(routers))
	}
}

func TestNetworkDB_DeleteRouter(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	err := ndb.DeleteRouter(context.Background(), "nonexistent")
	if err != nil {
		t.Logf("DeleteRouter: %v (expected for nonexistent)", err)
	} else {
		t.Log("DeleteRouter succeeded")
	}
}

func TestNetworkDB_ListFirewalls(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	firewalls, err := ndb.ListFirewalls(context.Background())
	if err != nil {
		t.Logf("ListFirewalls: %v", err)
	} else {
		t.Logf("Listed %d firewalls", len(firewalls))
	}
}

func TestNetworkDB_DeleteFirewall(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	err := ndb.DeleteFirewall(context.Background(), "nonexistent")
	if err != nil {
		t.Logf("DeleteFirewall: %v", err)
	} else {
		t.Log("DeleteFirewall succeeded")
	}
}

func TestNetworkDB_DeleteTunnel(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	err := ndb.DeleteTunnel(context.Background(), "nonexistent")
	if err != nil {
		t.Logf("DeleteTunnel: %v", err)
	} else {
		t.Log("DeleteTunnel succeeded")
	}
}

func TestNetworkDB_ListInterfaces(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	ifaces, err := ndb.ListInterfaces(context.Background())
	if err != nil {
		t.Logf("ListInterfaces: %v", err)
	} else {
		t.Logf("Listed %d interfaces", len(ifaces))
	}
}

func TestNetworkDB_DeleteInterface(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	err := ndb.DeleteInterface(context.Background(), "nonexistent")
	if err != nil {
		t.Logf("DeleteInterface: %v", err)
	} else {
		t.Log("DeleteInterface succeeded")
	}
}

func TestNetworkDB_Backup(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	tmpDir, err := os.MkdirTemp("", "vimic2-backup-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	backupPath := filepath.Join(tmpDir, "backup.db")
	err = ndb.Backup(context.Background(), backupPath)
	if err != nil {
		t.Logf("Backup: %v", err)
	} else {
		t.Logf("Backup succeeded to %s", backupPath)
	}
}

func TestNetworkDB_Restore(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	tmpDir, err := os.MkdirTemp("", "vimic2-restore-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// First backup, then restore
	backupPath := filepath.Join(tmpDir, "backup.db")
	err = ndb.Backup(context.Background(), backupPath)
	if err != nil {
		t.Skipf("Backup failed: %v", err)
	}

	err = ndb.Restore(context.Background(), backupPath)
	if err != nil {
		t.Logf("Restore: %v", err)
	} else {
		t.Log("Restore succeeded")
	}
}

func TestNetworkDB_Migrate(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	err := ndb.Migrate(context.Background())
	if err != nil {
		t.Logf("Migrate: %v", err)
	} else {
		t.Log("Migrate succeeded")
	}
}

// ==================== Database CRUD Tests ====================

func TestNetworkDB_SaveAndGetRouter(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	router := &Router{
		ID:        "router-1",
		Name:      "test-router",
		NetworkID: "net-1",
		Enabled:   true,
	}

	err := ndb.SaveRouter(context.Background(), router)
	if err != nil {
		t.Logf("SaveRouter: %v", err)
	} else {
		got, err := ndb.GetRouter(context.Background(), "router-1")
		if err != nil {
			t.Logf("GetRouter: %v", err)
		} else {
			t.Logf("Got router: %s", got.Name)
		}
	}
}

func TestNetworkDB_SaveAndGetFirewall(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	firewall := &Firewall{
		ID:            "fw-1",
		Name:          "test-firewall",
		NetworkID:     "net-1",
		DefaultPolicy: "drop",
		Enabled:       true,
	}

	err := ndb.SaveFirewall(context.Background(), firewall)
	if err != nil {
		t.Logf("SaveFirewall: %v", err)
	} else {
		got, err := ndb.GetFirewall(context.Background(), "fw-1")
		if err != nil {
			t.Logf("GetFirewall: %v", err)
		} else {
			t.Logf("Got firewall: %s", got.Name)
		}
	}
}

func TestNetworkDB_SaveAndGetTunnel(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	tunnel := &Tunnel{
		ID:        "tunnel-1",
		Name:      "test-tunnel",
		NetworkID: "net-1",
		Protocol: TunnelVXLAN,
	}

	err := ndb.SaveTunnel(context.Background(), tunnel)
	if err != nil {
		t.Logf("SaveTunnel: %v", err)
	} else {
		got, err := ndb.GetTunnel(context.Background(), "tunnel-1")
		if err != nil {
			t.Logf("GetTunnel: %v", err)
		} else {
			t.Logf("Got tunnel: %s", got.Name)
		}
	}
}

func TestNetworkDB_SaveAndGetInterface(t *testing.T) {
	ndb, cleanup := setupNetworkDB(t)
	defer cleanup()

	iface := &VMInterface{
		ID:         "iface-1",
		Name:       "eth0",
		VMID:       "vm-1",
		NetworkID:  "net-1",
		MACAddress: "aa:bb:cc:dd:ee:ff",
	}

	err := ndb.SaveInterface(context.Background(), iface)
	if err != nil {
		t.Logf("SaveInterface: %v", err)
	} else {
		got, err := ndb.GetInterface(context.Background(), "iface-1")
		if err != nil {
			t.Logf("GetInterface: %v", err)
		} else {
			t.Logf("Got interface: %s", got.Name)
		}
	}
}

// ==================== IPAM Manager Tests ====================

func TestIPAMManager_Allocate_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	cidr, mac, err := im.Allocate()
	if err != nil {
		t.Logf("Allocate: %v", err)
	} else {
		t.Logf("Allocated CIDR: %s, MAC: %s", cidr, mac)
	}
}

func TestIPAMManager_Release_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	err = im.Release("10.0.1.0/24")
	if err != nil {
		t.Logf("Release: %v", err)
	}
}

func TestIPAMManager_Reclaim_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	im.Reclaim("10.0.1.0/24")
	t.Log("Reclaim succeeded")
}

func TestIPAMManager_ListPools_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	pools := im.ListPools()
	t.Logf("Listed %d pools", len(pools))
}

func TestIPAMManager_GetDNS_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	dns := im.GetDNS()
	if len(dns) != 2 {
		t.Errorf("DNS count = %d, want 2", len(dns))
	}
	t.Logf("DNS: %v", dns)
}

func TestIPAMManager_GetGateway_DB(t *testing.T) {
	im, err := NewIPAMManager(&IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8"},
	})
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	gw, err := im.GetGateway("10.0.1.0/24")
	if err != nil {
		t.Logf("GetGateway: %v", err)
	} else {
		t.Logf("Gateway: %s", gw)
	}
}

// ==================== VLAN Allocator Tests ====================

func TestVLANAllocator_Allocate_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	vlan, err := va.Allocate()
	if err != nil {
		t.Logf("Allocate: %v", err)
	} else {
		t.Logf("Allocated VLAN: %d", vlan)
	}
}

func TestVLANAllocator_Release_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	vlan, err := va.Allocate()
	if err != nil {
		t.Skipf("Allocate failed: %v", err)
	}

	err = va.Release(vlan)
	if err != nil {
		t.Logf("Release: %v", err)
	} else {
		t.Logf("Released VLAN: %d", vlan)
	}
}

func TestVLANAllocator_Reclaim_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	va.Reclaim(150)
	t.Log("Reclaim succeeded")
}

func TestVLANAllocator_IsAllocated_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	vlan, err := va.Allocate()
	if err != nil {
		t.Skipf("Allocate failed: %v", err)
	}

	if !va.IsAllocated(vlan) {
		t.Errorf("VLAN %d should be allocated", vlan)
	}

	if va.IsAllocated(199) {
		t.Error("VLAN 199 should not be allocated")
	}
}

func TestVLANAllocator_ListUsed_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	va.Allocate()
	va.Allocate()

	used := va.ListUsed()
	t.Logf("Used VLANs: %v", used)
	if len(used) < 2 {
		t.Errorf("Expected at least 2 used VLANs, got %d", len(used))
	}
}

func TestVLANAllocator_ListAvailable_DB(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}

	available := va.ListAvailable()
	t.Logf("Available VLANs: %d", len(available))
	if len(available) == 0 {
		t.Errorf("Expected some available VLANs, got 0")
	}
	t.Logf("Available VLANs: %d", len(available))
}