// Package database provides SQLite persistence for Vimic2
package database

import (
	"os"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	// Create temp database
	tmpfile, err := os.CreateTemp("", "vimic2_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	db, err := NewDB(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Test Host operations
	host := &Host{
		ID:         "test-host-1",
		Name:       "test-host",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/home/user/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	err = db.SaveHost(host)
	if err != nil {
		t.Fatalf("SaveHost failed: %v", err)
	}

	retrieved, err := db.GetHost("test-host-1")
	if err != nil {
		t.Fatalf("GetHost failed: %v", err)
	}
	if retrieved.Name != "test-host" {
		t.Errorf("Expected 'test-host', got '%s'", retrieved.Name)
	}
	if retrieved.Address != "192.168.1.100" {
		t.Errorf("Expected '192.168.1.100', got '%s'", retrieved.Address)
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}
	if len(hosts) != 1 {
		t.Errorf("Expected 1 host, got %d", len(hosts))
	}

	// Test Cluster operations
	cluster := &Cluster{
		ID:     "test-cluster-1",
		Name:   "test-cluster",
		Status: "pending",
		Config: &ClusterConfig{
			MinNodes: 1,
			MaxNodes: 10,
		},
	}

	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	retrievedCluster, err := db.GetCluster("test-cluster-1")
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}
	if retrievedCluster.Name != "test-cluster" {
		t.Errorf("Expected 'test-cluster', got '%s'", retrievedCluster.Name)
	}
	if retrievedCluster.Config.MinNodes != 1 {
		t.Errorf("Expected MinNodes 1, got %d", retrievedCluster.Config.MinNodes)
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}

	// Test Node operations
	node := &Node{
		ID:        "test-node-1",
		ClusterID: "test-cluster-1",
		HostID:    "test-host-1",
		Name:      "test-node",
		Role:      "worker",
		State:     "running",
		IP:        "192.168.1.101",
		Config: &NodeConfig{
			CPU:      2,
			MemoryMB: 2048,
			DiskGB:   20,
		},
	}

	err = db.SaveNode(node)
	if err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	retrievedNode, err := db.GetNode("test-node-1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrievedNode.Name != "test-node" {
		t.Errorf("Expected 'test-node', got '%s'", retrievedNode.Name)
	}
	if retrievedNode.Role != "worker" {
		t.Errorf("Expected 'worker', got '%s'", retrievedNode.Role)
	}

	nodes, err := db.ListClusterNodes("test-cluster-1")
	if err != nil {
		t.Fatalf("ListClusterNodes failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	// Test UpdateNodeState
	err = db.UpdateNodeState("test-node-1", "stopped", "")
	if err != nil {
		t.Fatalf("UpdateNodeState failed: %v", err)
	}

	updatedNode, _ := db.GetNode("test-node-1")
	if updatedNode.State != "stopped" {
		t.Errorf("Expected 'stopped', got '%s'", updatedNode.State)
	}

	// Test Metric operations
	metric := &Metric{
		NodeID: "test-node-1",
		CPU:    45.5,
		Memory: 67.8,
		Disk:   30.0,
	}

	err = db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("SaveMetric failed: %v", err)
	}

	latest, err := db.GetLatestMetric("test-node-1")
	if err != nil {
		t.Fatalf("GetLatestMetric failed: %v", err)
	}
	if latest.CPU != 45.5 {
		t.Errorf("Expected CPU 45.5, got %.1f", latest.CPU)
	}

	// Test GetNodeMetrics
	metrics, err := db.GetNodeMetrics("test-node-1", time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("GetNodeMetrics failed: %v", err)
	}
	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
	}

	// Test Alert operations
	alert := &Alert{
		ID:        "test-alert-1",
		RuleID:    "high-cpu",
		NodeID:    "test-node-1",
		NodeName:  "test-node",
		Metric:    "cpu",
		Value:     95.0,
		Threshold: 90.0,
		Message:   "High CPU on test-node (95%)",
		FiredAt:  time.Now(),
	}

	err = db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("SaveAlert failed: %v", err)
	}

	alerts, err := db.GetActiveAlerts()
	if err != nil {
		t.Fatalf("GetActiveAlerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}

	nodeAlerts, err := db.GetNodeAlerts("test-node-1")
	if err != nil {
		t.Fatalf("GetNodeAlerts failed: %v", err)
	}
	if len(nodeAlerts) != 1 {
		t.Errorf("Expected 1 node alert, got %d", len(nodeAlerts))
	}

	// Test UpdateClusterStatus
	err = db.UpdateClusterStatus("test-cluster-1", "running")
	if err != nil {
		t.Fatalf("UpdateClusterStatus failed: %v", err)
	}

	updatedCluster, _ := db.GetCluster("test-cluster-1")
	if updatedCluster.Status != "running" {
		t.Errorf("Expected 'running', got '%s'", updatedCluster.Status)
	}

	// Test CleanupOldMetrics
	_, err = db.CleanupOldMetrics(time.Hour)
	if err != nil {
		t.Fatalf("CleanupOldMetrics failed: %v", err)
	}

	// Test DeleteNode
	err = db.DeleteNode("test-node-1")
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	nodes, _ = db.ListAllNodes()
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes after delete, got %d", len(nodes))
	}

	// Test DeleteCluster
	err = db.DeleteCluster("test-cluster-1")
	if err != nil {
		t.Fatalf("DeleteCluster failed: %v", err)
	}

	clusters, _ = db.ListClusters()
	if len(clusters) != 0 {
		t.Errorf("Expected 0 clusters after delete, got %d", len(clusters))
	}

	// Test DeleteHost
	err = db.DeleteHost("test-host-1")
	if err != nil {
		t.Fatalf("DeleteHost failed: %v", err)
	}

	hosts, _ = db.ListHosts()
	if len(hosts) != 0 {
		t.Errorf("Expected 0 hosts after delete, got %d", len(hosts))
	}
}

func TestClusterConfig(t *testing.T) {
	cfg := &ClusterConfig{
		MinNodes:      3,
		MaxNodes:      10,
		AutoScale:     true,
		ScaleOnCPU:    70.0,
		ScaleOnMemory: 80.0,
		CooldownSec:   300,
		Network: &NetworkConfig{
			Type: "nat",
			CIDR: "192.168.100.0/24",
		},
		NodeDefaults: &NodeConfig{
			CPU:      2,
			MemoryMB: 2048,
			DiskGB:   20,
		},
	}

	if cfg.MinNodes != 3 {
		t.Errorf("Expected MinNodes 3, got %d", cfg.MinNodes)
	}
	if cfg.MaxNodes != 10 {
		t.Errorf("Expected MaxNodes 10, got %d", cfg.MaxNodes)
	}
	if !cfg.AutoScale {
		t.Error("Expected AutoScale to be true")
	}
	if cfg.Network.Type != "nat" {
		t.Errorf("Expected Network.Type 'nat', got '%s'", cfg.Network.Type)
	}
}

func TestNodeDefaults(t *testing.T) {
	cfg := &NodeConfig{
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
	}

	if cfg.CPU != 4 {
		t.Errorf("Expected CPU 4, got %d", cfg.CPU)
	}
	if cfg.MemoryMB != 8192 {
		t.Errorf("Expected Memory 8192, got %d", cfg.MemoryMB)
	}
}
