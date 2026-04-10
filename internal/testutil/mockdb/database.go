// Package mockdb provides mock database for testing
package mockdb

import (
	"errors"
	"sync"
	"time"
)

// MockDB provides a mock database for testing
type MockDB struct {
	clusters    map[string]*Cluster
	hosts       map[string]*Host
	nodes       map[string]*Node
	pipelines   map[string]*Pipeline
	runners     map[string]*Runner
	networks    map[string]*Network
	pools       map[string]*Pool
	metrics     map[string][]*Metric
	alerts      map[string]*Alert
	mu          sync.RWMutex
	errMode     bool // If true, return errors
	failNextOp  bool // Fail the next operation
}

// Cluster represents a cluster record
type Cluster struct {
	ID        string
	Name      string
	Status    string
	Config    *ClusterConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	MinNodes  int
	MaxNodes  int
	AutoScale bool
}

// Host represents a host record
type Host struct {
	ID         string
	Name       string
	Address    string
	Port       int
	User       string
	SSHKeyPath string
	HVType     string
	CreatedAt  time.Time
}

// Node represents a node record
type Node struct {
	ID        string
	ClusterID string
	HostID    string
	Name      string
	Role      string
	State     string
	IP        string
	Config    *NodeConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NodeConfig represents node configuration
type NodeConfig struct {
	CPU    int
	Memory int
	Disk   int
	Image  string
}

// Pipeline represents a pipeline record
type Pipeline struct {
	ID     string
	Status string
	Config map[string]interface{}
}

// Runner represents a runner record
type Runner struct {
	ID     string
	Status string
	Config map[string]interface{}
}

// Network represents a network record
type Network struct {
	ID     string
	Status string
	Config map[string]interface{}
}

// Pool represents a pool record
type Pool struct {
	ID        string
	Name      string
	Available int
	Busy      int
}

// Metric represents a metric record
type Metric struct {
	NodeID     string
	CPU        float64
	Memory     float64
	Disk       float64
	NetworkRX  float64
	NetworkTX  float64
	RecordedAt time.Time
}

// Alert represents an alert record
type Alert struct {
	ID        string
	NodeID    string
	Type      string
	Message   string
	Severity  string
	CreatedAt time.Time
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		clusters:  make(map[string]*Cluster),
		hosts:     make(map[string]*Host),
		nodes:     make(map[string]*Node),
		pipelines: make(map[string]*Pipeline),
		runners:   make(map[string]*Runner),
		networks:  make(map[string]*Network),
		pools:     make(map[string]*Pool),
		metrics:   make(map[string][]*Metric),
		alerts:    make(map[string]*Alert),
	}
}

// SetErrorMode enables or disables error mode
func (m *MockDB) SetErrorMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errMode = enabled
}

// FailNext fails the next operation
func (m *MockDB) FailNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNextOp = true
}

// Cluster operations

func (m *MockDB) SaveCluster(cluster *Cluster) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save cluster")
	}

	m.clusters[cluster.ID] = cluster
	return nil
}

func (m *MockDB) GetCluster(id string) (*Cluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get cluster")
	}

	cluster, ok := m.clusters[id]
	if !ok {
		return nil, errors.New("cluster not found")
	}
	return cluster, nil
}

func (m *MockDB) ListClusters() ([]*Cluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: list clusters")
	}

	var clusters []*Cluster
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}
	return clusters, nil
}

func (m *MockDB) DeleteCluster(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: delete cluster")
	}

	delete(m.clusters, id)
	return nil
}

// Host operations

func (m *MockDB) SaveHost(host *Host) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save host")
	}

	m.hosts[host.ID] = host
	return nil
}

func (m *MockDB) GetHost(id string) (*Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get host")
	}

	host, ok := m.hosts[id]
	if !ok {
		return nil, errors.New("host not found")
	}
	return host, nil
}

func (m *MockDB) ListHosts() ([]*Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: list hosts")
	}

	var hosts []*Host
	for _, h := range m.hosts {
		hosts = append(hosts, h)
	}
	return hosts, nil
}

func (m *MockDB) DeleteHost(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: delete host")
	}

	delete(m.hosts, id)
	return nil
}

// Node operations

func (m *MockDB) SaveNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save node")
	}

	m.nodes[node.ID] = node
	return nil
}

func (m *MockDB) GetNode(id string) (*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get node")
	}

	node, ok := m.nodes[id]
	if !ok {
		return nil, errors.New("node not found")
	}
	return node, nil
}

func (m *MockDB) ListClusterNodes(clusterID string) ([]*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: list cluster nodes")
	}

	var nodes []*Node
	for _, n := range m.nodes {
		if n.ClusterID == clusterID {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (m *MockDB) DeleteNode(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: delete node")
	}

	delete(m.nodes, id)
	return nil
}

// Metric operations

func (m *MockDB) SaveMetric(metric *Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save metric")
	}

	m.metrics[metric.NodeID] = append(m.metrics[metric.NodeID], metric)
	return nil
}

func (m *MockDB) GetNodeMetrics(nodeID string, since time.Time) ([]*Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get metrics")
	}

	var metrics []*Metric
	for _, m := range m.metrics[nodeID] {
		if m.RecordedAt.After(since) {
			metrics = append(metrics, m)
		}
	}
	return metrics, nil
}

// Alert operations

func (m *MockDB) SaveAlert(alert *Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save alert")
	}

	m.alerts[alert.ID] = alert
	return nil
}

func (m *MockDB) GetAlert(id string) (*Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get alert")
	}

	alert, ok := m.alerts[id]
	if !ok {
		return nil, errors.New("alert not found")
	}
	return alert, nil
}

func (m *MockDB) ListAlerts() ([]*Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: list alerts")
	}

	var alerts []*Alert
	for _, a := range m.alerts {
		alerts = append(alerts, a)
	}
	return alerts, nil
}

// Pool operations

func (m *MockDB) SavePool(pool *Pool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return errors.New("mock error: save pool")
	}

	m.pools[pool.ID] = pool
	return nil
}

func (m *MockDB) GetPool(id string) (*Pool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNextOp {
		m.failNextOp = false
		return nil, errors.New("mock error: get pool")
	}

	pool, ok := m.pools[id]
	if !ok {
		return nil, errors.New("pool not found")
	}
	return pool, nil
}

// Count returns total counts for testing
func (m *MockDB) Count() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]int{
		"clusters": len(m.clusters),
		"hosts":    len(m.hosts),
		"nodes":    len(m.nodes),
		"alerts":   len(m.alerts),
		"pools":    len(m.pools),
	}
}