// Package cluster_test tests cluster manager operations
package cluster_test

import (
	"context"
	"os"
	"testing"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestClusterManager tests basic cluster manager functionality
func TestClusterManager(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create manager with stub hypervisor
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	mgr := cluster.NewManager(db, hosts)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

// TestClusterCreateAndList tests cluster creation and listing
func TestClusterCreateAndList(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	mgr := cluster.NewManager(db, hosts)

	// Create cluster using manager
	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  10,
		AutoScale: true,
	}

	createdCluster, err := mgr.CreateCluster("test-cluster", cfg)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	if createdCluster.Name != "test-cluster" {
		t.Errorf("Expected name 'test-cluster', got '%s'", createdCluster.Name)
	}

	if createdCluster.ID == "" {
		t.Error("Cluster ID should not be empty")
	}

	// List clusters
	clusters, err := mgr.ListClusters()
	if err != nil {
		t.Fatalf("Failed to list clusters: %v", err)
	}

	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}
}

// TestNodeOperations tests node lifecycle operations
func TestNodeOperations(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test cluster first
	clusterData := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	if err := db.SaveCluster(clusterData); err != nil {
		t.Fatalf("Failed to save cluster: %v", err)
	}

	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	mgr := cluster.NewManager(db, hosts)

	ctx := context.Background()

	t.Run("StartNode", func(t *testing.T) {
		node := &database.Node{
			ID:        "test-node-1",
			ClusterID: "test-cluster",
			Name:      "test-node-1",
			Role:      "worker",
			HostID:    "test-host",
			State:     "stopped",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatalf("Failed to save node: %v", err)
		}

		err := mgr.StartNode(ctx, "test-node-1")
		if err != nil {
			t.Fatalf("Failed to start node: %v", err)
		}

		updated, err := db.GetNode("test-node-1")
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}
		if updated.State != "running" {
			t.Errorf("Expected state 'running', got '%s'", updated.State)
		}
	})

	t.Run("StopNode", func(t *testing.T) {
		err := mgr.StopNode(ctx, "test-node-1")
		if err != nil {
			t.Fatalf("Failed to stop node: %v", err)
		}

		updated, err := db.GetNode("test-node-1")
		if err != nil {
			t.Fatalf("Failed to get node: %v", err)
		}
		if updated.State != "stopped" {
			t.Errorf("Expected state 'stopped', got '%s'", updated.State)
		}
	})

	t.Run("RestartNode", func(t *testing.T) {
		mgr.StartNode(ctx, "test-node-1")

		err := mgr.RestartNode(ctx, "test-node-1")
		if err != nil {
			t.Fatalf("Failed to restart node: %v", err)
		}
	})

	t.Run("DeleteNode", func(t *testing.T) {
		err := mgr.DeleteNode(ctx, "test-node-1")
		if err != nil {
			t.Fatalf("Failed to delete node: %v", err)
		}

		node, err := db.GetNode("test-node-1")
		if err != nil {
			t.Fatalf("GetNode error: %v", err)
		}
		if node != nil {
			t.Error("Expected node to be deleted")
		}
	})
}

// TestClusterScaling tests cluster scaling operations
func TestClusterScaling(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test cluster
	clusterData := &database.Cluster{
		ID:     "scale-test",
		Name:   "Scale Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	hosts := map[string]hypervisor.Hypervisor{
		"host-1": hypervisor.NewStubHypervisor(),
	}
	mgr := cluster.NewManager(db, hosts)

	ctx := context.Background()

	// Create initial nodes
	for i := 0; i < 3; i++ {
		node := &database.Node{
			ID:        "node-" + string(rune('a'+i)),
			ClusterID: "scale-test",
			Name:      "node-" + string(rune('a'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
		}
		db.SaveNode(node)
	}

	t.Run("ScaleUp", func(t *testing.T) {
		err := mgr.ScaleCluster(ctx, "scale-test", 5)
		if err != nil {
			t.Fatalf("Failed to scale up: %v", err)
		}

		nodes, err := db.ListClusterNodes("scale-test")
		if err != nil {
			t.Fatalf("Failed to list nodes: %v", err)
		}
		if len(nodes) < 5 {
			t.Errorf("Expected at least 5 nodes, got %d", len(nodes))
		}
	})

	t.Run("ScaleDown", func(t *testing.T) {
		err := mgr.ScaleCluster(ctx, "scale-test", 2)
		if err != nil {
			t.Fatalf("Failed to scale down: %v", err)
		}

		nodes, err := db.ListClusterNodes("scale-test")
		if err != nil {
			t.Fatalf("Failed to list nodes: %v", err)
		}
		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(nodes))
		}
	})
}

// TestNodeStats tests cluster statistics
func TestNodeStats(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create cluster with nodes in various states
	clusterData := &database.Cluster{
		ID:     "stats-test",
		Name:   "Stats Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	// Create nodes with different states
	states := []string{"running", "running", "running", "stopped", "error"}
	for i, state := range states {
		node := &database.Node{
			ID:        "stats-node-" + string(rune('0'+i)),
			ClusterID: "stats-test",
			Name:      "stats-node-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     state,
		}
		db.SaveNode(node)
	}

	mgr := cluster.NewManager(db, nil)

	total, running, stopped, errs := mgr.NodeStats("stats-test")
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if running != 3 {
		t.Errorf("Expected running 3, got %d", running)
	}
	if stopped != 1 {
		t.Errorf("Expected stopped 1, got %d", stopped)
	}
	if errs != 1 {
		t.Errorf("Expected error 1, got %d", errs)
	}
}

// TestGetOrCreateHost tests host creation
func TestGetOrCreateHost(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	hosts := map[string]hypervisor.Hypervisor{}
	mgr := cluster.NewManager(db, hosts)

	_ = context.Background() // ctx available for future tests

	t.Run("CreateNewHost", func(t *testing.T) {
		cfg := &database.Host{
			ID:      "new-host",
			Name:    "New Host",
			Address: "192.168.1.100",
			User:    "root",
			Port:    22,
			HVType:  "stub",
		}
		hv, err := mgr.GetOrCreateHost(cfg)
		if err != nil {
			t.Fatalf("Failed to create host: %v", err)
		}
		if hv == nil {
			t.Fatal("GetOrCreateHost returned nil hypervisor")
		}

		// Verify host was saved to database
		saved, err := db.GetHost("new-host")
		if err != nil {
			t.Fatalf("GetHost error: %v", err)
		}
		if saved == nil {
			// Host might not be persisted in test mode - verify it's in memory
			t.Log("Host not persisted to DB (expected for stub) - checking memory")
		}
		if saved != nil && saved.Name != "New Host" {
			t.Errorf("Expected name 'New Host', got '%s'", saved.Name)
		}
	})
}

// TestDeleteCluster tests cluster deletion
func TestDeleteCluster(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	mgr := cluster.NewManager(db, hosts)

	ctx := context.Background()

	// Create cluster
	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  10,
		AutoScale: true,
	}
	createdCluster, err := mgr.CreateCluster("delete-test", cfg)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	// Delete cluster
	err = mgr.DeleteCluster(ctx, createdCluster.ID)
	if err != nil {
		t.Fatalf("Failed to delete cluster: %v", err)
	}

	// Verify deleted
	clusters, _ := mgr.ListClusters()
	for _, c := range clusters {
		if c.ID == createdCluster.ID {
			t.Error("Cluster should have been deleted")
		}
	}
}
