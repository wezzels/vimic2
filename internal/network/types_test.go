// Package network provides unit tests for network types
package network

import (
	"encoding/json"
	"testing"
	"time"
)

// TestOVSClient_Creation tests OVS client creation
func TestOVSClient_Creation(t *testing.T) {
	ovs := NewOVSClient()

	if ovs == nil {
		t.Fatal("NewOVSClient returned nil")
	}
	if ovs.vswitchdPath != "ovs-vsctl" {
		t.Errorf("expected vswitchdPath 'ovs-vsctl', got %s", ovs.vswitchdPath)
	}
	if ovs.ofctlPath != "ovs-ofctl" {
		t.Errorf("expected ofctlPath 'ovs-ofctl', got %s", ovs.ofctlPath)
	}
}

// TestIsolatedNetwork tests isolated network structure
func TestIsolatedNetwork_Create(t *testing.T) {
	network := &IsolatedNetwork{
		ID:         "network-1",
		PipelineID: "pipeline-1",
		BridgeName: "br-pipeline-1",
		VLANID:     100,
		CIDR:       "10.100.1.0/24",
		Gateway:    "10.100.1.1",
		DNS:        []string{"8.8.8.8", "8.8.4.4"},
		VMs:        []string{"vm-1", "vm-2"},
		CreatedAt:  time.Now(),
	}

	if network.ID != "network-1" {
		t.Errorf("expected ID network-1, got %s", network.ID)
	}
	if network.VLANID != 100 {
		t.Errorf("expected VLAN 100, got %d", network.VLANID)
	}
	if len(network.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(network.DNS))
	}
	if len(network.VMs) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(network.VMs))
	}
}

// TestNetworkConfig tests network configuration
func TestNetworkConfig_Create(t *testing.T) {
	config := &NetworkConfig{
		VLANStart:       100,
		VLANEnd:         200,
		BaseCIDR:        "10.0.0.0/8",
		DNS:             []string{"8.8.8.8"},
		OVSBridge:       "br-vimic2",
		FirewallBackend: "iptables",
	}

	if config.VLANStart != 100 {
		t.Errorf("expected VLAN start 100, got %d", config.VLANStart)
	}
	if config.VLANEnd != 200 {
		t.Errorf("expected VLAN end 200, got %d", config.VLANEnd)
	}
	if config.OVSBridge != "br-vimic2" {
		t.Errorf("expected bridge br-vimic2, got %s", config.OVSBridge)
	}
}

// TestVLANAllocator_New tests VLAN allocator creation
func TestVLANAllocator_New(t *testing.T) {
	alloc, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatalf("failed to create VLAN allocator: %v", err)
	}

	if alloc == nil {
		t.Fatal("VLAN allocator is nil")
	}
}

// TestVLANAllocator_InvalidRange tests invalid VLAN range
func TestVLANAllocator_InvalidRange(t *testing.T) {
	_, err := NewVLANAllocator(200, 100) // Start > End
	if err == nil {
		t.Error("expected error for invalid VLAN range")
	}
}

// TestIPAMConfig tests IPAM configuration
func TestIPAMConfig_Create(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.100.0.0/16",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	if config.BaseCIDR != "10.100.0.0/16" {
		t.Errorf("expected CIDR 10.100.0.0/16, got %s", config.BaseCIDR)
	}
	if len(config.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(config.DNS))
	}
}

// TestFirewallRule tests firewall rule structure
func TestFirewallRule_Create(t *testing.T) {
	rule := &FirewallRule{
		ID:          "rule-1",
		Name:        "allow-https",
		Direction:   "ingress",
		Protocol:    "tcp",
		DestPort:    443,
		SourceCIDR:  "0.0.0.0/0",
		Action:      "accept",
		Priority:    100,
		Enabled:     true,
	}

	if rule.ID != "rule-1" {
		t.Errorf("expected rule-1, got %s", rule.ID)
	}
	if rule.Direction != "ingress" {
		t.Errorf("expected ingress direction, got %s", rule.Direction)
	}
	if rule.DestPort != 443 {
		t.Errorf("expected port 443, got %d", rule.DestPort)
	}
}

