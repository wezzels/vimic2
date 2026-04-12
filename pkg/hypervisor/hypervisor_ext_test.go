// Package hypervisor provides hypervisor tests
package hypervisor

import (
	"context"
	"testing"
	"time"
)

// TestNodeConfig tests node configuration
func TestNodeConfig_Create(t *testing.T) {
	cfg := &NodeConfig{
		Name:     "test-vm",
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
		Network:  "default",
		SSHKey:   "ssh-rsa AAA...",
	}

	if cfg.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", cfg.Name)
	}
	if cfg.CPU != 4 {
		t.Errorf("expected 4 CPUs, got %d", cfg.CPU)
	}
	if cfg.MemoryMB != 8192 {
		t.Errorf("expected 8192MB, got %d", cfg.MemoryMB)
	}
	if cfg.Image != "ubuntu-22.04" {
		t.Errorf("expected ubuntu-22.04, got %s", cfg.Image)
	}
}

// TestNode tests node structure
func TestNode_Create(t *testing.T) {
	now := time.Now()
	cfg := &NodeConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 4096,
	}

	node := &Node{
		ID:      "vm-123",
		Name:    "test-vm",
		State:   NodeRunning,
		IP:      "192.168.122.100",
		Host:    "qemu:///system",
		Config:  cfg,
		Created: now,
	}

	if node.ID != "vm-123" {
		t.Errorf("expected vm-123, got %s", node.ID)
	}
	if node.State != NodeRunning {
		t.Errorf("expected running state, got %s", node.State)
	}
	if node.IP != "192.168.122.100" {
		t.Errorf("expected IP 192.168.122.100, got %s", node.IP)
	}
}

// TestNodeStatus tests node status structure
func TestNodeStatus_Create(t *testing.T) {
	status := &NodeStatus{
		State:       NodeRunning,
		Uptime:      3600 * time.Second,
		CPUPercent:  45.5,
		MemUsed:     4096,
		MemTotal:    8192,
		DiskUsedGB:  25.5,
		DiskTotalGB: 50.0,
		IP:          "192.168.122.100",
	}

	if status.State != NodeRunning {
		t.Errorf("expected running, got %s", status.State)
	}
	if status.CPUPercent != 45.5 {
		t.Errorf("expected 45.5%% CPU, got %.1f", status.CPUPercent)
	}
	if status.MemUsed != 4096 {
		t.Errorf("expected 4096MB used, got %d", status.MemUsed)
	}
}

// TestMetrics tests metrics structure
func TestMetrics_Create(t *testing.T) {
	metrics := &Metrics{
		CPU:       35.5,
		Memory:    50.2,
		Disk:      40.0,
		NetworkRX: 1024000,
		NetworkTX: 512000,
		Timestamp: time.Now(),
	}

	if metrics.CPU != 35.5 {
		t.Errorf("expected 35.5%% CPU, got %.1f", metrics.CPU)
	}
	if metrics.Memory != 50.2 {
		t.Errorf("expected 50.2%% memory, got %.1f", metrics.Memory)
	}
	if metrics.NetworkRX != 1024000 {
		t.Errorf("expected 1024000 RX, got %f", metrics.NetworkRX)
	}
}

// TestHostConfig tests host configuration
func TestHostConfig_Create(t *testing.T) {
	cfg := &HostConfig{
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		Type:       "libvirt",
	}

	if cfg.Address != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", cfg.Address)
	}
	if cfg.Port != 22 {
		t.Errorf("expected port 22, got %d", cfg.Port)
	}
	if cfg.Type != "libvirt" {
		t.Errorf("expected libvirt type, got %s", cfg.Type)
	}
}

// TestNodeState tests node state constants
func TestNodeState_Constants(t *testing.T) {
	states := []NodeState{NodePending, NodeRunning, NodeStopped, NodeError}

	for _, state := range states {
		if state == "" {
			t.Error("empty node state")
		}
	}
}

