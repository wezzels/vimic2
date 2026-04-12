// Package mockhv_test tests the mock hypervisor
package mockhv_test

import (
	"context"
	"testing"

	"github.com/stsgym/vimic2/internal/testutil/mockhv"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestMockHypervisor_Create tests hypervisor creation
func TestMockHypervisor_Create(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	if hv == nil {
		t.Fatal("hypervisor should not be nil")
	}
}

// TestMockHypervisor_CreateNode tests node creation
func TestMockHypervisor_CreateNode(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}

	node, err := hv.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create node: %v", err)
	}

	if node.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", node.Name)
	}
	if node.State != hypervisor.NodeRunning {
		t.Errorf("expected running state, got %s", node.State)
	}
	if node.IP == "" {
		t.Error("expected IP address")
	}
}

// TestMockHypervisor_ListNodes tests node listing
func TestMockHypervisor_ListNodes(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	// Create nodes
	cfg1 := &hypervisor.NodeConfig{Name: "vm-1", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	cfg2 := &hypervisor.NodeConfig{Name: "vm-2", CPU: 2, MemoryMB: 2048, DiskGB: 20}

	_, _ = hv.CreateNode(context.Background(), cfg1)
	_, _ = hv.CreateNode(context.Background(), cfg2)

	nodes, err := hv.ListNodes(context.Background())
	if err != nil {
		t.Fatalf("failed to list nodes: %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

// TestMockHypervisor_GetNode tests node retrieval
func TestMockHypervisor_GetNode(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	retrieved, err := hv.GetNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	if retrieved.ID != node.ID {
		t.Errorf("expected ID %s, got %s", node.ID, retrieved.ID)
	}
}

// TestMockHypervisor_DeleteNode tests node deletion
func TestMockHypervisor_DeleteNode(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	err := hv.DeleteNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to delete node: %v", err)
	}

	_, err = hv.GetNode(context.Background(), node.ID)
	if err == nil {
		t.Error("node should be deleted")
	}
}

// TestMockHypervisor_StartStop tests start/stop operations
func TestMockHypervisor_StartStop(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	// Stop
	err := hv.StopNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to stop node: %v", err)
	}

	retrieved, _ := hv.GetNode(context.Background(), node.ID)
	if retrieved.State != hypervisor.NodeStopped {
		t.Errorf("expected stopped state, got %s", retrieved.State)
	}

	// Start
	err = hv.StartNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to start node: %v", err)
	}

	retrieved, _ = hv.GetNode(context.Background(), node.ID)
	if retrieved.State != hypervisor.NodeRunning {
		t.Errorf("expected running state, got %s", retrieved.State)
	}

	// Restart
	err = hv.RestartNode(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to restart node: %v", err)
	}
}

// TestMockHypervisor_GetMetrics tests metrics retrieval
func TestMockHypervisor_GetMetrics(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	metrics, err := hv.GetMetrics(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}

	if metrics.CPU < 0 || metrics.CPU > 100 {
		t.Errorf("invalid CPU percentage: %f", metrics.CPU)
	}
	if metrics.Memory < 0 || metrics.Memory > 100 {
		t.Errorf("invalid memory percentage: %f", metrics.Memory)
	}
}

// TestMockHypervisor_GetNodeStatus tests status retrieval
func TestMockHypervisor_GetNodeStatus(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	status, err := hv.GetNodeStatus(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.State != hypervisor.NodeRunning {
		t.Errorf("expected running state, got %s", status.State)
	}
	if status.IP == "" {
		t.Error("expected IP address")
	}
}

// TestMockHypervisor_ErrorMode tests error mode
func TestMockHypervisor_ErrorMode(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	// Enable error mode
	hv.SetErrorMode(true)

	// Operations should fail
	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	_, err := hv.CreateNode(context.Background(), cfg)
	if err == nil {
		t.Error("expected error in error mode")
	}

	// Disable error mode
	hv.SetErrorMode(false)

	// Operations should succeed
	_, err = hv.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockHypervisor_FailNext tests fail next
func TestMockHypervisor_FailNext(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	// Set up fail next
	hv.FailNext()

	// This operation should fail
	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	_, err := hv.CreateNode(context.Background(), cfg)
	if err == nil {
		t.Error("expected error from fail next")
	}

	// This operation should succeed (failNext was consumed)
	_, err = hv.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockHypervisor_SetNodeState tests setting node state
func TestMockHypervisor_SetNodeState(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	// Set state
	err := hv.SetNodeState(node.ID, hypervisor.NodeStopped)
	if err != nil {
		t.Fatalf("failed to set state: %v", err)
	}

	retrieved, _ := hv.GetNode(context.Background(), node.ID)
	if retrieved.State != hypervisor.NodeStopped {
		t.Errorf("expected stopped, got %s", retrieved.State)
	}
}

// TestMockHypervisor_SetNodeIP tests setting node IP
func TestMockHypervisor_SetNodeIP(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	// Set IP
	err := hv.SetNodeIP(node.ID, "10.100.1.10")
	if err != nil {
		t.Fatalf("failed to set IP: %v", err)
	}

	retrieved, _ := hv.GetNode(context.Background(), node.ID)
	if retrieved.IP != "10.100.1.10" {
		t.Errorf("expected 10.100.1.10, got %s", retrieved.IP)
	}
}

// TestMockHypervisor_Count tests node counting
func TestMockHypervisor_Count(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}

	// Create 5 nodes
	for i := 0; i < 5; i++ {
		_, _ = hv.CreateNode(context.Background(), cfg)
	}

	if hv.Count() != 5 {
		t.Errorf("expected 5 nodes, got %d", hv.Count())
	}
}

// TestMockHypervisor_CountByState tests counting by state
func TestMockHypervisor_CountByState(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}

	// Create 5 running nodes
	for i := 0; i < 5; i++ {
		node, _ := hv.CreateNode(context.Background(), cfg)
		if i < 2 {
			// Stop 2 nodes
			_ = hv.SetNodeState(node.ID, hypervisor.NodeStopped)
		}
	}

	running := hv.CountByState(hypervisor.NodeRunning)
	stopped := hv.CountByState(hypervisor.NodeStopped)

	if running != 3 {
		t.Errorf("expected 3 running nodes, got %d", running)
	}
	if stopped != 2 {
		t.Errorf("expected 2 stopped nodes, got %d", stopped)
	}
}

// TestMockHypervisor_Close tests closing
func TestMockHypervisor_Close(t *testing.T) {
	hv := mockhv.NewMockHypervisor()

	cfg := &hypervisor.NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	_, _ = hv.CreateNode(context.Background(), cfg)

	// Close should clear nodes
	err := hv.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	if hv.Count() != 0 {
		t.Errorf("expected 0 nodes after close, got %d", hv.Count())
	}
}

// TestMockHypervisor_CreateWithNodes tests factory function
func TestMockHypervisor_CreateWithNodes(t *testing.T) {
	configs := []*hypervisor.NodeConfig{
		{Name: "vm-1", CPU: 2, MemoryMB: 2048, DiskGB: 20},
		{Name: "vm-2", CPU: 2, MemoryMB: 2048, DiskGB: 20},
		{Name: "vm-3", CPU: 2, MemoryMB: 2048, DiskGB: 20},
	}

	hv, err := mockhv.CreateWithNodes(configs)
	if err != nil {
		t.Fatalf("failed to create with nodes: %v", err)
	}

	if hv.Count() != 3 {
		t.Errorf("expected 3 nodes, got %d", hv.Count())
	}
}

// TestMockHypervisorFactory tests factory creation
func TestMockHypervisorFactory(t *testing.T) {
	factory := mockhv.NewMockHypervisorFactory()

	cfg := &hypervisor.HostConfig{
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		Type:       "libvirt",
	}

	hv, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("failed to create hypervisor: %v", err)
	}

	if hv == nil {
		t.Error("hypervisor should not be nil")
	}
}

// TestMockHypervisorFactory_WithDelay tests factory with delay
func TestMockHypervisorFactory_WithDelay(t *testing.T) {
	factory := &mockhv.MockHypervisorFactory{
		DefaultDelay: 0, // No delay for tests
	}

	cfg := &hypervisor.HostConfig{
		Address: "192.168.1.100",
		Port:    22,
		User:    "root",
		Type:    "libvirt",
	}

	hv, err := factory.Create(cfg)
	if err != nil {
		t.Fatalf("failed to create hypervisor: %v", err)
	}

	if hv == nil {
		t.Error("hypervisor should not be nil")
	}
}
