// Package mockovs_test tests the mock OVS client
package mockovs_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/testutil/mockovs"
)

// TestMockOVSClient_Create tests client creation
func TestMockOVSClient_Create(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	if client == nil {
		t.Fatal("client should not be nil")
	}
}

// TestMockOVSClient_CreateBridge tests bridge creation
func TestMockOVSClient_CreateBridge(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	err := client.CreateBridge("br-test")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}

	if !client.BridgeExists("br-test") {
		t.Error("bridge should exist")
	}
}

// TestMockOVSClient_DeleteBridge tests bridge deletion
func TestMockOVSClient_DeleteBridge(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	err := client.DeleteBridge("br-test")
	if err != nil {
		t.Fatalf("failed to delete bridge: %v", err)
	}

	if client.BridgeExists("br-test") {
		t.Error("bridge should not exist")
	}
}

// TestMockOVSClient_AddPort tests port addition
func TestMockOVSClient_AddPort(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	err := client.AddPort("br-test", "vnet0")
	if err != nil {
		t.Fatalf("failed to add port: %v", err)
	}

	if !client.PortExists("vnet0") {
		t.Error("port should exist")
	}
}

// TestMockOVSClient_DeletePort tests port deletion
func TestMockOVSClient_DeletePort(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")

	err := client.DeletePort("br-test", "vnet0")
	if err != nil {
		t.Fatalf("failed to delete port: %v", err)
	}

	if client.PortExists("vnet0") {
		t.Error("port should not exist")
	}
}

