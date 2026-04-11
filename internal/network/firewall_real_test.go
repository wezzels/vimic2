//go:build integration

// Package network provides integration tests for firewall
package network

import (
	"os"
	"testing"
)

// TestIntegration_Firewall_CreateIsolationRules tests creating isolation rules
func TestIntegration_Firewall_CreateIsolationRules(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for firewall operations")
	}

	// Try iptables first, then nftables
	fm, err := NewFirewallManager(FirewallBackendIPTables)
	if err != nil {
		// Try nftables
		fm, err = NewFirewallManager(FirewallBackendNFTables)
		if err != nil {
			t.Skipf("no firewall backend available: %v", err)
		}
	}

	bridgeName := "test-br-isolated"
	cidr := "10.50.0.0/24"
	vlanID := 500

	// Create isolation rules
	err = fm.CreateIsolationRules(bridgeName, cidr, vlanID)
	if err != nil {
		t.Fatalf("CreateIsolationRules failed: %v", err)
	}

	t.Logf("Created isolation rules for %s (VLAN %d)", bridgeName, vlanID)

	// List rules
	rules := fm.ListRules()
	for chain, chainRules := range rules {
		t.Logf("Chain %s: %d rules", chain, len(chainRules))
	}

	// Clean up
	err = fm.DeleteIsolationRules(bridgeName, cidr, vlanID)
	if err != nil {
		t.Logf("DeleteIsolationRules failed (cleanup): %v", err)
	}
}

// TestIntegration_Firewall_AllowDenyTraffic tests allow/deny rules
func TestIntegration_Firewall_AllowDenyTraffic(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for firewall operations")
	}

	fm, err := NewFirewallManager("")
	if err != nil {
		t.Skipf("no firewall backend available: %v", err)
	}

	sourceCIDR := "10.60.0.0/24"
	destCIDR := "10.70.0.0/24"
	ports := []int{80, 443}

	// Allow traffic
	err = fm.AllowTraffic(sourceCIDR, destCIDR, ports)
	if err != nil {
		t.Fatalf("AllowTraffic failed: %v", err)
	}

	t.Logf("Allowed traffic from %s to %s on ports %v", sourceCIDR, destCIDR, ports)

	// Deny traffic
	err = fm.DenyTraffic(sourceCIDR, destCIDR)
	if err != nil {
		t.Logf("DenyTraffic failed: %v", err)
	}
}

// TestIntegration_Firewall_BackendDetection tests backend detection
func TestIntegration_Firewall_BackendDetection(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for firewall operations")
	}

	// Auto-detect
	fm, err := NewFirewallManager("")
	if err != nil {
		t.Skipf("no firewall backend available: %v", err)
	}

	backend := fm.GetBackend()
	t.Logf("Detected firewall backend: %s", backend)

	if backend == FirewallBackendStub {
		t.Log("Using stub backend (no real firewall)")
	}
}

// TestIntegration_Firewall_IPTables tests iptables specifically
func TestIntegration_Firewall_IPTables(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for firewall operations")
	}

	fm, err := NewFirewallManager(FirewallBackendIPTables)
	if err != nil {
		t.Skipf("iptables not available: %v", err)
	}

	bridgeName := "test-br-iptables"
	cidr := "10.80.0.0/24"
	vlanID := 800

	err = fm.CreateIsolationRules(bridgeName, cidr, vlanID)
	if err != nil {
		t.Fatalf("CreateIsolationRules failed: %v", err)
	}

	// Verify backend
	if fm.GetBackend() != FirewallBackendIPTables {
		t.Errorf("expected iptables backend, got %s", fm.GetBackend())
	}

	// Clean up
	fm.DeleteIsolationRules(bridgeName, cidr, vlanID)
}

// TestIntegration_Firewall_NFTables tests nftables specifically
func TestIntegration_Firewall_NFTables(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for firewall operations")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Skipf("nftables not available: %v", err)
	}

	bridgeName := "test-br-nftables"
	cidr := "10.90.0.0/24"
	vlanID := 900

	err = fm.CreateIsolationRules(bridgeName, cidr, vlanID)
	if err != nil {
		t.Fatalf("CreateIsolationRules failed: %v", err)
	}

	// Verify backend
	if fm.GetBackend() != FirewallBackendNFTables {
		t.Errorf("expected nftables backend, got %s", fm.GetBackend())
	}

	// Clean up
	fm.DeleteIsolationRules(bridgeName, cidr, vlanID)
}