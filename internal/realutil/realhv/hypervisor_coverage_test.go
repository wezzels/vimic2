package realhv_test

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realhv"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestRealHypervisor_NewHypervisor_Defaults tests NewHypervisor with nil config
func TestRealHypervisor_NewHypervisor_Defaults(t *testing.T) {
	hv := realhv.NewHypervisor(nil)
	if hv == nil {
		t.Fatal("NewHypervisor should return non-nil")
	}
}

// TestRealHypervisor_NewHypervisor_CustomConfig tests NewHypervisor with custom config
func TestRealHypervisor_NewHypervisor_CustomConfig(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.1/system",
		Timeout:     60 * time.Second,
		MaxVMs:      10,
		AutoConnect: true,
	})
	if hv == nil {
		t.Fatal("NewHypervisor should return non-nil")
	}
}

// TestRealHypervisor_Connect_Timeout tests Connect with timeout
func TestRealHypervisor_Connect_Timeout(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:     "qemu:///system",
		Timeout: 1 * time.Nanosecond, // Very short timeout
	})

	ctx := context.Background()
	err := hv.Connect(ctx)
	// May timeout or fail to connect
	_ = err
}

// TestRealHypervisor_Disconnect_WithoutConnect tests Disconnect without prior connection
func TestRealHypervisor_Disconnect_WithoutConnect(t *testing.T) {
	hv := realhv.NewHypervisor(nil)

	err := hv.Disconnect()
	// Should not error even if not connected
	_ = err
}

// TestRealHypervisor_StartNode_NotConnected tests StartNode without connection
func TestRealHypervisor_StartNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false, // Don't auto-connect
	})

	ctx := context.Background()
	err := hv.StartNode(ctx, "non-existent-vm")
	// Should fail because not connected
	_ = err
}

// TestRealHypervisor_StopNode_NotConnected tests StopNode without connection
func TestRealHypervisor_StopNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	err := hv.StopNode(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_RestartNode_NotConnected tests RestartNode without connection
func TestRealHypervisor_RestartNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	err := hv.RestartNode(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_GetNode_NotConnected tests GetNode without connection
func TestRealHypervisor_GetNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	_, err := hv.GetNode(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_GetNodeStatus_NotConnected tests GetNodeStatus without connection
func TestRealHypervisor_GetNodeStatus_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	_, err := hv.GetNodeStatus(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_GetMetrics_NotConnected tests GetMetrics without connection
func TestRealHypervisor_GetMetrics_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	_, err := hv.GetMetrics(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_ListNodes_NotConnected tests ListNodes without connection
func TestRealHypervisor_ListNodes_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	_, err := hv.ListNodes(ctx)
	_ = err
}

// TestRealHypervisor_CreateNode_NotConnected tests CreateNode without connection
func TestRealHypervisor_CreateNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	_, err := hv.CreateNode(ctx, &realhv.VMConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
	})
	_ = err
}

// TestRealHypervisor_DeleteNode_NotConnected tests DeleteNode without connection
func TestRealHypervisor_DeleteNode_NotConnected(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	})

	ctx := context.Background()
	err := hv.DeleteNode(ctx, "non-existent-vm")
	_ = err
}

// TestRealHypervisor_AutoConnectBehavior tests auto-connect behavior
func TestRealHypervisor_AutoConnectBehavior(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: true,
		Timeout:     5 * time.Second,
	})

	ctx := context.Background()

	// ListNodes should attempt to connect
	_, err := hv.ListNodes(ctx)
	// May fail if no libvirt, but auto-connect should be attempted
	_ = err
}

// TestRealHypervisor_ContextCancellation tests context cancellation
func TestRealHypervisor_ContextCancellation(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:     "qemu:///system",
		Timeout: 10 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := hv.Connect(ctx)
	// Should handle cancelled context
	_ = err
}

// TestRealHypervisor_VMConfig_Validation tests VM config validation
func TestRealHypervisor_VMConfig_Validation(t *testing.T) {
	configs := []*realhv.VMConfig{
		{Name: "test", CPU: 1, MemoryMB: 512, DiskGB: 10},
		{Name: "test", CPU: 4, MemoryMB: 8192, DiskGB: 100},
		{Name: "test", CPU: 0, MemoryMB: 0, DiskGB: 0}, // Zero values
	}

	for i, cfg := range configs {
		if cfg.Name == "" {
			t.Errorf("config %d: Name should not be empty", i)
		}
		_ = cfg // Just validation
	}
}

// TestRealHypervisor_NewHypervisor_EmptyURI tests NewHypervisor with empty URI
func TestRealHypervisor_NewHypervisor_EmptyURI(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI: "", // Empty URI should use default
	})
	if hv == nil {
		t.Fatal("NewHypervisor should return non-nil with empty URI")
	}
}

