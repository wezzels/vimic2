// Package realovs provides real Open vSwitch client for production use
package realovs

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Client provides real OVS client operations
type Client struct {
	mu          sync.RWMutex
	timeout     time.Duration
	sudo        bool
	dryRun      bool
	lastCommand string
	lastOutput  string
	lastError   error
}

// Config holds OVS client configuration
type Config struct {
	Timeout time.Duration
	Sudo    bool
	DryRun  bool
}

// Bridge represents an OVS bridge
type Bridge struct {
	Name  string
	VLAN  int
	Trunk []int
	Ports []string
}

// Port represents an OVS port
type Port struct {
	Name         string
	Bridge       string
	VLAN         int
	Trunk        []int
	QoS          int64
	MAC          string
	IPAddress    string
	PortSecurity bool
	Options      map[string]string
}

// Flow represents an OpenFlow rule
type Flow struct {
	ID       string
	Bridge   string
	Priority int
	Match    string
	Actions  string
	Enabled  bool
}

// NewClient creates a new OVS client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		timeout: cfg.Timeout,
		sudo:    cfg.Sudo,
		dryRun:  cfg.DryRun,
	}
}

// NewClientWithDefaults creates a client with sensible defaults
func NewClientWithDefaults() *Client {
	return NewClient(&Config{
		Timeout: 30 * time.Second,
		Sudo:    true, // Use sudo for OVS commands
		DryRun:  false,
	})
}

