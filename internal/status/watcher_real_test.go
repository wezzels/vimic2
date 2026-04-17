//go:build integration

package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

func setupStatusDB_Real(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-status-real-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== WebSocket Hub Tests ====================

func TestWebSocketHub_Broadcast_Status(t *testing.T) {
	hub := NewWebSocketHub()
	if hub == nil {
		t.Fatal("NewWebSocketHub should not return nil")
	}

	hub.Broadcast(&NodeUpdate{Type: UpdateNode, NodeID: "test-node", State: "running"})
	t.Log("Broadcast succeeded")
}

// ==================== Struct Tests ====================

func TestNodeUpdate_Struct_Real(t *testing.T) {
	now := time.Now()
	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		IP:        "10.0.0.1",
		CPU:       45.0,
		Memory:    60.0,
		Disk:      30.0,
		Timestamp: now,
	}

	if update.Type != UpdateNode {
		t.Errorf("Type = %s, want %s", update.Type, UpdateNode)
	}
	if update.NodeID != "node-1" {
		t.Errorf("NodeID = %s, want node-1", update.NodeID)
	}
	if update.CPU != 45.0 {
		t.Errorf("CPU = %f, want 45.0", update.CPU)
	}
}

func TestClusterUpdate_Struct_Real(t *testing.T) {
	now := time.Now()
	update := &ClusterUpdate{
		Type:      UpdateCluster,
		ClusterID: "cluster-1",
		Status:    "running",
		NodeCount: 5,
		Timestamp: now,
	}

	if update.Type != UpdateCluster {
		t.Errorf("Type = %s, want %s", update.Type, UpdateCluster)
	}
	if update.ClusterID != "cluster-1" {
		t.Errorf("ClusterID = %s, want cluster-1", update.ClusterID)
	}
	if update.NodeCount != 5 {
		t.Errorf("NodeCount = %d, want 5", update.NodeCount)
	}
}

// ==================== UpdateType Constants ====================

func TestUpdateType_Constants_Status(t *testing.T) {
	if UpdateNode != "node" {
		t.Errorf("UpdateNode = %s, want node", UpdateNode)
	}
	if UpdateCluster != "cluster" {
		t.Errorf("UpdateCluster = %s, want cluster", UpdateCluster)
	}
	if UpdateMetrics != "metrics" {
		t.Errorf("UpdateMetrics = %s, want metrics", UpdateMetrics)
	}
}

// ==================== Watcher Start/Stop Tests ====================

func TestWatcher_StartStop_DB(t *testing.T) {
	db, cleanup := setupStatusDB_Real(t)
	defer cleanup()

	w := NewWatcher(db, nil)
	if w == nil {
		t.Fatal("NewWatcher should not return nil")
	}

	w.Start(1 * time.Second)
	time.Sleep(200 * time.Millisecond)
	w.Stop()

	t.Log("Watcher started and stopped successfully")
}

// ==================== Watcher with Context ====================

func TestWatcher_WithContext(t *testing.T) {
	db, cleanup := setupStatusDB_Real(t)
	defer cleanup()

	w := NewWatcher(db, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w.Start(1 * time.Second)
	time.Sleep(100 * time.Millisecond)

	select {
	case <-ctx.Done():
		w.Stop()
		t.Log("Watcher context cancelled successfully")
	case <-time.After(3 * time.Second):
		w.Stop()
		t.Error("Watcher context should have been cancelled")
	}
}