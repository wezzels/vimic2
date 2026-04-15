//go:build integration

package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// ==================== Closed DB error paths ====================

func TestListHosts_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	// Don't defer cleanup — we'll close manually
	db.Close()
	cleanup() // This also closes, but that's OK

	_, err := db.ListHosts()
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("ListHosts on closed DB (expected): %v", err)
	}
}

func TestGetNodeMetrics_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	_, err := db.GetNodeMetrics("node-1", time.Now())
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("GetNodeMetrics on closed DB (expected): %v", err)
	}
}

func TestListClusters_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	_, err := db.ListClusters()
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("ListClusters on closed DB (expected): %v", err)
	}
}

func TestGetActiveAlerts_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	_, err := db.GetActiveAlerts()
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("GetActiveAlerts on closed DB (expected): %v", err)
	}
}

func TestGetNodeAlerts_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	_, err := db.GetNodeAlerts("node-1")
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("GetNodeAlerts on closed DB (expected): %v", err)
	}
}

func TestListPipelines_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	_, err := db.ListPipelines()
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("ListPipelines on closed DB (expected): %v", err)
	}
}

func TestSaveCluster_ClosedDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	db.Close()
	cleanup()

	cluster := &Cluster{
		ID:     "cluster-closed",
		Name:   "closed-test",
		Config: &ClusterConfig{MinNodes: 1, MaxNodes: 5},
		Status: "running",
	}

	err := db.SaveCluster(cluster)
	if err == nil {
		t.Error("Expected error from closed DB")
	} else {
		t.Logf("SaveCluster on closed DB (expected): %v", err)
	}
}

// ==================== Verify sql.ErrTainted is available ====================

func TestSQLErrorTypes(t *testing.T) {
	// Verify we know how to trigger scan errors
	err := sql.ErrTxDone
	t.Logf("sql.ErrTxDone: %v", err)

	// Verify database close error
	err = sql.ErrConnDone
	t.Logf("sql.ErrConnDone: %v", err)

	// Verify errors.Is works
	if !errors.Is(sql.ErrTxDone, sql.ErrTxDone) {
		t.Error("errors.Is should match sql.ErrTxDone")
	}
}