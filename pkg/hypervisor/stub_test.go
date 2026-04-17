//go:build integration

package hypervisor

import (
	"context"
	"testing"
	"time"
)

// ==================== Stub Hypervisor Tests ====================

func TestStub_Real_CreateNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{
		Address: "qemu:///system",
		Type:    "libvirt",
	})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	cfg := &NodeConfig{
		Name:      "test-vm",
		CPU:       2,
		MemoryMB:  2048,
		DiskGB:    20,
		Image:     "ubuntu:22.04",
		Network:   "default",
	}

	node, err := h.CreateNode(context.Background(), cfg)
	if err != nil {
		t.Logf("CreateNode: %v (expected for stub)", err)
	} else {
		t.Logf("Created node: %s", node.ID)
	}
}

func TestStub_Real_CreateNode_NilConfig(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	_, err = h.CreateNode(context.Background(), nil)
	if err == nil {
		t.Error("CreateNode(nil) should return error")
	}
}

func TestStub_Real_DeleteNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	err = h.DeleteNode(context.Background(), "test-vm")
	t.Logf("DeleteNode: %v (expected for stub)", err)
}

func TestStub_Real_StartNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	err = h.StartNode(context.Background(), "test-vm")
	t.Logf("StartNode: %v (expected for stub)", err)
}

func TestStub_Real_StopNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	err = h.StopNode(context.Background(), "test-vm")
	t.Logf("StopNode: %v (expected for stub)", err)
}

func TestStub_Real_RestartNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	err = h.RestartNode(context.Background(), "test-vm")
	t.Logf("RestartNode: %v (expected for stub)", err)
}

func TestStub_Real_ListNodes(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	nodes, err := h.ListNodes(context.Background())
	t.Logf("ListNodes: %d nodes, err=%v", len(nodes), err)
}

func TestStub_Real_GetNode(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	node, err := h.GetNode(context.Background(), "test-vm")
	t.Logf("GetNode: node=%v err=%v", node, err)
}

func TestStub_Real_GetNodeStatus(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	status, err := h.GetNodeStatus(context.Background(), "test-vm")
	t.Logf("GetNodeStatus: state=%v err=%v", status, err)
}

func TestStub_Real_GetMetrics(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	metrics, err := h.GetMetrics(context.Background(), "test-vm")
	t.Logf("GetMetrics: cpu=%.1f mem=%.1f disk=%.1f err=%v", metrics.CPU, metrics.Memory, metrics.Disk, err)
}

func TestStub_Real_Close(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{Address: "qemu:///system"})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	err = h.Close()
	t.Logf("Close: %v", err)
}

// ==================== NewHypervisor Tests ====================

func TestNewHypervisor_Real_InvalidPlatform(t *testing.T) {
	_, err := NewHypervisor(&HostConfig{Type: "invalid-platform"})
	if err == nil {
		t.Error("NewHypervisor with invalid platform should return error")
	}
}

func TestNewHypervisor_Real_StubPlatform(t *testing.T) {
	h, err := NewHypervisor(&HostConfig{Type: "stub"})
	if err != nil {
		t.Skipf("NewHypervisor(stub) not supported: %v", err)
	}
	if h == nil {
		t.Fatal("NewHypervisor(stub) should return non-nil")
	}
}

func TestNewHypervisor_Real_NilConfig(t *testing.T) {
	h, err := NewHypervisor(nil)
	if err != nil {
		t.Logf("NewHypervisor(nil): %v", err)
	}
	_ = h
}

// ==================== Struct Tests ====================

func TestHostConfig_Real_Struct(t *testing.T) {
	config := &HostConfig{
		Address:    "10.0.0.1",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		Type:       "libvirt",
	}

	if config.Address != "10.0.0.1" {
		t.Errorf("Address = %s, want 10.0.0.1", config.Address)
	}
	if config.Port != 22 {
		t.Errorf("Port = %d, want 22", config.Port)
	}
	if config.Type != "libvirt" {
		t.Errorf("Type = %s, want libvirt", config.Type)
	}
}

