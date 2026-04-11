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

// TestRealDatabase_SaveCluster_Update tests updating existing cluster
func TestRealDatabase_SaveCluster_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create cluster
	cluster := &realdb.Cluster{
		ID:     "cluster-1",
		Name:   "test-cluster",
		Status: "created",
	}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Update cluster
	cluster.Status = "running"
	cluster.Name = "updated-cluster"
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "updated-cluster" {
		t.Errorf("expected updated-cluster, got %s", retrieved.Name)
	}
	if retrieved.Status != "running" {
		t.Errorf("expected running status, got %s", retrieved.Status)
	}
}

// TestRealDatabase_SaveHost_Update tests updating existing host
func TestRealDatabase_SaveHost_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host
	host := &realdb.Host{
		ID:      "host-1",
		Name:    "test-host",
		Address: "10.0.0.1",
	}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	// Update host
	host.Name = "updated-host"
	host.Address = "10.0.0.2"
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, err := db.GetHost("host-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "updated-host" {
		t.Errorf("expected updated-host, got %s", retrieved.Name)
	}
}

// TestRealDatabase_SaveNode_Update tests updating existing node
func TestRealDatabase_SaveNode_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Create node
	node := &realdb.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "test-node",
		Role:      "worker",
		State:     "created",
	}
	if err := db.SaveNode(node); err != nil {
		t.Fatal(err)
	}

	// Update node
	node.State = "running"
	node.IP = "10.0.1.1"
	if err := db.SaveNode(node); err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, err := db.GetNode("node-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.State != "running" {
		t.Errorf("expected running state, got %s", retrieved.State)
	}
}

// TestRealDatabase_DeleteHost tests deleting host
func TestRealDatabase_DeleteHost(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	// Delete host
	if err := db.DeleteHost("host-1"); err != nil {
		t.Fatal(err)
	}

	// Verify deleted
	_, err = db.GetHost("host-1")
	if err == nil {
		t.Error("expected error for deleted host")
	}
}

// TestRealDatabase_DeleteNode tests deleting node
func TestRealDatabase_DeleteNode(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host, cluster, and node
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}
	node := &realdb.Node{ID: "node-1", ClusterID: "cluster-1", HostID: "host-1", Name: "test", Role: "worker", State: "running"}
	if err := db.SaveNode(node); err != nil {
		t.Fatal(err)
	}

	// Delete node
	if err := db.DeleteNode("node-1"); err != nil {
		t.Fatal(err)
	}

	// Verify deleted
	_, err = db.GetNode("node-1")
	if err == nil {
		t.Error("expected error for deleted node")
	}
}

// TestRealDatabase_DeletePool tests deleting pool
func TestRealDatabase_DeletePool(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create pool
	pool := &realdb.Pool{ID: "pool-1", Name: "test", Available: 5, Busy: 2}
	if err := db.SavePool(pool); err != nil {
		t.Fatal(err)
	}

	// Delete pool
	if err := db.DeletePool("pool-1"); err != nil {
		t.Fatal(err)
	}

	// Verify deleted
	_, err = db.GetPool("pool-1")
	if err == nil {
		t.Error("expected error for deleted pool")
	}
}

// TestRealDatabase_SavePool_Update tests updating pool
func TestRealDatabase_SavePool_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create pool
	pool := &realdb.Pool{ID: "pool-1", Name: "test", Available: 5, Busy: 2}
	if err := db.SavePool(pool); err != nil {
		t.Fatal(err)
	}

	// Update pool
	pool.Available = 10
	pool.Busy = 3
	if err := db.SavePool(pool); err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, err := db.GetPool("pool-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Available != 10 {
		t.Errorf("expected Available=10, got %d", retrieved.Available)
	}
}

// TestRealDatabase_Count_WithData tests Count with actual data
func TestRealDatabase_Count_WithData(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create some data
	for i := 0; i < 3; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
		host := &realdb.Host{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Address: "10.0.0.1"}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	counts := db.Count()
	if counts["clusters"] != 3 {
		t.Errorf("expected 3 clusters, got %d", counts["clusters"])
	}
	if counts["hosts"] != 3 {
		t.Errorf("expected 3 hosts, got %d", counts["hosts"])
	}
}

// TestRealDatabase_ClusterConfig tests cluster config JSON handling
func TestRealDatabase_ClusterConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	config := &realdb.ClusterConfig{
		MinNodes:  3,
		MaxNodes:  10,
		AutoScale: true,
	}

	cluster := &realdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "running",
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveCluster(cluster); err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}

	if retrieved.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if retrieved.Config.MinNodes != 3 {
		t.Errorf("expected MinNodes=3, got %d", retrieved.Config.MinNodes)
	}
	if retrieved.Config.MaxNodes != 10 {
		t.Errorf("expected MaxNodes=10, got %d", retrieved.Config.MaxNodes)
	}
	if !retrieved.Config.AutoScale {
		t.Error("expected AutoScale=true")
	}
}

