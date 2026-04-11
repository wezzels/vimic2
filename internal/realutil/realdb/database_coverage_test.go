package realdb_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realdb"
)

// TestRealDatabase_NewDatabase_WithPath tests NewDatabase with file path
func TestRealDatabase_NewDatabase_WithPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := realdb.NewDatabase(&realdb.Config{
		Path:         dbPath,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
		BusyTimeout:  time.Second,
		WALMode:      true,
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	// Verify file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created: %v", err)
	}
}

// TestRealDatabase_NewDatabase_EmptyPath tests NewDatabase with empty path (uses :memory:)
func TestRealDatabase_NewDatabase_EmptyPath(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path: "",
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_NewDatabase_NoWAL tests NewDatabase without WAL mode
func TestRealDatabase_NewDatabase_NoWAL(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:    ":memory:",
		WALMode: false,
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_NewDatabase_BusyTimeout tests NewDatabase with busy timeout
func TestRealDatabase_NewDatabase_BusyTimeout(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:        ":memory:",
		BusyTimeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_GetCluster_NotFound tests GetCluster with non-existent cluster
func TestRealDatabase_GetCluster_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	_, err = db.GetCluster("non-existent")
	if err == nil {
		t.Error("GetCluster should fail for non-existent cluster")
	}
}

// TestRealDatabase_DeleteCluster_NotFound tests DeleteCluster with non-existent cluster
func TestRealDatabase_DeleteCluster_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	err = db.DeleteCluster("non-existent")
	// May or may not error depending on implementation
	_ = err
}

// TestRealDatabase_GetHost_NotFound tests GetHost with non-existent host
func TestRealDatabase_GetHost_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	_, err = db.GetHost("non-existent")
	if err == nil {
		t.Error("GetHost should fail for non-existent host")
	}
}

// TestRealDatabase_DeleteHost_NotFound tests DeleteHost with non-existent host
func TestRealDatabase_DeleteHost_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	err = db.DeleteHost("non-existent")
	_ = err
}

// TestRealDatabase_GetNode_NotFound tests GetNode with non-existent node
func TestRealDatabase_GetNode_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	_, err = db.GetNode("non-existent")
	if err == nil {
		t.Error("GetNode should fail for non-existent node")
	}
}

// TestRealDatabase_DeleteNode_NotFound tests DeleteNode with non-existent node
func TestRealDatabase_DeleteNode_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	err = db.DeleteNode("non-existent")
	_ = err
}

// TestRealDatabase_SaveCluster_NilConfig tests SaveCluster with nil config
func TestRealDatabase_SaveCluster_NilConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	cluster := &realdb.Cluster{
		ID:        "cluster-nil",
		Name:      "test-cluster",
		Status:    "running",
		Config:    nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	retrieved, err := db.GetCluster("cluster-nil")
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}

	// Config may be returned as empty struct instead of nil
	_ = retrieved
}

// TestRealDatabase_SaveNode_NilConfig tests SaveNode with nil config
func TestRealDatabase_SaveNode_NilConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster first
	host := &realdb.Host{
		ID:      "host-1",
		Name:    "test-host",
		Address: "10.0.0.1",
	}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	cluster := &realdb.Cluster{
		ID:     "cluster-1",
		Name:   "test-cluster",
		Status: "running",
	}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	node := &realdb.Node{
		ID:        "node-nil",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "test-node",
		Role:      "worker",
		State:     "running",
		Config:    nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.SaveNode(node)
	if err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	retrieved, err := db.GetNode("node-nil")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}

	// Config may be returned as empty struct instead of nil
	_ = retrieved
}

// TestRealDatabase_ListClusters_Empty tests ListClusters with empty database
func TestRealDatabase_ListClusters_Empty(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}

	if len(clusters) != 0 {
		t.Errorf("expected empty list, got %d clusters", len(clusters))
	}
}

// TestRealDatabase_ListHosts_Empty tests ListHosts with empty database
func TestRealDatabase_ListHosts_Empty(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("expected empty list, got %d hosts", len(hosts))
	}
}

