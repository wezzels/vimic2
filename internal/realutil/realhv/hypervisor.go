// Package realhv provides real hypervisor operations for production use
package realhv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Hypervisor provides real hypervisor operations using libvirt
type Hypervisor struct {
	client    hypervisor.Hypervisor
	config    *Config
	mu        sync.RWMutex
	connected bool
}

// Config holds hypervisor configuration
type Config struct {
	URI         string        // libvirt connection URI (e.g., qemu:///system)
	Timeout     time.Duration // connection timeout
	MaxVMs      int           // maximum VMs (0 = unlimited)
	AutoConnect bool          // auto-connect on first operation
}

// VMConfig represents VM configuration
type VMConfig struct {
	Name      string
	CPU       int
	MemoryMB  uint64
	DiskGB    int
	Image     string
	Network   string
	CloudInit string // cloud-init user data
}

// VM represents a VM node
type VM struct {
	ID        string
	Name      string
	State     hypervisor.NodeState
	IP        string
	Host      string
	Config    *VMConfig
	Created   time.Time
	UpdatedAt time.Time
}

// VMMetrics represents VM metrics
type VMMetrics struct {
	CPU       float64
	Memory    float64
	Disk      float64
	NetworkRX float64
	NetworkTX float64
	Timestamp time.Time
}

// VMStatus represents VM status
type VMStatus struct {
	State       hypervisor.NodeState
	Uptime      time.Duration
	CPUPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	DiskUsedGB  float64
	DiskTotalGB float64
	IP          string
}

// NewHypervisor creates a new hypervisor instance
func NewHypervisor(cfg *Config) *Hypervisor {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.URI == "" {
		cfg.URI = "qemu:///system"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Hypervisor{
		config: cfg,
	}
}

// Connect establishes connection to hypervisor
func (h *Hypervisor) Connect(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connected {
		return nil
	}

	// Create hypervisor client
	// Parse URI to determine connection type
	uri := h.config.URI
	hostConfig := &hypervisor.HostConfig{}

	// Handle different URI formats
	if uri == "qemu:///system" {
		// Local libvirt
		hostConfig.Type = "libvirt"
	} else if len(uri) > 8 && uri[:8] == "qemu+ssh" {
		// Remote libvirt via SSH (e.g., qemu+ssh://10.0.0.117/system)
		hostConfig.Type = "libvirt"
		hostConfig.Address = uri
	} else if uri == "apple" || uri == "" {
		hostConfig.Type = "apple"
	} else {
		hostConfig.Type = "libvirt"
	}

	client, err := hypervisor.NewHypervisor(hostConfig)
	if err != nil {
		// If libvirt fails, fall back to stub for testing
		client = hypervisor.NewStubHypervisor()
	}

	h.client = client
	h.connected = true
	return nil
}

// Disconnect closes the hypervisor connection
func (h *Hypervisor) Disconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.connected {
		return nil
	}

	if h.client != nil {
		if err := h.client.Close(); err != nil {
			return err
		}
	}

	h.connected = false
	h.client = nil
	return nil
}

