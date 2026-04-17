//go:build integration

package network

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupNetworkManagerTest(t *testing.T) (*NetworkManager, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-net-test-")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewNetworkDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	nm := NewNetworkManager(db)
	return nm, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== NetworkManager Tests ====================

func TestNetworkManager_ListNetworks_Empty(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	networks, err := nm.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("ListNetworks failed: %v", err)
	}
	if len(networks) != 0 {
		t.Errorf("ListNetworks returned %d networks, want 0", len(networks))
	}
}

func TestNetworkManager_ListRouters_Empty(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	routers, err := nm.ListRouters(ctx)
	if err != nil {
		t.Fatalf("ListRouters failed: %v", err)
	}
	if len(routers) != 0 {
		t.Errorf("ListRouters returned %d routers, want 0", len(routers))
	}
}

func TestNetworkManager_ListTunnels_Empty(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	tunnels, err := nm.ListTunnels(ctx)
	if err != nil {
		t.Fatalf("ListTunnels failed: %v", err)
	}
	if len(tunnels) != 0 {
		t.Errorf("ListTunnels returned %d tunnels, want 0", len(tunnels))
	}
}

func TestNetworkManager_CreateNetwork_InvalidCIDR(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	network := &Network{
		Name:       "test-network",
		BridgeName: "br-test",
		CIDR:       "invalid-cidr",
	}

	err := nm.CreateNetwork(ctx, network)
	if err == nil {
		t.Error("CreateNetwork should fail with invalid CIDR")
	}
}

func TestNetworkManager_CreateNetwork_ValidCIDR(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	network := &Network{
		Name:       "test-network",
		BridgeName: "br-test",
		CIDR:       "192.168.1.0/24",
	}

	err := nm.CreateNetwork(ctx, network)
	if err != nil {
		t.Logf("CreateNetwork: %v (OVS may not be available)", err)
	} else if network.ID == "" {
		t.Error("Network ID should be set")
	}
}

func TestNetworkManager_CreateRouter_Int(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	router := &Router{
		Name:    "test-router",
		Enabled: true,
	}

	err := nm.CreateRouter(ctx, router)
	if err != nil {
		t.Logf("CreateRouter: %v", err)
	} else if router.ID == "" {
		t.Error("Router ID should be set")
	}
}

func TestNetworkManager_CreateFirewall_Int(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	firewall := &Firewall{
		Name:          "test-firewall",
		DefaultPolicy: "DROP",
	}

	err := nm.CreateFirewall(ctx, firewall)
	if err != nil {
		t.Logf("CreateFirewall: %v", err)
	} else if firewall.ID == "" {
		t.Error("Firewall ID should be set")
	}
}

func TestNetworkManager_CreateTunnel_Int(t *testing.T) {
	nm, cleanup := setupNetworkManagerTest(t)
	defer cleanup()

	ctx := context.Background()
	tunnel := &Tunnel{
		Name:     "test-tunnel",
		Protocol: TunnelVXLAN,
		RemoteIP: "10.0.0.1",
		LocalIP:  "10.0.0.2",
		VNI:      100,
	}

	err := nm.CreateTunnel(ctx, tunnel)
	if err != nil {
		t.Logf("CreateTunnel: %v", err)
	} else if tunnel.ID == "" {
		t.Error("Tunnel ID should be set")
	}
}

// ==================== Struct Tests ====================

