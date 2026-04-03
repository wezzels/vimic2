// Package network provides network isolation for CI/CD pipelines
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// IsolationManager manages network isolation for pipelines
type IsolationManager struct {
	db        types.PipelineDB
	ovs       *OVSClient
	vlanAlloc *VLANAllocator
	ipam      *IPAMManager
	firewall  *FirewallManager
	stateFile string
	networks  map[string]*IsolatedNetwork
	mu        sync.RWMutex
}

// IsolatedNetwork represents an isolated network for a pipeline
type IsolatedNetwork struct {
	ID          string     `json:"id"`
	PipelineID  string     `json:"pipeline_id"`
	BridgeName  string     `json:"bridge_name"`
	VLANID      int        `json:"vlan_id"`
	CIDR        string     `json:"cidr"`
	Gateway     string     `json:"gateway"`
	DNS         []string   `json:"dns"`
	VMs         []string   `json:"vms"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	VLANStart      int      `json:"vlan_start"`
	VLANEnd        int      `json:"vlan_end"`
	BaseCIDR       string   `json:"base_cidr"`
	DNS            []string `json:"dns"`
	OVSBridge      string   `json:"ovs_bridge"`
	FirewallBackend string  `json:"firewall_backend"`
}

// NewIsolationManager creates a new isolation manager
func NewIsolationManager(db types.PipelineDB, config *NetworkConfig) (*IsolationManager, error) {
	im := &IsolationManager{
		db:       db,
		networks: make(map[string]*IsolatedNetwork),
	}

	// Create OVS client
	ovs := NewOVSClient()
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
	firewall, err := NewFirewallManager(FirewallBackend(config.FirewallBackend))
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

func (im *IsolationManager) loadState() error {
	stateFile := im.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var networks []*IsolatedNetwork
	if err := json.Unmarshal(data, &networks); err != nil {
		return err
	}

	for _, network := range networks {
		im.networks[network.ID] = network
	}

	return nil
}

func (im *IsolationManager) saveState() error {
	stateFile := im.getStateFile()

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

	return ioutil.WriteFile(stateFile, data, 0644)
}

func (im *IsolationManager) getStateFile() string {
	if im.stateFile != "" {
		return im.stateFile
	}
	return "/var/lib/vimic2/network-state.json"
}

// CreateNetwork creates a new isolated network for a pipeline
func (im *IsolationManager) CreateNetwork(ctx context.Context, pipelineID string) (*IsolatedNetwork, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

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
	if err := im.ovs.CreateBridge(bridgeName); err != nil {
		im.vlanAlloc.Release(vlanID)
		im.ipam.Release(cidr)
		return nil, fmt.Errorf("failed to create bridge: %w", err)
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
	im.saveState()

	return network, nil
}

// DestroyNetwork destroys a network
func (im *IsolationManager) DestroyNetwork(ctx context.Context, networkID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, exists := im.networks[networkID]
	if !exists {
		return fmt.Errorf("network not found: %s", networkID)
	}

	// Delete OVS bridge
	if err := im.ovs.DeleteBridge(network.BridgeName); err != nil {
		return fmt.Errorf("failed to delete bridge: %w", err)
	}

	// Release VLAN
	im.vlanAlloc.Release(network.VLANID)

	// Release CIDR
	im.ipam.Release(network.CIDR)

	// Mark as destroyed
	now := time.Now()
	network.DestroyedAt = &now

	delete(im.networks, networkID)
	im.saveState()

	return nil
}

// GetNetwork returns a network by ID
func (im *IsolationManager) GetNetwork(networkID string) (*IsolatedNetwork, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	network, exists := im.networks[networkID]
	if !exists {
		return nil, fmt.Errorf("network not found: %s", networkID)
	}

	return network, nil
}

// AddVMToNetwork adds a VM to a network
func (im *IsolationManager) AddVMToNetwork(networkID, vmID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, exists := im.networks[networkID]
	if !exists {
		return fmt.Errorf("network not found: %s", networkID)
	}

	network.VMs = append(network.VMs, vmID)
	im.saveState()

	return nil
}

// RemoveVMFromNetwork removes a VM from a network
func (im *IsolationManager) RemoveVMFromNetwork(networkID, vmID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	network, exists := im.networks[networkID]
	if !exists {
		return fmt.Errorf("network not found: %s", networkID)
	}

	newVMs := []string{}
	for _, id := range network.VMs {
		if id != vmID {
			newVMs = append(newVMs, id)
		}
	}
	network.VMs = newVMs
	im.saveState()

	return nil
}

// ListNetworks returns all networks
func (im *IsolationManager) ListNetworks() ([]*IsolatedNetwork, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	networks := make([]*IsolatedNetwork, 0, len(im.networks))
	for _, network := range im.networks {
		networks = append(networks, network)
	}

	return networks, nil
}

// CreateNetwork implements types.NetworkManagerInterface
func (im *IsolationManager) CreateNetworkContext(ctx context.Context, config *types.NetworkConfig) (string, error) {
	network, err := im.CreateNetwork(ctx, "pipeline")
	if err != nil {
		return "", err
	}
	return network.ID, nil
}

// DestroyNetwork implements types.NetworkManagerInterface
func (im *IsolationManager) DestroyNetworkContext(ctx context.Context, networkID string) error {
	return im.DestroyNetwork(ctx, networkID)
}

// GetNetwork implements types.NetworkManagerInterface
func (im *IsolationManager) GetNetworkByID(networkID string) (*types.NetworkConfig, error) {
	network, err := im.GetNetwork(networkID)
	if err != nil {
		return nil, err
	}
	return &types.NetworkConfig{
		VLAN: network.VLANID,
		CIDR: network.CIDR,
	}, nil
}

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
