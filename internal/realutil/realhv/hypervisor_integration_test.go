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

// TestRealHypervisor_Connect_10_0_0_99 tests connection to libvirt on 10.0.0.99
func TestRealHypervisor_Connect_10_0_0_99(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.99/system",
		Timeout:     10 * time.Second,
		AutoConnect: false,
	})

	ctx := context.Background()

	err := hv.Connect(ctx)
	if err != nil {
		t.Fatalf("libvirt connection failed: %v", err)
	}
	defer hv.Disconnect()

	if !hv.IsConnected() {
		t.Error("should be connected")
	}
}

// TestRealHypervisor_Connect_10_0_0_117 tests connection to libvirt on 10.0.0.117
func TestRealHypervisor_Connect_10_0_0_117(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: false,
	})

	ctx := context.Background()

	err := hv.Connect(ctx)
	if err != nil {
		t.Fatalf("libvirt connection failed: %v", err)
	}
	defer hv.Disconnect()

	if !hv.IsConnected() {
		t.Error("should be connected")
	}
}

// TestRealHypervisor_ListNodes_10_0_0_99 tests listing VMs on 10.0.0.99
func TestRealHypervisor_ListNodes_10_0_0_99(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.99/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	t.Logf("Found %d VMs on 10.0.0.99", len(nodes))
}

// TestRealHypervisor_ListNodes_10_0_0_117 tests listing VMs on 10.0.0.117
func TestRealHypervisor_ListNodes_10_0_0_117(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	// Stub hypervisor returns empty list without libvirt
	t.Logf("Found %d VMs on 10.0.0.117 (stub mode)", len(nodes))
}

// TestRealHypervisor_GetNode_10_0_0_117 tests getting VM details
func TestRealHypervisor_GetNode_10_0_0_117(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	// List nodes first
	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}

	if len(nodes) == 0 {
		t.Skip("no VMs found (stub mode)")
	}

	// Get first VM
	node, err := hv.GetNode(ctx, nodes[0].ID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}

	if node.Name != nodes[0].Name {
		t.Errorf("expected name %s, got %s", nodes[0].Name, node.Name)
	}
}

// TestRealHypervisor_Factory_Remote tests factory for remote hosts
func TestRealHypervisor_Factory_Remote(t *testing.T) {
	factory := realhv.NewHypervisorFactory()

	// Test creating remote hypervisor for 10.0.0.99
	hv99 := factory.CreateRemote("10.0.0.99")
	if hv99 == nil {
		t.Error("factory should create hypervisor for 10.0.0.99")
	}

	// Test creating remote hypervisor for 10.0.0.117
	hv117 := factory.CreateRemote("10.0.0.117")
	if hv117 == nil {
		t.Error("factory should create hypervisor for 10.0.0.117")
	}
}
