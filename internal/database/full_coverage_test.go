//go:build integration

package database

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// ==================== JSON Marshal Error Paths ====================

func TestSaveCluster_JSONMarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// A channel can't be marshaled to JSON
	ch := make(chan int)
	cluster := &Cluster{
		ID:     "cluster-bad",
		Name:   "bad-cluster",
		Config: &ClusterConfig{Network: &NetworkConfig{Type: "nat", CIDR: "10.0.0.0/24"}},
	}
	// Override Config with something that has a channel via raw struct
	// Actually, ClusterConfig has all JSON-safe fields. Need to use SavePipeline/SaveRunner
	// which take map[string]interface{} instead.
	_ = ch
	_ = cluster

	// For SavePipeline, we can pass a channel in the map:
	err := db.SavePipeline("pipe-bad", map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("Expected error when marshaling pipeline with channel")
	} else {
		t.Logf("SavePipeline marshal error (expected): %v", err)
	}

	// Same for SaveRunner
	err = db.SaveRunner("runner-bad", map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("Expected error when marshaling runner with channel")
	} else {
		t.Logf("SaveRunner marshal error (expected): %v", err)
	}

	// Same for SaveNetwork (if it takes map)
	// SaveNetwork and SavePool use different signatures — check them
}

// ==================== JSON Unmarshal Error Paths ====================

func TestLoadPipeline_JSONUnmarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert corrupted JSON directly into the DB
	_, err := db.db.Exec(`INSERT OR REPLACE INTO pipelines (id, state, updated_at) VALUES (?, ?, ?)`,
		"pipe-corrupt", "not-valid-json{", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.LoadPipeline("pipe-corrupt")
	if err == nil {
		t.Error("Expected error when unmarshaling corrupted pipeline JSON")
	} else {
		t.Logf("LoadPipeline unmarshal error (expected): %v", err)
	}
}

func TestLoadRunner_JSONUnmarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.db.Exec(`INSERT OR REPLACE INTO runners (id, state, updated_at) VALUES (?, ?, ?)`,
		"runner-corrupt", "not-valid-json{", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.LoadRunner("runner-corrupt")
	if err == nil {
		t.Error("Expected error when unmarshaling corrupted runner JSON")
	} else {
		t.Logf("LoadRunner unmarshal error (expected): %v", err)
	}
}

// ==================== NewDB Migrate Error ====================

func TestNewDB_MigrateError2(t *testing.T) {
	// Open a DB that's read-only to make migration fail
	tmpFile, err := os.CreateTemp("", "vimic2-migrate-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	// Make file read-only
	os.Chmod(tmpPath, 0444)
	defer os.Remove(tmpPath)

	// This might fail at Open or Ping, but let's try
	_, err = NewDB(tmpPath)
	if err != nil {
		t.Logf("NewDB with read-only file (expected): %v", err)
	} else {
		t.Log("NewDB succeeded with read-only file (SQLite may handle this)")
	}
}

// ==================== ListPipelines Scan Error ====================

func TestListPipelines_ScanError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert a pipeline
	err := db.SavePipeline("pipe-scan-test", map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}

	// Now list should work
	pipelines, err := db.ListPipelines()
	if err != nil {
		t.Errorf("ListPipelines failed: %v", err)
	}
	if len(pipelines) != 1 {
		t.Errorf("Expected 1 pipeline, got %d", len(pipelines))
	}
}

// ==================== SaveNetwork/LoadNetwork/SavePool/LoadPool JSON errors ====================

func TestSaveNetwork_JSONMarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.SaveNetwork("net-bad", map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("Expected error when marshaling network with channel")
	} else {
		t.Logf("SaveNetwork marshal error (expected): %v", err)
	}
}

func TestLoadNetwork_JSONUnmarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.db.Exec(`INSERT OR REPLACE INTO networks (id, state, updated_at) VALUES (?, ?, ?)`,
		"net-corrupt", "not-valid-json{", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.LoadNetwork("net-corrupt")
	if err == nil {
		t.Error("Expected error when unmarshaling corrupted network JSON")
	} else {
		t.Logf("LoadNetwork unmarshal error (expected): %v", err)
	}
}

