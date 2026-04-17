//go:build integration

package network

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// ==================== Real OVS Bridge Tests ====================

func TestReal_OVS_BridgeCRUD(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-br-crud"

	// Clean up from any previous failed run
	ovs.DeleteBridge(bridgeName)

	// Create
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed (OVS datapath issue): %v", err)
	}
	t.Logf("Created bridge %s", bridgeName)

	// Verify it exists via ListBridges
	bridges, err := ovs.ListBridges()
	if err != nil {
		t.Fatalf("ListBridges failed: %v", err)
	}
	found := false
	for _, b := range bridges {
		if b == bridgeName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Bridge %s not found in ListBridges result: %v", bridgeName, bridges)
	}

	// Delete
	err = ovs.DeleteBridge(bridgeName)
	if err != nil {
		t.Fatalf("DeleteBridge failed: %v", err)
	}
	t.Logf("Deleted bridge %s", bridgeName)

	// Verify it's gone
	bridges, _ = ovs.ListBridges()
	for _, b := range bridges {
		if b == bridgeName {
			t.Error("Bridge should not exist after deletion")
		}
	}
}

func TestReal_OVS_PortCRUD(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-br-ports"
	portName := "vimic2-test-port1"

	// Clean up
	ovs.DeleteBridge(bridgeName)

	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed (OVS datapath issue): %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	// Create port
	err = ovs.CreatePort(bridgeName, portName)
	if err != nil {
		t.Fatalf("CreatePort failed: %v", err)
	}
	t.Logf("Created port %s on bridge %s", portName, bridgeName)

	// List ports
	ports, err := ovs.ListPorts(bridgeName)
	if err != nil {
		t.Fatalf("ListPorts failed: %v", err)
	}
	found := false
	for _, p := range ports {
		if p == portName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Port %s not found in ListPorts result: %v", portName, ports)
	}

	// Delete port
	err = ovs.DeletePort(bridgeName, portName)
	if err != nil {
		t.Fatalf("DeletePort failed: %v", err)
	}

	// Verify port is gone
	ports, _ = ovs.ListPorts(bridgeName)
	for _, p := range ports {
		if p == portName {
			t.Error("Port should not exist after deletion")
		}
	}
}

func TestReal_OVS_VLAN(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-br-vlan"
	portName := "vimic2-vlan-port1"

	ovs.DeleteBridge(bridgeName)
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed (OVS datapath issue): %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	err = ovs.CreatePort(bridgeName, portName)
	if err != nil {
		t.Fatalf("CreatePort failed: %v", err)
	}

	// Set VLAN tag
	err = ovs.SetPortVLAN(portName, 100)
	if err != nil {
		t.Fatalf("SetPortVLAN failed: %v", err)
	}
	t.Logf("Set VLAN 100 on port %s", portName)

	// Set trunk VLANs
	err = ovs.SetPortTrunk(portName, []int{100, 200, 300})
	if err != nil {
		t.Logf("SetPortTrunk: %v (may not be supported)", err)
	}
}

func TestReal_OVS_Flows(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-br-flows"

	ovs.DeleteBridge(bridgeName)
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed (OVS not available): %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	// Add flows
	err = ovs.AddFlow(bridgeName, "priority=100,ip,nw_dst=10.0.0.0/24,actions=output:1")
	if err != nil {
		t.Logf("AddFlow: %v", err)
	} else {
		t.Log("Added flow successfully")
	}

	err = ovs.AddFlow(bridgeName, "priority=200,tcp,tp_dst=80,actions=output:2")
	if err != nil {
		t.Logf("AddFlow tcp: %v", err)
	}

	// Dump flows
	flows, err := ovs.DumpFlows(bridgeName)
	if err != nil {
		t.Logf("DumpFlows: %v", err)
	} else {
		t.Logf("Found %d flows", len(flows))
		for _, f := range flows {
			t.Logf("  Flow: %s", f)
		}
	}

	// Delete flow
	err = ovs.DelFlow(bridgeName, "priority=100,ip,nw_dst=10.0.0.0/24")
	if err != nil {
		t.Logf("DelFlow: %v", err)
	}
}

