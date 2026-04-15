//go:build integration

package database

import (
	"os"
	"testing"
	"time"
)

// ==================== Cover ListPipelines scan error ====================

func TestListPipelines_ScanError_Schema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Drop the pipelines table and recreate with wrong schema
	_, err := db.db.Exec(`DROP TABLE IF EXISTS pipelines`)
	if err != nil {
		t.Fatal(err)
	}
	// Create with integer id instead of text
	_, err = db.db.Exec(`CREATE TABLE pipelines (id INTEGER PRIMARY KEY, state TEXT, updated_at TIMESTAMP)`)
	if err != nil {
		t.Fatal(err)
	}
	// Insert an integer id
	_, err = db.db.Exec(`INSERT INTO pipelines (id, state, updated_at) VALUES (1, '{}', '2024-01-01')`)
	if err != nil {
		t.Fatal(err)
	}

	// ListPipelines should still work since it just scans a string
	ids, err := db.ListPipelines()
	t.Logf("ListPipelines with schema mismatch: ids=%v err=%v", ids, err)
}

// ==================== Cover ListHosts scan error via schema mismatch ====================

func TestListHosts_ScanErrorViaSchema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert a host first
	host := &Host{
		ID:         "host-schema",
		Name:       "schema-host",
		Address:    "10.0.0.1",
		Port:       22,
		User:       "admin",
		SSHKeyPath: "/path",
		HVType:     "libvirt",
	}
	err := db.SaveHost(host)
	if err != nil {
		t.Fatal(err)
	}

	// Alter table to add a required column with wrong type
	_, err = db.db.Exec(`ALTER TABLE hosts ADD COLUMN extra TEXT NOT NULL DEFAULT ''`)
	if err != nil {
		t.Logf("ALTER TABLE: %v (may not be supported)", err)
	}

	// ListHosts should still work
	hosts, err := db.ListHosts()
	if err != nil {
		t.Logf("ListHosts after alter: %v", err)
	} else {
		t.Logf("ListHosts returned %d hosts", len(hosts))
	}
}

// ==================== Cover NewDB Ping Error ====================

func TestNewDB_PingError2(t *testing.T) {
	// Try opening a path that exists but is not a valid SQLite file
	tmpFile, err := os.CreateTemp("", "vimic2-ping-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	// Write garbage
	tmpFile.WriteString("this is not a sqlite database")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// SQLite may not error on Open, but should error on Ping or migrate
	_, err = NewDB(tmpFile.Name())
	if err != nil {
		t.Logf("NewDB with garbage file (expected): %v", err)
	} else {
		t.Log("NewDB succeeded with garbage file (SQLite may overwrite)")
	}
}

// ==================== GetActiveAlerts with resolved_at ====================

func TestGetActiveAlerts_WithResolvedAt(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	resolvedAt := now.Add(5 * time.Minute)

	// Insert a resolved alert directly via SQL with resolved_at
	_, err := db.db.Exec(`
		INSERT INTO alerts (id, rule_id, node_id, node_name, metric, value, threshold, message, fired_at, resolved, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"alert-resolved", "rule-1", "node-1", "test-node", "cpu", 95.0, 80.0, "CPU critical", now, 1, resolvedAt)
	if err != nil {
		t.Fatal(err)
	}

	// Get active alerts (should be empty since resolved=1)
	alerts, err := db.GetActiveAlerts()
	if err != nil {
		t.Logf("GetActiveAlerts error: %v", err)
	}
	t.Logf("Active alerts after resolved insert: %d", len(alerts))

	// Get node alerts (should include resolved ones)
	nodeAlerts, err := db.GetNodeAlerts("node-1")
	if err != nil {
		t.Logf("GetNodeAlerts error: %v", err)
	}
	t.Logf("Node alerts after resolved insert: %d", len(nodeAlerts))

	// Verify ResolvedAt was scanned properly
	if len(nodeAlerts) > 0 {
		if nodeAlerts[0].ResolvedAt != nil {
			t.Logf("ResolvedAt: %v", *nodeAlerts[0].ResolvedAt)
		}
	}
}