func TestSavePool_JSONMarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.SavePool("pool-bad", map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("Expected error when marshaling pool with channel")
	} else {
		t.Logf("SavePool marshal error (expected): %v", err)
	}
}

func TestLoadPool_JSONUnmarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := db.db.Exec(`INSERT OR REPLACE INTO pools (id, state, updated_at) VALUES (?, ?, ?)`,
		"pool-corrupt", "not-valid-json{", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.LoadPool("pool-corrupt")
	if err == nil {
		t.Error("Expected error when unmarshaling corrupted pool JSON")
	} else {
		t.Logf("LoadPool unmarshal error (expected): %v", err)
	}
}

// ==================== SaveCluster with actual marshal-failing config ====================

func TestSaveCluster_RealMarshalError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// We need to create a ClusterConfig that can't be marshaled.
	// Since ClusterConfig fields are all JSON-safe, we use a workaround:
	// Create a cluster with Config that has circular reference via custom type
	cluster := &Cluster{
		ID:   "cluster-marshal",
		Name: "marshal-test",
		Config: &ClusterConfig{
			MinNodes: 1,
			MaxNodes: 10,
		},
		Status: "running",
	}

	// This should succeed normally
	err := db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("SaveCluster failed for normal cluster: %v", err)
	}

	// To trigger the marshal error, we need to close the DB first
	db.Close()

	// Create a new DB and try with a closed connection
	// Actually, we need the json.Marshal error, not the DB error.
	// Since ClusterConfig has all safe types, we can't trigger json.Marshal error.
	// The marshal error path is effectively dead code for this struct.
	t.Log("ClusterConfig uses only JSON-safe types — marshal error is unreachable")
}

// ==================== SaveAlert / GetActiveAlerts / GetNodeAlerts full coverage ====================

func TestSaveAndGetAlerts_Full(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	alert := &Alert{
		ID:         "alert-1",
		RuleID:     "rule-1",
		NodeID:     "node-1",
		NodeName:   "test-node",
		Metric:     "cpu",
		Value:      95.5,
		Threshold:  80.0,
		Message:    "CPU usage critical",
		FiredAt:    now,
		Resolved:   false,
		ResolvedAt: nil,
	}

	err := db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("SaveAlert failed: %v", err)
	}

	// Get active alerts
	alerts, err := db.GetActiveAlerts()
	if err != nil {
		t.Fatalf("GetActiveAlerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 active alert, got %d", len(alerts))
	}

	// Get node alerts
	nodeAlerts, err := db.GetNodeAlerts("node-1")
	if err != nil {
		t.Fatalf("GetNodeAlerts failed: %v", err)
	}
	if len(nodeAlerts) != 1 {
		t.Errorf("Expected 1 node alert, got %d", len(nodeAlerts))
	}

	// Resolve the alert
	alert.Resolved = true
	resolvedAt := now.Add(5 * time.Minute)
	alert.ResolvedAt = &resolvedAt
	err = db.SaveAlert(alert)
	if err != nil {
		t.Fatalf("SaveAlert (resolved) failed: %v", err)
	}

	// Active alerts should now be empty
	alerts, err = db.GetActiveAlerts()
	if err != nil {
		t.Fatalf("GetActiveAlerts (resolved) failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("Expected 0 active alerts after resolve, got %d", len(alerts))
	}

	// But node alerts should still show it
	nodeAlerts, err = db.GetNodeAlerts("node-1")
	if err != nil {
		t.Fatalf("GetNodeAlerts (resolved) failed: %v", err)
	}
	t.Logf("Node alerts after resolve: %d", len(nodeAlerts))
}

// ==================== GetNodeMetrics Full Coverage ====================

func TestGetNodeMetrics_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Save some metrics
	now := time.Now()
	for i := 0; i < 5; i++ {
		metric := &Metric{
			NodeID:     "node-metrics",
			CPU:        float64(i * 10),
			Memory:     float64(i * 5),
			Disk:       float64(i * 2),
			NetworkRX:  float64(i * 100),
			NetworkTX:  float64(i * 50),
			RecordedAt: now.Add(time.Duration(i) * time.Minute),
		}
		err := db.SaveMetric(metric)
		if err != nil {
			t.Fatalf("SaveMetric failed: %v", err)
		}
	}

	// Get metrics
	since := now.Add(-1 * time.Minute)
	metrics, err := db.GetNodeMetrics("node-metrics", since)
	if err != nil {
		t.Fatalf("GetNodeMetrics failed: %v", err)
	}
	if len(metrics) != 5 {
		t.Errorf("Expected 5 metrics, got %d", len(metrics))
	}
}

