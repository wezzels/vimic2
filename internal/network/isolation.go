// Package network provides network isolation for CI/CD pipelines
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// IsolationManager manages network isolation for pipelines
type IsolationManager struct {
	db           types.PipelineDB
	ovs          *OVSClient
	vlanAlloc    *VLANAllocator
	ipam         *IPAMManager
	firewall     *FirewallManager
	stateFile    string
	networks     map[string]*IsolatedNetwork
	mu           sync.RWMutex
}

// IsolatedNetwork represents an isolated network for a pipeline
type IsolatedNetwork struct {
	ID          string       `json:"id"`
	PipelineID  string       `json:"pipeline_id"`
	BridgeName  string       `json:"bridge_name"`
	VLANID      int          `json:"vlan_id"`
	CIDR        string       `json:"cidr"`
	Gateway     string       `json:"gateway"`
	DNS         []string     `json:"dns"`
	VMs         []string     `json:"vms"`
	CreatedAt   time.Time    `json:"created_at"`
	DestroyedAt *time.Time   `json:"destroyed_at,omitempty"`
}

// NewIsolationManager creates a new isolation manager
func NewIsolationManager(db *pipeline.PipelineDB, config *NetworkConfig) (*IsolationManager, error) {
	im := &IsolationManager{
		db:       db,
		networks: make(map[string]*IsolatedNetwork),
	}

	// Create OVS client
	ovsConfig := &OVSConfig{
		OVSPath: config.OVSPath,
	}
	ovs, err := NewOVSClient(ovsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OVS client: %w", err)
	}
	im.ovs = ovs

	// Create VLAN allocator
	vlanAlloc, err := NewVLANAllocator(config.VLANStart, config.VLANEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to create VLAN allocator: %w", err)
	}
	im.vlanAlloc = vlanAlloc

	// Create IPAM manager
	ipamConfig := &IPAMConfig{
		BaseCIDR: config.BaseCIDR,
		DNS:      config.DNS,
	}
	ipam, err := NewIPAMManager(ipamConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPAM manager: %w", err)
	}
	im.ipam = ipam

	// Create firewall manager
	firewall, err := NewFirewallManager(config.FirewallBackend)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall manager: %w", err)
	}
	im.firewall = firewall

	// Load state
	if err := im.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return im, nil
}

// loadState loads network state from disk
func (im *IsolationManager) loadState() error {
	stateFile := im.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var networks []*IsolatedNetwork
	if err := json.Unmarshal(data, &networks); err != nil {
		return err
	}

	for _, network := range networks {
		im.networks[network.ID] = network
		// Reclaim VLAN
		im.vlanAlloc.Reclaim(network.VLANID)
		// Reclaim CIDR
		im.ipam.Reclaim(network.CIDR)
	}

	return nil
}

