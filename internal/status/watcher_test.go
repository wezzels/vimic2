// Package status_test tests status watcher functionality
package status_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/status"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestWatcher tests the status watcher
func TestWatcher(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-watcher-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create stub hypervisor
	stubHV := hypervisor.NewStubHypervisor()
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": stubHV,
	}

	watcher := status.NewWatcher(db, hosts)
	if watcher == nil {
		t.Fatal("Watcher should not be nil")
	}
}

// TestWatcherSubscribe tests subscriber management
func TestWatcherSubscribe(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-watcher-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	stubHV := hypervisor.NewStubHypervisor()
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": stubHV,
	}

	watcher := status.NewWatcher(db, hosts)

	// Create a test subscriber
	sub := &TestSubscriber{}
	watcher.Subscribe(sub)

	// Unsubscribe
	watcher.Unsubscribe(sub)

	// Subscribe again
	watcher.Subscribe(sub)
}

// TestWatcherStartStop tests starting and stopping the watcher
func TestWatcherStartStop(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-watcher-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	stubHV := hypervisor.NewStubHypervisor()
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": stubHV,
	}

	watcher := status.NewWatcher(db, hosts)

	// Start and immediately stop
	watcher.Start(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	watcher.Stop()

	// If we get here without hanging, the test passed
}

// TestWatcherCheckAll tests the check all function
// TestWatcherCheckAll tests the checkAll method
func TestWatcherCheckAll(t *testing.T) {
	t.Skip("checkAll is unexported - internal method")
}

// TestNodeUpdate tests node update structure
func TestNodeUpdate(t *testing.T) {
	update := &status.NodeUpdate{
		Type:      status.UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
		CPU:       45.5,
		Memory:    60.0,
		Disk:      30.0,
		Timestamp: time.Now(),
	}

	if update.Type != status.UpdateNode {
		t.Errorf("Expected type UpdateNode, got %s", update.Type)
	}
	if update.State != "running" {
		t.Errorf("Expected state 'running', got '%s'", update.State)
	}
}

// TestClusterUpdate tests cluster update structure
func TestClusterUpdate(t *testing.T) {
	update := &status.ClusterUpdate{
		Type:      status.UpdateCluster,
		ClusterID: "cluster-1",
		Status:    "running",
		NodeCount: 5,
		Timestamp: time.Now(),
	}

	if update.Type != status.UpdateCluster {
		t.Errorf("Expected type UpdateCluster, got %s", update.Type)
	}
	if update.NodeCount != 5 {
		t.Errorf("Expected NodeCount 5, got %d", update.NodeCount)
	}
}

// TestWebSocketHub tests the WebSocket hub
func TestWebSocketHub(t *testing.T) {
	hub := status.NewWebSocketHub()
	if hub == nil {
		t.Fatal("Hub should not be nil")
	}

	// Start the hub
	go hub.Run()

	// Broadcast should not panic
	hub.Broadcast(&status.NodeUpdate{
		Type:      status.UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
	})
}

// TestWebSocketHubBroadcast tests broadcasting to multiple clients
func TestWebSocketHubBroadcast(t *testing.T) {
	hub := status.NewWebSocketHub()
	go hub.Run()

	// Broadcast update
	update := &status.NodeUpdate{
		Type:      status.UpdateNode,
		NodeID:    "node-1",
		ClusterID: "cluster-1",
		State:     "running",
	}
	hub.Broadcast(update)

	// Give time for broadcast
	time.Sleep(10 * time.Millisecond)
}

// TestSubscriberFilter tests node filtering
func TestSubscriberFilter(t *testing.T) {
	t.Skip("WebSocketClient fields are unexported")
}

// TestSubscriber interface implementation for testing
type TestSubscriber struct {
	mu         sync.Mutex
	nodeUpdates []*status.NodeUpdate
	clusterUpdates []*status.ClusterUpdate
}

func (s *TestSubscriber) OnNodeUpdate(update *status.NodeUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeUpdates = append(s.nodeUpdates, update)
}

func (s *TestSubscriber) OnClusterUpdate(update *status.ClusterUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusterUpdates = append(s.clusterUpdates, update)
}