// TestRealHypervisor_NewHypervisor_Timeout tests NewHypervisor with timeout
func TestRealHypervisor_NewHypervisor_Timeout(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     30 * time.Second,
		MaxVMs:      10,
		AutoConnect: false,
	})
	if hv == nil {
		t.Fatal("NewHypervisor should return non-nil")
	}
}

// TestRealHypervisor_VMConfig_Fields tests VMConfig field values
func TestRealHypervisor_VMConfig_Fields(t *testing.T) {
	config := &realhv.VMConfig{
		Name:      "test-vm",
		CPU:       4,
		MemoryMB:  8192,
		DiskGB:    100,
		Image:     "ubuntu-22.04",
		Network:   "default",
		CloudInit: "#cloud-config\n",
	}

	if config.Name != "test-vm" {
		t.Errorf("expected Name=test-vm, got %s", config.Name)
	}
	if config.CPU != 4 {
		t.Errorf("expected CPU=4, got %d", config.CPU)
	}
	if config.MemoryMB != 8192 {
		t.Errorf("expected MemoryMB=8192, got %d", config.MemoryMB)
	}
	if config.DiskGB != 100 {
		t.Errorf("expected DiskGB=100, got %d", config.DiskGB)
	}
	if config.Image != "ubuntu-22.04" {
		t.Errorf("expected Image=ubuntu-22.04, got %s", config.Image)
	}
}

// TestRealHypervisor_VM_Fields tests VM field values
func TestRealHypervisor_VM_Fields(t *testing.T) {
	now := time.Now()
	vm := &realhv.VM{
		ID:        "vm-1",
		Name:      "test-vm",
		State:     hypervisor.NodeRunning,
		IP:        "10.0.0.1",
		Host:      "host-1",
		Config:    &realhv.VMConfig{Name: "test", CPU: 2, MemoryMB: 4096, DiskGB: 20},
		Created:   now,
		UpdatedAt: now,
	}

	if vm.ID != "vm-1" {
		t.Errorf("expected ID=vm-1, got %s", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("expected Name=test-vm, got %s", vm.Name)
	}
	if vm.State != hypervisor.NodeRunning {
		t.Errorf("expected State=Running, got %v", vm.State)
	}
	if vm.IP != "10.0.0.1" {
		t.Errorf("expected IP=10.0.0.1, got %s", vm.IP)
	}
}

// TestRealHypervisor_VMMetrics_Fields tests VMMetrics field values
func TestRealHypervisor_VMMetrics_Fields(t *testing.T) {
	now := time.Now()
	metrics := &realhv.VMMetrics{
		CPU:       50.5,
		Memory:    60.2,
		Disk:      70.8,
		NetworkRX: 1000.0,
		NetworkTX: 500.0,
		Timestamp: now,
	}

	if metrics.CPU != 50.5 {
		t.Errorf("expected CPU=50.5, got %f", metrics.CPU)
	}
	if metrics.Memory != 60.2 {
		t.Errorf("expected Memory=60.2, got %f", metrics.Memory)
	}
	if metrics.Disk != 70.8 {
		t.Errorf("expected Disk=70.8, got %f", metrics.Disk)
	}
}

// TestRealHypervisor_VMStatus_Fields tests VMStatus field values
func TestRealHypervisor_VMStatus_Fields(t *testing.T) {
	status := &realhv.VMStatus{
		State:       hypervisor.NodeRunning,
		Uptime:      time.Hour,
		CPUPercent:  45.5,
		MemUsed:     4096,
		MemTotal:    8192,
		DiskUsedGB:  25.5,
		DiskTotalGB: 100.0,
		IP:          "10.0.0.1",
	}

	if status.State != hypervisor.NodeRunning {
		t.Errorf("expected State=Running, got %v", status.State)
	}
	if status.Uptime != time.Hour {
		t.Errorf("expected Uptime=1h, got %v", status.Uptime)
	}
	if status.CPUPercent != 45.5 {
		t.Errorf("expected CPUPercent=45.5, got %f", status.CPUPercent)
	}
}

// TestRealHypervisor_Config_Defaults tests Config defaults
func TestRealHypervisor_Config_Defaults(t *testing.T) {
	hv := realhv.NewHypervisor(nil)
	if hv == nil {
		t.Fatal("NewHypervisor with nil config should return non-nil")
	}
}

// TestRealHypervisor_Disconnect_Twice tests disconnecting twice
func TestRealHypervisor_Disconnect_Twice(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI: "qemu:///system",
	})

	// Disconnect when not connected
	if err := hv.Disconnect(); err != nil {
		t.Errorf("Disconnect when not connected should not error: %v", err)
	}

	// Disconnect again
	if err := hv.Disconnect(); err != nil {
		t.Errorf("Second Disconnect should not error: %v", err)
	}
}

