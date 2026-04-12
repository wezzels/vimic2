//go:build integration

// Package cluster provides integration tests for cluster management
package cluster

import (
	"context"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestIntegration_Cluster_CreateGet tests creating and getting clusters
func TestIntegration_Cluster_CreateGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}

	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	cfg := &database.ClusterConfig{
		MinNodes:      1,
		MaxNodes:      5,
		AutoScale:     false,
		ScaleOnCPU:    80.0,
		ScaleOnMemory: 80.0,
		CooldownSec:   300,
	}

	cluster, err := mgr.CreateCluster("test-cluster-1", cfg)
	if err != nil {
		t.Fatalf("CreateCluster failed: %v", err)
	}

	t.Logf("Created cluster: %s (ID: %s)", cluster.Name, cluster.ID)

	retrieved, err := mgr.GetCluster(cluster.ID)
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}

	if retrieved.ID != cluster.ID {
		t.Errorf("expected ID %s, got %s", cluster.ID, retrieved.ID)
	}

	t.Logf("Retrieved cluster: %s", retrieved.Name)
}

// TestIntegration_Cluster_List tests listing clusters
func TestIntegration_Cluster_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  3,
		AutoScale: false,
	}

	for i := 0; i < 3; i++ {
		mgr.CreateCluster("list-cluster-"+string(rune('0'+i)), cfg)
	}

	clusters, err := mgr.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}

	if len(clusters) < 3 {
		t.Errorf("expected at least 3 clusters, got %d", len(clusters))
	}

	t.Logf("Listed %d clusters", len(clusters))
}

// TestIntegration_Cluster_Delete tests deleting a cluster
func TestIntegration_Cluster_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  1,
		AutoScale: false,
	}
	cluster, _ := mgr.CreateCluster("delete-cluster-1", cfg)

	ctx := context.Background()
	err := mgr.DeleteCluster(ctx, cluster.ID)
	if err != nil {
		t.Fatalf("DeleteCluster failed: %v", err)
	}

	t.Logf("Cluster deleted successfully")
}

// TestIntegration_Cluster_AddNode tests adding nodes
func TestIntegration_Cluster_AddNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  3,
		AutoScale: false,
	}
	cluster, _ := mgr.CreateCluster("node-cluster-1", cfg)

	ctx := context.Background()
	nodeCfg := &database.NodeConfig{
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
	}

	node, err := mgr.AddNode(ctx, cluster.ID, "test-host", "test-node-1", "worker", nodeCfg)
	if err != nil {
		t.Skipf("AddNode failed (expected without hypervisor): %v", err)
	}

	t.Logf("Added node: %s", node.Name)
}

// TestIntegration_Cluster_NodeStats tests node statistics
func TestIntegration_Cluster_NodeStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  1,
		AutoScale: false,
	}
	cluster, _ := mgr.CreateCluster("stats-cluster-1", cfg)

	total, running, stopped, errored := mgr.NodeStats(cluster.ID)
	t.Logf("Stats: total=%d, running=%d, stopped=%d, error=%d", total, running, stopped, errored)
}

// TestIntegration_Cluster_ListHosts tests listing hosts
func TestIntegration_Cluster_ListHosts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	hostList, err := mgr.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}

	t.Logf("Listed %d hosts", len(hostList))
}