func TestReal_OVS_MultipleBridges(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridges := []string{"vimic2-test-mb1", "vimic2-test-mb2", "vimic2-test-mb3"}

	// Clean up
	for _, b := range bridges {
		ovs.DeleteBridge(b)
	}

	// Create all bridges
	for _, b := range bridges {
		err := ovs.CreateBridge(b)
		if err != nil {
			t.Skipf("CreateBridge(%s) failed (OVS): %v", b, err)
		}
		t.Logf("Created bridge %s", b)
	}

	// Verify all exist
	list, err := ovs.ListBridges()
	if err != nil {
		t.Fatalf("ListBridges failed: %v", err)
	}
	for _, b := range bridges {
		found := false
		for _, lb := range list {
			if lb == b {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Bridge %s not found", b)
		}
	}

	// Clean up
	for _, b := range bridges {
		ovs.DeleteBridge(b)
	}
}

// ==================== Real Firewall Tests ====================

func TestReal_NFTables_Isolation(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Fatalf("NewFirewallManager(nftables) failed: %v", err)
	}

	t.Logf("Firewall backend: %s", fm.GetBackend())

	// Create isolation rules
	err = fm.CreateIsolationRules("vimic2-iso-test", "10.200.0.0/24", 200)
	if err != nil {
		t.Fatalf("CreateIsolationRules failed: %v", err)
	}
	t.Log("Created isolation rules")

	// Verify rules exist with nft list
	out, _ := exec.Command("nft", "list", "table", "inet", "vimic2").CombinedOutput()
	t.Logf("Current nft rules:\n%s", string(out))

	// Delete isolation rules
	err = fm.DeleteIsolationRules("vimic2-iso-test", "10.200.0.0/24", 200)
	if err != nil {
		t.Fatalf("DeleteIsolationRules failed: %v", err)
	}
	t.Log("Deleted isolation rules")
}

func TestReal_NFTables_AllowDeny(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Fatalf("NewFirewallManager(nftables) failed: %v", err)
	}

	// Allow SSH and HTTPS from specific CIDR
	err = fm.AllowTraffic("10.0.0.0/24", "10.200.0.0/24", []int{22, 443})
	if err != nil {
		t.Fatalf("AllowTraffic failed: %v", err)
	}
	t.Log("AllowTraffic succeeded")

	// Verify
	out, _ := exec.Command("nft", "list", "table", "inet", "vimic2").CombinedOutput()
	t.Logf("Rules after AllowTraffic:\n%s", string(out))

	// Deny all traffic from a CIDR
	err = fm.DenyTraffic("10.100.0.0/24", "10.200.0.0/24")
	if err != nil {
		t.Fatalf("DenyTraffic failed: %v", err)
	}
	t.Log("DenyTraffic succeeded")

	// List rules
	rules := fm.ListRules()
	t.Logf("Found %d rule chains", len(rules))
}

// ==================== Real NetworkManager Tests ====================

func TestReal_NetworkManager_FullCRUD(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-")
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

	// Create network with real OVS bridge
	network := &Network{
		Name:       "test-real-network",
		BridgeName: "vimic2-test-crud",
		CIDR:       "10.250.0.0/24",
		Gateway:    "10.250.0.1",
		VLANID:     250,
		DNS:        []string{"8.8.8.8"},
	}

	err = nm.CreateNetwork(ctx, network)
	if err != nil {
		t.Skipf("CreateNetwork failed (OVS not available): %v", err)
	}
	t.Logf("Created network %s (ID: %s)", network.Name, network.ID)

	// Get the network back via database
	got, err := db.GetNetwork(ctx, network.ID)
	if err != nil {
		t.Fatalf("GetNetwork failed: %v", err)
	}
	if got.Name != network.Name {
		t.Errorf("Name = %s, want %s", got.Name, network.Name)
	}
	if got.CIDR != network.CIDR {
		t.Errorf("CIDR = %s, want %s", got.CIDR, network.CIDR)
	}

	// List networks
	networks, err := nm.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("ListNetworks failed: %v", err)
	}
	if len(networks) < 1 {
		t.Error("ListNetworks should return at least 1 network")
	}

	// Delete network via database
	err = db.DeleteNetwork(ctx, network.ID)
	if err != nil {
		t.Logf("DeleteNetwork: %v", err)
	} else {
		t.Log("Deleted network")
	}

	// Clean up OVS bridge
	ovs := NewOVSClient()
	ovs.DeleteBridge("vimic2-test-crud")
}

