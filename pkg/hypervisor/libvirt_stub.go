//go:build !libvirt
// +build !libvirt

// Package hypervisor provides cross-platform virtualization support
// This is a stub implementation for testing without libvirt
package hypervisor

import (
	"context"
	"fmt"
	"time"
)

// LibvirtHypervisorStub implements Hypervisor without requiring libvirt
type LibvirtHypervisorStub struct {
	uri     string
	storage string
}

func newLibvirtHypervisor(cfg *HostConfig) (*LibvirtHypervisorStub, error) {
	return &LibvirtHypervisorStub{
		uri:     "qemu:///system",
		storage: "/var/lib/libvirt/images",
	}, nil
}

func (h *LibvirtHypervisorStub) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("node config required")
	}
	return &Node{
		ID:      fmt.Sprintf("vm-%d", time.Now().UnixNano()),
		Name:    cfg.Name,
		State:   NodeRunning,
		IP:      fmt.Sprintf("192.168.122.%d", time.Now().Unix()%254+1),
		Host:    h.uri,
		Config:  cfg,
		Created: time.Now(),
	}, nil
}

func (h *LibvirtHypervisorStub) DeleteNode(ctx context.Context, id string) error {
	return nil
}

func (h *LibvirtHypervisorStub) StartNode(ctx context.Context, id string) error {
	return nil
}

func (h *LibvirtHypervisorStub) StopNode(ctx context.Context, id string) error {
	return nil
}

func (h *LibvirtHypervisorStub) RestartNode(ctx context.Context, id string) error {
	return nil
}

func (h *LibvirtHypervisorStub) ListNodes(ctx context.Context) ([]*Node, error) {
	return []*Node{}, nil
}

func (h *LibvirtHypervisorStub) GetNode(ctx context.Context, id string) (*Node, error) {
	return nil, fmt.Errorf("node not found")
}

func (h *LibvirtHypervisorStub) GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error) {
	return &NodeStatus{
		State:      NodeRunning,
		Uptime:     time.Hour,
		CPUPercent: 25.0,
		MemUsed:    1024 * 1024 * 1024,
		MemTotal:   2 * 1024 * 1024 * 1024,
	}, nil
}

func (h *LibvirtHypervisorStub) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	return &Metrics{
		CPU:       25.0,
		Memory:    50.0,
		Disk:      30.0,
		NetworkRX: 100.0,
		NetworkTX: 50.0,
		Timestamp: time.Now(),
	}, nil
}

func (h *LibvirtHypervisorStub) Close() error {
	return nil
}
