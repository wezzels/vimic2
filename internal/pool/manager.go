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

// poolConfig represents pool configuration
type poolConfig struct {
	Template     string `json:"template"`
	MinSize      int    `json:"min_size"`
	MaxSize      int    `json:"max_size"`
	CPU          int    `json:"cpu"`
	Memory       int    `json:"memory"`
	DiskSize     int64  `json:"disk_size"`
	PreAllocated int    `json:"pre_allocated"`
}

// VM represents a virtual machine
type VM struct {
	ID          string       `json:"id"`
	PoolID      string       `json:"pool_id"`
	TemplateID  string       `json:"template_id"`
	Name        string       `json:"name"`
	IP          string       `json:"ip"`
	MAC         string       `json:"mac"`
	CPU         int          `json:"cpu"`
	Memory      int          `json:"memory"`
	State       VMState      `json:"state"`
	OverlayID   string       `json:"overlay_id"`
	CreatedAt   time.Time    `json:"created_at"`
	DestroyedAt *time.Time   `json:"destroyed_at,omitempty"`
}

// Overlay represents a copy-on-write overlay
type Overlay struct {
	ID          string       `json:"id"`
	TemplateID  string       `json:"template_id"`
	VMID        string       `json:"vm_id"`
	Path        string       `json:"path"`
	ActualSize  int64        `json:"actual_size"`
	CreatedAt   time.Time    `json:"created_at"`
	DestroyedAt *time.Time   `json:"destroyed_at,omitempty"`
}

// Pool represents a VM pool
type Pool struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	TemplateID  string       `json:"template_id"`
	MinSize     int          `json:"min_size"`
	MaxSize     int          `json:"max_size"`
	CurrentSize int          `json:"current_size"`
	CPU         int          `json:"cpu"`
	Memory      int          `json:"memory"`
	CreatedAt   time.Time    `json:"created_at"`
	VMs         []string     `json:"vms"`
}

// VMState represents VM state
type VMState string

const (
	VMStateCreating  VMState = "creating"
	VMStateRunning   VMState = "running"
	VMStateIdle      VMState = "idle"
	VMStateBusy      VMState = "busy"
	VMStateStopping  VMState = "stopping"
	VMStateStopped   VMState = "stopped"
	VMStateDestroyed VMState = "destroyed"
	VMStateError     VMState = "error"
)

// VMEvent represents a VM state change event
type VMEvent struct {
	VM       *VM       `json:"vm"`
	OldState VMState   `json:"old_state"`
	NewState VMState   `json:"new_state"`
	Timestamp time.Time `json:"timestamp"`
	PoolID   string    `json:"pool_id"`
}

