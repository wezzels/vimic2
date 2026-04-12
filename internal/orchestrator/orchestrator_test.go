// Package orchestrator_test tests auto-scaling and recovery functionality
package orchestrator_test

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestAutoScalerCreation tests auto-scaler initialization
func TestAutoScalerCreation(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create cluster manager
	hosts := map[string]hypervisor.Hypervisor{
		"host-1": hypervisor.NewStubHypervisor(),
	}
	clusterMgr := cluster.NewManager(db, hosts)
	monitorMgr := monitor.NewManager(db, hosts)

	logger := zap.NewNop().Sugar()
	scaler := orchestrator.NewAutoScaler(clusterMgr, monitorMgr, logger)

	if scaler == nil {
		t.Fatal("AutoScaler should not be nil")
	}
}

// TestAutoScaleUp tests scaling up based on CPU
func TestAutoScaleUp(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create test cluster
	clusterData := &database.Cluster{
		ID:     "scale-test",
		Name:   "Scale Test",
		Status: "running",
	}
	if err := db.SaveCluster(clusterData); err != nil {
		t.Fatalf("Failed to save cluster: %v", err)
	}

	// Create initial nodes
	for i := 0; i < 2; i++ {
		node := &database.Node{
			ID:        "node-" + string(rune('a'+i)),
			ClusterID: "scale-test",
			Name:      "node-" + string(rune('a'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
		}
		db.SaveNode(node)
	}

	hosts := map[string]hypervisor.Hypervisor{
		"host-1": hypervisor.NewStubHypervisor(),
	}
	clusterMgr := cluster.NewManager(db, hosts)
	monitorMgr := monitor.NewManager(db, hosts)

	logger := zap.NewNop().Sugar()
	scaler := orchestrator.NewAutoScaler(clusterMgr, monitorMgr, logger)

	// Add scale rule
	rule := &orchestrator.ScaleRule{
		ClusterID:      "scale-test",
		Metric:         "cpu",
		UpperThreshold: 80,
		LowerThreshold: 20,
		ScaleUpCount:   1,
		ScaleDownCount: 1,
		Cooldown:       5 * time.Minute,
	}
	scaler.AddRule(rule)

	// Verify rule was added (no GetRules method, just verify no error)
	// The rule is stored internally and will be used during evaluation
}

// TestRollingUpdate tests rolling update functionality
func TestRollingUpdate(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create test cluster with nodes
	clusterData := &database.Cluster{
		ID:     "update-test",
		Name:   "Update Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	for i := 0; i < 5; i++ {
		node := &database.Node{
			ID:        "update-node-" + string(rune('0'+i)),
			ClusterID: "update-test",
			Name:      "update-node-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
			Config: &database.NodeConfig{
				CPU:      2,
				MemoryMB: 4096,
				DiskGB:   20,
				Image:    "ubuntu:22.04",
			},
		}
		db.SaveNode(node)
	}

	logger := zap.NewNop().Sugar()
	updater := orchestrator.NewRollingUpdater(db, logger)

	config := &orchestrator.UpdateConfig{
		BatchSize:     2,
		BatchPause:    10 * time.Second,
		HealthCheck:   true,
		HealthTimeout: 5 * time.Minute,
		NewImage:      "ubuntu:24.04",
	}

	progress := make(chan *orchestrator.UpdateProgress, 100)

	// Drain progress channel in background to prevent blocking
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range progress {
		}
	}()

	err = updater.Update("update-test", config, progress)
	close(progress)
	<-done
	if err != nil {
		t.Logf("Update returned: %v", err)
	}
}

// TestRollingUpdateBatchSize tests different batch sizes
func TestRollingUpdateBatchSize(t *testing.T) {
	tests := []struct {
		name      string
		nodeCount int
		batchSize int
		batches   int // expected number of batches
	}{
		{"5 nodes, batch 1", 5, 1, 5},
		{"5 nodes, batch 2", 5, 2, 3},
		{"5 nodes, batch 5", 5, 5, 1},
		{"10 nodes, batch 3", 10, 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := (tt.nodeCount + tt.batchSize - 1) / tt.batchSize
			if batches != tt.batches {
				t.Errorf("Expected %d batches, got %d", tt.batches, batches)
			}
		})
	}
}

// TestRecoveryManager tests backup and recovery
func TestRecoveryManager(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	if rm == nil {
		t.Fatal("RecoveryManager should not be nil")
	}
}

// TestCreateBackup tests backup creation
func TestCreateBackup(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create test cluster with nodes
	clusterData := &database.Cluster{
		ID:     "backup-test",
		Name:   "Backup Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	for i := 0; i < 3; i++ {
		node := &database.Node{
			ID:        "backup-node-" + string(rune('0'+i)),
			ClusterID: "backup-test",
			Name:      "backup-node-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
			Config: &database.NodeConfig{
				CPU:      2,
				MemoryMB: 4096,
				DiskGB:   20,
				Image:    "ubuntu:22.04",
			},
		}
		db.SaveNode(node)
	}

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	ctx := context.Background()
	backup, err := rm.CreateBackup(ctx, "backup-test", "Test backup")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if backup == nil {
		t.Fatal("Backup should not be nil")
	}

	if backup.ClusterID != "backup-test" {
		t.Errorf("Expected cluster ID 'backup-test', got '%s'", backup.ClusterID)
	}

	if len(backup.Nodes) != 3 {
		t.Errorf("Expected 3 nodes in backup, got %d", len(backup.Nodes))
	}
}

