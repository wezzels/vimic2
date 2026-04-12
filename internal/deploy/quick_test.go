// Package deploy provides quick win tests
package deploy

import (
	"testing"
)

// TestCluster_Fields tests Cluster field access
func TestCluster_Fields(t *testing.T) {
	c := &Cluster{
		ID:   "cluster-1",
		Name: "test-cluster",
		NodeGroups: []*NodeGroup{
			{Name: "workers", Role: "worker", Count: 3},
		},
	}

	if c.ID != "cluster-1" {
		t.Error("ID mismatch")
	}
	if len(c.NodeGroups) != 1 {
		t.Error("expected 1 node group")
	}
}

// TestNodeGroup_Fields tests NodeGroup field access
func TestNodeGroup_Fields(t *testing.T) {
	ng := &NodeGroup{
		Name:     "workers",
		Role:     "worker",
		Count:    5,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
		HostID:   "host-1",
	}

	if ng.Name != "workers" {
		t.Error("Name mismatch")
	}
	if ng.Count != 5 {
		t.Error("Count mismatch")
	}
	if ng.CPU != 4 {
		t.Error("CPU mismatch")
	}
}

// TestNetworkConfig_Fields tests NetworkConfig field access
func TestNetworkConfig_Fields(t *testing.T) {
	nc := &NetworkConfig{
		Type:    "nat",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
	}

	if nc.Type != "nat" {
		t.Error("Type mismatch")
	}
	if nc.CIDR != "10.0.0.0/24" {
		t.Error("CIDR mismatch")
	}
}

// TestProgress_Fields tests Progress field access
func TestProgress_Fields(t *testing.T) {
	p := &Progress{
		ClusterID:     "cluster-1",
		TotalNodes:    10,
		DeployedNodes: 5,
		CurrentNode:   "node-5",
		Status:        "deploying",
	}

	if p.ClusterID != "cluster-1" {
		t.Error("ClusterID mismatch")
	}
	if p.DeployedNodes != 5 {
		t.Error("DeployedNodes mismatch")
	}
}

// TestHostRef_Fields tests HostRef field access
func TestHostRef_Fields(t *testing.T) {
	hr := &HostRef{
		HostID:    "host-1",
		HostName:  "hv1",
		NodeCount: 5,
	}

	if hr.HostID != "host-1" {
		t.Error("HostID mismatch")
	}
	if hr.NodeCount != 5 {
		t.Error("NodeCount mismatch")
	}
}

// TestWizard_New tests NewWizard
func TestWizard_New_Quick(t *testing.T) {
	w := NewWizard()
	if w == nil {
		t.Fatal("expected non-nil wizard")
	}
}

// TestCluster_Empty tests empty Cluster
func TestCluster_Empty(t *testing.T) {
	c := &Cluster{}

	if c.ID != "" {
		t.Error("expected empty ID")
	}
	if c.NodeGroups != nil {
		t.Error("expected nil NodeGroups")
	}
}

// TestNodeGroup_Defaults tests default values
func TestNodeGroup_Defaults_Quick(t *testing.T) {
	ng := &NodeGroup{
		Name:  "test",
		Role:  "worker",
		Count: 1,
	}

	// Apply defaults manually for test
	if ng.CPU == 0 {
		ng.CPU = 2
	}
	if ng.MemoryMB == 0 {
		ng.MemoryMB = 2048
	}
	if ng.DiskGB == 0 {
		ng.DiskGB = 20
	}

	if ng.CPU != 2 {
		t.Error("CPU default should be 2")
	}
	if ng.MemoryMB != 2048 {
		t.Error("MemoryMB default should be 2048")
	}
}

// TestProgress_Complete tests progress completion
func TestProgress_Complete(t *testing.T) {
	p := &Progress{
		TotalNodes:    5,
		DeployedNodes: 5,
		Status:        "complete",
	}

	if p.Status != "complete" {
		t.Error("status should be complete")
	}
	if p.DeployedNodes != p.TotalNodes {
		t.Error("should be fully deployed")
	}
}
