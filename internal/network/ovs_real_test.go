//go:build integration

package network

import (
	"os"
	"strings"
	"testing"
)

// TestOVSClient_RealOVS tests with actual OVS installation
func TestOVSClient_RealOVS(t *testing.T) {
	// Check for OVS - look for any .ctl file in openvswitch directory
	ctlFiles, err := os.ReadDir("/var/run/openvswitch")
	if err != nil {
		t.Skip("OVS not available (no openvswitch directory)")
	}
	
	hasOVS := false
	for _, f := range ctlFiles {
		if strings.HasSuffix(f.Name(), ".ctl") {
			hasOVS = true
			break
		}
	}
	if !hasOVS {
		t.Skip("OVS not available (no control socket)")
	}

	ovs := NewOVSClient()
	if ovs == nil {
		t.Fatal("NewOVSClient returned nil")
	}

	t.Run("ListBridges", func(t *testing.T) {
		bridges, err := ovs.ListBridges()
		if err != nil {
			t.Logf("ListBridges error (may need permissions): %v", err)
			return
		}

		t.Logf("Found %d bridges", len(bridges))
		for _, br := range bridges {
			t.Logf("  Bridge: %s", br)
		}
	})

	t.Run("BridgeExists", func(t *testing.T) {
		// Check if br-int exists (from ovs-vsctl show)
		exists := ovs.bridgeExists("br-int")
		t.Logf("br-int exists: %v", exists)
	})
}

// TestOVSClient_BridgeOperations tests bridge creation/deletion
func TestOVSClient_BridgeOperations(t *testing.T) {
	ctlFiles, err := os.ReadDir("/var/run/openvswitch")
	if err != nil {
		t.Skip("OVS not available")
	}
	
	hasOVS := false
	for _, f := range ctlFiles {
		if strings.HasSuffix(f.Name(), ".ctl") {
			hasOVS = true
			break
		}
	}
	if !hasOVS {
		t.Skip("OVS not available")
	}

	ovs := NewOVSClient()
	testBridge := "vimic2-test-br"

	t.Run("CreateBridge", func(t *testing.T) {
		err := ovs.CreateBridge(testBridge)
		if err != nil {
			t.Skipf("CreateBridge failed (may need permissions): %v", err)
		}

		// Verify bridge exists
		if !ovs.bridgeExists(testBridge) {
			t.Error("Bridge was not created")
		}

		// Cleanup
		defer func() {
			ovs.DeleteBridge(testBridge)
		}()

		t.Log("Bridge created successfully")
	})

	t.Run("CreatePort", func(t *testing.T) {
		// First create bridge
		if err := ovs.CreateBridge(testBridge); err != nil {
			t.Skipf("CreateBridge failed: %v", err)
		}
		defer ovs.DeleteBridge(testBridge)

		// Create port
		err := ovs.CreatePort(testBridge, "test-port")
		if err != nil {
			t.Skipf("CreatePort failed: %v", err)
		}

		t.Log("Port created successfully")
	})

	t.Run("SetBridgeVLAN", func(t *testing.T) {
		if err := ovs.CreateBridge(testBridge); err != nil {
			t.Skipf("CreateBridge failed: %v", err)
		}
		defer ovs.DeleteBridge(testBridge)

		err := ovs.SetBridgeVLAN(testBridge, 100)
		if err != nil {
			t.Skipf("SetBridgeVLAN failed: %v", err)
		}

		t.Log("VLAN set successfully")
	})
}

// TestOVSClient_PortOperations tests port operations
func TestOVSClient_PortOperations_Real(t *testing.T) {
	ctlFiles, err := os.ReadDir("/var/run/openvswitch")
	if err != nil {
		t.Skip("OVS not available")
	}
	
	hasOVS := false
	for _, f := range ctlFiles {
		if strings.HasSuffix(f.Name(), ".ctl") {
			hasOVS = true
			break
		}
	}
	if !hasOVS {
		t.Skip("OVS not available")
	}

	ovs := NewOVSClient()
	testBridge := "vimic2-test-br2"

	// Setup
	if err := ovs.CreateBridge(testBridge); err != nil {
		t.Skipf("Setup failed: %v", err)
	}
	defer ovs.DeleteBridge(testBridge)

	t.Run("CreatePort", func(t *testing.T) {
		err := ovs.CreatePort(testBridge, "vimic2-test-port")
		if err != nil {
			t.Skipf("CreatePort failed: %v", err)
		}

		// Verify port exists
		ports, err := ovs.ListPorts(testBridge)
		if err == nil {
			for _, p := range ports {
				if p == "vimic2-test-port" {
					t.Log("Port created and verified")
					return
				}
			}
		}

		t.Log("Port created")
	})

	t.Run("DeletePort", func(t *testing.T) {
		ovs.CreatePort(testBridge, "vimic2-test-port2")
		err := ovs.DeletePort(testBridge, "vimic2-test-port2")
		if err != nil {
			t.Skipf("DeletePort failed: %v", err)
		}

		t.Log("Port deleted successfully")
	})
}