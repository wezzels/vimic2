// Package mockdb_test tests the mock database
package mockdb_test

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/testutil/mockdb"
)

// TestMockDB_Create tests database creation
func TestMockDB_Create(t *testing.T) {
	db := mockdb.NewMockDB()

	if db == nil {
		t.Fatal("database should not be nil")
	}
}

// TestMockDB_Cluster tests cluster operations
func TestMockDB_Cluster(t *testing.T) {
	db := mockdb.NewMockDB()

	cluster := &mockdb.Cluster{
		ID:        "cluster-1",
		Name:      "test-cluster",
		Status:    "running",
		CreatedAt: time.Now(),
	}

	err := db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("failed to save cluster: %v", err)
	}

	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}

	if retrieved.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", retrieved.Name)
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatalf("failed to list clusters: %v", err)
	}

	if len(clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(clusters))
	}

	err = db.DeleteCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to delete cluster: %v", err)
	}

	_, err = db.GetCluster("cluster-1")
	if err == nil {
		t.Error("cluster should be deleted")
	}
}

// TestMockDB_Host tests host operations
func TestMockDB_Host(t *testing.T) {
	db := mockdb.NewMockDB()

	host := &mockdb.Host{
		ID:        "host-1",
		Name:      "worker-1",
		Address:   "192.168.1.100",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	err := db.SaveHost(host)
	if err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	retrieved, err := db.GetHost("host-1")
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	if retrieved.Name != "worker-1" {
		t.Errorf("expected worker-1, got %s", retrieved.Name)
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts: %v", err)
	}

	if len(hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(hosts))
	}
}

// TestMockDB_Node tests node operations
func TestMockDB_Node(t *testing.T) {
	db := mockdb.NewMockDB()

	node := &mockdb.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
		Name:      "node-1",
		Role:      "worker",
		State:     "running",
		IP:        "192.168.122.10",
		CreatedAt: time.Now(),
	}

	err := db.SaveNode(node)
	if err != nil {
		t.Fatalf("failed to save node: %v", err)
	}

	retrieved, err := db.GetNode("node-1")
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	if retrieved.Name != "node-1" {
		t.Errorf("expected node-1, got %s", retrieved.Name)
	}

	nodes, err := db.ListClusterNodes("cluster-1")
	if err != nil {
		t.Fatalf("failed to list cluster nodes: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(nodes))
	}
}

// TestMockDB_Metric tests metric operations
func TestMockDB_Metric(t *testing.T) {
	db := mockdb.NewMockDB()

	metric := &mockdb.Metric{
		NodeID:     "node-1",
		CPU:        25.5,
		Memory:     30.2,
		Disk:       40.0,
		NetworkRX:  1024000,
		NetworkTX:  512000,
		RecordedAt: time.Now(),
	}

	err := db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("failed to save metric: %v", err)
	}

	retrieved, err := db.GetNodeMetrics("node-1", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("expected 1 metric, got %d", len(retrieved))
	}
}

// TestMockDB_Alert tests alert operations
func TestMockDB_Alert(t *testing.T) {
	db := mockdb.NewMockDB()

	alert := &mockdb.Alert{
		ID:        "alert-1",
		NodeID:     "node-1",
		Type:      "cpu_high",
		Message:   "CPU usage above 80%",
		Severity:  "warning",
		CreatedAt: time.Now(),
	}

	err := db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("failed to save alert: %v", err)
	}

	retrieved, err := db.GetAlert("alert-1")
	if err != nil {
		t.Fatalf("failed to get alert: %v", err)
	}

	if retrieved.Message != "CPU usage above 80%" {
		t.Errorf("expected alert message, got %s", retrieved.Message)
	}

	alerts, err := db.ListAlerts()
	if err != nil {
		t.Fatalf("failed to list alerts: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
}

// TestMockDB_Pool tests pool operations
func TestMockDB_Pool(t *testing.T) {
	db := mockdb.NewMockDB()

	pool := &mockdb.Pool{
		ID:        "pool-1",
		Name:      "test-pool",
		Available: 5,
		Busy:      3,
	}

	err := db.SavePool(pool)
	if err != nil {
		t.Fatalf("failed to save pool: %v", err)
	}

	retrieved, err := db.GetPool("pool-1")
	if err != nil {
		t.Fatalf("failed to get pool: %v", err)
	}

	if retrieved.Name != "test-pool" {
		t.Errorf("expected test-pool, got %s", retrieved.Name)
	}
}

// TestMockDB_Count tests count function
func TestMockDB_Count(t *testing.T) {
	db := mockdb.NewMockDB()

	// Add some data
	_ = db.SaveCluster(&mockdb.Cluster{ID: "c1", Name: "cluster-1"})
	_ = db.SaveCluster(&mockdb.Cluster{ID: "c2", Name: "cluster-2"})
	_ = db.SaveHost(&mockdb.Host{ID: "h1", Name: "host-1"})
	_ = db.SaveNode(&mockdb.Node{ID: "n1", Name: "node-1"})

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

// TestMockDB_ErrorMode tests error mode
func TestMockDB_ErrorMode(t *testing.T) {
	db := mockdb.NewMockDB()

	// Enable error mode
	db.SetErrorMode(true)

	// Operations should fail
	err := db.SaveCluster(&mockdb.Cluster{ID: "c1"})
	if err == nil {
		t.Error("expected error in error mode")
	}

	// Disable error mode
	db.SetErrorMode(false)

	// Operations should succeed
	err = db.SaveCluster(&mockdb.Cluster{ID: "c1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockDB_FailNext tests fail next
func TestMockDB_FailNext(t *testing.T) {
	db := mockdb.NewMockDB()

	// Set up fail next
	db.FailNext()

	// This operation should fail
	err := db.SaveCluster(&mockdb.Cluster{ID: "c1"})
	if err == nil {
		t.Error("expected error from fail next")
	}

	// This operation should succeed (failNext was consumed)
	err = db.SaveCluster(&mockdb.Cluster{ID: "c1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}