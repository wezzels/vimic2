// Package cluster tests cluster manager operations
package cluster

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestNewManager tests manager creation
func TestNewManager_Basic(t *testing.T) {
	mgr := NewManager(nil, nil)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestManager_DeleteCluster tests cluster deletion (requires DB)
func TestManager_DeleteCluster_Basic(t *testing.T) {
	t.Skip("requires database setup")
}

// TestManager_NodeStats tests node statistics (requires DB)
func TestManager_NodeStats_Basic(t *testing.T) {
	t.Skip("requires database setup")
}

// TestClusterConfig tests cluster configuration
func TestClusterConfig_Basic(t *testing.T) {
	cfg := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  10,
		AutoScale: true,
		NodeDefaults: &database.NodeConfig{
			CPU:      4,
			MemoryMB: 8192,
			DiskGB:   100,
		},
	}

	if cfg.MinNodes != 1 {
		t.Errorf("expected MinNodes 1, got %d", cfg.MinNodes)
	}

	if cfg.MaxNodes != 10 {
		t.Errorf("expected MaxNodes 10, got %d", cfg.MaxNodes)
	}

	if !cfg.AutoScale {
		t.Error("expected AutoScale to be true")
	}

	if cfg.NodeDefaults == nil {
		t.Fatal("expected non-nil NodeDefaults")
	}

	if cfg.NodeDefaults.CPU != 4 {
		t.Errorf("expected CPU 4, got %d", cfg.NodeDefaults.CPU)
	}
}

// TestNodeStruct tests Node struct
func TestNodeStruct_Basic(t *testing.T) {
	node := &database.Node{
		ID:        "node-1",
		Name:      "worker-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		State:     "running",
		Role:      "worker",
		IP:        "192.168.1.100",
		CreatedAt: time.Now(),
	}

	if node.ID != "node-1" {
		t.Errorf("expected node-1, got %s", node.ID)
	}

	if node.State != "running" {
		t.Errorf("expected running, got %s", node.State)
	}

	if node.IP != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", node.IP)
	}
}

// TestHostStruct tests Host struct
func TestHostStruct_Basic(t *testing.T) {
	host := &database.Host{
		ID:         "host-1",
		Name:       "hypervisor-1",
		Address:    "192.168.1.10",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if host.ID != "host-1" {
		t.Errorf("expected host-1, got %s", host.ID)
	}

	if host.Address != "192.168.1.10" {
		t.Errorf("expected address 192.168.1.10, got %s", host.Address)
	}

	if host.HVType != "libvirt" {
		t.Errorf("expected HVType libvirt, got %s", host.HVType)
	}
}

// TestNodeConfig tests node configuration struct
func TestNodeConfig_Basic(t *testing.T) {
	cfg := &database.NodeConfig{
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
	}

	if cfg.CPU != 4 {
		t.Errorf("expected CPU 4, got %d", cfg.CPU)
	}

	if cfg.MemoryMB != 8192 {
		t.Errorf("expected MemoryMB 8192, got %d", cfg.MemoryMB)
	}

	if cfg.DiskGB != 100 {
		t.Errorf("expected DiskGB 100, got %d", cfg.DiskGB)
	}

	if cfg.Image != "ubuntu-22.04" {
		t.Errorf("expected Image ubuntu-22.04, got %s", cfg.Image)
	}
}

// TestNetworkConfig tests network configuration
func TestNetworkConfig_Basic(t *testing.T) {
	cfg := &database.NetworkConfig{
		Type: "nat",
		CIDR: "192.168.122.0/24",
	}

	if cfg.Type != "nat" {
		t.Errorf("expected nat, got %s", cfg.Type)
	}

	if cfg.CIDR != "192.168.122.0/24" {
		t.Errorf("expected CIDR 192.168.122.0/24, got %s", cfg.CIDR)
	}
}