// TestIsolatedNetwork_JSON tests JSON marshaling
func TestIsolatedNetwork_JSON(t *testing.T) {
	network := &IsolatedNetwork{
		ID:         "network-1",
		PipelineID: "pipeline-1",
		BridgeName: "br-pipeline-1",
		VLANID:     100,
		CIDR:       "10.100.1.0/24",
		Gateway:    "10.100.1.1",
		DNS:        []string{"8.8.8.8"},
		VMs:        []string{"vm-1"},
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("failed to marshal network: %v", err)
	}

	var network2 IsolatedNetwork
	if err := json.Unmarshal(data, &network2); err != nil {
		t.Fatalf("failed to unmarshal network: %v", err)
	}

	if network2.ID != network.ID {
		t.Errorf("expected ID %s, got %s", network.ID, network2.ID)
	}
	if network2.VLANID != network.VLANID {
		t.Errorf("expected VLAN %d, got %d", network.VLANID, network2.VLANID)
	}
}

// TestFirewallRule_JSON tests firewall rule JSON
func TestFirewallRule_JSON(t *testing.T) {
	rule := &FirewallRule{
		ID:          "rule-1",
		Name:        "allow-https",
		Direction:   "ingress",
		Protocol:    "tcp",
		DestPort:    443,
		SourceCIDR:  "0.0.0.0/0",
		Action:      "accept",
		Priority:    100,
		Enabled:     true,
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("failed to marshal rule: %v", err)
	}

	var rule2 FirewallRule
	if err := json.Unmarshal(data, &rule2); err != nil {
		t.Fatalf("failed to unmarshal rule: %v", err)
	}

	if rule2.ID != rule.ID {
		t.Errorf("expected ID %s, got %s", rule.ID, rule2.ID)
	}
	if rule2.DestPort != rule.DestPort {
		t.Errorf("expected port %d, got %d", rule.DestPort, rule2.DestPort)
	}
}

// TestNetworkTypes tests various network types
func TestNetworkTypes_VLAN(t *testing.T) {
	// Valid VLAN range
	validVLANs := []int{1, 100, 500, 1000, 4094}
	for _, vlan := range validVLANs {
		if vlan < 1 || vlan > 4094 {
			t.Errorf("VLAN %d outside valid range 1-4094", vlan)
		}
	}

	// Invalid VLAN range
	invalidVLANs := []int{0, -1, 4095, 5000}
	for _, vlan := range invalidVLANs {
		if vlan >= 1 && vlan <= 4094 {
			t.Errorf("VLAN %d should be invalid", vlan)
		}
	}
}

// TestTunnel tests tunnel structure
func TestTunnel_Create(t *testing.T) {
	tunnel := &Tunnel{
		ID:        "tunnel-1",
		Name:      "vxlan-to-remote",
		Protocol:  TunnelVXLAN,
		LocalIP:   "10.0.0.1",
		RemoteIP:  "10.0.0.2",
		VNI:       100,
		DestPort:  4789,
		NetworkID: "network-1",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	if tunnel.ID != "tunnel-1" {
		t.Errorf("expected tunnel-1, got %s", tunnel.ID)
	}
	if tunnel.Protocol != TunnelVXLAN {
		t.Errorf("expected VXLAN protocol, got %s", tunnel.Protocol)
	}
	if tunnel.VNI != 100 {
		t.Errorf("expected VNI 100, got %d", tunnel.VNI)
	}
}

// TestVMInterface tests VM interface structure
func TestVMInterface_Create(t *testing.T) {
	iface := &VMInterface{
		ID:           "iface-1",
		VMID:         "vm-1",
		Name:         "eth0",
		MACAddress:   "52:54:00:12:34:56",
		IPAddress:    "10.100.1.10",
		NetworkID:    "network-1",
		VLANID:       100,
		MTU:          1500,
		Bandwidth:    1000,
		State:        InterfaceUp,
		PortSecurity: true,
	}

	if iface.ID != "iface-1" {
		t.Errorf("expected iface-1, got %s", iface.ID)
	}
	if iface.Name != "eth0" {
		t.Errorf("expected eth0, got %s", iface.Name)
	}
	if iface.State != InterfaceUp {
		t.Errorf("expected InterfaceUp, got %s", iface.State)
	}
}

// TestTunnelProtocol tests tunnel protocol constants
func TestTunnelProtocol_Constants(t *testing.T) {
	protocols := []TunnelProtocol{TunnelVXLAN, TunnelGRE, TunnelGeneve}

	for _, p := range protocols {
		if p == "" {
			t.Error("empty tunnel protocol")
		}
	}
}

// TestInterfaceState tests interface state constants
func TestInterfaceState_Constants(t *testing.T) {
	states := []InterfaceState{InterfaceUp, InterfaceDown}

	for _, s := range states {
		if s == "" {
			t.Error("empty interface state")
		}
	}
}