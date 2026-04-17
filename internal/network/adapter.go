package network

import (
	"context"
	"fmt"

	"github.com/stsgym/vimic2/internal/types"
)

// NetworkManagerAdapter wraps NetworkManager to implement types.NetworkManagerInterface.
type NetworkManagerAdapter struct {
	*NetworkManager
}

// NewNetworkManagerAdapter creates an adapter
func NewNetworkManagerAdapter(nm *NetworkManager) *NetworkManagerAdapter {
	return &NetworkManagerAdapter{NetworkManager: nm}
}

// CreateNetwork creates a network from a types.NetworkConfig
func (a *NetworkManagerAdapter) CreateNetwork(config *types.NetworkConfig) (string, error) {
	net := &Network{
		Name:       config.CIDR,
		VLANID:     config.VLAN,
		CIDR:       config.CIDR,
		Gateway:    config.Gateway,
		DNS:        config.DNS,
		BridgeName: "",
	}
	if config.Isolated {
		net.Type = NetworkTypeVLAN
	} else {
		net.Type = NetworkTypeBridge
	}
	if err := a.NetworkManager.CreateNetwork(context.Background(), net); err != nil {
		return "", err
	}
	return net.ID, nil
}

// DestroyNetwork destroys a network by ID
func (a *NetworkManagerAdapter) DestroyNetwork(networkID string) error {
	// NetworkManager doesn't have DestroyNetwork yet
	return fmt.Errorf("destroy network not yet implemented")
}

// GetNetwork returns a NetworkConfig for the given network ID
func (a *NetworkManagerAdapter) GetNetwork(networkID string) (*types.NetworkConfig, error) {
	// NetworkManager doesn't have GetNetwork by ID yet; use ListNetworks and filter
	networks, err := a.NetworkManager.ListNetworks(context.Background())
	if err != nil {
		return nil, err
	}
	for _, n := range networks {
		if n.ID == networkID {
			return &types.NetworkConfig{
				VLAN:     n.VLANID,
				CIDR:     n.CIDR,
				Gateway:  n.Gateway,
				DNS:      n.DNS,
				Isolated: n.Type == NetworkTypeVLAN,
			}, nil
		}
	}
	return nil, fmt.Errorf("network not found: %s", networkID)
}

// Verify interface satisfaction
var _ types.NetworkManagerInterface = (*NetworkManagerAdapter)(nil)