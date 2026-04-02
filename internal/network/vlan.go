// Package network provides VLAN allocation
package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// VLANAllocator manages VLAN ID allocation
type VLANAllocator struct {
	start   int
	end     int
	used    map[int]bool
	mu      sync.RWMutex
	stateFile string
}

// NewVLANAllocator creates a new VLAN allocator
func NewVLANAllocator(start, end int) (*VLANAllocator, error) {
	if start < 1 || start > 4094 {
		return nil, fmt.Errorf("invalid VLAN start: %d (must be 1-4094)", start)
	}
	if end < start || end > 4094 {
		return nil, fmt.Errorf("invalid VLAN end: %d (must be >= start and <= 4094)", end)
	}

	va := &VLANAllocator{
		start: start,
		end:   end,
		used:  make(map[int]bool),
	}

	// Load state
	if err := va.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return va, nil
}

// loadState loads VLAN state from disk
func (va *VLANAllocator) loadState() error {
	stateFile := va.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var used map[int]bool
	if err := json.Unmarshal(data, &used); err != nil {
		return err
	}

	va.used = used
	if va.used == nil {
		va.used = make(map[int]bool)
	}

	return nil
}

// saveState saves VLAN state to disk
func (va *VLANAllocator) saveState() error {
	va.mu.RLock()
	defer va.mu.RUnlock()

	data, err := json.MarshalIndent(va.used, "", "  ")
	if err != nil {
		return err
	}

	stateFile := va.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (va *VLANAllocator) getStateFile() string {
	if va.stateFile != "" {
		return va.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "vlan-state.json")
}

// SetStateFile sets the state file path
func (va *VLANAllocator) SetStateFile(path string) {
	va.stateFile = path
}

// Allocate allocates a VLAN ID
func (va *VLANAllocator) Allocate() (int, error) {
	va.mu.Lock()
	defer va.mu.Unlock()

	// Find next available VLAN
	for vlan := va.start; vlan <= va.end; vlan++ {
		if !va.used[vlan] {
			va.used[vlan] = true
			// Save state
			if err := va.saveState(); err != nil {
				delete(va.used, vlan)
				return 0, fmt.Errorf("failed to save state: %w", err)
			}
			return vlan, nil
		}
	}

	return 0, fmt.Errorf("no available VLAN IDs (range: %d-%d)", va.start, va.end)
}

// Release releases a VLAN ID
func (va *VLANAllocator) Release(vlan int) error {
	va.mu.Lock()
	defer va.mu.Unlock()

	if !va.used[vlan] {
		return fmt.Errorf("VLAN %d is not allocated", vlan)
	}

	delete(va.used, vlan)

	// Save state
	if err := va.saveState(); err != nil {
		va.used[vlan] = true
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// Reclaim marks a VLAN as used (for loading existing state)
func (va *VLANAllocator) Reclaim(vlan int) {
	va.mu.Lock()
	defer va.mu.Unlock()
	va.used[vlan] = true
}

// IsAllocated checks if a VLAN is allocated
func (va *VLANAllocator) IsAllocated(vlan int) bool {
	va.mu.RLock()
	defer va.mu.RUnlock()
	return va.used[vlan]
}

// Used returns the number of used VLANs
func (va *VLANAllocator) Used() int {
	va.mu.RLock()
	defer va.mu.RUnlock()
	return len(va.used)
}

// Available returns the number of available VLANs
func (va *VLANAllocator) Available() int {
	va.mu.RLock()
	defer va.mu.RUnlock()
	return (va.end - va.start + 1) - len(va.used)
}

// ListUsed returns all used VLAN IDs
func (va *VLANAllocator) ListUsed() []int {
	va.mu.RLock()
	defer va.mu.RUnlock()

	used := make([]int, 0, len(va.used))
	for vlan := range va.used {
		used = append(used, vlan)
	}
	return used
}

// ListAvailable returns all available VLAN IDs
func (va *VLANAllocator) ListAvailable() []int {
	va.mu.RLock()
	defer va.mu.RUnlock()

	available := make([]int, 0)
	for vlan := va.start; vlan <= va.end; vlan++ {
		if !va.used[vlan] {
			available = append(available, vlan)
		}
	}
	return available
}

// Reset resets the VLAN allocator
func (va *VLANAllocator) Reset() error {
	va.mu.Lock()
	defer va.mu.Unlock()

	va.used = make(map[int]bool)

	// Save state
	if err := va.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}