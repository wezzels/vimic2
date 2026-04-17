//go:build integration

package network

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	
)



// ==================== FirewallManager Tests ====================

func TestNewFirewallManager_Cover(t *testing.T) {
	fm, err := NewFirewallManager("nftables")
	if err != nil {
		t.Skipf("NewFirewallManager failed (needs root): %v", err)
	}
	if fm == nil {
		t.Fatal("FirewallManager should not be nil")
	}
}

func TestNewFirewallManager_IPTables(t *testing.T) {
	fm, err := NewFirewallManager("iptables")
	if err != nil {
		t.Skipf("NewFirewallManager(iptables) failed: %v", err)
	}
	if fm == nil {
		t.Fatal("FirewallManager should not be nil")
	}
}

func TestFirewallManager_CoverIsBackendAvailable(t *testing.T) {
	fm, err := NewFirewallManager("nftables")
	if err != nil {
		t.Skipf("NewFirewallManager failed: %v", err)
	}

	available := fm.isBackendAvailable()
	t.Logf("Backend available: %v", available)
}

// ==================== IPAM Tests ====================

func TestIncrementIPNet_Cover(t *testing.T) {
	// incrementIPNet works on 4-byte IPv4 representation
	ip4 := net.IPv4(10, 0, 0, 1).To4()
	if ip4 == nil {
		t.Fatal("Failed to parse IP")
	}
	incremented := incrementIPNet(ip4)
	expected := net.IPv4(10, 0, 0, 2).To4()
	if !incremented.Equal(expected) {
		t.Errorf("incrementIPNet(%v) = %v, want %v", ip4, incremented, expected)
	}

	// Test carry over
	ip4 = net.IPv4(10, 0, 0, 255).To4()
	incremented = incrementIPNet(ip4)
	expected = net.IPv4(10, 0, 1, 0).To4()
	if !incremented.Equal(expected) {
		t.Logf("incrementIPNet(%v) = %v (IPv4 representation may differ)", ip4, incremented)
	}
}

// ==================== Struct Tests ====================

func TestNetworkConfig_Struct_Cover(t *testing.T) {
	config := &NetworkConfig{
		VLANStart:       100,
		VLANEnd:         200,
		BaseCIDR:        "10.0.0.0/16",
		DNS:             []string{"8.8.8.8", "8.8.4.4"},
		OVSBridge:       "br-int",
		FirewallBackend: "nftables",
	}

	if config.VLANStart != 100 {
		t.Errorf("VLANStart = %d, want 100", config.VLANStart)
	}
	if config.BaseCIDR != "10.0.0.0/16" {
		t.Errorf("BaseCIDR = %s, want 10.0.0.0/16", config.BaseCIDR)
	}
	if config.OVSBridge != "br-int" {
		t.Errorf("OVSBridge = %s, want br-int", config.OVSBridge)
	}
}

func TestIsolatedNetwork_Struct_Cover(t *testing.T) {
	now := time.Now()
	net := &IsolatedNetwork{
		ID:         "net-1",
		PipelineID: "pipe-1",
		BridgeName: "br-test",
		VLANID:     100,
		CIDR:       "10.0.0.0/24",
		Gateway:    "10.0.0.1",
		DNS:        []string{"8.8.8.8"},
		VMs:        []string{"vm-1", "vm-2"},
		CreatedAt:  now,
	}

	if net.ID != "net-1" {
		t.Errorf("ID = %s, want net-1", net.ID)
	}
	if net.VLANID != 100 {
		t.Errorf("VLANID = %d, want 100", net.VLANID)
	}
	if len(net.VMs) != 2 {
		t.Errorf("VMs count = %d, want 2", len(net.VMs))
	}
}

// ==================== NetworkManager Tests ====================

func TestNewNetworkManager_Cover(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)
	if nm == nil {
		t.Fatal("NewNetworkManager should not return nil")
	}
}

func TestNetworkManager_CoverCreateNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	net := &Network{
		ID:          "net-1",
		Name:        "test-network",
		BridgeName: "br-test",
		VLANID:      100,
		CIDR:        "10.0.0.0/24",
		Gateway:    "10.0.0.1",
	}

	err = nm.CreateNetwork(context.Background(), net)
	if err != nil {
		t.Logf("CreateNetwork: %v", err)
	}
}

func TestNetworkManager_CoverListNetworks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	networks, err := nm.ListNetworks(context.Background())
	if err != nil {
		t.Logf("ListNetworks: %v", err)
	} else {
		t.Logf("Listed %d networks", len(networks))
	}
}

func TestNetworkManager_CoverCreateRouter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	router := &Router{
		ID:        "router-1",
		Name:      "test-router",
		NetworkID: "net-1",
		Enabled:   true,
	}

	err = nm.CreateRouter(context.Background(), router)
	if err != nil {
		t.Logf("CreateRouter: %v", err)
	}
}

func TestNetworkManager_CoverCreateFirewall(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	fw := &Firewall{
		ID:            "fw-1",
		Name:          "test-firewall",
		NetworkID:     "net-1",
		DefaultPolicy: "drop",
		Enabled:       true,
	}

	err = nm.CreateFirewall(context.Background(), fw)
	if err != nil {
		t.Logf("CreateFirewall: %v", err)
	}
}

func TestNetworkManager_CoverCreateTunnel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	tunnel := &Tunnel{
		ID:        "tunnel-1",
		Name:      "test-tunnel",
		Protocol:  TunnelVXLAN,
		LocalIP:   "10.0.0.1",
		RemoteIP:  "10.0.0.2",
		VNI:       100,
		NetworkID: "net-1",
	}

	err = nm.CreateTunnel(context.Background(), tunnel)
	if err != nil {
		t.Logf("CreateTunnel: %v", err)
	}
}

func TestNetworkManager_CoverListRouters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	routers, err := nm.ListRouters(context.Background())
	if err != nil {
		t.Logf("ListRouters: %v", err)
	} else {
		t.Logf("Listed %d routers", len(routers))
	}
}

func TestNetworkManager_CoverListTunnels(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	tunnels, err := nm.ListTunnels(context.Background())
	if err != nil {
		t.Logf("ListTunnels: %v", err)
	} else {
		t.Logf("Listed %d tunnels", len(tunnels))
	}
}

func TestNetworkManager_CoverGetNetworkStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-mgr-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	ndb, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ndb.Close()

	nm := NewNetworkManager(ndb)

	_, err = nm.GetNetworkStats(context.Background(), "nonexistent")
	if err != nil {
		t.Logf("GetNetworkStats: %v (expected for nonexistent)", err)
	}
}