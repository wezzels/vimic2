// Package deploy provides tests for deployment wizard
package deploy

import (
	"testing"
	"time"
)

// TestWizard_Create tests wizard creation
func TestWizard_Create(t *testing.T) {
	wizard := NewWizard()

	if wizard == nil {
		t.Fatal("wizard should not be nil")
	}
	if wizard.cluster == nil {
		t.Error("cluster should not be nil")
	}
	if wizard.cluster.ID == "" {
		t.Error("cluster ID should be generated")
	}
}

// TestCluster tests cluster structure
func TestCluster_Create(t *testing.T) {
	cluster := &Cluster{
		ID:         "cluster-1",
		Name:       "test-cluster",
		Hosts:      []*HostRef{},
		NodeGroups: []*NodeGroup{},
		Network: &NetworkConfig{
			Type:    "nat",
			Name:    "default",
			CIDR:    "10.0.0.0/24",
			Gateway: "10.0.0.1",
		},
		Status:     "pending",
		DeployedAt: time.Time{},
	}

	if cluster.ID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", cluster.ID)
	}
	if cluster.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", cluster.Name)
	}
	if cluster.Network == nil {
		t.Error("network should not be nil")
	}
	if cluster.Network.Type != "nat" {
		t.Errorf("expected nat network type, got %s", cluster.Network.Type)
	}
}

// TestHostRef tests host reference structure
func TestHostRef_Create(t *testing.T) {
	host := &HostRef{
		HostID:    "host-1",
		HostName:  "worker-node",
		NodeCount: 3,
	}

	if host.HostID != "host-1" {
		t.Errorf("expected host-1, got %s", host.HostID)
	}
	if host.HostName != "worker-node" {
		t.Errorf("expected worker-node, got %s", host.HostName)
	}
	if host.NodeCount != 3 {
		t.Errorf("expected 3 nodes, got %d", host.NodeCount)
	}
}

// TestNodeGroup tests node group structure
func TestNodeGroup_Create(t *testing.T) {
	group := &NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    5,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
		HostID:   "host-1",
	}

	if group.Name != "workers" {
		t.Errorf("expected workers, got %s", group.Name)
	}
	if group.Role != "worker" {
		t.Errorf("expected worker role, got %s", group.Role)
	}
	if group.Count != 5 {
		t.Errorf("expected 5 nodes, got %d", group.Count)
	}
	if group.CPU != 4 {
		t.Errorf("expected 4 CPUs, got %d", group.CPU)
	}
	if group.MemoryMB != 8192 {
		t.Errorf("expected 8192MB memory, got %d", group.MemoryMB)
	}
}

// TestNetworkConfig tests network configuration
func TestNetworkConfig_Create(t *testing.T) {
	config := &NetworkConfig{
		Type:    "bridge",
		Name:    "br-cluster",
		CIDR:    "172.20.0.0/16",
		Gateway: "172.20.0.1",
	}

	if config.Type != "bridge" {
		t.Errorf("expected bridge type, got %s", config.Type)
	}
	if config.Name != "br-cluster" {
		t.Errorf("expected br-cluster, got %s", config.Name)
	}
	if config.CIDR != "172.20.0.0/16" {
		t.Errorf("expected 172.20.0.0/16 CIDR, got %s", config.CIDR)
	}
	if config.Gateway != "172.20.0.1" {
		t.Errorf("expected gateway 172.20.0.1, got %s", config.Gateway)
	}
}

// TestWizard_SetName tests setting cluster name
func TestWizard_SetName(t *testing.T) {
	wizard := NewWizard()

	wizard.SetName("my-cluster")

	if wizard.cluster.Name != "my-cluster" {
		t.Errorf("expected my-cluster, got %s", wizard.cluster.Name)
	}
}

// TestWizard_AddNodeGroup tests adding node groups
func TestWizard_AddNodeGroup(t *testing.T) {
	wizard := NewWizard()

	group := &NodeGroup{
		Name:     "control-plane",
		Role:     "control-plane",
		Count:    1,
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}

	wizard.AddNodeGroup(group)

	if len(wizard.cluster.NodeGroups) != 1 {
		t.Errorf("expected 1 node group, got %d", len(wizard.cluster.NodeGroups))
	}
	if wizard.cluster.NodeGroups[0].Name != "control-plane" {
		t.Errorf("expected control-plane, got %s", wizard.cluster.NodeGroups[0].Name)
	}
}

// TestWizard_SetNetwork tests setting network configuration
func TestWizard_SetNetwork(t *testing.T) {
	wizard := NewWizard()

	config := &NetworkConfig{
		Type:    "custom",
		Name:    "br-custom",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
	}

	wizard.SetNetwork(config)

	if wizard.cluster.Network == nil {
		t.Fatal("network should not be nil")
	}
	if wizard.cluster.Network.Type != "custom" {
		t.Errorf("expected custom type, got %s", wizard.cluster.Network.Type)
	}
}

// TestWizard_NextStep tests wizard step progression
func TestWizard_NextStep(t *testing.T) {
	wizard := NewWizard()

	initialStep := wizard.step
	wizard.NextStep()

	if wizard.step != initialStep+1 {
		t.Errorf("expected step %d, got %d", initialStep+1, wizard.step)
	}
}

// TestWizard_PrevStep tests wizard step regression
func TestWizard_PrevStep(t *testing.T) {
	wizard := NewWizard()
	wizard.step = 2

	wizard.PrevStep()

	if wizard.step != 1 {
		t.Errorf("expected step 1, got %d", wizard.step)
	}
}

// TestWizard_GetStep tests getting current step
func TestWizard_GetStep(t *testing.T) {
	wizard := NewWizard()
	wizard.step = 3

	step := wizard.GetStep()

	if step != 3 {
		t.Errorf("expected step 3, got %d", step)
	}
}

// TestWizard_GetCluster tests getting cluster configuration
func TestWizard_GetCluster(t *testing.T) {
	wizard := NewWizard()
	wizard.cluster.Name = "test-cluster"

	cluster := wizard.GetCluster()

	if cluster == nil {
		t.Fatal("cluster should not be nil")
	}
	if cluster.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", cluster.Name)
	}
}

// TestWizard_Validate tests cluster validation
func TestWizard_Validate(t *testing.T) {
	wizard := NewWizard()

	// Empty cluster should fail validation
	err := wizard.Validate()
	if err == nil {
		t.Error("expected validation error for empty cluster")
	}

	// Add required fields
	wizard.SetName("test-cluster")
	wizard.AddNodeGroup(&NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    1,
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	})
	wizard.SetNetwork(&NetworkConfig{
		Type:    "nat",
		Name:    "default",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
	})

	// Now should validate
	err = wizard.Validate()
	if err != nil {
		t.Errorf("validation should pass: %v", err)
	}
}

// TestNodeGroup_JSON tests node group structure
func TestNodeGroup_JSON(t *testing.T) {
	group := &NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    5,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
	}

	if group.Name != "workers" {
		t.Errorf("expected workers, got %s", group.Name)
	}
}

// TestCluster_JSON tests cluster structure
func TestCluster_JSON(t *testing.T) {
	cluster := &Cluster{
		ID:         "cluster-1",
		Name:       "test-cluster",
		Hosts:      []*HostRef{},
		NodeGroups: []*NodeGroup{},
		Status:     "pending",
	}

	if cluster.ID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", cluster.ID)
	}
}
