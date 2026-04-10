// Package realhv_test tests the real hypervisor
package realhv_test

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realhv"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestRealHypervisor_Create tests hypervisor creation
func TestRealHypervisor_Create(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	if hv == nil {
		t.Fatal("hypervisor should not be nil")
	}
}

// TestRealHypervisor_Config tests configuration
func TestRealHypervisor_Config(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     30 * time.Second,
		MaxVMs:      10,
		AutoConnect: false,
	})

	if hv == nil {
		t.Fatal("hypervisor should not be nil")
	}
}

// TestRealHypervisor_NotConnected tests unconnected state
func TestRealHypervisor_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	if hv.IsConnected() {
		t.Error("should not be connected initially")
	}
}

// TestRealHypervisor_Types tests type definitions
func TestRealHypervisor_Types(t *testing.T) {
	// Test VMConfig
	cfg := &realhv.VMConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
		Network:  "default",
	}

	if cfg.Name != "test-vm" {
		t.Error("VMConfig name mismatch")
	}
	if cfg.CPU != 2 {
		t.Error("VMConfig CPU mismatch")
	}

	// Test VM
	vm := &realhv.VM{
		ID:      "vm-1",
		Name:    "test-vm",
		State:   hypervisor.NodeRunning,
		IP:      "192.168.122.10",
		Host:    "localhost",
		Config:  cfg,
		Created: time.Now(),
	}

	if vm.ID != "vm-1" {
		t.Error("VM ID mismatch")
	}
	if vm.State != hypervisor.NodeRunning {
		t.Error("VM state mismatch")
	}

	// Test VMStatus
	status := &realhv.VMStatus{
		State:       hypervisor.NodeRunning,
		Uptime:      time.Hour,
		CPUPercent:  25.5,
		MemUsed:     1024,
		MemTotal:    2048,
		DiskUsedGB:  10.0,
		DiskTotalGB: 20.0,
		IP:          "192.168.122.10",
	}

	if status.State != hypervisor.NodeRunning {
		t.Error("VMStatus state mismatch")
	}
	if status.CPUPercent != 25.5 {
		t.Error("VMStatus CPU mismatch")
	}

	// Test VMMetrics
	metrics := &realhv.VMMetrics{
		CPU:       30.0,
		Memory:    40.0,
		Disk:      50.0,
		NetworkRX: 1024000,
		NetworkTX: 512000,
		Timestamp: time.Now(),
	}

	if metrics.CPU != 30.0 {
		t.Error("VMMetrics CPU mismatch")
	}
}

// TestRealHypervisor_SetTimeout tests timeout setting
func TestRealHypervisor_SetTimeout(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	hv.SetTimeout(60 * time.Second)
	// No error means success
}

// TestRealHypervisor_SetMaxVMs tests max VMs setting
func TestRealHypervisor_SetMaxVMs(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	hv.SetMaxVMs(10)
	// No error means success
}

// TestRealHypervisor_AutoConnect tests auto-connect behavior
func TestRealHypervisor_AutoConnect(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	// With AutoConnect, operations will try to connect
	// But we can't test actual connection without libvirt

	if hv.IsConnected() {
		t.Error("should not be connected yet")
	}
}

// TestRealHypervisor_Factory tests factory
func TestRealHypervisor_Factory(t *testing.T) {
	factory := realhv.NewHypervisorFactory()

	if factory == nil {
		t.Fatal("factory should not be nil")
	}

	hv := factory.Create(nil)
	if hv == nil {
		t.Error("created hypervisor should not be nil")
	}

	hv = factory.CreateWithURI("qemu:///system")
	if hv == nil {
		t.Error("created hypervisor with URI should not be nil")
	}

	hv = factory.CreateRemote("192.168.1.100")
	if hv == nil {
		t.Error("created remote hypervisor should not be nil")
	}
}

