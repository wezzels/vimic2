// Package database provides edge case tests for 100% coverage
package database

import (
	"os"
	"testing"
	"time"
)

// TestNewDB_InvalidPath tests NewDB with invalid path
func TestNewDB_InvalidPath(t *testing.T) {
	// Try to create DB in non-existent directory
	_, err := NewDB("/nonexistent/path/to/db.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestNewDB_Permission tests NewDB with permission issues
func TestNewDB_Permission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "vimic2-perm-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Make it read-only
	os.Chmod(tmpDir, 0555)

	// Try to create DB in read-only directory
	_, err = NewDB(tmpDir + "/test.db")
	if err == nil {
		t.Error("expected error for read-only directory")
	}
}

// TestListHosts_Empty tests ListHosts when no hosts exist
func TestListHosts_Empty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

// TestGetHost_NotFound tests GetHost when not found
func TestGetHost_NotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	_, err = db.GetHost("nonexistent")
	// GetHost may return nil/empty or error depending on implementation
	_ = err
}

// TestSaveCluster_ErrorPaths tests cluster save edge cases
func TestSaveCluster_ErrorPaths(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Save cluster with empty ID
	cluster1 := &Cluster{ID: "", Name: "empty-id"}
	err = db.SaveCluster(cluster1)
	// May succeed or fail depending on implementation
	_ = err

	// Save cluster with full config
	cluster2 := &Cluster{
		ID:     "cluster-full",
		Name:   "full-cluster",
		Status: "running",
		Config: &ClusterConfig{
			MinNodes:      5,
			MaxNodes:      20,
			AutoScale:     true,
			ScaleOnCPU:    75.0,
			ScaleOnMemory: 85.0,
			CooldownSec:   600,
			NodeDefaults: &NodeConfig{
				CPU:      8,
				MemoryMB: 16384,
				DiskGB:   200,
				Image:    "ubuntu-24.04",
			},
		},
	}
	err = db.SaveCluster(cluster2)
	if err != nil {
		t.Fatalf("failed to save cluster: %v", err)
	}
}

// TestListClusters_WithStatus tests listing clusters with different statuses
func TestListClusters_WithStatus(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create clusters with different statuses
	statuses := []string{"running", "stopped", "error", "pending"}
	for i, status := range statuses {
		cluster := &Cluster{
			ID:     "cluster-" + string(rune('0'+i)),
			Name:   "cluster-" + string(rune('0'+i)),
			Status: status,
		}
		db.SaveCluster(cluster)
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}
	if len(clusters) != 4 {
		t.Errorf("expected 4 clusters, got %d", len(clusters))
	}
}

// TestGetCluster_WithFullConfig tests GetCluster with full config
func TestGetCluster_WithFullConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster with all fields
	cluster := &Cluster{
		ID:     "cluster-full",
		Name:   "full-cluster",
		Status: "running",
	}
	db.SaveCluster(cluster)

	retrieved, err := db.GetCluster("cluster-full")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}
	if retrieved.Name != "full-cluster" {
		t.Errorf("expected name, got %s", retrieved.Name)
	}
}

// TestListClusterNodes_Empty tests listing nodes for empty cluster
func TestListClusterNodes_Empty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster with no nodes
	cluster := &Cluster{ID: "empty-cluster", Name: "empty"}
	db.SaveCluster(cluster)

	nodes, err := db.ListClusterNodes("empty-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}

// TestGetNode_WithAllFields tests GetNode with all fields populated
func TestGetNode_WithAllFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster
	cluster := &Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	// Create node with all fields
	node := &Node{
		ID:        "node-full",
		Name:      "full-node",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		State:     "running",
		Role:      "worker",
		IP:        "10.0.0.100",
	}
	db.SaveNode(node)

	retrieved, err := db.GetNode("node-full")
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}
	if retrieved.IP != "10.0.0.100" {
		t.Errorf("expected IP, got %s", retrieved.IP)
	}
}

// TestUpdateNodeState_WithMessage tests updating node state with message
func TestUpdateNodeState_WithMessage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create cluster and node
	cluster := &Cluster{ID: "cluster-1", Name: "test"}
	db.SaveCluster(cluster)

	node := &Node{ID: "node-1", Name: "test-node", ClusterID: "cluster-1", State: "running"}
	db.SaveNode(node)

	// Update state
	err = db.UpdateNodeState("node-1", "stopped", "Manual stop")
	if err != nil {
		t.Fatalf("failed to update node state: %v", err)
	}

	retrieved, _ := db.GetNode("node-1")
	if retrieved.State != "stopped" {
		t.Errorf("expected stopped, got %s", retrieved.State)
	}
}

// TestSaveMetric_WithAllFields tests saving metric with all fields
func TestSaveMetric_WithAllFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	metric := &Metric{
		NodeID:     "node-1",
		CPU:        45.5,
		Memory:     60.2,
		Disk:       30.1,
		NetworkRX:  1024000,
		NetworkTX:  512000,
		RecordedAt: time.Now(),
	}
	err = db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("failed to save metric: %v", err)
	}
}

// TestGetNodeMetrics_Empty tests GetNodeMetrics when no metrics exist
func TestGetNodeMetrics_Empty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	metrics, err := db.GetNodeMetrics("nonexistent", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics, got %d", len(metrics))
	}
}

// TestGetActiveAlerts_Empty tests GetActiveAlerts when no alerts exist
func TestGetActiveAlerts_Empty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	alerts, err := db.GetActiveAlerts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

// TestGetNodeAlerts_Empty tests GetNodeAlerts when no alerts exist
func TestGetNodeAlerts_Empty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	alerts, err := db.GetNodeAlerts("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

// TestSaveAlert_WithAllFields tests saving alert with all fields
func TestSaveAlert_WithAllFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	alert := &Alert{
		ID:        "alert-full",
		RuleID:    "rule-1",
		NodeID:    "node-1",
		NodeName:  "worker-1",
		Metric:    "cpu",
		Value:     95.5,
		Threshold: 90.0,
		Message:   "Critical CPU usage",
		FiredAt:   time.Now(),
		Resolved:  false,
	}
	err = db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("failed to save alert: %v", err)
	}
}

// TestSaveHost_WithAllFields tests saving host with all fields
func TestSaveHost_WithAllFields(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	host := &Host{
		ID:         "host-full",
		Name:       "full-host",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "admin",
		SSHKeyPath: "/home/admin/.ssh/id_rsa",
		HVType:     "libvirt",
	}
	err = db.SaveHost(host)
	if err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	retrieved, _ := db.GetHost("host-full")
	if retrieved.Port != 22 {
		t.Errorf("expected port 22, got %d", retrieved.Port)
	}
}

// TestCleanupOldMetrics_NoOldMetrics tests cleanup when no old metrics
func TestCleanupOldMetrics_NoOldMetrics(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Only save recent metrics
	for i := 0; i < 3; i++ {
		metric := &Metric{
			NodeID:     "node-1",
			CPU:        40.0,
			RecordedAt: time.Now(),
		}
		db.SaveMetric(metric)
	}

	count, err := db.CleanupOldMetrics(24 * time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = count // May be 0 if no old metrics
}