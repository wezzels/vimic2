// Package mockhv provides mock hypervisor for testing
package mockhv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// MockHypervisor provides a mock hypervisor for testing
type MockHypervisor struct {
	nodes      map[string]*hypervisor.Node
	metrics    map[string]*hypervisor.Metrics
	mu         sync.RWMutex
	errMode    bool
	failNextOp bool
	// Behavior controls
	CreateDelay time.Duration
	DeleteDelay time.Duration
	StartDelay  time.Duration
	StopDelay   time.Duration
}

// NewMockHypervisor creates a new mock hypervisor
func NewMockHypervisor() *MockHypervisor {
	return &MockHypervisor{
		nodes:   make(map[string]*hypervisor.Node),
		metrics: make(map[string]*hypervisor.Metrics),
	}
}

// CreateNode creates a new node
func (m *MockHypervisor) CreateNode(ctx context.Context, cfg *hypervisor.NodeConfig) (*hypervisor.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, fmt.Errorf("mock error: create node")
	}

	// Simulate delay
	if m.CreateDelay > 0 {
		time.Sleep(m.CreateDelay)
	}

	node := &hypervisor.Node{
		ID:      fmt.Sprintf("vm-%d", time.Now().UnixNano()),
		Name:    cfg.Name,
		State:   hypervisor.NodeRunning,
		IP:      fmt.Sprintf("192.168.122.%d", time.Now().Unix()%254+1),
		Host:    "mock",
		Config:  cfg,
		Created: time.Now(),
	}

	m.nodes[node.ID] = node

	// Initialize metrics
	m.metrics[node.ID] = &hypervisor.Metrics{
		CPU:       25.0,
		Memory:    30.0,
		Disk:      40.0,
		NetworkRX: 1024000,
		NetworkTX: 512000,
		Timestamp: time.Now(),
	}

	return node, nil
}

// DeleteNode deletes a node
func (m *MockHypervisor) DeleteNode(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return fmt.Errorf("mock error: delete node")
	}

	// Simulate delay
	if m.DeleteDelay > 0 {
		time.Sleep(m.DeleteDelay)
	}

	delete(m.nodes, id)
	delete(m.metrics, id)
	return nil
}

// StartNode starts a node
func (m *MockHypervisor) StartNode(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return fmt.Errorf("mock error: start node")
	}

	// Simulate delay
	if m.StartDelay > 0 {
		time.Sleep(m.StartDelay)
	}

	node, ok := m.nodes[id]
	if !ok {
		return fmt.Errorf("node not found: %s", id)
	}

	node.State = hypervisor.NodeRunning
	return nil
}

// StopNode stops a node
func (m *MockHypervisor) StopNode(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return fmt.Errorf("mock error: stop node")
	}

	// Simulate delay
	if m.StopDelay > 0 {
		time.Sleep(m.StopDelay)
	}

	node, ok := m.nodes[id]
	if !ok {
		return fmt.Errorf("node not found: %s", id)
	}

	node.State = hypervisor.NodeStopped
	return nil
}

// RestartNode restarts a node
func (m *MockHypervisor) RestartNode(ctx context.Context, id string) error {
	if err := m.StopNode(ctx, id); err != nil {
		return err
	}
	return m.StartNode(ctx, id)
}

// ListNodes lists all nodes
func (m *MockHypervisor) ListNodes(ctx context.Context) ([]*hypervisor.Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, fmt.Errorf("mock error: list nodes")
	}

	var nodes []*hypervisor.Node
	for _, n := range m.nodes {
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// GetNode gets a node by ID
func (m *MockHypervisor) GetNode(ctx context.Context, id string) (*hypervisor.Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, fmt.Errorf("mock error: get node")
	}

	node, ok := m.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

// GetNodeStatus gets node status
func (m *MockHypervisor) GetNodeStatus(ctx context.Context, id string) (*hypervisor.NodeStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, fmt.Errorf("mock error: get status")
	}

	node, ok := m.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	return &hypervisor.NodeStatus{
		State:       node.State,
		Uptime:      time.Since(node.Created),
		CPUPercent:  25.0,
		MemUsed:     2048,
		MemTotal:    8192,
		DiskUsedGB:  10.0,
		DiskTotalGB: 50.0,
		IP:          node.IP,
	}, nil
}

// GetMetrics gets node metrics
func (m *MockHypervisor) GetMetrics(ctx context.Context, id string) (*hypervisor.Metrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, fmt.Errorf("mock error: get metrics")
	}

	metrics, ok := m.metrics[id]
	if !ok {
		return nil, fmt.Errorf("metrics not found: %s", id)
	}

	// Update timestamp
	metrics.Timestamp = time.Now()
	return metrics, nil
}

// Close closes the hypervisor
func (m *MockHypervisor) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodes = make(map[string]*hypervisor.Node)
	m.metrics = make(map[string]*hypervisor.Metrics)
	return nil
}

// SetErrorMode enables or disables error mode
func (m *MockHypervisor) SetErrorMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errMode = enabled
}

// FailNext fails the next operation
func (m *MockHypervisor) FailNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNextOp = true
}

// SetNodeState sets a node's state directly
func (m *MockHypervisor) SetNodeState(id string, state hypervisor.NodeState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	if !ok {
		return fmt.Errorf("node not found: %s", id)
	}

	node.State = state
	return nil
}

// SetNodeIP sets a node's IP directly
func (m *MockHypervisor) SetNodeIP(id string, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	if !ok {
		return fmt.Errorf("node not found: %s", id)
	}

	node.IP = ip
	return nil
}

// SetMetrics sets node metrics directly
func (m *MockHypervisor) SetMetrics(id string, metrics *hypervisor.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[id] = metrics
	return nil
}

// Count returns total node count
func (m *MockHypervisor) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.nodes)
}

// CountByState returns node count by state
func (m *MockHypervisor) CountByState(state hypervisor.NodeState) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, node := range m.nodes {
		if node.State == state {
			count++
		}
	}
	return count
}

// GetNodeIDs returns all node IDs
func (m *MockHypervisor) GetNodeIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.nodes))
	for id := range m.nodes {
		ids = append(ids, id)
	}
	return ids
}

// MockHypervisorFactory creates mock hypervisors
type MockHypervisorFactory struct {
	DefaultDelay time.Duration
	ErrorMode    bool
}

// NewMockHypervisorFactory creates a new factory
func NewMockHypervisorFactory() *MockHypervisorFactory {
	return &MockHypervisorFactory{}
}

// Create creates a new mock hypervisor
func (f *MockHypervisorFactory) Create(cfg *hypervisor.HostConfig) (hypervisor.Hypervisor, error) {
	m := NewMockHypervisor()
	if f.DefaultDelay > 0 {
		m.CreateDelay = f.DefaultDelay
		m.DeleteDelay = f.DefaultDelay
		m.StartDelay = f.DefaultDelay
		m.StopDelay = f.DefaultDelay
	}
	m.errMode = f.ErrorMode
	return m, nil
}

// CreateWithNodes creates a mock hypervisor with pre-existing nodes
func CreateWithNodes(nodes []*hypervisor.NodeConfig) (*MockHypervisor, error) {
	m := NewMockHypervisor()
	ctx := context.Background()

	for _, cfg := range nodes {
		_, err := m.CreateNode(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}