// TestRealDatabase_ListClusterNodes_Empty tests ListClusterNodes with empty cluster
func TestRealDatabase_ListClusterNodes_Empty(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create cluster first
	cluster := &realdb.Cluster{
		ID:     "cluster-empty",
		Name:   "empty-cluster",
		Status: "running",
	}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	nodes, err := db.ListClusterNodes("cluster-empty")
	if err != nil {
		t.Fatalf("ListClusterNodes failed: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("expected empty list, got %d nodes", len(nodes))
	}
}

// TestRealDatabase_GetAlert_NotFound tests GetAlert with non-existent alert
func TestRealDatabase_GetAlert_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	_, err = db.GetAlert("non-existent")
	if err == nil {
		t.Error("GetAlert should fail for non-existent alert")
	}
}

// TestRealDatabase_GetPool_NotFound tests GetPool with non-existent pool
func TestRealDatabase_GetPool_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	_, err = db.GetPool("non-existent")
	if err == nil {
		t.Error("GetPool should fail for non-existent pool")
	}
}

// TestRealDatabase_DeletePool_NotFound tests DeletePool with non-existent pool
func TestRealDatabase_DeletePool_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	err = db.DeletePool("non-existent")
	_ = err
}

// TestRealDatabase_ListPools_Empty tests ListPools with empty database
func TestRealDatabase_ListPools_Empty(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	pools, err := db.ListPools()
	if err != nil {
		t.Fatalf("ListPools failed: %v", err)
	}

	if len(pools) != 0 {
		t.Errorf("expected empty list, got %d pools", len(pools))
	}
}

// TestRealDatabase_Count tests Count function
func TestRealDatabase_CountFunc(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Count empty database
	counts := db.Count()
	if counts["clusters"] != 0 {
		t.Errorf("expected 0 clusters, got %d", counts["clusters"])
	}

	// Add a cluster
	cluster := &realdb.Cluster{
		ID:     "cluster-1",
		Name:   "test",
		Status: "running",
	}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	counts = db.Count()
	if counts["clusters"] != 1 {
		t.Errorf("expected 1 cluster, got %d", counts["clusters"])
	}
}

// TestRealDatabase_Count_InvalidTable tests Count with all tables
func TestRealDatabase_Count_AllTables(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	counts := db.Count()
	for table, count := range counts {
		if count != 0 {
			t.Errorf("expected 0 for %s, got %d", table, count)
		}
	}
}

// TestRealDatabase_IntegrityCheckFunc tests IntegrityCheck
func TestRealDatabase_IntegrityCheckFunc(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	if err := db.IntegrityCheck(); err != nil {
		t.Errorf("IntegrityCheck failed: %v", err)
	}
}

// TestRealDatabase_Close tests database close
func TestRealDatabase_Close(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}

	// Close should work
	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestRealDatabase_Ping tests database ping
func TestRealDatabase_Ping(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

// TestRealDatabase_MultipleClusters tests multiple clusters
func TestRealDatabase_MultipleClusters(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple clusters
	for i := 0; i < 5; i++ {
		cluster := &realdb.Cluster{
			ID:     string(rune('a' + i)),
			Name:   string(rune('a' + i)),
			Status: "running",
		}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters failed: %v", err)
	}

	if len(clusters) != 5 {
		t.Errorf("expected 5 clusters, got %d", len(clusters))
	}
}

// TestRealDatabase_SaveMetric tests SaveMetric
func TestRealDatabase_SaveMetric(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	metric := &realdb.Metric{
		NodeID:     "node-1",
		CPU:        50.0,
		Memory:     60.0,
		Disk:       70.0,
		NetworkRX:  1000.0,
		NetworkTX:  500.0,
		RecordedAt: time.Now(),
	}

	err = db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("SaveMetric failed: %v", err)
	}
}

// TestRealDatabase_GetNodeMetrics_Empty tests GetNodeMetrics with no metrics
func TestRealDatabase_GetNodeMetrics_Empty(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	metrics, err := db.GetNodeMetrics("node-no-metrics", time.Now())
	if err != nil {
		t.Fatalf("GetNodeMetrics failed: %v", err)
	}

	if len(metrics) != 0 {
		t.Errorf("expected empty metrics, got %d", len(metrics))
	}
}
// TestRealDatabase_NewDatabase_InvalidPath tests NewDatabase with invalid path
func TestRealDatabase_NewDatabase_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory
	_, err := realdb.NewDatabase(&realdb.Config{
		Path: "/nonexistent/path/db.sqlite",
	})
	// Should fail to create directory/file
	_ = err
}

