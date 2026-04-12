// Package mocknet provides a mock NetworkManagerInterface for testing
package mocknet

import (
	"sync"

	"github.com/stsgym/vimic2/internal/types"
)

// MockNetworkManager implements types.NetworkManagerInterface for testing
type MockNetworkManager struct {
	mu       sync.RWMutex
	networks map[string]*types.NetworkConfig
	nextID   int
}

// NewMockNetworkManager creates a new mock network manager
func NewMockNetworkManager() *MockNetworkManager {
	return &MockNetworkManager{
		networks: make(map[string]*types.NetworkConfig),
		nextID:   1,
	}
}

// CreateNetwork creates a new network
func (m *MockNetworkManager) CreateNetwork(config *types.NetworkConfig) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID
	id := "net-" + randomID(8)
	m.nextID++

	// Store network
	m.networks[id] = config

	return id, nil
}

// DestroyNetwork destroys a network
func (m *MockNetworkManager) DestroyNetwork(networkID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.networks, networkID)
	return nil
}

// GetNetwork gets a network by ID
func (m *MockNetworkManager) GetNetwork(networkID string) (*types.NetworkConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, ok := m.networks[networkID]
	if !ok {
		// Return default config
		return &types.NetworkConfig{
			VLAN:     100,
			CIDR:     "10.0.0.0/24",
			Gateway:  "10.0.0.1",
			Isolated: false,
		}, nil
	}

	return config, nil
}

// AddNetwork adds a network to the manager (for testing)
func (m *MockNetworkManager) AddNetwork(id string, config *types.NetworkConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networks[id] = config
}

// ListNetworks lists all networks
func (m *MockNetworkManager) ListNetworks() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.networks))
	for id := range m.networks {
		ids = append(ids, id)
	}

	// If no networks, return default
	if len(ids) == 0 {
		ids = append(ids, "net-default")
	}

	return ids, nil
}

// Verify MockNetworkManager implements types.NetworkManagerInterface
var _ types.NetworkManagerInterface = (*MockNetworkManager)(nil)

// randomID generates a random ID string
func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