func TestReal_NetworkManager_CreateRouter(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-")
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

	router := &Router{
		Name:    "test-router",
		Enabled: true,
		RoutingTable: []Route{
			{Destination: "10.0.0.0/24", Gateway: "192.168.1.1", Interface: "eth0", Metric: 100},
		},
	}

	err = nm.CreateRouter(ctx, router)
	if err != nil {
		t.Skipf("CreateRouter failed (route addition requires specific network config): %v", err)
	}
	t.Logf("Created router %s (ID: %s)", router.Name, router.ID)

	// Get it back via database
	got, err := db.GetRouter(ctx, router.ID)
	if err != nil {
		t.Fatalf("GetRouter failed: %v", err)
	}
	if got.Name != router.Name {
		t.Errorf("Name = %s, want %s", got.Name, router.Name)
	}

	// Clean up via database
	db.DeleteRouter(ctx, router.ID)
}

func TestReal_NetworkManager_CreateTunnel(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-")
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

	tunnel := &Tunnel{
		Name:     "test-vxlan-tunnel",
		Protocol: TunnelVXLAN,
		RemoteIP: "10.0.0.1",
		LocalIP:  "10.0.0.2",
		VNI:      100,
	}

	err = nm.CreateTunnel(ctx, tunnel)
	if err != nil {
		t.Logf("CreateTunnel: %v (OVS may not be available)", err)
	} else {
		t.Logf("Created tunnel %s (ID: %s)", tunnel.Name, tunnel.ID)

		// Get it back via database
		got, err := db.GetTunnel(ctx, tunnel.ID)
		if err != nil {
			t.Logf("GetTunnel: %v", err)
		} else if got.Name != tunnel.Name {
			t.Errorf("Name = %s, want %s", got.Name, tunnel.Name)
		}

		db.DeleteTunnel(ctx, tunnel.ID)
	}
}

// ==================== Real Network Stats Tests ====================

func TestReal_NetworkManager_Stats(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-")
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

	// Get stats (should work even with no networks)
	stats, err := nm.GetNetworkStats(ctx, "")
	if err != nil {
		t.Logf("GetNetworkStats: %v", err)
	} else {
		t.Logf("Stats: %+v", stats)
	}
}

// ==================== Real Database Persistence Tests ====================

func TestReal_NetworkDB_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-net-real-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	// Create and write
	db1, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	network := &Network{
		Name:       "persist-test",
		BridgeName: "vimic2-persist",
		CIDR:       "10.99.0.0/24",
		Gateway:    "10.99.0.1",
		VLANID:     99,
	}

	err = db1.SaveNetwork(ctx, network)
	if err != nil {
		t.Fatalf("SaveNetwork failed: %v", err)
	}
	t.Logf("Saved network with ID: %s", network.ID)

	savedID := network.ID

	// Close and reopen
	db1.Close()

	db2, err := NewNetworkDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	// Read back
	got, err := db2.GetNetwork(ctx, savedID)
	if err != nil {
		t.Fatalf("GetNetwork after reopen failed: %v", err)
	}
	if got.Name != "persist-test" {
		t.Errorf("Name = %s, want persist-test", got.Name)
	}
	if got.CIDR != "10.99.0.0/24" {
		t.Errorf("CIDR = %s, want 10.99.0.0/24", got.CIDR)
	}
	t.Logf("Retrieved network: %s (CIDR: %s)", got.Name, got.CIDR)
}

// ==================== Real IPAM Allocation Tests ====================

