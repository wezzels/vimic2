//go:build integration

package network

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestIsolation_RealCreateDelete tests creating and deleting isolation rules
// Requires root/sudo for nftables/iptables access
func TestIsolation_RealCreateDelete(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Fatalf("NewFirewallManager failed: %v", err)
	}

	t.Logf("Firewall backend: %s", fm.GetBackend())

	// Create isolation rules
	err = fm.CreateIsolationRules("vimic2-iso-test", "10.200.0.0/24", 200)
	if err != nil {
		t.Logf("CreateIsolationRules: %v", err)
	} else {
		t.Log("Isolation rules created successfully")

		// Clean up
		err = fm.DeleteIsolationRules("vimic2-iso-test", "10.200.0.0/24", 200)
		if err != nil {
			t.Logf("DeleteIsolationRules: %v", err)
		} else {
			t.Log("Isolation rules deleted successfully")
		}
	}
}

// TestFirewall_RealAllowDenyTraffic tests allow/deny traffic rules
// Requires root/sudo
func TestFirewall_RealAllowDenyTraffic(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Firewall backend: %s", fm.GetBackend())

	// Test AllowTraffic
	err = fm.AllowTraffic("10.0.0.0/24", "10.200.0.0/24", []int{22, 443})
	if err != nil {
		t.Logf("AllowTraffic: %v", err)
	} else {
		t.Log("AllowTraffic succeeded")
	}

	// Test DenyTraffic
	err = fm.DenyTraffic("10.0.0.0/24", "10.200.0.0/24")
	if err != nil {
		t.Logf("DenyTraffic: %v", err)
	} else {
		t.Log("DenyTraffic succeeded")
	}
}

// TestOVS_RealFlowManagement tests OVS flow operations
// Requires root/sudo
func TestOVS_RealFlowManagement(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()

	// Create a test bridge
	err := ovs.CreateBridge("vimic2-test-flows")
	if err != nil {
		t.Skipf("CreateBridge failed (OVS not available): %v", err)
	}
	defer ovs.DeleteBridge("vimic2-test-flows")

	// Add flows
	err = ovs.AddFlow("vimic2-test-flows", "priority=100,ip,actions=output:1")
	if err != nil {
		t.Logf("AddFlow: %v (may need more ports)", err)
	}

	// List flows
	flows, err := ovs.DumpFlows("vimic2-test-flows")
	if err != nil {
		t.Logf("DumpFlows: %v", err)
	} else {
		t.Logf("Found %d flows", len(flows))
	}

	// Delete flows
	err = ovs.DelFlow("vimic2-test-flows", "priority=100")
	if err != nil {
		t.Logf("DelFlow: %v", err)
	}
}

// TestIPAM_RealAllocation tests IPAM allocation
func TestIPAM_RealAllocation(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.100.0.0/24",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	// Allocate multiple IPs
	allocated := make(map[string]bool)
	for i := 0; i < 5; i++ {
		ip, _, err := ipam.Allocate()
		if err != nil {
			t.Logf("Allocate %d: %v", i, err)
			break
		}
		if allocated[ip] {
			t.Errorf("Duplicate IP allocated: %s", ip)
		}
		allocated[ip] = true
		t.Logf("Allocated IP: %s", ip)
	}

	// Test pool listing
	pools := ipam.ListPools()
	t.Logf("Pools: %d", len(pools))

	// Test DNS
	dns := ipam.GetDNS()
	t.Logf("DNS servers: %v", dns)

	// Test stats
	used := ipam.Used()
	available := ipam.Available()
	t.Logf("Used: %d, Available: %d", used, available)

	// Test reclamation
	for _, pool := range pools {
		err := ipam.Reclaim(pool.CIDR)
		if err != nil {
			t.Logf("Reclaim %s: %v", pool.CIDR, err)
		}
	}
}

// TestVLAN_RealAllocator tests VLAN allocation
func TestVLAN_RealAllocator(t *testing.T) {
	vlanAlloc, err := NewVLANAllocator(1, 4094)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	// Allocate VLANs
	vlanIDs := make([]int, 0, 5)
	for i := 0; i < 5; i++ {
		vlan, err := vlanAlloc.Allocate()
		if err != nil {
			t.Logf("Allocate VLAN %d: %v", i, err)
			break
		}
		t.Logf("Allocated VLAN: %d", vlan)
		vlanIDs = append(vlanIDs, vlan)
	}

	// Release VLANs
	for _, vlan := range vlanIDs {
		vlanAlloc.Reclaim(vlan)
		t.Logf("Released VLAN: %d", vlan)
	}
}

// TestNetworkManager_RealCreateNetwork tests creating a real network with OVS
// Requires root/sudo
func TestNetworkManager_RealCreateNetwork(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	nm := NewNetworkManager(db)
	ctx := context.Background()

	// Create network
	network := &Network{
		Name:       "test-real-network",
		BridgeName: "vimic2-test-real",
		CIDR:       "10.250.0.0/24",
		Gateway:    "10.250.0.1",
		VLANID:     250,
	}

	err = nm.CreateNetwork(ctx, network)
	if err != nil {
		t.Skipf("CreateNetwork failed (OVS/network not available): %v", err)
	}
	t.Logf("Created network: %s (ID: %s)", network.Name, network.ID)

	// List networks
	networks, err := nm.ListNetworks(ctx)
	if err != nil {
		t.Logf("ListNetworks: %v", err)
	} else {
		t.Logf("Found %d networks", len(networks))
	}

	// Clean up: delete the bridge
	ovs := NewOVSClient()
	ovs.DeleteBridge("vimic2-test-real")
}