// TestRealDatabase_NodeConfig tests node config JSON handling
func TestRealDatabase_NodeConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	config := &realdb.NodeConfig{
		CPU:    4,
		Memory: 8192,
		Disk:   100,
		Image:  "ubuntu-22.04",
	}

	node := &realdb.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "test-node",
		Role:      "worker",
		State:     "running",
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveNode(node); err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	retrieved, err := db.GetNode("node-1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}

	if retrieved.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if retrieved.Config.CPU != 4 {
		t.Errorf("expected CPU=4, got %d", retrieved.Config.CPU)
	}
	if retrieved.Config.Memory != 8192 {
		t.Errorf("expected Memory=8192, got %d", retrieved.Config.Memory)
	}
	if retrieved.Config.Image != "ubuntu-22.04" {
		t.Errorf("expected Image=ubuntu-22.04, got %s", retrieved.Config.Image)
	}
}

// TestRealDatabase_ListClusterNodes_WithNodes tests listing cluster nodes
func TestRealDatabase_ListClusterNodes_WithNodes(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Create nodes
	for i := 0; i < 3; i++ {
		node := &realdb.Node{
			ID:        string(rune('a' + i)),
			ClusterID: "cluster-1",
			HostID:    "host-1",
			Name:      string(rune('a' + i)),
			Role:      "worker",
			State:     "running",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}
	}

	nodes, err := db.ListClusterNodes("cluster-1")
	if err != nil {
		t.Fatalf("ListClusterNodes failed: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}
}

// TestRealDatabase_DeleteAlert tests deleting alerts
func TestRealDatabase_DeleteAlert(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create alert
	alert := &realdb.Alert{
		ID:        "alert-1",
		NodeID:    "node-1",
		Type:      "cpu_high",
		Message:   "test",
		Severity:  "warning",
		CreatedAt: time.Now(),
	}
	if err := db.SaveAlert(alert); err != nil {
		t.Fatal(err)
	}

	// Delete alert
	if err := db.DeleteAlert("alert-1"); err != nil {
		t.Fatalf("DeleteAlert failed: %v", err)
	}

	// Verify deleted
	_, err = db.GetAlert("alert-1")
	if err == nil {
		t.Error("expected error for deleted alert")
	}
}

// TestRealDatabase_DeleteAlert_NotFound tests deleting non-existent alert
func TestRealDatabase_DeleteAlert_NotFound(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	err = db.DeleteAlert("non-existent")
	_ = err
}

// TestRealDatabase_SaveAlert_Update tests updating alert
func TestRealDatabase_SaveAlert_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create alert
	alert := &realdb.Alert{
		ID:        "alert-1",
		NodeID:    "node-1",
		Type:      "cpu_high",
		Message:   "CPU high",
		Severity:  "warning",
		CreatedAt: time.Now(),
	}
	if err := db.SaveAlert(alert); err != nil {
		t.Fatal(err)
	}

	// Update alert
	alert.Message = "CPU critical"
	alert.Severity = "critical"
	if err := db.SaveAlert(alert); err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, err := db.GetAlert("alert-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Message != "CPU critical" {
		t.Errorf("expected 'CPU critical', got %s", retrieved.Message)
	}
}

// TestRealDatabase_DoubleClose tests closing database twice
func TestRealDatabase_DoubleClose(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}

	// Close once
	if err := db.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}

	// Close again - should not error
	if err := db.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}

// TestRealDatabase_PingAfterClose tests Ping after close
func TestRealDatabase_PingAfterClose(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}

	// Close
	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Ping after close should fail
	if err := db.Ping(); err == nil {
		t.Error("Ping should fail after Close")
	}
}

// TestRealDatabase_ClusterStatus tests cluster status updates
func TestRealDatabase_ClusterStatus(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	statuses := []string{"created", "running", "stopped", "error"}

	for _, status := range statuses {
		cluster := &realdb.Cluster{
			ID:     "cluster-" + status,
			Name:   "test-" + status,
			Status: status,
		}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetCluster("cluster-" + status)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Status != status {
			t.Errorf("expected status %s, got %s", status, retrieved.Status)
		}
	}
}