// saveState saves network state to disk
func (im *IsolationManager) saveState() error {
	im.mu.RLock()
	defer im.mu.RUnlock()

	networks := make([]*IsolatedNetwork, 0, len(im.networks))
	for _, network := range im.networks {
		networks = append(networks, network)
	}

	data, err := json.MarshalIndent(networks, "", "  ")
	if err != nil {
		return err
	}

	stateFile := im.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (im *IsolationManager) getStateFile() string {
	if im.stateFile != "" {
		return im.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "network-state.json")
}

// SetStateFile sets the state file path
func (im *IsolationManager) SetStateFile(path string) {
	im.stateFile = path
}

// CreateNetwork creates an isolated network for a pipeline
func (im *IsolationManager) CreateNetwork(ctx context.Context, pipelineID string) (*IsolatedNetwork, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Check if network already exists for this pipeline
	for _, network := range im.networks {
		if network.PipelineID == pipelineID {
			return network, nil
		}
	}

	// Allocate VLAN
	vlanID, err := im.vlanAlloc.Allocate()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate VLAN: %w", err)
	}

	// Allocate CIDR
	cidr, gateway, err := im.ipam.Allocate()
	if err != nil {
		im.vlanAlloc.Release(vlanID)
		return nil, fmt.Errorf("failed to allocate CIDR: %w", err)
	}

	// Create bridge name
	bridgeName := fmt.Sprintf("vimic-br-%d", vlanID)

	// Create OVS bridge
	if err := im.ovs.CreateBridge(bridgeName, vlanID); err != nil {
		im.vlanAlloc.Release(vlanID)
		im.ipam.Release(cidr)
		return nil, fmt.Errorf("failed to create bridge: %w", err)
	}

	// Configure firewall rules
	if err := im.firewall.CreateIsolationRules(bridgeName, cidr, vlanID); err != nil {
		im.ovs.DeleteBridge(bridgeName)
		im.vlanAlloc.Release(vlanID)
		im.ipam.Release(cidr)
		return nil, fmt.Errorf("failed to create firewall rules: %w", err)
	}

	network := &IsolatedNetwork{
		ID:         generateNetworkID(pipelineID),
		PipelineID: pipelineID,
		BridgeName: bridgeName,
		VLANID:     vlanID,
		CIDR:       cidr,
		Gateway:    gateway,
		DNS:        im.ipam.GetDNS(),
		VMs:        []string{},
		CreatedAt:  time.Now(),
	}

	im.networks[network.ID] = network

	// Save to database
	dbNetwork := &pipeline.Network{
		ID:         network.ID,
		PipelineID: network.PipelineID,
		BridgeName: network.BridgeName,
		VLANID:     network.VLANID,
		CIDR:       network.CIDR,
		Gateway:    network.Gateway,
		CreatedAt:  network.CreatedAt,
	}
	if err := im.db.SaveNetwork(ctx, dbNetwork); err != nil {
		im.DeleteNetwork(ctx, network.ID)
		return nil, fmt.Errorf("failed to save network: %w", err)
	}

	// Save state
	if err := im.saveState(); err != nil {
		im.DeleteNetwork(ctx, network.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return network, nil
}

// DeleteNetwork deletes an isolated network
func (im *IsolationManager) DeleteNetwork(ctx context.Context, networkID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, ok := im.networks[networkID]
	if !ok {
		return fmt.Errorf("network not found: %s", networkID)
	}

	// Check for active VMs
	if len(network.VMs) > 0 {
		return fmt.Errorf("network has active VMs: %d", len(network.VMs))
	}

	// Delete firewall rules
	im.firewall.DeleteIsolationRules(network.BridgeName, network.CIDR, network.VLANID)

	// Delete OVS bridge
	im.ovs.DeleteBridge(network.BridgeName)

	// Release VLAN
	im.vlanAlloc.Release(network.VLANID)

	// Release CIDR
	im.ipam.Release(network.CIDR)

	// Mark as destroyed
	now := time.Now()
	network.DestroyedAt = &now

	// Delete from database
	if err := im.db.DeleteNetwork(ctx, networkID); err != nil {
		return fmt.Errorf("failed to delete network from database: %w", err)
	}

	// Delete from memory
	delete(im.networks, networkID)

	// Save state
	if err := im.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetNetwork returns a network by ID
func (im *IsolationManager) GetNetwork(networkID string) (*IsolatedNetwork, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	network, ok := im.networks[networkID]
	if !ok {
		return nil, fmt.Errorf("network not found: %s", networkID)
	}

	return network, nil
}

// GetNetworkByPipeline returns the network for a pipeline
func (im *IsolationManager) GetNetworkByPipeline(pipelineID string) (*IsolatedNetwork, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	for _, network := range im.networks {
		if network.PipelineID == pipelineID {
			return network, nil
		}
	}

	return nil, fmt.Errorf("network not found for pipeline: %s", pipelineID)
}

// ListNetworks returns all networks
func (im *IsolationManager) ListNetworks() []*IsolatedNetwork {
	im.mu.RLock()
	defer im.mu.RUnlock()

	networks := make([]*IsolatedNetwork, 0, len(im.networks))
	for _, network := range im.networks {
		networks = append(networks, network)
	}
	return networks
}

// AddVMToNetwork adds a VM to a network
func (im *IsolationManager) AddVMToNetwork(ctx context.Context, networkID, vmID, mac string) (string, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, ok := im.networks[networkID]
	if !ok {
		return "", fmt.Errorf("network not found: %s", networkID)
	}

	// Allocate IP for VM
	ip, err := im.ipam.AllocateIP(network.CIDR, mac)
	if err != nil {
		return "", fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Add VM to network
	network.VMs = append(network.VMs, vmID)

	// Save state
	if err := im.saveState(); err != nil {
		im.ipam.ReleaseIP(network.CIDR, ip)
		// Remove VM from network
		for i, id := range network.VMs {
			if id == vmID {
				network.VMs = append(network.VMs[:i], network.VMs[i+1:]...)
				break
			}
		}
		return "", fmt.Errorf("failed to save state: %w", err)
	}

	return ip, nil
}

// RemoveVMFromNetwork removes a VM from a network
func (im *IsolationManager) RemoveVMFromNetwork(ctx context.Context, networkID, vmID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, ok := im.networks[networkID]
	if !ok {
		return fmt.Errorf("network not found: %s", networkID)
	}

	// Remove VM from network
	for i, id := range network.VMs {
		if id == vmID {
			network.VMs = append(network.VMs[:i], network.VMs[i+1:]...)
			break
		}
	}

	// Save state
	if err := im.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// ConnectVMToNetwork connects a VM to a network using OVS
func (im *IsolationManager) ConnectVMToNetwork(vmName, bridgeName, mac string) error {
	// Add port to OVS bridge
	cmd := exec.Command("ovs-vsctl", "add-port", bridgeName, vmName+"-eth0",
		"--", "set", "Interface", vmName+"-eth0",
		"external-ids:vm="+vmName,
		"external-ids:mac="+mac)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add port to bridge: %w: %s", err, output)
	}

	return nil
}

// DisconnectVMFromNetwork disconnects a VM from a network
func (im *IsolationManager) DisconnectVMFromNetwork(vmName, bridgeName string) error {
	// Remove port from OVS bridge
	cmd := exec.Command("ovs-vsctl", "del-port", bridgeName, vmName+"-eth0")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove port from bridge: %w: %s", err, output)
	}

	return nil
}

// GetNetworkStats returns network statistics
func (im *IsolationManager) GetNetworkStats() map[string]int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	stats := map[string]int{
		"total_networks": len(im.networks),
		"active_vlans":   im.vlanAlloc.Used(),
		"active_cidrs":   im.ipam.Used(),
	}

	activeVMs := 0
	for _, network := range im.networks {
		activeVMs += len(network.VMs)
	}
	stats["active_vms"] = activeVMs

	return stats
}

// Cleanup cleans up destroyed networks
func (im *IsolationManager) Cleanup(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	cleaned := 0
	for id, network := range im.networks {
		if network.DestroyedAt != nil {
			delete(im.networks, id)
			cleaned++
		}
	}

	// Save state
	if err := im.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	fmt.Printf("[IsolationManager] Cleaned up %d destroyed networks\n", cleaned)
	return nil
}

// Close closes the isolation manager
func (im *IsolationManager) Close() error {
	return im.saveState()
}

// Helper functions

func generateNetworkID(pipelineID string) string {
	return fmt.Sprintf("net-%s-%s", pipelineID[:8], randomString(4))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	BaseCIDR        string   `json:"base_cidr"`
	VLANStart       int      `json:"vlan_start"`
	VLANEnd         int      `json:"vlan_end"`
	DNS             []string `json:"dns"`
	OVSPath         string   `json:"ovs_path"`
	FirewallBackend string   `json:"firewall_backend"`
}