//go:build integration
// +build integration

// Package pool provides integration tests with real database for state tracking
package pool

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestStateTracker_GetVM_RealDB tests VM state retrieval with real database
func TestStateTracker_GetVM_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Set a VM state
	vm := &VMState{
		ID:        "vm-1",
		Name:      "test-vm-1",
		Status:    "running",
		IPAddress: "10.0.0.1",
		PoolName:  "test-pool",
		Template:  "ubuntu-22.04",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tracker.SetVM(vm)

	// Get the VM state
	retrieved, err := tracker.GetVM("vm-1")
	if err != nil {
		t.Fatalf("failed to get VM: %v", err)
	}

	if retrieved.ID != vm.ID {
		t.Errorf("expected VM ID %s, got %s", vm.ID, retrieved.ID)
	}
	if retrieved.Name != vm.Name {
		t.Errorf("expected VM name %s, got %s", vm.Name, retrieved.Name)
	}
	if retrieved.Status != vm.Status {
		t.Errorf("expected status %s, got %s", vm.Status, retrieved.Status)
	}
}

// TestStateTracker_SetVM_RealDB tests VM state setting with real database
func TestStateTracker_SetVM_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Set multiple VM states
	for i := 0; i < 3; i++ {
		vm := &VMState{
			ID:        "vm-" + string(rune('a'+i)),
			Name:      "test-vm-" + string(rune('a'+i)),
			Status:    "running",
			PoolName:  "test-pool",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tracker.SetVM(vm)
	}

	// List all VMs
	vms := tracker.ListVMs()
	if len(vms) < 3 {
		t.Errorf("expected at least 3 VMs, got %d", len(vms))
	}
}

// TestStateTracker_DeleteVM_RealDB tests VM state deletion with real database
func TestStateTracker_DeleteVM_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Set a VM state
	vm := &VMState{
		ID:        "vm-1",
		Name:      "test-vm-1",
		Status:    "running",
		PoolName:  "test-pool",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tracker.SetVM(vm)

	// Delete the VM
	tracker.DeleteVM("vm-1")

	// Verify it's deleted
	_, err = tracker.GetVM("vm-1")
	if err == nil {
		t.Error("expected error when getting deleted VM")
	}
}

// TestStateTracker_ListVMs_RealDB tests listing VM states with real database
func TestStateTracker_ListVMs_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Set multiple VMs with different statuses
	vmIDs := []string{"vm-1", "vm-2", "vm-3"}
	for _, id := range vmIDs {
		vm := &VMState{
			ID:        id,
			Name:      "test-" + id,
			Status:    "running",
			PoolName:  "test-pool",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tracker.SetVM(vm)
	}

	// List all VMs
	vms := tracker.ListVMs()
	if len(vms) < 3 {
		t.Errorf("expected at least 3 VMs, got %d", len(vms))
	}

	// Verify all VMs are in the list
	vmMap := make(map[string]bool)
	for _, vm := range vms {
		vmMap[vm.ID] = true
	}
	for _, id := range vmIDs {
		if !vmMap[id] {
			t.Errorf("expected VM %s in list", id)
		}
	}
}

// TestStateTracker_Subscribe_RealDB tests state event subscription
func TestStateTracker_Subscribe_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Subscribe to events
	ch := tracker.Subscribe("test-subscriber")
	if ch == nil {
		t.Error("expected non-nil channel")
	}

	// Unsubscribe
	tracker.Unsubscribe("test-subscriber", ch)
}

// TestStateTracker_Persistence tests that state persists across restarts
func TestStateTracker_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	stateFile := filepath.Join(tmpDir, "state.json")

	// First session: create state
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}

		tracker, err := NewStateTracker(db, stateFile)
		if err != nil {
			t.Fatalf("failed to create state tracker: %v", err)
		}

		vm := &VMState{
			ID:        "vm-persistent",
			Name:      "persistent-vm",
			Status:    "running",
			PoolName:  "test-pool",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tracker.SetVM(vm)
		tracker.SaveState()

		db.Close()
	}

	// Second session: verify state persisted
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer db.Close()

		tracker, err := NewStateTracker(db, stateFile)
		if err != nil {
			t.Fatalf("failed to create state tracker: %v", err)
		}

		vm, err := tracker.GetVM("vm-persistent")
		if err != nil {
			t.Fatalf("failed to get persistent VM: %v", err)
		}

		if vm.Name != "persistent-vm" {
			t.Errorf("expected VM name persistent-vm, got %s", vm.Name)
		}
	}
}

// TestVMState_Transitions tests VM state transitions
func TestVMState_Transitions_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-state-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	stateFile := filepath.Join(tmpDir, "state.json")
	tracker, err := NewStateTracker(db, stateFile)
	if err != nil {
		t.Fatalf("failed to create state tracker: %v", err)
	}

	// Create VM in creating state
	vm := &VMState{
		ID:        "vm-1",
		Name:      "test-vm",
		Status:    "creating",
		PoolName:  "test-pool",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tracker.SetVM(vm)

	// Transition to running
	vm.Status = "running"
	vm.UpdatedAt = time.Now()
	tracker.SetVM(vm)

	// Verify transition
	retrieved, err := tracker.GetVM("vm-1")
	if err != nil {
		t.Fatalf("failed to get VM: %v", err)
	}

	if retrieved.Status != "running" {
		t.Errorf("expected status running, got %s", retrieved.Status)
	}

	// Transition to stopped
	vm.Status = "stopped"
	vm.UpdatedAt = time.Now()
	tracker.SetVM(vm)

	retrieved, err = tracker.GetVM("vm-1")
	if err != nil {
		t.Fatalf("failed to get VM: %v", err)
	}

	if retrieved.Status != "stopped" {
		t.Errorf("expected status stopped, got %s", retrieved.Status)
	}
}