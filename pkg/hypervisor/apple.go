// Package hypervisor provides cross-platform virtualization support
package hypervisor

import (
	"context"
	"fmt"
)

// AppleHypervisor implements Hypervisor for macOS using Apple Virtualization Framework
type AppleHypervisor struct{}

func newAppleHypervisor(cfg *HostConfig) (*AppleHypervisor, error) {
	return &AppleHypervisor{}, nil
}

func (h *AppleHypervisor) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	return nil, fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) DeleteNode(ctx context.Context, id string) error {
	return fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) StartNode(ctx context.Context, id string) error {
	return fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) StopNode(ctx context.Context, id string) error {
	return fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) RestartNode(ctx context.Context, id string) error {
	return fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) ListNodes(ctx context.Context) ([]*Node, error) {
	return []*Node{}, nil
}

func (h *AppleHypervisor) GetNode(ctx context.Context, id string) (*Node, error) {
	return nil, fmt.Errorf("node not found")
}

func (h *AppleHypervisor) GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error) {
	return nil, fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	return nil, fmt.Errorf("Apple Virtualization not supported on Linux")
}

func (h *AppleHypervisor) Close() error {
	return nil
}
