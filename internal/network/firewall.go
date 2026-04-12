// Package network provides firewall management
package network

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// FirewallBackend represents the firewall backend type
type FirewallBackend string

const (
	FirewallBackendIPTables FirewallBackend = "iptables"
	FirewallBackendNFTables FirewallBackend = "nftables"
	FirewallBackendStub     FirewallBackend = "stub"
)

// FirewallManager manages firewall rules for network isolation
type FirewallManager struct {
	backend FirewallBackend
	rules   map[string][]string // Chain -> Rules
	mu      sync.RWMutex
}

// NewFirewallManager creates a new firewall manager
func NewFirewallManager(backend FirewallBackend) (*FirewallManager, error) {
	fm := &FirewallManager{
		backend: backend,
		rules:   make(map[string][]string),
	}

	// Detect backend if not specified
	if backend == "" {
		fm.detectBackend()
	}

	// Verify backend is available
	if !fm.isBackendAvailable() {
		return nil, fmt.Errorf("firewall backend not available: %s", fm.backend)
	}

	// Initialize chains
	if err := fm.initializeChains(); err != nil {
		return nil, fmt.Errorf("failed to initialize chains: %w", err)
	}

	return fm, nil
}

// detectBackend detects the available firewall backend
func (fm *FirewallManager) detectBackend() {
	// Prefer nftables over iptables
	if _, err := exec.LookPath("nft"); err == nil {
		fm.backend = FirewallBackendNFTables
	} else if _, err := exec.LookPath("iptables"); err == nil {
		fm.backend = FirewallBackendIPTables
	} else {
		fm.backend = FirewallBackendStub
	}
}

// isBackendAvailable checks if the backend is available
func (fm *FirewallManager) isBackendAvailable() bool {
	switch fm.backend {
	case FirewallBackendIPTables:
		_, err := exec.LookPath("iptables")
		return err == nil
	case FirewallBackendNFTables:
		_, err := exec.LookPath("nft")
		return err == nil
	case FirewallBackendStub:
		return true
	default:
		return false
	}
}

// initializeChains initializes the firewall chains
func (fm *FirewallManager) initializeChains() error {
	switch fm.backend {
	case FirewallBackendIPTables:
		return fm.initIPTablesChains()
	case FirewallBackendNFTables:
		return fm.initNFTablesChains()
	case FirewallBackendStub:
		return nil
	default:
		return fmt.Errorf("unsupported backend: %s", fm.backend)
	}
}

// initIPTablesChains initializes iptables chains
func (fm *FirewallManager) initIPTablesChains() error {
	// Create VIMIC2 chain
	cmd := exec.Command("iptables", "-t", "filter", "-N", "VIMIC2")
	if err := cmd.Run(); err != nil {
		// Chain might already exist, ignore error
	}

	// Jump to VIMIC2 chain from FORWARD
	cmd = exec.Command("iptables", "-t", "filter", "-I", "FORWARD", "-j", "VIMIC2")
	cmd.Run() // Ignore error if already exists

	// Default policies
	cmd = exec.Command("iptables", "-t", "filter", "-A", "VIMIC2", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add conntrack rule: %w: %s", err, output)
	}

	return nil
}

// initNFTablesChains initializes nftables chains
func (fm *FirewallManager) initNFTablesChains() error {
	// Create table
	cmd := exec.Command("nft", "add", "table", "inet", "vimic2")
	if err := cmd.Run(); err != nil {
		// Table might already exist, ignore error
	}

	// Create forward chain
	cmd = exec.Command("nft", "add", "chain", "inet", "vimic2", "forward", "{ type filter hook forward priority 0 \\; }")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create forward chain: %w: %s", err, output)
	}

	// Allow established connections
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", "forward", "ct", "state", "established,related", "accept")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add conntrack rule: %w: %s", err, output)
	}

	return nil
}

// CreateIsolationRules creates firewall rules for network isolation
func (fm *FirewallManager) CreateIsolationRules(bridgeName, cidr string, vlanID int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Parse CIDR
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CIDR: %s", cidr)
	}
	subnet := parts[0]

	// Create chain for this network
	chainName := fmt.Sprintf("VIMIC2_%d", vlanID)

	switch fm.backend {
	case FirewallBackendIPTables:
		return fm.createIPTablesIsolation(bridgeName, subnet, vlanID, chainName)
	case FirewallBackendNFTables:
		return fm.createNFTablesIsolation(bridgeName, subnet, vlanID, chainName)
	case FirewallBackendStub:
		// Stub mode - just track rules
		fm.rules[chainName] = []string{
			fmt.Sprintf("allow-established"),
			fmt.Sprintf("allow-dns"),
			fmt.Sprintf("drop-inter-network"),
		}
		return nil
	default:
		return fmt.Errorf("unsupported backend: %s", fm.backend)
	}
}

