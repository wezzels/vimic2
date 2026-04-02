// Package cluster provides cluster management functionality
package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Manager handles cluster operations
type Manager struct {
	db    *database.DB
	hosts map[string]hypervisor.Hypervisor
}

// NewManager creates a new cluster manager
func NewManager(db *database.DB, hosts map[string]hypervisor.Hypervisor) *Manager {
	return &Manager{
		db:    db,
		hosts: hosts,
	}
}

// ============== Cluster Operations ==============

// CreateCluster creates a new cluster
func (m *Manager) CreateCluster(name string, cfg *database.ClusterConfig) (*database.Cluster, error) {
	cluster := &database.Cluster{
		ID:     uuid.New().String(),
		Name:   name,
		Config: cfg,
		Status: "pending",
	}

	if err := m.db.SaveCluster(cluster); err != nil {
		return nil, fmt.Errorf("failed to save cluster: %w", err)
	}

	return cluster, nil
}

// GetCluster retrieves a cluster by ID
func (m *Manager) GetCluster(id string) (*database.Cluster, error) {
	return m.db.GetCluster(id)
}

// ListClusters returns all clusters
func (m *Manager) ListClusters() ([]*database.Cluster, error) {
	return m.db.ListClusters()
}

// DeleteCluster deletes a cluster and all its nodes
func (m *Manager) DeleteCluster(ctx context.Context, id string) error {
	// Get all nodes in the cluster
	nodes, err := m.db.ListClusterNodes(id)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// Delete each node from its hypervisor
	for _, node := range nodes {
		if hv, ok := m.hosts[node.HostID]; ok {
			if err := hv.DeleteNode(ctx, node.ID); err != nil {
				// Log but continue
				fmt.Printf("Failed to delete node %s from hypervisor: %v\n", node.Name, err)
			}
		}
		if err := m.db.DeleteNode(node.ID); err != nil {
			fmt.Printf("Failed to delete node %s from db: %v\n", node.Name, err)
		}
	}

	// Delete the cluster
	return m.db.DeleteCluster(id)
}

// DeployCluster deploys all nodes in a cluster
func (m *Manager) DeployCluster(ctx context.Context, clusterID string) error {
	cluster, err := m.db.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}
	if cluster == nil {
		return fmt.Errorf("cluster not found")
	}

	// Update status to deploying
	if err := m.db.UpdateClusterStatus(clusterID, "deploying"); err != nil {
		return err
	}

	// Get cluster nodes
	nodes, err := m.db.ListClusterNodes(clusterID)
	if err != nil {
		m.db.UpdateClusterStatus(clusterID, "error")
		return err
	}

	// Provision each node
	for _, nodeRef := range nodes {
		if err := m.provisionNode(ctx, cluster, nodeRef); err != nil {
			m.db.UpdateClusterStatus(clusterID, "error")
			return fmt.Errorf("failed to provision node %s: %w", nodeRef.Name, err)
		}
	}

	// Update status to running
	return m.db.UpdateClusterStatus(clusterID, "running")
}

// provisionNode creates a VM for a node reference
func (m *Manager) provisionNode(ctx context.Context, cluster *database.Cluster, nodeRef *database.Node) error {
	// Get the hypervisor for this node's host
	hv, ok := m.hosts[nodeRef.HostID]
	if !ok {
		return fmt.Errorf("host not found: %s", nodeRef.HostID)
	}

	// Merge cluster defaults with node config
	nodeCfg := &hypervisor.NodeConfig{
		Name:     nodeRef.Name,
		CPU:      2,
		MemoryMB: 2048,
		DiskGB:   20,
		Image:    "ubuntu-22.04",
		Network:  "default",
	}

	if cluster.Config != nil && cluster.Config.NodeDefaults != nil {
		nodeCfg.CPU = cluster.Config.NodeDefaults.CPU
		nodeCfg.MemoryMB = cluster.Config.NodeDefaults.MemoryMB
		nodeCfg.DiskGB = cluster.Config.NodeDefaults.DiskGB
		if cluster.Config.NodeDefaults.Image != "" {
			nodeCfg.Image = cluster.Config.NodeDefaults.Image
		}
	}

	if nodeRef.Config != nil {
		if nodeRef.Config.CPU > 0 {
			nodeCfg.CPU = nodeRef.Config.CPU
		}
		if nodeRef.Config.MemoryMB > 0 {
			nodeCfg.MemoryMB = nodeRef.Config.MemoryMB
		}
		if nodeRef.Config.DiskGB > 0 {
			nodeCfg.DiskGB = nodeRef.Config.DiskGB
		}
	}

	// Create the VM
	hvNode, err := hv.CreateNode(ctx, nodeCfg)
	if err != nil {
		m.db.UpdateNodeState(nodeRef.ID, "error", "")
		return fmt.Errorf("failed to create VM: %w", err)
	}

	// Create database node from hypervisor response
	node := &database.Node{
		ID:        nodeRef.ID,
		ClusterID: cluster.ID,
		HostID:    nodeRef.HostID,
		Name:      hvNode.Name,
		Role:      nodeRef.Role,
		State:     "running",
		IP:        hvNode.IP,
	}

	// Convert hypervisor config to database config
	if hvNode.Config != nil {
		node.Config = &database.NodeConfig{
			CPU:      hvNode.Config.CPU,
			MemoryMB: hvNode.Config.MemoryMB,
			DiskGB:   hvNode.Config.DiskGB,
			Image:    hvNode.Config.Image,
		}
	}

	if err := m.db.SaveNode(node); err != nil {
		// Try to cleanup
		hv.DeleteNode(ctx, node.ID)
		return fmt.Errorf("failed to save node: %w", err)
	}

	// Update state to running
	m.db.UpdateNodeState(node.ID, "running", node.IP)

	return nil
}

