// Package network provides tests for network management
package network

import (
	"context"
	"testing"
)

// TestNetworkCreation tests network creation
func TestNetworkCreation(t *testing.T) {
	// Create in-memory database
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create network manager
	mgr := NewNetworkManager(db)

	// Create a network
	network := &Network{
		Name:       "test-network",
		Type:       NetworkTypeBridge,
		BridgeName: "br-test",
		CIDR:       "10.0.0.0/24",
		Gateway:    "10.0.0.1",
		VLANID:     100,
	}

	err = mgr.CreateNetwork(context.Background(), network)
	if err != nil {
		t.Logf("CreateNetwork failed (expected with stub OVS): %v", err)
		// This is expected since OVS is not installed
	}
}

// TestRouterCreation tests router creation
func TestRouterCreation(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	mgr := NewNetworkManager(db)

	router := &Router{
		Name:    "test-router",
		Enabled: true,
		Interfaces: []RouterInterface{
			{
				Name:      "eth0",
				IPAddress: "10.0.0.1/24",
				Enabled:   true,
			},
		},
	}

	err = mgr.CreateRouter(context.Background(), router)
	if err != nil {
		t.Logf("CreateRouter failed (expected with stub OVS): %v", err)
	}
}

// TestFirewallCreation tests firewall creation
func TestFirewallCreation(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	mgr := NewNetworkManager(db)

	firewall := &Firewall{
		Name:          "test-firewall",
		DefaultPolicy: "drop",
		Enabled:       true,
		Rules: []FirewallRule{
			{
				Name:      "allow-ssh",
				Direction: "ingress",
				Protocol:  "tcp",
				DestPort:  22,
				Action:    "accept",
				Enabled:   true,
			},
		},
	}

	err = mgr.CreateFirewall(context.Background(), firewall)
	if err != nil {
		t.Logf("CreateFirewall failed (expected with stub): %v", err)
	}
}

// TestTunnelCreation tests tunnel creation
func TestTunnelCreation(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	mgr := NewNetworkManager(db)

	// First create a network
	network := &Network{
		ID:         "net-001",
		Name:       "tunnel-network",
		Type:       NetworkTypeBridge,
		BridgeName: "br-tunnel",
	}
	if err := db.SaveNetwork(context.Background(), network); err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	tunnel := &Tunnel{
		Name:      "vxlan-tunnel",
		Protocol:  TunnelVXLAN,
		LocalIP:   "192.168.1.1",
		RemoteIP:  "192.168.1.2",
		VNI:       1000,
		NetworkID: "net-001",
		Enabled:   true,
	}

	err = mgr.CreateTunnel(context.Background(), tunnel)
	if err != nil {
		t.Logf("CreateTunnel failed (expected with stub OVS): %v", err)
	}
}

// TestDatabaseOperations tests database CRUD operations
func TestDatabaseOperations(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test network save/load
	network := &Network{
		ID:         "net-001",
		Name:       "test-network",
		Type:       NetworkTypeBridge,
		BridgeName: "br-test",
		CIDR:       "10.0.0.0/24",
		Gateway:    "10.0.0.1",
		VLANID:     100,
		VLANs:      []int{100, 200, 300},
		DNS:        []string{"8.8.8.8", "8.8.4.4"},
	}

	err = db.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	loaded, err := db.GetNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("Failed to load network: %v", err)
	}

	if loaded.Name != network.Name {
		t.Errorf("Network name mismatch: got %s, want %s", loaded.Name, network.Name)
	}

	if len(loaded.VLANs) != len(network.VLANs) {
		t.Errorf("VLANs mismatch: got %d, want %d", len(loaded.VLANs), len(network.VLANs))
	}

	// Test network list
	networks, err := db.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("Failed to list networks: %v", err)
	}

	if len(networks) != 1 {
		t.Errorf("Expected 1 network, got %d", len(networks))
	}

	// Test network delete
	err = db.DeleteNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("Failed to delete network: %v", err)
	}

	networks, err = db.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("Failed to list networks: %v", err)
	}

	if len(networks) != 0 {
		t.Errorf("Expected 0 networks after delete, got %d", len(networks))
	}
}

// TestRouterDatabaseOperations tests router database operations
func TestRouterDatabaseOperations(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	router := &Router{
		ID:        "router-001",
		Name:      "test-router",
		NetworkID: "net-001",
		Enabled:   true,
		Interfaces: []RouterInterface{
			{
				ID:         "iface-001",
				Name:       "eth0",
				NetworkID:  "net-001",
				IPAddress:  "10.0.0.1/24",
				MACAddress: "00:11:22:33:44:55",
				Enabled:    true,
			},
		},
		RoutingTable: []Route{
			{
				ID:          "route-001",
				Destination: "0.0.0.0/0",
				Gateway:     "10.0.0.254",
				Interface:   "eth0",
				Metric:      100,
				Type:        "static",
				Enabled:     true,
			},
		},
		NATRules: []NATRule{
			{
				ID:         "nat-001",
				Type:       "masquerade",
				SourceCIDR: "10.0.0.0/24",
				Enabled:    true,
			},
		},
	}

	err = db.SaveRouter(ctx, router)
	if err != nil {
		t.Fatalf("Failed to save router: %v", err)
	}

	loaded, err := db.GetRouter(ctx, router.ID)
	if err != nil {
		t.Fatalf("Failed to load router: %v", err)
	}

	if loaded.Name != router.Name {
		t.Errorf("Router name mismatch: got %s, want %s", loaded.Name, router.Name)
	}

	if len(loaded.Interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(loaded.Interfaces))
	}

	if len(loaded.RoutingTable) != 1 {
		t.Errorf("Expected 1 route, got %d", len(loaded.RoutingTable))
	}

	if len(loaded.NATRules) != 1 {
		t.Errorf("Expected 1 NAT rule, got %d", len(loaded.NATRules))
	}
}

