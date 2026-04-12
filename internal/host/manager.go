// Package host provides multi-host hypervisor management
package host

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Manager manages connections to multiple hypervisor hosts
type Manager struct {
	hosts map[string]*HostConnection
	db    *database.DB
	// defaultHV is kept for future use when multi-hypervisor support is added
	// nolint:unused
	defaultHV hypervisor.Hypervisor
}

// HostConnection represents a connection to a hypervisor host
type HostConnection struct {
	ID      string
	Name    string
	Address string
	Port    int
	User    string
	SSHKey  string
	client  *ssh.Client
	hv      hypervisor.Hypervisor
	IsLocal bool
}

// NewManager creates a new host manager
func NewManager(db *database.DB) *Manager {
	return &Manager{
		hosts: make(map[string]*HostConnection),
		db:    db,
	}
}

// AddHost adds and connects to a new host
func (m *Manager) AddHost(cfg *database.Host) (*HostConnection, error) {
	// Check if already connected
	if conn, ok := m.hosts[cfg.ID]; ok {
		return conn, nil
	}

	conn := &HostConnection{
		ID:      cfg.ID,
		Name:    cfg.Name,
		Address: cfg.Address,
		Port:    cfg.Port,
		User:    cfg.User,
		SSHKey:  cfg.SSHKeyPath,
		IsLocal: m.isLocalAddress(cfg.Address),
	}

	if conn.IsLocal {
		// Use local hypervisor
		hv, err := hypervisor.NewHypervisor(&hypervisor.HostConfig{
			Address:    cfg.Address,
			Port:       cfg.Port,
			User:       cfg.User,
			SSHKeyPath: cfg.SSHKeyPath,
			Type:       cfg.HVType,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create local hypervisor: %w", err)
		}
		conn.hv = hv
	} else {
		// Connect via SSH
		if err := m.connectSSH(conn); err != nil {
			return nil, fmt.Errorf("failed to SSH to host: %w", err)
		}
	}

	m.hosts[cfg.ID] = conn
	return conn, nil
}

// connectSSH establishes an SSH connection to a remote host
func (m *Manager) connectSSH(conn *HostConnection) error {
	auth := []ssh.AuthMethod{}

	if conn.SSHKey != "" {
		key, err := os.ReadFile(conn.SSHKey)
		if err != nil {
			return fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse SSH key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	// Try password auth if no key
	if conn.User == "root" {
		// Could prompt for password here
	}

	addr := fmt.Sprintf("%s:%d", conn.Address, conn.Port)
	if conn.Port == 0 {
		addr = fmt.Sprintf("%s:22", conn.Address)
	}

	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            conn.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}

	conn.client = client
	return nil
}

// isLocalAddress checks if an address refers to the local machine
func (m *Manager) isLocalAddress(addr string) bool {
	// Empty address means local libvirt (qemu:///system)
	if addr == "" {
		return true
	}
	if addr == "localhost" || addr == "127.0.0.1" || addr == "::1" {
		return true
	}

	// Check if it's the local IP
	if net.ParseIP(addr) != nil {
		ifaces, err := net.Interfaces()
		if err != nil {
			return false
		}
		for _, iface := range ifaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addrs {
				if ip, ok := a.(*net.IPNet); ok && ip.IP.Equal(net.ParseIP(addr)) {
					return true
				}
			}
		}
	}

	return false
}

// GetHypervisor returns the hypervisor for a host
func (m *Manager) GetHypervisor(hostID string) (hypervisor.Hypervisor, error) {
	if conn, ok := m.hosts[hostID]; ok {
		return conn.hv, nil
	}

	// Try to load from database
	host, err := m.db.GetHost(hostID)
	if err != nil || host == nil {
		return nil, fmt.Errorf("host not found: %s", hostID)
	}

	conn, err := m.AddHost(host)
	if err != nil {
		return nil, err
	}

	return conn.hv, nil
}

// GetConnection returns a host connection
func (m *Manager) GetConnection(hostID string) (*HostConnection, error) {
	if conn, ok := m.hosts[hostID]; ok {
		return conn, nil
	}
	return nil, fmt.Errorf("host not connected: %s", hostID)
}

// RemoveHost disconnects and removes a host
func (m *Manager) RemoveHost(hostID string) error {
	if conn, ok := m.hosts[hostID]; ok {
		if conn.client != nil {
			conn.client.Close()
		}
		if conn.hv != nil {
			conn.hv.Close()
		}
		delete(m.hosts, hostID)
	}
	return nil
}

// ListHosts returns all connected hosts
func (m *Manager) ListHosts() []*HostConnection {
	var list []*HostConnection
	for _, conn := range m.hosts {
		list = append(list, conn)
	}
	return list
}

// ExecSSH executes a command on a remote host via SSH
func (c *HostConnection) ExecSSH(cmd string) ([]byte, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected via SSH")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return session.CombinedOutput(cmd)
}

// ExecLocal executes a command locally
func ExecLocal(cmd string, args ...string) ([]byte, error) {
	c := exec.Command(cmd, args...)
	return c.CombinedOutput()
}

// RefreshConnections re-establishes connections to all hosts
func (m *Manager) RefreshConnections() error {
	hosts, err := m.db.ListHosts()
	if err != nil {
		return err
	}

	for _, host := range hosts {
		if _, ok := m.hosts[host.ID]; !ok {
			if _, err := m.AddHost(host); err != nil {
				// Log but continue
				fmt.Printf("Failed to reconnect to %s: %v\n", host.Name, err)
			}
		}
	}

	return nil
}

// GetHostStatus returns the status of a host connection
func (c *HostConnection) GetStatus() string {
	if c.client != nil {
		return "connected"
	}
	if c.hv != nil {
		return "active"
	}
	return "disconnected"
}

// HostInfo holds information about a host
type HostInfo struct {
	ID        string
	Name      string
	Address   string
	Status    string
	Nodes     int
	CPUUsage  float64
	MemUsage  float64
	DiskUsage float64
}

// GetHostInfo returns detailed information about a host
func (m *Manager) GetHostInfo(hostID string) (*HostInfo, error) {
	conn, err := m.GetConnection(hostID)
	if err != nil {
		return nil, err
	}

	info := &HostInfo{
		ID:      conn.ID,
		Name:    conn.Name,
		Address: conn.Address,
		Status:  conn.GetStatus(),
	}

	// Get node count and metrics
	if conn.hv != nil {
		nodes, err := conn.hv.ListNodes(context.Background())
		if err == nil {
			info.Nodes = len(nodes)
		}
	}

	return info, nil
}

// BestHostSelector helps select the best host for a new node
type BestHostSelector struct {
	hosts *Manager
}

// NewBestHostSelector creates a selector
func NewBestHostSelector(hosts *Manager) *BestHostSelector {
	return &BestHostSelector{hosts: hosts}
}

// SelectHost selects the host with the most available resources
func (s *BestHostSelector) SelectHost() (string, error) {
	var bestID string
	var bestScore float64

	for id := range s.hosts.hosts {
		info, err := s.hosts.GetHostInfo(id)
		if err != nil {
			continue
		}

		// Simple score: lower usage = better
		score := 100.0 - info.CPUUsage - info.MemUsage
		if score > bestScore {
			bestScore = score
			bestID = id
		}
	}

	if bestID == "" {
		// Default to first host
		for id := range s.hosts.hosts {
			return id, nil
		}
		return "", fmt.Errorf("no hosts available")
	}

	return bestID, nil
}