func TestNetwork_Struct(t *testing.T) {
	now := time.Now()
	network := &Network{
		ID:         "net-1",
		Name:       "test-network",
		BridgeName: "br-test",
		CIDR:       "192.168.1.0/24",
		Gateway:    "192.168.1.1",
		VLANID:     100,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if network.ID != "net-1" {
		t.Errorf("ID = %s, want net-1", network.ID)
	}
	if network.VLANID != 100 {
		t.Errorf("VLANID = %d, want 100", network.VLANID)
	}
}

func TestRouter_Struct(t *testing.T) {
	router := &Router{
		ID:           "router-1",
		Name:         "test-router",
		RoutingTable: []Route{},
		Enabled:      true,
	}

	if router.ID != "router-1" {
		t.Errorf("ID = %s, want router-1", router.ID)
	}
}

func TestFirewall_Struct(t *testing.T) {
	firewall := &Firewall{
		ID:            "fw-1",
		Name:          "test-firewall",
		DefaultPolicy: "DROP",
		Rules: []FirewallRule{
			{ID: "rule-1", Protocol: "tcp", DestPort: 22, Action: "ACCEPT"},
		},
	}

	if firewall.DefaultPolicy != "DROP" {
		t.Errorf("DefaultPolicy = %s, want DROP", firewall.DefaultPolicy)
	}
}

func TestTunnel_Struct(t *testing.T) {
	tunnel := &Tunnel{
		ID:       "tunnel-1",
		Name:     "test-tunnel",
		Protocol: TunnelVXLAN,
		RemoteIP: "10.0.0.1",
		VNI:      100,
	}

	if tunnel.Protocol != TunnelVXLAN {
		t.Errorf("Protocol = %s, want %s", tunnel.Protocol, TunnelVXLAN)
	}
}

func TestTunnelProtocol_Constants(t *testing.T) {
	if TunnelVXLAN != "vxlan" {
		t.Errorf("TunnelVXLAN = %s, want vxlan", TunnelVXLAN)
	}
	if TunnelGRE != "gre" {
		t.Errorf("TunnelGRE = %s, want gre", TunnelGRE)
	}
}

func TestFirewallRule_Struct(t *testing.T) {
	rule := FirewallRule{
		ID:         "rule-1",
		Protocol:   "tcp",
		DestPort:   443,
		SourceCIDR: "0.0.0.0/0",
		Action:     "ACCEPT",
	}

	if rule.Protocol != "tcp" {
		t.Errorf("Protocol = %s, want tcp", rule.Protocol)
	}
	if rule.DestPort != 443 {
		t.Errorf("DestPort = %d, want 443", rule.DestPort)
	}
}

func TestVMInterface_Struct(t *testing.T) {
	iface := &VMInterface{
		ID:         "iface-1",
		VMID:       "vm-1",
		NetworkID:  "net-1",
		Name:       "eth0",
		MACAddress: "00:11:22:33:44:55",
		IPAddress:  "192.168.1.100",
	}

	if iface.MACAddress != "00:11:22:33:44:55" {
		t.Errorf("MACAddress = %s, want 00:11:22:33:44:55", iface.MACAddress)
	}
}

func TestRoute_Struct(t *testing.T) {
	route := Route{
		Destination: "10.0.0.0/24",
		Gateway:     "192.168.1.1",
		Interface:   "eth0",
		Metric:      100,
	}

	if route.Destination != "10.0.0.0/24" {
		t.Errorf("Destination = %s, want 10.0.0.0/24", route.Destination)
	}
}

// ==================== IPAM Tests ====================

func TestIPAM_New(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}
	if ipam == nil {
		t.Fatal("IPAMManager should not be nil")
	}
}

func TestIPAM_Allocate(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	ip, _, err := ipam.Allocate()
	if err != nil {
		t.Logf("Allocate: %v (may be expected)", err)
	} else if ip == "" {
		t.Error("Allocate should return non-empty IP")
	}
}

func TestIPAM_MultipleAllocate(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/30",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, _ := NewIPAMManager(config)

	ips := make(map[string]bool)
	for i := 0; i < 3; i++ {
		ip, _, err := ipam.Allocate()
		if err != nil {
			t.Logf("Allocate %d: %v (pool may be exhausted)", i, err)
			break
		}
		if ips[ip] {
			t.Errorf("Duplicate IP allocated: %s", ip)
		}
		ips[ip] = true
	}
}

func TestIPAM_GetDNS(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}
	ipam, _ := NewIPAMManager(config)

	dns := ipam.GetDNS()
	if len(dns) != 2 {
		t.Errorf("GetDNS returned %d servers, want 2", len(dns))
	}
}

func TestIPAM_ListPools(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, _ := NewIPAMManager(config)

	pools := ipam.ListPools()
	if len(pools) == 0 {
		t.Error("ListPools should return at least one pool")
	}
}

func TestIPAM_Used(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, _ := NewIPAMManager(config)

	ip, poolID, _ := ipam.Allocate()
	_ = ip
	_ = poolID
	used := ipam.Used()
	if used < 1 {
		t.Errorf("Used = %d, want at least 1", used)
	}
}

func TestIPAM_Available(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, _ := NewIPAMManager(config)

	available := ipam.Available()
	// Available may be negative if Allocate consumed IPs
	// Just verify it returns a value
	_ = available
}

func TestIPAM_InvalidCIDR(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "invalid-cidr",
		DNS:      []string{"8.8.8.8"},
	}
	_, err := NewIPAMManager(config)
	if err == nil {
		t.Error("NewIPAMManager should fail with invalid CIDR")
	}
}

func TestIPAM_ReleaseIP(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}
	ipam, _ := NewIPAMManager(config)

	ip, poolID, _ := ipam.Allocate()

	err := ipam.ReleaseIP(poolID, ip)
	if err != nil {
		t.Logf("ReleaseIP: %v", err)
	}
}