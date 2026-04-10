// Package mockovs provides mock Open vSwitch client for testing
package mockovs

import (
	"fmt"
	"sync"
)

// MockOVSClient provides a mock OVS client for testing
type MockOVSClient struct {
	bridges   map[string]*Bridge
	ports     map[string]*Port
	flows     map[string][]*Flow
	mu        sync.RWMutex
	errMode   bool
	failNext  bool
	// Behavior controls
	CommandDelay int64 // milliseconds
}

// Bridge represents an OVS bridge
type Bridge struct {
	Name    string
	VLAN    int
	Ports   []string
	Trunk   []int
}

// Port represents an OVS port
type Port struct {
	Name      string
	Bridge    string
	VLAN      int
	Trunk     []int
	QoS       int64 // Mbps
	MAC       string
	IPAddress  string
	PortSecurity bool
}

// Flow represents an OpenFlow rule
type Flow struct {
	ID        string
	Bridge    string
	Priority  int
	Match     string
	Actions   string
	Enabled   bool
}

// NewMockOVSClient creates a new mock OVS client
func NewMockOVSClient() *MockOVSClient {
	return &MockOVSClient{
		bridges: make(map[string]*Bridge),
		ports:   make(map[string]*Port),
		flows:   make(map[string][]*Flow),
	}
}

// CreateBridge creates an OVS bridge
func (m *MockOVSClient) CreateBridge(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: create bridge")
	}

	if _, exists := m.bridges[name]; exists {
		return nil // Already exists
	}

	m.bridges[name] = &Bridge{
		Name:  name,
		Ports: []string{},
	}

	return nil
}

// DeleteBridge deletes an OVS bridge
func (m *MockOVSClient) DeleteBridge(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: delete bridge")
	}

	delete(m.bridges, name)

	// Delete associated ports
	for portName, port := range m.ports {
		if port.Bridge == name {
			delete(m.ports, portName)
		}
	}

	return nil
}

// SetBridgeVLAN sets VLAN on a bridge
func (m *MockOVSClient) SetBridgeVLAN(name string, vlan int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set bridge vlan")
	}

	bridge, exists := m.bridges[name]
	if !exists {
		return fmt.Errorf("bridge not found: %s", name)
	}

	bridge.VLAN = vlan
	return nil
}

// SetBridgeTrunk sets trunk VLANs on a bridge
func (m *MockOVSClient) SetBridgeTrunk(name string, vlans []int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set bridge trunk")
	}

	bridge, exists := m.bridges[name]
	if !exists {
		return fmt.Errorf("bridge not found: %s", name)
	}

	bridge.Trunk = vlans
	return nil
}

// AddPort adds a port to a bridge
func (m *MockOVSClient) AddPort(bridgeName, portName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: add port")
	}

	bridge, exists := m.bridges[bridgeName]
	if !exists {
		return fmt.Errorf("bridge not found: %s", bridgeName)
	}

	// Add port to bridge
	bridge.Ports = append(bridge.Ports, portName)

	// Create port
	m.ports[portName] = &Port{
		Name:   portName,
		Bridge: bridgeName,
	}

	return nil
}

// DeletePort removes a port from a bridge
func (m *MockOVSClient) DeletePort(bridgeName, portName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: delete port")
	}

	bridge, exists := m.bridges[bridgeName]
	if !exists {
		return fmt.Errorf("bridge not found: %s", bridgeName)
	}

	// Remove port from bridge
	newPorts := []string{}
	for _, p := range bridge.Ports {
		if p != portName {
			newPorts = append(newPorts, p)
		}
	}
	bridge.Ports = newPorts

	// Delete port
	delete(m.ports, portName)

	return nil
}

// SetPortVLAN sets VLAN on a port
func (m *MockOVSClient) SetPortVLAN(portName string, vlan int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set port vlan")
	}

	port, exists := m.ports[portName]
	if !exists {
		return fmt.Errorf("port not found: %s", portName)
	}

	port.VLAN = vlan
	return nil
}

// SetPortTrunk sets trunk VLANs on a port
func (m *MockOVSClient) SetPortTrunk(portName string, vlans []int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set port trunk")
	}

	port, exists := m.ports[portName]
	if !exists {
		return fmt.Errorf("port not found: %s", portName)
	}

	port.Trunk = vlans
	return nil
}

