// Package pool provides VM state tracking
package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// StateTracker manages VM state in memory with persistence
type StateTracker struct {
	db           types.PipelineDB
	cache        map[string]*VMState
	stateFile    string
	eventChan    chan StateEvent
	subscribers  map[string][]chan StateEvent
	mu           sync.RWMutex
}

// VMState represents the state of a VM
type VMState struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	PoolID          string       `json:"pool_id"`
	TemplateID      string       `json:"template_id"`
	OverlayID       string       `json:"overlay_id"`
	Status          string       `json:"status"`
	IPAddress       string       `json:"ip_address"`
	MACAddress      string       `json:"mac_address"`
	CPU             int          `json:"cpu"`
	Memory          int          `json:"memory"`
	DiskUsed        int64        `json:"disk_used"`
	DiskTotal       int64        `json:"disk_total"`
	CPUUsage        float64      `json:"cpu_usage"`
	MemoryUsage     float64      `json:"memory_usage"`
	NetworkRx       int64        `json:"network_rx"`
	NetworkTx       int64        `json:"network_tx"`
	RunnerConnected bool         `json:"runner_connected"`
	CurrentJob      string       `json:"current_job"`
	LastHeartbeat   time.Time    `json:"last_heartbeat"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	TransitionCount int          `json:"transition_count"`
	History         []StateTransition `json:"history"`
}

// StateTransition represents a state transition
type StateTransition struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

// StateEvent represents a state change event
type StateEvent struct {
	VMID      string       `json:"vm_id"`
	OldState  string       `json:"old_state"`
	NewState  string       `json:"new_state"`
	Timestamp time.Time    `json:"timestamp"`
	Reason    string       `json:"reason"`
}

// NewStateTracker creates a new state tracker
func NewStateTracker(db *pipeline.PipelineDB, stateFile string) (*StateTracker, error) {
	st := &StateTracker{
		db:          db,
		cache:       make(map[string]*VMState),
		stateFile:   stateFile,
		eventChan:   make(chan StateEvent, 1000),
		subscribers: make(map[string][]chan StateEvent),
	}

	// Load state from disk
	if err := st.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Start event processor
	go st.processEvents()

	return st, nil
}

// loadState loads state from disk
func (st *StateTracker) loadState() error {
	data, err := ioutil.ReadFile(st.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var states []*VMState
	if err := json.Unmarshal(data, &states); err != nil {
		return err
	}

	for _, state := range states {
		st.cache[state.ID] = state
	}

	return nil
}

// saveState saves state to disk
func (st *StateTracker) saveState() error {
	st.mu.RLock()
	defer st.mu.RUnlock()

	states := make([]*VMState, 0, len(st.cache))
	for _, state := range st.cache {
		states = append(states, state)
	}

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(st.stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(st.stateFile, data, 0644)
}

// GetState returns the state of a VM
func (st *StateTracker) GetState(vmID string) (*VMState, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	state, ok := st.cache[vmID]
	if !ok {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return state, nil
}

// SetState sets the state of a VM
func (st *StateTracker) SetState(vmID, newState, reason string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.cache[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	oldState := state.Status
	now := time.Now()

	// Record transition
	transition := StateTransition{
		From:      oldState,
		To:        newState,
		Timestamp: now,
		Reason:    reason,
	}

	state.History = append(state.History, transition)
	state.TransitionCount++
	state.Status = newState
	state.UpdatedAt = now

	// Emit event
	st.emitEvent(StateEvent{
		VMID:      vmID,
		OldState:  oldState,
		NewState:  newState,
		Timestamp: now,
		Reason:    reason,
	})

	// Save to database
	ctx := context.Background()
	if err := st.db.UpdateVMState(ctx, vmID, pipeline.VMState(newState)); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	// Save state
	if err := st.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// CreateState creates a new VM state entry
func (st *StateTracker) CreateState(vm *VM) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if _, ok := st.cache[vm.ID]; ok {
		return fmt.Errorf("VM already exists: %s", vm.ID)
	}

	now := time.Now()
	state := &VMState{
		ID:              vm.ID,
		Name:            vm.Name,
		PoolID:          vm.PoolID,
		TemplateID:      vm.TemplateID,
		OverlayID:       vm.OverlayID,
		Status:          string(vm.State),
		IPAddress:       vm.IP,
		MACAddress:      vm.MAC,
		CPU:             vm.CPU,
		Memory:          vm.Memory,
		RunnerConnected: false,
		LastHeartbeat:   now,
		CreatedAt:       now,
		UpdatedAt:       now,
		TransitionCount: 0,
		History:         []StateTransition{},
	}

	st.cache[vm.ID] = state

	// Save state
	if err := st.saveState(); err != nil {
		delete(st.cache, vm.ID)
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// DeleteState removes a VM state entry
func (st *StateTracker) DeleteState(vmID string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if _, ok := st.cache[vmID]; !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	delete(st.cache, vmID)

	// Save state
	if err := st.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// UpdateMetrics updates VM metrics
func (st *StateTracker) UpdateMetrics(vmID string, metrics *VMMetrics) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.cache[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	state.CPUUsage = metrics.CPUUsage
	state.MemoryUsage = metrics.MemoryUsage
	state.DiskUsed = metrics.DiskUsed
	state.DiskTotal = metrics.DiskTotal
	state.NetworkRx = metrics.NetworkRx
	state.NetworkTx = metrics.NetworkTx
	state.UpdatedAt = time.Now()

	// Save state
	if err := st.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// VMMetrics represents VM metrics
type VMMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsed    int64   `json:"disk_used"`
	DiskTotal   int64   `json:"disk_total"`
	NetworkRx   int64   `json:"network_rx"`
	NetworkTx   int64   `json:"network_tx"`
}

// UpdateRunnerStatus updates runner connection status
func (st *StateTracker) UpdateRunnerStatus(vmID string, connected bool, currentJob string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.cache[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	state.RunnerConnected = connected
	state.CurrentJob = currentJob
	state.LastHeartbeat = time.Now()
	state.UpdatedAt = time.Now()

	// Update state based on runner status
	if connected && state.Status == string(VMStateIdle) {
		state.Status = string(VMStateBusy)
	} else if !connected && state.Status == string(VMStateBusy) {
		state.Status = string(VMStateIdle)
	}

	// Save state
	if err := st.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// ListByState returns VMs in a specific state
func (st *StateTracker) ListByState(status string) []*VMState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var result []*VMState
	for _, state := range st.cache {
		if state.Status == status {
			result = append(result, state)
		}
	}
	return result
}

// ListByPool returns VMs in a specific pool
func (st *StateTracker) ListByPool(poolID string) []*VMState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var result []*VMState
	for _, state := range st.cache {
		if state.PoolID == poolID {
			result = append(result, state)
		}
	}
	return result
}

// ListAll returns all VM states
func (st *StateTracker) ListAll() []*VMState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	result := make([]*VMState, 0, len(st.cache))
	for _, state := range st.cache {
		result = append(result, state)
	}
	return result
}

// Subscribe subscribes to state change events
func (st *StateTracker) Subscribe(vmID string) <-chan StateEvent {
	st.mu.Lock()
	defer st.mu.Unlock()

	ch := make(chan StateEvent, 100)
	st.subscribers[vmID] = append(st.subscribers[vmID], ch)
	return ch
}

// Unsubscribe unsubscribes from state change events
func (st *StateTracker) Unsubscribe(vmID string, ch <-chan StateEvent) {
	st.mu.Lock()
	defer st.mu.Unlock()

	subs := st.subscribers[vmID]
	for i, sub := range subs {
		if sub == ch {
			st.subscribers[vmID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

// emitEvent emits a state change event
func (st *StateTracker) emitEvent(event StateEvent) {
	select {
	case st.eventChan <- event:
	default:
		// Channel full, drop event
	}
}

// processEvents processes state change events
func (st *StateTracker) processEvents() {
	for event := range st.eventChan {
		st.mu.RLock()
		subs := st.subscribers[event.VMID]
		st.mu.RUnlock()

		// Send to subscribers
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetStats returns state statistics
func (st *StateTracker) GetStats() map[string]int {
	st.mu.RLock()
	defer st.mu.RUnlock()

	stats := make(map[string]int)
	for _, state := range st.cache {
		stats[state.Status]++
	}
	return stats
}

// GetTransitionHistory returns transition history for a VM
func (st *StateTracker) GetTransitionHistory(vmID string) ([]StateTransition, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	state, ok := st.cache[vmID]
	if !ok {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return state.History, nil
}

// PruneHistory prunes old history entries
func (st *StateTracker) PruneHistory(vmID string, maxEntries int) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	state, ok := st.cache[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	if len(state.History) > maxEntries {
		state.History = state.History[len(state.History)-maxEntries:]
	}

	// Save state
	if err := st.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// Close closes the state tracker
func (st *StateTracker) Close() error {
	close(st.eventChan)
	return st.saveState()
}