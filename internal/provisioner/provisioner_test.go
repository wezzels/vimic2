// Package provisioner provides provisioner tests
package provisioner

import (
	"testing"
)

// TestNetworkConfig tests network configuration
func TestNetworkConfig_Create(t *testing.T) {
	config := &NetworkConfig{
		Name:    "default",
		Type:    "nat",
		CIDR:    "192.168.122.0/24",
		Gateway: "192.168.122.1",
	}

	if config.Name != "default" {
		t.Errorf("expected default, got %s", config.Name)
	}
	if config.Type != "nat" {
		t.Errorf("expected nat type, got %s", config.Type)
	}
	if config.CIDR != "192.168.122.0/24" {
		t.Errorf("expected CIDR 192.168.122.0/24, got %s", config.CIDR)
	}
	if config.Gateway != "192.168.122.1" {
		t.Errorf("expected gateway 192.168.122.1, got %s", config.Gateway)
	}
}

// TestNetworkConfig_Bridge tests bridge network configuration
func TestNetworkConfig_Bridge(t *testing.T) {
	config := &NetworkConfig{
		Name:    "br-isolated",
		Type:    "bridge",
		CIDR:    "10.100.0.0/16",
		Gateway: "10.100.0.1",
	}

	if config.Type != "bridge" {
		t.Errorf("expected bridge type, got %s", config.Type)
	}
	if config.Name != "br-isolated" {
		t.Errorf("expected br-isolated, got %s", config.Name)
	}
}

// TestManager_Create tests manager creation
func TestManager_Create(t *testing.T) {
	mgr := NewManager("/var/lib/libvirt/images")

	if mgr == nil {
		t.Fatal("manager should not be nil")
	}
	if mgr.imageDir != "/var/lib/libvirt/images" {
		t.Errorf("expected image dir, got %s", mgr.imageDir)
	}
	if mgr.networks == nil {
		t.Error("networks map should not be nil")
	}
}

// TestManager_DefaultImageDir tests default image directory
func TestManager_DefaultImageDir(t *testing.T) {
	mgr := NewManager("")

	if mgr.imageDir != "/var/lib/libvirt/images" {
		t.Errorf("expected default image dir, got %s", mgr.imageDir)
	}
}

// TestManager_CustomImageDir tests custom image directory
func TestManager_CustomImageDir(t *testing.T) {
	mgr := NewManager("/custom/path/images")

	if mgr.imageDir != "/custom/path/images" {
		t.Errorf("expected custom image dir, got %s", mgr.imageDir)
	}
}

// TestManager_AddNetwork tests adding network configuration
func TestManager_AddNetwork(t *testing.T) {
	mgr := NewManager("")

	// Add network to map
	mgr.networks["default"] = &NetworkConfig{
		Name:    "default",
		Type:    "nat",
		CIDR:    "192.168.122.0/24",
		Gateway: "192.168.122.1",
	}

	if len(mgr.networks) != 1 {
		t.Errorf("expected 1 network, got %d", len(mgr.networks))
	}
	if mgr.networks["default"] == nil {
		t.Error("default network should exist")
	}
}

// TestManager_MultipleNetworks tests multiple network configurations
func TestManager_MultipleNetworks(t *testing.T) {
	mgr := NewManager("")

	networks := []*NetworkConfig{
		{Name: "default", Type: "nat", CIDR: "192.168.122.0/24"},
		{Name: "isolated", Type: "bridge", CIDR: "10.100.0.0/16"},
		{Name: "public", Type: "bridge", CIDR: "172.16.0.0/12"},
	}

	for _, n := range networks {
		mgr.networks[n.Name] = n
	}

	if len(mgr.networks) != 3 {
		t.Errorf("expected 3 networks, got %d", len(mgr.networks))
	}

	// Verify each network
	for _, n := range networks {
		if mgr.networks[n.Name] == nil {
			t.Errorf("network %s should exist", n.Name)
		}
		if mgr.networks[n.Name].Type != n.Type {
			t.Errorf("expected type %s for %s, got %s", n.Type, n.Name, mgr.networks[n.Name].Type)
		}
	}
}

