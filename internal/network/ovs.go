// Package network provides Open vSwitch client for network management
package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// OVSClient provides Open vSwitch operations
type OVSClient struct {
	vswitchdPath string
	ofctlPath    string
	ipPath       string
	ip2netnsPath string
	iptablesPath string
}

// NewOVSClient creates a new OVS client
func NewOVSClient() *OVSClient {
	return &OVSClient{
		vswitchdPath: "ovs-vsctl",
		ofctlPath:    "ovs-ofctl",
		ipPath:       "ip",
		iptablesPath: "iptables",
	}
}

// CreateBridge creates an OVS bridge
func (o *OVSClient) CreateBridge(name string) error {
	// Check if bridge already exists
	if o.bridgeExists(name) {
		return nil
	}

	// Create the bridge
	cmd := exec.Command(o.vswitchdPath, "--", "add-br", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge %s: %w", name, err)
	}

	// Bring up the bridge
	if err := o.upInterface(name); err != nil {
		return fmt.Errorf("failed to bring up bridge %s: %w", name, err)
	}

	return nil
}

// DeleteBridge deletes an OVS bridge
func (o *OVSClient) DeleteBridge(name string) error {
	cmd := exec.Command(o.vswitchdPath, "--", "del-br", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete bridge %s: %w", name, err)
	}
	return nil
}

// SetBridgeVLAN sets a single VLAN on a bridge
func (o *OVSClient) SetBridgeVLAN(name string, vlanID int) error {
	// Set the bridge as an access port with VLAN tag
	cmd := exec.Command(o.vswitchdPath, "--", "set", "bridge", name,
		fmt.Sprintf("other_config:vlan-default-nid=%d", vlanID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set VLAN %d on bridge %s: %w", vlanID, name, err)
	}
	return nil
}

// SetBridgeTrunk sets multiple VLANs on a bridge (trunk mode)
func (o *OVSClient) SetBridgeTrunk(name string, vlans []int) error {
	// Convert VLANs to string list
	vlanStrs := make([]string, len(vlans))
	for i, v := range vlans {
		vlanStrs[i] = strconv.Itoa(v)
	}

	// Set as trunk port
	cmd := exec.Command(o.vswitchdPath, "--", "set", "bridge", name,
		fmt.Sprintf("other_config:trunk=%s", strings.Join(vlanStrs, ",")))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set trunk VLANs on bridge %s: %w", name, err)
	}
	return nil
}

// CreatePort creates a port on a bridge
func (o *OVSClient) CreatePort(bridgeName, portName string) error {
	cmd := exec.Command(o.vswitchdPath, "--", "add-port", bridgeName, portName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create port %s on bridge %s: %w", portName, bridgeName, err)
	}
	return nil
}

// DeletePort deletes a port from a bridge
func (o *OVSClient) DeletePort(bridgeName, portName string) error {
	cmd := exec.Command(o.vswitchdPath, "--", "del-port", bridgeName, portName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete port %s from bridge %s: %w", portName, bridgeName, err)
	}
	return nil
}

// SetPortVLAN sets a VLAN tag on a port
func (o *OVSClient) SetPortVLAN(portName string, vlanID int) error {
	cmd := exec.Command(o.vswitchdPath, "--", "set", "port", portName,
		fmt.Sprintf("tag=%d", vlanID))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set VLAN %d on port %s: %w", vlanID, portName, err)
	}
	return nil
}

// SetPortTrunk sets trunk VLANs on a port
func (o *OVSClient) SetPortTrunk(portName string, vlans []int) error {
	vlanStrs := make([]string, len(vlans))
	for i, v := range vlans {
		vlanStrs[i] = strconv.Itoa(v)
	}

	cmd := exec.Command(o.vswitchdPath, "--", "set", "port", portName,
		fmt.Sprintf("trunks=%s", strings.Join(vlanStrs, ",")))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set trunk VLANs on port %s: %w", portName, err)
	}
	return nil
}

// SetPortQoS sets QoS rate limiting on a port
func (o *OVSClient) SetPortQoS(portName string, bandwidthMbps int64) error {
	// Convert Mbps to Kbps
	bandwidthKbps := bandwidthMbps * 1000

	// Set ingress policing
	cmd := exec.Command(o.vswitchdPath, "--", "set", "interface", portName,
		fmt.Sprintf("ingress_policing_rate=%d", bandwidthKbps))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set QoS on port %s: %w", portName, err)
	}

	// Set burst rate (10% over)
	cmd = exec.Command(o.vswitchdPath, "--", "set", "interface", portName,
		fmt.Sprintf("ingress_policing_burst=%d", bandwidthKbps*110/100))
	return cmd.Run()
}

