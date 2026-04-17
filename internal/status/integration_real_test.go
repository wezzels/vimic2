//go:build integration

package status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// realSubscriber collects status updates
type realSubscriber struct {
	mu       sync.Mutex
	updates  []*NodeUpdate
	clusters []*ClusterUpdate
}

func (m *realSubscriber) OnNodeUpdate(u *NodeUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updates = append(m.updates, u)
}

func (m *realSubscriber) OnClusterUpdate(u *ClusterUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clusters = append(m.clusters, u)
}

func (m *realSubscriber) getUpdates() []*NodeUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updates
}

func (m *realSubscriber) getClusters() []*ClusterUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.clusters
}

func setupStatusTest(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-status-test-")
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

// ==================== Watcher Creation Tests ====================

func TestNewWatcher_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)
	if w == nil {
		t.Fatal("NewWatcher should not return nil")
	}
}

func TestWatcher_SubscribeUnsubscribe_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	sub := &realSubscriber{}
	w.Subscribe(sub)

	if len(w.subs) != 1 {
		t.Errorf("subscribers = %d, want 1", len(w.subs))
	}

	w.Unsubscribe(sub)

	if len(w.subs) != 0 {
		t.Errorf("subscribers = %d, want 0", len(w.subs))
	}
}

func TestWatcher_MultipleSubscribers_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	sub1 := &realSubscriber{}
	sub2 := &realSubscriber{}
	w.Subscribe(sub1)
	w.Subscribe(sub2)

	if len(w.subs) != 2 {
		t.Errorf("subscribers = %d, want 2", len(w.subs))
	}
}

// ==================== NodeUpdate Tests ====================

func TestNodeUpdate_Serialization_Real(t *testing.T) {
	now := time.Now()
	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		IP:        "10.0.0.1",
		CPU:       45.2,
		Memory:    60.0,
		Disk:      30.5,
		Timestamp: now,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got NodeUpdate
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.NodeID != "node-1" {
		t.Errorf("NodeID = %s, want node-1", got.NodeID)
	}
	if got.State != "running" {
		t.Errorf("State = %s, want running", got.State)
	}
	if got.CPU != 45.2 {
		t.Errorf("CPU = %f, want 45.2", got.CPU)
	}
	if got.Type != UpdateNode {
		t.Errorf("Type = %s, want %s", got.Type, UpdateNode)
	}
}

func TestClusterUpdate_Serialization_Real(t *testing.T) {
	now := time.Now()
	update := &ClusterUpdate{
		Type:       UpdateCluster,
		ClusterID:  "cluster-1",
		Status:     "healthy",
		NodeCount:  3,
		Timestamp:  now,
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got ClusterUpdate
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ClusterID != "cluster-1" {
		t.Errorf("ClusterID = %s, want cluster-1", got.ClusterID)
	}
	if got.Status != "healthy" {
		t.Errorf("Status = %s, want healthy", got.Status)
	}
	if got.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want 3", got.NodeCount)
	}
}

// ==================== UpdateType Tests ====================

func TestUpdateType_Constants_Real(t *testing.T) {
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

func TestWatcher_StartStop_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	w.Start(10 * time.Second)
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	t.Log("Watcher started and stopped successfully")
}

func TestWatcher_StartStop_Quick_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	w.Start(1 * time.Hour)
	w.Stop()

	t.Log("Quick start/stop completed")
}

// ==================== Database Integration Tests ====================

func TestWatcher_WithDatabase_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	// Create a cluster
	cluster := &database.Cluster{
		Name:   "test-cluster",
		Status: "healthy",
	}
	err := db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("SaveCluster failed: %v", err)
	}

	// Create a node
	node := &database.Node{
		ClusterID: cluster.ID,
		Name:      "test-node",
		State:     "running",
		IP:        "10.0.0.1",
	}
	err = db.SaveNode(node)
	if err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	// Create watcher
	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	sub := &realSubscriber{}
	w.Subscribe(sub)

	w.Start(1 * time.Hour)
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	t.Logf("Created cluster %s with node %s", cluster.ID, node.ID)
}

func TestWatcher_EmptyDatabase_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	sub := &realSubscriber{}
	w.Subscribe(sub)

	w.Start(1 * time.Hour)
	time.Sleep(100 * time.Millisecond)
	w.Stop()

	updates := sub.getUpdates()
	if len(updates) != 0 {
		t.Errorf("Expected no updates, got %d", len(updates))
	}
}

// ==================== NodeUpdate Content Tests ====================

func TestNodeUpdate_Fields_Real(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		update NodeUpdate
	}{
		{
			name: "running node",
			update: NodeUpdate{
				Type:      UpdateNode,
				NodeID:    "node-1",
				ClusterID: "cluster-1",
				State:     "running",
				IP:        "10.0.0.1",
				CPU:       75.0,
				Memory:    80.0,
				Disk:      45.0,
				Timestamp: now,
			},
		},
		{
			name: "stopped node",
			update: NodeUpdate{
				Type:      UpdateNode,
				NodeID:    "node-2",
				ClusterID: "cluster-1",
				State:     "stopped",
				IP:        "10.0.0.2",
				CPU:       0,
				Memory:    0,
				Disk:      0,
				Timestamp: now,
			},
		},
		{
			name: "error node",
			update: NodeUpdate{
				Type:      UpdateNode,
				NodeID:    "node-3",
				ClusterID: "cluster-2",
				State:     "error",
				IP:        "",
				CPU:       0,
				Memory:    0,
				Disk:      0,
				Timestamp: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.update)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			var got NodeUpdate
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if got.NodeID != tt.update.NodeID {
				t.Errorf("NodeID = %s, want %s", got.NodeID, tt.update.NodeID)
			}
			if got.State != tt.update.State {
				t.Errorf("State = %s, want %s", got.State, tt.update.State)
			}
		})
	}
}

// ==================== ClusterUpdate Content Tests ====================

func TestClusterUpdate_Fields_Real(t *testing.T) {
	now := time.Now()
	update := ClusterUpdate{
		Type:       UpdateCluster,
		ClusterID:  "cluster-1",
		Status:     "degraded",
		NodeCount:  5,
		Timestamp:  now,
	}

	if update.Type != UpdateCluster {
		t.Errorf("Type = %s, want %s", update.Type, UpdateCluster)
	}
	if update.Status != "degraded" {
		t.Errorf("Status = %s, want degraded", update.Status)
	}
	if update.NodeCount != 5 {
		t.Errorf("NodeCount = %d, want 5", update.NodeCount)
	}
}

// ==================== Metrics Update Tests ====================

func TestMetricsUpdate_Type_Real(t *testing.T) {
	if UpdateMetrics != "metrics" {
		t.Errorf("UpdateMetrics = %s, want metrics", UpdateMetrics)
	}
}

func TestNodeUpdate_MetricsType_Real(t *testing.T) {
	update := NodeUpdate{
		Type: UpdateMetrics,
	}
	if update.Type != UpdateMetrics {
		t.Errorf("Type = %s, want metrics", update.Type)
	}
}

// ==================== Context Cancellation ====================

func TestWatcher_ContextCancellation_Real(t *testing.T) {
	db, cleanup := setupStatusTest(t)
	defer cleanup()

	hosts := make(map[string]hypervisor.Hypervisor)
	w := NewWatcher(db, hosts)

	w.Start(100 * time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	w.Stop()

	t.Log("Watcher stopped after context cancellation")
}