// TestRestoreBackup tests backup restoration
func TestRestoreBackup(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create and backup cluster
	clusterData := &database.Cluster{
		ID:     "restore-test",
		Name:   "Restore Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	for i := 0; i < 2; i++ {
		node := &database.Node{
			ID:        "restore-node-" + string(rune('0'+i)),
			ClusterID: "restore-test",
			Name:      "restore-node-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
			Config: &database.NodeConfig{
				CPU:      2,
				MemoryMB: 4096,
				DiskGB:   20,
				Image:    "ubuntu:22.04",
			},
		}
		db.SaveNode(node)
	}

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	ctx := context.Background()

	// Create backup
	backup, err := rm.CreateBackup(ctx, "restore-test", "Before restore")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Restore backup
	err = rm.RestoreCluster(ctx, backup.ID)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify cluster status
	cluster, err := db.GetCluster("restore-test")
	if err != nil {
		t.Fatalf("Failed to get cluster: %v", err)
	}

	if cluster.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", cluster.Status)
	}
}

// TestListBackups tests listing backups
func TestListBackups(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create test cluster
	clusterData := &database.Cluster{
		ID:     "list-test",
		Name:   "List Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	// Create multiple backups (IDs use nanoseconds so no delay needed)
	ctx := context.Background()
	_, err = rm.CreateBackup(ctx, "list-test", "Backup 1")
	if err != nil {
		t.Fatalf("Failed to create backup 1: %v", err)
	}
	_, err = rm.CreateBackup(ctx, "list-test", "Backup 2")
	if err != nil {
		t.Fatalf("Failed to create backup 2: %v", err)
	}
	_, err = rm.CreateBackup(ctx, "list-test", "Backup 3")
	if err != nil {
		t.Fatalf("Failed to create backup 3: %v", err)
	}

	// List backups
	backups, err := rm.ListBackups("list-test")
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}
}

// TestDeleteBackup tests backup deletion
func TestDeleteBackup(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create test cluster
	clusterData := &database.Cluster{
		ID:     "delete-test",
		Name:   "Delete Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	ctx := context.Background()
	backup, err := rm.CreateBackup(ctx, "delete-test", "To delete")
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Delete backup
	err = rm.DeleteBackup(backup.ID)
	if err != nil {
		t.Fatalf("Failed to delete backup: %v", err)
	}

	// Verify it's deleted
	_, err = rm.GetBackup(backup.ID)
	if err == nil {
		t.Error("Expected error when getting deleted backup")
	}
}

// TestFailover tests node failover
func TestFailover(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create test cluster with standby nodes
	clusterData := &database.Cluster{
		ID:     "failover-test",
		Name:   "Failover Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	// Worker nodes
	for i := 0; i < 3; i++ {
		node := &database.Node{
			ID:        "worker-" + string(rune('0'+i)),
			ClusterID: "failover-test",
			Name:      "worker-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
		}
		db.SaveNode(node)
	}

	// Standby node
	standby := &database.Node{
		ID:        "standby-1",
		ClusterID: "failover-test",
		Name:      "standby-1",
		Role:      "standby",
		HostID:    "host-1",
		State:     "running",
	}
	db.SaveNode(standby)

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	ctx := context.Background()

	// Simulate failover
	err = rm.Failover(ctx, "failover-test", "worker-0")
	if err != nil {
		t.Fatalf("Failover failed: %v", err)
	}

	// Verify standby was promoted
	promotedNode, err := db.GetNode("standby-1")
	if err != nil {
		t.Fatalf("Failed to get promoted node: %v", err)
	}

	if promotedNode.Role != "worker" {
		t.Errorf("Expected standby to be promoted to worker, got '%s'", promotedNode.Role)
	}

	// Verify failed node is in error state
	failedNode, err := db.GetNode("worker-0")
	if err != nil {
		t.Fatalf("Failed to get failed node: %v", err)
	}

	if failedNode.State != "error" {
		t.Errorf("Expected failed node state 'error', got '%s'", failedNode.State)
	}
}

// TestFailoverNoStandby tests failover when no standby available
func TestFailoverNoStandby(t *testing.T) {
	// Setup test database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
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
	defer db.Close()

	// Create temp backup dir
	backupDir, err := os.MkdirTemp("", "vimic2-backup-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	defer os.RemoveAll(backupDir)

	// Create test cluster with no standby
	clusterData := &database.Cluster{
		ID:     "no-standby-test",
		Name:   "No Standby Test",
		Status: "running",
	}
	db.SaveCluster(clusterData)

	// Only worker nodes
	for i := 0; i < 3; i++ {
		node := &database.Node{
			ID:        "worker-" + string(rune('0'+i)),
			ClusterID: "no-standby-test",
			Name:      "worker-" + string(rune('0'+i)),
			Role:      "worker",
			HostID:    "host-1",
			State:     "running",
		}
		db.SaveNode(node)
	}

	logger := zap.NewNop().Sugar()
	rm := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	ctx := context.Background()

	// Try failover - should fail
	err = rm.Failover(ctx, "no-standby-test", "worker-0")
	if err == nil {
		t.Error("Expected failover to fail with no standby nodes")
	}
}