// ScaleCluster scales a cluster up or down
func (m *Manager) ScaleCluster(ctx context.Context, clusterID string, desiredCount int) error {
	cluster, err := m.db.GetCluster(clusterID)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("cluster not found")
	}

	nodes, err := m.db.ListClusterNodes(clusterID)
	if err != nil {
		return err
	}

	currentCount := len(nodes)
	diff := desiredCount - currentCount

	if diff > 0 {
		// Scale up - add nodes
		for i := 0; i < diff; i++ {
			nodeRef := &database.Node{
				ID:        uuid.New().String(),
				ClusterID: clusterID,
				HostID:    nodes[i%len(nodes)].HostID, // Distribute across hosts
				Name:      fmt.Sprintf("%s-worker-%d", cluster.Name, currentCount+i+1),
				Role:      "worker",
				State:     "pending",
			}
			if err := m.db.SaveNode(nodeRef); err != nil {
				return err
			}
			if err := m.provisionNode(ctx, cluster, nodeRef); err != nil {
				return err
			}
		}
	} else if diff < 0 {
		// Scale down - remove nodes
		nodesToRemove := nodes[currentCount+diff:]
		for _, node := range nodesToRemove {
			if hv, ok := m.hosts[node.HostID]; ok {
				if err := hv.DeleteNode(ctx, node.ID); err != nil {
					fmt.Printf("Warning: failed to delete node %s: %v\n", node.Name, err)
				}
			}
			if err := m.db.DeleteNode(node.ID); err != nil {
				fmt.Printf("Warning: failed to remove node %s from db: %v\n", node.Name, err)
			}
		}
	}

	return nil
}

// ============== Node Operations ==============

// GetNode retrieves a node by ID
func (m *Manager) GetNode(id string) (*database.Node, error) {
	return m.db.GetNode(id)
}

// GetNodeStatus returns current status from hypervisor
func (m *Manager) GetNodeStatus(ctx context.Context, nodeID string) (*hypervisor.NodeStatus, error) {
	node, err := m.db.GetNode(nodeID)
	if err != nil || node == nil {
		return nil, fmt.Errorf("node not found")
	}

	hv, ok := m.hosts[node.HostID]
	if !ok {
		return nil, fmt.Errorf("host not found")
	}

	return hv.GetNodeStatus(ctx, nodeID)
}

// StartNode starts a VM
func (m *Manager) StartNode(ctx context.Context, nodeID string) error {
	node, err := m.db.GetNode(nodeID)
	if err != nil || node == nil {
		return fmt.Errorf("node not found")
	}

	hv, ok := m.hosts[node.HostID]
	if !ok {
		return fmt.Errorf("host not found")
	}

	if err := hv.StartNode(ctx, nodeID); err != nil {
		m.db.UpdateNodeState(nodeID, "error", node.IP)
		return err
	}

	// Get updated status to get IP
	status, err := hv.GetNodeStatus(ctx, nodeID)
	ip := node.IP
	if err == nil && status != nil {
		ip = status.IP
	}

	return m.db.UpdateNodeState(nodeID, "running", ip)
}

// StopNode stops a VM
func (m *Manager) StopNode(ctx context.Context, nodeID string) error {
	node, err := m.db.GetNode(nodeID)
	if err != nil || node == nil {
		return fmt.Errorf("node not found")
	}

	hv, ok := m.hosts[node.HostID]
	if !ok {
		return fmt.Errorf("host not found")
	}

	if err := hv.StopNode(ctx, nodeID); err != nil {
		m.db.UpdateNodeState(nodeID, "error", node.IP)
		return err
	}

	return m.db.UpdateNodeState(nodeID, "stopped", node.IP)
}