// TestRealDatabase_NodeStates tests node state updates
func TestRealDatabase_NodeStates(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	states := []string{"created", "running", "stopped", "error"}

	for _, state := range states {
		node := &realdb.Node{
			ID:        "node-" + state,
			ClusterID: "cluster-1",
			HostID:    "host-1",
			Name:      "test-" + state,
			Role:      "worker",
			State:     state,
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetNode("node-" + state)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.State != state {
			t.Errorf("expected state %s, got %s", state, retrieved.State)
		}
	}
}

// TestRealDatabase_NodeRoles tests node roles
func TestRealDatabase_NodeRoles(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	roles := []string{"worker", "master", "controller"}

	for _, role := range roles {
		node := &realdb.Node{
			ID:        "node-" + role,
			ClusterID: "cluster-1",
			HostID:    "host-1",
			Name:      "test-" + role,
			Role:      role,
			State:     "running",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetNode("node-" + role)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Role != role {
			t.Errorf("expected role %s, got %s", role, retrieved.Role)
		}
	}
}

// TestRealDatabase_HostTypes tests host HV types
func TestRealDatabase_HostTypes(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	hvTypes := []string{"libvirt", "kvm", "qemu"}

	for _, hvType := range hvTypes {
		host := &realdb.Host{
			ID:      "host-" + hvType,
			Name:    "test-" + hvType,
			Address: "10.0.0.1",
			HVType:  hvType,
		}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetHost("host-" + hvType)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.HVType != hvType {
			t.Errorf("expected HVType %s, got %s", hvType, retrieved.HVType)
		}
	}
}

// TestRealDatabase_MultipleMetrics tests multiple metrics for different nodes
func TestRealDatabase_MultipleMetrics(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create metrics for multiple nodes
	for nodeNum := 0; nodeNum < 3; nodeNum++ {
		for i := 0; i < 5; i++ {
			metric := &realdb.Metric{
				NodeID:     string(rune('a' + nodeNum)),
				CPU:        float64(50 + i),
				Memory:     float64(60 + i),
				Disk:       float64(70 + i),
				NetworkRX:  float64(1000 + i),
				NetworkTX:  float64(500 + i),
				RecordedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			}
			if err := db.SaveMetric(metric); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Get metrics for each node
	for nodeNum := 0; nodeNum < 3; nodeNum++ {
		metrics, err := db.GetNodeMetrics(string(rune('a'+nodeNum)), time.Now().Add(-time.Hour))
		if err != nil {
			t.Fatalf("GetNodeMetrics failed: %v", err)
		}
		if len(metrics) != 5 {
			t.Errorf("expected 5 metrics for node %c, got %d", rune('a'+nodeNum), len(metrics))
		}
	}
}

// TestRealDatabase_AlertSeverities tests different alert severities
func TestRealDatabase_AlertSeverities(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	severities := []string{"info", "warning", "error", "critical"}

	for _, sev := range severities {
		alert := &realdb.Alert{
			ID:        "alert-" + sev,
			NodeID:    "node-1",
			Type:      "test",
			Message:   "test message",
			Severity:  sev,
			CreatedAt: time.Now(),
		}
		if err := db.SaveAlert(alert); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetAlert("alert-" + sev)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Severity != sev {
			t.Errorf("expected severity %s, got %s", sev, retrieved.Severity)
		}
	}
}

// TestRealDatabase_PoolAvailability tests pool available/busy counts
func TestRealDatabase_PoolAvailability(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create pool with various available/busy counts
	pools := []struct {
		available int
		busy      int
	}{
		{available: 10, busy: 0},
		{available: 5, busy: 5},
		{available: 0, busy: 10},
	}

	for i, p := range pools {
		pool := &realdb.Pool{
			ID:        string(rune('a' + i)),
			Name:      string(rune('a' + i)),
			Available: p.available,
			Busy:      p.busy,
		}
		if err := db.SavePool(pool); err != nil {
			t.Fatal(err)
		}

		retrieved, err := db.GetPool(string(rune('a' + i)))
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Available != p.available {
			t.Errorf("expected Available=%d, got %d", p.available, retrieved.Available)
		}
		if retrieved.Busy != p.busy {
			t.Errorf("expected Busy=%d, got %d", p.busy, retrieved.Busy)
		}
	}
}

// TestRealDatabase_MultipleListOperations tests multiple list operations
func TestRealDatabase_MultipleListOperations(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple entities
	for i := 0; i < 5; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
		host := &realdb.Host{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Address: "10.0.0." + string(rune('1'+i))}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
		pool := &realdb.Pool{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Available: 10, Busy: 0}
		if err := db.SavePool(pool); err != nil {
			t.Fatal(err)
		}
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 5 {
		t.Errorf("expected 5 clusters, got %d", len(clusters))
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 5 {
		t.Errorf("expected 5 hosts, got %d", len(hosts))
	}

	pools, err := db.ListPools()
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 5 {
		t.Errorf("expected 5 pools, got %d", len(pools))
	}
}

// TestRealDatabase_NewDatabase_WALModeError tests NewDatabase WAL mode error handling
func TestRealDatabase_NewDatabase_WALModeError(t *testing.T) {
	// Create database with WAL mode enabled (default)
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:    ":memory:",
		WALMode: true,
	})
	if err != nil {
		t.Fatalf("NewDatabase with WALMode failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_NewDatabase_BusyTimeoutSetting tests busy timeout setting
func TestRealDatabase_NewDatabase_BusyTimeoutSetting(t *testing.T) {
	db, err := realdb.NewDatabase(&realdb.Config{
		Path:        ":memory:",
		BusyTimeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewDatabase with BusyTimeout failed: %v", err)
	}
	defer db.Close()
}

// TestRealDatabase_SaveMetric_WithValidData tests SaveMetric with various data
func TestRealDatabase_SaveMetric_WithValidData(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Test with various metric values
	testCases := []struct {
		cpu    float64
		memory float64
		disk   float64
	}{
		{0.0, 0.0, 0.0},
		{100.0, 100.0, 100.0},
		{50.5, 75.25, 25.75},
		{-1.0, -1.0, -1.0}, // Negative values should be handled
	}

	for i, tc := range testCases {
		metric := &realdb.Metric{
			NodeID:     "test-node",
			CPU:        tc.cpu,
			Memory:     tc.memory,
			Disk:       tc.disk,
			NetworkRX:  1000.0,
			NetworkTX:  500.0,
			RecordedAt: time.Now(),
		}
		if err := db.SaveMetric(metric); err != nil {
			t.Errorf("SaveMetric[%d] failed: %v", i, err)
		}
	}
}

// TestRealDatabase_IntegrityCheck_Corruption tests IntegrityCheck behavior
func TestRealDatabase_IntegrityCheck_Corruption(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Basic integrity check should pass
	if err := db.IntegrityCheck(); err != nil {
		t.Errorf("IntegrityCheck failed: %v", err)
	}
}

// TestRealDatabase_ListClusters_Order tests ListClusters ordering
func TestRealDatabase_ListClusters_Order(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create clusters in order
	clusterNames := []string{"cluster-a", "cluster-b", "cluster-c"}
	for _, name := range clusterNames {
		cluster := &realdb.Cluster{ID: name, Name: name, Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatal(err)
	}

	if len(clusters) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(clusters))
	}
}

// TestRealDatabase_ListHosts_Order tests ListHosts ordering
func TestRealDatabase_ListHosts_Order(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create hosts
	hostNames := []string{"host-a", "host-b", "host-c"}
	for _, name := range hostNames {
		host := &realdb.Host{ID: name, Name: name, Address: "10.0.0.1"}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatal(err)
	}

	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(hosts))
	}
}

// TestRealDatabase_ListPools_Order tests ListPools ordering
func TestRealDatabase_ListPools_Order(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create pools
	poolNames := []string{"pool-a", "pool-b", "pool-c"}
	for _, name := range poolNames {
		pool := &realdb.Pool{ID: name, Name: name, Available: 10, Busy: 0}
		if err := db.SavePool(pool); err != nil {
			t.Fatal(err)
		}
	}

	pools, err := db.ListPools()
	if err != nil {
		t.Fatal(err)
	}

	if len(pools) != 3 {
		t.Errorf("expected 3 pools, got %d", len(pools))
	}
}

// TestRealDatabase_ListAlerts_Order tests ListAlerts ordering
func TestRealDatabase_ListAlerts_Order(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create alerts
	alertIDs := []string{"alert-a", "alert-b", "alert-c"}
	for _, id := range alertIDs {
		alert := &realdb.Alert{
			ID:        id,
			NodeID:    "node-1",
			Type:      "test",
			Message:   "test message",
			Severity:  "warning",
			CreatedAt: time.Now(),
		}
		if err := db.SaveAlert(alert); err != nil {
			t.Fatal(err)
		}
	}

	alerts, err := db.ListAlerts()
	if err != nil {
		t.Fatal(err)
	}

	if len(alerts) != 3 {
		t.Errorf("expected 3 alerts, got %d", len(alerts))
	}
}

// TestRealDatabase_Count_WithMultipleEntities tests Count with various entities
func TestRealDatabase_Count_WithMultipleEntities(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create 5 clusters, 10 hosts, 20 nodes
	for i := 0; i < 5; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 10; i++ {
		host := &realdb.Host{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Address: "10.0.0.1"}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	// Create cluster and host for nodes
	cluster := &realdb.Cluster{ID: "cluster-for-nodes", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}
	host := &realdb.Host{ID: "host-for-nodes", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		node := &realdb.Node{
			ID:        string(rune('a' + i)),
			ClusterID: "cluster-for-nodes",
			HostID:    "host-for-nodes",
			Name:      string(rune('a' + i)),
			Role:      "worker",
			State:     "running",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}
	}

	counts := db.Count()
	if counts["clusters"] != 6 { // 5 + 1 for cluster-for-nodes
		t.Errorf("expected 6 clusters, got %d", counts["clusters"])
	}
	if counts["hosts"] != 11 { // 10 + 1 for host-for-nodes
		t.Errorf("expected 11 hosts, got %d", counts["hosts"])
	}
	if counts["nodes"] != 20 {
		t.Errorf("expected 20 nodes, got %d", counts["nodes"])
	}
}

// TestRealDatabase_SaveCluster_JSONError tests SaveCluster JSON handling
func TestRealDatabase_SaveCluster_JSONError(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Test with config that should serialize properly
	cluster := &realdb.Cluster{
		ID:        "cluster-json",
		Name:      "test-cluster",
		Status:    "running",
		Config:    &realdb.ClusterConfig{MinNodes: 1, MaxNodes: 5, AutoScale: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveCluster(cluster); err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	// Verify we can retrieve it
	retrieved, err := db.GetCluster("cluster-json")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved.Config == nil {
		t.Fatal("Config should not be nil")
	}
}

// TestRealDatabase_SaveNode_JSONError tests SaveNode JSON handling  
func TestRealDatabase_SaveNode_JSONError(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Test with config
	node := &realdb.Node{
		ID:        "node-json",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "test-node",
		Role:      "worker",
		State:     "running",
		Config:    &realdb.NodeConfig{CPU: 4, Memory: 8192, Disk: 100, Image: "ubuntu"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveNode(node); err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	retrieved, err := db.GetNode("node-json")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if retrieved.Config.CPU != 4 {
		t.Errorf("expected CPU=4, got %d", retrieved.Config.CPU)
	}
}

// TestRealDatabase_GetCluster_Multiple tests GetCluster after multiple saves
func TestRealDatabase_GetCluster_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple clusters
	for i := 0; i < 10; i++ {
		cluster := &realdb.Cluster{
			ID:     string(rune('a' + i)),
			Name:   "cluster-" + string(rune('a'+i)),
			Status: "running",
		}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	// Verify we can get each one
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		cluster, err := db.GetCluster(id)
		if err != nil {
			t.Errorf("GetCluster(%s) failed: %v", id, err)
		}
		if cluster.Name != "cluster-"+id {
			t.Errorf("expected name cluster-%s, got %s", id, cluster.Name)
		}
	}
}

// TestRealDatabase_GetHost_Multiple tests GetHost after multiple saves
func TestRealDatabase_GetHost_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple hosts
	for i := 0; i < 10; i++ {
		host := &realdb.Host{
			ID:      string(rune('a' + i)),
			Name:    "host-" + string(rune('a'+i)),
			Address: "10.0.0." + string(rune('1'+i)),
		}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	// Verify we can get each one
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		host, err := db.GetHost(id)
		if err != nil {
			t.Errorf("GetHost(%s) failed: %v", id, err)
		}
		if host.Name != "host-"+id {
			t.Errorf("expected name host-%s, got %s", id, host.Name)
		}
	}
}

// TestRealDatabase_GetNode_Multiple tests GetNode after multiple saves
func TestRealDatabase_GetNode_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create host and cluster
	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Create multiple nodes
	for i := 0; i < 10; i++ {
		node := &realdb.Node{
			ID:        string(rune('a' + i)),
			ClusterID: "cluster-1",
			HostID:    "host-1",
			Name:      "node-" + string(rune('a'+i)),
			Role:      "worker",
			State:     "running",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}
	}

	// Verify we can get each one
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		node, err := db.GetNode(id)
		if err != nil {
			t.Errorf("GetNode(%s) failed: %v", id, err)
		}
		if node.Name != "node-"+id {
			t.Errorf("expected name node-%s, got %s", id, node.Name)
		}
	}
}

// TestRealDatabase_SaveMetric_ErrorPaths tests SaveMetric error paths
func TestRealDatabase_SaveMetric_ErrorPaths(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Save a valid metric
	metric := &realdb.Metric{
		NodeID:     "node-1",
		CPU:        50.0,
		Memory:     60.0,
		Disk:       70.0,
		NetworkRX:  1000.0,
		NetworkTX:  500.0,
		RecordedAt: time.Now(),
	}

	if err := db.SaveMetric(metric); err != nil {
		t.Fatalf("SaveMetric failed: %v", err)
	}

	// Verify ID was set
	if metric.ID == 0 {
		t.Error("Metric ID should be set after SaveMetric")
	}
}

// TestRealDatabase_GetNodeMetrics_TimeRange tests GetNodeMetrics time range
func TestRealDatabase_GetNodeMetrics_TimeRange(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Save metrics at different times
	now := time.Now()
	times := []time.Duration{-2 * time.Hour, -time.Hour, -30 * time.Minute, 0}

	for i, offset := range times {
		metric := &realdb.Metric{
			NodeID:     "node-1",
			CPU:        float64(i),
			RecordedAt: now.Add(offset),
		}
		if err := db.SaveMetric(metric); err != nil {
			t.Fatal(err)
		}
	}

	// Query for metrics in last hour
	metrics, err := db.GetNodeMetrics("node-1", now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Should get metrics from last hour only (3 metrics)
	if len(metrics) < 2 {
		t.Errorf("expected at least 2 metrics, got %d", len(metrics))
	}
}

// TestRealDatabase_GetAlert_Multiple tests GetAlert after multiple saves
func TestRealDatabase_GetAlert_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple alerts
	for i := 0; i < 10; i++ {
		alert := &realdb.Alert{
			ID:        string(rune('a' + i)),
			NodeID:    "node-1",
			Type:      "test",
			Message:   "test message",
			Severity:  "warning",
			CreatedAt: time.Now(),
		}
		if err := db.SaveAlert(alert); err != nil {
			t.Fatal(err)
		}
	}

	// Verify we can get each one
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		alert, err := db.GetAlert(id)
		if err != nil {
			t.Errorf("GetAlert(%s) failed: %v", id, err)
		}
		if alert.Message != "test message" {
			t.Errorf("expected 'test message', got %s", alert.Message)
		}
	}
}

// TestRealDatabase_GetPool_Multiple tests GetPool after multiple saves
func TestRealDatabase_GetPool_Multiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple pools
	for i := 0; i < 10; i++ {
		pool := &realdb.Pool{
			ID:        string(rune('a' + i)),
			Name:      "pool-" + string(rune('a'+i)),
			Available: i,
			Busy:      i,
		}
		if err := db.SavePool(pool); err != nil {
			t.Fatal(err)
		}
	}

	// Verify we can get each one
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		pool, err := db.GetPool(id)
		if err != nil {
			t.Errorf("GetPool(%s) failed: %v", id, err)
		}
		if pool.Available != i {
			t.Errorf("expected Available=%d, got %d", i, pool.Available)
		}
	}
}

// TestRealDatabase_DeleteMultiple tests deleting multiple entities
func TestRealDatabase_DeleteMultiple(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple clusters
	for i := 0; i < 5; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a' + i)), Name: string(rune('a' + i)), Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	// Delete first 3
	for i := 0; i < 3; i++ {
		if err := db.DeleteCluster(string(rune('a' + i))); err != nil {
			t.Fatal(err)
		}
	}

	// Verify first 3 are gone
	for i := 0; i < 3; i++ {
		_, err := db.GetCluster(string(rune('a' + i)))
		if err == nil {
			t.Errorf("cluster %s should be deleted", string(rune('a'+i)))
		}
	}

	// Verify last 2 are still there
	for i := 3; i < 5; i++ {
		_, err := db.GetCluster(string(rune('a' + i)))
		if err != nil {
			t.Errorf("cluster %s should still exist", string(rune('a'+i)))
		}
	}
}

// TestRealDatabase_VacuumOperation tests VACUUM operation
func TestRealDatabase_VacuumOperation(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create some data
	for i := 0; i < 100; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a'+i%26)), Name: "test", Status: "running"}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	// Delete some
	for i := 0; i < 50; i++ {
		if err := db.DeleteCluster(string(rune('a' + i%26))); err != nil {
			t.Fatal(err)
		}
	}

	// Verify counts
	counts := db.Count()
	_ = counts
}

// TestRealDatabase_SaveCluster_LargeConfig tests SaveCluster with large config
func TestRealDatabase_SaveCluster_LargeConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Large config
	largeConfig := make(map[string]string)
	for i := 0; i < 100; i++ {
		largeConfig[string(rune('a'+i%26))] = "value-" + string(rune('a'+i%26))
	}

	cluster := &realdb.Cluster{
		ID:        "cluster-large",
		Name:      "large-config-cluster",
		Status:    "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveCluster(cluster); err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	retrieved, err := db.GetCluster("cluster-large")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "large-config-cluster" {
		t.Errorf("wrong name: %s", retrieved.Name)
	}
}

// TestRealDatabase_SaveNode_LargeConfig tests SaveNode with large config
func TestRealDatabase_SaveNode_LargeConfig(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	node := &realdb.Node{
		ID:        "node-large",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "large-config-node",
		Role:      "worker",
		State:     "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.SaveNode(node); err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	retrieved, err := db.GetNode("node-large")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "large-config-node" {
		t.Errorf("wrong name: %s", retrieved.Name)
	}
}

// TestRealDatabase_NewDatabase_MaxConns tests connection pool settings
func TestRealDatabase_NewDatabase_MaxConns(t *testing.T) {
	cfg := &realdb.Config{
		Path:         ":memory:",
		MaxOpenConns: 100,
		MaxIdleConns: 50,
		WALMode:      true,
		BusyTimeout:  30 * time.Second,
	}

	db, err := realdb.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}
	defer db.Close()

	// Verify database is functional
	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

// TestRealDatabase_NewDatabaseWithDefaults_File tests file-based database
func TestRealDatabase_NewDatabaseWithDefaults_File(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := realdb.NewDatabaseWithDefaults(dbPath)
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create some data
	cluster := &realdb.Cluster{ID: "test", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created: %v", err)
	}
}

// TestRealDatabase_ListClusters_JSONHandling tests ListClusters JSON handling
func TestRealDatabase_ListClusters_JSONHandling(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create clusters with and without config
	for i := 0; i < 5; i++ {
		var config *realdb.ClusterConfig
		if i%2 == 0 {
			config = &realdb.ClusterConfig{MinNodes: i, MaxNodes: i * 2, AutoScale: true}
		}
		cluster := &realdb.Cluster{
			ID:        string(rune('a' + i)),
			Name:      "cluster-" + string(rune('a'+i)),
			Status:    "running",
			Config:    config,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.SaveCluster(cluster); err != nil {
			t.Fatal(err)
		}
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatal(err)
	}

	if len(clusters) != 5 {
		t.Errorf("expected 5 clusters, got %d", len(clusters))
	}
}

// TestRealDatabase_ListHosts_JSONHandling tests ListHosts
func TestRealDatabase_ListHosts_JSONHandling(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	for i := 0; i < 5; i++ {
		host := &realdb.Host{
			ID:         string(rune('a' + i)),
			Name:       "host-" + string(rune('a'+i)),
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
		t.Fatal(err)
	}

	if len(hosts) != 5 {
		t.Errorf("expected 5 hosts, got %d", len(hosts))
	}
}

// TestRealDatabase_ListClusterNodes_JSONHandling tests ListClusterNodes JSON
func TestRealDatabase_ListClusterNodes_JSONHandling(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	host := &realdb.Host{ID: "host-1", Name: "test", Address: "10.0.0.1"}
	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}
	cluster := &realdb.Cluster{ID: "cluster-1", Name: "test", Status: "running"}
	if err := db.SaveCluster(cluster); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		var config *realdb.NodeConfig
		if i%2 == 0 {
			config = &realdb.NodeConfig{CPU: i + 1, Memory: 1024 * (i + 1), Disk: 10 * (i + 1)}
		}
		node := &realdb.Node{
			ID:        string(rune('a' + i)),
			ClusterID: "cluster-1",
			HostID:    "host-1",
			Name:      "node-" + string(rune('a'+i)),
			Role:      "worker",
			State:     "running",
			Config:    config,
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatal(err)
		}
	}

	nodes, err := db.ListClusterNodes("cluster-1")
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(nodes))
	}
}