// EnablePortSecurity enables anti-spoofing on a port
func (o *OVSClient) EnablePortSecurity(portName, macAddress, ipAddress string) error {
	// Set port security with MAC and IP
	cmd := exec.Command(o.vswitchdPath, "--", "set", "interface", portName,
		fmt.Sprintf("port_security=%s,%s", macAddress, ipAddress))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable port security on %s: %w", portName, err)
	}

	// Enable port security on the port
	cmd = exec.Command(o.vswitchdPath, "--", "set", "port", portName,
		"port_security=true")
	return cmd.Run()
}

// CreateTunnelPort creates a tunnel port on a bridge
func (o *OVSClient) CreateTunnelPort(tunnel *Tunnel) error {
	var interfaceType string
	var options []string

	switch tunnel.Protocol {
	case TunnelVXLAN:
		interfaceType = "vxlan"
		options = []string{
			fmt.Sprintf("remote_ip=%s", tunnel.RemoteIP),
			fmt.Sprintf("key=%d", tunnel.VNI),
		}
	case TunnelGRE:
		interfaceType = "gre"
		options = []string{
			fmt.Sprintf("remote_ip=%s", tunnel.RemoteIP),
			fmt.Sprintf("key=%d", tunnel.VNI),
		}
	case TunnelGeneve:
		interfaceType = "geneve"
		options = []string{
			fmt.Sprintf("remote_ip=%s", tunnel.RemoteIP),
			fmt.Sprintf("key=%d", tunnel.VNI),
		}
	default:
		interfaceType = "vxlan"
		options = []string{
			fmt.Sprintf("remote_ip=%s", tunnel.RemoteIP),
			fmt.Sprintf("key=%d", tunnel.VNI),
		}
	}

	// Create tunnel interface
	args := []string{"--", "add-port", tunnel.NetworkID, tunnel.ID, "--",
		"set", "interface", tunnel.ID, fmt.Sprintf("type=%s", interfaceType)}

	for _, opt := range options {
		args = append(args, fmt.Sprintf("options:%s", opt))
	}

	cmd := exec.Command(o.vswitchdPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tunnel port: %w", err)
	}

	return nil
}

// CreateRouterNamespace creates a network namespace for a router
func (o *OVSClient) CreateRouterNamespace(routerID string) error {
	// Create network namespace
	cmd := exec.Command(o.ipPath, "netns", "add", routerID)
	if err := cmd.Run(); err != nil {
		// Namespace might already exist
		if !strings.Contains(err.Error(), "File exists") {
			return fmt.Errorf("failed to create namespace %s: %w", routerID, err)
		}
	}

	// Bring up loopback in namespace
	cmd = exec.Command(o.ipPath, "netns", "exec", routerID, "ip", "link", "set", "lo", "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up loopback: %w", err)
	}

	return nil
}

// DeleteRouterNamespace deletes a router's network namespace
func (o *OVSClient) DeleteRouterNamespace(routerID string) error {
	cmd := exec.Command(o.ipPath, "netns", "delete", routerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete namespace %s: %w", routerID, err)
	}
	return nil
}

// CreateRouterInterface creates a veth pair for router interface
func (o *OVSClient) CreateRouterInterface(routerID string, iface *RouterInterface) error {
	// Create veth pair
	vethHost := fmt.Sprintf("%s-host", iface.ID)
	vethRouter := fmt.Sprintf("%s-router", iface.ID)

	cmd := exec.Command(o.ipPath, "link", "add", vethHost, "type", "veth",
		"peer", "name", vethRouter)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}

	// Move router side to namespace
	cmd = exec.Command(o.ipPath, "link", "set", vethRouter, "netns", routerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move veth to namespace: %w", err)
	}

	// Set IP address on router interface
	cmd = exec.Command(o.ipPath, "netns", "exec", routerID,
		"ip", "addr", "add", iface.IPAddress, "dev", vethRouter)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP on router interface: %w", err)
	}

	// Bring up router interface
	cmd = exec.Command(o.ipPath, "netns", "exec", routerID,
		"ip", "link", "set", vethRouter, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up router interface: %w", err)
	}

	// Connect host side to OVS bridge
	if iface.NetworkID != "" {
		cmd = exec.Command(o.vswitchdPath, "--", "add-port", iface.NetworkID, vethHost)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add port to bridge: %w", err)
		}
	}

	// Bring up host side
	if err := o.upInterface(vethHost); err != nil {
		return fmt.Errorf("failed to bring up host veth: %w", err)
	}

	return nil
}

