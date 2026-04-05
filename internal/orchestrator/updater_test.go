// Package orchestrator_test tests rolling update functionality
package orchestrator_test

import (
	"context"
	"os"
	"testing"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"go.uber.org/zap"
)

// createTestUpdater creates a test rolling updater
func createTestUpdater(t *testing.T) (*orchestrator.RollingUpdater, *database.DB) {
	tmpFile, err := os.CreateTemp("", "vimic2-updater-test-*.db")
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
	stubHV := cluster.NewStubHypervisor()
	hosts := map[string]cluster.Hypervisor{
		"test-host": stubHV,
	}
	clusterMgr := cluster.NewManager(db, hosts)

	// Create logger
	sugar := zap.NewExample().Sugar()

	updater := orchestrator.NewRollingUpdater(clusterMgr, sugar)

	return updater, db
}

func TestRollingUpdaterCreate(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	if updater == nil {
		t.Fatal("Updater should not be nil")
	}
}

func TestRollingUpdaterUpdateStrategy(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	strategy := &orchestrator.UpdateStrategy{
		BatchSize:      2,
		MaxUnavailable: 1,
		WaitBetween:    5,
	}

	// Test setting strategy
	updater.SetStrategy(strategy)

	// Verify nodes are handled correctly
	nodes := []string{"node-1", "node-2", "node-3", "node-4", "node-5"}
	
	// Calculate batches
	batches := updater.CalculateBatches(nodes, 2)
	if len(batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(batches))
	}

	if len(batches[0]) != 2 {
		t.Errorf("Expected batch 0 size 2, got %d", len(batches[0]))
	}
}

func TestRollingUpdaterBatching(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	tests := []struct {
		nodes     []string
		batchSize int
		expected  int
	}{
		{[]string{"n1", "n2", "n3"}, 2, 2},
		{[]string{"n1", "n2", "n3", "n4"}, 2, 2},
		{[]string{"n1", "n2", "n3", "n4", "n5"}, 2, 3},
		{[]string{"n1"}, 2, 1},
		{[]string{}, 2, 0},
	}

	for _, tt := range tests {
		batches := updater.CalculateBatches(tt.nodes, tt.batchSize)
		if len(batches) != tt.expected {
			t.Errorf("For %d nodes with batch size %d, expected %d batches, got %d",
				len(tt.nodes), tt.batchSize, tt.expected, len(batches))
		}
	}
}

func TestRollingUpdaterHealthCheck(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	ctx := context.Background()

	// Test with valid node
	healthy, err := updater.IsNodeHealthy(ctx, "test-node")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Stub hypervisor should return healthy
	if !healthy {
		t.Error("Node should be healthy")
	}
}

func TestRollingUpdaterDrainNode(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	ctx := context.Background()

	// Test draining a node
	err := updater.DrainNode(ctx, "test-node")
	if err != nil {
		t.Errorf("DrainNode failed: %v", err)
	}
}

func TestRollingUpdaterUpgradeNode(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	ctx := context.Background()

	// Test upgrading a node
	err := updater.UpgradeNode(ctx, "test-node", "v2.0.0")
	if err != nil {
		t.Errorf("UpgradeNode failed: %v", err)
	}
}

func TestRollingUpdaterRestoreNode(t *testing.T) {
	updater, db := createTestUpdater(t)
	defer db.Close()

	ctx := context.Background()

	// Test restoring a node after upgrade
	err := updater.RestoreNode(ctx, "test-node")
	if err != nil {
		t.Errorf("RestoreNode failed: %v", err)
	}
}

func TestUpdateStrategyValidation(t *testing.T) {
	tests := []struct {
		name      string
		strategy  *orchestrator.UpdateStrategy
		wantValid bool
	}{
		{
			name: "valid strategy",
			strategy: &orchestrator.UpdateStrategy{
				BatchSize:      2,
				MaxUnavailable: 1,
				WaitBetween:    5,
			},
			wantValid: true,
		},
		{
			name: "zero batch size",
			strategy: &orchestrator.UpdateStrategy{
				BatchSize:      0,
				MaxUnavailable: 1,
				WaitBetween:    5,
			},
			wantValid: false,
		},
		{
			name: "negative wait",
			strategy: &orchestrator.UpdateStrategy{
				BatchSize:      2,
				MaxUnavailable: 1,
				WaitBetween:    -1,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.strategy.IsValid()
			if valid != tt.wantValid {
				t.Errorf("Expected valid=%v, got %v", tt.wantValid, valid)
			}
		})
	}
}