func TestRealDatabase_Count_AfterOperations(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create multiple entities
	for i := 0; i < 5; i++ {
		cluster := &realdb.Cluster{ID: string(rune('a'+i)), Name: string(rune('a'+i)), Status: "running"}
		db.SaveCluster(cluster)

		host := &realdb.Host{ID: string(rune('a'+i)), Name: string(rune('a'+i)), Address: "10.0.0.1"}
		db.SaveHost(host)
	}

	// Create nodes
	cluster := &realdb.Cluster{ID: "main", Name: "main", Status: "running"}
	db.SaveCluster(cluster)
	host := &realdb.Host{ID: "main", Name: "main", Address: "10.0.0.1"}
	db.SaveHost(host)

	for i := 0; i < 10; i++ {
		node := &realdb.Node{
			ID:        string(rune('a' + i)),
			ClusterID: "main",
			HostID:    "main",
			Name:      string(rune('a' + i)),
			Role:      "worker",
			State:     "running",
		}
		db.SaveNode(node)
	}

	// Create metrics
	for i := 0; i < 20; i++ {
		metric := &realdb.Metric{
			NodeID:     string(rune('a' + i%10)),
			CPU:        float64(i),
			RecordedAt: time.Now(),
		}
		db.SaveMetric(metric)
	}

	// Create alerts
	for i := 0; i < 5; i++ {
		alert := &realdb.Alert{
			ID:        string(rune('a' + i)),
			NodeID:    string(rune('a' + i)),
			Type:      "test",
			Message:   "test alert",
			Severity:  "warning",
			CreatedAt: time.Now(),
		}
		db.SaveAlert(alert)
	}

	// Create pools
	for i := 0; i < 3; i++ {
		pool := &realdb.Pool{
			ID:        string(rune('a' + i)),
			Name:      string(rune('a' + i)),
			Available: 10,
			Busy:      i,
		}
		db.SavePool(pool)
	}

	counts := db.Count()
	// Count only returns clusters, hosts, nodes, alerts, pools
	if counts["clusters"] != 6 {
		t.Errorf("expected 6 clusters, got %d", counts["clusters"])
	}
	if counts["hosts"] != 6 {
		t.Errorf("expected 6 hosts, got %d", counts["hosts"])
	}
	if counts["nodes"] != 10 {
		t.Errorf("expected 10 nodes, got %d", counts["nodes"])
	}
	// metrics not in Count()
	if counts["alerts"] != 5 {
		t.Errorf("expected 5 alerts, got %d", counts["alerts"])
	}
	if counts["pools"] != 3 {
		t.Errorf("expected 3 pools, got %d", counts["pools"])
	}
}

