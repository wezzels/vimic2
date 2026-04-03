// Package pool provides VM pool management
package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// PoolManager manages VM pools with QEMU backing files
type PoolManager struct {
	db            types.PipelineDB
	templateMgr   *TemplateManager
	config        map[string]poolConfig
	vms           map[string]*VM
	overlays      map[string]*Overlay
	pools         map[string]*Pool
	mu            sync.RWMutex
	stateFile     string
	eventChan     chan VMEvent
}

// Pool represents a VM pool
type Pool struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TemplateID  string    `json:"template_id"`
	MinSize     int       `json:"min_size"`
	MaxSize     int       `json:"max_size"`
	CurrentSize int       `json:"current_size"`
	CPU         int       `json:"cpu"`
	Memory      int       `json:"memory"`
	VMs         []string  `json:"vms"`
	CreatedAt   time.Time `json:"created_at"`
}

// VM represents a virtual machine
type VM struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PoolID    string    `json:"pool_id"`
	Status    string    `json:"status"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// Overlay represents a copy-on-write overlay image
type Overlay struct {
	ID          string     `json:"id"`
	TemplateID  string     `json:"template_id"`
	VMID        string     `json:"vm_id"`
	Path        string     `json:"path"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

type poolConfig struct {
	Template string `json:"template"`
	MinSize  int    `json:"min_size"`
	MaxSize  int    `json:"max_size"`
	CPU      int    `json:"cpu"`
	Memory   int    `json:"memory"`
}

// VMStatus represents VM status
type VMStatus string

const (
	VMStatusCreating  VMStatus = "creating"
	VMStatusRunning    VMStatus = "running"
	VMStatusIdle       VMStatus = "idle"
	VMStatusBusy       VMStatus = "busy"
	VMStatusStopping   VMStatus = "stopping"
	VMStatusStopped    VMStatus = "stopped"
	VMStatusDestroyed  VMStatus = "destroyed"
	VMStatusError      VMStatus = "error"
)

// VMEvent represents a VM state change event
type VMEvent struct {
	VM        *VM       `json:"vm"`
	OldStatus VMStatus  `json:"old_status"`
	NewStatus VMStatus  `json:"new_status"`
	Timestamp time.Time `json:"timestamp"`
	PoolID    string    `json:"pool_id"`
}

