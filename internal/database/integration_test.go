// Package database provides integration tests
package database

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestIntegration_FullWorkflow tests complete database workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-integration-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// === HOSTS ===
	t.Run("HostsWorkflow", func(t *testing.T) {
		hosts := []*Host{
			{ID: "host-1", Name: "hv1", Address: "192.168.1.1", Port: 22, User: "admin", HVType: "libvirt"},
			{ID: "host-2", Name: "hv2", Address: "192.168.1.2", Port: 22, User: "admin", HVType: "libvirt"},
			{ID: "host-3", Name: "hv3", Address: "192.168.1.3", Port: 22, User: "admin", HVType: "hyperv"},
		}

		for _, h := range hosts {
			if err := db.SaveHost(h); err != nil {
				t.Fatalf("failed to save host %s: %v", h.ID, err)
			}
		}

		allHosts, err := db.ListHosts()
		if err != nil {
			t.Fatalf("failed to list hosts: %v", err)
		}
		if len(allHosts) != 3 {
			t.Errorf("expected 3 hosts, got %d", len(allHosts))
		}

		for _, h := range hosts {
			retrieved, err := db.GetHost(h.ID)
			if err != nil {
				t.Fatalf("failed to get host %s: %v", h.ID, err)
			}
			if retrieved.Name != h.Name {
				t.Errorf("host name mismatch: expected %s, got %s", h.Name, retrieved.Name)
			}
		}

		if err := db.DeleteHost("host-3"); err != nil {
			t.Fatalf("failed to delete host: %v", err)
		}

		remaining, _ := db.ListHosts()
		if len(remaining) != 2 {
			t.Errorf("expected 2 hosts after delete, got %d", len(remaining))
		}
	})

	// === CLUSTERS ===
	t.Run("ClustersWorkflow", func(t *testing.T) {
		clusters := []*Cluster{
			{ID: "cluster-1", Name: "prod", Status: "running", Config: &ClusterConfig{MinNodes: 3, MaxNodes: 10}},
			{ID: "cluster-2", Name: "dev", Status: "running", Config: &ClusterConfig{MinNodes: 1, MaxNodes: 5}},
		}

		for _, c := range clusters {
			if err := db.SaveCluster(c); err != nil {
				t.Fatalf("failed to save cluster %s: %v", c.ID, err)
			}
		}

		allClusters, err := db.ListClusters()
		if err != nil {
			t.Fatalf("failed to list clusters: %v", err)
		}
		if len(allClusters) != 2 {
			t.Errorf("expected 2 clusters, got %d", len(allClusters))
		}

		prod, err := db.GetCluster("cluster-1")
		if err != nil {
			t.Fatalf("failed to get cluster: %v", err)
		}
		if prod.Config == nil {
			t.Fatal("expected config to be non-nil")
		}
		if prod.Config.MinNodes != 3 {
			t.Errorf("expected MinNodes 3, got %d", prod.Config.MinNodes)
		}

		if err := db.UpdateClusterStatus("cluster-2", "stopped"); err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		dev, _ := db.GetCluster("cluster-2")
		if dev.Status != "stopped" {
			t.Errorf("expected stopped, got %s", dev.Status)
		}

		if err := db.DeleteCluster("cluster-2"); err != nil {
			t.Fatalf("failed to delete cluster: %v", err)
		}
	})

	// === NODES ===
	t.Run("NodesWorkflow", func(t *testing.T) {
		cluster := &Cluster{ID: "cluster-nodes", Name: "test"}
		db.SaveCluster(cluster)

		nodes := []*Node{
			{ID: "node-1", Name: "worker-1", ClusterID: "cluster-nodes", State: "running", Role: "worker"},
			{ID: "node-2", Name: "worker-2", ClusterID: "cluster-nodes", State: "running", Role: "worker"},
			{ID: "node-3", Name: "master-1", ClusterID: "cluster-nodes", State: "running", Role: "master"},
		}

		for _, n := range nodes {
			if err := db.SaveNode(n); err != nil {
				t.Fatalf("failed to save node %s: %v", n.ID, err)
			}
		}

		clusterNodes, err := db.ListClusterNodes("cluster-nodes")
		if err != nil {
			t.Fatalf("failed to list cluster nodes: %v", err)
		}
		if len(clusterNodes) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(clusterNodes))
		}

		allNodes, err := db.ListAllNodes()
		if err != nil {
			t.Fatalf("failed to list all nodes: %v", err)
		}
		if len(allNodes) != 3 {
			t.Errorf("expected 3 total nodes, got %d", len(allNodes))
		}

		node, err := db.GetNode("node-1")
		if err != nil {
			t.Fatalf("failed to get node: %v", err)
		}
		if node.Role != "worker" {
			t.Errorf("expected worker role, got %s", node.Role)
		}

		if err := db.UpdateNodeState("node-1", "stopped", "Manual stop"); err != nil {
			t.Fatalf("failed to update node state: %v", err)
		}

		updated, _ := db.GetNode("node-1")
		if updated.State != "stopped" {
			t.Errorf("expected stopped, got %s", updated.State)
		}

		if err := db.DeleteNode("node-3"); err != nil {
			t.Fatalf("failed to delete node: %v", err)
		}

		remaining, _ := db.ListClusterNodes("cluster-nodes")
		if len(remaining) != 2 {
			t.Errorf("expected 2 nodes after delete, got %d", len(remaining))
		}
	})

	// === METRICS ===
	t.Run("MetricsWorkflow", func(t *testing.T) {
		now := time.Now()
		metrics := []*Metric{
			{NodeID: "node-1", CPU: 45.5, Memory: 60.2, Disk: 30.1, NetworkRX: 1024, NetworkTX: 512, RecordedAt: now},
			{NodeID: "node-1", CPU: 50.0, Memory: 62.0, Disk: 30.5, NetworkRX: 2048, NetworkTX: 1024, RecordedAt: now.Add(1 * time.Minute)},
			{NodeID: "node-1", CPU: 55.5, Memory: 65.0, Disk: 31.0, NetworkRX: 3072, NetworkTX: 1536, RecordedAt: now.Add(2 * time.Minute)},
		}

		for _, m := range metrics {
			if err := db.SaveMetric(m); err != nil {
				t.Fatalf("failed to save metric: %v", err)
			}
		}

		latest, err := db.GetLatestMetric("node-1")
		if err != nil {
			t.Fatalf("failed to get latest metric: %v", err)
		}
		if latest == nil {
			t.Fatal("expected non-nil metric")
		}

		nodeMetrics, err := db.GetNodeMetrics("node-1", time.Now().Add(-1*time.Hour))
		if err != nil {
			t.Fatalf("failed to get node metrics: %v", err)
		}
		if len(nodeMetrics) < 3 {
			t.Errorf("expected at least 3 metrics, got %d", len(nodeMetrics))
		}

		count, err := db.CleanupOldMetrics(24 * time.Hour)
		if err != nil {
			t.Fatalf("failed to cleanup metrics: %v", err)
		}
		_ = count
	})

	// === ALERTS ===
	t.Run("AlertsWorkflow", func(t *testing.T) {
		alerts := []*Alert{
			{ID: "alert-1", NodeID: "node-1", Metric: "cpu", Value: 85.0, Threshold: 80.0, Message: "High CPU", FiredAt: time.Now()},
			{ID: "alert-2", NodeID: "node-2", Metric: "memory", Value: 92.0, Threshold: 90.0, Message: "High Memory", FiredAt: time.Now()},
			{ID: "alert-3", NodeID: "node-1", Metric: "disk", Value: 95.0, Threshold: 85.0, Message: "High Disk", FiredAt: time.Now()},
		}

		for _, a := range alerts {
			if err := db.SaveAlert(a); err != nil {
				t.Fatalf("failed to save alert %s: %v", a.ID, err)
			}
		}

		active, err := db.GetActiveAlerts()
		if err != nil {
			t.Fatalf("failed to get active alerts: %v", err)
		}
		if len(active) < 3 {
			t.Errorf("expected at least 3 alerts, got %d", len(active))
		}

		nodeAlerts, err := db.GetNodeAlerts("node-1")
		if err != nil {
			t.Fatalf("failed to get node alerts: %v", err)
		}
		if len(nodeAlerts) < 2 {
			t.Errorf("expected at least 2 alerts for node-1, got %d", len(nodeAlerts))
		}
	})
}

