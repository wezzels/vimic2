//go:build integration
// +build integration

// Package realhv_test tests real hypervisor with libvirt connection
package realhv_test

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realhv"
)

// TestRealHypervisor_Connect_Integration tests connection to libvirt
func TestRealHypervisor_Connect_Integration(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     10 * time.Second,
		AutoConnect: false,
	})

	ctx := context.Background()

	// Try to connect
	err := hv.Connect(ctx)
	if err != nil {
		// Connection failed - libvirt might not be running
		t.Skipf("libvirt connection failed (expected in CI): %v", err)
	}
	defer hv.Disconnect()

	if !hv.IsConnected() {
		t.Error("should be connected")
	}
}

// TestRealHypervisor_CreateNode_Integration tests VM creation
func TestRealHypervisor_CreateNode_Integration(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     30 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	// Create a test VM
	cfg := &realhv.VMConfig{
		Name:     "test-vm-integration",
		CPU:      1,
		MemoryMB: 512,
		DiskGB:   5,
		Image:    "cirros",
	}

	vm, err := hv.CreateNode(ctx, cfg)
	if err != nil {
		t.Skipf("VM creation failed (expected in CI without libvirt): %v", err)
	}
	defer hv.DeleteNode(ctx, vm.ID)

	if vm.ID == "" {
		t.Error("VM ID should not be empty")
	}
	if vm.Name != cfg.Name {
		t.Errorf("expected name %s, got %s", cfg.Name, vm.Name)
	}
}

// TestRealHypervisor_ListNodes_Integration tests listing VMs
func TestRealHypervisor_ListNodes_Integration(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Skipf("ListNodes failed (expected in CI without libvirt): %v", err)
	}

	// Just verify we got a list
	t.Logf("Found %d VMs", len(nodes))
}

// TestRealHypervisor_GetNodeStatus_Integration tests getting VM status
func TestRealHypervisor_GetNodeStatus_Integration(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	// First create a VM
	cfg := &realhv.VMConfig{
		Name:     "test-vm-status",
		CPU:      1,
		MemoryMB: 512,
		DiskGB:   5,
		Image:    "cirros",
	}

	vm, err := hv.CreateNode(ctx, cfg)
	if err != nil {
		t.Skipf("VM creation failed (expected in CI without libvirt): %v", err)
	}
	defer hv.DeleteNode(ctx, vm.ID)

	// Get status
	status, err := hv.GetNodeStatus(ctx, vm.ID)
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}

	if status.State == "" {
		t.Error("status state should not be empty")
	}
}

// TestRealHypervisor_GetMetrics_Integration tests getting VM metrics
func TestRealHypervisor_GetMetrics_Integration(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu:///system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	// First create a VM
	cfg := &realhv.VMConfig{
		Name:     "test-vm-metrics",
		CPU:      1,
		MemoryMB: 512,
		DiskGB:   5,
		Image:    "cirros",
	}

	vm, err := hv.CreateNode(ctx, cfg)
	if err != nil {
		t.Skipf("VM creation failed (expected in CI without libvirt): %v", err)
	}
	defer hv.DeleteNode(ctx, vm.ID)

	// Get metrics
	metrics, err := hv.GetMetrics(ctx, vm.ID)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}

	if metrics.Timestamp.IsZero() {
		t.Error("metrics timestamp should not be zero")
	}
}