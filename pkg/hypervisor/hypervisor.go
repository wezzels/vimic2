// Package hypervisor provides cross-platform virtualization support
package hypervisor

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// Hypervisor is the interface for VM management
type Hypervisor interface {
	// Node lifecycle
	CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error)
	DeleteNode(ctx context.Context, id string) error
	StartNode(ctx context.Context, id string) error
	StopNode(ctx context.Context, id string) error
	RestartNode(ctx context.Context, id string) error

	// Node queries
	ListNodes(ctx context.Context) ([]*Node, error)
	GetNode(ctx context.Context, id string) (*Node, error)
	GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error)

	// Metrics
	GetMetrics(ctx context.Context, id string) (*Metrics, error)

	// Connection
	Close() error
}

// NodeState represents the state of a node
type NodeState string

const (
	NodePending NodeState = "pending"
	NodeRunning NodeState = "running"
	NodeStopped NodeState = "stopped"
	NodeError   NodeState = "error"
)

// NodeConfig holds configuration for creating a new node
type NodeConfig struct {
	Name     string `json:"name"`
	CPU      int    `json:"cpu"`
	MemoryMB uint64 `json:"memory_mb"`
	DiskGB   int    `json:"disk_gb"`
	Image    string `json:"image"`
	Network  string `json:"network"`
	SSHKey   string `json:"ssh_key"`
}

// Node represents a virtual machine node
type Node struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	State   NodeState `json:"state"`
	IP      string    `json:"ip"`
	Host    string    `json:"host"`
	Config  *NodeConfig `json:"config"`
	Created time.Time `json:"created"`
}

// NodeStatus holds current status information
type NodeStatus struct {
	State       NodeState `json:"state"`
	Uptime      time.Duration `json:"uptime"`
	CPUPercent  float64    `json:"cpu_percent"`
	MemUsed     uint64     `json:"mem_used"`
	MemTotal    uint64     `json:"mem_total"`
	DiskUsedGB  float64    `json:"disk_used_gb"`
	DiskTotalGB float64    `json:"disk_total_gb"`
	IP          string     `json:"ip"`
}

// Metrics holds resource usage metrics
type Metrics struct {
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Disk      float64   `json:"disk"`
	NetworkRX float64   `json:"network_rx"`
	NetworkTX float64   `json:"network_tx"`
	Timestamp time.Time `json:"timestamp"`
}

// HostConfig holds configuration for connecting to a hypervisor host
type HostConfig struct {
	Address    string `json:"address"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	SSHKeyPath string `json:"ssh_key_path"`
	Type       string `json:"type"` // libvirt, hyperv, apple
}

// NewHypervisor creates a hypervisor for the current platform
func NewHypervisor(cfg *HostConfig) (Hypervisor, error) {
	if cfg == nil {
		cfg = &HostConfig{}
	}

	switch cfg.Type {
	case "libvirt":
		return newLibvirtHypervisor(cfg)
	case "hyperv":
		return newWindowsHypervisor(cfg)
	case "apple":
		return newAppleHypervisor(cfg)
	case "":
		// Auto-detect based on OS
		switch runtime.GOOS {
		case "linux":
			return newLibvirtHypervisor(cfg)
		case "darwin":
			return newAppleHypervisor(cfg)
		case "windows":
			return newWindowsHypervisor(cfg)
		}
	}

	return nil, fmt.Errorf("unsupported hypervisor type: %s", cfg.Type)
}

// NewStubHypervisor creates a stub hypervisor for testing/development
func NewStubHypervisor() Hypervisor {
	return &StubHypervisor{
		nodes: make(map[string]*Node),
	}
}

// StubHypervisor is a stub implementation for testing
type StubHypervisor struct {
	nodes map[string]*Node
}

func (s *StubHypervisor) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	node := &Node{
		ID:      fmt.Sprintf("vm-%d", len(s.nodes)+1),
		Name:    cfg.Name,
		State:   NodeRunning,
		IP:      fmt.Sprintf("192.168.122.%d", len(s.nodes)+10),
		Config:  cfg,
		Created: time.Now(),
	}
	s.nodes[node.ID] = node
	return node, nil
}

func (s *StubHypervisor) DeleteNode(ctx context.Context, id string) error {
	delete(s.nodes, id)
	return nil
}

func (s *StubHypervisor) StartNode(ctx context.Context, id string) error {
	if n, ok := s.nodes[id]; ok {
		n.State = NodeRunning
	}
	return nil
}

func (s *StubHypervisor) StopNode(ctx context.Context, id string) error {
	if n, ok := s.nodes[id]; ok {
		n.State = NodeStopped
	}
	return nil
}

func (s *StubHypervisor) RestartNode(ctx context.Context, id string) error {
	if n, ok := s.nodes[id]; ok {
		n.State = NodeRunning
	}
	return nil
}

func (s *StubHypervisor) ListNodes(ctx context.Context) ([]*Node, error) {
	var result []*Node
	for _, n := range s.nodes {
		result = append(result, n)
	}
	return result, nil
}

func (s *StubHypervisor) GetNode(ctx context.Context, id string) (*Node, error) {
	return s.nodes[id], nil
}

func (s *StubHypervisor) GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error) {
	n, ok := s.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found")
	}
	return &NodeStatus{
		State:       n.State,
		Uptime:      time.Hour,
		CPUPercent:  25.0,
		MemUsed:     1024 * 1024 * 1024,
		MemTotal:    2 * 1024 * 1024 * 1024,
		DiskUsedGB:  10.0,
		DiskTotalGB: 50.0,
		IP:          n.IP,
	}, nil
}

func (s *StubHypervisor) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	return &Metrics{
		CPU:       25.0,
		Memory:    50.0,
		Disk:      30.0,
		NetworkRX: 100.0,
		NetworkTX: 50.0,
		Timestamp: time.Now(),
	}, nil
}

func (s *StubHypervisor) Close() error {
	return nil
}
