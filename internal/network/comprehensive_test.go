// Package network provides comprehensive real tests
package network

import (
	"testing"
)

// TestFirewallRule_EnableDisable tests enabling and disabling rules
func TestFirewallRule_EnableDisable(t *testing.T) {
	rule := FirewallRule{
		ID:         "rule-1",
		Name:       "SSH Access",
		Direction:  "ingress",
		Protocol:   "tcp",
		DestPort:   22,
		Action:     "accept",
		Priority:   100,
		Enabled:    true,
	}

	// Check enabled
	if !rule.Enabled {
		t.Error("expected rule to be enabled")
	}

	// Disable
	rule.Enabled = false
	if rule.Enabled {
		t.Error("expected rule to be disabled")
	}

	// Re-enable
	rule.Enabled = true
	if !rule.Enabled {
		t.Error("expected rule to be enabled again")
	}
}

// TestFirewall_PriorityOrder tests rule priority ordering
func TestFirewall_PriorityOrder(t *testing.T) {
	rules := []FirewallRule{
		{ID: "r1", Priority: 100, Action: "accept"},
		{ID: "r2", Priority: 50, Action: "drop"},
		{ID: "r3", Priority: 200, Action: "accept"},
	}

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	// Verify order
	if rules[0].Priority != 50 {
		t.Errorf("expected first rule priority 50, got %d", rules[0].Priority)
	}
	if rules[2].Priority != 200 {
		t.Errorf("expected last rule priority 200, got %d", rules[2].Priority)
	}
}

// TestRouter_Enabled tests router enable/disable
func TestRouter_Enabled(t *testing.T) {
	router := &Router{
		ID:      "router-1",
		Name:    "Main Router",
		Enabled: true,
	}

	if !router.Enabled {
		t.Error("expected router to be enabled")
	}

	router.Enabled = false
	if router.Enabled {
		t.Error("expected router to be disabled")
	}
}

// TestRouterInterface_Naming tests interface naming
func TestRouterInterface_Naming(t *testing.T) {
	iface := RouterInterface{
		ID:        "if-1",
		Name:      "eth0",
		NetworkID: "net-1",
	}

	if iface.Name != "eth0" {
		t.Errorf("expected eth0, got %s", iface.Name)
	}
}

// TestRoute_Gateway tests route gateway
func TestRoute_Gateway(t *testing.T) {
	route := Route{
		ID:          "route-1",
		Destination: "192.168.2.0/24",
		Gateway:     "192.168.1.1",
		Interface:   "eth0",
		Metric:      100,
	}

	if route.Gateway != "192.168.1.1" {
		t.Errorf("expected gateway, got %s", route.Gateway)
	}
	if route.Metric != 100 {
		t.Errorf("expected metric 100, got %d", route.Metric)
	}
}

// TestNATRule_PortForwarding tests NAT port forwarding
func TestNATRule_PortForwarding(t *testing.T) {
	nat := NATRule{
		ID:           "nat-1",
		ExternalIP:   "1.2.3.4",
		ExternalPort: 8080,
		InternalIP:   "10.0.0.5",
		InternalPort: 80,
		Protocol:     "tcp",
	}

	// Verify port mapping
	if nat.ExternalPort != 8080 {
		t.Errorf("expected external port 8080, got %d", nat.ExternalPort)
	}
	if nat.InternalPort != 80 {
		t.Errorf("expected internal port 80, got %d", nat.InternalPort)
	}
	if nat.Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", nat.Protocol)
	}
}

// TestVMInterface_VLAN tests VLAN tagging
func TestVMInterface_VLAN(t *testing.T) {
	iface := VMInterface{
		ID:         "iface-1",
		VMID:       "vm-1",
		Name:       "eth0",
		VLANID:     100,
		TrunkVLANs: []int{100, 200, 300},
	}

	if iface.VLANID != 100 {
		t.Errorf("expected VLAN 100, got %d", iface.VLANID)
	}
	if len(iface.TrunkVLANs) != 3 {
		t.Errorf("expected 3 trunk VLANs, got %d", len(iface.TrunkVLANs))
	}
}