// AddRoute adds a route to a router's routing table
func (o *OVSClient) AddRoute(routerID string, route Route) error {
	args := []string{"netns", "exec", routerID, "ip", "route", "add"}

	if route.Destination != "" {
		args = append(args, route.Destination)
	}

	if route.Gateway != "" {
		args = append(args, "via", route.Gateway)
	}

	if route.Interface != "" {
		args = append(args, "dev", route.Interface)
	}

	if route.Metric > 0 {
		args = append(args, "metric", strconv.Itoa(route.Metric))
	}

	cmd := exec.Command(o.ipPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add route: %w", err)
	}

	return nil
}

// DeleteRoute deletes a route from a router's routing table
func (o *OVSClient) DeleteRoute(routerID string, route Route) error {
	args := []string{"netns", "exec", routerID, "ip", "route", "del", route.Destination}

	cmd := exec.Command(o.ipPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	return nil
}

// CreateFirewallChain creates an iptables chain for a firewall
func (o *OVSClient) CreateFirewallChain(firewallID, defaultPolicy string) error {
	chainName := fmt.Sprintf("FW_%s", firewallID)

	// Create new chain
	cmd := exec.Command(o.iptablesPath, "-t", "filter", "-N", chainName)
	if err := cmd.Run(); err != nil {
		// Chain might already exist
		if !strings.Contains(err.Error(), "File exists") {
			return fmt.Errorf("failed to create chain: %w", err)
		}
	}

	// Set default policy
	cmd = exec.Command(o.iptablesPath, "-t", "filter", "-P", chainName, strings.ToUpper(defaultPolicy))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set default policy: %w", err)
	}

	return nil
}

// AddFirewallRule adds an iptables rule to a firewall chain
func (o *OVSClient) AddFirewallRule(firewallID string, rule FirewallRule) error {
	chainName := fmt.Sprintf("FW_%s", firewallID)

	args := []string{"-t", "filter", "-A", chainName}

	if rule.Direction == "ingress" {
		args = append(args, "-i", "eth0") // Input interface
	} else {
		args = append(args, "-o", "eth0") // Output interface
	}

	if rule.Protocol != "" && rule.Protocol != "all" {
		args = append(args, "-p", rule.Protocol)
	}

	if rule.SourceCIDR != "" {
		args = append(args, "-s", rule.SourceCIDR)
	}

	if rule.DestCIDR != "" {
		args = append(args, "-d", rule.DestCIDR)
	}

	if rule.SourcePort > 0 {
		args = append(args, "--sport", strconv.Itoa(rule.SourcePort))
	}

	if rule.DestPort > 0 {
		args = append(args, "--dport", strconv.Itoa(rule.DestPort))
	}

	if rule.Log {
		args = append(args, "-j", "LOG", "--log-prefix", fmt.Sprintf("[%s] ", rule.Name))
	}

	switch strings.ToLower(rule.Action) {
	case "accept":
		args = append(args, "-j", "ACCEPT")
	case "drop":
		args = append(args, "-j", "DROP")
	case "reject":
		args = append(args, "-j", "REJECT")
	}

	cmd := exec.Command(o.iptablesPath, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add firewall rule: %w", err)
	}

	return nil
}

// DeleteFirewallChain deletes an iptables chain
func (o *OVSClient) DeleteFirewallChain(firewallID string) error {
	chainName := fmt.Sprintf("FW_%s", firewallID)

	// Flush chain
	cmd := exec.Command(o.iptablesPath, "-t", "filter", "-F", chainName)
	cmd.Run() // Ignore errors

	// Delete chain
	cmd = exec.Command(o.iptablesPath, "-t", "filter", "-X", chainName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete chain: %w", err)
	}

	return nil
}

