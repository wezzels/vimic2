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

// TestRealHypervisor_OperationsWithoutConnect tests operations without connection
func TestRealHypervisor_OperationsWithoutConnect(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		AutoConnect: false,
	})

	ctx := context.Background()

	// These should fail because not connected
	_, err := hv.CreateNode(ctx, &realhv.VMConfig{Name: "test"})
	if err == nil {
		t.Error("CreateNode should fail without connection")
	}

	err = hv.StartNode(ctx, "vm-1")
	if err == nil {
		t.Error("StartNode should fail without connection")
	}

	err = hv.StopNode(ctx, "vm-1")
	if err == nil {
		t.Error("StopNode should fail without connection")
	}

	_, err = hv.ListNodes(ctx)
	if err == nil {
		t.Error("ListNodes should fail without connection")
	}

	_, err = hv.GetNode(ctx, "vm-1")
	if err == nil {
		t.Error("GetNode should fail without connection")
	}
}