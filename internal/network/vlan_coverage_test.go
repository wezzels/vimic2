// Package network provides VLAN tests
package network

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewVLANAllocator tests VLAN allocator creation
func TestNewVLANAllocator(t *testing.T) {
	va, err := NewVLANAllocator(100, 200)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	if va.start != 100 {
		t.Errorf("expected start 100, got %d", va.start)
	}
	if va.end != 200 {
		t.Errorf("expected end 200, got %d", va.end)
	}
}

// TestNewVLANAllocator_InvalidRange tests invalid VLAN ranges
func TestNewVLANAllocator_InvalidRange(t *testing.T) {
	tests := []struct {
		name  string
		start int
		end   int
	}{
		{"Start too low", 0, 100},
		{"Start too high", 5000, 6000},
		{"End before start", 200, 100},
		{"End too high", 100, 5000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewVLANAllocator(tt.start, tt.end)
			if err == nil {
				t.Errorf("expected error for start=%d, end=%d", tt.start, tt.end)
			}
		})
	}
}

// TestVLANAllocator_AllocateReal tests VLAN allocation
func TestVLANAllocator_Allocate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	// Allocate first VLAN
	vlan1, err := va.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if vlan1 < 100 || vlan1 > 110 {
		t.Errorf("allocated VLAN %d out of range", vlan1)
	}

	// Allocate second VLAN
	vlan2, err := va.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if vlan1 == vlan2 {
		t.Error("allocated same VLAN twice")
	}
}

// TestVLANAllocator_ReleaseReal tests VLAN release
func TestVLANAllocator_Release(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-release-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	// Allocate and release
	vlan, err := va.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	err = va.Release(vlan)
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	// Should be able to allocate again
	vlan2, err := va.Allocate()
	if err != nil {
		t.Fatalf("Allocate after release failed: %v", err)
	}

	if vlan2 != vlan {
		t.Errorf("expected same VLAN after release, got %d vs %d", vlan, vlan2)
	}
}

// TestVLANAllocator_ReleaseNotAllocated tests releasing unallocated VLAN
func TestVLANAllocator_ReleaseNotAllocated(t *testing.T) {
	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	err = va.Release(105)
	if err == nil {
		t.Error("expected error for releasing unallocated VLAN")
	}
}

// TestVLANAllocator_IsAllocatedCheck tests IsAllocated check
func TestVLANAllocator_IsAllocated(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-is-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	vlan, _ := va.Allocate()

	if !va.IsAllocated(vlan) {
		t.Errorf("VLAN %d should be allocated", vlan)
	}

	if va.IsAllocated(999) {
		t.Error("VLAN 999 should not be allocated")
	}
}

// TestVLANAllocator_UsedCount tests Used count
func TestVLANAllocator_Used(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-used-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	if va.Used() != 0 {
		t.Errorf("expected 0 used, got %d", va.Used())
	}

	va.Allocate()
	va.Allocate()

	if va.Used() != 2 {
		t.Errorf("expected 2 used, got %d", va.Used())
	}
}

// TestVLANAllocator_AvailableCount tests Available count
func TestVLANAllocator_Available(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-avail-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	total := 11 // 100-110 inclusive
	if va.Available() != total {
		t.Errorf("expected %d available, got %d", total, va.Available())
	}

	va.Allocate()
	va.Allocate()

	if va.Available() != total-2 {
		t.Errorf("expected %d available after 2 allocations, got %d", total-2, va.Available())
	}
}

// TestVLANAllocator_ListUsed tests ListUsed
func TestVLANAllocator_ListUsed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-list-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	va.Allocate()
	va.Allocate()

	used := va.ListUsed()
	if len(used) != 2 {
		t.Errorf("expected 2 used VLANs, got %d", len(used))
	}
}

// TestVLANAllocator_ListAvailable tests ListAvailable
func TestVLANAllocator_ListAvailable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-list-avail-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	available := va.ListAvailable()
	if len(available) != 11 {
		t.Errorf("expected 11 available VLANs, got %d", len(available))
	}
}

// TestVLANAllocator_Reclaim tests Reclaim
func TestVLANAllocator_Reclaim(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-reclaim-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	// Reclaim a VLAN
	va.Reclaim(105)

	if !va.IsAllocated(105) {
		t.Error("VLAN 105 should be allocated after reclaim")
	}

	// Should skip reclaimed VLAN
	vlan, err := va.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if vlan == 105 {
		t.Error("should not allocate reclaimed VLAN")
	}
}

// TestVLANAllocator_ResetPool tests Reset
func TestVLANAllocator_Reset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-reset-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	va, err := NewVLANAllocator(100, 110)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	// Allocate some VLANs
	va.Allocate()
	va.Allocate()
	va.Allocate()

	if va.Used() != 3 {
		t.Errorf("expected 3 used, got %d", va.Used())
	}

	// Reset
	err = va.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	if va.Used() != 0 {
		t.Errorf("expected 0 used after reset, got %d", va.Used())
	}
}

// TestVLANAllocator_Exhaust tests exhausting VLAN pool
func TestVLANAllocator_Exhaust(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vlan-exhaust-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := filepath.Join(tmpDir, "vlan-state.json")

	// Small range
	va, err := NewVLANAllocator(100, 102)
	if err != nil {
		t.Fatalf("NewVLANAllocator failed: %v", err)
	}

	va.SetStateFile(stateFile)

	// Allocate all 3 VLANs
	va.Allocate()
	va.Allocate()
	va.Allocate()

	// Should fail
	_, err = va.Allocate()
	if err == nil {
		t.Error("expected error when pool exhausted")
	}
}