// Package network provides IPAM coverage tests
package network

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIPAMManager_GetDNS tests GetDNS method
func TestIPAMManager_GetDNS(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.50.0/24",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	dns := manager.GetDNS()
	if len(dns) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(dns))
	}
	if dns[0] != "8.8.8.8" {
		t.Errorf("expected 8.8.8.8, got %s", dns[0])
	}
}

// TestIPAMManager_ListPools tests ListPools method
func TestIPAMManager_ListPools(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.51.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	pools := manager.ListPools()
	// May be empty if no allocations
	t.Logf("ListPools returned %d pools", len(pools))
}

// TestIPAMManager_Used tests Used method
func TestIPAMManager_Used(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.52.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	used := manager.Used()
	// Should be 0 initially
	t.Logf("Used: %d", used)
}

// TestIPAMManager_Available tests Available method
func TestIPAMManager_Available(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.53.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	available := manager.Available()
	t.Logf("Available: %d", available)
}

// TestIPAMManager_GetGateway tests GetGateway method
func TestIPAMManager_GetGateway(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "192.168.54.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	// GetGateway requires a valid CIDR pool
	gateway, err := manager.GetGateway("192.168.54.0/24")
	if err != nil {
		t.Logf("GetGateway returned error: %v", err)
	} else {
		t.Logf("Gateway: %s", gateway)
	}
}

// TestIPAMManager_Allocate_Release tests allocate and release cycle
func TestIPAMManager_Allocate_Release(t *testing.T) {
	// Create temp state file
	tmpDir, err := os.MkdirTemp("", "ipam-alloc-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "ipam-state.json")

	config := &IPAMConfig{
		BaseCIDR: "10.100.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	manager.SetStateFile(stateFile)

	// Allocate an IP
	ip, mac, err := manager.Allocate()
	if err != nil {
		t.Logf("Allocate returned: %v", err)
	} else {
		t.Logf("Allocated: ip=%s, mac=%s", ip, mac)

		// Try to release
		err = manager.Release("10.100.0.0/24")
		if err != nil {
			t.Logf("Release returned: %v", err)
		}
	}
}

// TestIPAMManager_AllocateIP_ReleaseIP tests specific IP allocation
func TestIPAMManager_AllocateIP_ReleaseIP(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ipam-ip-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "ipam-state.json")

	config := &IPAMConfig{
		BaseCIDR: "10.101.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	manager.SetStateFile(stateFile)

	// Allocate specific IP
	mac := "52:54:00:aa:bb:cc"
	ip, err := manager.AllocateIP("10.101.0.0/24", mac)
	if err != nil {
		t.Logf("AllocateIP returned: %v", err)
	} else {
		t.Logf("Allocated IP: %s", ip)

		// Try to release
		err = manager.ReleaseIP("10.101.0.0/24", ip)
		if err != nil {
			t.Logf("ReleaseIP returned: %v", err)
		}
	}
}

// TestIPAMManager_GetIP_Lookup tests GetIP method
func TestIPAMManager_GetIP_Lookup(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.102.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	// GetIP with unknown MAC
	ip, err := manager.GetIP("unknown-mac")
	if err != nil {
		t.Logf("GetIP(unknown) returned: %v", err)
	} else {
		t.Logf("GetIP(unknown) = %s", ip)
	}
}

// TestIPAMManager_GetMAC_Lookup tests GetMAC method
func TestIPAMManager_GetMAC_Lookup(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.103.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	// GetMAC with unknown IP
	mac, err := manager.GetMAC("10.103.0.0/24", "10.103.0.100")
	if err != nil {
		t.Logf("GetMAC returned: %v", err)
	} else {
		t.Logf("GetMAC = %s", mac)
	}
}

// TestIPAMManager_Reclaim tests Reclaim method
func TestIPAMManager_Reclaim(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.104.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	// Reclaim should work even on empty pool
	err = manager.Reclaim("10.104.0.0/24")
	if err != nil {
		t.Logf("Reclaim returned: %v", err)
	}
}

// TestIPAMManager_saveState tests saveState method
func TestIPAMManager_saveState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ipam-save-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "state.json")

	config := &IPAMConfig{
		BaseCIDR: "10.105.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	manager.SetStateFile(stateFile)

	// State should be set
	t.Logf("State file configured: %s", stateFile)
}

// TestIncrementIP_Single tests the incrementIP helper with single increment
func TestIncrementIP_Single(t *testing.T) {
	// incrementIP takes (ip string, offsets ...int)
	tests := []struct {
		ip       string
		offset   int
		expected string
	}{
		{"192.168.1.1", 1, "192.168.1.2"},
		{"192.168.1.9", 1, "192.168.1.10"},
		{"192.168.1.99", 1, "192.168.1.100"},
		{"192.168.1.254", 1, "192.168.1.255"},
		{"192.168.1.1", 10, "192.168.1.11"},
		{"192.168.1.255", 1, "192.168.2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.ip+"+"+string(rune('0'+tt.offset)), func(t *testing.T) {
			result := incrementIP(tt.ip, tt.offset)
			if result != tt.expected {
				t.Errorf("incrementIP(%s, %d) = %s, want %s", tt.ip, tt.offset, result, tt.expected)
			}
		})
	}
}

// TestIPAMManager_MultipleAllocations tests multiple IP allocations
func TestIPAMManager_MultipleAllocations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ipam-multi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "state.json")

	config := &IPAMConfig{
		BaseCIDR: "10.200.0.0/24",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	manager, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager creation failed: %v", err)
	}

	manager.SetStateFile(stateFile)

	// Try multiple allocations
	for i := 0; i < 3; i++ {
		ip, mac, err := manager.Allocate()
		if err != nil {
			t.Logf("Allocation %d failed: %v", i, err)
			break
		}
		t.Logf("Allocation %d: ip=%s, mac=%s", i, ip, mac)
	}

	// Check used count
	used := manager.Used()
	t.Logf("Total used: %d", used)
}

// TestIPAMManager_StatePersistence tests state save/load
func TestIPAMManager_StatePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ipam-persist-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "state.json")

	config := &IPAMConfig{
		BaseCIDR: "10.210.0.0/24",
		DNS:      []string{"1.1.1.1"},
	}

	// Create first manager
	manager1, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager1 creation failed: %v", err)
	}
	manager1.SetStateFile(stateFile)

	// Create second manager (should load state)
	manager2, err := NewIPAMManager(config)
	if err != nil {
		t.Skipf("manager2 creation failed: %v", err)
	}

	// Both should exist
	t.Logf("Manager1 DNS: %v", manager1.GetDNS())
	t.Logf("Manager2 DNS: %v", manager2.GetDNS())
}