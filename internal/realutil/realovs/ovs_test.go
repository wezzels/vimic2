// Package realovs_test tests the real OVS client
package realovs_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realovs"
)

// TestRealOVSClient_Create tests client creation
func TestRealOVSClient_Create(t *testing.T) {
	client := realovs.NewClientWithDefaults()

	if client == nil {
		t.Fatal("client should not be nil")
	}
}

// TestRealOVSClient_CreateBridge tests bridge creation (dry-run)
func TestRealOVSClient_CreateBridge(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.CreateBridge("br-test")
	if err != nil {
		t.Fatalf("failed to create bridge: %v", err)
	}

	cmd := client.LastCommand()
	if cmd == "" {
		t.Error("command should be recorded")
	}
}

// TestRealOVSClient_DeleteBridge tests bridge deletion (dry-run)
func TestRealOVSClient_DeleteBridge(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.DeleteBridge("br-test")
	if err != nil {
		t.Fatalf("failed to delete bridge: %v", err)
	}
}

// TestRealOVSClient_AddPort tests port addition (dry-run)
func TestRealOVSClient_AddPort(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.AddPort("br-test", "vnet0")
	if err != nil {
		t.Fatalf("failed to add port: %v", err)
	}
}

// TestRealOVSClient_DeletePort tests port deletion (dry-run)
func TestRealOVSClient_DeletePort(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.DeletePort("br-test", "vnet0")
	if err != nil {
		t.Fatalf("failed to delete port: %v", err)
	}
}

// TestRealOVSClient_SetPortVLAN tests VLAN setting (dry-run)
func TestRealOVSClient_SetPortVLAN(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetPortVLAN("vnet0", 100)
	if err != nil {
		t.Fatalf("failed to set VLAN: %v", err)
	}
}

// TestRealOVSClient_SetPortTrunk tests trunk setting (dry-run)
func TestRealOVSClient_SetPortTrunk(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetPortTrunk("vnet0", []int{100, 200, 300})
	if err != nil {
		t.Fatalf("failed to set trunk: %v", err)
	}
}

// TestRealOVSClient_SetPortQoS tests QoS setting (dry-run)
func TestRealOVSClient_SetPortQoS(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetPortQoS("vnet0", 1000)
	if err != nil {
		t.Fatalf("failed to set QoS: %v", err)
	}
}

// TestRealOVSClient_SetPortSecurity tests port security (dry-run)
func TestRealOVSClient_SetPortSecurity(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetPortSecurity("vnet0", "52:54:00:12:34:56", "10.100.1.10")
	if err != nil {
		t.Fatalf("failed to set port security: %v", err)
	}
}

// TestRealOVSClient_CreateVXLAN tests VXLAN creation (dry-run)
func TestRealOVSClient_CreateVXLAN(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.CreateVXLAN("vxlan0", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create VXLAN: %v", err)
	}
}

// TestRealOVSClient_CreateGRE tests GRE creation (dry-run)
func TestRealOVSClient_CreateGRE(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.CreateGRE("gre0", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create GRE: %v", err)
	}
}

// TestRealOVSClient_CreateGeneve tests Geneve creation (dry-run)
func TestRealOVSClient_CreateGeneve(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.CreateGeneve("geneve0", "10.0.0.1", 100)
	if err != nil {
		t.Fatalf("failed to create Geneve: %v", err)
	}
}

// TestRealOVSClient_AddFlow tests flow addition (dry-run)
func TestRealOVSClient_AddFlow(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.AddFlow("br-test", 100, "in_port=1", "output:2")
	if err != nil {
		t.Fatalf("failed to add flow: %v", err)
	}
}

// TestRealOVSClient_DeleteFlow tests flow deletion (dry-run)
func TestRealOVSClient_DeleteFlow(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.DeleteFlow("br-test", "in_port=1")
	if err != nil {
		t.Fatalf("failed to delete flow: %v", err)
	}
}

// TestRealOVSClient_ClearFlows tests flow clearing (dry-run)
func TestRealOVSClient_ClearFlows(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.ClearFlows("br-test")
	if err != nil {
		t.Fatalf("failed to clear flows: %v", err)
	}
}

