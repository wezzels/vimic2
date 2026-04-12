// Package hypervisor provides cross-platform virtualization support
package hypervisor

import (
	"context"
	"testing"
	"time"
)

func TestStubHypervisor(t *testing.T) {
	hv := NewStubHypervisor()

	// Test CreateNode
	cfg := &NodeConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
	}

	node, err := hv.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	if node.Name != "test-vm" {
		t.Errorf("Expected name 'test-vm', got '%s'", node.Name)
	}
	if node.State != NodeRunning {
		t.Errorf("Expected state running, got %s", node.State)
	}
	if node.IP == "" {
		t.Error("Expected IP address to be set")
	}

	// Test ListNodes
	nodes, err := hv.ListNodes(context.Background())
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	// Test GetNode
	retrieved, err := hv.GetNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrieved.Name != "test-vm" {
		t.Errorf("Expected 'test-vm', got '%s'", retrieved.Name)
	}

	// Test GetNodeStatus
	status, err := hv.GetNodeStatus(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}
	if status.CPUPercent != 25.0 {
		t.Errorf("Expected 25%% CPU, got %%%.1f", status.CPUPercent)
	}

	// Test GetMetrics
	metrics, err := hv.GetMetrics(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}
	if metrics.CPU != 25.0 {
		t.Errorf("Expected 25%% CPU, got %%%.1f", metrics.CPU)
	}

	// Test StopNode
	err = hv.StopNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("StopNode failed: %v", err)
	}

	stoppedNode, _ := hv.GetNode(context.Background(), node.ID)
	if stoppedNode.State != NodeStopped {
		t.Errorf("Expected stopped state, got %s", stoppedNode.State)
	}

	// Test StartNode
	err = hv.StartNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("StartNode failed: %v", err)
	}

	startedNode, _ := hv.GetNode(context.Background(), node.ID)
	if startedNode.State != NodeRunning {
		t.Errorf("Expected running state, got %s", startedNode.State)
	}

	// Test RestartNode
	err = hv.RestartNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("RestartNode failed: %v", err)
	}

	// Test DeleteNode
	err = hv.DeleteNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	nodes, _ = hv.ListNodes(context.Background())
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes after delete, got %d", len(nodes))
	}
}

func TestNodeConfig(t *testing.T) {
	cfg := &NodeConfig{
		Name:     "test",
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
		Network:  "default",
	}

	if cfg.CPU != 4 {
		t.Errorf("Expected CPU 4, got %d", cfg.CPU)
	}
	if cfg.MemoryMB != 8192 {
		t.Errorf("Expected Memory 8192, got %d", cfg.MemoryMB)
	}
	if cfg.DiskGB != 100 {
		t.Errorf("Expected Disk 100, got %d", cfg.DiskGB)
	}
}

func TestMetrics(t *testing.T) {
	m := &Metrics{
		CPU:       50.5,
		Memory:    75.2,
		Disk:      30.0,
		NetworkRX: 100.5,
		NetworkTX: 50.3,
		Timestamp: time.Now(),
	}

	if m.CPU != 50.5 {
		t.Errorf("Expected CPU 50.5, got %.1f", m.CPU)
	}
	if m.Memory != 75.2 {
		t.Errorf("Expected Memory 75.2, got %.1f", m.Memory)
	}
}

func TestNodeState(t *testing.T) {
	states := []NodeState{NodePending, NodeRunning, NodeStopped, NodeError}

	for _, state := range states {
		if state == "" {
			t.Errorf("State should not be empty")
		}
	}

	if NodeRunning != "running" {
		t.Errorf("Expected 'running', got '%s'", NodeRunning)
	}
}