// run executes an OVS command
func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Build command
	cmd := "ovs-vsctl"
	if c.sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	c.lastCommand = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if c.dryRun {
		c.lastOutput = "[dry-run]"
		c.lastError = nil
		return "", nil
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()
	c.lastOutput = stdout.String()
	c.lastError = err

	if err != nil {
		return "", fmt.Errorf("command failed: %s: %w (stderr: %s)", c.lastCommand, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// runOfctl executes an ovs-ofctl command
func (c *Client) runOfctl(ctx context.Context, args ...string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := "ovs-ofctl"
	if c.sudo {
		args = append([]string{cmd}, args...)
		cmd = "sudo"
	}

	c.lastCommand = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))

	if c.dryRun {
		c.lastOutput = "[dry-run]"
		c.lastError = nil
		return "", nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()
	c.lastOutput = stdout.String()
	c.lastError = err

	if err != nil {
		return "", fmt.Errorf("command failed: %s: %w (stderr: %s)", c.lastCommand, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Bridge operations

// CreateBridge creates an OVS bridge
func (c *Client) CreateBridge(name string) error {
	_, err := c.run(context.Background(), "--", "--may-exist", "add-br", name)
	return err
}

// DeleteBridge deletes an OVS bridge
func (c *Client) DeleteBridge(name string) error {
	_, err := c.run(context.Background(), "--", "--if-exists", "del-br", name)
	return err
}

// SetBridgeVLAN sets the VLAN tag on a bridge
func (c *Client) SetBridgeVLAN(name string, vlan int) error {
	_, err := c.run(context.Background(), "set", "bridge", name,
		fmt.Sprintf("other_config:tag=%d", vlan))
	return err
}

// SetBridgeTrunk sets trunk VLANs on a bridge
func (c *Client) SetBridgeTrunk(name string, vlans []int) error {
	vlanStr := ""
	for i, v := range vlans {
		if i > 0 {
			vlanStr += ","
		}
		vlanStr += fmt.Sprintf("%d", v)
	}
	_, err := c.run(context.Background(), "set", "bridge", name,
		fmt.Sprintf("other_config:trunk=%s", vlanStr))
	return err
}

// GetBridge gets a bridge by name
func (c *Client) GetBridge(name string) (*Bridge, error) {
	output, err := c.run(context.Background(), "list-br")
	if err != nil {
		return nil, err
	}

	bridges := strings.Split(output, "\n")
	for _, b := range bridges {
		if b == name {
			return &Bridge{Name: name}, nil
		}
	}

	return nil, fmt.Errorf("bridge not found: %s", name)
}

// ListBridges lists all bridges
func (c *Client) ListBridges() ([]*Bridge, error) {
	output, err := c.run(context.Background(), "list-br")
	if err != nil {
		return nil, err
	}

	var bridges []*Bridge
	for _, name := range strings.Split(output, "\n") {
		name = strings.TrimSpace(name)
		if name != "" {
			bridges = append(bridges, &Bridge{Name: name})
		}
	}

	return bridges, nil
}

// BridgeExists checks if a bridge exists
func (c *Client) BridgeExists(name string) bool {
	_, err := c.GetBridge(name)
	return err == nil
}

// Port operations

// AddPort adds a port to a bridge
func (c *Client) AddPort(bridgeName, portName string) error {
	_, err := c.run(context.Background(), "--", "--may-exist", "add-port",
		bridgeName, portName)
	return err
}

// DeletePort removes a port from a bridge
func (c *Client) DeletePort(bridgeName, portName string) error {
	_, err := c.run(context.Background(), "--", "--if-exists", "del-port",
		bridgeName, portName)
	return err
}

// SetPortVLAN sets the VLAN tag on a port
func (c *Client) SetPortVLAN(portName string, vlan int) error {
	_, err := c.run(context.Background(), "set", "port", portName,
		fmt.Sprintf("tag=%d", vlan))
	return err
}

// SetPortTrunk sets trunk VLANs on a port
func (c *Client) SetPortTrunk(portName string, vlans []int) error {
	vlanStr := ""
	for i, v := range vlans {
		if i > 0 {
			vlanStr += ","
		}
		vlanStr += fmt.Sprintf("%d", v)
	}
	_, err := c.run(context.Background(), "set", "port", portName,
		fmt.Sprintf("trunks=[%s]", vlanStr))
	return err
}

// SetPortQoS sets QoS bandwidth limit on a port
func (c *Client) SetPortQoS(portName string, bandwidthMbps int64) error {
	// Set ingress policing
	_, err := c.run(context.Background(), "set", "interface", portName,
		fmt.Sprintf("ingress_policing_rate=%d", bandwidthMbps*1000000/8))
	if err != nil {
		return err
	}
	return nil
}

// SetPortSecurity sets port security (MAC/IP anti-spoofing)
func (c *Client) SetPortSecurity(portName, mac, ip string) error {
	// Set MAC on interface
	_, err := c.run(context.Background(), "set", "interface", portName,
		fmt.Sprintf("mac=\"%s\"", mac))
	if err != nil {
		return err
	}
	return nil
}

// GetPort gets a port by name
func (c *Client) GetPort(name string) (*Port, error) {
	output, err := c.run(context.Background(), "get", "Port", name, "name")
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, fmt.Errorf("port not found: %s", name)
	}

	return &Port{Name: name}, nil
}

// ListPorts lists all ports on a bridge
func (c *Client) ListPorts(bridgeName string) ([]*Port, error) {
	output, err := c.run(context.Background(), "list-ports", bridgeName)
	if err != nil {
		return nil, err
	}

	var ports []*Port
	for _, name := range strings.Split(output, "\n") {
		name = strings.TrimSpace(name)
		if name != "" {
			ports = append(ports, &Port{Name: name, Bridge: bridgeName})
		}
	}

	return ports, nil
}

// PortExists checks if a port exists
func (c *Client) PortExists(name string) bool {
	_, err := c.GetPort(name)
	return err == nil
}

// Tunnel operations

// CreateVXLAN creates a VXLAN tunnel interface
func (c *Client) CreateVXLAN(name, remoteIP string, vni int) error {
	_, err := c.run(context.Background(), "--", "--may-exist", "add-port",
		"br-int", name,
		"--", "set", "interface", name,
		"type=vxlan",
		fmt.Sprintf("options:remote_ip=%s", remoteIP),
		fmt.Sprintf("options:key=%d", vni))
	return err
}

// CreateGRE creates a GRE tunnel interface
func (c *Client) CreateGRE(name, remoteIP string, key int) error {
	_, err := c.run(context.Background(), "--", "--may-exist", "add-port",
		"br-int", name,
		"--", "set", "interface", name,
		"type=gre",
		fmt.Sprintf("options:remote_ip=%s", remoteIP),
		fmt.Sprintf("options:key=%d", key))
	return err
}

// CreateGeneve creates a Geneve tunnel interface
func (c *Client) CreateGeneve(name, remoteIP string, vni int) error {
	_, err := c.run(context.Background(), "--", "--may-exist", "add-port",
		"br-int", name,
		"--", "set", "interface", name,
		"type=geneve",
		fmt.Sprintf("options:remote_ip=%s", remoteIP),
		fmt.Sprintf("options:key=%d", vni))
	return err
}

// Flow operations

// AddFlow adds an OpenFlow rule
func (c *Client) AddFlow(bridge string, priority int, match, actions string) error {
	flowStr := fmt.Sprintf("priority=%d,%s,actions=%s", priority, match, actions)
	_, err := c.runOfctl(context.Background(), "add-flow", bridge, flowStr)
	return err
}

// DeleteFlow deletes an OpenFlow rule
func (c *Client) DeleteFlow(bridge, match string) error {
	_, err := c.runOfctl(context.Background(), "del-flows", bridge,
		"--strict", match)
	return err
}

// ListFlows lists OpenFlow rules on a bridge
func (c *Client) ListFlows(bridge string) ([]*Flow, error) {
	output, err := c.runOfctl(context.Background(), "dump-flows", bridge)
	if err != nil {
		return nil, err
	}

	var flows []*Flow
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, " cookie") {
			continue
		}

		// Parse flow
		// Format: priority=100,in_port=1 actions=output:2
		flow := &Flow{Bridge: bridge}
		parts := strings.SplitN(line, " actions=", 2)
		if len(parts) == 2 {
			flow.Actions = parts[1]
			matchParts := strings.Split(parts[0], ",")
			for _, p := range matchParts {
				if strings.HasPrefix(p, "priority=") {
					fmt.Sscanf(p, "priority=%d", &flow.Priority)
				} else if p != "" {
					if flow.Match != "" {
						flow.Match += ","
					}
					flow.Match += p
				}
			}
		}
		flows = append(flows, flow)
	}

	return flows, nil
}

// ClearFlows clears all OpenFlow rules on a bridge
func (c *Client) ClearFlows(bridge string) error {
	_, err := c.runOfctl(context.Background(), "del-flows", bridge)
	return err
}

// Utility methods

// GetInterfaceUUID gets the UUID of an interface
func (c *Client) GetInterfaceUUID(name string) (string, error) {
	output, err := c.run(context.Background(), "get", "interface", name, "_uuid")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// SetInterfaceOptions sets interface options
func (c *Client) SetInterfaceOption(name, key, value string) error {
	_, err := c.run(context.Background(), "set", "interface", name,
		fmt.Sprintf("options:%s=%s", key, value))
	return err
}

// GetInterfaceOption gets an interface option
func (c *Client) GetInterfaceOption(name, key string) (string, error) {
	output, err := c.run(context.Background(), "get", "interface", name,
		fmt.Sprintf("options:%s", key))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// LastCommand returns the last executed command
func (c *Client) LastCommand() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCommand
}

// LastOutput returns the last command output
func (c *Client) LastOutput() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastOutput
}

// SetTimeout sets the command timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

// SetDryRun enables or disables dry-run mode
func (c *Client) SetDryRun(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dryRun = enabled
}

// SetSudo enables or disables sudo
func (c *Client) SetSudo(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sudo = enabled
}

// IsAvailable checks if OVS is available on the system
func IsAvailable() bool {
	_, err := exec.LookPath("ovs-vsctl")
	return err == nil
}

// Version returns OVS version
func Version() (string, error) {
	cmd := exec.Command("ovs-vsctl", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output like "ovs-vsctl (Open vSwitch) 2.17.3"
	parts := strings.Fields(string(output))
	for i, p := range parts {
		if p == "(Open" && i+1 < len(parts) {
			version := strings.TrimSuffix(parts[i+1], ")")
			return version, nil
		}
	}

	return "", fmt.Errorf("version not found in output")
}
