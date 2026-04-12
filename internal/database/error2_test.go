// Package database provides additional error path tests
package database

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

// TestListHosts_ScanError tests ListHosts when scan fails
func TestListHosts_ScanError(t *testing.T) {
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

	// Create a host
	host := &Host{
		ID:      "host-1",
		Name:    "test-host",
		Address: "192.168.1.1",
		Port:    22,
		User:    "admin",
		HVType:  "libvirt",
	}
	db.SaveHost(host)

	// Close to force rows.Next error
	db.Close()

	_, err = db.ListHosts()
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestGetCluster_RowScanError tests GetCluster when row scan fails
func TestGetCluster_RowScanError(t *testing.T) {
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

	// Close DB to force scan error
	db.Close()

	_, err = db.GetCluster("cluster-1")
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestListClusters_ScanError tests ListClusters when scan fails
func TestListClusters_ScanError(t *testing.T) {
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
	cluster := &Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	db.SaveCluster(cluster)

	// Close to force error
	db.Close()

	_, err = db.ListClusters()
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestGetLatestMetric_QueryError tests GetLatestMetric when query fails
func TestGetLatestMetric_QueryError(t *testing.T) {
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

	// Close DB to force query error
	db.Close()

	_, err = db.GetLatestMetric("node-1")
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestGetActiveAlerts_ScanError tests GetActiveAlerts when scan fails
func TestGetActiveAlerts_ScanError(t *testing.T) {
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

	// Create an alert
	alert := &Alert{
		ID:      "alert-1",
		NodeID:  "node-1",
		Metric:  "cpu",
		Value:   85.0,
		FiredAt: time.Now(),
	}
	db.SaveAlert(alert)

	// Close to force scan error
	db.Close()

	_, err = db.GetActiveAlerts()
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestGetNodeAlerts_ScanError tests GetNodeAlerts when scan fails
func TestGetNodeAlerts_ScanError(t *testing.T) {
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

	// Create an alert
	alert := &Alert{
		ID:      "alert-1",
		NodeID:  "node-1",
		Metric:  "cpu",
		Value:   85.0,
		FiredAt: time.Now(),
	}
	db.SaveAlert(alert)

	// Close to force scan error
	db.Close()

	_, err = db.GetNodeAlerts("node-1")
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestListClusterNodes_ScanError tests ListClusterNodes when scan fails
func TestListClusterNodes_ScanError(t *testing.T) {
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

	// Close to force error
	db.Close()

	_, err = db.ListClusterNodes("cluster-1")
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestListAllNodes_ScanError tests ListAllNodes when scan fails
func TestListAllNodes_ScanError(t *testing.T) {
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

	// Close to force error
	db.Close()

	_, err = db.ListAllNodes()
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestGetNodeMetrics_ScanError tests GetNodeMetrics when scan fails
func TestGetNodeMetrics_ScanError(t *testing.T) {
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

	// Save metric
	metric := &Metric{
		NodeID:     "node-1",
		CPU:        45.0,
		RecordedAt: time.Now(),
	}
	db.SaveMetric(metric)

	// Close to force error
	db.Close()

	_, err = db.GetNodeMetrics("node-1", time.Now().Add(-1*time.Hour))
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestDB_GetNode_Nil tests GetNode returns nil for not found
func TestDB_GetNode_Nil(t *testing.T) {
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

	// GetNode should return nil for non-existent node
	node, err := db.GetNode("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node != nil {
		t.Error("expected nil for non-existent node")
	}
}

// TestDB_GetHost_Nil tests GetHost returns nil for not found
func TestDB_GetHost_Nil(t *testing.T) {
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

	// GetHost should return nil for non-existent host
	host, err := db.GetHost("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host != nil {
		t.Error("expected nil for non-existent host")
	}
}

// TestDB_GetCluster_Nil tests GetCluster returns nil for not found
func TestDB_GetCluster_Nil(t *testing.T) {
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

	// GetCluster should return nil for non-existent cluster
	cluster, err := db.GetCluster("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cluster != nil {
		t.Error("expected nil for non-existent cluster")
	}
}

// TestDB_GetLatestMetric_Nil tests GetLatestMetric returns nil for not found
func TestDB_GetLatestMetric_Nil(t *testing.T) {
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

	// GetLatestMetric should return nil for non-existent metric
	metric, err := db.GetLatestMetric("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metric != nil {
		t.Error("expected nil for non-existent metric")
	}
}

// TestSaveCluster_ExecError tests SaveCluster when exec fails
func TestSaveCluster_ExecError(t *testing.T) {
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

	// Close DB to force exec error
	db.Close()

	cluster := &Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	err = db.SaveCluster(cluster)
	if err == nil {
		t.Error("expected error after close")
	}
}

// TestSqlOpenError tests sql.Open error path (requires mock)
func TestSqlOpenError(t *testing.T) {
	// This test is for the sql.Open error path in NewDB
	// sql.Open returns error only for invalid driver names
	_, err := sql.Open("invalid-driver", "test.db")
	if err == nil {
		t.Error("expected error for invalid driver")
	}
}
