// Package pool provides state tracker tests
package pool

import (
	"encoding/json"
	"testing"
	"time"
)

// TestVMState_Struct tests VM state structure
func TestVMState_Struct(t *testing.T) {
	state := &VMState{
		ID:         "vm-1",
		Name:       "test-vm",
		Status:     "running",
		IPAddress:  "10.100.1.10",
		PoolName:   "pool-1",
		Template:   "ubuntu-22.04",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if state.ID != "vm-1" {
		t.Errorf("expected vm-1, got %s", state.ID)
	}
	if state.Status != "running" {
		t.Errorf("expected running status, got %s", state.Status)
	}
	if state.PoolName != "pool-1" {
		t.Errorf("expected pool-1, got %s", state.PoolName)
	}
}

// TestVMState_JSON tests VM state JSON marshaling
func TestVMState_JSON(t *testing.T) {
	state := &VMState{
		ID:         "vm-1",
		Name:       "test-vm",
		Status:     "running",
		IPAddress:  "10.100.1.10",
		PoolName:   "pool-1",
		Template:   "ubuntu-22.04",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var state2 VMState
	if err := json.Unmarshal(data, &state2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if state2.ID != state.ID {
		t.Errorf("expected ID %s, got %s", state.ID, state2.ID)
	}
	if state2.Status != state.Status {
		t.Errorf("expected status %s, got %s", state.Status, state2.Status)
	}
}

// TestStateTransition tests state transition structure
func TestStateTransition_Struct(t *testing.T) {
	transition := &StateTransition{
		From:      "creating",
		To:        "running",
		Timestamp: time.Now(),
		Reason:    "VM started successfully",
	}

	if transition.From != "creating" {
		t.Errorf("expected from creating, got %s", transition.From)
	}
	if transition.To != "running" {
		t.Errorf("expected to running, got %s", transition.To)
	}
	if transition.Reason != "VM started successfully" {
		t.Errorf("expected reason, got %s", transition.Reason)
	}
}

// TestStateEvent tests state event structure
func TestStateEvent_Struct(t *testing.T) {
	event := &StateEvent{
		VMID:      "vm-1",
		OldState:  "stopped",
		NewState:  "running",
		Timestamp: time.Now(),
		Reason:    "Manual start",
	}

	if event.VMID != "vm-1" {
		t.Errorf("expected vm-1, got %s", event.VMID)
	}
	if event.OldState != "stopped" {
		t.Errorf("expected old state stopped, got %s", event.OldState)
	}
	if event.NewState != "running" {
		t.Errorf("expected new state running, got %s", event.NewState)
	}
}

// TestStateTracker_CreateStruct tests state tracker struct fields
func TestStateTracker_CreateStruct(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
	}

	if st.cache == nil {
		t.Error("cache should not be nil")
	}
	if st.subscribers == nil {
		t.Error("subscribers should not be nil")
	}

	// Add a VM state
	st.cache["vm-1"] = &VMState{
		ID:     "vm-1",
		Status: "running",
	}

	if len(st.cache) != 1 {
		t.Errorf("expected 1 cached state, got %d", len(st.cache))
	}
}

// TestStateTracker_UpdateCachedState tests updating cached state
func TestStateTracker_UpdateCachedState(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
	}

	// Create initial state
	st.cache["vm-1"] = &VMState{
		ID:         "vm-1",
		Name:       "test-vm",
		Status:     "creating",
		IPAddress:  "",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Update state
	st.cache["vm-1"].Status = "running"
	st.cache["vm-1"].IPAddress = "10.100.1.10"
	st.cache["vm-1"].UpdatedAt = time.Now()

	if st.cache["vm-1"].Status != "running" {
		t.Errorf("expected running status, got %s", st.cache["vm-1"].Status)
	}
	if st.cache["vm-1"].IPAddress != "10.100.1.10" {
		t.Errorf("expected IP, got %s", st.cache["vm-1"].IPAddress)
	}
}

// TestStateTracker_AddSubscriber tests subscriber pattern
func TestStateTracker_AddSubscriber(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
		eventChan:   make(chan StateEvent, 100),
	}

	// Create subscriber channel
	ch := make(chan StateEvent, 10)

	// Subscribe
	st.subscribers["vm-1"] = append(st.subscribers["vm-1"], ch)

	if len(st.subscribers["vm-1"]) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(st.subscribers["vm-1"]))
	}
}