// GetBridgeStats returns statistics for a bridge
func (o *OVSClient) GetBridgeStats(bridgeName string) (*NetworkStats, error) {
	stats := &NetworkStats{BridgeName: bridgeName}

	// Get interface statistics using ovs-vsctl
	cmd := exec.Command(o.vswitchdPath, "--columns=statistics", "list", "interface")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Parse statistics
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "rx_bytes") {
			stats.RxBytes = parseStatistic(line, "rx_bytes")
		}
		if strings.Contains(line, "tx_bytes") {
			stats.TxBytes = parseStatistic(line, "tx_bytes")
		}
		if strings.Contains(line, "rx_packets") {
			stats.RxPackets = parseStatistic(line, "rx_packets")
		}
		if strings.Contains(line, "tx_packets") {
			stats.TxPackets = parseStatistic(line, "tx_packets")
		}
		if strings.Contains(line, "rx_errors") {
			stats.RxErrors = parseStatistic(line, "rx_errors")
		}
		if strings.Contains(line, "tx_errors") {
			stats.TxErrors = parseStatistic(line, "tx_errors")
		}
	}

	// Get port count
	cmd = exec.Command(o.vswitchdPath, "--columns=ports", "list", "bridge", bridgeName)
	output, err = cmd.Output()
	if err == nil {
		ports := strings.Split(string(output), "\n")
		stats.ConnectedPorts = len(ports) - 1 // Subtract bridge itself
	}

	// Get flow count
	cmd = exec.Command(o.ofctlPath, "dump-flows", bridgeName)
	output, _ = cmd.Output()
	stats.FlowCount = strings.Count(string(output), "\n")

	return stats, nil
}

// bridgeExists checks if a bridge exists
func (o *OVSClient) bridgeExists(name string) bool {
	cmd := exec.Command(o.vswitchdPath, "br-exists", name)
	err := cmd.Run()
	return err == nil
}

// upInterface brings up a network interface
func (o *OVSClient) upInterface(name string) error {
	cmd := exec.Command(o.ipPath, "link", "set", name, "up")
	return cmd.Run()
}

// parseStatistic extracts a statistic value from OVS output
func parseStatistic(line, key string) int64 {
	idx := strings.Index(line, key)
	if idx == -1 {
		return 0
	}

	valueStr := line[idx+len(key):]
	valueStr = strings.TrimLeft(valueStr, "=:")

	// Extract number
	var numStr string
	for _, c := range valueStr {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		} else {
			break
		}
	}

	val, _ := strconv.ParseInt(numStr, 10, 64)
	return val
}

// ListBridges lists all OVS bridges
func (o *OVSClient) ListBridges() ([]string, error) {
	cmd := exec.Command(o.vswitchdPath, "list-br")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list bridges: %w", err)
	}

	bridges := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(bridges) == 1 && bridges[0] == "" {
		return []string{}, nil
	}
	return bridges, nil
}

// ListPorts lists all ports on a bridge
func (o *OVSClient) ListPorts(bridgeName string) ([]string, error) {
	cmd := exec.Command(o.vswitchdPath, "list-ports", bridgeName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	ports := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(ports) == 1 && ports[0] == "" {
		return []string{}, nil
	}
	return ports, nil
}

// GetPortInfo returns information about a port
func (o *OVSClient) GetPortInfo(portName string) (map[string]string, error) {
	cmd := exec.Command(o.vswitchdPath, "--columns=name,tag,trunks,interfaces", "list", "port", portName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get port info: %w", err)
	}

	info := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			info[key] = value
		}
	}

	return info, nil
}

// DumpFlows dumps all flows from a bridge
func (o *OVSClient) DumpFlows(bridgeName string) ([]string, error) {
	cmd := exec.Command(o.ofctlPath, "dump-flows", bridgeName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to dump flows: %w", err)
	}

	flows := strings.Split(strings.TrimSpace(string(output)), "\n")
	return flows, nil
}

// AddFlow adds a flow entry to a bridge
func (o *OVSClient) AddFlow(bridgeName, flow string) error {
	cmd := exec.Command(o.ofctlPath, "add-flow", bridgeName, flow)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add flow: %w", err)
	}
	return nil
}

// DelFlow deletes a flow entry from a bridge
func (o *OVSClient) DelFlow(bridgeName, flow string) error {
	cmd := exec.Command(o.ofctlPath, "del-flows", bridgeName, flow)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}
	return nil
}

// Execute runs a command and returns combined output
func (o *OVSClient) Execute(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}