// TestMockOVSClient_SetPortVLAN tests VLAN setting
func TestMockOVSClient_SetPortVLAN(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")

	err := client.SetPortVLAN("vnet0", 100)
	if err != nil {
		t.Fatalf("failed to set VLAN: %v", err)
	}

	port, err := client.GetPort("vnet0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}

	if port.VLAN != 100 {
		t.Errorf("expected VLAN 100, got %d", port.VLAN)
	}
}

// TestMockOVSClient_SetPortTrunk tests trunk setting
func TestMockOVSClient_SetPortTrunk(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")

	err := client.SetPortTrunk("vnet0", []int{100, 200, 300})
	if err != nil {
		t.Fatalf("failed to set trunk: %v", err)
	}

	port, err := client.GetPort("vnet0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}

	if len(port.Trunk) != 3 {
		t.Errorf("expected 3 trunk VLANs, got %d", len(port.Trunk))
	}
}

// TestMockOVSClient_SetPortQoS tests QoS setting
func TestMockOVSClient_SetPortQoS(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")

	err := client.SetPortQoS("vnet0", 1000)
	if err != nil {
		t.Fatalf("failed to set QoS: %v", err)
	}

	port, err := client.GetPort("vnet0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}

	if port.QoS != 1000 {
		t.Errorf("expected QoS 1000, got %d", port.QoS)
	}
}

// TestMockOVSClient_SetPortSecurity tests port security
func TestMockOVSClient_SetPortSecurity(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")

	err := client.SetPortSecurity("vnet0", "52:54:00:12:34:56", "10.100.1.10")
	if err != nil {
		t.Fatalf("failed to set port security: %v", err)
	}

	port, err := client.GetPort("vnet0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}

	if port.MAC != "52:54:00:12:34:56" {
		t.Errorf("expected MAC, got %s", port.MAC)
	}
	if port.IPAddress != "10.100.1.10" {
		t.Errorf("expected IP, got %s", port.IPAddress)
	}
}

// TestMockOVSClient_CreateVXLAN tests VXLAN creation
func TestMockOVSClient_CreateVXLAN(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	err := client.CreateVXLAN("vxlan0", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create VXLAN: %v", err)
	}

	if !client.PortExists("vxlan0") {
		t.Error("VXLAN port should exist")
	}
}

// TestMockOVSClient_CreateGRE tests GRE creation
func TestMockOVSClient_CreateGRE(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	err := client.CreateGRE("gre0", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create GRE: %v", err)
	}

	if !client.PortExists("gre0") {
		t.Error("GRE port should exist")
	}
}

// TestMockOVSClient_AddFlow tests flow addition
func TestMockOVSClient_AddFlow(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")

	err := client.AddFlow("br-test", 100, "in_port=1", "output:2")
	if err != nil {
		t.Fatalf("failed to add flow: %v", err)
	}

	flows, err := client.ListFlows("br-test")
	if err != nil {
		t.Fatalf("failed to list flows: %v", err)
	}

	if len(flows) != 1 {
		t.Errorf("expected 1 flow, got %d", len(flows))
	}
}

// TestMockOVSClient_DeleteFlow tests flow deletion
func TestMockOVSClient_DeleteFlow(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddFlow("br-test", 100, "in_port=1", "output:2")

	flows, _ := client.ListFlows("br-test")
	flowID := flows[0].ID

	err := client.DeleteFlow("br-test", flowID)
	if err != nil {
		t.Fatalf("failed to delete flow: %v", err)
	}

	flows, _ = client.ListFlows("br-test")
	if len(flows) != 0 {
		t.Errorf("expected 0 flows, got %d", len(flows))
	}
}

// TestMockOVSClient_ErrorMode tests error mode
func TestMockOVSClient_ErrorMode(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	// Enable error mode
	client.SetErrorMode(true)

	// Operations should fail
	err := client.CreateBridge("br-test")
	if err == nil {
		t.Error("expected error in error mode")
	}

	// Disable error mode
	client.SetErrorMode(false)

	// Operations should succeed
	err = client.CreateBridge("br-test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockOVSClient_FailNext tests fail next
func TestMockOVSClient_FailNext(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	// Set up fail next
	client.FailNext()

	// This operation should fail
	err := client.CreateBridge("br-test")
	if err == nil {
		t.Error("expected error from fail next")
	}

	// This operation should succeed (failNext was consumed)
	err = client.CreateBridge("br-test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockOVSClient_Count tests counting
func TestMockOVSClient_Count(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.CreateBridge("br-test2")
	_ = client.AddPort("br-test", "vnet0")
	_ = client.AddPort("br-test", "vnet1")

	counts := client.Count()

	if counts["bridges"] != 2 {
		t.Errorf("expected 2 bridges, got %d", counts["bridges"])
	}
	if counts["ports"] != 2 {
		t.Errorf("expected 2 ports, got %d", counts["ports"])
	}
}

// TestMockOVSClient_ListBridges tests bridge listing
func TestMockOVSClient_ListBridges(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test1")
	_ = client.CreateBridge("br-test2")
	_ = client.CreateBridge("br-test3")

	bridges, err := client.ListBridges()
	if err != nil {
		t.Fatalf("failed to list bridges: %v", err)
	}

	if len(bridges) != 3 {
		t.Errorf("expected 3 bridges, got %d", len(bridges))
	}
}

// TestMockOVSClient_ListPorts tests port listing
func TestMockOVSClient_ListPorts(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")
	_ = client.AddPort("br-test", "vnet0")
	_ = client.AddPort("br-test", "vnet1")

	ports, err := client.ListPorts()
	if err != nil {
		t.Fatalf("failed to list ports: %v", err)
	}

	if len(ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(ports))
	}
}

// TestMockOVSClient_CreateTestBridge tests test helper
func TestMockOVSClient_CreateTestBridge(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	err := client.CreateTestBridge("br-test")
	if err != nil {
		t.Fatalf("failed to create test bridge: %v", err)
	}

	bridge, err := client.GetBridge("br-test")
	if err != nil {
		t.Fatalf("failed to get bridge: %v", err)
	}

	if bridge.VLAN != 100 {
		t.Errorf("expected VLAN 100, got %d", bridge.VLAN)
	}
}

// TestMockOVSClient_CreateTestPort tests test helper
func TestMockOVSClient_CreateTestPort(t *testing.T) {
	client := mockovs.NewMockOVSClient()

	_ = client.CreateBridge("br-test")

	err := client.CreateTestPort("br-test", "vnet0")
	if err != nil {
		t.Fatalf("failed to create test port: %v", err)
	}

	port, err := client.GetPort("vnet0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}

	if port.VLAN != 100 {
		t.Errorf("expected VLAN 100, got %d", port.VLAN)
	}
}