// TestIntegration_ConcurrentOperations tests concurrent database access
func TestIntegration_ConcurrentOperations(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-concurrent-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	db.SaveCluster(&Cluster{ID: "cluster-concurrent", Name: "test"})

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				node := &Node{
					ID:        "node-" + string(rune('A'+idx)) + "-" + string(rune('0'+j)),
					Name:      "node",
					ClusterID: "cluster-concurrent",
					State:     "running",
				}
				if err := db.SaveNode(node); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := db.ListClusters()
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent operation failed: %v", err)
		}
	}
}

// TestIntegration_LargeDataset tests with many records
func TestIntegration_LargeDataset(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-large-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	db.SaveCluster(&Cluster{ID: "cluster-large", Name: "large-test"})

	start := time.Now()
	for i := 0; i < 100; i++ {
		nodeID := "node-" + string(rune('A'+i/26)) + string(rune('A'+i%26))
		node := &Node{
			ID:        nodeID,
			Name:      "node",
			ClusterID: "cluster-large",
			State:     "running",
		}
		if err := db.SaveNode(node); err != nil {
			t.Fatalf("failed to save node %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	t.Logf("Inserted 100 nodes in %v", elapsed)

	nodes, err := db.ListAllNodes()
	if err != nil {
		t.Fatalf("failed to list nodes: %v", err)
	}
	if len(nodes) != 100 {
		t.Errorf("expected 100 nodes, got %d", len(nodes))
	}

	start = time.Now()
	for i := 0; i < 1000; i++ {
		metric := &Metric{
			NodeID:     nodes[i%100].ID,
			CPU:        float64(i % 100),
			Memory:     float64((i + 20) % 100),
			RecordedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := db.SaveMetric(metric); err != nil {
			t.Fatalf("failed to save metric %d: %v", i, err)
		}
	}
	elapsed = time.Since(start)
	t.Logf("Inserted 1000 metrics in %v", elapsed)

	metrics, err := db.GetNodeMetrics(nodes[0].ID, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	if len(metrics) < 10 {
		t.Errorf("expected at least 10 metrics, got %d", len(metrics))
	}
}
