// Package mockpool provides a mock PoolManagerInterface for testing
package mockpool

import (
	"sync"

	"github.com/stsgym/vimic2/internal/types"
)

// MockPoolManager implements types.PoolManagerInterface for testing
type MockPoolManager struct {
	mu    sync.RWMutex
	pools map[string]*types.PoolState
	vms   map[string]*types.VMState
}

// NewMockPoolManager creates a new mock pool manager
func NewMockPoolManager() *MockPoolManager {
	return &MockPoolManager{
		pools: make(map[string]*types.PoolState),
		vms:   make(map[string]*types.VMState),
	}
}

// AllocateVM allocates a VM from a pool
func (m *MockPoolManager) AllocateVM(poolName string) (*types.VMState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new VM
	vmID := "vm-" + poolName + "-" + randomID(8)
	vm := &types.VMState{
		ID:        vmID,
		Name:      vmID,
		Status:    "running",
		IPAddress: "10.0.0." + randomID(3),
	}

	m.vms[vmID] = vm

	// Update pool
	pool, ok := m.pools[poolName]
	if !ok {
		pool = &types.PoolState{
			Name:      poolName,
			Available: 10,
			Busy:      0,
		}
		m.pools[poolName] = pool
	}

	pool.Available--
	pool.Busy++

	return vm, nil
}

// ReleaseVM releases a VM back to its pool
func (m *MockPoolManager) ReleaseVM(vmID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.vms[vmID]
	if !ok {
		return nil
	}

	// Find pool and update
	for _, pool := range m.pools {
		pool.Available++
		pool.Busy--
	}

	delete(m.vms, vmID)
	return nil
}

// GetPool gets a pool by name
func (m *MockPoolManager) GetPool(name string) (*types.PoolState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, ok := m.pools[name]
	if !ok {
		return &types.PoolState{
			Name:      name,
			Available: 10,
			Busy:      0,
		}, nil
	}

	return pool, nil
}

// ListPools lists all pools
func (m *MockPoolManager) ListPools() ([]*types.PoolState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pools := make([]*types.PoolState, 0, len(m.pools))
	for _, p := range m.pools {
		pools = append(pools, p)
	}

	// If no pools, return default
	if len(pools) == 0 {
		pools = append(pools, &types.PoolState{
			Name:      "default",
			Available: 10,
			Busy:      0,
		})
	}

	return pools, nil
}

// AddPool adds a pool to the manager (for testing)
func (m *MockPoolManager) AddPool(name string, available, busy int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pools[name] = &types.PoolState{
		Name:      name,
		Available: available,
		Busy:      busy,
	}
}

// randomID generates a random ID string
func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// Verify MockPoolManager implements types.PoolManagerInterface
var _ types.PoolManagerInterface = (*MockPoolManager)(nil)