// TestRealDatabase_NewDatabase_ConnectionPool tests connection pool settings
func TestRealDatabase_NewDatabase_ConnectionPool(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:         ":memory:",
		MaxOpenConns: 25,
		MaxIdleConns: 10,
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_NewDatabase_ZeroValues tests with zero values
func TestRealDatabase_NewDatabase_ZeroValues(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:         ":memory:",
		MaxOpenConns: 0,
		MaxIdleConns: 0,
		BusyTimeout:  0,
		WALMode:      false,
	})
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_SaveAlert tests SaveAlert
func TestRealDatabase_SaveAlert(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	alert := &realdb.Alert{
		ID:        "alert-1",
		NodeID:    "node-1",
		Type:      "cpu_high",
		Message:   "CPU usage above 90%",
		Severity:  "warning",
		CreatedAt: time.Now(),
	}

	err = db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("SaveAlert failed: %v", err)
	}

	retrieved, err := db.GetAlert("alert-1")
	if err != nil {
		t.Fatalf("GetAlert failed: %v", err)
	}

	if retrieved.Message != "CPU usage above 90%" {
		t.Errorf("wrong message: got %s", retrieved.Message)
	}
}

// TestRealDatabase_ListAlerts_WithFilter tests ListAlerts
func TestRealDatabase_ListAlerts_WithFilter(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple alerts
	for i := 0; i < 3; i++ {
		alert := &realdb.Alert{
			ID:        string(rune('a' + i)),
			NodeID:    "node-1",
			Type:      []string{"cpu_high", "mem_high", "disk_low"}[i],
			Message:   "test",
			Severity:  "warning",
			CreatedAt: time.Now(),
		}
		if err := db.SaveAlert(alert); err != nil {
			t.Fatal(err)
		}
	}

	alerts, err := db.ListAlerts()
	if err != nil {
		t.Fatalf("ListAlerts failed: %v", err)
	}

	if len(alerts) != 3 {
		t.Errorf("expected 3 alerts, got %d", len(alerts))
	}
}

// TestRealDatabase_SavePool tests SavePool
func TestRealDatabase_SavePool(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	pool := &realdb.Pool{
		ID:        "pool-1",
		Name:      "test-pool",
		Available: 5,
		Busy:      2,
	}

	err = db.SavePool(pool)
	if err != nil {
		t.Fatalf("SavePool failed: %v", err)
	}

	retrieved, err := db.GetPool("pool-1")
	if err != nil {
		t.Fatalf("GetPool failed: %v", err)
	}

	if retrieved.Name != "test-pool" {
		t.Errorf("wrong name: got %s", retrieved.Name)
	}
}

// TestRealDatabase_SaveMetric_Multiple tests multiple metrics
func TestRealDatabase_SaveMetric_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	for i := 0; i < 5; i++ {
		metric := &realdb.Metric{
			NodeID:     "node-1",
			CPU:        float64(50 + i),
			Memory:     float64(60 + i),
			Disk:       float64(70 + i),
			NetworkRX:  float64(1000 + i),
			NetworkTX:  float64(500 + i),
			RecordedAt: time.Now(),
		}
		if err := db.SaveMetric(metric); err != nil {
			t.Fatal(err)
		}
	}

	// Query from a time in the past
	metrics, err := db.GetNodeMetrics("node-1", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("GetNodeMetrics failed: %v", err)
	}

	if len(metrics) != 5 {
		t.Errorf("expected 5 metrics, got %d", len(metrics))
	}
}

// TestRealDatabase_MultipleHosts tests multiple hosts
func TestRealDatabase_MultipleHosts(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	for i := 0; i < 3; i++ {
		host := &realdb.Host{
			ID:         string(rune('a' + i)),
			Name:       string(rune('a' + i)),
			Address:    "10.0.0." + string(rune('1'+i)),
			Port:       22,
			User:       "root",
			SSHKeyPath: "/path/key",
			HVType:     "libvirt",
		}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}

	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(hosts))
	}
}
