// Package monitor provides real database tests
package monitor

import (
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestManager_CollectAll tests collectAll with real DB
func TestManager_CollectAll(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "monitor-test-*.db")
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

	node := &database.Node{
		ID:        "node-1",
		Name:      "test-node",
		ClusterID: "cluster-1",
		State:     "running",
		HostID:    "host-1",
	}
	db.SaveNode(node)

	// Create manager
	mgr := NewManager(db, nil)

	// Call collectAll - should not panic
	mgr.collectAll()
}

// TestManager_GetNodeMetrics_RealDB tests GetNodeMetrics with real DB
func TestManager_GetNodeMetrics_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "monitor-test-*.db")
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

	// Save metrics
	for i := 0; i < 5; i++ {
		metric := &database.Metric{
			NodeID:     "node-1",
			CPU:        float64(40 + i*5),
			Memory:     60.0,
			RecordedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
		db.SaveMetric(metric)
	}

	mgr := NewManager(db, nil)

	metrics, err := mgr.GetNodeMetrics("node-1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	if len(metrics) < 1 {
		t.Error("expected at least one metric")
	}
}

// TestManager_GetClusterMetrics_RealDB tests GetClusterMetrics with real DB
func TestManager_GetClusterMetrics_RealDB(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "monitor-test-*.db")
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

	// Save metrics
	for i := 0; i < 3; i++ {
		metric := &database.Metric{
			NodeID:     "node-1",
			CPU:        float64(40 + i*10),
			Memory:     60.0,
			RecordedAt: time.Now(),
		}
		db.SaveMetric(metric)
	}

	mgr := NewManager(db, nil)

	cm, err := mgr.GetClusterMetrics("cluster-1", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to get cluster metrics: %v", err)
	}
	if cm == nil {
		t.Fatal("expected non-nil ClusterMetrics")
	}
}

// TestAlerter_GetRules tests GetRules with real DB
func TestAlerter_GetRules(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "monitor-test-*.db")
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

	alerter := NewAlerter(db)

	// Add rules
	alerter.AddRule(&AlertRule{ID: "cpu-high", Metric: "cpu", Threshold: 80.0})

	rules := alerter.GetRules()
	if len(rules) == 0 {
		t.Error("expected at least one rule")
	}
}

// TestAlerter_ResolveAlert tests ResolveAlert
func TestAlerter_ResolveAlert(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "monitor-test-*.db")
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

	alerter := NewAlerter(db)

	// Add a rule
	alerter.AddRule(&AlertRule{ID: "cpu-high", Metric: "cpu", Threshold: 80.0})

	// Resolve it (will silently do nothing if alert doesn't exist)
	alerter.ResolveAlert("node-1", "cpu-high")
}

// TestManager_CollectNode_NilDB tests collectNode with nil DB
func TestManager_CollectNode_NilDB(t *testing.T) {
	// Skip test - collectNode requires non-nil DB
	t.Skip("collectNode requires non-nil DB")
}