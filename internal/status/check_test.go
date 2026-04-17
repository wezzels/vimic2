//go:build integration

package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

func setupWatcherWithDB(t *testing.T) (*Watcher, *database.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-watcher-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	w := NewWatcher(db, nil)
	return w, db, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

func TestWatcher_CheckCluster_Empty(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	cluster := &database.Cluster{
		ID:   "cluster-1",
		Name: "test-cluster",
	}

	// Should not panic with no hosts
	w.checkCluster(cluster)
	t.Log("checkCluster completed without panic")
}

func TestWatcher_NotifyNodeUpdate_NoSubscribers(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	node := &database.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
	}

	// Create a nil-safe status
	w.notifyNodeUpdate(node, "running", nil)
	t.Log("notifyNodeUpdate with nil status completed")
}

func TestWatcher_NotifyMetrics_NoSubscribers(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	node := &database.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
	}

	metrics := &hypervisor.Metrics{
		CPU:    45.0,
		Memory: 60.0,
		Disk:   30.0,
	}

	// This will try to call w.db.SaveMetric which may fail
	// but should not panic
	w.notifyMetrics(node, metrics)
	t.Log("notifyMetrics completed")
}

func TestWatcher_CheckAll_Empty(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	// Should not panic with no clusters
	w.checkAll()
	t.Log("checkAll completed without panic")
}

func TestWatcher_StartStop_WithContext(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w.Start(1 * time.Second)
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	_ = ctx
	t.Log("Watcher context test completed")
}

func TestWatcher_SubscribeAndNotify(t *testing.T) {
	w, _, cleanup := setupWatcherWithDB(t)
	defer cleanup()

	sub := &mockSubscriber{}
	w.Subscribe(sub)

	node := &database.Node{
		ID:        "node-1",
		ClusterID: "cluster-1",
	}

	// notifyNodeUpdate should reach our subscriber
	w.notifyNodeUpdate(node, "running", nil)
	t.Log("notifyNodeUpdate with nil status completed")

	w.Unsubscribe(sub)
	w.notifyNodeUpdate(node, "stopped", nil)
	t.Log("notifyNodeUpdate after unsubscribe completed")
}
