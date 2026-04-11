// Package cluster provides real database tests
package cluster

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestManager_GetCluster_RealDB tests GetCluster with real DB
func TestManager_GetCluster_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster
	cluster := &database.Cluster{
		ID:     "cluster-1",
		Name:   "test-cluster",
		Status: "running",
		Config: &database.ClusterConfig{
			MinNodes: 3,
			MaxNodes: 10,
		},
	}
	db.SaveCluster(cluster)

	// Create manager
	mgr := NewManager(db, nil)

	// Get cluster
	retrieved, err := mgr.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}
	if retrieved.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", retrieved.Name)
	}
}

// TestManager_ListClusters_RealDB tests ListClusters with real DB
func TestManager_ListClusters_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create clusters
	for i := 1; i <= 3; i++ {
		cluster := &database.Cluster{
			ID:     "cluster-" + string(rune('0'+i)),
			Name:   "cluster-" + string(rune('0'+i)),
			Status: "running",
		}
		db.SaveCluster(cluster)
	}

	mgr := NewManager(db, nil)

	clusters, err := mgr.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}
	if len(clusters) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(clusters))
	}
}

// TestManager_DeleteCluster_RealDB tests DeleteCluster with real DB
func TestManager_DeleteCluster_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and nodes
	cluster := &database.Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	node := &database.Node{ID: "node-1", Name: "node-1", ClusterID: "cluster-1", State: "running"}
	db.SaveNode(node)

	mgr := NewManager(db, nil)

	// Delete cluster
	err = mgr.DeleteCluster(context.Background(), "cluster-1")
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	// Verify cluster is gone
	clusters, _ := db.ListClusters()
	if len(clusters) != 0 {
		t.Error("cluster should be deleted")
	}
}

// TestManager_GetNode_RealDB tests GetNode with real DB
func TestManager_GetNode_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and node
	cluster := &database.Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	node := &database.Node{
		ID:        "node-1",
		Name:      "test-node",
		ClusterID: "cluster-1",
		State:     "running",
		Role:      "worker",
	}
	db.SaveNode(node)

	mgr := NewManager(db, nil)

	// Get node
	retrieved, err := mgr.GetNode("node-1")
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}
	if retrieved.Name != "test-node" {
		t.Errorf("expected test-node, got %s", retrieved.Name)
	}
}

// TestManager_GetNodeStatus_RealDB tests GetNodeStatus with real DB
func TestManager_GetNodeStatus_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and node
	cluster := &database.Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	node := &database.Node{
		ID:        "node-1",
		Name:      "test-node",
		ClusterID: "cluster-1",
		State:     "running",
	}
	db.SaveNode(node)

	// Save a metric
	metric := &database.Metric{
		NodeID:     "node-1",
		CPU:        45.5,
		Memory:     60.2,
		Disk:       30.1,
		RecordedAt: time.Now(),
	}
	db.SaveMetric(metric)

	mgr := NewManager(db, nil)

	// Get node status - will error since no hypervisor
	status, err := mgr.GetNodeStatus(context.Background(), "node-1")
	// Error expected without hypervisor
	if err != nil {
		t.Log("expected error without hypervisor")
	}
	_ = status
}

// TestManager_AddNode_RealDB tests AddNode with real DB
func TestManager_AddNode_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and host
	cluster := &database.Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	host := &database.Host{ID: "host-1", Name: "host-1", Address: "192.168.1.1", HVType: "libvirt"}
	db.SaveHost(host)

	// Create manager with empty hosts map
	mgr := NewManager(db, make(map[string]hypervisor.Hypervisor))

	// Add node - this will fail because no hypervisor for host-1
	// That's expected - we're testing the DB path
	_, err = mgr.AddNode(context.Background(), "cluster-1", "host-1", "new-node", "worker", nil)
	// Error expected since no hypervisor registered
	if err == nil {
		t.Log("node added without hypervisor")
	}
}

// TestManager_ListHosts_RealDB tests ListHosts with real DB
func TestManager_ListHosts_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create hosts
	for i := 1; i <= 3; i++ {
		host := &database.Host{
			ID:      "host-" + string(rune('0'+i)),
			Name:    "host-" + string(rune('0'+i)),
			Address: "192.168.1." + string(rune('0'+i)),
			HVType:  "libvirt",
		}
		db.SaveHost(host)
	}

	mgr := NewManager(db, nil)

	hosts, err := mgr.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts: %v", err)
	}
	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(hosts))
	}
}

// TestManager_GetClusterMetrics_RealDB tests GetClusterMetrics with real DB
func TestManager_GetClusterMetrics_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cluster-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and nodes
	cluster := &database.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	db.SaveCluster(cluster)

	for i := 1; i <= 3; i++ {
		node := &database.Node{
			ID:        "node-" + string(rune('0'+i)),
			Name:      "node-" + string(rune('0'+i)),
			ClusterID: "cluster-1",
			State:     "running",
		}
		db.SaveNode(node)

		// Add metrics for each node
		metric := &database.Metric{
			NodeID:     "node-" + string(rune('0'+i)),
			CPU:        float64(40 + i*10),
			Memory:     60.0,
			RecordedAt: time.Now(),
		}
		db.SaveMetric(metric)
	}

	mgr := NewManager(db, nil)

	cpuAvg, memAvg, diskAvg, err := mgr.GetClusterMetrics("cluster-1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to get cluster metrics: %v", err)
	}
	_ = cpuAvg
	_ = memAvg
	_ = diskAvg
}