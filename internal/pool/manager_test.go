// Package pool provides pool manager tests
package pool

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPoolConfig tests pool configuration
func TestPoolConfig_Create(t *testing.T) {
	config := &poolConfig{
		Template: "template-1",
		MinSize:  2,
		MaxSize:  10,
		CPU:      4,
		Memory:   8192,
	}

	if config.Template != "template-1" {
		t.Errorf("expected template-1, got %s", config.Template)
	}
	if config.MinSize != 2 {
		t.Errorf("expected min size 2, got %d", config.MinSize)
	}
	if config.MaxSize != 10 {
		t.Errorf("expected max size 10, got %d", config.MaxSize)
	}
}

// TestPool tests pool structure
func TestPool_Create(t *testing.T) {
	pool := &Pool{
		ID:          "pool-1",
		Name:        "test-pool",
		TemplateID:  "template-1",
		MinSize:     2,
		MaxSize:     10,
		CurrentSize: 5,
		CPU:         4,
		Memory:      8192,
		VMs:         []string{"vm-1", "vm-2", "vm-3", "vm-4", "vm-5"},
		CreatedAt:   time.Now(),
	}

	if pool.ID != "pool-1" {
		t.Errorf("expected ID pool-1, got %s", pool.ID)
	}
	if pool.MinSize != 2 {
		t.Errorf("expected min size 2, got %d", pool.MinSize)
	}
	if len(pool.VMs) != 5 {
		t.Errorf("expected 5 VMs, got %d", len(pool.VMs))
	}
}

// TestVM tests VM structure
func TestVM_Create(t *testing.T) {
	vm := &VM{
		ID:        "vm-1",
		Name:      "test-vm",
		PoolID:    "pool-1",
		Status:    "running",
		IPAddress: "10.100.1.10",
		CreatedAt: time.Now(),
	}

	if vm.ID != "vm-1" {
		t.Errorf("expected ID vm-1, got %s", vm.ID)
	}
	if vm.Status != "running" {
		t.Errorf("expected status running, got %s", vm.Status)
	}
}

// TestVMStatus tests VM status values
func TestVMStatus_Valid(t *testing.T) {
	statuses := []VMStatus{
		VMStatusCreating,
		VMStatusRunning,
		VMStatusIdle,
		VMStatusBusy,
		VMStatusStopping,
		VMStatusStopped,
		VMStatusDestroyed,
		VMStatusError,
	}

	for _, status := range statuses {
		if status == "" {
			t.Error("empty status in valid list")
		}
	}
}

// TestOverlay tests overlay structure
func TestOverlay_Create(t *testing.T) {
	overlay := &Overlay{
		ID:         "overlay-1",
		TemplateID: "template-1",
		VMID:       "vm-1",
		Path:       "/var/lib/vimic2/overlays/vm-1.qcow2",
		CreatedAt:  time.Now(),
	}

	if overlay.ID != "overlay-1" {
		t.Errorf("expected overlay-1, got %s", overlay.ID)
	}
	if overlay.VMID != "vm-1" {
		t.Errorf("expected vm-1, got %s", overlay.VMID)
	}
}

// TestVMEvent tests VM event
func TestVMEvent_Create(t *testing.T) {
	vm := &VM{
		ID:     "vm-1",
		Name:   "test-vm",
		PoolID: "pool-1",
		Status: "running",
	}

	event := VMEvent{
		VM:        vm,
		OldStatus: VMStatusCreating,
		NewStatus: VMStatusRunning,
		Timestamp: time.Now(),
		PoolID:    "pool-1",
	}

	if event.VM.ID != "vm-1" {
		t.Errorf("expected vm-1, got %s", event.VM.ID)
	}
	if event.OldStatus != VMStatusCreating {
		t.Errorf("expected creating, got %s", event.OldStatus)
	}
	if event.NewStatus != VMStatusRunning {
		t.Errorf("expected running, got %s", event.NewStatus)
	}
}

// TestPoolJSON tests JSON marshaling
func TestPool_JSON(t *testing.T) {
	pool := &Pool{
		ID:          "pool-1",
		Name:        "test-pool",
		TemplateID:  "template-1",
		MinSize:     2,
		MaxSize:     10,
		CurrentSize: 5,
		CPU:         4,
		Memory:      8192,
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(pool)
	if err != nil {
		t.Fatalf("failed to marshal pool: %v", err)
	}

	var pool2 Pool
	if err := json.Unmarshal(data, &pool2); err != nil {
		t.Fatalf("failed to unmarshal pool: %v", err)
	}

	if pool2.ID != pool.ID {
		t.Errorf("expected ID %s, got %s", pool.ID, pool2.ID)
	}
	if pool2.Name != pool.Name {
		t.Errorf("expected name %s, got %s", pool.Name, pool2.Name)
	}
}

// TestVMJSON tests VM JSON marshaling
func TestVM_JSON(t *testing.T) {
	vm := &VM{
		ID:        "vm-1",
		Name:      "test-vm",
		PoolID:    "pool-1",
		Status:    "running",
		IPAddress: "10.100.1.10",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(vm)
	if err != nil {
		t.Fatalf("failed to marshal VM: %v", err)
	}

	var vm2 VM
	if err := json.Unmarshal(data, &vm2); err != nil {
		t.Fatalf("failed to unmarshal VM: %v", err)
	}

	if vm2.ID != vm.ID {
		t.Errorf("expected ID %s, got %s", vm.ID, vm2.ID)
	}
	if vm2.Status != vm.Status {
		t.Errorf("expected status %s, got %s", vm.Status, vm2.Status)
	}
}

// TestVMStatusTransitions tests VM status transitions
func TestVMStatus_Transitions(t *testing.T) {
	// Valid transitions
	transitions := []struct {
		from VMStatus
		to   VMStatus
	}{
		{VMStatusCreating, VMStatusRunning},
		{VMStatusRunning, VMStatusIdle},
		{VMStatusIdle, VMStatusBusy},
		{VMStatusBusy, VMStatusIdle},
		{VMStatusRunning, VMStatusStopping},
		{VMStatusStopping, VMStatusStopped},
		{VMStatusStopped, VMStatusDestroyed},
	}

	for _, tt := range transitions {
		if tt.from == "" || tt.to == "" {
			t.Errorf("empty status in transition from %s to %s", tt.from, tt.to)
		}
	}
}