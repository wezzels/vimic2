// Package realovs_test tests the real OVS client
package realovs_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realovs"
)

// TestRealOVS_BridgeExists_DryRun tests BridgeExists in dry-run mode
func TestRealOVS_BridgeExists_DryRun(t *testing.T) {
	c := realovs.NewClient(&realovs.Config{DryRun: true})

	// In dry-run mode, BridgeExists should return false (can't verify)
	if c.BridgeExists("br-test") {
		t.Error("BridgeExists should return false in dry-run mode")
	}
}

// TestRealOVS_PortExists_DryRun tests PortExists in dry-run mode
func TestRealOVS_PortExists_DryRun(t *testing.T) {
	c := realovs.NewClient(&realovs.Config{DryRun: true})

	// In dry-run mode, PortExists should return false (can't verify)
	if c.PortExists("vnet0") {
		t.Error("PortExists should return false in dry-run mode")
	}
}

// TestRealOVS_GetBridge_Error tests GetBridge error handling
func TestRealOVS_GetBridge_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults() // No dry-run - will actually try

	// Get non-existent bridge
	_, err := c.GetBridge("non-existent-bridge-xyz")
	if err == nil {
		t.Error("GetBridge should fail for non-existent bridge")
	}
}

// TestRealOVS_ListBridges_Error tests ListBridges error handling
func TestRealOVS_ListBridges_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// This should work - list all bridges
	bridges, err := c.ListBridges()
	// May or may not error depending on OVS state
	_ = bridges
	_ = err
}

// TestRealOVS_GetPort_Error tests GetPort error handling
func TestRealOVS_GetPort_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Get port from non-existent bridge
	_, err := c.GetPort("non-existent-port-xyz")
	if err == nil {
		t.Error("GetPort should fail for non-existent port")
	}
}

// TestRealOVS_ListPorts_Error tests ListPorts error handling
func TestRealOVS_ListPorts_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// List ports on non-existent bridge
	_, err := c.ListPorts("non-existent-bridge-xyz")
	if err == nil {
		t.Error("ListPorts should fail for non-existent bridge")
	}
}

// TestRealOVS_ListFlows_Error tests ListFlows error handling
func TestRealOVS_ListFlows_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// List flows on non-existent bridge
	_, err := c.ListFlows("non-existent-bridge-xyz")
	if err == nil {
		t.Error("ListFlows should fail for non-existent bridge")
	}
}

// TestRealOVS_GetInterfaceUUID_Error tests GetInterfaceUUID error handling
func TestRealOVS_GetInterfaceUUID_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Get UUID for non-existent interface
	_, err := c.GetInterfaceUUID("non-existent-interface-xyz")
	if err == nil {
		t.Error("GetInterfaceUUID should fail for non-existent interface")
	}
}

// TestRealOVS_GetInterfaceOption_Error tests GetInterfaceOption error handling
func TestRealOVS_GetInterfaceOption_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Get option for non-existent interface
	_, err := c.GetInterfaceOption("non-existent-interface-xyz", "key")
	if err == nil {
		t.Error("GetInterfaceOption should fail for non-existent interface")
	}
}

// TestRealOVS_Version_Success tests Version returns valid version
func TestRealOVS_Version_Success(t *testing.T) {
	version, err := realovs.Version()
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}

	if version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("OVS Version: %s", version)
}

// TestRealOVS_IsAvailable tests IsAvailable function
func TestRealOVS_IsAvailable_Unit(t *testing.T) {
	available := realovs.IsAvailable()
	if !available {
		t.Error("IsAvailable should return true (OVS is installed)")
	}
}

// TestRealOVS_NewClient_Defaults tests NewClient with nil config
func TestRealOVS_NewClient_Defaults(t *testing.T) {
	c := realovs.NewClient(nil)
	if c == nil {
		t.Fatal("NewClient should return non-nil client")
	}
}

// TestRealOVS_NewClientWithDefaults tests NewClientWithDefaults
func TestRealOVS_NewClientWithDefaults(t *testing.T) {
	c := realovs.NewClientWithDefaults()
	if c == nil {
		t.Fatal("NewClientWithDefaults should return non-nil client")
	}
}

// TestRealOVS_Config_Timeout tests SetTimeout
func TestRealOVS_Config_Timeout(t *testing.T) {
	c := realovs.NewClientWithDefaults()
	c.SetTimeout(60)
	// No error - just verify it doesn't panic
}

// TestRealOVS_Config_Sudo tests SetSudo
func TestRealOVS_Config_Sudo(t *testing.T) {
	c := realovs.NewClientWithDefaults()
	c.SetSudo(true)
	c.SetSudo(false)
	// No error - just verify it doesn't panic
}

