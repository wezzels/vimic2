//go:build integration

// Package network provides integration tests for VLAN allocator
package network

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// newIsolatedVLANAllocator creates a VLAN allocator with isolated state for integration tests
func newIsolatedVLANAllocator(t *testing.T, start, end int) *VLANAllocator {
	tmpDir, err := os.MkdirTemp("", "vlan-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	stateFile := filepath.Join(tmpDir, "vlan-state.json")
	va := &VLANAllocator{
		start:     start,
		end:       end,
		used:      make(map[int]bool),
		stateFile: stateFile,
	}
	return va
}

// TestIntegration_VLAN_AllocateRelease tests allocating and releasing VLANs
func TestIntegration_VLAN_AllocateRelease(t *testing.T) {
	allocator := newIsolatedVLANAllocator(t, 100, 200)

	// Allocate VLAN
	vlan1, err := allocator.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if vlan1 < 100 || vlan1 > 200 {
		t.Errorf("allocated VLAN %d out of range [100, 200]", vlan1)
	}

	// Allocate another
	vlan2, err := allocator.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if vlan2 == vlan1 {
		t.Errorf("allocated same VLAN twice")
	}

	// Release first VLAN
	err = allocator.Release(vlan1)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Allocate again - should get the released one
	vlan3, err := allocator.Allocate()
	if err != nil {
		t.Fatalf("Allocate after release failed: %v", err)
	}

	t.Logf("Allocated VLANs: %d, %d, %d", vlan1, vlan2, vlan3)
}

// TestIntegration_VLAN_ConcurrentAllocation tests concurrent VLAN allocation
func TestIntegration_VLAN_ConcurrentAllocation(t *testing.T) {
	allocator := newIsolatedVLANAllocator(t, 1000, 1100) // 101 VLANs

	var wg sync.WaitGroup
	results := make(chan int, 50)
	errors := make(chan error, 50)

	// Allocate 50 VLANs concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vlan, err := allocator.Allocate()
			if err != nil {
				errors <- err
				return
			}
			results <- vlan
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent allocation error: %v", err)
	}

	// Collect unique VLANs
	uniqueVLANs := make(map[int]bool)
	for vlan := range results {
		uniqueVLANs[vlan] = true
	}

	if len(uniqueVLANs) != 50 {
		t.Errorf("expected 50 unique VLANs, got %d", len(uniqueVLANs))
	}

	t.Logf("Successfully allocated %d unique VLANs concurrently", len(uniqueVLANs))
}

// TestIntegration_VLAN_Exhaustion tests VLAN pool exhaustion
func TestIntegration_VLAN_Exhaustion(t *testing.T) {
	allocator := newIsolatedVLANAllocator(t, 2000, 2010) // 11 VLANs

	// Allocate all VLANs
	for i := 0; i < 11; i++ {
		_, err := allocator.Allocate()
		if err != nil {
			t.Fatalf("allocation %d failed: %v", i, err)
		}
	}

	// Should be exhausted - one more should fail
	_, err := allocator.Allocate()
	if err == nil {
		t.Error("expected error when pool exhausted")
	} else {
		t.Logf("correctly got error when exhausted: %v", err)
	}

	// Release one VLAN
	err = allocator.Release(2000)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Should be able to allocate again
	vlan, err := allocator.Allocate()
	if err != nil {
		t.Fatalf("Allocate after release failed: %v", err)
	}
	t.Logf("Allocated VLAN %d after release", vlan)
}

// TestIntegration_VLAN_Status tests status methods
func TestIntegration_VLAN_Status(t *testing.T) {
	allocator := newIsolatedVLANAllocator(t, 3000, 3020)

	// Check initial status
	if allocator.Used() != 0 {
		t.Errorf("expected 0 used, got %d", allocator.Used())
	}

	if allocator.Available() != 21 {
		t.Errorf("expected 21 available, got %d", allocator.Available())
	}

	// Allocate some VLANs
	vlan1, _ := allocator.Allocate()
	vlan2, _ := allocator.Allocate()
	vlan3, _ := allocator.Allocate()

	if allocator.Used() != 3 {
		t.Errorf("expected 3 used, got %d", allocator.Used())
	}

	if allocator.Available() != 18 {
		t.Errorf("expected 18 available, got %d", allocator.Available())
	}

	// Check IsAllocated
	if !allocator.IsAllocated(vlan1) {
		t.Errorf("VLAN %d should be allocated", vlan1)
	}
	if !allocator.IsAllocated(vlan2) {
		t.Errorf("VLAN %d should be allocated", vlan2)
	}
	if !allocator.IsAllocated(vlan3) {
		t.Errorf("VLAN %d should be allocated", vlan3)
	}
	if allocator.IsAllocated(3005) {
		t.Error("VLAN 3005 should not be allocated")
	}

	// Check ListUsed
	usedList := allocator.ListUsed()
	if len(usedList) != 3 {
		t.Errorf("expected 3 used VLANs, got %d", len(usedList))
	}

	// Check ListAvailable
	availableList := allocator.ListAvailable()
	if len(availableList) != 18 {
		t.Errorf("expected 18 available VLANs, got %d", len(availableList))
	}
}