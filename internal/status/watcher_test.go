// Package status provides tests for status watcher
package status

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestNodeUpdate tests node update structure
func TestNodeUpdate_Create(t *testing.T) {
	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		IP:        "10.0.0.1",
		CPU:       45.5,
		Memory:    60.2,
		Disk:      30.1,
		Timestamp: time.Now(),
	}

	if update.Type != UpdateNode {
		t.Errorf("expected UpdateNode, got %s", update.Type)
	}
	if update.NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", update.NodeID)
	}
	if update.State != "running" {
		t.Errorf("expected running state, got %s", update.State)
	}
	if update.CPU != 45.5 {
		t.Errorf("expected CPU 45.5, got %f", update.CPU)
	}
}

// TestClusterUpdate tests cluster update structure
func TestClusterUpdate_Create(t *testing.T) {
	update := &ClusterUpdate{
		Type:      UpdateCluster,
		ClusterID: "cluster-1",
		Status:    "healthy",
		NodeCount: 5,
		Timestamp: time.Now(),
	}

	if update.Type != UpdateCluster {
		t.Errorf("expected UpdateCluster, got %s", update.Type)
	}
	if update.ClusterID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", update.ClusterID)
	}
	if update.Status != "healthy" {
		t.Errorf("expected healthy status, got %s", update.Status)
	}
	if update.NodeCount != 5 {
		t.Errorf("expected 5 nodes, got %d", update.NodeCount)
	}
}

// TestUpdateType tests update type constants
func TestUpdateType_Constants(t *testing.T) {
	types := []UpdateType{UpdateNode, UpdateCluster, UpdateMetrics}

	for _, ut := range types {
		if ut == "" {
			t.Error("empty update type")
		}
	}
}

// TestNodeUpdate_JSON tests JSON marshaling
func TestNodeUpdate_JSON(t *testing.T) {
	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		IP:        "10.0.0.1",
		CPU:       45.5,
		Memory:    60.2,
		Disk:      30.1,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var update2 NodeUpdate
	if err := json.Unmarshal(data, &update2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if update2.NodeID != update.NodeID {
		t.Errorf("expected NodeID %s, got %s", update.NodeID, update2.NodeID)
	}
	if update2.CPU != update.CPU {
		t.Errorf("expected CPU %f, got %f", update.CPU, update2.CPU)
	}
}

// TestClusterUpdate_JSON tests JSON marshaling
func TestClusterUpdate_JSON(t *testing.T) {
	update := &ClusterUpdate{
		Type:      UpdateCluster,
		ClusterID: "cluster-1",
		Status:    "healthy",
		NodeCount: 5,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var update2 ClusterUpdate
	if err := json.Unmarshal(data, &update2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if update2.ClusterID != update.ClusterID {
		t.Errorf("expected ClusterID %s, got %s", update.ClusterID, update2.ClusterID)
	}
	if update2.NodeCount != update.NodeCount {
		t.Errorf("expected NodeCount %d, got %d", update.NodeCount, update2.NodeCount)
	}
}

// TestSubscriberInterface tests subscriber interface
func TestSubscriberInterface(t *testing.T) {
	// Create a mock subscriber
	var _ Subscriber = &mockSubscriber{}
}

type mockSubscriber struct {
	nodeUpdates    []*NodeUpdate
	clusterUpdates []*ClusterUpdate
}

func (m *mockSubscriber) OnNodeUpdate(u *NodeUpdate) {
	m.nodeUpdates = append(m.nodeUpdates, u)
}

func (m *mockSubscriber) OnClusterUpdate(u *ClusterUpdate) {
	m.clusterUpdates = append(m.clusterUpdates, u)
}

// TestMockSubscriber tests mock subscriber
func TestMockSubscriber(t *testing.T) {
	mock := &mockSubscriber{
		nodeUpdates:    make([]*NodeUpdate, 0),
		clusterUpdates: make([]*ClusterUpdate, 0),
	}

	// Test node update
	nodeUpdate := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		Timestamp: time.Now(),
	}
	mock.OnNodeUpdate(nodeUpdate)

	if len(mock.nodeUpdates) != 1 {
		t.Errorf("expected 1 node update, got %d", len(mock.nodeUpdates))
	}
	if mock.nodeUpdates[0].NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", mock.nodeUpdates[0].NodeID)
	}

	// Test cluster update
	clusterUpdate := &ClusterUpdate{
		Type:      UpdateCluster,
		ClusterID: "cluster-1",
		Status:    "healthy",
		NodeCount: 5,
		Timestamp: time.Now(),
	}
	mock.OnClusterUpdate(clusterUpdate)

	if len(mock.clusterUpdates) != 1 {
		t.Errorf("expected 1 cluster update, got %d", len(mock.clusterUpdates))
	}
	if mock.clusterUpdates[0].ClusterID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", mock.clusterUpdates[0].ClusterID)
	}
}

// TestWatcher_Create tests watcher creation
func TestWatcher_Create(t *testing.T) {
	db := &database.DB{}
	hosts := make(map[string]hypervisor.Hypervisor)

	watcher := NewWatcher(db, hosts)
	if watcher == nil {
		t.Fatal("expected non-nil watcher")
	}
}

