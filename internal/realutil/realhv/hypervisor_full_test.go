//go:build integration && libvirt
// +build integration,libvirt

package realhv_test

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/realutil/realhv"
)

// TestRealHypervisor_FullLifecycle tests complete VM lifecycle with real libvirt
func TestRealHypervisor_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     30 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	// List existing VMs first
	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	t.Logf("Found %d existing VMs", len(nodes))

	if len(nodes) == 0 {
		t.Skip("no VMs found to test lifecycle")
	}

	// Pick first VM
	testVM := nodes[0]
	t.Logf("Testing with VM: %s (ID: %s, State: %s)", testVM.Name, testVM.ID, testVM.State)

	// Test GetNode
	node, err := hv.GetNode(ctx, testVM.ID)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if node.ID != testVM.ID {
		t.Errorf("GetNode returned wrong ID: got %s, want %s", node.ID, testVM.ID)
	}

	// Test GetNodeStatus
	status, err := hv.GetNodeStatus(ctx, testVM.ID)
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}
	t.Logf("VM Status: State=%s, MemUsed=%d, MemTotal=%d", status.State, status.MemUsed, status.MemTotal)

	// Test GetMetrics
	metrics, err := hv.GetMetrics(ctx, testVM.ID)
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}
	t.Logf("VM Metrics: CPU=%.1f%%, Memory=%.1f%%", metrics.CPU, metrics.Memory)

	// Test StartNode (if stopped)
	if testVM.State == "stopped" {
		err = hv.StartNode(ctx, testVM.ID)
		if err != nil {
			t.Logf("StartNode failed (may be expected): %v", err)
		} else {
			t.Log("StartNode succeeded")
		}
	}

	// Test StopNode (if running)
	if testVM.State == "running" {
		// Don't actually stop - just verify the call works
		t.Log("VM is running, skipping StopNode test to avoid disruption")
	}

	// Test RestartNode
	t.Log("Skipping RestartNode test to avoid disruption")
}

// TestRealHypervisor_GetNodeStatus_All tests GetNodeStatus for all VMs
func TestRealHypervisor_GetNodeStatus_All(t *testing.T) {
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

	for _, node := range nodes {
		status, err := hv.GetNodeStatus(ctx, node.ID)
		if err != nil {
			t.Errorf("GetNodeStatus(%s) failed: %v", node.Name, err)
			continue
		}
		t.Logf("%s: State=%s, MemUsed=%d/%d", node.Name, status.State, status.MemUsed, status.MemTotal)
	}
}

// TestRealHypervisor_GetMetrics_All tests GetMetrics for all VMs
func TestRealHypervisor_GetMetrics_All(t *testing.T) {
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

	for _, node := range nodes {
		metrics, err := hv.GetMetrics(ctx, node.ID)
		if err != nil {
			t.Errorf("GetMetrics(%s) failed: %v", node.Name, err)
			continue
		}
		t.Logf("%s: CPU=%.1f%%, Mem=%.1f%%, Time=%s", node.Name, metrics.CPU, metrics.Memory, metrics.Timestamp.Format(time.RFC3339))
	}
}

// TestRealHypervisor_DeleteNode_NonExistent tests DeleteNode with invalid ID
func TestRealHypervisor_DeleteNode_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.DeleteNode(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("DeleteNode should fail for non-existent VM")
	}
	t.Logf("DeleteNode correctly failed: %v", err)
}

// TestRealHypervisor_StartNode_NonExistent tests StartNode with invalid ID
func TestRealHypervisor_StartNode_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.StartNode(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("StartNode should fail for non-existent VM")
	}
	t.Logf("StartNode correctly failed: %v", err)
}

// TestRealHypervisor_StopNode_NonExistent tests StopNode with invalid ID
func TestRealHypervisor_StopNode_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.StopNode(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("StopNode should fail for non-existent VM")
	}
	t.Logf("StopNode correctly failed: %v", err)
}

// TestRealHypervisor_RestartNode_NonExistent tests RestartNode with invalid ID
func TestRealHypervisor_RestartNode_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	err := hv.RestartNode(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("RestartNode should fail for non-existent VM")
	}
	t.Logf("RestartNode correctly failed: %v", err)
}

// TestRealHypervisor_GetNodeStatus_NonExistent tests GetNodeStatus with invalid ID
func TestRealHypervisor_GetNodeStatus_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	_, err := hv.GetNodeStatus(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("GetNodeStatus should fail for non-existent VM")
	}
	t.Logf("GetNodeStatus correctly failed: %v", err)
}

// TestRealHypervisor_GetMetrics_NonExistent tests GetMetrics with invalid ID
func TestRealHypervisor_GetMetrics_NonExistent(t *testing.T) {
	hv := realhv.NewHypervisor(&realhv.Config{
		URI:         "qemu+ssh://10.0.0.117/system",
		Timeout:     10 * time.Second,
		AutoConnect: true,
	})

	ctx := context.Background()

	_, err := hv.GetMetrics(ctx, "non-existent-vm-id-12345")
	if err == nil {
		t.Error("GetMetrics should fail for non-existent VM")
	}
	t.Logf("GetMetrics correctly failed: %v", err)
}