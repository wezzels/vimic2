//go:build integration
// +build integration

// Package realovs_test tests the real OVS client with actual OVS
package realovs_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realovs"
)

// TestRealOVS_CreateBridge_Integration tests bridge creation with real OVS
func TestRealOVS_CreateBridge_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create bridge
	err := client.CreateBridge("br-test-int")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}

	// Verify bridge exists
	if !client.BridgeExists("br-test-int") {
		t.Error("bridge should exist")
	}

	// Cleanup
	err = client.DeleteBridge("br-test-int")
	if err != nil {
		t.Fatalf("failed to delete bridge: %v", err)
	}
}

// TestRealOVS_BridgeOperations_Integration tests bridge operations
func TestRealOVS_BridgeOperations_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create bridge
	err := client.CreateBridge("br-test-ops")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer client.DeleteBridge("br-test-ops")

	// Set VLAN
	err = client.SetBridgeVLAN("br-test-ops", 100)
	if err != nil {
		t.Errorf("failed to set VLAN: %v", err)
	}

	// Set trunk
	err = client.SetBridgeTrunk("br-test-ops", []int{200, 300})
	if err != nil {
		t.Errorf("failed to set trunk: %v", err)
	}

	// List bridges
	bridges, err := client.ListBridges()
	if err != nil {
		t.Fatalf("failed to list bridges: %v", err)
	}

	found := false
	for _, b := range bridges {
		if b.Name == "br-test-ops" {
			found = true
			break
		}
	}
	if !found {
		t.Error("bridge should be in list")
	}
}

// TestRealOVS_PortOperations_Integration tests port operations
func TestRealOVS_PortOperations_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create bridge
	err := client.CreateBridge("br-test-port")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer client.DeleteBridge("br-test-port")

	// Add port
	err = client.AddPort("br-test-port", "vnet0")
	if err != nil {
		t.Fatalf("failed to add port: %v", err)
	}

	// Set VLAN
	err = client.SetPortVLAN("vnet0", 100)
	if err != nil {
		t.Errorf("failed to set port VLAN: %v", err)
	}

	// Set QoS
	err = client.SetPortQoS("vnet0", 1000)
	if err != nil {
		t.Errorf("failed to set port QoS: %v", err)
	}

	// List ports
	ports, err := client.ListPorts("br-test-port")
	if err != nil {
		t.Fatalf("failed to list ports: %v", err)
	}

	found := false
	for _, p := range ports {
		if p.Name == "vnet0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("port should be in list")
	}

	// Delete port
	err = client.DeletePort("br-test-port", "vnet0")
	if err != nil {
		t.Fatalf("failed to delete port: %v", err)
	}
}

// TestRealOVS_FlowOperations_Integration tests flow operations
func TestRealOVS_FlowOperations_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create bridge
	err := client.CreateBridge("br-test-flow")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer client.DeleteBridge("br-test-flow")

	// Add flow
	err = client.AddFlow("br-test-flow", 100, "in_port=1", "output:2")
	if err != nil {
		t.Fatalf("failed to add flow: %v", err)
	}

	// List flows
	flows, err := client.ListFlows("br-test-flow")
	if err != nil {
		t.Fatalf("failed to list flows: %v", err)
	}

	if len(flows) < 1 {
		t.Error("should have at least one flow")
	}

	// Clear flows
	err = client.ClearFlows("br-test-flow")
	if err != nil {
		t.Fatalf("failed to clear flows: %v", err)
	}
}

// TestRealOVS_VXLAN_Integration tests VXLAN tunnel creation
func TestRealOVS_VXLAN_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Ensure integration bridge exists
	_ = client.CreateBridge("br-int")

	// Create VXLAN
	err := client.CreateVXLAN("vxlan-test", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create VXLAN: %v", err)
	}

	// Cleanup
	_ = client.DeletePort("br-int", "vxlan-test")
}

// TestRealOVS_GRE_Integration tests GRE tunnel creation
func TestRealOVS_GRE_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Ensure integration bridge exists
	_ = client.CreateBridge("br-int")

	// Create GRE
	err := client.CreateGRE("gre-test", "10.0.0.2", 200)
	if err != nil {
		t.Fatalf("failed to create GRE: %v", err)
	}

	// Cleanup
	_ = client.DeletePort("br-int", "gre-test")
}

// TestRealOVS_Version tests version detection
func TestRealOVS_Version(t *testing.T) {
	version, err := realovs.Version()
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}

	if version == "" {
		t.Error("version should not be empty")
	}

	t.Logf("OVS version: %s", version)
}

// TestRealOVS_IsAvailable tests availability check
func TestRealOVS_IsAvailable(t *testing.T) {
	if !realovs.IsAvailable() {
		t.Error("OVS should be available")
	}
}

// TestRealOVS_MultipleBridges_Integration tests multiple bridges
func TestRealOVS_MultipleBridges_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create multiple bridges
	for i := 0; i < 3; i++ {
		bridgeName := "br-test-multi-" + string(rune('0'+i))
		err := client.CreateBridge(bridgeName)
		if err != nil {
			t.Fatalf("failed to create bridge %s: %v", bridgeName, err)
		}
		defer client.DeleteBridge(bridgeName)
	}

	// List bridges
	bridges, err := client.ListBridges()
	if err != nil {
		t.Fatalf("failed to list bridges: %v", err)
	}

	// Should have at least 3 bridges we created
	count := 0
	for _, b := range bridges {
		if b.Name == "br-test-multi-0" || b.Name == "br-test-multi-1" || b.Name == "br-test-multi-2" {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 test bridges, found %d", count)
	}
}

// TestRealOVS_PortSecurity_Integration tests port security
func TestRealOVS_PortSecurity_Integration(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	// Create bridge
	err := client.CreateBridge("br-test-sec")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}
	defer client.DeleteBridge("br-test-sec")

	// Add port
	err = client.AddPort("br-test-sec", "vnet-sec")
	if err != nil {
		t.Fatalf("failed to add port: %v", err)
	}
	defer client.DeletePort("br-test-sec", "vnet-sec")

	// Set port security
	err = client.SetPortSecurity("vnet-sec", "52:54:00:12:34:56", "10.100.1.10")
	if err != nil {
		t.Errorf("failed to set port security: %v", err)
	}
}