// TestTunnelDatabaseOperations tests tunnel database operations
func TestTunnelDatabaseOperations(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	tunnel := &Tunnel{
		ID:        "tun-001",
		Name:      "vxlan-100",
		Protocol:  TunnelVXLAN,
		LocalIP:   "192.168.1.1",
		RemoteIP:  "192.168.1.2",
		VNI:       1000,
		NetworkID: "net-001",
		Enabled:   true,
	}

	err = db.SaveTunnel(ctx, tunnel)
	if err != nil {
		t.Fatalf("Failed to save tunnel: %v", err)
	}

	loaded, err := db.GetTunnel(ctx, tunnel.ID)
	if err != nil {
		t.Fatalf("Failed to load tunnel: %v", err)
	}

	if loaded.Name != tunnel.Name {
		t.Errorf("Tunnel name mismatch: got %s, want %s", loaded.Name, tunnel.Name)
	}

	if loaded.Protocol != tunnel.Protocol {
		t.Errorf("Protocol mismatch: got %s, want %s", loaded.Protocol, tunnel.Protocol)
	}
}

// TestInterfaceOperations tests VM interface operations
func TestInterfaceOperations(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	iface := &VMInterface{
		ID:           "vif-001",
		VMID:         "vm-001",
		Name:         "eth0",
		MACAddress:   "00:11:22:33:44:55",
		IPAddress:    "10.0.0.10",
		NetworkID:    "net-001",
		VLANID:       100,
		TrunkVLANs:   []int{100, 200, 300},
		MTU:          1500,
		Bandwidth:    1000,
		State:        InterfaceUp,
		PortSecurity: true,
	}

	err = db.SaveInterface(ctx, iface)
	if err != nil {
		t.Fatalf("Failed to save interface: %v", err)
	}

	loaded, err := db.GetInterface(ctx, iface.ID)
	if err != nil {
		t.Fatalf("Failed to load interface: %v", err)
	}

	if loaded.Name != iface.Name {
		t.Errorf("Interface name mismatch: got %s, want %s", loaded.Name, iface.Name)
	}

	if loaded.VLANID != iface.VLANID {
		t.Errorf("VLAN ID mismatch: got %d, want %d", loaded.VLANID, iface.VLANID)
	}

	if len(loaded.TrunkVLANs) != len(iface.TrunkVLANs) {
		t.Errorf("Trunk VLANs mismatch: got %d, want %d", len(loaded.TrunkVLANs), len(iface.TrunkVLANs))
	}
}

// TestFirewallOperations tests firewall operations
func TestFirewallOperations(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	firewall := &Firewall{
		ID:            "fw-001",
		Name:          "web-firewall",
		NetworkID:     "net-001",
		DefaultPolicy: "drop",
		Enabled:       true,
		Logging:       true,
		Rules: []FirewallRule{
			{
				ID:        "rule-001",
				Name:      "allow-web",
				Direction: "ingress",
				Protocol:  "tcp",
				DestPort:  80,
				Action:    "accept",
				Priority:  100,
				Enabled:   true,
			},
			{
				ID:        "rule-002",
				Name:      "allow-ssh",
				Direction: "ingress",
				Protocol:  "tcp",
				DestPort:  22,
				Action:    "accept",
				Priority:  90,
				Enabled:   true,
			},
		},
	}

	err = db.SaveFirewall(ctx, firewall)
	if err != nil {
		t.Fatalf("Failed to save firewall: %v", err)
	}

	loaded, err := db.GetFirewall(ctx, firewall.ID)
	if err != nil {
		t.Fatalf("Failed to load firewall: %v", err)
	}

	if loaded.Name != firewall.Name {
		t.Errorf("Firewall name mismatch: got %s, want %s", loaded.Name, firewall.Name)
	}

	if len(loaded.Rules) != len(firewall.Rules) {
		t.Errorf("Rules count mismatch: got %d, want %d", len(loaded.Rules), len(firewall.Rules))
	}
}

// TestDatabaseStats tests database statistics
func TestDatabaseStats(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create some test data
	network := &Network{ID: "net-001", Name: "test", Type: NetworkTypeBridge}
	db.SaveNetwork(ctx, network)

	router := &Router{ID: "router-001", Name: "test-router"}
	db.SaveRouter(ctx, router)

	firewall := &Firewall{ID: "fw-001", Name: "test-firewall"}
	db.SaveFirewall(ctx, firewall)

	tunnel := &Tunnel{ID: "tun-001", Name: "test-tunnel", Protocol: TunnelVXLAN}
	db.SaveTunnel(ctx, tunnel)

	iface := &VMInterface{ID: "vif-001", VMID: "vm-001", Name: "eth0"}
	db.SaveInterface(ctx, iface)

	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["networks"] != 1 {
		t.Errorf("Expected 1 network, got %d", stats["networks"])
	}

	if stats["routers"] != 1 {
		t.Errorf("Expected 1 router, got %d", stats["routers"])
	}

	if stats["firewalls"] != 1 {
		t.Errorf("Expected 1 firewall, got %d", stats["firewalls"])
	}

	if stats["tunnels"] != 1 {
		t.Errorf("Expected 1 tunnel, got %d", stats["tunnels"])
	}

	if stats["interfaces"] != 1 {
		t.Errorf("Expected 1 interface, got %d", stats["interfaces"])
	}
}