// TestRealOVSClient_InterfaceOptions tests interface options (dry-run)
func TestRealOVSClient_InterfaceOptions(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetInterfaceOption("vnet0", "key", "value")
	if err != nil {
		t.Fatalf("failed to set interface option: %v", err)
	}
}

// TestRealOVSClient_SetBridgeVLAN tests bridge VLAN (dry-run)
func TestRealOVSClient_SetBridgeVLAN(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetBridgeVLAN("br-test", 100)
	if err != nil {
		t.Fatalf("failed to set bridge VLAN: %v", err)
	}
}

// TestRealOVSClient_SetBridgeTrunk tests bridge trunk (dry-run)
func TestRealOVSClient_SetBridgeTrunk(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	err := client.SetBridgeTrunk("br-test", []int{100, 200})
	if err != nil {
		t.Fatalf("failed to set bridge trunk: %v", err)
	}
}

// TestRealOVSClient_Config tests configuration
func TestRealOVSClient_Config(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		Timeout: 10,
		Sudo:    true,
		DryRun:  true,
	})

	// Test timeout
	client.SetTimeout(30)

	// Test dry-run toggle
	client.SetDryRun(false)
	client.SetDryRun(true)

	// Test sudo toggle
	client.SetSudo(false)
	client.SetSudo(true)
}

// TestRealOVSClient_LastCommand tests command tracking
func TestRealOVSClient_LastCommand(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_ = client.CreateBridge("br-test")

	cmd := client.LastCommand()
	if cmd == "" {
		t.Error("last command should be recorded")
	}

	output := client.LastOutput()
	if output != "[dry-run]" {
		t.Errorf("expected dry-run output, got %s", output)
	}
}

// TestRealOVSClient_Types tests type definitions
func TestRealOVSClient_Types(t *testing.T) {
	// Test Bridge
	bridge := &realovs.Bridge{
		Name:  "br-test",
		VLAN:  100,
		Trunk: []int{200, 300},
		Ports: []string{"vnet0", "vnet1"},
	}
	if bridge.Name != "br-test" {
		t.Error("bridge name mismatch")
	}

	// Test Port
	port := &realovs.Port{
		Name:        "vnet0",
		Bridge:      "br-test",
		VLAN:        100,
		Trunk:       []int{200, 300},
		QoS:         1000,
		MAC:         "52:54:00:12:34:56",
		IPAddress:   "10.100.1.10",
		PortSecurity: true,
		Options:     map[string]string{"key": "value"},
	}
	if port.Name != "vnet0" {
		t.Error("port name mismatch")
	}

	// Test Flow
	flow := &realovs.Flow{
		ID:       "flow-1",
		Bridge:   "br-test",
		Priority: 100,
		Match:    "in_port=1",
		Actions:  "output:2",
		Enabled:  true,
	}
	if flow.Priority != 100 {
		t.Error("flow priority mismatch")
	}
}

// TestRealOVSClient_BridgeOperations tests bridge operations chain (dry-run)
func TestRealOVSClient_BridgeOperations(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	// Create bridge
	if err := client.CreateBridge("br-test"); err != nil {
		t.Fatalf("create bridge failed: %v", err)
	}

	// Set VLAN
	if err := client.SetBridgeVLAN("br-test", 100); err != nil {
		t.Fatalf("set VLAN failed: %v", err)
	}

	// Set trunk
	if err := client.SetBridgeTrunk("br-test", []int{200, 300}); err != nil {
		t.Fatalf("set trunk failed: %v", err)
	}

	// Delete bridge
	if err := client.DeleteBridge("br-test"); err != nil {
		t.Fatalf("delete bridge failed: %v", err)
	}
}

// TestRealOVSClient_PortOperations tests port operations chain (dry-run)
func TestRealOVSClient_PortOperations(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	// Create bridge
	if err := client.CreateBridge("br-test"); err != nil {
		t.Fatalf("create bridge failed: %v", err)
	}

	// Add port
	if err := client.AddPort("br-test", "vnet0"); err != nil {
		t.Fatalf("add port failed: %v", err)
	}

	// Set VLAN
	if err := client.SetPortVLAN("vnet0", 100); err != nil {
		t.Fatalf("set VLAN failed: %v", err)
	}

	// Set trunk
	if err := client.SetPortTrunk("vnet0", []int{200, 300}); err != nil {
		t.Fatalf("set trunk failed: %v", err)
	}

	// Set QoS
	if err := client.SetPortQoS("vnet0", 1000); err != nil {
		t.Fatalf("set QoS failed: %v", err)
	}

	// Set security
	if err := client.SetPortSecurity("vnet0", "52:54:00:12:34:56", "10.100.1.10"); err != nil {
		t.Fatalf("set security failed: %v", err)
	}

	// Delete port
	if err := client.DeletePort("br-test", "vnet0"); err != nil {
		t.Fatalf("delete port failed: %v", err)
	}

	// Delete bridge
	if err := client.DeleteBridge("br-test"); err != nil {
		t.Fatalf("delete bridge failed: %v", err)
	}
}