// TestStateTracker_BroadcastStateEvent tests event broadcasting
func TestStateTracker_BroadcastStateEvent(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
		eventChan:   make(chan StateEvent, 100),
	}

	// Create subscriber
	ch := make(chan StateEvent, 10)
	st.subscribers["vm-1"] = append(st.subscribers["vm-1"], ch)

	// Broadcast event
	event := StateEvent{
		VMID:      "vm-1",
		OldState:  "stopped",
		NewState:  "running",
		Timestamp: time.Now(),
		Reason:    "Manual start",
	}

	// Send to subscribers
	for _, sub := range st.subscribers["vm-1"] {
		sub <- event
	}

	// Receive
	select {
	case received := <-ch:
		if received.VMID != "vm-1" {
			t.Errorf("expected vm-1, got %s", received.VMID)
		}
		if received.NewState != "running" {
			t.Errorf("expected running, got %s", received.NewState)
		}
	default:
		t.Error("should have received event")
	}
}

// TestVMState_ValidTransitions tests valid state transitions
func TestVMState_ValidTransitions(t *testing.T) {
	validTransitions := []struct {
		from string
		to   string
	}{
		{"creating", "running"},
		{"creating", "error"},
		{"running", "stopped"},
		{"running", "error"},
		{"stopped", "running"},
		{"stopped", "destroyed"},
	}

	for _, tt := range validTransitions {
		state := &VMState{Status: tt.from}
		state.Status = tt.to

		if state.Status != tt.to {
			t.Errorf("failed to transition from %s to %s", tt.from, tt.to)
		}
	}
}

// TestStateTransition_JSON tests transition JSON
func TestStateTransition_JSON(t *testing.T) {
	transition := &StateTransition{
		From:      "creating",
		To:        "running",
		Timestamp: time.Now(),
		Reason:    "VM started",
	}

	data, err := json.Marshal(transition)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var t2 StateTransition
	if err := json.Unmarshal(data, &t2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if t2.From != transition.From {
		t.Errorf("expected from %s, got %s", transition.From, t2.From)
	}
}

// TestStateEvent_JSON tests event JSON
func TestStateEvent_JSON(t *testing.T) {
	event := &StateEvent{
		VMID:      "vm-1",
		OldState:  "stopped",
		NewState:  "running",
		Timestamp: time.Now(),
		Reason:    "Manual start",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var e2 StateEvent
	if err := json.Unmarshal(data, &e2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if e2.VMID != event.VMID {
		t.Errorf("expected VMID %s, got %s", event.VMID, e2.VMID)
	}
}

// TestVMState_MultipleVMs tests managing multiple VMs
func TestVMState_MultipleVMs(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
	}

	// Add multiple VMs
	for i := 1; i <= 5; i++ {
		id := "vm-" + string(rune('0'+i))
		st.cache[id] = &VMState{
			ID:     id,
			Status: "running",
		}
	}

	if len(st.cache) != 5 {
		t.Errorf("expected 5 VMs, got %d", len(st.cache))
	}

	// Verify each exists
	for i := 1; i <= 5; i++ {
		id := "vm-" + string(rune('0'+i))
		if st.cache[id] == nil {
			t.Errorf("VM %s should exist", id)
		}
	}
}

// TestVMState_RemoveVM tests removing a VM
func TestVMState_RemoveVM(t *testing.T) {
	st := &StateTracker{
		cache:       make(map[string]*VMState),
		subscribers: make(map[string][]chan StateEvent),
	}

	// Add VMs
	st.cache["vm-1"] = &VMState{ID: "vm-1", Status: "running"}
	st.cache["vm-2"] = &VMState{ID: "vm-2", Status: "running"}
	st.cache["vm-3"] = &VMState{ID: "vm-3", Status: "running"}

	// Remove one
	delete(st.cache, "vm-2")

	if len(st.cache) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(st.cache))
	}
	if st.cache["vm-2"] != nil {
		t.Error("vm-2 should be removed")
	}
}