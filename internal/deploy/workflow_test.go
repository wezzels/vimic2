// Package deploy provides additional deployment tests
package deploy

import (
	"testing"
)

// TestWizard_FullWorkflow tests full wizard workflow
func TestWizard_FullWorkflow(t *testing.T) {
	wizard := NewWizard()

	// Step 1: Set name
	wizard.SetName("production-cluster")
	cluster := wizard.GetCluster()
	if cluster.Name != "production-cluster" {
		t.Errorf("expected production-cluster, got %s", cluster.Name)
	}

	// Step 2: Configure network
	wizard.SetNetwork(&NetworkConfig{
		Type:    "bridge",
		Name:    "br0",
		CIDR:    "192.168.100.0/24",
		Gateway: "192.168.100.1",
	})
	if wizard.cluster.Network.Type != "bridge" {
		t.Errorf("expected bridge, got %s", wizard.cluster.Network.Type)
	}

	// Step 3: Add node groups
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

	// Step through wizard
	initialStep := wizard.GetStep()
	wizard.NextStep()
	if wizard.GetStep() <= initialStep {
		t.Error("expected step to increase")
	}

	wizard.PrevStep()
	if wizard.GetStep() != initialStep {
		t.Error("expected step to return to initial")
	}
}

// TestWizard_RemoveNodeGroup tests removal
func TestWizard_RemoveNodeGroup(t *testing.T) {
	wizard := NewWizard()

	wizard.AddNodeGroup(&NodeGroup{Name: "group-1"})
	wizard.AddNodeGroup(&NodeGroup{Name: "group-2"})
	wizard.AddNodeGroup(&NodeGroup{Name: "group-3"})

	// Remove middle
	wizard.RemoveNodeGroup(1)

	if len(wizard.cluster.NodeGroups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(wizard.cluster.NodeGroups))
	}

	// Check order
	if wizard.cluster.NodeGroups[0].Name != "group-1" {
		t.Errorf("expected group-1, got %s", wizard.cluster.NodeGroups[0].Name)
	}
	if wizard.cluster.NodeGroups[1].Name != "group-3" {
		t.Errorf("expected group-3, got %s", wizard.cluster.NodeGroups[1].Name)
	}
}

// TestWizard_Validate tests validation
func TestWizard_Validate_Tests(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Wizard)
		expectErr bool
	}{
		{
			name: "Empty name",
			setup: func(w *Wizard) {
				// Leave name empty
			},
			expectErr: true,
		},
		{
			name: "Valid config",
			setup: func(w *Wizard) {
				w.SetName("test")
				w.AddNodeGroup(&NodeGroup{
					Name:     "workers",
					Role:     "worker",
					Count:    1,
					CPU:      2,
					MemoryMB: 4096,
					DiskGB:   20,
					Image:    "ubuntu-22.04",
				})
				w.SetNetwork(&NetworkConfig{
					Type:    "nat",
					Name:    "default",
					CIDR:    "10.0.0.0/24",
					Gateway: "10.0.0.1",
				})
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := NewWizard()
			tt.setup(wizard)

			err := wizard.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestNetworkConfigDefaults tests network defaults
func TestNetworkConfig_Defaults(t *testing.T) {
	networks := []struct {
		name   string
		config NetworkConfig
	}{
		{"NAT", NetworkConfig{Type: "nat", Name: "default", CIDR: "10.0.0.0/24"}},
		{"Bridge", NetworkConfig{Type: "bridge", Name: "br0", CIDR: "192.168.1.0/24"}},
		{"Custom", NetworkConfig{Type: "custom", CIDR: "172.16.0.0/16"}},
	}

	for _, tt := range networks {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Type == "" {
				t.Error("expected non-empty type")
			}
			if tt.config.CIDR == "" {
				t.Error("expected non-empty CIDR")
			}
		})
	}
}

// TestNodeGroupDefaults tests node group defaults
func TestNodeGroup_Defaults(t *testing.T) {
	groups := []struct {
		name  string
		group NodeGroup
	}{
		{"Master", NodeGroup{Name: "master", Role: "master", Count: 3}},
		{"Worker", NodeGroup{Name: "worker", Role: "worker", Count: 5}},
	}

	for _, tt := range groups {
		t.Run(tt.name, func(t *testing.T) {
			if tt.group.Role == "" {
				t.Error("expected non-empty role")
			}
		})
	}
}

// TestClusterStatus tests cluster status values
func TestCluster_Status(t *testing.T) {
	statuses := []string{"pending", "deploying", "running", "error", "stopped"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cluster := &Cluster{Status: status}
			if cluster.Status != status {
				t.Errorf("expected %s, got %s", status, cluster.Status)
			}
		})
	}
}

// TestProgressStatus tests progress status
func TestProgress_Status(t *testing.T) {
	progress := &Progress{
		ClusterID:     "cluster-1",
		TotalNodes:    10,
		DeployedNodes: 7,
		CurrentNode:   "node-7",
		Status:        "deploying",
	}

	if progress.TotalNodes != 10 {
		t.Errorf("expected 10, got %d", progress.TotalNodes)
	}
	if progress.DeployedNodes >= progress.TotalNodes {
		t.Error("deployment not complete")
	}
}