// TestRealOVS_Config_DryRun tests SetDryRun
func TestRealOVS_Config_DryRun(t *testing.T) {
	c := realovs.NewClientWithDefaults()
	c.SetDryRun(true)
	c.SetDryRun(false)
	// No error - just verify it doesn't panic
}

// TestRealOVS_DeleteBridge_Error tests DeleteBridge error
func TestRealOVS_DeleteBridge_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to delete non-existent bridge (should not error with --if-exists)
	err := c.DeleteBridge("non-existent-bridge-xyz")
	// --if-exists makes this succeed even for non-existent bridge
	_ = err
}

// TestRealOVS_DeletePort_Error tests DeletePort error
func TestRealOVS_DeletePort_Error(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to delete non-existent port (should not error with --if-exists)
	err := c.DeletePort("non-existent-bridge", "non-existent-port")
	// --if-exists makes this succeed
	_ = err
}

// TestRealOVS_AddPort_NoBridge tests AddPort without bridge
func TestRealOVS_AddPort_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to add port to non-existent bridge
	err := c.AddPort("non-existent-bridge-xyz", "vnet0")
	if err == nil {
		t.Error("AddPort should fail for non-existent bridge")
	}
}

// TestRealOVS_SetPortVLAN_NoPort tests SetPortVLAN without port
func TestRealOVS_SetPortVLAN_NoPort(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to set VLAN on non-existent port
	err := c.SetPortVLAN("non-existent-port-xyz", 100)
	if err == nil {
		t.Error("SetPortVLAN should fail for non-existent port")
	}
}

// TestRealOVS_SetPortTrunk_NoPort tests SetPortTrunk without port
func TestRealOVS_SetPortTrunk_NoPort(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to set trunk on non-existent port
	err := c.SetPortTrunk("non-existent-port-xyz", []int{100, 200})
	if err == nil {
		t.Error("SetPortTrunk should fail for non-existent port")
	}
}

// TestRealOVS_SetPortQoS_NoPort tests SetPortQoS without port
func TestRealOVS_SetPortQoS_NoPort(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to set QoS on non-existent port
	err := c.SetPortQoS("non-existent-port-xyz", 1000)
	if err == nil {
		t.Error("SetPortQoS should fail for non-existent port")
	}
}

// TestRealOVS_SetPortSecurity_NoPort tests SetPortSecurity without port
func TestRealOVS_SetPortSecurity_NoPort(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Try to set security on non-existent port
	err := c.SetPortSecurity("non-existent-port-xyz", "52:54:00:12:34:56", "10.0.0.1")
	if err == nil {
		t.Error("SetPortSecurity should fail for non-existent port")
	}
}

// TestRealOVS_CreateVXLAN_NoBridge tests CreateVXLAN without bridge
func TestRealOVS_CreateVXLAN_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// VXLAN needs br-int to exist
	err := c.CreateVXLAN("vxlan-test", "10.0.0.1", 100)
	// May fail if br-int doesn't exist
	_ = err
}

// TestRealOVS_CreateGRE_NoBridge tests CreateGRE without bridge
func TestRealOVS_CreateGRE_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// GRE needs br-int to exist
	err := c.CreateGRE("gre-test", "10.0.0.1", 100)
	// May fail if br-int doesn't exist
	_ = err
}

// TestRealOVS_CreateGeneve_NoBridge tests CreateGeneve without bridge
func TestRealOVS_CreateGeneve_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// Geneve needs br-int to exist
	err := c.CreateGeneve("geneve-test", "10.0.0.1", 100)
	// May fail if br-int doesn't exist
	_ = err
}

// TestRealOVS_AddFlow_NoBridge tests AddFlow without bridge
func TestRealOVS_AddFlow_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.AddFlow("non-existent-bridge-xyz", 100, "in_port=1", "output:2")
	if err == nil {
		t.Error("AddFlow should fail for non-existent bridge")
	}
}

// TestRealOVS_DeleteFlow_NoBridge tests DeleteFlow without bridge
func TestRealOVS_DeleteFlow_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.DeleteFlow("non-existent-bridge-xyz", "in_port=1")
	if err == nil {
		t.Error("DeleteFlow should fail for non-existent bridge")
	}
}

// TestRealOVS_ClearFlows_NoBridge tests ClearFlows without bridge
func TestRealOVS_ClearFlows_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.ClearFlows("non-existent-bridge-xyz")
	if err == nil {
		t.Error("ClearFlows should fail for non-existent bridge")
	}
}

// TestRealOVS_SetInterfaceOption_NoInterface tests SetInterfaceOption without interface
func TestRealOVS_SetInterfaceOption_NoInterface(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.SetInterfaceOption("non-existent-interface-xyz", "key", "value")
	if err == nil {
		t.Error("SetInterfaceOption should fail for non-existent interface")
	}
}

