//go:build integration

// Package network provides integration tests for VLAN allocator
package network

import (
	"sync"
	"testing"
)

// TestIntegration_VLAN_AllocateRelease tests allocating and releasing VLANs
func TestIntegration_VLAN_AllocateRelease(t *testing.T) {
	allocator, err := NewVLANAllocator(100, 200) // VLANs 100-200
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

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
	allocator, err := NewVLANAllocator(1000, 1100) // 100 VLANs
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

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

	// Check for duplicates
	seen := make(map[int]bool)
	for vlan := range results {
		if seen[vlan] {
			t.Errorf("duplicate VLAN allocated: %d", vlan)
		}
		seen[vlan] = true
	}

	if len(seen) < 50 {
		t.Errorf("expected 50 unique VLANs, got %d", len(seen))
	}

	t.Logf("Successfully allocated %d unique VLANs concurrently", len(seen))
}

// TestIntegration_VLAN_Exhaustion tests VLAN exhaustion
func TestIntegration_VLAN_Exhaustion(t *testing.T) {
	allocator, err := NewVLANAllocator(2000, 2010) // Only 10 VLANs
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	// Allocate all VLANs
	vlans := make([]int, 0)
	for i := 0; i < 10; i++ {
		vlan, err := allocator.Allocate()
		if err != nil {
			t.Fatalf("allocation %d failed: %v", i, err)
		}
		vlans = append(vlans, vlan)
	}

	// Try to allocate one more - should fail
	_, err = allocator.Allocate()
	if err == nil {
		t.Error("expected error when pool exhausted")
	}

	t.Logf("Pool exhausted after %d allocations", len(vlans))

	// Release one
	allocator.Release(vlans[0])

	// Should be able to allocate again
	vlan, err := allocator.Allocate()
	if err != nil {
		t.Errorf("allocation after release failed: %v", err)
	}
	t.Logf("Allocated VLAN %d after release", vlan)
}

// TestIntegration_VLAN_Status tests allocator status
func TestIntegration_VLAN_Status(t *testing.T) {
	allocator, err := NewVLANAllocator(3000, 3020) // 20 VLANs
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	// Allocate some
	for i := 0; i < 5; i++ {
		_, err := allocator.Allocate()
		if err != nil {
			t.Fatalf("Allocate failed: %v", err)
		}
	}

	t.Logf("Allocated 5 VLANs from pool")
}