// Package orchestrator_test tests auto-scaling functionality
package orchestrator_test

import (
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"github.com/stsgym/vimic2/pkg/hypervisor"
	"go.uber.org/zap"
)

// createTestAutoScaler creates a test auto-scaler with mocks
func createTestAutoScaler(t *testing.T) (*orchestrator.AutoScaler, *database.DB) {
	tmpFile, err := os.CreateTemp("", "vimic2-autoscaler-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
		Config: &database.ClusterConfig{
			MinNodes:  1,
			MaxNodes:  10,
			AutoScale: true,
		},
	}
	db.SaveCluster(testCluster)

	// Create stub hypervisor and cluster manager
	stubHV := hypervisor.NewStubHypervisor()
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": stubHV,
	}
	clusterMgr := cluster.NewManager(db, hosts)
	
	// Create monitor manager
	monitorMgr := monitor.NewManager(db, nil)
	
	// Create logger
	sugar := zap.NewExample().Sugar()

	autoScaler := orchestrator.NewAutoScaler(clusterMgr, monitorMgr, sugar)
	autoScaler.SetDB(db)

	return autoScaler, db
}

func TestAutoScalerAddRule(t *testing.T) {
	autoScaler, db := createTestAutoScaler(t)
	defer db.Close()

	rule := &orchestrator.ScaleRule{
		ClusterID:      "test-cluster",
		Metric:         "cpu",
		UpperThreshold: 80.0,
		LowerThreshold: 20.0,
		ScaleUpCount:   2,
		ScaleDownCount: 1,
		Cooldown:       5 * time.Minute,
		Enabled:        true,
	}

	autoScaler.AddRule(rule)

	retrieved := autoScaler.GetRule("test-cluster")
	if retrieved == nil {
		t.Fatal("Rule should not be nil")
	}
	if retrieved.UpperThreshold != 80.0 {
		t.Errorf("Expected UpperThreshold 80.0, got %f", retrieved.UpperThreshold)
	}
	if retrieved.Metric != "cpu" {
		t.Errorf("Expected Metric 'cpu', got '%s'", retrieved.Metric)
	}
}

func TestAutoScalerRemoveRule(t *testing.T) {
	autoScaler, db := createTestAutoScaler(t)
	defer db.Close()

	rule := &orchestrator.ScaleRule{
		ClusterID:      "test-cluster",
		Metric:         "cpu",
		UpperThreshold: 80.0,
		LowerThreshold: 20.0,
		Enabled:        true,
	}

	autoScaler.AddRule(rule)
	autoScaler.RemoveRule("test-cluster")

	retrieved := autoScaler.GetRule("test-cluster")
	if retrieved != nil {
		t.Error("Rule should be nil after removal")
	}
}

func TestAutoScalerStartStop(t *testing.T) {
	autoScaler, db := createTestAutoScaler(t)
	defer db.Close()

	// Add a rule
	rule := &orchestrator.ScaleRule{
		ClusterID:      "test-cluster",
		Metric:         "cpu",
		UpperThreshold: 80.0,
		LowerThreshold: 20.0,
		Enabled:        false, // Disabled so it doesn't actually try to scale
	}
	autoScaler.AddRule(rule)

	// Start and immediately stop
	autoScaler.Start(100 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	autoScaler.Stop()

	// If we get here without hanging, the test passed
}

func TestAutoScalerEvaluate(t *testing.T) {
	autoScaler, db := createTestAutoScaler(t)
	defer db.Close()

	// Add nodes to the cluster
	for i := 0; i < 3; i++ {
		node := &database.Node{
			ID:        "node-" + string(rune('a'+i)),
			ClusterID: "test-cluster",
			Name:      "node-" + string(rune('a'+i)),
			Role:      "worker",
			HostID:    "test-host",
			State:     "running",
		}
		db.SaveNode(node)
	}

	// Add a disabled rule (we'll manually test evaluation)
	rule := &orchestrator.ScaleRule{
		ClusterID:      "test-cluster",
		Metric:         "cpu",
		UpperThreshold: 80.0,
		LowerThreshold: 20.0,
		ScaleUpCount:   1,
		ScaleDownCount: 1,
		Cooldown:       1 * time.Second,
		Enabled:        false,
	}
	autoScaler.AddRule(rule)

	// Test that evaluation doesn't panic with disabled rule
	autoScaler.Start(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	autoScaler.Stop()
}

func TestScaleRuleDefaults(t *testing.T) {
	// Test that cooldown defaults are set correctly
	rule := &orchestrator.ScaleRule{
		ClusterID:      "test",
		UpperThreshold: 80,
		LowerThreshold: 20,
	}

	if !rule.LastScaleUp.IsZero() {
		t.Error("LastScaleUp should be zero initially")
	}
	if !rule.LastScaleDown.IsZero() {
		t.Error("LastScaleDown should be zero initially")
	}
}

func TestAutoScalerMultipleClusters(t *testing.T) {
	autoScaler, db := createTestAutoScaler(t)
	defer db.Close()

	// Create additional clusters
	for i := 0; i < 3; i++ {
		clusterID := "cluster-" + string(rune('0'+i))
		testCluster := &database.Cluster{
			ID:     clusterID,
			Name:   "Cluster " + string(rune('0'+i)),
			Status: "running",
			Config: &database.ClusterConfig{
				MinNodes:  1,
				MaxNodes:  5,
				AutoScale: true,
			},
		}
		db.SaveCluster(testCluster)

		rule := &orchestrator.ScaleRule{
			ClusterID:      clusterID,
			Metric:         "cpu",
			UpperThreshold: 80.0,
			LowerThreshold: 20.0,
			Enabled:        true,
		}
		autoScaler.AddRule(rule)
	}

	// Verify all rules exist
	for i := 0; i < 3; i++ {
		clusterID := "cluster-" + string(rune('0'+i))
		rule := autoScaler.GetRule(clusterID)
		if rule == nil {
			t.Errorf("Rule for %s should exist", clusterID)
		}
	}
}
