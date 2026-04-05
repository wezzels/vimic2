// Package orchestrator_test tests backup/restore functionality
package orchestrator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/orchestrator"
	"go.uber.org/zap"
)

// createTestRecoveryManager creates a test recovery manager
func createTestRecoveryManager(t *testing.T) (*orchestrator.RecoveryManager, *database.DB, string) {
	tmpFile, err := os.CreateTemp("", "vimic2-recovery-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Create temp backup directory
	backupDir, err := os.MkdirTemp("", "vimic2-backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	logger := zap.NewExample().Sugar()
	manager := orchestrator.NewRecoveryManagerWithLogger(db, backupDir, logger)

	return manager, db, backupDir
}

func TestRecoveryManagerCreate(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestCreateBackup(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create test nodes
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

	ctx := context.Background()
	backup, err := manager.CreateBackup(ctx, "test-cluster", "test backup")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	if backup.ClusterID != "test-cluster" {
		t.Errorf("Expected ClusterID 'test-cluster', got '%s'", backup.ClusterID)
	}

	if len(backup.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(backup.Nodes))
	}
}

func TestListBackups(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create initial backup
	ctx := context.Background()
	_, err := manager.CreateBackup(ctx, "test-cluster", "backup 1")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Create another backup
	_, err = manager.CreateBackup(ctx, "test-cluster", "backup 2")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// List backups
	backups, err := manager.ListBackups("test-cluster")
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}
}

func TestGetBackup(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create backup
	ctx := context.Background()
	created, err := manager.CreateBackup(ctx, "test-cluster", "test")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Get backup
	loaded, err := manager.GetBackup(created.ID)
	if err != nil {
		t.Fatalf("GetBackup failed: %v", err)
	}

	if loaded.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, loaded.ID)
	}
}

func TestRestoreCluster(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create test nodes
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

	// Create backup
	ctx := context.Background()
	backup, err := manager.CreateBackup(ctx, "test-cluster", "test")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Delete the cluster
	db.DeleteCluster("test-cluster")

	// Restore
	err = manager.RestoreCluster(ctx, backup.ID)
	if err != nil {
		t.Fatalf("RestoreCluster failed: %v", err)
	}

	// Verify cluster exists
	restored, err := db.GetCluster("test-cluster")
	if err != nil {
		t.Fatalf("GetCluster failed: %v", err)
	}
	if restored == nil {
		t.Fatal("Cluster should exist after restore")
	}
}

func TestDeleteBackup(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create backup
	ctx := context.Background()
	backup, err := manager.CreateBackup(ctx, "test-cluster", "test")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Delete backup
	err = manager.DeleteBackup(backup.ID)
	if err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}

	// Verify deleted
	_, err = manager.GetBackup(backup.ID)
	if err == nil {
		t.Error("Expected error when getting deleted backup")
	}
}

func TestFailover(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	// Create worker and standby nodes
	db.SaveNode(&database.Node{
		ID:        "worker-1",
		ClusterID: "test-cluster",
		Name:      "worker-1",
		Role:      "worker",
		HostID:    "test-host",
		State:     "running",
	})

	db.SaveNode(&database.Node{
		ID:        "standby-1",
		ClusterID: "test-cluster",
		Name:      "standby-1",
		Role:      "standby",
		HostID:    "test-host",
		State:     "running",
	})

	ctx := context.Background()
	err := manager.Failover(ctx, "test-cluster", "worker-1")
	if err != nil {
		t.Fatalf("Failover failed: %v", err)
	}

	// Verify standby was promoted
	standby, _ := db.GetNode("standby-1")
	if standby.Role != "worker" {
		t.Errorf("Expected standby to be promoted to worker, got '%s'", standby.Role)
	}
}

func TestBackupNodeHelpers(t *testing.T) {
	node := &orchestrator.BackupNode{
		Config: map[string]interface{}{
			"cpu":       4,
			"memory_mb": 8192.0,
			"disk_gb":   100.0,
		},
	}

	if node.CPU() != 4 {
		t.Errorf("Expected CPU 4, got %d", node.CPU())
	}
	if node.MemoryMB() != 8192 {
		t.Errorf("Expected MemoryMB 8192, got %d", node.MemoryMB())
	}
	if node.DiskGB() != 100 {
		t.Errorf("Expected DiskGB 100, got %d", node.DiskGB())
	}
}

func TestBackupNodeHelpersFloat(t *testing.T) {
	node := &orchestrator.BackupNode{
		Config: map[string]interface{}{
			"cpu":       4.0,
			"memory_mb": 8192.0,
			"disk_gb":   100.0,
		},
	}

	if node.CPU() != 4 {
		t.Errorf("Expected CPU 4, got %d", node.CPU())
	}
}

func TestBackupNodeHelpersNil(t *testing.T) {
	node := &orchestrator.BackupNode{
		Config: nil,
	}

	if node.CPU() != 0 {
		t.Errorf("Expected CPU 0, got %d", node.CPU())
	}
	if node.MemoryMB() != 0 {
		t.Errorf("Expected MemoryMB 0, got %d", node.MemoryMB())
	}
	if node.DiskGB() != 0 {
		t.Errorf("Expected DiskGB 0, got %d", node.DiskGB())
	}
}

func TestRestoreToNonexistentCluster(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster and backup
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	ctx := context.Background()
	backup, err := manager.CreateBackup(ctx, "test-cluster", "test")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Delete cluster completely
	db.DeleteCluster("test-cluster")

	// Restore should recreate the cluster
	err = manager.RestoreCluster(ctx, backup.ID)
	if err != nil {
		t.Fatalf("RestoreCluster failed: %v", err)
	}
}

func TestBackupFileLocation(t *testing.T) {
	manager, db, backupDir := createTestRecoveryManager(t)
	defer db.Close()
	defer os.RemoveAll(backupDir)

	// Create test cluster
	testCluster := &database.Cluster{
		ID:     "test-cluster",
		Name:   "Test Cluster",
		Status: "running",
	}
	db.SaveCluster(testCluster)

	ctx := context.Background()
	backup, err := manager.CreateBackup(ctx, "test-cluster", "test")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify file exists
	filename := backup.ID + ".json"
	expectedPath := filepath.Join(backupDir, filename)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Backup file should exist at %s", expectedPath)
	}
}
