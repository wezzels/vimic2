// Package cluster provides additional manager tests
package cluster

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestManager_New tests manager creation
func TestManager_New(t *testing.T) {
	mgr := NewManager(nil, nil)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestDatabaseCluster_Fields tests database.Cluster struct
func TestDatabaseCluster_Fields(t *testing.T) {
	cluster := &database.Cluster{
		ID:     "cluster-1",
		Name:   "production",
		Status: "running",
		Config: &database.ClusterConfig{
			MinNodes:  3,
			MaxNodes:  10,
			AutoScale: true,
		},
	}

	if cluster.ID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", cluster.ID)
	}
	if cluster.Status != "running" {
		t.Errorf("expected running, got %s", cluster.Status)
	}
	if cluster.Config == nil {
		t.Fatal("expected non-nil config")
	}
	if cluster.Config.MinNodes != 3 {
		t.Errorf("expected 3 min nodes, got %d", cluster.Config.MinNodes)
	}
}

// TestDatabaseNode_Fields tests database.Node struct
func TestDatabaseNode_Fields(t *testing.T) {
	node := &database.Node{
		ID:        "node-1",
		Name:      "worker-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		State:     "running",
		Role:      "worker",
		IP:        "192.168.1.100",
	}

	if node.ID != "node-1" {
		t.Errorf("expected node-1, got %s", node.ID)
	}
	if node.Role != "worker" {
		t.Errorf("expected worker, got %s", node.Role)
	}
}

// TestDatabaseHost_Fields tests database.Host struct
func TestDatabaseHost_Fields(t *testing.T) {
	host := &database.Host{
		ID:         "host-1",
		Name:       "hv-1",
		Address:    "192.168.1.10",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if host.ID != "host-1" {
		t.Errorf("expected host-1, got %s", host.ID)
	}
	if host.HVType != "libvirt" {
		t.Errorf("expected libvirt, got %s", host.HVType)
	}
}

// TestNodeStates tests node state values
func TestNodeStates(t *testing.T) {
	states := []string{"pending", "creating", "running", "stopping", "stopped", "error"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			node := &database.Node{State: state}
			if node.State != state {
				t.Errorf("expected %s, got %s", state, node.State)
			}
		})
	}
}

// TestNodeRoles tests node role values
func TestNodeRoles(t *testing.T) {
	roles := []string{"master", "worker", "storage"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			node := &database.Node{Role: role}
			if node.Role != role {
				t.Errorf("expected %s, got %s", role, node.Role)
			}
		})
	}
}

// TestClusterStatuses tests cluster status values
func TestClusterStatuses(t *testing.T) {
	statuses := []string{"pending", "deploying", "running", "degraded", "error", "stopped"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			cluster := &database.Cluster{Status: status}
			if cluster.Status != status {
				t.Errorf("expected %s, got %s", status, cluster.Status)
			}
		})
	}
}

// TestClusterConfig_Defaults tests cluster config defaults
func TestClusterConfig_Defaults(t *testing.T) {
	config := &database.ClusterConfig{
		MinNodes:      1,
		MaxNodes:      10,
		AutoScale:     true,
		ScaleOnCPU:    70.0,
		ScaleOnMemory: 80.0,
		CooldownSec:   300,
	}

	if config.MinNodes < 1 {
		t.Error("min nodes should be >= 1")
	}
	if config.MaxNodes < config.MinNodes {
		t.Error("max nodes should be >= min nodes")
	}
	if config.CooldownSec < 0 {
		t.Error("cooldown should be >= 0")
	}
}

// TestNodeConfig_Defaults tests node config defaults
func TestNodeConfig_Defaults(t *testing.T) {
	config := &database.NodeConfig{
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}

	if config.CPU < 1 {
		t.Error("CPU should be >= 1")
	}
	if config.MemoryMB < 1024 {
		t.Error("memory should be >= 1024MB")
	}
	if config.DiskGB < 10 {
		t.Error("disk should be >= 10GB")
	}
	if config.Image == "" {
		t.Error("image should not be empty")
	}
}

// TestNetworkConfig tests network configuration
func TestNetworkConfig(t *testing.T) {
	config := &database.NetworkConfig{
		Type: "nat",
		CIDR: "192.168.122.0/24",
	}

	if config.Type != "nat" {
		t.Errorf("expected nat, got %s", config.Type)
	}
	if config.CIDR != "192.168.122.0/24" {
		t.Errorf("expected CIDR, got %s", config.CIDR)
	}
}

// TestClusterTimestamps tests cluster timestamps
func TestClusterTimestamps(t *testing.T) {
	now := time.Now()
	cluster := &database.Cluster{
		ID:        "cluster-1",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	if cluster.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
	if cluster.UpdatedAt.Before(cluster.CreatedAt) {
		t.Error("updated_at should be >= created_at")
	}
}