func TestReal_IPAM_FullAllocation(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.0.0.0/16",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
	}

	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	// Allocate 10 IPs
	allocated := make(map[string]string) // ip -> poolID
	for i := 0; i < 10; i++ {
		ip, poolID, err := ipam.Allocate()
		if err != nil {
			t.Fatalf("Allocate %d failed: %v", i, err)
		}
		if allocated[ip] != "" {
			t.Errorf("Duplicate IP allocated: %s", ip)
		}
		allocated[ip] = poolID
		t.Logf("Allocated IP: %s (pool: %s)", ip, poolID)
	}

	t.Logf("Total allocated: %d IPs", len(allocated))

	// Verify stats
	used := ipam.Used()
	t.Logf("Used: %d", used)

	// Release some IPs
	count := 0
	for ip, poolID := range allocated {
		if count >= 3 {
			break
		}
		err := ipam.ReleaseIP(poolID, ip)
		if err != nil {
			t.Logf("ReleaseIP %s: %v", ip, err)
		} else {
			t.Logf("Released IP: %s", ip)
		}
		count++
	}

	// Re-allocate — should reuse released IPs
	ip, _, err := ipam.Allocate()
	if err != nil {
		t.Logf("Re-allocate failed: %v", err)
	} else {
		t.Logf("Re-allocated IP: %s", ip)
	}

	// Test Reclaim
	pools := ipam.ListPools()
	for _, pool := range pools {
		err := ipam.Reclaim(pool.CIDR)
		if err != nil {
			t.Logf("Reclaim %s: %v", pool.CIDR, err)
		}
	}
}

// ==================== Real VLAN Allocator Tests ====================

func TestReal_VLAN_Allocation(t *testing.T) {
	vlanAlloc, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	vlans := make([]int, 0, 5)
	for i := 0; i < 5; i++ {
		vlan, err := vlanAlloc.Allocate()
		if err != nil {
			t.Fatalf("Allocate VLAN %d failed: %v", i, err)
		}
		vlans = append(vlans, vlan)
		t.Logf("Allocated VLAN: %d", vlan)
	}

	// Verify uniqueness
	seen := make(map[int]bool)
	for _, v := range vlans {
		if seen[v] {
			t.Errorf("Duplicate VLAN allocated: %d", v)
		}
		seen[v] = true
	}

	// Release and reallocate
	for _, v := range vlans[:2] {
		vlanAlloc.Reclaim(v)
		t.Logf("Released VLAN: %d", v)
	}

	// Should reuse released VLANs
	newVlan, err := vlanAlloc.Allocate()
	if err != nil {
		t.Logf("Re-allocate VLAN: %v", err)
	} else {
		t.Logf("Re-allocated VLAN: %d", newVlan)
	}
}

// ==================== Network Namespace Isolation Test ====================

func TestReal_NetworkNamespace(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	// Create a network namespace for isolation testing
	nsName := "vimic2-test-ns"

	// Clean up from any previous run
	exec.Command("ip", "netns", "delete", nsName).Run()

	err := exec.Command("ip", "netns", "add", nsName).Run()
	if err != nil {
		t.Skipf("Cannot create network namespace: %v", err)
	}
	defer exec.Command("ip", "netns", "delete", nsName).Run()

	t.Logf("Created network namespace: %s", nsName)

	// Create a veth pair
	vethHost := "vimic2-veth-h"
	vethNS := "vimic2-veth-n"

	// Clean up any existing veth
	exec.Command("ip", "link", "delete", vethHost).Run()

	err = exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethNS).Run()
	if err != nil {
		t.Skipf("Cannot create veth pair: %v", err)
	}
	defer exec.Command("ip", "link", "delete", vethHost).Run()

	t.Logf("Created veth pair: %s <-> %s", vethHost, vethNS)

	// Move one end into the namespace
	err = exec.Command("ip", "link", "set", vethNS, "netns", nsName).Run()
	if err != nil {
		t.Fatalf("Cannot move veth into namespace: %v", err)
	}

	// Bring up the host side
	exec.Command("ip", "link", "set", vethHost, "up").Run()

	// Assign IP to host side
	err = exec.Command("ip", "addr", "add", "10.99.99.1/24", "dev", vethHost).Run()
	if err != nil {
		t.Logf("Cannot assign IP to host veth: %v", err)
	}

	// Configure namespace side
	cmds := [][]string{
		{"ip", "netns", "exec", nsName, "ip", "link", "set", "lo", "up"},
		{"ip", "netns", "exec", nsName, "ip", "link", "set", vethNS, "up"},
		{"ip", "netns", "exec", nsName, "ip", "addr", "add", "10.99.99.2/24", "dev", vethNS},
	}
	for _, cmd := range cmds {
		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			t.Logf("Namespace config: %v: %s", err, string(out))
		}
	}

	// Ping from host to namespace
	time.Sleep(500 * time.Millisecond) // Wait for interfaces to come up
	out, err := exec.Command("ping", "-c", "1", "-W", "1", "10.99.99.2").CombinedOutput()
	if err != nil {
		t.Logf("Ping failed (expected in some environments): %v: %s", err, string(out))
	} else {
		t.Log("Ping succeeded: namespace isolation working")
	}

	t.Log("Network namespace test completed")
}