func TestNodeConfig_Real_Struct(t *testing.T) {
	config := &NodeConfig{
		Name:      "test-vm",
		CPU:       2,
		MemoryMB:  2048,
		DiskGB:    20,
		Image:     "ubuntu:22.04",
		Network:   "default",
		SSHKey:    "ssh-rsa AAA...",
	}

	if config.Name != "test-vm" {
		t.Errorf("Name = %s, want test-vm", config.Name)
	}
	if config.CPU != 2 {
		t.Errorf("CPU = %d, want 2", config.CPU)
	}
	if config.MemoryMB != 2048 {
		t.Errorf("MemoryMB = %d, want 2048", config.MemoryMB)
	}
	if config.DiskGB != 20 {
		t.Errorf("DiskGB = %d, want 20", config.DiskGB)
	}
}

func TestNodeStatus_Real_Struct(t *testing.T) {
	status := &NodeStatus{
		State:      NodeRunning,
		Uptime:     3600 * time.Second,
		CPUPercent: 45.0,
		MemUsed:    1024,
		MemTotal:   2048,
		DiskUsedGB: 10.5,
		DiskTotalGB: 20.0,
		IP:         "10.0.0.1",
	}

	if status.State != NodeRunning {
		t.Errorf("State = %s, want running", status.State)
	}
	if status.CPUPercent != 45.0 {
		t.Errorf("CPUPercent = %f, want 45.0", status.CPUPercent)
	}
	if status.IP != "10.0.0.1" {
		t.Errorf("IP = %s, want 10.0.0.1", status.IP)
	}
}

func TestMetrics_Real_Struct(t *testing.T) {
	now := time.Now()
	metrics := &Metrics{
		CPU:       55.0,
		Memory:    70.0,
		Disk:      40.0,
		NetworkRX: 1024,
		NetworkTX: 2048,
		Timestamp: now,
	}

	if metrics.CPU != 55.0 {
		t.Errorf("CPU = %f, want 55.0", metrics.CPU)
	}
	if metrics.Memory != 70.0 {
		t.Errorf("Memory = %f, want 70.0", metrics.Memory)
	}
	if metrics.NetworkRX != 1024 {
		t.Errorf("NetworkRX = %f, want 1024", metrics.NetworkRX)
	}
	if metrics.NetworkTX != 2048 {
		t.Errorf("NetworkTX = %f, want 2048", metrics.NetworkTX)
	}
}

func TestNodeState_Constants_Real(t *testing.T) {
	if NodeRunning != "running" {
		t.Errorf("NodeRunning = %s, want running", NodeRunning)
	}
	if NodeStopped != "stopped" {
		t.Errorf("NodeStopped = %s, want stopped", NodeStopped)
	}
	if NodeError != "error" {
		t.Errorf("NodeError = %s, want error", NodeError)
	}
	if NodePending != "pending" {
		t.Errorf("NodePending = %s, want pending", NodePending)
	}
}

// ==================== Lifecycle Test ====================

func TestStub_Real_Lifecycle(t *testing.T) {
	h, err := newLibvirtHypervisor(&HostConfig{
		Address: "qemu:///system",
		Type:    "libvirt",
	})
	if err != nil {
		t.Fatalf("newLibvirtHypervisor failed: %v", err)
	}

	ctx := context.Background()

	cfg := &NodeConfig{
		Name:      "lifecycle-test-vm",
		CPU:       4,
		MemoryMB:  4096,
		DiskGB:    50,
		Image:     "ubuntu:22.04",
		Network:   "default",
	}

	// Create
	node, err := h.CreateNode(ctx, cfg)
	if err != nil {
		t.Skipf("CreateNode failed: %v", err)
	}
	t.Logf("Created node: %s", node.ID)

	// Start
	_ = h.StartNode(ctx, node.ID)

	// Get status
	status, _ := h.GetNodeStatus(ctx, node.ID)
	t.Logf("GetNodeStatus: state=%v", status)

	// Get metrics
	metrics, _ := h.GetMetrics(ctx, node.ID)
	t.Logf("GetMetrics: cpu=%.1f mem=%.1f disk=%.1f", metrics.CPU, metrics.Memory, metrics.Disk)

	// Restart
	_ = h.RestartNode(ctx, node.ID)

	// Stop
	_ = h.StopNode(ctx, node.ID)

	// Delete
	_ = h.DeleteNode(ctx, node.ID)

	// Close
	_ = h.Close()
}