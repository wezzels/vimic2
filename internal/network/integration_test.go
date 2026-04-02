// Package network provides integration tests for network management
package network

import (
	"context"
	"testing"
	"time"
)

// TestNetworkLifecycle tests creating, updating, and deleting networks
func TestNetworkLifecycle(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Test 1: Create network
	network := &Network{
		ID:         "net-test-001",
		Name:       "test-network",
		Type:       NetworkTypeBridge,
		BridgeName: "br-test",
		CIDR:       "10.100.0.0/24",
		Gateway:    "10.100.0.1",
		DNS:        []string{"8.8.8.8", "8.8.4.4"},
		VLANID:     100,
		VLANs:      []int{100, 200, 300},
	}

	// Save to database
	err = db.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatalf("Failed to save network: %v", err)
	}

	// Retrieve and verify
	loaded, err := db.GetNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("Failed to get network: %v", err)
	}

	if loaded.Name != network.Name {
		t.Errorf("Network name mismatch: got %s, want %s", loaded.Name, network.Name)
	}

	if loaded.CIDR != network.CIDR {
		t.Errorf("CIDR mismatch: got %s, want %s", loaded.CIDR, network.CIDR)
	}

	if loaded.VLANID != network.VLANID {
		t.Errorf("VLAN ID mismatch: got %d, want %d", loaded.VLANID, network.VLANID)
	}

	// Test 2: Update network
	loaded.Description = "Updated test network"
	loaded.DHCPEnabled = true
	loaded.DHCPStart = "10.100.0.100"
	loaded.DHCPEnd = "10.100.0.200"
	loaded.UpdatedAt = time.Now()

	err = db.SaveNetwork(ctx, loaded)
	if err != nil {
		t.Fatalf("Failed to update network: %v", err)
	}

	// Verify update
	updated, err := db.GetNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("Failed to get updated network: %v", err)
	}

	if updated.Description != "Updated test network" {
		t.Errorf("Description not updated")
	}

	if !updated.DHCPEnabled {
		t.Errorf("DHCP should be enabled")
	}

	// Test 3: Delete network
	err = db.DeleteNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("Failed to delete network: %v", err)
	}

	// Verify deletion
	_, err = db.GetNetwork(ctx, network.ID)
	if err == nil {
		t.Error("Network should be deleted")
	}
}

