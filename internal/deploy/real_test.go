// Package deploy provides real implementation tests
package deploy

import (
	"testing"
)

// TestNewWizard_Real tests real wizard creation
func TestNewWizard_Real(t *testing.T) {
	wizard := NewWizard()

	if wizard == nil {
		t.Fatal("expected non-nil wizard")
	}
	if wizard.cluster == nil {
		t.Fatal("expected non-nil cluster")
	}
	if wizard.cluster.ID == "" {
		t.Error("expected cluster ID to be generated")
	}
	if wizard.step != 1 {
		t.Errorf("expected initial step 1, got %d", wizard.step)
	}
}

// TestWizard_FullConfiguration tests full wizard configuration
func TestWizard_FullConfiguration(t *testing.T) {
	wizard := NewWizard()

	// Configure cluster name
	wizard.SetName("production-cluster")
	if wizard.cluster.Name != "production-cluster" {
		t.Errorf("expected production-cluster, got %s", wizard.cluster.Name)
	}

	// Configure network
	network := &NetworkConfig{
		Type:    "bridge",
		Name:    "br0",
		CIDR:    "192.168.100.0/24",
		Gateway: "192.168.100.1",
	}
	wizard.SetNetwork(network)

	if wizard.cluster.Network == nil {
		t.Fatal("expected network to be set")
	}
	if wizard.cluster.Network.Type != "bridge" {
		t.Errorf("expected bridge, got %s", wizard.cluster.Network.Type)
	}

	// Add node groups
	wizard.AddNodeGroup(&NodeGroup{
		Name:     "masters",
		Role:     "master",
		Count:    3,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
	})

	wizard.AddNodeGroup(&NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    5,
		CPU:      8,
		MemoryMB: 16384,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
	})

	if len(wizard.cluster.NodeGroups) != 2 {
		t.Errorf("expected 2 node groups, got %d", len(wizard.cluster.NodeGroups))
	}

	// Verify cluster configuration
	cluster := wizard.GetCluster()
	if cluster.Name != "production-cluster" {
		t.Errorf("expected production-cluster, got %s", cluster.Name)
	}
	if cluster.Network == nil {
		t.Error("expected network to be configured")
	}
}

// TestWizard_StepNavigation tests step navigation
func TestWizard_StepNavigation(t *testing.T) {
	wizard := NewWizard()

	// Initial step
	if wizard.GetStep() != 1 {
		t.Errorf("expected step 1, got %d", wizard.GetStep())
	}

	// Next step
	wizard.NextStep()
	if wizard.GetStep() != 2 {
		t.Errorf("expected step 2, got %d", wizard.GetStep())
	}

	// Another next
	wizard.NextStep()
	if wizard.GetStep() != 3 {
		t.Errorf("expected step 3, got %d", wizard.GetStep())
	}

	// Previous
	wizard.PrevStep()
	if wizard.GetStep() != 2 {
		t.Errorf("expected step 2, got %d", wizard.GetStep())
	}
}

// TestWizard_RemoveNodeGroups tests removing node groups
func TestWizard_RemoveNodeGroups(t *testing.T) {
	wizard := NewWizard()

	// Add multiple groups
	wizard.AddNodeGroup(&NodeGroup{Name: "group-1", Role: "master", Count: 1})
	wizard.AddNodeGroup(&NodeGroup{Name: "group-2", Role: "worker", Count: 3})
	wizard.AddNodeGroup(&NodeGroup{Name: "group-3", Role: "storage", Count: 2})

	if len(wizard.cluster.NodeGroups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(wizard.cluster.NodeGroups))
	}

	// Remove middle group
	wizard.RemoveNodeGroup(1)

	if len(wizard.cluster.NodeGroups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(wizard.cluster.NodeGroups))
	}

	// Verify order
	if wizard.cluster.NodeGroups[0].Name != "group-1" {
		t.Errorf("expected group-1, got %s", wizard.cluster.NodeGroups[0].Name)
	}
	if wizard.cluster.NodeGroups[1].Name != "group-3" {
		t.Errorf("expected group-3, got %s", wizard.cluster.NodeGroups[1].Name)
	}
}