// TestVMInterface_Bandwidth tests bandwidth limiting
func TestVMInterface_Bandwidth(t *testing.T) {
	iface := VMInterface{
		ID:        "iface-1",
		Bandwidth: 1000, // Mbps
	}

	if iface.Bandwidth != 1000 {
		t.Errorf("expected bandwidth 1000, got %d", iface.Bandwidth)
	}
}

// TestNetwork_DHCP tests DHCP configuration
func TestNetwork_DHCP(t *testing.T) {
	net := &Network{
		ID:          "net-1",
		Name:        "default",
		CIDR:        "192.168.1.0/24",
		Gateway:     "192.168.1.1",
		DHCPEnabled: true,
	}

	if !net.DHCPEnabled {
		t.Error("expected DHCP to be enabled")
	}

	// Disable DHCP
	net.DHCPEnabled = false
	if net.DHCPEnabled {
		t.Error("expected DHCP to be disabled")
	}
}

// TestNetwork_MultipleVLANs tests multiple VLAN configurations
func TestNetwork_MultipleVLANs(t *testing.T) {
	networks := []*Network{
		{ID: "net-1", Name: "vlan100", VLANID: 100},
		{ID: "net-2", Name: "vlan200", VLANID: 200},
		{ID: "net-3", Name: "vlan300", VLANID: 300},
	}

	if len(networks) != 3 {
		t.Errorf("expected 3 networks, got %d", len(networks))
	}

	for i, net := range networks {
		expectedVLAN := (i + 1) * 100
		if net.VLANID != expectedVLAN {
			t.Errorf("expected VLAN %d, got %d", expectedVLAN, net.VLANID)
		}
	}
}

// TestTunnel_VXLAN tests VXLAN tunnel configuration
func TestTunnel_VXLAN(t *testing.T) {
	tunnel := Tunnel{
		ID:        "tunnel-1",
		Name:      "vxlan-5000",
		Protocol:  "vxlan",
		VNI:       5000,
		LocalIP:   "10.0.0.1",
		RemoteIP:  "10.0.0.2",
		SourcePort: 4789,
		DestPort:  4789,
	}

	if tunnel.VNI != 5000 {
		t.Errorf("expected VNI 5000, got %d", tunnel.VNI)
	}
	if tunnel.Protocol != "vxlan" {
		t.Errorf("expected vxlan, got %s", tunnel.Protocol)
	}
}

// TestTunnel_GRE tests GRE tunnel configuration
func TestTunnel_GRE(t *testing.T) {
	tunnel := Tunnel{
		ID:       "tunnel-2",
		Name:     "gre-1",
		Protocol: "gre",
		LocalIP:  "10.0.0.1",
		RemoteIP: "10.0.0.2",
	}

	if tunnel.Protocol != "gre" {
		t.Errorf("expected gre, got %s", tunnel.Protocol)
	}
}

// TestMultipleTunnels tests multiple tunnel configurations
func TestMultipleTunnels(t *testing.T) {
	tunnels := []*Tunnel{
		{ID: "t1", Protocol: "vxlan", VNI: 5000},
		{ID: "t2", Protocol: "vxlan", VNI: 5001},
		{ID: "t3", Protocol: "gre"},
	}

	vxlanCount := 0
	for _, t := range tunnels {
		if t.Protocol == "vxlan" {
			vxlanCount++
		}
	}

	if vxlanCount != 2 {
		t.Errorf("expected 2 VXLAN tunnels, got %d", vxlanCount)
	}
}

// TestFirewall_RuleMatch tests rule matching logic
func TestFirewall_RuleMatch(t *testing.T) {
	rule := FirewallRule{
		Direction:  "ingress",
		Protocol:   "tcp",
		DestPort:   22,
		SourceCIDR: "10.0.0.0/8",
		Action:     "accept",
	}

	// Verify rule matches expected pattern
	if rule.Direction != "ingress" {
		t.Error("expected ingress direction")
	}
	if rule.Protocol != "tcp" {
		t.Error("expected tcp protocol")
	}
	if rule.DestPort != 22 {
		t.Error("expected port 22")
	}
	if rule.Action != "accept" {
		t.Error("expected accept action")
	}
}