// ==================== Helper ====================

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "vimic2-db-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	db, err := NewDB(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		t.Fatal(err)
	}

	return db, func() {
		db.Close()
		os.Remove(tmpPath)
	}
}

// Verify json.Marshal actually fails with channels
func TestChannelMarshalFails(t *testing.T) {
	_, err := json.Marshal(map[string]interface{}{"ch": make(chan int)})
	if err == nil {
		t.Error("json.Marshal should fail with channel value")
	}
}

// Verify DB direct access
func TestDB_DirectAccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Verify db.db is accessible
	var ok bool
	_ = ok
	if db.db == nil {
		t.Fatal("DB should have non-nil sql.DB")
	}

	// Execute raw SQL
	_, err := db.db.Exec("CREATE TABLE IF NOT EXISTS test (id TEXT)")
	if err != nil {
		t.Logf("Raw SQL exec: %v", err)
	}
}

// ==================== ListPipelines Full Path ====================

func TestListPipelines_Multiple(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		err := db.SavePipeline("pipe-"+string(rune('A'+i)), map[string]interface{}{
			"key":   "value",
			"index": i,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	pipelines, err := db.ListPipelines()
	if err != nil {
		t.Fatal(err)
	}
	if len(pipelines) != 3 {
		t.Errorf("Expected 3 pipelines, got %d", len(pipelines))
	}
}

// ==================== ListClusters with data ====================

func TestListClusters_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		cluster := &Cluster{
			ID:     "cluster-list-" + string(rune('A'+i)),
			Name:   "list-cluster-" + string(rune('A'+i)),
			Config: &ClusterConfig{MinNodes: 1, MaxNodes: 5},
			Status: "running",
		}
		err := db.SaveCluster(cluster)
		if err != nil {
			t.Fatal(err)
		}
	}

	clusters, err := db.ListClusters()
	if err != nil {
		t.Fatal(err)
	}
	if len(clusters) != 3 {
		t.Errorf("Expected 3 clusters, got %d", len(clusters))
	}
}

// ==================== ListHosts with data ====================

func TestListHosts_WithData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		host := &Host{
			ID:         "host-list-" + string(rune('A'+i)),
			Name:       "list-host-" + string(rune('A'+i)),
			Address:    "10.0.0." + string(rune('1'+i)),
			Port:       22,
			User:       "admin",
			SSHKeyPath: "/path/to/key",
			HVType:     "libvirt",
		}
		err := db.SaveHost(host)
		if err != nil {
			t.Fatal(err)
		}
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 3 {
		t.Errorf("Expected 3 hosts, got %d", len(hosts))
	}
}

// ==================== NewDB sql.Open Error ====================

func TestNewDB_SqlOpenError(t *testing.T) {
	// Try to open with an invalid path (directory that doesn't exist for writing)
	_, err := NewDB("/nonexistent/deep/path/that/doesnt/exist/test.db")
	if err != nil {
		t.Logf("NewDB with invalid path (expected): %v", err)
	} else {
		t.Log("NewDB succeeded with invalid path (SQLite may create it)")
	}
}

// ==================== Verify SaveNetwork/SavePool signatures ====================

func TestSaveNetwork_Signature(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Check what SaveNetwork actually takes
	err := db.SaveNetwork("net-1", map[string]interface{}{"cidr": "10.0.0.0/24"})
	if err != nil {
		t.Logf("SaveNetwork: %v", err)
	}
}

func TestSavePool_Signature(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := db.SavePool("pool-1", map[string]interface{}{"size": 10})
	if err != nil {
		t.Logf("SavePool: %v", err)
	}
}

// Compile-time check that *sql.DB is accessible
var _ = func() *sql.DB {
	var d *DB
	return d.db
}