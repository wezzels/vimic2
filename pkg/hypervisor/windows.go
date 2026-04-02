// Package hypervisor provides cross-platform virtualization support
package hypervisor

import (
	"context"
	"fmt"
)

// WindowsHypervisor implements Hypervisor for Windows using Hyper-V
type WindowsHypervisor struct{}

func newWindowsHypervisor(cfg *HostConfig) (*WindowsHypervisor, error) {
	return &WindowsHypervisor{}, nil
}

func (h *WindowsHypervisor) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	return nil, fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) DeleteNode(ctx context.Context, id string) error {
	return fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) StartNode(ctx context.Context, id string) error {
	return fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) StopNode(ctx context.Context, id string) error {
	return fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) RestartNode(ctx context.Context, id string) error {
	return fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) ListNodes(ctx context.Context) ([]*Node, error) {
	return []*Node{}, nil
}

func (h *WindowsHypervisor) GetNode(ctx context.Context, id string) (*Node, error) {
	return nil, fmt.Errorf("node not found")
}

func (h *WindowsHypervisor) GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error) {
	return nil, fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	return nil, fmt.Errorf("Hyper-V not supported on Linux")
}

func (h *WindowsHypervisor) Close() error {
	return nil
}