// TestRouterWithInterfaces tests router creation with multiple interfaces
func TestRouterWithInterfaces(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create networks first
	net1 := &Network{ID: "net-1", Name: "lan", CIDR: "10.0.0.0/24", Type: NetworkTypeBridge}
	net2 := &Network{ID: "net-2", Name: "wan", CIDR: "192.168.1.0/24", Type: NetworkTypeBridge}
	db.SaveNetwork(ctx, net1)
	db.SaveNetwork(ctx, net2)

	// Create router with interfaces
	router := &Router{
		ID:        "router-001",
		Name:      "edge-router",
		NetworkID: "net-1",
		Enabled:   true,
		Interfaces: []RouterInterface{
			{
				ID:         "eth0",
				Name:       "eth0",
				NetworkID:  "net-1",
				IPAddress:  "10.0.0.1/24",
				MACAddress: "00:11:22:33:44:55",
				VLANID:     100,
				Enabled:    true,
			},
			{
				ID:         "eth1",
				Name:       "eth1",
				NetworkID:  "net-2",
				IPAddress:  "192.168.1.1/24",
				MACAddress: "00:11:22:33:44:56",
				Enabled:    true,
			},
		},
		RoutingTable: []Route{
			{
				ID:          "route-default",
				Destination: "0.0.0.0/0",
				Gateway:     "192.168.1.254",
				Interface:   "eth1",
				Metric:      100,
				Type:        "static",
				Enabled:     true,
			},
			{
				ID:          "route-lan",
				Destination: "10.0.0.0/24",
				Interface:   "eth0",
				Metric:      50,
				Type:        "connected",
				Enabled:     true,
			},
		},
		NATRules: []NATRule{
			{
				ID:         "nat-masq",
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

	// Verify router
	loaded, err := db.GetRouter(ctx, router.ID)
	if err != nil {
		t.Fatalf("Failed to get router: %v", err)
	}

	if len(loaded.Interfaces) != 2 {
		t.Errorf("Expected 2 interfaces, got %d", len(loaded.Interfaces))
	}

	if len(loaded.RoutingTable) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(loaded.RoutingTable))
	}

	if len(loaded.NATRules) != 1 {
		t.Errorf("Expected 1 NAT rule, got %d", len(loaded.NATRules))
	}

	// Verify interface details
	iface0 := loaded.Interfaces[0]
	if iface0.IPAddress != "10.0.0.1/24" {
		t.Errorf("Interface 0 IP mismatch: got %s", iface0.IPAddress)
	}

	if iface0.VLANID != 100 {
		t.Errorf("Interface 0 VLAN mismatch: got %d", iface0.VLANID)
	}
}

// TestTunnelBetweenNetworks tests creating tunnels between networks
func TestTunnelBetweenNetworks(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create two networks
	net1 := &Network{ID: "net-site-a", Name: "site-a", CIDR: "10.1.0.0/16", Type: NetworkTypeBridge}
	net2 := &Network{ID: "net-site-b", Name: "site-b", CIDR: "10.2.0.0/16", Type: NetworkTypeBridge}
	db.SaveNetwork(ctx, net1)
	db.SaveNetwork(ctx, net2)

	// Create VXLAN tunnel
	vxlanTunnel := &Tunnel{
		ID:        "tun-vxlan-001",
		Name:      "site-a-to-site-b",
		Protocol:  TunnelVXLAN,
		LocalIP:   "192.168.100.1",
		RemoteIP:  "192.168.100.2",
		VNI:       5000,
		NetworkID: "net-site-a",
		Enabled:   true,
	}

	err = db.SaveTunnel(ctx, vxlanTunnel)
	if err != nil {
		t.Fatalf("Failed to save VXLAN tunnel: %v", err)
	}

	// Create GRE tunnel
	greTunnel := &Tunnel{
		ID:        "tun-gre-001",
		Name:      "backup-link",
		Protocol:  TunnelGRE,
		LocalIP:   "192.168.200.1",
		RemoteIP:  "192.168.200.2",
		VNI:       100,
		NetworkID: "net-site-a",
		Enabled:   true,
	}

	err = db.SaveTunnel(ctx, greTunnel)
	if err != nil {
		t.Fatalf("Failed to save GRE tunnel: %v", err)
	}

	// List tunnels
	tunnels, err := db.ListTunnels(ctx)
	if err != nil {
		t.Fatalf("Failed to list tunnels: %v", err)
	}

	if len(tunnels) != 2 {
		t.Errorf("Expected 2 tunnels, got %d", len(tunnels))
	}

	// Verify VXLAN tunnel
	vxlan, err := db.GetTunnel(ctx, "tun-vxlan-001")
	if err != nil {
		t.Fatalf("Failed to get VXLAN tunnel: %v", err)
	}

	if vxlan.Protocol != TunnelVXLAN {
		t.Errorf("Expected VXLAN protocol, got %s", vxlan.Protocol)
	}

	if vxlan.VNI != 5000 {
		t.Errorf("VNI mismatch: got %d, want 5000", vxlan.VNI)
	}
}

// TestFirewallRules tests creating firewalls with rules
func TestFirewallRules(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create network
	net := &Network{ID: "net-fw", Name: "protected-net", CIDR: "10.50.0.0/24", Type: NetworkTypeBridge}
	db.SaveNetwork(ctx, net)

	// Create firewall with rules
	firewall := &Firewall{
		ID:            "fw-001",
		Name:          "web-firewall",
		NetworkID:     "net-fw",
		DefaultPolicy: "drop",
		Enabled:       true,
		Logging:       true,
		Rules: []FirewallRule{
			{
				ID:         "rule-allow-web",
				Name:       "allow-web",
				Direction:  "ingress",
				Protocol:   "tcp",
				DestCIDR:   "10.50.0.0/24",
				DestPort:   80,
				Action:     "accept",
				Priority:   100,
				Enabled:    true,
				Log:        true,
			},
			{
				ID:         "rule-allow-https",
				Name:       "allow-https",
				Direction:  "ingress",
				Protocol:   "tcp",
				DestCIDR:   "10.50.0.0/24",
				DestPort:   443,
				Action:     "accept",
				Priority:   90,
				Enabled:    true,
			},
			{
				ID:         "rule-allow-ssh",
				Name:       "allow-ssh",
				Direction:  "ingress",
				Protocol:   "tcp",
				SourceCIDR: "10.0.0.0/8",
				DestCIDR:   "10.50.0.0/24",
				DestPort:   22,
				Action:     "accept",
				Priority:   80,
				Enabled:    true,
			},
			{
				ID:         "rule-deny-all",
				Name:       "deny-all-ingress",
				Direction:  "ingress",
				Protocol:   "all",
				Action:     "drop",
				Priority:   1000,
				Enabled:    true,
			},
		},
	}

	err = db.SaveFirewall(ctx, firewall)
	if err != nil {
		t.Fatalf("Failed to save firewall: %v", err)
	}

	// Verify firewall
	loaded, err := db.GetFirewall(ctx, firewall.ID)
	if err != nil {
		t.Fatalf("Failed to get firewall: %v", err)
	}

	if loaded.DefaultPolicy != "drop" {
		t.Errorf("Expected drop policy, got %s", loaded.DefaultPolicy)
	}

	if len(loaded.Rules) != 4 {
		t.Errorf("Expected 4 rules, got %d", len(loaded.Rules))
	}

	// Verify rule priorities (should be evaluated in order)
	rules := loaded.Rules
	if rules[0].Priority != 100 {
		t.Errorf("First rule priority should be 100, got %d", rules[0].Priority)
	}
}

// TestVMInterfaceAssignment tests assigning VM interfaces to networks
func TestVMInterfaceAssignment(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create network
	net := &Network{
		ID:         "net-vm-test",
		Name:       "vm-network",
		CIDR:       "10.200.0.0/24",
		Gateway:    "10.200.0.1",
		Type:       NetworkTypeBridge,
		VLANID:     200,
	}
	db.SaveNetwork(ctx, net)

	// Create VM interface
	iface := &VMInterface{
		ID:           "vif-001",
		VMID:         "vm-001",
		Name:         "eth0",
		MACAddress:   "52:54:00:12:34:56",
		IPAddress:    "10.200.0.10",
		NetworkID:    "net-vm-test",
		VLANID:       200,
		MTU:          1500,
		Bandwidth:    1000, // 1 Gbps
		State:        InterfaceUp,
		PortSecurity: true,
	}

	err = db.SaveInterface(ctx, iface)
	if err != nil {
		t.Fatalf("Failed to save interface: %v", err)
	}

	// Verify interface
	loaded, err := db.GetInterface(ctx, iface.ID)
	if err != nil {
		t.Fatalf("Failed to get interface: %v", err)
	}

	if loaded.VMID != "vm-001" {
		t.Errorf("VM ID mismatch: got %s", loaded.VMID)
	}

	if loaded.NetworkID != "net-vm-test" {
		t.Errorf("Network ID mismatch: got %s", loaded.NetworkID)
	}

	if loaded.Bandwidth != 1000 {
		t.Errorf("Bandwidth mismatch: got %d", loaded.Bandwidth)
	}

	// Create trunk interface (multiple VLANs)
	trunkIface := &VMInterface{
		ID:          "vif-002",
		VMID:        "vm-001",
		Name:        "eth1",
		MACAddress:  "52:54:00:12:34:57",
		NetworkID:   "net-vm-test",
		TrunkVLANs:  []int{100, 200, 300, 400},
		MTU:         9000, // Jumbo frames
		State:       InterfaceUp,
	}

	err = db.SaveInterface(ctx, trunkIface)
	if err != nil {
		t.Fatalf("Failed to save trunk interface: %v", err)
	}

	// Verify trunk
	trunkLoaded, err := db.GetInterface(ctx, trunkIface.ID)
	if err != nil {
		t.Fatalf("Failed to get trunk interface: %v", err)
	}

	if len(trunkLoaded.TrunkVLANs) != 4 {
		t.Errorf("Expected 4 trunk VLANs, got %d", len(trunkLoaded.TrunkVLANs))
	}

	if trunkLoaded.MTU != 9000 {
		t.Errorf("MTU should be 9000 for jumbo frames, got %d", trunkLoaded.MTU)
	}
}

// TestComplexTopology tests a complex network topology
func TestComplexTopology(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create multiple networks
	networks := []*Network{
		{ID: "net-mgmt", Name: "management", CIDR: "10.0.0.0/24", Type: NetworkTypeBridge, VLANID: 10},
		{ID: "net-prod", Name: "production", CIDR: "10.100.0.0/16", Type: NetworkTypeBridge, VLANID: 100},
		{ID: "net-dmz", Name: "dmz", CIDR: "10.200.0.0/24", Type: NetworkTypeBridge, VLANID: 200},
		{ID: "net-storage", Name: "storage", CIDR: "10.50.0.0/24", Type: NetworkTypeBridge, VLANID: 50},
	}

	for _, net := range networks {
		err = db.SaveNetwork(ctx, net)
		if err != nil {
			t.Fatalf("Failed to save network %s: %v", net.ID, err)
		}
	}

	// Create router connecting networks
	router := &Router{
		ID:      "router-core",
		Name:    "core-router",
		Enabled: true,
		Interfaces: []RouterInterface{
			{ID: "eth0", Name: "eth0", NetworkID: "net-mgmt", IPAddress: "10.0.0.1/24", Enabled: true},
			{ID: "eth1", Name: "eth1", NetworkID: "net-prod", IPAddress: "10.100.0.1/16", Enabled: true},
			{ID: "eth2", Name: "eth2", NetworkID: "net-dmz", IPAddress: "10.200.0.1/24", Enabled: true},
			{ID: "eth3", Name: "eth3", NetworkID: "net-storage", IPAddress: "10.50.0.1/24", Enabled: true},
		},
		RoutingTable: []Route{
			{ID: "default", Destination: "0.0.0.0/0", Gateway: "10.0.0.254", Interface: "eth0", Metric: 100, Type: "static", Enabled: true},
		},
	}

	err = db.SaveRouter(ctx, router)
	if err != nil {
		t.Fatalf("Failed to save router: %v", err)
	}

	// Create tunnels
	tunnels := []*Tunnel{
		{ID: "tun-1", Name: "prod-to-dmz", Protocol: TunnelVXLAN, LocalIP: "10.100.0.1", RemoteIP: "10.200.0.1", VNI: 100, NetworkID: "net-prod", Enabled: true},
		{ID: "tun-2", Name: "storage-backup", Protocol: TunnelGRE, LocalIP: "10.50.0.1", RemoteIP: "10.50.0.2", VNI: 50, NetworkID: "net-storage", Enabled: true},
	}

	for _, tun := range tunnels {
		err = db.SaveTunnel(ctx, tun)
		if err != nil {
			t.Fatalf("Failed to save tunnel %s: %v", tun.ID, err)
		}
	}

	// Create firewall for DMZ
	firewall := &Firewall{
		ID:            "fw-dmz",
		Name:          "dmz-firewall",
		NetworkID:     "net-dmz",
		DefaultPolicy: "drop",
		Enabled:       true,
		Rules: []FirewallRule{
			{ID: "allow-http", Name: "allow-http", Direction: "ingress", Protocol: "tcp", DestPort: 80, Action: "accept", Priority: 100, Enabled: true},
			{ID: "allow-https", Name: "allow-https", Direction: "ingress", Protocol: "tcp", DestPort: 443, Action: "accept", Priority: 90, Enabled: true},
		},
	}

	err = db.SaveFirewall(ctx, firewall)
	if err != nil {
		t.Fatalf("Failed to save firewall: %v", err)
	}

	// Verify topology
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["networks"] != 4 {
		t.Errorf("Expected 4 networks, got %d", stats["networks"])
	}

	if stats["routers"] != 1 {
		t.Errorf("Expected 1 router, got %d", stats["routers"])
	}

	if stats["tunnels"] != 2 {
		t.Errorf("Expected 2 tunnels, got %d", stats["tunnels"])
	}

	if stats["firewalls"] != 1 {
		t.Errorf("Expected 1 firewall, got %d", stats["firewalls"])
	}

	t.Logf("Complex topology: %d networks, %d routers, %d tunnels, %d firewalls",
		stats["networks"], stats["routers"], stats["tunnels"], stats["firewalls"])
}

// TestNATConfiguration tests NAT rule configuration
func TestNATConfiguration(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create router with various NAT rules
	router := &Router{
		ID:      "router-nat",
		Name:    "nat-router",
		Enabled: true,
		Interfaces: []RouterInterface{
			{ID: "eth0", Name: "eth0", IPAddress: "10.0.0.1/24", Enabled: true},
			{ID: "eth1", Name: "eth1", IPAddress: "192.168.1.100/24", Enabled: true},
		},
		NATRules: []NATRule{
			{
				ID:         "nat-1",
				Type:       "masquerade",
				SourceCIDR: "10.0.0.0/24",
				Enabled:    true,
			},
			{
				ID:           "nat-2",
				Type:         "dnat",
				ExternalIP:   "192.168.1.100",
				ExternalPort: 80,
				InternalIP:   "10.0.0.10",
				InternalPort: 8080,
				Protocol:     "tcp",
				Enabled:      true,
			},
			{
				ID:         "nat-3",
				Type:       "snat",
				SourceCIDR: "10.0.0.0/24",
				ExternalIP: "192.168.1.100",
				Enabled:    true,
			},
		},
	}

	err = db.SaveRouter(ctx, router)
	if err != nil {
		t.Fatalf("Failed to save router: %v", err)
	}

	// Verify NAT rules
	loaded, err := db.GetRouter(ctx, router.ID)
	if err != nil {
		t.Fatalf("Failed to get router: %v", err)
	}

	if len(loaded.NATRules) != 3 {
		t.Errorf("Expected 3 NAT rules, got %d", len(loaded.NATRules))
	}

	// Verify DNAT rule
	dnatRule := loaded.NATRules[1]
	if dnatRule.Type != "dnat" {
		t.Errorf("Expected DNAT type, got %s", dnatRule.Type)
	}

	if dnatRule.ExternalPort != 80 {
		t.Errorf("External port should be 80, got %d", dnatRule.ExternalPort)
	}

	if dnatRule.InternalPort != 8080 {
		t.Errorf("Internal port should be 8080, got %d", dnatRule.InternalPort)
	}
}

// TestBackupRestore tests network configuration backup/restore
func TestBackupRestore(t *testing.T) {
	db, err := NewNetworkDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create configuration
	net := &Network{ID: "net-1", Name: "backup-test", CIDR: "10.0.0.0/24", Type: NetworkTypeBridge}
	router := &Router{ID: "router-1", Name: "test-router", Enabled: true}
	firewall := &Firewall{ID: "fw-1", Name: "test-fw", DefaultPolicy: "drop", Enabled: true}
	tunnel := &Tunnel{ID: "tun-1", Name: "test-tunnel", Protocol: TunnelVXLAN, LocalIP: "1.1.1.1", RemoteIP: "2.2.2.2", VNI: 100, Enabled: true}

	db.SaveNetwork(ctx, net)
	db.SaveRouter(ctx, router)
	db.SaveFirewall(ctx, firewall)
	db.SaveTunnel(ctx, tunnel)

	// Test database stats
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	expectedCounts := map[string]int{
		"networks":  1,
		"routers":   1,
		"firewalls": 1,
		"tunnels":   1,
	}

	for key, expected := range expectedCounts {
		if stats[key] != expected {
			t.Errorf("Stats %s: got %d, want %d", key, stats[key], expected)
		}
	}

	// Verify all entities exist
	_, err = db.GetNetwork(ctx, "net-1")
	if err != nil {
		t.Errorf("Network should exist: %v", err)
	}

	_, err = db.GetRouter(ctx, "router-1")
	if err != nil {
		t.Errorf("Router should exist: %v", err)
	}

	_, err = db.GetFirewall(ctx, "fw-1")
	if err != nil {
		t.Errorf("Firewall should exist: %v", err)
	}

	_, err = db.GetTunnel(ctx, "tun-1")
	if err != nil {
		t.Errorf("Tunnel should exist: %v", err)
	}
}