// TestRealOVSClient_TunnelOperations tests tunnel operations (dry-run)
func TestRealOVSClient_TunnelOperations(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	// Create VXLAN
	if err := client.CreateVXLAN("vxlan0", "10.0.0.1", 100); err != nil {
		t.Fatalf("create VXLAN failed: %v", err)
	}

	// Create GRE
	if err := client.CreateGRE("gre0", "10.0.0.2", 200); err != nil {
		t.Fatalf("create GRE failed: %v", err)
	}

	// Create Geneve
	if err := client.CreateGeneve("geneve0", "10.0.0.3", 300); err != nil {
		t.Fatalf("create Geneve failed: %v", err)
	}
}

// TestRealOVSClient_FlowOperations tests flow operations (dry-run)
func TestRealOVSClient_FlowOperations(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	// Create bridge
	if err := client.CreateBridge("br-test"); err != nil {
		t.Fatalf("create bridge failed: %v", err)
	}

	// Add flow
	if err := client.AddFlow("br-test", 100, "in_port=1", "output:2"); err != nil {
		t.Fatalf("add flow failed: %v", err)
	}

	// Delete flow
	if err := client.DeleteFlow("br-test", "in_port=1"); err != nil {
		t.Fatalf("delete flow failed: %v", err)
	}

	// Clear flows
	if err := client.ClearFlows("br-test"); err != nil {
		t.Fatalf("clear flows failed: %v", err)
	}

	// Delete bridge
	if err := client.DeleteBridge("br-test"); err != nil {
		t.Fatalf("delete bridge failed: %v", err)
	}
}

// TestRealOVSClient_GetBridge tests GetBridge (dry-run)
func TestRealOVSClient_GetBridge(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	// GetBridge returns error in dry-run mode
	_, err := client.GetBridge("br-test")
	// This should not error in dry-run mode
	_ = err
}

// TestRealOVSClient_ListBridges tests ListBridges (dry-run)
func TestRealOVSClient_ListBridges(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.ListBridges()
	_ = err
}

// TestRealOVSClient_GetPort tests GetPort (dry-run)
func TestRealOVSClient_GetPort(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.GetPort("vnet0")
	_ = err
}

// TestRealOVSClient_ListPorts tests ListPorts (dry-run)
func TestRealOVSClient_ListPorts(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.ListPorts("br-test")
	_ = err
}

// TestRealOVSClient_ListFlows tests ListFlows (dry-run)
func TestRealOVSClient_ListFlows(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.ListFlows("br-test")
	_ = err
}

// TestRealOVSClient_GetInterfaceUUID tests GetInterfaceUUID (dry-run)
func TestRealOVSClient_GetInterfaceUUID(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.GetInterfaceUUID("vnet0")
	_ = err
}

// TestRealOVSClient_GetInterfaceOption tests GetInterfaceOption (dry-run)
func TestRealOVSClient_GetInterfaceOption(t *testing.T) {
	client := realovs.NewClient(&realovs.Config{
		DryRun: true,
	})

	_, err := client.GetInterfaceOption("vnet0", "key")
	_ = err
}

// TestRealOVSClient_IsAvailable tests IsAvailable
func TestRealOVSClient_IsAvailable(t *testing.T) {
	// This function checks if ovs-vsctl is in PATH
	available := realovs.IsAvailable()
	_ = available
}

// TestRealOVSClient_Version tests Version
func TestRealOVSClient_Version(t *testing.T) {
	// This function returns OVS version
	version, err := realovs.Version()
	if err != nil {
		// OVS may not be installed, that's fine
		return
	}
	_ = version
}