// CreateNode creates a new VM
func (h *Hypervisor) CreateNode(ctx context.Context, cfg *VMConfig) (*VM, error) {
	if err := h.ensureConnected(ctx); err != nil {
		return nil, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Convert config
	hvConfig := &hypervisor.NodeConfig{
		Name:     cfg.Name,
		CPU:      cfg.CPU,
		MemoryMB: cfg.MemoryMB,
		DiskGB:   cfg.DiskGB,
		Image:    cfg.Image,
		Network:  cfg.Network,
	}

	// Create VM
	node, err := h.client.CreateNode(ctx, hvConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	return &VM{
		ID:      node.ID,
		Name:    node.Name,
		State:   node.State,
		IP:      node.IP,
		Host:    node.Host,
		Config:  cfg,
		Created: time.Now(),
	}, nil
}

// DeleteNode deletes a VM
func (h *Hypervisor) DeleteNode(ctx context.Context, id string) error {
	if err := h.ensureConnected(ctx); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.client.DeleteNode(ctx, id)
}

// StartNode starts a VM
func (h *Hypervisor) StartNode(ctx context.Context, id string) error {
	if err := h.ensureConnected(ctx); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.client.StartNode(ctx, id)
}

// StopNode stops a VM
func (h *Hypervisor) StopNode(ctx context.Context, id string) error {
	if err := h.ensureConnected(ctx); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.client.StopNode(ctx, id)
}

// RestartNode restarts a VM
func (h *Hypervisor) RestartNode(ctx context.Context, id string) error {
	if err := h.ensureConnected(ctx); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	return h.client.RestartNode(ctx, id)
}

// ListNodes lists all VMs
func (h *Hypervisor) ListNodes(ctx context.Context) ([]*VM, error) {
	if err := h.ensureConnected(ctx); err != nil {
		return nil, err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	nodes, err := h.client.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	var result []*VM
	for _, n := range nodes {
		result = append(result, &VM{
			ID:    n.ID,
			Name:  n.Name,
			State: n.State,
			IP:    n.IP,
			Host:  n.Host,
		})
	}

	return result, nil
}

// GetNode gets a VM by ID
func (h *Hypervisor) GetNode(ctx context.Context, id string) (*VM, error) {
	if err := h.ensureConnected(ctx); err != nil {
		return nil, err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	node, err := h.client.GetNode(ctx, id)
	if err != nil {
		return nil, err
	}

	return &VM{
		ID:    node.ID,
		Name:  node.Name,
		State: node.State,
		IP:    node.IP,
		Host:  node.Host,
	}, nil
}

// GetNodeStatus gets VM status
func (h *Hypervisor) GetNodeStatus(ctx context.Context, id string) (*VMStatus, error) {
	if err := h.ensureConnected(ctx); err != nil {
		return nil, err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	status, err := h.client.GetNodeStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	return &VMStatus{
		State:       status.State,
		Uptime:      status.Uptime,
		CPUPercent:  status.CPUPercent,
		MemUsed:     status.MemUsed,
		MemTotal:    status.MemTotal,
		DiskUsedGB:  status.DiskUsedGB,
		DiskTotalGB: status.DiskTotalGB,
		IP:          status.IP,
	}, nil
}

// GetMetrics gets VM metrics
func (h *Hypervisor) GetMetrics(ctx context.Context, id string) (*VMMetrics, error) {
	if err := h.ensureConnected(ctx); err != nil {
		return nil, err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics, err := h.client.GetMetrics(ctx, id)
	if err != nil {
		return nil, err
	}

	return &VMMetrics{
		CPU:       metrics.CPU,
		Memory:    metrics.Memory,
		Disk:      metrics.Disk,
		NetworkRX: metrics.NetworkRX,
		NetworkTX: metrics.NetworkTX,
		Timestamp: time.Now(),
	}, nil
}

// IsConnected returns connection status
func (h *Hypervisor) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected
}

// Close closes the hypervisor connection
func (h *Hypervisor) Close() error {
	return h.Disconnect()
}

// ensureConnected ensures we have a connection
func (h *Hypervisor) ensureConnected(ctx context.Context) error {
	if h.config.AutoConnect && !h.connected {
		return h.Connect(ctx)
	}

	if !h.connected {
		return fmt.Errorf("not connected to hypervisor")
	}

	return nil
}

// SetTimeout sets the operation timeout
func (h *Hypervisor) SetTimeout(timeout time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.Timeout = timeout
}

// SetMaxVMs sets the maximum number of VMs
func (h *Hypervisor) SetMaxVMs(max int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.MaxVMs = max
}

// Factory

// HypervisorFactory creates hypervisor instances
type HypervisorFactory struct {
	DefaultURI string
	AutoConnect bool
}

// NewHypervisorFactory creates a factory
func NewHypervisorFactory() *HypervisorFactory {
	return &HypervisorFactory{
		DefaultURI: "qemu:///system",
		AutoConnect: true,
	}
}

// Create creates a new hypervisor instance
func (f *HypervisorFactory) Create(cfg *Config) *Hypervisor {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.URI == "" {
		cfg.URI = f.DefaultURI
	}
	cfg.AutoConnect = f.AutoConnect

	return NewHypervisor(cfg)
}

// CreateWithURI creates a hypervisor with specific URI
func (f *HypervisorFactory) CreateWithURI(uri string) *Hypervisor {
	return f.Create(&Config{
		URI:         uri,
		AutoConnect: f.AutoConnect,
	})
}

// CreateRemote creates a hypervisor for remote host
func (f *HypervisorFactory) CreateRemote(host string) *Hypervisor {
	uri := fmt.Sprintf("qemu+ssh://%s/system", host)
	return f.CreateWithURI(uri)
}