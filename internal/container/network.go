package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

// NetworkManager manages Docker networks for container isolation
type NetworkManager struct {
	client *client.Client
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(cli *client.Client) *NetworkManager {
	return &NetworkManager{client: cli}
}

// CreateNetwork creates an isolated network for a pipeline
func (nm *NetworkManager) CreateNetwork(ctx context.Context, pipelineID string, subnet string) (*NetworkInfo, error) {
	name := fmt.Sprintf("vimic2-%s", pipelineID)

	// Check if network already exists
	networks, err := nm.client.NetworkList(ctx, NetworkListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	for _, n := range networks {
		if n.Name == name {
			return &NetworkInfo{
				ID:   n.ID,
				Name: n.Name,
				Mode: "bridge",
			}, nil
		}
	}

	// Create new network
	ipam := &IPAMConfig{}
	if subnet != "" {
		ipam.Subnet = subnet
	}

	resp, err := nm.client.NetworkCreate(ctx, name, NetworkCreateOptions{
		Driver: "bridge",
		IPAM:   ipam,
		Labels: map[string]string{
			"vimic2.pipeline": pipelineID,
			"vimic2.type":     "ephemeral",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	return &NetworkInfo{
		ID:   resp.ID,
		Name: name,
		Mode: "bridge",
	}, nil
}

// GetNetwork returns network info by name
func (nm *NetworkManager) GetNetwork(ctx context.Context, name string) (*NetworkInfo, error) {
	network, err := nm.client.NetworkInspect(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("network %s not found: %w", name, err)
	}

	return &NetworkInfo{
		ID:   network.ID,
		Name: network.Name,
		Mode: network.Driver,
	}, nil
}

// RemoveNetwork removes an ephemeral network
func (nm *NetworkManager) RemoveNetwork(ctx context.Context, name string) error {
	if !strings.HasPrefix(name, "vimic2-") {
		return fmt.Errorf("refusing to remove non-vimic2 network: %s", name)
	}

	err := nm.client.NetworkRemove(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}
	return nil
}

// NetworkInfo represents Docker network information
type NetworkInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mode string `json:"mode"`
}
