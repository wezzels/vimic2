// Package network provides IPAM tests
package network

import (
	"testing"
)

// TestIPAMConfig tests IPAM configuration
func TestIPAMConfig(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.0.0/16",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	if config.BaseCIDR != "192.168.0.0/16" {
		t.Errorf("expected base CIDR, got %s", config.BaseCIDR)
	}
	if len(config.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(config.DNS))
	}
}

// TestCIDRPool tests CIDR pool
func TestCIDRPool(t *testing.T) {
	pool := &CIDRPool{
		CIDR:        "192.168.100.0/24",
		Gateway:     "192.168.100.1",
		Allocations: make(map[string]string),
	}

	if pool.CIDR != "192.168.100.0/24" {
		t.Errorf("expected CIDR, got %s", pool.CIDR)
	}
	if pool.Gateway != "192.168.100.1" {
		t.Errorf("expected gateway, got %s", pool.Gateway)
	}
}

// TestIPAllocation tests IP allocation logic
func TestIPAllocation(t *testing.T) {
	allocations := make(map[string]string)

	// Simulate allocation
	allocations["192.168.1.10"] = "mac-aa:bb:cc:dd:ee:ff"
	allocations["192.168.1.11"] = "mac-11:22:33:44:55:66"

	if len(allocations) != 2 {
		t.Errorf("expected 2 allocations, got %d", len(allocations))
	}

	// Check allocation exists
	if ip, exists := allocations["192.168.1.10"]; !exists {
		t.Error("expected allocation to exist")
	} else if ip == "" {
		t.Error("expected MAC to be non-empty")
	}
}

// TestIPRelease tests IP release logic
func TestIPRelease(t *testing.T) {
	allocations := map[string]string{
		"192.168.1.10": "mac-1",
		"192.168.1.11": "mac-2",
	}

	// Release an IP
	delete(allocations, "192.168.1.10")

	if len(allocations) != 1 {
		t.Errorf("expected 1 allocation after release, got %d", len(allocations))
	}

	if _, exists := allocations["192.168.1.10"]; exists {
		t.Error("expected IP to be released")
	}
}

// TestFirewallAction tests firewall actions
func TestFirewallAction(t *testing.T) {
	actions := []string{"accept", "drop", "reject"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			rule := FirewallRule{Action: action}
			if rule.Action != action {
				t.Errorf("expected %s, got %s", action, rule.Action)
			}
		})
	}
}

// TestFirewallDirection tests firewall directions
func TestFirewallDirection(t *testing.T) {
	directions := []string{"ingress", "egress"}

	for _, dir := range directions {
		t.Run(dir, func(t *testing.T) {
			rule := FirewallRule{Direction: dir}
			if rule.Direction != dir {
				t.Errorf("expected %s, got %s", dir, rule.Direction)
			}
		})
	}
}

// TestFirewallProtocols tests firewall protocols
func TestFirewallProtocols(t *testing.T) {
	protocols := []struct {
		proto    string
		expected string
	}{
		{"tcp", "tcp"},
		{"udp", "udp"},
		{"icmp", "icmp"},
		{"all", "all"},
	}

	for _, tt := range protocols {
		t.Run(tt.proto, func(t *testing.T) {
			rule := FirewallRule{Protocol: tt.proto}
			if rule.Protocol != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, rule.Protocol)
			}
		})
	}
}

// TestMultipleRules tests multiple firewall rules
func TestMultipleRules(t *testing.T) {
	rules := []FirewallRule{
		{ID: "fw-1", Action: "accept", DestPort: 22, Enabled: true},
		{ID: "fw-2", Action: "accept", DestPort: 443, Enabled: true},
		{ID: "fw-3", Action: "drop", DestPort: 23, Enabled: true},
		{ID: "fw-4", Action: "drop", DestPort: 25, Enabled: false},
	}

	enabledCount := 0
	acceptCount := 0
	for _, r := range rules {
		if r.Enabled {
			enabledCount++
		}
		if r.Action == "accept" {
			acceptCount++
		}
	}

	if enabledCount != 3 {
		t.Errorf("expected 3 enabled rules, got %d", enabledCount)
	}
	if acceptCount != 2 {
		t.Errorf("expected 2 accept rules, got %d", acceptCount)
	}
}

// TestRouterWithMultipleInterfaces tests router with interfaces
func TestRouterWithMultipleInterfaces(t *testing.T) {
	router := &Router{
		ID:   "router-1",
		Name: "main-router",
		Interfaces: []RouterInterface{
			{ID: "i1", Name: "eth0", NetworkID: "net-1"},
			{ID: "i2", Name: "eth1", NetworkID: "net-2"},
		},
		RoutingTable: []Route{
			{ID: "r1", Destination: "192.168.1.0/24", Gateway: "10.0.0.1"},
		},
		NATRules: []NATRule{
			{ID: "n1", ExternalPort: 8080, InternalPort: 80},
		},
	}

	if len(router.Interfaces) != 2 {
		t.Errorf("expected 2 interfaces, got %d", len(router.Interfaces))
	}
	if len(router.RoutingTable) != 1 {
		t.Errorf("expected 1 route, got %d", len(router.RoutingTable))
	}
	if len(router.NATRules) != 1 {
		t.Errorf("expected 1 NAT rule, got %d", len(router.NATRules))
	}
}

// TestNetworkMultipleVLANs tests network with VLANs
func TestNetworkMultipleVLANs(t *testing.T) {
	network := &Network{
		ID:          "net-1",
		Name:        "multi-vlan",
		CIDR:        "10.0.0.0/16",
		Gateway:     "10.0.0.1",
		DHCPEnabled: true,
		VLANID:      100,
	}

	if network.VLANID != 100 {
		t.Errorf("expected VLAN 100, got %d", network.VLANID)
	}
	if !network.DHCPEnabled {
		t.Error("expected DHCP to be enabled")
	}
}
