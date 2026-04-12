//go:build integration

// Package network provides integration tests with real OVS
package network

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestIntegration_OVS_CreateDeleteBridge tests creating and deleting a bridge
func TestIntegration_OVS_CreateDeleteBridge(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)

	// Create bridge
	err := client.CreateBridge(bridgeName)
	if err != nil {
		t.Fatalf("CreateBridge failed: %v", err)
	}

	// Verify bridge exists
	bridges, err := client.ListBridges()
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
		t.Errorf("bridge %s not found in list", bridgeName)
	}

	// Delete bridge
	err = client.DeleteBridge(bridgeName)
	if err != nil {
		t.Fatalf("DeleteBridge failed: %v", err)
	}

	// Verify deleted
	bridges, _ = client.ListBridges()
	for _, b := range bridges {
		if b == bridgeName {
			t.Errorf("bridge %s still exists after delete", bridgeName)
		}
	}
}

// TestIntegration_OVS_CreateDeletePort tests creating and deleting ports
func TestIntegration_OVS_CreateDeletePort(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)
	portName := fmt.Sprintf("test-port-%d", time.Now().Unix()%10000)

	// Setup: create bridge
	if err := client.CreateBridge(bridgeName); err != nil {
		t.Fatalf("CreateBridge failed: %v", err)
	}
	defer client.DeleteBridge(bridgeName)

	// Create port
	err := client.CreatePort(bridgeName, portName)
	if err != nil {
		t.Fatalf("CreatePort failed: %v", err)
	}

	// Verify port exists
	ports, err := client.ListPorts(bridgeName)
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
		t.Errorf("port %s not found in list", portName)
	}

	// Delete port
	err = client.DeletePort(bridgeName, portName)
	if err != nil {
		t.Fatalf("DeletePort failed: %v", err)
	}

	// Verify deleted
	ports, _ = client.ListPorts(bridgeName)
	for _, p := range ports {
		if p == portName {
			t.Errorf("port %s still exists after delete", portName)
		}
	}
}

// TestIntegration_OVS_SetPortVLAN tests setting VLAN on a port
func TestIntegration_OVS_SetPortVLAN(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)
	portName := fmt.Sprintf("test-port-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)
	client.CreatePort(bridgeName, portName)
	defer client.DeletePort(bridgeName, portName)

	// Set VLAN
	err := client.SetPortVLAN(portName, 100)
	if err != nil {
		t.Fatalf("SetPortVLAN failed: %v", err)
	}

	// Verify VLAN was set
	info, err := client.GetPortInfo(portName)
	if err != nil {
		t.Fatalf("GetPortInfo failed: %v", err)
	}
	t.Logf("Port info: %v", info)
}

// TestIntegration_OVS_SetPortTrunk tests setting trunk ports
func TestIntegration_OVS_SetPortTrunk(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)
	portName := fmt.Sprintf("test-port-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)
	client.CreatePort(bridgeName, portName)
	defer client.DeletePort(bridgeName, portName)

	// Set trunk
	err := client.SetPortTrunk(portName, []int{100, 200, 300})
	if err != nil {
		t.Fatalf("SetPortTrunk failed: %v", err)
	}
}

// TestIntegration_OVS_BridgeVLAN tests bridge VLAN configuration
func TestIntegration_OVS_BridgeVLAN(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)

	// Set bridge VLAN
	err := client.SetBridgeVLAN(bridgeName, 50)
	if err != nil {
		t.Fatalf("SetBridgeVLAN failed: %v", err)
	}

	// Set bridge trunk
	err = client.SetBridgeTrunk(bridgeName, []int{10, 20, 30})
	if err != nil {
		t.Fatalf("SetBridgeTrunk failed: %v", err)
	}
}

// TestIntegration_OVS_QoS tests setting QoS on a port
func TestIntegration_OVS_SetPortQoS(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)
	portName := fmt.Sprintf("test-port-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)
	client.CreatePort(bridgeName, portName)
	defer client.DeletePort(bridgeName, portName)

	// Set QoS (100 Mbps)
	err := client.SetPortQoS(portName, 100)
	if err != nil {
		t.Fatalf("SetPortQoS failed: %v", err)
	}
}

// TestIntegration_OVS_PortSecurity tests enabling port security
func TestIntegration_OVS_EnablePortSecurity(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)
	portName := fmt.Sprintf("test-port-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)
	client.CreatePort(bridgeName, portName)
	defer client.DeletePort(bridgeName, portName)

	// Enable port security
	err := client.EnablePortSecurity(portName, "00:11:22:33:44:55", "10.0.0.100")
	if err != nil {
		t.Fatalf("EnablePortSecurity failed: %v", err)
	}
}

// TestIntegration_OVS_Flows tests OpenFlow operations
func TestIntegration_OVS_Flows(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)

	// Add flow
	flow := "actions=NORMAL"
	err := client.AddFlow(bridgeName, flow)
	if err != nil {
		t.Fatalf("AddFlow failed: %v", err)
	}

	// List flows
	flows, err := client.DumpFlows(bridgeName)
	if err != nil {
		t.Fatalf("DumpFlows failed: %v", err)
	}
	found := false
	for _, f := range flows {
		if strings.Contains(f, "NORMAL") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("flow not found in dump")
	}

	// Delete flow
	err = client.DelFlow(bridgeName, flow)
	if err != nil {
		t.Fatalf("DelFlow failed: %v", err)
	}
}

// TestIntegration_OVS_GetBridgeStats tests getting bridge statistics
func TestIntegration_OVS_GetBridgeStats(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root/sudo for OVS operations")
	}

	client := NewOVSClient()
	bridgeName := fmt.Sprintf("test-br-%d", time.Now().Unix()%10000)

	// Setup
	client.CreateBridge(bridgeName)
	defer client.DeleteBridge(bridgeName)

	// Get stats
	stats, err := client.GetBridgeStats(bridgeName)
	if err != nil {
		t.Fatalf("GetBridgeStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("expected non-nil stats")
	}

	t.Logf("Bridge stats: %+v", stats)
}
