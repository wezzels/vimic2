// Package pool provides VM state tracking
package pool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// StateTracker manages VM state in memory with persistence
type StateTracker struct {
	db          types.PipelineDB
	cache       map[string]*VMState
	stateFile   string
	eventChan   chan StateEvent
	subscribers map[string][]chan StateEvent
	mu          sync.RWMutex
}

// VMState represents the state of a VM
type VMState struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	IPAddress string    `json:"ip_address"`
	PoolName  string    `json:"pool_name"`
	Template  string    `json:"template"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	VMID      string    `json:"vm_id"`
	OldState  string    `json:"old_state"`
	NewState  string    `json:"new_state"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

// NewStateTracker creates a new state tracker
func NewStateTracker(db types.PipelineDB, stateFile string) (*StateTracker, error) {
	st := &StateTracker{
		db:          db,
		cache:       make(map[string]*VMState),
		stateFile:   stateFile,
		eventChan:   make(chan StateEvent, 1000),
		subscribers: make(map[string][]chan StateEvent),
	}

	// Load state from disk
	if err := st.loadState(); err != nil {
		return nil, err
	}

	return st, nil
}

func (st *StateTracker) loadState() error {
	if st.stateFile == "" {
		return nil
	}

	data, err := ioutil.ReadFile(st.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
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

func (st *StateTracker) saveState() error {
	if st.stateFile == "" {
		return nil
	}

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

	return ioutil.WriteFile(st.stateFile, data, 0644)
}

// GetVM returns VM state by ID
func (st *StateTracker) GetVM(vmID string) (*VMState, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	state, ok := st.cache[vmID]
	if !ok {
		return nil, ErrVMNotFound
	}

	return state, nil
}

// SetVM sets VM state
func (st *StateTracker) SetVM(vm *VMState) error {
	st.mu.Lock()
	st.cache[vm.ID] = vm
	st.mu.Unlock()
	return st.saveState()
}

// DeleteVM removes VM state
func (st *StateTracker) DeleteVM(vmID string) error {
	st.mu.Lock()
	delete(st.cache, vmID)
	st.mu.Unlock()
	return st.saveState()
}

// ListVMs returns all VMs
func (st *StateTracker) ListVMs() []*VMState {
	st.mu.RLock()
	defer st.mu.RUnlock()

	vms := make([]*VMState, 0, len(st.cache))
	for _, vm := range st.cache {
		vms = append(vms, vm)
	}

	return vms
}

// Subscribe subscribes to state events
func (st *StateTracker) Subscribe(id string) chan StateEvent {
	st.mu.Lock()
	defer st.mu.Unlock()

	ch := make(chan StateEvent, 100)
	st.subscribers[id] = append(st.subscribers[id], ch)
	return ch
}

// Unsubscribe unsubscribes from state events
func (st *StateTracker) Unsubscribe(id string, ch chan StateEvent) {
	st.mu.Lock()
	defer st.mu.Unlock()

	subs, ok := st.subscribers[id]
	if !ok {
		return
	}

	for i, sub := range subs {
		if sub == ch {
			st.subscribers[id] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

// ErrVMNotFound is returned when a VM is not found
var ErrVMNotFound = fmt.Errorf("vm not found")

// SaveState explicitly saves state to disk
func (st *StateTracker) SaveState() error {
	return st.saveState()
}
