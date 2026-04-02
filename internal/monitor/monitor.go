// Package monitor provides metrics collection and alerting
package monitor

import (
	"context"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// Manager handles metrics collection
type Manager struct {
	db    *database.DB
	hosts map[string]hypervisor.Hypervisor
}

// NewManager creates a new monitor manager
func NewManager(db *database.DB, hosts map[string]hypervisor.Hypervisor) *Manager {
	return &Manager{
		db:    db,
		hosts: hosts,
	}
}

// StartCollection starts background metrics collection
func (m *Manager) StartCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectAll()
		case <-ctx.Done():
			return
		}
	}
}

func (m *Manager) collectAll() {
	clusters, _ := m.db.ListClusters()
	for _, cluster := range clusters {
		nodes, _ := m.db.ListClusterNodes(cluster.ID)
		for _, node := range nodes {
			if node.State != "running" {
				continue
			}
			m.collectNode(node)
		}
	}
}

func (m *Manager) collectNode(node *database.Node) {
	hv, ok := m.hosts[node.HostID]
	if !ok {
		return
	}

	metrics, err := hv.GetMetrics(context.Background(), node.ID)
	if err != nil {
		return
	}

	// Save to database
	m.db.SaveMetric(&database.Metric{
		NodeID:     node.ID,
		CPU:        metrics.CPU,
		Memory:     metrics.Memory,
		Disk:       metrics.Disk,
		NetworkRX:  metrics.NetworkRX,
		NetworkTX:  metrics.NetworkTX,
		RecordedAt: time.Now(),
	})
}

// GetNodeMetrics returns metrics for a node
func (m *Manager) GetNodeMetrics(nodeID string, since time.Duration) ([]*database.Metric, error) {
	return m.db.GetNodeMetrics(nodeID, time.Now().Add(-since))
}

// GetClusterMetrics returns aggregated cluster metrics
func (m *Manager) GetClusterMetrics(clusterID string, since time.Duration) (*ClusterMetrics, error) {
	nodes, err := m.db.ListClusterNodes(clusterID)
	if err != nil {
		return nil, err
	}

	cm := &ClusterMetrics{
		NodeMetrics: make(map[string][]*database.Metric),
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

		cm.NodeMetrics[node.ID] = metrics

		// Calculate averages
		for _, m := range metrics {
			totalCPU += m.CPU
			totalMem += m.Memory
			totalDisk += m.Disk
			count++
		}
	}

	if count > 0 {
		cm.CPUAvg = totalCPU / float64(count)
		cm.MemAvg = totalMem / float64(count)
		cm.DiskAvg = totalDisk / float64(count)
	}

	return cm, nil
}

// ClusterMetrics holds aggregated metrics for a cluster
type ClusterMetrics struct {
	NodeMetrics map[string][]*database.Metric
	CPUAvg      float64
	MemAvg      float64
	DiskAvg     float64
}