// NewPoolManager creates a new pool manager
func NewPoolManager(db types.PipelineDB, templateMgr *TemplateManager, configPath string) (*PoolManager, error) {
	pm := &PoolManager{
		db:          db,
		templateMgr: templateMgr,
		vms:         make(map[string]*VM),
		overlays:    make(map[string]*Overlay),
		pools:       make(map[string]*Pool),
		eventChan:   make(chan VMEvent, 1000),
	}

	// Load configuration
	if configPath != "" {
		if err := pm.loadConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Load state from disk
	if err := pm.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Start event processor
	go pm.processEvents()

	return pm, nil
}

func (pm *PoolManager) loadConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var config struct {
		Pools map[string]poolConfig `json:"pools"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	pm.config = config.Pools
	return nil
}

func (pm *PoolManager) loadState() error {
	stateFile := pm.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var state struct {
		VMs      []*VM      `json:"vms"`
		Overlays []*Overlay `json:"overlays"`
		Pools    []*Pool    `json:"pools"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	for _, vm := range state.VMs {
		pm.vms[vm.ID] = vm
	}
	for _, overlay := range state.Overlays {
		pm.overlays[overlay.ID] = overlay
	}
	for _, pool := range state.Pools {
		pm.pools[pool.Name] = pool
	}

	return nil
}

func (pm *PoolManager) saveState() error {
	stateFile := pm.getStateFile()

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var state struct {
		VMs      []*VM      `json:"vms"`
		Overlays []*Overlay `json:"overlays"`
		Pools    []*Pool    `json:"pools"`
	}

	for _, vm := range pm.vms {
		state.VMs = append(state.VMs, vm)
	}
	for _, overlay := range pm.overlays {
		state.Overlays = append(state.Overlays, overlay)
	}
	for _, pool := range pm.pools {
		state.Pools = append(state.Pools, pool)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

func (pm *PoolManager) getStateFile() string {
	if pm.stateFile != "" {
		return pm.stateFile
	}
	return "/var/lib/vimic2/pool-state.json"
}

func (pm *PoolManager) processEvents() {
	for event := range pm.eventChan {
		// Process events asynchronously
		_ = event
	}
}

// CreatePool creates a new VM pool
func (pm *PoolManager) CreatePool(ctx context.Context, name, templateID string, minSize, maxSize, cpu, memory int) (*Pool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.pools[name]; exists {
		return nil, fmt.Errorf("pool already exists: %s", name)
	}

	pool := &Pool{
		ID:          generateID("pool"),
		Name:        name,
		TemplateID:  templateID,
		MinSize:     minSize,
		MaxSize:     maxSize,
		CurrentSize: 0,
		CPU:         cpu,
		Memory:      memory,
		VMs:         []string{},
		CreatedAt:   time.Now(),
	}

	pm.pools[name] = pool
	pm.saveState()

	return pool, nil
}

// GetPool returns a pool by name
func (pm *PoolManager) GetPool(name string) (*Pool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pool, exists := pm.pools[name]
	if !exists {
		return nil, fmt.Errorf("pool not found: %s", name)
	}

	return pool, nil
}

// ListPools returns all pools
func (pm *PoolManager) ListPools() ([]*types.PoolState, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pools := make([]*types.PoolState, 0, len(pm.pools))
	for _, pool := range pm.pools {
		pools = append(pools, &types.PoolState{
			Name:      pool.Name,
			Capacity:  pool.MaxSize,
			Available: pool.MaxSize - pool.CurrentSize,
			Busy:      pool.CurrentSize,
		})
	}
	return pools, nil
}

// DeletePool deletes a pool
func (pm *PoolManager) DeletePool(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, exists := pm.pools[name]
	if !exists {
		return fmt.Errorf("pool not found: %s", name)
	}

	// Check for active VMs
	if len(pool.VMs) > 0 {
		return fmt.Errorf("pool has active VMs")
	}

	delete(pm.pools, name)
	pm.saveState()

	return nil
}

// AllocateVM allocates a VM from a pool
func (pm *PoolManager) AllocateVM(poolName string) (*types.VMState, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, exists := pm.pools[poolName]
	if !exists {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	vm := &VM{
		ID:        generateID("vm"),
		Name:      fmt.Sprintf("vm-%s", generateID("")),
		PoolID:    pool.ID,
		Status:     "creating",
		CreatedAt: time.Now(),
	}

	pm.vms[vm.ID] = vm
	pool.VMs = append(pool.VMs, vm.ID)
	pool.CurrentSize++
	pm.saveState()

	return &types.VMState{
		ID:     vm.ID,
		Name:   vm.Name,
		Status: vm.Status,
	}, nil
}

// ReleaseVM releases a VM back to the pool
func (pm *PoolManager) ReleaseVM(vmID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	vm, exists := pm.vms[vmID]
	if !exists {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	// Find pool and remove VM
	for _, pool := range pm.pools {
		if pool.ID == vm.PoolID {
			newVMs := []string{}
			for _, id := range pool.VMs {
				if id != vmID {
					newVMs = append(newVMs, id)
				}
			}
			pool.VMs = newVMs
			pool.CurrentSize--
			break
		}
	}

	delete(pm.vms, vmID)
	pm.saveState()

	return nil
}

// AllocateVM implements types.PoolManagerInterface
func (pm *PoolManager) AllocateVMContext(ctx context.Context, poolName string) (*types.VMState, error) {
	return pm.AllocateVM(poolName)
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// AllocateVM implements types.PoolManagerInterface
func (pm *PoolManager) AllocateVMContext(ctx context.Context, poolName string) (*types.VMState, error) {
	return pm.AllocateVM(poolName)
}