// createIPTablesIsolation creates iptables isolation rules
func (fm *FirewallManager) createIPTablesIsolation(bridgeName, subnet string, vlanID int, chainName string) error {
	// Create chain
	cmd := exec.Command("iptables", "-t", "filter", "-N", chainName)
	if err := cmd.Run(); err != nil {
		// Chain might already exist, ignore error
	}

	// Allow established connections
	cmd = exec.Command("iptables", "-t", "filter", "-A", chainName, "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add established rule: %w: %s", err, output)
	}

	// Allow DNS
	cmd = exec.Command("iptables", "-t", "filter", "-A", chainName, "-p", "udp", "--dport", "53", "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add DNS rule: %w: %s", err, output)
	}

	// Allow intra-network traffic (same VLAN)
	cmd = exec.Command("iptables", "-t", "filter", "-A", chainName, "-s", subnet+"/24", "-d", subnet+"/24", "-j", "ACCEPT")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add intra-network rule: %w: %s", err, output)
	}

	// Drop inter-network traffic
	cmd = exec.Command("iptables", "-t", "filter", "-A", chainName, "-s", subnet+"/24", "-j", "DROP")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add drop rule: %w: %s", err, output)
	}

	// Jump to chain from VIMIC2
	cmd = exec.Command("iptables", "-t", "filter", "-A", "VIMIC2", "-i", bridgeName, "-j", chainName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add jump rule: %w: %s", err, output)
	}

	// Track rules
	fm.rules[chainName] = []string{
		fmt.Sprintf("established: %s", subnet),
		fmt.Sprintf("dns: %s", subnet),
		fmt.Sprintf("intra-network: %s", subnet),
		fmt.Sprintf("drop-inter: %s", subnet),
	}

	return nil
}

// createNFTablesIsolation creates nftables isolation rules
func (fm *FirewallManager) createNFTablesIsolation(bridgeName, subnet string, vlanID int, chainName string) error {
	// Create chain
	cmd := exec.Command("nft", "add", "chain", "inet", "vimic2", chainName)
	if err := cmd.Run(); err != nil {
		// Chain might already exist, ignore error
	}

	// Allow established connections
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", chainName, "ct", "state", "established,related", "accept")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add established rule: %w: %s", err, output)
	}

	// Allow DNS
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", chainName, "udp", "dport", "53", "accept")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add DNS rule: %w: %s", err, output)
	}

	// Allow intra-network traffic
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", chainName, "ip", "saddr", subnet+"/24", "ip", "daddr", subnet+"/24", "accept")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add intra-network rule: %w: %s", err, output)
	}

	// Drop inter-network traffic
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", chainName, "ip", "saddr", subnet+"/24", "drop")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add drop rule: %w: %s", err, output)
	}

	// Jump to chain from forward
	cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", "forward", "iifname", bridgeName, "jump", chainName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add jump rule: %w: %s", err, output)
	}

	// Track rules
	fm.rules[chainName] = []string{
		fmt.Sprintf("established: %s", subnet),
		fmt.Sprintf("dns: %s", subnet),
		fmt.Sprintf("intra-network: %s", subnet),
		fmt.Sprintf("drop-inter: %s", subnet),
	}

	return nil
}

// DeleteIsolationRules deletes firewall rules for a network
func (fm *FirewallManager) DeleteIsolationRules(bridgeName, cidr string, vlanID int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	chainName := fmt.Sprintf("VIMIC2_%d", vlanID)

	switch fm.backend {
	case FirewallBackendIPTables:
		return fm.deleteIPTablesIsolation(bridgeName, vlanID, chainName)
	case FirewallBackendNFTables:
		return fm.deleteNFTablesIsolation(bridgeName, vlanID, chainName)
	case FirewallBackendStub:
		delete(fm.rules, chainName)
		return nil
	default:
		return fmt.Errorf("unsupported backend: %s", fm.backend)
	}
}

// deleteIPTablesIsolation deletes iptables isolation rules
func (fm *FirewallManager) deleteIPTablesIsolation(bridgeName string, vlanID int, chainName string) error {
	// Remove jump from VIMIC2
	cmd := exec.Command("iptables", "-t", "filter", "-D", "VIMIC2", "-i", bridgeName, "-j", chainName)
	cmd.Run() // Ignore error if rule doesn't exist

	// Flush chain
	cmd = exec.Command("iptables", "-t", "filter", "-F", chainName)
	cmd.Run() // Ignore error if chain is empty

	// Delete chain
	cmd = exec.Command("iptables", "-t", "filter", "-X", chainName)
	cmd.Run() // Ignore error if chain doesn't exist

	// Remove from tracked rules
	delete(fm.rules, chainName)

	return nil
}