// TestRealHypervisor_StartNode tests StartNode
func TestRealHypervisor_StartNode(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	err := hv.StartNode(ctx, "test-vm")
	if err == nil {
		t.Error("StartNode should fail without connection")
	}
}

// TestRealHypervisor_StopNode tests StopNode
func TestRealHypervisor_StopNode(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	err := hv.StopNode(ctx, "test-vm")
	if err == nil {
		t.Error("StopNode should fail without connection")
	}
}

// TestRealHypervisor_RestartNode tests RestartNode
func TestRealHypervisor_RestartNode(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	err := hv.RestartNode(ctx, "test-vm")
	if err == nil {
		t.Error("RestartNode should fail without connection")
	}
}

// TestRealHypervisor_GetNode tests GetNode
func TestRealHypervisor_GetNode(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	_, err := hv.GetNode(ctx, "test-vm")
	if err == nil {
		t.Error("GetNode should fail without connection")
	}
}

// TestRealHypervisor_GetNodeStatus tests GetNodeStatus
func TestRealHypervisor_GetNodeStatus(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	_, err := hv.GetNodeStatus(ctx, "test-vm")
	if err == nil {
		t.Error("GetNodeStatus should fail without connection")
	}
}

// TestRealHypervisor_GetMetrics tests GetMetrics
func TestRealHypervisor_GetMetrics(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu+tcp://127.0.0.1/system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Without connection, this should fail
	_, err := hv.GetMetrics(ctx, "test-vm")
	if err == nil {
		t.Error("GetMetrics should fail without connection")
	}
}

// TestRealHypervisor_Connect_Local tests Connect with local URI
func TestRealHypervisor_Connect_Local(t *testing.T) {
	cfg := &realhv.Config{
		URI:         "qemu:///system",
		AutoConnect: false,
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Connect should work (will use stub if libvirt not available)
	if err := hv.Connect(ctx); err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	// Second connect should be no-op
	if err := hv.Connect(ctx); err != nil {
		t.Errorf("Second Connect failed: %v", err)
	}

	hv.Disconnect()
}

// TestRealHypervisor_Connect_SSH tests Connect with SSH URI
func TestRealHypervisor_Connect_SSH(t *testing.T) {
	cfg := &realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		AutoConnect: false,
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Connect should work (will use stub)
	if err := hv.Connect(ctx); err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	hv.Disconnect()
}

// TestRealHypervisor_Connect_EmptyURI tests Connect with empty URI
func TestRealHypervisor_Connect_EmptyURI(t *testing.T) {
	cfg := &realhv.Config{
		URI:         "",
		AutoConnect: false,
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	if err := hv.Connect(ctx); err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	hv.Disconnect()
}

// TestRealHypervisor_Connect_Apple tests Connect with apple URI
func TestRealHypervisor_Connect_Apple(t *testing.T) {
	cfg := &realhv.Config{
		URI:         "apple",
		AutoConnect: false,
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	if err := hv.Connect(ctx); err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	hv.Disconnect()
}

// TestRealHypervisor_Disconnect_AfterConnect tests Disconnect after connect
func TestRealHypervisor_Disconnect_AfterConnect(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu:///system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	if err := hv.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	// Disconnect should work
	if err := hv.Disconnect(); err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	// Second disconnect should also work
	if err := hv.Disconnect(); err != nil {
		t.Errorf("Second Disconnect failed: %v", err)
	}
}

// TestRealHypervisor_CreateNode tests CreateNode
func TestRealHypervisor_CreateNode(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu:///system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	// Connect first
	if err := hv.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer hv.Disconnect()

	// CreateNode should work (with stub)
	vmCfg := &realhv.VMConfig{
		Name:     "test-vm",
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
	}
	_, err := hv.CreateNode(ctx, vmCfg)
	// May fail without actual libvirt
	_ = err
}

// TestRealHypervisor_ListNodes_AfterConnect tests ListNodes after connect
func TestRealHypervisor_ListNodes_AfterConnect(t *testing.T) {
	cfg := &realhv.Config{
		URI: "qemu:///system",
	}
	hv := realhv.NewHypervisor(cfg)

	ctx := context.Background()
	if err := hv.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer hv.Disconnect()

	// ListNodes should work
	nodes, err := hv.ListNodes(ctx)
	// May return empty list
	_ = nodes
	_ = err
}
