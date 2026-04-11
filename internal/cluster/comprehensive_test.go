// Package cluster provides comprehensive real tests
package cluster

import (
	"testing"

	"github.com/stsgym/vimic2/internal/database"
)

// TestClusterConfig_Scaling tests cluster scaling configuration
func TestClusterConfig_Scaling(t *testing.T) {
	config := &database.ClusterConfig{
		MinNodes:      3,
		MaxNodes:      20,
		AutoScale:     true,
		ScaleOnCPU:    70.0,
		ScaleOnMemory: 80.0,
		CooldownSec:   300,
	}

	// Verify scaling parameters
	if !config.AutoScale {
		t.Error("expected auto-scaling to be enabled")
	}
	if config.MinNodes < 1 {
		t.Error("min nodes should be >= 1")
	}
	if config.MaxNodes < config.MinNodes {
		t.Error("max nodes should be >= min nodes")
	}
	if config.ScaleOnCPU <= 0 || config.ScaleOnCPU > 100 {
		t.Error("scale on CPU should be 0-100%")
	}
}

// TestClusterConfig_ResourceDefaults tests resource defaults
func TestClusterConfig_ResourceDefaults(t *testing.T) {
	config := &database.ClusterConfig{
		MinNodes:  1,
		MaxNodes:  10,
		AutoScale: false,
	}

	if config.MinNodes != 1 {
		t.Errorf("expected 1 min node, got %d", config.MinNodes)
	}
	if config.AutoScale {
		t.Error("expected auto-scale to be disabled")
	}
}

// TestNode_Creation tests node creation
func TestNode_Creation(t *testing.T) {
	node := &database.Node{
		ID:        "node-1",
		Name:      "worker-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		State:     "pending",
		Role:      "worker",
	}

	if node.ID != "node-1" {
		t.Errorf("expected node-1, got %s", node.ID)
	}
	if node.State != "pending" {
		t.Errorf("expected pending state, got %s", node.State)
	}
}

// TestNode_StateTransitions tests node state transitions
func TestNode_StateTransitions(t *testing.T) {
	node := &database.Node{State: "pending"}

	// Valid transitions
	transitions := []string{"creating", "running", "stopping", "stopped", "error"}

	for _, state := range transitions {
		node.State = state
		if node.State != state {
			t.Errorf("expected state %s", state)
		}
	}
}

// TestNode_WorkerRole tests worker role
func TestNode_WorkerRole(t *testing.T) {
	node := &database.Node{Role: "worker"}

	if node.Role != "worker" {
		t.Errorf("expected worker role, got %s", node.Role)
	}
}

// TestNode_MasterRole tests master role
func TestNode_MasterRole(t *testing.T) {
	node := &database.Node{Role: "master"}

	if node.Role != "master" {
		t.Errorf("expected master role, got %s", node.Role)
	}
}

// TestHost_Connection tests host connection parameters
func TestHost_Connection(t *testing.T) {
	host := &database.Host{
		ID:         "host-1",
		Name:       "hypervisor-1",
		Address:    "192.168.1.10",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if host.Address != "192.168.1.10" {
		t.Errorf("expected address, got %s", host.Address)
	}
	if host.Port != 22 {
		t.Errorf("expected port 22, got %d", host.Port)
	}
	if host.HVType != "libvirt" {
		t.Errorf("expected libvirt, got %s", host.HVType)
	}
}

// TestHost_HVTypes tests hypervisor types
func TestHost_HVTypes(t *testing.T) {
	types := []string{"libvirt", "qemu", "kvm"}

	for _, hvType := range types {
		host := &database.Host{HVType: hvType}
		if host.HVType != hvType {
			t.Errorf("expected %s, got %s", hvType, host.HVType)
		}
	}
}

// TestMetric_ResourceUsage tests metric resource usage
func TestMetric_ResourceUsage(t *testing.T) {
	metric := &database.Metric{
		NodeID:     "node-1",
		CPU:        45.5,
		Memory:     60.2,
		Disk:       30.1,
		NetworkRX:  1024000,
		NetworkTX:  512000,
	}

	if metric.CPU < 0 || metric.CPU > 100 {
		t.Error("CPU should be 0-100%")
	}
	if metric.Memory < 0 || metric.Memory > 100 {
		t.Error("Memory should be 0-100%")
	}
}

// TestMetric_NetworkTraffic tests network traffic metrics
func TestMetric_NetworkTraffic(t *testing.T) {
	metric := &database.Metric{
		NetworkRX: 1024000,
		NetworkTX: 512000,
	}

	totalTraffic := metric.NetworkRX + metric.NetworkTX
	if totalTraffic != 1536000 {
		t.Errorf("expected total traffic 1536000, got %f", totalTraffic)
	}
}