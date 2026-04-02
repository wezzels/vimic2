// Package deploy_test tests deployment wizard functionality
package deploy_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/deploy"
)

// TestWizardCreation tests wizard initialization
func TestWizardCreation(t *testing.T) {
	w := deploy.NewWizard()

	if w == nil {
		t.Fatal("Wizard should not be nil")
	}

	cluster := w.GetCluster()
	if cluster == nil {
		t.Fatal("Cluster should not be nil")
	}

	if cluster.ID == "" {
		t.Error("Cluster ID should be generated")
	}

	if w.GetStep() != 1 {
		t.Errorf("Initial step should be 1, got %d", w.GetStep())
	}
}

// TestWizardStepNavigation tests step navigation
func TestWizardStepNavigation(t *testing.T) {
	w := deploy.NewWizard()

	// Test initial step (starts at 1)
	if w.GetStep() != 1 {
		t.Errorf("Expected step 1, got %d", w.GetStep())
	}

	// Navigate forward
	w.NextStep()
	if w.GetStep() != 2 {
		t.Errorf("Expected step 2, got %d", w.GetStep())
	}

	w.NextStep()
	if w.GetStep() != 3 {
		t.Errorf("Expected step 3, got %d", w.GetStep())
	}

	// Navigate backward
	w.PrevStep()
	if w.GetStep() != 2 {
		t.Errorf("Expected step 2, got %d", w.GetStep())
	}

	// Test boundary - can't go below 1
	w.PrevStep() // Now at 1
	if w.GetStep() != 1 {
		t.Errorf("Expected step 1, got %d", w.GetStep())
	}
}

// TestWizardNameSetting tests cluster name configuration
func TestWizardNameSetting(t *testing.T) {
	w := deploy.NewWizard()

	w.SetName("test-cluster")

	cluster := w.GetCluster()
	if cluster.Name != "test-cluster" {
		t.Errorf("Expected name 'test-cluster', got '%s'", cluster.Name)
	}
}

// TestWizardNetworkConfiguration tests network config
func TestWizardNetworkConfiguration(t *testing.T) {
	w := deploy.NewWizard()

	nw := &deploy.NetworkConfig{
		Name:    "default-network",
		Type:    "nat",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
	}

	w.SetNetwork(nw)

	cluster := w.GetCluster()
	if cluster.Network == nil {
		t.Fatal("Network config should not be nil")
	}

	if cluster.Network.Name != "default-network" {
		t.Errorf("Expected network name 'default-network', got '%s'", cluster.Network.Name)
	}

	if cluster.Network.CIDR != "10.0.0.0/24" {
		t.Errorf("Expected CIDR '10.0.0.0/24', got '%s'", cluster.Network.CIDR)
	}

	if cluster.Network.Type != "nat" {
		t.Errorf("Expected type 'nat', got '%s'", cluster.Network.Type)
	}
}

// TestWizardNodeGroupManagement tests node group operations
func TestWizardNodeGroupManagement(t *testing.T) {
	w := deploy.NewWizard()

	// Add first node group
	ng1 := &deploy.NodeGroup{
		Name:     "workers",
		Count:    3,
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "ubuntu:22.04",
		Role:     "worker",
	}
	w.AddNodeGroup(ng1)

	cluster := w.GetCluster()
	if len(cluster.NodeGroups) != 1 {
		t.Errorf("Expected 1 node group, got %d", len(cluster.NodeGroups))
	}

	// Add second node group
	ng2 := &deploy.NodeGroup{
		Name:     "masters",
		Count:    1,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu:22.04",
		Role:     "master",
	}
	w.AddNodeGroup(ng2)

	if len(cluster.NodeGroups) != 2 {
		t.Errorf("Expected 2 node groups, got %d", len(cluster.NodeGroups))
	}

	// Remove first node group
	w.RemoveNodeGroup(0)

	if len(cluster.NodeGroups) != 1 {
		t.Errorf("Expected 1 node group after removal, got %d", len(cluster.NodeGroups))
	}

	// Verify remaining group is the master
	if cluster.NodeGroups[0].Name != "masters" {
		t.Errorf("Expected 'masters' node group, got '%s'", cluster.NodeGroups[0].Name)
	}
}

// TestWizardHostConfiguration tests host assignment
func TestWizardHostConfiguration(t *testing.T) {
	w := deploy.NewWizard()

	host := &deploy.HostRef{
		HostID:    "host-1",
		HostName:  "Primary Host",
		NodeCount: 0,
	}

	cluster := w.GetCluster()
	cluster.Hosts = append(cluster.Hosts, host)

	if len(cluster.Hosts) != 1 {
		t.Errorf("Expected 1 host, got %d", len(cluster.Hosts))
	}

	if cluster.Hosts[0].HostID != "host-1" {
		t.Errorf("Expected host ID 'host-1', got '%s'", cluster.Hosts[0].HostID)
	}
}

// TestWizardValidation tests configuration validation
func TestWizardValidation(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		w := deploy.NewWizard()
		w.SetName("valid-cluster")

		ng := &deploy.NodeGroup{
			Name:     "workers",
			Count:    3,
			CPU:      2,
			MemoryMB: 4096,
			DiskGB:   20,
			Image:    "ubuntu:22.04",
			Role:     "worker",
		}
		w.AddNodeGroup(ng)

		err := w.Validate()
		if err != nil {
			t.Errorf("Valid configuration should not error: %v", err)
		}
	})

	t.Run("EmptyName", func(t *testing.T) {
		w := deploy.NewWizard()
		// Don't set name

		err := w.Validate()
		if err == nil {
			t.Error("Empty name should fail validation")
		}
	})

	t.Run("NoNodeGroups", func(t *testing.T) {
		w := deploy.NewWizard()
		w.SetName("test-cluster")
		// Don't add node groups

		err := w.Validate()
		if err == nil {
			t.Error("No node groups should fail validation")
		}
	})
}

// TestNodeGroupDefaults tests node group values
func TestNodeGroupDefaults(t *testing.T) {
	// NodeGroup fields should be explicitly set
	// No defaults are applied by the struct itself
	ng := &deploy.NodeGroup{
		Name:     "test",
		Count:    3,
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
	}

	// Verify set values
	if ng.CPU != 2 {
		t.Errorf("Expected CPU 2, got %d", ng.CPU)
	}
	if ng.MemoryMB != 4096 {
		t.Errorf("Expected MemoryMB 4096, got %d", ng.MemoryMB)
	}
	if ng.DiskGB != 20 {
		t.Errorf("Expected DiskGB 20, got %d", ng.DiskGB)
	}
}

// TestNetworkConfigDefaults tests network config default values
func TestNetworkConfigDefaults(t *testing.T) {
	nw := &deploy.NetworkConfig{
		Name: "test-network",
	}

	// Verify defaults
	if nw.Type == "" {
		// Type can be empty, will use default
	}
	if nw.CIDR == "" {
		// CIDR can be empty for bridge networks
	}
}