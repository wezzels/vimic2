// Package realdb_test tests the real database
package realdb_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realdb"
)

// TestRealDatabase_Create tests database creation
func TestRealDatabase_Create(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("database should not be nil")
	}

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

// TestRealDatabase_Cluster tests cluster operations
func TestRealDatabase_Cluster(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	cluster := &realdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "running",
		Config:    &realdb.ClusterConfig{MinNodes: 1, MaxNodes: 10, AutoScale: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save
	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("failed to save cluster: %v", err)
	}

	// Get
	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}

	if retrieved.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", retrieved.Name)
	}
	if retrieved.Config.MinNodes != 1 {
		t.Errorf("expected MinNodes 1, got %d", retrieved.Config.MinNodes)
	}

	// List
	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}

	// Delete
	err = db.DeleteCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	// Verify deletion
	_, err = db.GetCluster("cluster-1")
	if err == nil {
		t.Error("cluster should be deleted")
	}
}

// TestRealDatabase_Host tests host operations
func TestRealDatabase_Host(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	host := &realdb.Host{
		ID:         "host-1",
		Name:       "worker-1",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		HVType:     "libvirt",
		CreatedAt:  time.Now(),
	}

	// Save
	err = db.SaveHost(host)
	if err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	// Get
	retrieved, err := db.GetHost("host-1")
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	if retrieved.Name != "worker-1" {
		t.Errorf("expected worker-1, got %s", retrieved.Name)
	}
	if retrieved.Address != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", retrieved.Address)
	}

	// List
	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts: %v", err)
	}
	if len(hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hosts))
	}

	// Delete
	err = db.DeleteHost("host-1")
	if err != nil {
		t.Fatalf("failed to delete host: %v", err)
	}
}

