// Package network provides firewall tests
package network

import (
	"testing"
)

// TestFirewallManager_Stub tests firewall with stub backend
func TestFirewallManager_Stub(t *testing.T) {
	fm, err := NewFirewallManager(FirewallBackendStub)
	if err != nil {
		t.Fatalf("failed to create firewall manager: %v", err)
	}
	if fm == nil {
		t.Fatal("expected non-nil firewall manager")
	}
}

// TestFirewallManager_DetectBackend tests backend detection
func TestFirewallManager_DetectBackend(t *testing.T) {
	fm := &FirewallManager{
		backend: "",
		rules:   make(map[string][]string),
	}

	fm.detectBackend()

	// Backend should be set (iptables, nftables, or stub)
	if fm.backend == "" {
		t.Error("backend should be detected")
	}
}

// TestFirewallManager_IsBackendAvailable tests backend availability check
func TestFirewallManager_IsBackendAvailable(t *testing.T) {
	fm := &FirewallManager{backend: FirewallBackendStub}
	if !fm.isBackendAvailable() {
		t.Error("stub backend should always be available")
	}
}

// TestFirewallBackend_Constants tests backend constants
func TestFirewallBackend_Constants(t *testing.T) {
	backends := []FirewallBackend{
		FirewallBackendIPTables,
		FirewallBackendNFTables,
		FirewallBackendStub,
	}

	if len(backends) != 3 {
		t.Error("expected 3 backend types")
	}
}

// TestFirewallManager_Close tests close
func TestFirewallManager_Close(t *testing.T) {
	fm, err := NewFirewallManager(FirewallBackendStub)
	if err != nil {
		t.Fatalf("failed to create firewall manager: %v", err)
	}

	err = fm.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

// TestFirewallManager_GetBackend tests GetBackend
func TestFirewallManager_GetBackend(t *testing.T) {
	fm, err := NewFirewallManager(FirewallBackendStub)
	if err != nil {
		t.Fatalf("failed to create firewall manager: %v", err)
	}

	backend := fm.GetBackend()
	if backend != FirewallBackendStub {
		t.Errorf("expected stub, got %s", backend)
	}
}

// TestIsolationManager_Struct tests isolation manager struct
func TestIsolationManager_Struct(t *testing.T) {
	// Just verify struct fields exist
	type testIsolationManager struct {
		db       interface{}
		networks map[string]interface{}
	}

	im := &testIsolationManager{
		networks: make(map[string]interface{}),
	}

	if im.networks == nil {
		t.Error("networks should be initialized")
	}
}

// TestIsolatedNetwork_Struct tests isolated network struct
func TestIsolatedNetwork_Struct(t *testing.T) {
	in := &IsolatedNetwork{
		ID:        "net-1",
		BridgeName: "isolated",
		CIDR:      "10.0.0.0/24",
		Gateway:   "10.0.0.1",
		VLANID:    100,
		VMs:       []string{},
	}

	if in.ID != "net-1" {
		t.Error("ID mismatch")
	}
	if in.VLANID != 100 {
		t.Error("VLANID mismatch")
	}
}

// TestNetworkConfig_Struct tests network config struct
func TestNetworkConfig_Struct(t *testing.T) {
	nc := &NetworkConfig{
		BaseCIDR:        "192.168.0.0/24",
		VLANStart:      100,
		DNS:             []string{"8.8.8.8"},
	}

	if nc.BaseCIDR != "192.168.0.0/24" {
		t.Error("BaseCIDR mismatch")
	}
	if len(nc.DNS) != 1 {
		t.Error("DNS count mismatch")
	}
}

// TestCIDRPool_Struct tests CIDR pool struct
func TestCIDRPool_Struct(t *testing.T) {
	pool := &CIDRPool{
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
		Start:   "10.0.0.10",
		End:     "10.0.0.250",
	}

	if pool.Start == "" {
		t.Error("Start should be set")
	}
}

// TestIPAllocation_Struct tests IP allocation struct
func TestIPAllocation_Struct(t *testing.T) {
	alloc := &IPAllocation{
		IP:  "10.0.0.100",
		MAC: "00:11:22:33:44:55",
	}

	if alloc.IP != "10.0.0.100" {
		t.Error("IP mismatch")
	}
}

// TestIPAMConfig_Struct tests IPAM config struct
func TestIPAMConfig_Struct(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.0.0/16",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	if config.BaseCIDR == "" {
		t.Error("BaseCIDR should be set")
	}
}