// TestRealHypervisor_FactoryConfig tests factory config
func TestRealHypervisor_FactoryConfig(t *testing.T) {
	factory := &realhv.HypervisorFactory{
		DefaultURI:  "qemu+tcp://192.168.1.100/system",
		AutoConnect: false,
	}

	hv := factory.Create(nil)
	if hv == nil {
		t.Error("created hypervisor should not be nil")
	}
}

// TestRealHypervisor_Disconnect tests disconnect
func TestRealHypervisor_Disconnect(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	// Disconnect when not connected is fine
	err := hv.Disconnect()
	if err != nil {
		t.Errorf("disconnect should succeed when not connected: %v", err)
	}
}

// TestRealHypervisor_Close tests close
func TestRealHypervisor_Close(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	err := hv.Close()
	if err != nil {
		t.Errorf("close should succeed: %v", err)
	}
}

// TestRealHypervisor_Disconnect_Connected tests disconnect when connected
func TestRealHypervisor_Disconnect_Connected(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	err := hv.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if !hv.IsConnected() {
		t.Error("should be connected")
	}

	err = hv.Disconnect()
	if err != nil {
		t.Errorf("disconnect should succeed: %v", err)
	}

	if hv.IsConnected() {
		t.Error("should be disconnected")
	}
}

// TestRealHypervisor_CreateNode_WithStub tests create with stub
func TestRealHypervisor_CreateNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	cfg := &realhv.VMConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}

	vm, err := hv.CreateNode(ctx, cfg)
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	if vm.ID == "" {
		t.Error("VM ID should not be empty")
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected test-vm, got %s", vm.Name)
	}
	if vm.State == "" {
		t.Error("VM state should not be empty")
	}
}

// TestRealHypervisor_DeleteNode_WithStub tests delete with stub
func TestRealHypervisor_DeleteNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.DeleteNode(ctx, "vm-1")
	if err != nil {
		t.Errorf("DeleteNode should succeed with stub: %v", err)
	}
}

// TestRealHypervisor_StartNode_WithStub tests start with stub
func TestRealHypervisor_StartNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.StartNode(ctx, "vm-1")
	if err != nil {
		t.Errorf("StartNode should succeed with stub: %v", err)
	}
}

// TestRealHypervisor_StopNode_WithStub tests stop with stub
func TestRealHypervisor_StopNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.StopNode(ctx, "vm-1")
	if err != nil {
		t.Errorf("StopNode should succeed with stub: %v", err)
	}
}

// TestRealHypervisor_RestartNode_WithStub tests restart with stub
func TestRealHypervisor_RestartNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.RestartNode(ctx, "vm-1")
	if err != nil {
		t.Errorf("RestartNode should succeed with stub: %v", err)
	}
}

// TestRealHypervisor_GetNode_WithStub tests get node with stub
func TestRealHypervisor_GetNode_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	// Connect first
	err := hv.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create multiple nodes
	vm1, err := hv.CreateNode(ctx, &realhv.VMConfig{Name: "test-vm-1"})
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	vm2, err := hv.CreateNode(ctx, &realhv.VMConfig{Name: "test-vm-2"})
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	// Verify VMs were created with valid IDs
	if vm1.ID == "" || vm2.ID == "" {
		t.Error("VM ID should not be empty")
	}

	// List nodes
	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	// Stub hypervisor should maintain nodes in memory
	t.Logf("Created VMs: %s, %s; Found %d nodes", vm1.ID, vm2.ID, len(nodes))
}

// TestRealHypervisor_GetNodeStatus_WithStub tests get status with stub
func TestRealHypervisor_GetNodeStatus_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	status, err := hv.GetNodeStatus(ctx, "vm-1")
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}

	if status.State == "" {
		t.Error("status state should not be empty")
	}
}

// TestRealHypervisor_GetMetrics_WithStub tests get metrics with stub
func TestRealHypervisor_GetMetrics_WithStub(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: true,
	})

	ctx := context.Background()

	metrics, err := hv.GetMetrics(ctx, "vm-1")
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if metrics.Timestamp.IsZero() {
		t.Error("metrics timestamp should not be zero")
	}
}