// TestWatcher_Subscribe tests subscribing
func TestWatcher_Subscribe(t *testing.T) {
	db := &database.DB{}
	hosts := make(map[string]hypervisor.Hypervisor)
	watcher := NewWatcher(db, hosts)

	mock := &mockSubscriber{}
	watcher.Subscribe(mock)

	if len(watcher.subs) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(watcher.subs))
	}
}

// TestWatcher_Unsubscribe tests unsubscribing
func TestWatcher_Unsubscribe(t *testing.T) {
	db := &database.DB{}
	hosts := make(map[string]hypervisor.Hypervisor)
	watcher := NewWatcher(db, hosts)

	mock := &mockSubscriber{}
	watcher.Subscribe(mock)
	watcher.Unsubscribe(mock)

	if len(watcher.subs) != 0 {
		t.Errorf("expected 0 subscribers, got %d", len(watcher.subs))
	}
}

// TestWatcher_MultipleSubscribers tests multiple subscribers
func TestWatcher_MultipleSubscribers(t *testing.T) {
	db := &database.DB{}
	hosts := make(map[string]hypervisor.Hypervisor)
	watcher := NewWatcher(db, hosts)

	mock1 := &mockSubscriber{}
	mock2 := &mockSubscriber{}
	mock3 := &mockSubscriber{}

	watcher.Subscribe(mock1)
	watcher.Subscribe(mock2)
	watcher.Subscribe(mock3)

	if len(watcher.subs) != 3 {
		t.Errorf("expected 3 subscribers, got %d", len(watcher.subs))
	}

	watcher.Unsubscribe(mock2)

	if len(watcher.subs) != 2 {
		t.Errorf("expected 2 subscribers after unsubscribe, got %d", len(watcher.subs))
	}
}

// TestWatcher_Stop tests stopping the watcher
func TestWatcher_Stop(t *testing.T) {
	db := &database.DB{}
	hosts := make(map[string]hypervisor.Hypervisor)
	watcher := NewWatcher(db, hosts)

	// Stop should not panic
	watcher.Stop()
}

// TestWebSocketHub_Create tests hub creation
func TestWebSocketHub_Create(t *testing.T) {
	hub := NewWebSocketHub()
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.clients == nil {
		t.Error("expected initialized clients map")
	}
	if hub.broadcast == nil {
		t.Error("expected initialized broadcast channel")
	}
}

// TestWebSocketHub_Broadcast tests broadcasting
func TestWebSocketHub_Broadcast(t *testing.T) {
	hub := NewWebSocketHub()

	update := &NodeUpdate{
		Type:      UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		Timestamp: time.Now(),
	}

	// Broadcast should not block
	hub.Broadcast(update)
}

// TestWebSocketHub_Clients tests client management
func TestWebSocketHub_Clients(t *testing.T) {
	hub := NewWebSocketHub()

	client := &WebSocketClient{
		hub:        hub,
		send:       make(chan []byte, 256),
		nodeFilter: nil,
	}

	// Manually add/remove client
	hub.clients[client] = true
	if len(hub.clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(hub.clients))
	}

	delete(hub.clients, client)
	if len(hub.clients) != 0 {
		t.Errorf("expected 0 clients, got %d", len(hub.clients))
	}
}

// TestWebSocketClient_NodeFilter tests node filtering
func TestWebSocketClient_NodeFilter(t *testing.T) {
	hub := NewWebSocketHub()

	client := &WebSocketClient{
		hub:        hub,
		send:       make(chan []byte, 256),
		nodeFilter: []string{"node-1", "node-2"},
	}

	if len(client.nodeFilter) != 2 {
		t.Errorf("expected 2 node filters, got %d", len(client.nodeFilter))
	}
}

// TestContains tests contains helper
func TestContains(t *testing.T) {
	list := []string{"a", "b", "c"}

	if !contains(list, "a") {
		t.Error("expected to find 'a'")
	}
	if contains(list, "d") {
		t.Error("did not expect to find 'd'")
	}
	if !contains(list, "c") {
		t.Error("expected to find 'c'")
	}
}

// TestNodeUpdate_Fields tests all node update fields
func TestNodeUpdate_Fields(t *testing.T) {
	now := time.Now()
	update := &NodeUpdate{
		Type:      UpdateMetrics,
		NodeID:    "test-node",
		ClusterID: "test-cluster",
		State:     "stopped",
		IP:        "192.168.1.1",
		CPU:       75.5,
		Memory:    80.2,
		Disk:      45.0,
		Timestamp: now,
	}

	if update.Type != UpdateMetrics {
		t.Errorf("expected UpdateMetrics, got %s", update.Type)
	}
	if update.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", update.IP)
	}
	if update.Disk != 45.0 {
		t.Errorf("expected Disk 45.0, got %f", update.Disk)
	}
}

// TestClusterUpdate_Fields tests all cluster update fields
func TestClusterUpdate_Fields(t *testing.T) {
	now := time.Now()
	update := &ClusterUpdate{
		Type:      UpdateCluster,
		ClusterID: "prod-cluster",
		Status:    "degraded",
		NodeCount: 10,
		Timestamp: now,
	}

	if update.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", update.Status)
	}
	if update.NodeCount != 10 {
		t.Errorf("expected 10 nodes, got %d", update.NodeCount)
	}
}