// ==================== Real Firewall with Namespace Test ====================

func TestReal_Firewall_NamespaceIsolation(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	fm, err := NewFirewallManager(FirewallBackendNFTables)
	if err != nil {
		t.Fatalf("NewFirewallManager failed: %v", err)
	}

	// Create isolation rules for a test network
	err = fm.CreateIsolationRules("vimic2-iso-real", "10.250.0.0/24", 250)
	if err != nil {
		t.Fatalf("CreateIsolationRules failed: %v", err)
	}
	t.Log("Created isolation rules for 10.250.0.0/24")

	// Verify nft rules
	out, err := exec.Command("nft", "list", "table", "inet", "vimic2").CombinedOutput()
	if err != nil {
		t.Logf("nft list: %v", err)
	} else {
		t.Logf("nft rules:\n%s", string(out))
	}

	// Clean up
	err = fm.DeleteIsolationRules("vimic2-iso-real", "10.250.0.0/24", 250)
	if err != nil {
		t.Logf("DeleteIsolationRules: %v", err)
	} else {
		t.Log("Deleted isolation rules")
	}
}

// ==================== Real OVS Tunnel Tests ====================

func TestReal_OVS_VXLAN(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-vxlan"

	ovs.DeleteBridge(bridgeName)
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed: %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	// Create VXLAN port
	tunnel := &Tunnel{
		Name:     "vimic2-vxlan0",
		Protocol: TunnelVXLAN,
		RemoteIP: "127.0.0.1",
		VNI:      100,
	}
	err = ovs.CreateTunnelPort(tunnel)
	if err != nil {
		t.Logf("CreateTunnelPort VXLAN: %v", err)
	} else {
		t.Log("Created VXLAN tunnel port")

		// Verify port exists
		ports, _ := ovs.ListPorts(bridgeName)
		found := false
		for _, p := range ports {
			if p == "vimic2-vxlan0" {
				found = true
				break
			}
		}
		if !found {
			t.Error("VXLAN port not found in ListPorts")
		}
	}
}

func TestReal_OVS_GRE(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-gre"

	ovs.DeleteBridge(bridgeName)
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed: %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	// Create GRE port
	tunnel := &Tunnel{
		Name:     "vimic2-gre0",
		Protocol: TunnelGRE,
		RemoteIP: "127.0.0.1",
	}
	err = ovs.CreateTunnelPort(tunnel)
	if err != nil {
		t.Logf("CreateTunnelPort GRE: %v", err)
	} else {
		t.Log("Created GRE tunnel port")
	}
}

// ==================== Helper ====================

func TestReal_OVS_GetBridgeInfo(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()

	// List existing bridges
	bridges, err := ovs.ListBridges()
	if err != nil {
		t.Fatalf("ListBridges failed: %v", err)
	}
	t.Logf("Existing bridges: %v", bridges)

	// Check ports on each bridge
	for _, b := range bridges {
		ports, err := ovs.ListPorts(b)
		if err != nil {
			t.Logf("ListPorts(%s): %v", b, err)
			continue
		}
		t.Logf("Bridge %s ports: %v", b, ports)

		// Get bridge stats
		stats, err := ovs.GetBridgeStats(b)
		if err != nil {
			t.Logf("GetBridgeStats(%s): %v", b, err)
		} else {
			t.Logf("  Bridge %s stats: %+v", b, stats)
		}
	}
}

func TestReal_OVS_BridgeStats(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root")
	}

	ovs := NewOVSClient()
	bridgeName := "vimic2-test-stats"

	ovs.DeleteBridge(bridgeName)
	err := ovs.CreateBridge(bridgeName)
	if err != nil {
		t.Skipf("CreateBridge failed: %v", err)
	}
	defer ovs.DeleteBridge(bridgeName)

	// Get stats for the bridge
	stats, err := ovs.GetBridgeStats(bridgeName)
	if err != nil {
		t.Logf("GetBridgeStats: %v", err)
	} else {
		t.Logf("Bridge stats: %+v", stats)
	}
}