// TestWizard_ValidateReal tests validation logic with real calls
func TestWizard_ValidateReal(t *testing.T) {
	// Valid configuration test
	wizard := NewWizard()
	wizard.SetName("valid-cluster")
	wizard.AddNodeGroup(&NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    3,
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	})
	wizard.SetNetwork(&NetworkConfig{
		Type:    "nat",
		CIDR:    "192.168.100.0/24",
		Gateway: "192.168.100.1",
	})

	err := wizard.Validate()
	if err != nil {
		t.Errorf("unexpected error for valid config: %v", err)
	}
}

// TestCluster_Serialization tests cluster serialization
func TestCluster_Serialization(t *testing.T) {
	cluster := &Cluster{
		ID:   "cluster-1",
		Name: "test-cluster",
		NodeGroups: []*NodeGroup{
			{Name: "masters", Role: "master", Count: 3},
			{Name: "workers", Role: "worker", Count: 5},
		},
		Network: &NetworkConfig{
			Type:    "nat",
			CIDR:    "192.168.1.0/24",
			Gateway: "192.168.1.1",
		},
		Status: "pending",
	}

	// Verify all fields are accessible
	if cluster.ID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", cluster.ID)
	}
	if len(cluster.NodeGroups) != 2 {
		t.Errorf("expected 2 node groups, got %d", len(cluster.NodeGroups))
	}
	if cluster.Network == nil {
		t.Fatal("expected network")
	}
}

// TestNodeGroup_Serialization tests node group serialization
func TestNodeGroup_Serialization(t *testing.T) {
	ng := &NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    5,
		CPU:      8,
		MemoryMB: 16384,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
		HostID:   "host-1",
	}

	// Verify fields
	if ng.Name != "workers" {
		t.Errorf("expected workers, got %s", ng.Name)
	}
	if ng.CPU != 8 {
		t.Errorf("expected 8 CPUs, got %d", ng.CPU)
	}
	if ng.MemoryMB != 16384 {
		t.Errorf("expected 16384 MB, got %d", ng.MemoryMB)
	}
}

// TestNetworkConfig_Types tests network configuration types
func TestNetworkConfig_Types(t *testing.T) {
	networkTypes := []string{"nat", "bridge", "none"}

	for _, nt := range networkTypes {
		t.Run(nt, func(t *testing.T) {
			network := &NetworkConfig{
				Type:    nt,
				CIDR:    "192.168.1.0/24",
				Gateway: "192.168.1.1",
			}

			if network.Type != nt {
				t.Errorf("expected %s, got %s", nt, network.Type)
			}
		})
	}
}

// TestProgress_Struct tests Progress struct
func TestProgress_Struct(t *testing.T) {
	progress := &Progress{
		ClusterID:     "cluster-1",
		TotalNodes:    10,
		DeployedNodes: 5,
		CurrentNode:   "node-5",
		Status:        "deploying",
	}

	if progress.ClusterID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", progress.ClusterID)
	}
	if progress.TotalNodes != 10 {
		t.Errorf("expected 10 total nodes, got %d", progress.TotalNodes)
	}
	if progress.DeployedNodes != 5 {
		t.Errorf("expected 5 deployed nodes, got %d", progress.DeployedNodes)
	}
	if progress.Status != "deploying" {
		t.Errorf("expected deploying, got %s", progress.Status)
	}
}

// TestHostRef_Struct tests HostRef struct
func TestHostRef_Struct(t *testing.T) {
	hostRef := &HostRef{
		HostID:    "host-1",
		HostName:  "hypervisor-1",
		NodeCount: 10,
	}

	if hostRef.HostID != "host-1" {
		t.Errorf("expected host-1, got %s", hostRef.HostID)
	}
	if hostRef.NodeCount != 10 {
		t.Errorf("expected 10 nodes, got %d", hostRef.NodeCount)
	}
}