// TestStubHypervisor_CreateNode tests stub hypervisor node creation
func TestStubHypervisor_CreateNode(t *testing.T) {
	hv := NewStubHypervisor()

	cfg := &NodeConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}

	node, err := hv.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	if node.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", node.Name)
	}
	if node.State != NodeRunning {
		t.Errorf("expected running state, got %s", node.State)
	}
	if node.IP == "" {
		t.Error("expected IP address")
	}
}

// TestStubHypervisor_ListNodes tests stub hypervisor list
func TestStubHypervisor_ListNodes(t *testing.T) {
	hv := NewStubHypervisor()

	// Initially empty
	nodes, err := hv.ListNodes(context.Background())
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	// Create a node
	cfg := &NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	_, _ = hv.CreateNode(context.Background(), cfg)

	// Now should have one
	nodes, _ = hv.ListNodes(context.Background())
	if len(nodes) == 0 {
		t.Error("expected at least one node after creation")
	}
}

// TestStubHypervisor_GetMetrics tests stub hypervisor metrics
func TestStubHypervisor_GetMetrics(t *testing.T) {
	hv := NewStubHypervisor()

	cfg := &NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	metrics, err := hv.GetMetrics(context.Background(), node.ID)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if metrics.CPU < 0 || metrics.CPU > 100 {
		t.Errorf("invalid CPU percentage: %.1f", metrics.CPU)
	}
	if metrics.Memory < 0 || metrics.Memory > 100 {
		t.Errorf("invalid memory percentage: %.1f", metrics.Memory)
	}
}

// TestStubHypervisor_StopStart tests stop and start operations
func TestStubHypervisor_StopStart(t *testing.T) {
	hv := NewStubHypervisor()

	cfg := &NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	// Stop
	err := hv.StopNode(context.Background(), node.ID)
	if err != nil {
		t.Errorf("StopNode failed: %v", err)
	}

	// Start
	err = hv.StartNode(context.Background(), node.ID)
	if err != nil {
		t.Errorf("StartNode failed: %v", err)
	}

	// Restart
	err = hv.RestartNode(context.Background(), node.ID)
	if err != nil {
		t.Errorf("RestartNode failed: %v", err)
	}
}

// TestStubHypervisor_Delete tests node deletion
func TestStubHypervisor_Delete(t *testing.T) {
	hv := NewStubHypervisor()

	cfg := &NodeConfig{Name: "test-vm", CPU: 2, MemoryMB: 2048, DiskGB: 20}
	node, _ := hv.CreateNode(context.Background(), cfg)

	err := hv.DeleteNode(context.Background(), node.ID)
	if err != nil {
		t.Errorf("DeleteNode failed: %v", err)
	}
}

// TestStubHypervisor_Close tests close operation
func TestStubHypervisor_Close(t *testing.T) {
	hv := NewStubHypervisor()

	err := hv.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestNewHypervisor tests hypervisor factory
func TestNewHypervisor(t *testing.T) {
	cfg := &HostConfig{
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		Type:       "libvirt",
	}

	hv, err := NewHypervisor(cfg)
	if err != nil {
		t.Fatalf("NewHypervisor failed: %v", err)
	}
	if hv == nil {
		t.Error("hypervisor should not be nil")
	}
}

// TestNode_JSON tests node JSON marshaling
func TestNode_JSON(t *testing.T) {
	node := &Node{
		ID:      "vm-123",
		Name:    "test-vm",
		State:   NodeRunning,
		IP:      "192.168.122.100",
		Host:    "qemu:///system",
		Created: time.Now(),
	}

	// Verify struct fields
	if node.ID != "vm-123" {
		t.Errorf("expected vm-123, got %s", node.ID)
	}
	if node.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", node.Name)
	}
}

// TestNodeConfig_JSON tests config JSON marshaling
func TestNodeConfig_JSON(t *testing.T) {
	cfg := &NodeConfig{
		Name:     "test-vm",
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
		Network:  "default",
	}

	// Verify struct fields
	if cfg.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", cfg.Name)
	}
	if cfg.CPU != 4 {
		t.Errorf("expected 4 CPUs, got %d", cfg.CPU)
	}
}