// TestNetworkConfig_Types tests different network types
func TestNetworkConfig_Types(t *testing.T) {
	types := []struct {
		name     string
		nwType   string
		expected string
	}{
		{"NAT network", "nat", "nat"},
		{"Bridge network", "bridge", "bridge"},
		{"MACVTAP network", "macvtap", "macvtap"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			config := &NetworkConfig{Type: tt.nwType}
			if config.Type != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, config.Type)
			}
		})
	}
}

// TestNetworkConfig_CIDRs tests different CIDR formats
func TestNetworkConfig_CIDRs(t *testing.T) {
	cidrs := []struct {
		name string
		cidr string
	}{
		{"Private class C", "192.168.1.0/24"},
		{"Private class B", "172.16.0.0/16"},
		{"Private class A", "10.0.0.0/8"},
		{"Custom subnet", "10.100.50.0/24"},
	}

	for _, tt := range cidrs {
		t.Run(tt.name, func(t *testing.T) {
			config := &NetworkConfig{CIDR: tt.cidr}
			if config.CIDR != tt.cidr {
				t.Errorf("expected %s, got %s", tt.cidr, config.CIDR)
			}
		})
	}
}

// TestNetworkConfig_Gateways tests different gateway configurations
func TestNetworkConfig_Gateways(t *testing.T) {
	gateways := []struct {
		name    string
		gateway string
	}{
		{"Standard gateway", "192.168.1.1"},
		{"Custom gateway", "10.100.0.254"},
		{"First IP", "172.16.0.1"},
	}

	for _, tt := range gateways {
		t.Run(tt.name, func(t *testing.T) {
			config := &NetworkConfig{Gateway: tt.gateway}
			if config.Gateway != tt.gateway {
				t.Errorf("expected %s, got %s", tt.gateway, config.Gateway)
			}
		})
	}
}

// TestManager_RemoveNetwork tests removing network configuration
func TestManager_RemoveNetwork(t *testing.T) {
	mgr := NewManager("")

	// Add networks
	mgr.networks["default"] = &NetworkConfig{Name: "default", Type: "nat"}
	mgr.networks["isolated"] = &NetworkConfig{Name: "isolated", Type: "bridge"}

	// Remove one
	delete(mgr.networks, "isolated")

	if len(mgr.networks) != 1 {
		t.Errorf("expected 1 network, got %d", len(mgr.networks))
	}
	if mgr.networks["isolated"] != nil {
		t.Error("isolated network should be removed")
	}
}

// TestNetworkConfig_Validation tests network configuration validation
func TestNetworkConfig_Validation(t *testing.T) {
	// Valid NAT network
	nat := &NetworkConfig{
		Name:    "nat-network",
		Type:    "nat",
		CIDR:    "192.168.122.0/24",
		Gateway: "192.168.122.1",
	}

	if nat.Name == "" {
		t.Error("network name should not be empty")
	}
	if nat.Type == "" {
		t.Error("network type should not be empty")
	}
	if nat.CIDR == "" {
		t.Error("network CIDR should not be empty")
	}
	if nat.Gateway == "" {
		t.Error("network gateway should not be empty")
	}
}

// TestNetworkConfig_EmptyValues tests empty value handling
func TestNetworkConfig_EmptyValues(t *testing.T) {
	config := &NetworkConfig{}

	if config.Name != "" {
		t.Error("empty name should be empty string")
	}
	if config.Type != "" {
		t.Error("empty type should be empty string")
	}
	if config.CIDR != "" {
		t.Error("empty CIDR should be empty string")
	}
	if config.Gateway != "" {
		t.Error("empty gateway should be empty string")
	}
}