// TestRealDatabase_Node tests node operations
func TestRealDatabase_Node(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create cluster first
	cluster := &realdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = db.SaveCluster(cluster)

	// Create host
	host := &realdb.Host{
		ID:         "host-1",
		Name:       "worker-1",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		HVType:     "libvirt",
		CreatedAt:  time.Now(),
	}
	_ = db.SaveHost(host)

	node := &realdb.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		Name:      "node-1",
		Role:      "worker",
		State:     "running",
		IP:        "192.168.122.10",
		Config:    &realdb.NodeConfig{CPU: 2, Memory: 2048, Disk: 20, Image: "ubuntu-22.04"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save
	err = db.SaveNode(node)
	if err != nil {
		t.Fatalf("failed to save node: %v", err)
	}

	// Get
	retrieved, err := db.GetNode("node-1")
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	if retrieved.Name != "node-1" {
		t.Errorf("expected node-1, got %s", retrieved.Name)
	}
	if retrieved.Config.CPU != 2 {
		t.Errorf("expected CPU 2, got %d", retrieved.Config.CPU)
	}

	// List by cluster
	nodes, err := db.ListClusterNodes("cluster-1")
	if err != nil {
		t.Fatalf("failed to list cluster nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(nodes))
	}

	// Delete
	err = db.DeleteNode("node-1")
	if err != nil {
		t.Fatalf("failed to delete node: %v", err)
	}
}

// TestRealDatabase_Metric tests metric operations
func TestRealDatabase_Metric(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	metric := &realdb.Metric{
		NodeID:     "node-1",
		CPU:        25.5,
		Memory:     30.2,
		Disk:       40.0,
		NetworkRX:  1024000,
		NetworkTX:  512000,
		RecordedAt: time.Now(),
	}

	// Save
	err = db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("failed to save metric: %v", err)
	}

	if metric.ID == 0 {
		t.Error("metric ID should be set after save")
	}

	// Get metrics
	metrics, err := db.GetNodeMetrics("node-1", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("expected 1 metric, got %d", len(metrics))
	}

	if metrics[0].CPU != 25.5 {
		t.Errorf("expected CPU 25.5, got %f", metrics[0].CPU)
	}
}

// TestRealDatabase_Alert tests alert operations
func TestRealDatabase_Alert(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	alert := &realdb.Alert{
		ID:        "alert-1",
		NodeID:    "node-1",
		Type:      "cpu_high",
		Message:   "CPU usage above 80%",
		Severity:  "warning",
		CreatedAt: time.Now(),
	}

	// Save
	err = db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("failed to save alert: %v", err)
	}

	// Get
	retrieved, err := db.GetAlert("alert-1")
	if err != nil {
		t.Fatalf("failed to get alert: %v", err)
	}

	if retrieved.Message != "CPU usage above 80%" {
		t.Errorf("expected message, got %s", retrieved.Message)
	}

	// List
	alerts, err := db.ListAlerts()
	if err != nil {
		t.Fatalf("failed to list alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}

	// Delete
	err = db.DeleteAlert("alert-1")
	if err != nil {
		t.Fatalf("failed to delete alert: %v", err)
	}
}

// TestRealDatabase_Pool tests pool operations
func TestRealDatabase_Pool(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	pool := &realdb.Pool{
		ID:        "pool-1",
		Name:      "test-pool",
		Available: 5,
		Busy:      3,
	}

	// Save
	err = db.SavePool(pool)
	if err != nil {
		t.Fatalf("failed to save pool: %v", err)
	}

	// Get
	retrieved, err := db.GetPool("pool-1")
	if err != nil {
		t.Fatalf("failed to get pool: %v", err)
	}

	if retrieved.Name != "test-pool" {
		t.Errorf("expected test-pool, got %s", retrieved.Name)
	}
	if retrieved.Available != 5 {
		t.Errorf("expected Available 5, got %d", retrieved.Available)
	}

	// List
	pools, err := db.ListPools()
	if err != nil {
		t.Fatalf("failed to list pools: %v", err)
	}
	if len(pools) != 1 {
		t.Errorf("expected 1 pool, got %d", len(pools))
	}

	// Delete
	err = db.DeletePool("pool-1")
	if err != nil {
		t.Fatalf("failed to delete pool: %v", err)
	}
}

// TestRealDatabase_Count tests count function
func TestRealDatabase_Count(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Add some data
	_ = db.SaveCluster(&realdb.Cluster{ID: "c1", Name: "cluster-1", Status: "running", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = db.SaveCluster(&realdb.Cluster{ID: "c2", Name: "cluster-2", Status: "running", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = db.SaveHost(&realdb.Host{ID: "h1", Name: "host-1", Address: "192.168.1.100", CreatedAt: time.Now()})
	_ = db.SaveNode(&realdb.Node{ID: "n1", Name: "node-1", Role: "worker", State: "running", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	counts := db.Count()

	if counts["clusters"] != 2 {
		t.Errorf("expected 2 clusters, got %d", counts["clusters"])
	}
	if counts["hosts"] != 1 {
		t.Errorf("expected 1 host, got %d", counts["hosts"])
	}
	if counts["nodes"] != 1 {
		t.Errorf("expected 1 node, got %d", counts["nodes"])
	}
}

// TestRealDatabase_Transaction tests transaction support
func TestRealDatabase_Transaction(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Insert within transaction
	_, err = tx.Exec("INSERT INTO clusters (id, name, status) VALUES (?, ?, ?)", "tx-1", "tx-cluster", "created")
	if err != nil {
		tx.Rollback()
		t.Fatalf("failed to insert: %v", err)
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Get
	retrieved, err := db.GetCluster("tx-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}
	if retrieved.Name != "tx-cluster" {
		t.Errorf("expected tx-cluster, got %s", retrieved.Name)
	}
}

// TestRealDatabase_IntegrityCheck tests integrity check
func TestRealDatabase_IntegrityCheck(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	err = db.IntegrityCheck()
	if err != nil {
		t.Errorf("integrity check failed: %v", err)
	}
}

// TestRealDatabase_Backup tests backup
func TestRealDatabase_Backup(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Add some data
	_ = db.SaveCluster(&realdb.Cluster{ID: "c1", Name: "cluster-1", Status: "running", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	// Create temp backup file
	tmpDir, err := os.MkdirTemp("", "realdb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backupPath := filepath.Join(tmpDir, "backup.db")

	err = db.Backup(backupPath)
	if err != nil {
		t.Fatalf("failed to backup: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}
}

// TestRealDatabase_Vacuum tests vacuum
func TestRealDatabase_Vacuum(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Add and delete data
	_ = db.SaveCluster(&realdb.Cluster{ID: "c1", Name: "cluster-1", Status: "running", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = db.DeleteCluster("c1")

	// Vacuum
	err = db.Vacuum()
	if err != nil {
		t.Errorf("failed to vacuum: %v", err)
	}
}

// TestRealDatabase_File tests file-based database
func TestRealDatabase_File(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "realdb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := realdb.NewDatabaseWithDefaults(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Add data
	cluster := &realdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("failed to save cluster: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file should exist")
	}
}

// TestRealDatabase_Stats tests database stats
func TestRealDatabase_Stats(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stats := db.Stats()

	// Basic checks
	if stats.OpenConnections < 0 {
		t.Error("open connections should not be negative")
	}
}

// TestRealDatabase_Update tests update operations
func TestRealDatabase_Update(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create cluster
	cluster := &realdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "created",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = db.SaveCluster(cluster)

	// Update cluster
	cluster.Status = "running"
	cluster.Name = "updated-cluster"
	cluster.UpdatedAt = time.Now()
	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("failed to update cluster: %v", err)
	}

	// Verify update
	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}

	if retrieved.Name != "updated-cluster" {
		t.Errorf("expected updated-cluster, got %s", retrieved.Name)
	}
	if retrieved.Status != "running" {
		t.Errorf("expected running, got %s", retrieved.Status)
	}
}

// TestRealDatabase_Concurrent tests concurrent access
func TestRealDatabase_Concurrent(t *testing.T) {
	db, err := realdb.NewDatabaseWithDefaults(":memory:")
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Concurrent writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			cluster := &realdb.Cluster{
				ID:        fmt.Sprintf("cluster-%d", id),
				Name:      fmt.Sprintf("test-cluster-%d", id),
				Status:    "running",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = db.SaveCluster(cluster)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all clusters were created
	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}

	if len(clusters) != 10 {
		t.Errorf("expected 10 clusters, got %d", len(clusters))
	}
}