// RestartNode restarts a VM
func (m *Manager) RestartNode(ctx context.Context, nodeID string) error {
	node, err := m.db.GetNode(nodeID)
	if err != nil || node == nil {
		return fmt.Errorf("node not found")
	}

	hv, ok := m.hosts[node.HostID]
	if !ok {
		return fmt.Errorf("host not found")
	}

	if err := hv.RestartNode(ctx, nodeID); err != nil {
		m.db.UpdateNodeState(nodeID, "error", node.IP)
		return err
	}

	// Get updated status to get IP
	status, err := hv.GetNodeStatus(ctx, nodeID)
	ip := node.IP
	if err == nil && status != nil {
		ip = status.IP
	}

	return m.db.UpdateNodeState(nodeID, "running", ip)
}

// DeleteNode deletes a VM
func (m *Manager) DeleteNode(ctx context.Context, nodeID string) error {
	node, err := m.db.GetNode(nodeID)
	if err != nil || node == nil {
		return fmt.Errorf("node not found")
	}

	hv, ok := m.hosts[node.HostID]
	if ok {
		if err := hv.DeleteNode(ctx, nodeID); err != nil {
			return fmt.Errorf("failed to delete from hypervisor: %w", err)
		}
	}

	return m.db.DeleteNode(nodeID)
}

// AddNode adds a new node to a cluster
func (m *Manager) AddNode(ctx context.Context, clusterID, hostID, name, role string, cfg *database.NodeConfig) (*database.Node, error) {
	cluster, err := m.db.GetCluster(clusterID)
	if err != nil || cluster == nil {
		return nil, fmt.Errorf("cluster not found")
	}

	node := &database.Node{
		ID:        uuid.New().String(),
		ClusterID: clusterID,
		HostID:    hostID,
		Name:      name,
		Role:      role,
		State:     "pending",
		Config:    cfg,
	}

	if err := m.db.SaveNode(node); err != nil {
		return nil, err
	}

	// Provision the node
	nodeRef := &database.Node{
		ID:        node.ID,
		ClusterID: clusterID,
		HostID:    hostID,
		Name:      name,
		Role:      role,
		Config:    cfg,
	}

	if err := m.provisionNode(ctx, cluster, nodeRef); err != nil {
		m.db.DeleteNode(node.ID)
		return nil, err
	}

	return m.db.GetNode(node.ID)
}

// GetOrCreateHost gets or creates a host connection
func (m *Manager) GetOrCreateHost(cfg *database.Host) (hypervisor.Hypervisor, error) {
	// Check if we already have a connection
	if hv, ok := m.hosts[cfg.ID]; ok {
		return hv, nil
	}

	// Save host to database if not exists
	if _, err := m.db.GetHost(cfg.ID); err == nil {
		// Host exists
	} else {
		m.db.SaveHost(cfg)
	}

	// Create hypervisor based on type
	switch cfg.HVType {
	case "libvirt", "qemu", "kvm":
		return hypervisor.NewHypervisor(&hypervisor.HostConfig{
			Address:    cfg.Address,
			Port:       cfg.Port,
			User:       cfg.User,
			Type:       cfg.HVType,
		})
	case "stub", "mock", "test":
		return hypervisor.NewStubHypervisor(), nil
	default:
		return nil, fmt.Errorf("unsupported hypervisor type: %s", cfg.HVType)
	}
}

// ListHosts returns all configured hosts
func (m *Manager) ListHosts() ([]*database.Host, error) {
	return m.db.ListHosts()
}

// NodeStats returns statistics for a cluster
func (m *Manager) NodeStats(clusterID string) (total, running, stopped, error int) {
	nodes, err := m.db.ListClusterNodes(clusterID)
	if err != nil {
		return 0, 0, 0, 0
	}

	for _, n := range nodes {
		total++
		switch n.State {
		case "running":
			running++
		case "stopped":
			stopped++
		case "error":
			error++
		}
	}
	return
}

// GetClusterMetrics returns aggregated metrics for a cluster
func (m *Manager) GetClusterMetrics(clusterID string, since time.Duration) (cpuAvg, memAvg, diskAvg float64, err error) {
	nodes, err := m.db.ListClusterNodes(clusterID)
	if err != nil {
		return 0, 0, 0, err
	}

	var totalCPU, totalMem, totalDisk float64
	var count int

	for _, node := range nodes {
		if node.State != "running" {
			continue
		}

		metrics, err := m.db.GetNodeMetrics(node.ID, time.Now().Add(-since))
		if err != nil || len(metrics) == 0 {
			continue
		}

		// Get latest metric for this node
		latest := metrics[len(metrics)-1]
		totalCPU += latest.CPU
		totalMem += latest.Memory
		totalDisk += latest.Disk
		count++
	}

	if count > 0 {
		cpuAvg = totalCPU / float64(count)
		memAvg = totalMem / float64(count)
		diskAvg = totalDisk / float64(count)
	}

	return cpuAvg, memAvg, diskAvg, nil
}