// NewPoolManager creates a new pool manager
func NewPoolManager(db *pipeline.PipelineDB, templateMgr *TemplateManager, configPath string) (*PoolManager, error) {
	pm := &PoolManager{
		db:         db,
		templateMgr: templateMgr,
		vms:        make(map[string]*VM),
		overlays:   make(map[string]*Overlay),
		pools:      make(map[string]*Pool),
		eventChan:  make(chan VMEvent, 1000),
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

	// Initialize pools from config
	for name, cfg := range pm.config {
		if _, ok := pm.pools[name]; !ok {
			pool := &Pool{
				ID:         generateID("pool"),
				Name:       name,
				TemplateID: cfg.Template,
				MinSize:    cfg.MinSize,
				MaxSize:    cfg.MaxSize,
				CPU:        cfg.CPU,
				Memory:     cfg.Memory,
				CreatedAt:  time.Now(),
				VMs:        []string{},
			}
			pm.pools[name] = pool
		}
	}

	// Start event processor
	go pm.processEvents()

	return pm, nil
}

// loadConfig loads pool configuration
func (pm *PoolManager) loadConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config file yet
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

// loadState loads VM state from disk
func (pm *PoolManager) loadState() error {
	stateFile := pm.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
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

// saveState saves VM state to disk
func (pm *PoolManager) saveState() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	state := struct {
		VMs      []*VM      `json:"vms"`
		Overlays []*Overlay `json:"overlays"`
		Pools    []*Pool    `json:"pools"`
	}{
		VMs:      make([]*VM, 0, len(pm.vms)),
		Overlays: make([]*Overlay, 0, len(pm.overlays)),
		Pools:    make([]*Pool, 0, len(pm.pools)),
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

	stateFile := pm.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (pm *PoolManager) getStateFile() string {
	if pm.stateFile != "" {
		return pm.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "vm-state.json")
}

// SetStateFile sets the state file path
func (pm *PoolManager) SetStateFile(path string) {
	pm.stateFile = path
}

// CreatePool creates a new VM pool
func (pm *PoolManager) CreatePool(ctx context.Context, name, templateID string, minSize, maxSize, cpu, memory int) (*Pool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if pool already exists
	if _, ok := pm.pools[name]; ok {
		return nil, fmt.Errorf("pool already exists: %s", name)
	}

	// Validate template
	template, err := pm.templateMgr.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
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
		CreatedAt:   time.Now(),
		VMs:         []string{},
	}

	pm.pools[name] = pool

	// Save to database
	dbPool := &pipeline.Pool{
		ID:          pool.ID,
		Name:        pool.Name,
		TemplateID:  pool.TemplateID,
		MinSize:     pool.MinSize,
		MaxSize:     pool.MaxSize,
		CurrentSize: pool.CurrentSize,
		CPU:         pool.CPU,
		Memory:      pool.Memory,
		CreatedAt:   pool.CreatedAt,
	}
	if err := pm.db.SavePool(ctx, dbPool); err != nil {
		delete(pm.pools, name)
		return nil, fmt.Errorf("failed to save pool: %w", err)
	}

	// Save state
	if err := pm.saveState(); err != nil {
		pm.db.DeletePool(ctx, pool.ID)
		delete(pm.pools, name)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return pool, nil
}

// GetPool returns a pool by name
func (pm *PoolManager) GetPool(name string) (*Pool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pool, ok := pm.pools[name]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", name)
	}

	return pool, nil
}

// ListPools returns all pools
func (pm *PoolManager) ListPools() []*Pool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pools := make([]*Pool, 0, len(pm.pools))
	for _, pool := range pm.pools {
		pools = append(pools, pool)
	}
	return pools
}

// AcquireVM acquires a VM from a pool
func (pm *PoolManager) AcquireVM(ctx context.Context, poolName string) (*VM, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, ok := pm.pools[poolName]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	// Check pool capacity
	if pool.CurrentSize >= pool.MaxSize {
		return nil, fmt.Errorf("pool at max capacity: %s (max: %d)", poolName, pool.MaxSize)
	}

	// Create VM
	vm, err := pm.createVM(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Update pool
	pool.CurrentSize++
	pool.VMs = append(pool.VMs, vm.ID)

	// Save to database
	if err := pm.db.UpdatePoolSize(ctx, pool.ID, 1); err != nil {
		pm.destroyVM(ctx, vm.ID)
		return nil, fmt.Errorf("failed to update pool size: %w", err)
	}

	// Save state
	if err := pm.saveState(); err != nil {
		pm.destroyVM(ctx, vm.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return vm, nil
}

// ReleaseVM releases a VM back to the pool
func (pm *PoolManager) ReleaseVM(ctx context.Context, vmID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	vm, ok := pm.vms[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	pool, ok := pm.pools[vm.PoolID]
	if !ok {
		return fmt.Errorf("pool not found: %s", vm.PoolID)
	}

	// Update VM state
	vm.State = VMStateIdle
	if err := pm.db.UpdateVMState(ctx, vmID, pipeline.VMState(vm.State)); err != nil {
		return fmt.Errorf("failed to update VM state: %w", err)
	}

	// Remove from pool's VM list
	for i, id := range pool.VMs {
		if id == vmID {
			pool.VMs = append(pool.VMs[:i], pool.VMs[i+1:]...)
			break
		}
	}

	// Check if we should destroy the VM
	if pool.CurrentSize > pool.MinSize {
		// Destroy VM if above minimum
		if err := pm.destroyVM(ctx, vmID); err != nil {
			return fmt.Errorf("failed to destroy VM: %w", err)
		}
		pool.CurrentSize--
		if err := pm.db.UpdatePoolSize(ctx, pool.ID, -1); err != nil {
			return fmt.Errorf("failed to update pool size: %w", err)
		}
	}

	// Save state
	if err := pm.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// createVM creates a new VM
func (pm *PoolManager) createVM(ctx context.Context, pool *Pool) (*VM, error) {
	vmID := generateID("vm")
	vmName := fmt.Sprintf("%s-%s", pool.Name, vmID[:8])

	// Create overlay
	overlay, err := pm.templateMgr.CreateOverlay(pool.TemplateID, vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to create overlay: %w", err)
	}

	// Create VM record
	vm := &VM{
		ID:         vmID,
		PoolID:     pool.ID,
		TemplateID: pool.TemplateID,
		Name:       vmName,
		CPU:        pool.CPU,
		Memory:     pool.Memory,
		State:      VMStateCreating,
		OverlayID:  overlay.ID,
		CreatedAt:  time.Now(),
	}

	pm.vms[vmID] = vm

	// Save to database
	dbVM := &pipeline.VM{
		ID:         vm.ID,
		PoolID:     vm.PoolID,
		TemplateID: vm.TemplateID,
		Name:       vm.Name,
		CPU:        vm.CPU,
		Memory:     vm.Memory,
		State:      pipeline.VMState(vm.State),
		OverlayID:  vm.OverlayID,
		CreatedAt:  vm.CreatedAt,
	}
	if err := pm.db.SaveVM(ctx, dbVM); err != nil {
		delete(pm.vms, vmID)
		pm.templateMgr.DeleteOverlay(overlay.ID)
		return nil, fmt.Errorf("failed to save VM: %w", err)
	}

	// Start VM creation in background
	go pm.startVM(vm, overlay)

	return vm, nil
}

// startVM starts a VM
func (pm *PoolManager) startVM(vm *VM, overlay *Overlay) {
	ctx := context.Background()

	// Generate MAC address
	mac := generateMAC()

	// Start VM with QEMU
	cmd := exec.CommandContext(ctx, "qemu-system-x86_64",
		"-name", vm.Name,
		"-machine", "accel=kvm,type=pc",
		"-cpu", "host",
		"-m", fmt.Sprintf("%d", vm.Memory),
		"-smp", fmt.Sprintf("%d", vm.CPU),
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", overlay.Path),
		"-netdev", fmt.Sprintf("bridge,id=net0,br=virbr0"),
		"-device", fmt.Sprintf("virtio-net-pci,netdev=net0,mac=%s", mac),
		"-nographic",
		"-serial", "mon:stdio",
		"-daemonize",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		pm.emitEvent(VMEvent{
			VM:       vm,
			OldState: VMStateCreating,
			NewState: VMStateError,
			Timestamp: time.Now(),
			PoolID:   vm.PoolID,
		})
		pm.db.UpdateVMState(ctx, vm.ID, pipeline.VMStateError)
		return
	}

	// Update VM state
	pm.mu.Lock()
	vm.State = VMStateRunning
	vm.MAC = mac
	pm.mu.Unlock()

	pm.db.UpdateVMState(ctx, vm.ID, pipeline.VMStateRunning)

	// Emit event
	pm.emitEvent(VMEvent{
		VM:        vm,
		OldState:  VMStateCreating,
		NewState:  VMStateRunning,
		Timestamp: time.Now(),
		PoolID:    vm.PoolID,
	})
}

// destroyVM destroys a VM
func (pm *PoolManager) destroyVM(ctx context.Context, vmID string) error {
	vm, ok := pm.vms[vmID]
	if !ok {
		return fmt.Errorf("VM not found: %s", vmID)
	}

	// Stop VM
	cmd := exec.CommandContext(ctx, "virsh", "destroy", vm.Name)
	cmd.Run() // Ignore error if VM not running

	// Undefine VM
	cmd = exec.CommandContext(ctx, "virsh", "undefine", vm.Name)
	cmd.Run() // Ignore error if VM not defined

	// Delete overlay
	if err := pm.templateMgr.DeleteOverlay(vm.OverlayID); err != nil {
		// Log error but continue
	}

	// Update state
	now := time.Now()
	vm.State = VMStateDestroyed
	vm.DestroyedAt = &now

	// Delete from database
	if err := pm.db.DeleteVM(ctx, vmID); err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	// Delete from memory
	delete(pm.vms, vmID)

	return nil
}

// GetVM returns a VM by ID
func (pm *PoolManager) GetVM(vmID string) (*VM, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	vm, ok := pm.vms[vmID]
	if !ok {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return vm, nil
}

// ListVMs returns all VMs in a pool
func (pm *PoolManager) ListVMs(poolName string) ([]*VM, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pool, ok := pm.pools[poolName]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}

	vms := make([]*VM, 0, len(pool.VMs))
	for _, vmID := range pool.VMs {
		if vm, ok := pm.vms[vmID]; ok {
			vms = append(vms, vm)
		}
	}

	return vms, nil
}

// GetVMStats returns VM statistics
func (pm *PoolManager) GetVMStats() map[string]int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := make(map[string]int)
	for _, vm := range pm.vms {
		stats[string(vm.State)]++
	}
	return stats
}

// Events returns the event channel
func (pm *PoolManager) Events() <-chan VMEvent {
	return pm.eventChan
}

// emitEvent emits a VM event
func (pm *PoolManager) emitEvent(event VMEvent) {
	select {
	case pm.eventChan <- event:
	default:
		// Channel full, drop event
	}
}

// processEvents processes VM events
func (pm *PoolManager) processEvents() {
	for event := range pm.eventChan {
		// Log event
		fmt.Printf("[VM Event] %s: %s -> %s (pool: %s)\n",
			event.VM.ID,
			event.OldState,
			event.NewState,
			event.PoolID,
		)

		// Could emit to external systems here (webhooks, etc.)
	}
}

// PreAllocateVMs pre-allocates VMs for a pool
func (pm *PoolManager) PreAllocateVMs(ctx context.Context, poolName string, count int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, ok := pm.pools[poolName]
	if !ok {
		return fmt.Errorf("pool not found: %s", poolName)
	}

	// Check capacity
	if pool.CurrentSize+count > pool.MaxSize {
		return fmt.Errorf("would exceed max pool size: current=%d, requested=%d, max=%d",
			pool.CurrentSize, count, pool.MaxSize)
	}

	// Create VMs
	for i := 0; i < count; i++ {
		vm, err := pm.createVM(ctx, pool)
		if err != nil {
			// Rollback created VMs
			for _, vmID := range pool.VMs {
				pm.destroyVM(ctx, vmID)
			}
			return fmt.Errorf("failed to create VM %d: %w", i, err)
		}
		pool.VMs = append(pool.VMs, vm.ID)
	}

	pool.CurrentSize += count

	// Save state
	if err := pm.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// Cleanup cleans up destroyed VMs
func (pm *PoolManager) Cleanup(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cleaned := 0
	for _, vm := range pm.vms {
		if vm.State == VMStateDestroyed {
			delete(pm.vms, vm.ID)
			cleaned++
		}
	}

	// Save state
	if err := pm.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("[PoolManager] Cleaned up %d destroyed VMs\n", cleaned)
	return nil
}

// Close closes the pool manager
func (pm *PoolManager) Close() error {
	close(pm.eventChan)
	return pm.saveState()
}

// Helper functions

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

func generateMAC() string {
	// Generate random MAC address (locally administered)
	return fmt.Sprintf("52:54:00:%02x:%02x:%02x",
		time.Now().UnixNano()&0xff,
		(time.Now().UnixNano()>>8)&0xff,
		(time.Now().UnixNano()>>16)&0xff,
	)
}