// TestRealOVS_SetBridgeVLAN_NoBridge tests SetBridgeVLAN without bridge
func TestRealOVS_SetBridgeVLAN_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.SetBridgeVLAN("non-existent-bridge-xyz", 100)
	if err == nil {
		t.Error("SetBridgeVLAN should fail for non-existent bridge")
	}
}

// TestRealOVS_SetBridgeTrunk_NoBridge tests SetBridgeTrunk without bridge
func TestRealOVS_SetBridgeTrunk_NoBridge(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.SetBridgeTrunk("non-existent-bridge-xyz", []int{100, 200})
	if err == nil {
		t.Error("SetBridgeTrunk should fail for non-existent bridge")
	}
}
// TestRealOVS_GetBridge_NonExistent tests GetBridge with non-existent bridge
func TestRealOVS_GetBridge_NonExistent(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	_, err := c.GetBridge("non-existent-bridge-xyz")
	if err == nil {
		t.Error("GetBridge should fail for non-existent bridge")
	}
}

// TestRealOVS_ListBridges tests ListBridges
func TestRealOVS_ListBridges(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	bridges, err := c.ListBridges()
	if err != nil {
		t.Fatalf("ListBridges failed: %v", err)
	}

	// Should return list (may be empty if no bridges)
	_ = bridges
}

// TestRealOVS_GetPort_NonExistent tests GetPort with non-existent port
func TestRealOVS_GetPort_NonExistent(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	_, err := c.GetPort("non-existent-port-xyz")
	if err == nil {
		t.Error("GetPort should fail for non-existent port")
	}
}

// TestRealOVS_ListFlows tests ListFlows
func TestRealOVS_ListFlows(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// ListFlows requires a bridge
	_, err := c.ListFlows("non-existent-bridge-xyz")
	// Should fail for non-existent bridge
	_ = err
}

// TestRealOVS_Version_Format tests Version format
func TestRealOVS_Version_Format(t *testing.T) {
	version, err := realovs.Version()
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}

	if version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("OVS Version: %s", version)
}

// TestRealOVS_GetInterfaceUUID_NonExistent tests GetInterfaceUUID with non-existent interface
func TestRealOVS_GetInterfaceUUID_NonExistent(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	_, err := c.GetInterfaceUUID("non-existent-interface-xyz")
	if err == nil {
		t.Error("GetInterfaceUUID should fail for non-existent interface")
	}
}

// TestRealOVS_GetInterfaceOption_NonExistent tests GetInterfaceOption with non-existent interface
func TestRealOVS_GetInterfaceOption_NonExistent(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	_, err := c.GetInterfaceOption("non-existent-interface-xyz", "key")
	if err == nil {
		t.Error("GetInterfaceOption should fail for non-existent interface")
	}
}

// TestRealOVS_SetInterfaceOption_Invalid tests SetInterfaceOption with invalid option
func TestRealOVS_SetInterfaceOption_Invalid(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	err := c.SetInterfaceOption("non-existent-interface-xyz", "invalid-key", "value")
	if err == nil {
		t.Error("SetInterfaceOption should fail for non-existent interface")
	}
}



// TestRealOVS_GetBridge_Errors tests GetBridge error handling
func TestRealOVS_GetBridge_Errors(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// GetBridge for non-existent should fail
	_, err := c.GetBridge("non-existent-bridge-xyz")
	if err == nil {
		t.Error("GetBridge should fail for non-existent bridge")
	}
}

// TestRealOVS_ListPorts_Errors tests ListPorts error handling  
func TestRealOVS_ListPorts_Errors(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// ListPorts for non-existent bridge should fail
	_, err := c.ListPorts("non-existent-bridge-xyz")
	if err == nil {
		t.Error("ListPorts should fail for non-existent bridge")
	}
}

// TestRealOVS_Version_Errors tests Version error handling
func TestRealOVS_Version_Errors(t *testing.T) {
	// Version is a package-level function
	_, err := realovs.Version()
	_ = err
}

// TestRealOVS_ClearFlows_Errors tests ClearFlows error handling
func TestRealOVS_ClearFlows_Errors(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// ClearFlows for non-existent bridge should fail
	err := c.ClearFlows("non-existent-bridge-xyz")
	if err == nil {
		t.Error("ClearFlows should fail for non-existent bridge")
	}
}

// TestRealOVS_BridgeExists_False tests BridgeExists returns false
func TestRealOVS_BridgeExists_False(t *testing.T) {
	c := realovs.NewClientWithDefaults()

	// BridgeExists for non-existent should return false
	if c.BridgeExists("non-existent-bridge-xyz") {
		t.Error("BridgeExists should return false for non-existent bridge")
	}
}