// SetPortQoS sets QoS on a port
func (m *MockOVSClient) SetPortQoS(portName string, bandwidthMbps int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set port qos")
	}

	port, exists := m.ports[portName]
	if !exists {
		return fmt.Errorf("port not found: %s", portName)
	}

	port.QoS = bandwidthMbps
	return nil
}

// SetPortSecurity sets port security
func (m *MockOVSClient) SetPortSecurity(portName string, mac, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: set port security")
	}

	port, exists := m.ports[portName]
	if !exists {
		return fmt.Errorf("port not found: %s", portName)
	}

	port.MAC = mac
	port.IPAddress = ip
	port.PortSecurity = true
	return nil
}

// CreateVXLAN creates a VXLAN tunnel
func (m *MockOVSClient) CreateVXLAN(name, remoteIP string, vni int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: create vxlan")
	}

	// Create as a port with VXLAN attributes
	m.ports[name] = &Port{
		Name: name,
		VLAN: vni,
	}

	return nil
}

// CreateGRE creates a GRE tunnel
func (m *MockOVSClient) CreateGRE(name, remoteIP string, key int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: create gre")
	}

	m.ports[name] = &Port{
		Name: name,
		VLAN: key,
	}

	return nil
}

// AddFlow adds an OpenFlow rule
func (m *MockOVSClient) AddFlow(bridge string, priority int, match, actions string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: add flow")
	}

	flow := &Flow{
		ID:       fmt.Sprintf("flow-%d", len(m.flows[bridge])+1),
		Bridge:   bridge,
		Priority: priority,
		Match:    match,
		Actions:  actions,
		Enabled:  true,
	}

	m.flows[bridge] = append(m.flows[bridge], flow)
	return nil
}

// DeleteFlow deletes an OpenFlow rule
func (m *MockOVSClient) DeleteFlow(bridge, flowID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return fmt.Errorf("mock error: delete flow")
	}

	flows := m.flows[bridge]
	newFlows := []*Flow{}
	for _, f := range flows {
		if f.ID != flowID {
			newFlows = append(newFlows, f)
		}
	}
	m.flows[bridge] = newFlows

	return nil
}

// ListFlows lists OpenFlow rules
func (m *MockOVSClient) ListFlows(bridge string) ([]*Flow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: list flows")
	}

	return m.flows[bridge], nil
}

// BridgeExists checks if a bridge exists
func (m *MockOVSClient) BridgeExists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.bridges[name]
	return exists
}

// PortExists checks if a port exists
func (m *MockOVSClient) PortExists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.ports[name]
	return exists
}

// GetBridge gets a bridge by name
func (m *MockOVSClient) GetBridge(name string) (*Bridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: get bridge")
	}

	bridge, exists := m.bridges[name]
	if !exists {
		return nil, fmt.Errorf("bridge not found: %s", name)
	}
	return bridge, nil
}

// GetPort gets a port by name
func (m *MockOVSClient) GetPort(name string) (*Port, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: get port")
	}

	port, exists := m.ports[name]
	if !exists {
		return nil, fmt.Errorf("port not found: %s", name)
	}
	return port, nil
}

// ListBridges lists all bridges
func (m *MockOVSClient) ListBridges() ([]*Bridge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: list bridges")
	}

	bridges := []*Bridge{}
	for _, b := range m.bridges {
		bridges = append(bridges, b)
	}
	return bridges, nil
}

// ListPorts lists all ports
func (m *MockOVSClient) ListPorts() ([]*Port, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode || m.failNext {
		m.failNext = false
		return nil, fmt.Errorf("mock error: list ports")
	}

	ports := []*Port{}
	for _, p := range m.ports {
		ports = append(ports, p)
	}
	return ports, nil
}

// SetErrorMode enables or disables error mode
func (m *MockOVSClient) SetErrorMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errMode = enabled
}

// FailNext fails the next operation
func (m *MockOVSClient) FailNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNext = true
}

// Count returns total counts
func (m *MockOVSClient) Count() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]int{
		"bridges": len(m.bridges),
		"ports":   len(m.ports),
		"flows":   len(m.flows),
	}
}

// CreateTestBridge creates a bridge for testing
func (m *MockOVSClient) CreateTestBridge(name string) error {
	if err := m.CreateBridge(name); err != nil {
		return err
	}
	return m.SetBridgeVLAN(name, 100)
}

// CreateTestPort creates a port for testing
func (m *MockOVSClient) CreateTestPort(bridgeName, portName string) error {
	if err := m.AddPort(bridgeName, portName); err != nil {
		return err
	}
	return m.SetPortVLAN(portName, 100)
}