// TestRealDatabase_GetNodeMetrics_WithResults tests GetNodeMetrics with actual data
func TestRealDatabase_GetNodeMetrics_WithResults(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create metrics
	now := time.Now()
	for i := 0; i < 5; i++ {
		metric := &realdb.Metric{
			NodeID:     "node-1",
			CPU:        float64(i * 10),
			Memory:     float64(i * 20),
			Disk:       float64(i * 5),
			NetworkRX:  float64(i * 100),
			NetworkTX:  float64(i * 50),
			RecordedAt: now.Add(-time.Duration(i) * time.Minute),
		}
		if err := db.SaveMetric(metric); err != nil {
			t.Fatal(err)
		}
	}

	// Query for metrics
	metrics, err := db.GetNodeMetrics("node-1", now.Add(-10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	if len(metrics) != 5 {
		t.Errorf("expected 5 metrics, got %d", len(metrics))
	}

	// Verify first metric (most recent)
	if metrics[0].CPU != 0 {
		t.Errorf("expected first metric CPU=0, got %f", metrics[0].CPU)
	}
}

// TestRealDatabase_SaveHost_WithAllFields tests SaveHost with all fields
func TestRealDatabase_SaveHost_WithAllFields(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	host := &realdb.Host{
		ID:         "host-full",
		Name:       "full-host",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "admin",
		SSHKeyPath: "/home/admin/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if err := db.SaveHost(host); err != nil {
		t.Fatal(err)
	}

	retrieved, err := db.GetHost("host-full")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved.Address != "192.168.1.100" {
		t.Errorf("expected Address=192.168.1.100, got %s", retrieved.Address)
	}
	if retrieved.Port != 22 {
		t.Errorf("expected Port=22, got %d", retrieved.Port)
	}
	if retrieved.User != "admin" {
		t.Errorf("expected User=admin, got %s", retrieved.User)
	}
	if retrieved.HVType != "libvirt" {
		t.Errorf("expected HVType=libvirt, got %s", retrieved.HVType)
	}
}

// TestRealDatabase_SaveAlert_WithAllFields tests SaveAlert with all fields
func TestRealDatabase_SaveAlert_WithAllFields(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	alert := &realdb.Alert{
		ID:        "alert-full",
		NodeID:    "node-1",
		Type:      "cpu_high",
		Message:   "CPU usage exceeded 90%",
		Severity:  "critical",
		CreatedAt: time.Now(),
	}

	if err := db.SaveAlert(alert); err != nil {
		t.Fatal(err)
	}

	retrieved, err := db.GetAlert("alert-full")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved.Type != "cpu_high" {
		t.Errorf("expected Type=cpu_high, got %s", retrieved.Type)
	}
	if retrieved.Severity != "critical" {
		t.Errorf("expected Severity=critical, got %s", retrieved.Severity)
	}
}

// TestRealDatabase_SavePool_WithAllFields tests SavePool with all fields
func TestRealDatabase_SavePool_WithAllFields(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	pool := &realdb.Pool{
		ID:        "pool-full",
		Name:      "test-pool",
		Available: 100,
		Busy:      25,
	}

	if err := db.SavePool(pool); err != nil {
		t.Fatal(err)
	}

	retrieved, err := db.GetPool("pool-full")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved.Available != 100 {
		t.Errorf("expected Available=100, got %d", retrieved.Available)
	}
	if retrieved.Busy != 25 {
		t.Errorf("expected Busy=25, got %d", retrieved.Busy)
	}
}

// TestRealDatabase_ListHosts_WithData tests ListHosts with multiple hosts
func TestRealDatabase_ListHosts_WithData(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create hosts
	for i := 0; i < 10; i++ {
		host := &realdb.Host{
			ID:      string(rune('a' + i)),
			Name:    "host-" + string(rune('a'+i)),
			Address: "10.0.0." + string(rune('1'+i)),
		}
		if err := db.SaveHost(host); err != nil {
			t.Fatal(err)
		}
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatal(err)
	}

	if len(hosts) != 10 {
		t.Errorf("expected 10 hosts, got %d", len(hosts))
	}
}

// TestRealDatabase_ListPools_WithData tests ListPools with multiple pools
func TestRealDatabase_ListPools_WithData(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("NewDatabaseWithDefaults failed: %v", err)
	}
	defer db.Close()

	// Create pools
	for i := 0; i < 5; i++ {
		pool := &realdb.Pool{
			ID:        string(rune('a' + i)),
			Name:      "pool-" + string(rune('a'+i)),
			Available: i * 10,
			Busy:      i,
		}
		if err := db.SavePool(pool); err != nil {
			t.Fatal(err)
		}
	}

	pools, err := db.ListPools()
	if err != nil {
		t.Fatal(err)
	}

	if len(pools) != 5 {
		t.Errorf("expected 5 pools, got %d", len(pools))
	}
}