// deleteNFTablesIsolation deletes nftables isolation rules
func (fm *FirewallManager) deleteNFTablesIsolation(bridgeName string, vlanID int, chainName string) error {
	// Delete chain (automatically removes rules)
	cmd := exec.Command("nft", "delete", "chain", "inet", "vimic2", chainName)
	cmd.Run() // Ignore error if chain doesn't exist

	// Remove from tracked rules
	delete(fm.rules, chainName)

	return nil
}

// ListRules lists all firewall rules
func (fm *FirewallManager) ListRules() map[string][]string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	rules := make(map[string][]string)
	for chain, chainRules := range fm.rules {
		rules[chain] = append([]string{}, chainRules...)
	}
	return rules
}

// AllowTraffic allows traffic between networks
func (fm *FirewallManager) AllowTraffic(sourceCIDR, destCIDR string, ports []int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	switch fm.backend {
	case FirewallBackendIPTables:
		return fm.allowIPTablesTraffic(sourceCIDR, destCIDR, ports)
	case FirewallBackendNFTables:
		return fm.allowNFTablesTraffic(sourceCIDR, destCIDR, ports)
	case FirewallBackendStub:
		return nil
	default:
		return fmt.Errorf("unsupported backend: %s", fm.backend)
	}
}

// allowIPTablesTraffic allows traffic between networks using iptables
func (fm *FirewallManager) allowIPTablesTraffic(sourceCIDR, destCIDR string, ports []int) error {
	for _, port := range ports {
		// Allow TCP
		cmd := exec.Command("iptables", "-t", "filter", "-I", "VIMIC2", "1",
			"-s", sourceCIDR, "-d", destCIDR,
			"-p", "tcp", "--dport", fmt.Sprintf("%d", port),
			"-j", "ACCEPT")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to allow TCP port %d: %w: %s", port, err, output)
		}

		// Allow UDP
		cmd = exec.Command("iptables", "-t", "filter", "-I", "VIMIC2", "1",
			"-s", sourceCIDR, "-d", destCIDR,
			"-p", "udp", "--dport", fmt.Sprintf("%d", port),
			"-j", "ACCEPT")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to allow UDP port %d: %w: %s", port, err, output)
		}
	}

	return nil
}

// allowNFTablesTraffic allows traffic between networks using nftables
func (fm *FirewallManager) allowNFTablesTraffic(sourceCIDR, destCIDR string, ports []int) error {
	for _, port := range ports {
		// Allow TCP
		cmd := exec.Command("nft", "add", "rule", "inet", "vimic2", "forward",
			"ip", "saddr", sourceCIDR, "ip", "daddr", destCIDR,
			"tcp", "dport", fmt.Sprintf("%d", port), "accept")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to allow TCP port %d: %w: %s", port, err, output)
		}

		// Allow UDP
		cmd = exec.Command("nft", "add", "rule", "inet", "vimic2", "forward",
			"ip", "saddr", sourceCIDR, "ip", "daddr", destCIDR,
			"udp", "dport", fmt.Sprintf("%d", port), "accept")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to allow UDP port %d: %w: %s", port, err, output)
		}
	}

	return nil
}

// DenyTraffic denies traffic between networks
func (fm *FirewallManager) DenyTraffic(sourceCIDR, destCIDR string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	switch fm.backend {
	case FirewallBackendIPTables:
		return fm.denyIPTablesTraffic(sourceCIDR, destCIDR)
	case FirewallBackendNFTables:
		return fm.denyNFTablesTraffic(sourceCIDR, destCIDR)
	case FirewallBackendStub:
		return nil
	default:
		return fmt.Errorf("unsupported backend: %s", fm.backend)
	}
}

// denyIPTablesTraffic denies traffic between networks using iptables
func (fm *FirewallManager) denyIPTablesTraffic(sourceCIDR, destCIDR string) error {
	cmd := exec.Command("iptables", "-t", "filter", "-I", "VIMIC2", "1",
		"-s", sourceCIDR, "-d", destCIDR,
		"-j", "DROP")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to deny traffic: %w: %s", err, output)
	}
	return nil
}

// denyNFTablesTraffic denies traffic between networks using nftables
func (fm *FirewallManager) denyNFTablesTraffic(sourceCIDR, destCIDR string) error {
	cmd := exec.Command("nft", "add", "rule", "inet", "vimic2", "forward",
		"ip", "saddr", sourceCIDR, "ip", "daddr", destCIDR, "drop")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to deny traffic: %w: %s", err, output)
	}
	return nil
}

// GetBackend returns the firewall backend
func (fm *FirewallManager) GetBackend() FirewallBackend {
	return fm.backend
}

// Close closes the firewall manager
func (fm *FirewallManager) Close() error {
	// Nothing to close
	return nil
}
