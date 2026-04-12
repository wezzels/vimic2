// Package status provides tests for status watcher
package status

import (
	"encoding/json"
	"testing"
	"time"
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
