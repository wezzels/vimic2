// Package network provides network management tests
package network

import (
	"testing"
)

// TestRouterInterfaceFields tests RouterInterface fields
func TestRouterInterface_Fields(t *testing.T) {
	iface := RouterInterface{
		ID:        "ri-1",
		Name:      "eth0",
		NetworkID: "net-1",
	}

	if iface.ID != "ri-1" {
		t.Errorf("expected ri-1, got %s", iface.ID)
	}
	if iface.Name != "eth0" {
		t.Errorf("expected eth0, got %s", iface.Name)
	}
}

// TestNATRuleFields tests NATRule fields
func TestNATRule_Fields(t *testing.T) {
	nat := NATRule{
		ID:           "nat-1",
		ExternalIP:   "1.2.3.4",
		ExternalPort: 8080,
		InternalIP:   "10.0.0.1",
		InternalPort: 80,
		Protocol:     "tcp",
	}

	if nat.ID != "nat-1" {
		t.Errorf("expected nat-1, got %s", nat.ID)
	}
	if nat.Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", nat.Protocol)
	}
}

// TestRouteFields tests Route fields
func TestRoute_Fields(t *testing.T) {
	route := Route{
		ID:          "route-1",
		Destination: "192.168.2.0/24",
		Gateway:     "192.168.1.1",
		Metric:      100,
	}

	if route.ID != "route-1" {
		t.Errorf("expected route-1, got %s", route.ID)
	}
	if route.Metric != 100 {
		t.Errorf("expected 100, got %d", route.Metric)
	}
}

// TestFirewallRuleFields tests FirewallRule fields
func TestFirewallRule_Fields(t *testing.T) {
	rule := FirewallRule{
		ID:         "fw-1",
		Name:       "Allow SSH",
		Direction:  "ingress",
		Protocol:   "tcp",
		SourceCIDR: "10.0.0.0/8",
		DestPort:   22,
		Action:     "accept",
		Priority:   100,
		Enabled:    true,
	}

	if rule.ID != "fw-1" {
		t.Errorf("expected fw-1, got %s", rule.ID)
	}
	if rule.DestPort != 22 {
		t.Errorf("expected 22, got %d", rule.DestPort)
	}
	if !rule.Enabled {
		t.Error("expected enabled")
	}
}

// TestMultipleFirewallRules tests multiple rules
func TestMultipleFirewallRules(t *testing.T) {
	rules := []FirewallRule{
		{ID: "fw-1", Action: "accept", DestPort: 22},
		{ID: "fw-2", Action: "accept", DestPort: 443},
		{ID: "fw-3", Action: "drop", DestPort: 23},
	}

	if len(rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rules))
	}

	acceptCount := 0
	for _, r := range rules {
		if r.Action == "accept" {
			acceptCount++
		}
	}
	if acceptCount != 2 {
		t.Errorf("expected 2 accept rules, got %d", acceptCount)
	}
}

// TestTunnelProtocol tests tunnel protocols
func TestTunnelProtocol(t *testing.T) {
	protocols := []TunnelProtocol{"vxlan", "gre", "ipip"}

	for _, proto := range protocols {
		t.Run(string(proto), func(t *testing.T) {
			tunnel := Tunnel{Protocol: proto}
			if tunnel.Protocol != proto {
				t.Errorf("expected %s, got %s", proto, tunnel.Protocol)
			}
		})
	}
}
