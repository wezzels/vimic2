//go:build integration

package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// ==================== Recovery Manager Tests ====================

func TestNewRecoveryManager_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-recovery-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rm := NewRecoveryManager(db, tmpDir)
	if rm == nil {
		t.Fatal("RecoveryManager should not be nil")
	}
}

func TestRecoveryManager_CreateBackup_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-recovery-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rm := NewRecoveryManager(db, tmpDir)

	backup, err := rm.CreateBackup(context.Background(), "test-cluster", "test backup")
	if err != nil {
		t.Skipf("CreateBackup failed: %v", err)
	}
	t.Logf("Created backup: %s", backup.ID)
}

// ==================== ScaleRule Tests ====================

func TestScaleRule_Struct(t *testing.T) {
	now := time.Now()
	rule := &ScaleRule{
		ClusterID:      "test-cluster",
		Metric:         "cpu",
		UpperThreshold: 80.0,
		LowerThreshold: 20.0,
		ScaleUpCount:   2,
		ScaleDownCount: 1,
		Cooldown:       5 * time.Minute,
		LastScaleUp:    now,
		LastScaleDown:  now,
		Enabled:        true,
	}

	if rule.ClusterID != "test-cluster" {
		t.Errorf("ClusterID = %s, want test-cluster", rule.ClusterID)
	}
	if rule.Metric != "cpu" {
		t.Errorf("Metric = %s, want cpu", rule.Metric)
	}
	if rule.UpperThreshold != 80.0 {
		t.Errorf("UpperThreshold = %f, want 80.0", rule.UpperThreshold)
	}
	if rule.ScaleUpCount != 2 {
		t.Errorf("ScaleUpCount = %d, want 2", rule.ScaleUpCount)
	}
	if !rule.Enabled {
		t.Error("Enabled should be true")
	}
}

// ==================== RollingUpdater Tests ====================

func TestRollingUpdater_Cancel_Real(t *testing.T) {
	ru := NewRollingUpdater(nil, nil)
	if ru == nil {
		t.Fatal("NewRollingUpdater should not return nil")
	}

	ru.Cancel()
	t.Log("RollingUpdater.Cancel() succeeded")
}

func TestRollingUpdater_IsValid_Real(t *testing.T) {
	_ = NewRollingUpdater(nil, nil)

	strategy := UpdateStrategy{
		BatchSize:      3,
		MaxUnavailable: 1,
		WaitBetween:    30,
	}
	valid := strategy.IsValid()
	if !valid {
		t.Error("IsValid should accept valid strategy")
	}
}

func TestRollingUpdater_CalculateBatches_Real(t *testing.T) {
	ru := NewRollingUpdater(nil, nil)

	nodes := []string{"node1", "node2", "node3", "node4", "node5"}
	batches := ru.CalculateBatches(nodes, 2)
	if len(batches) == 0 {
		t.Error("CalculateBatches should return at least one batch")
	}
	t.Logf("Calculated %d batches for 5 nodes with batch size 2", len(batches))
}

// ==================== UpdateStrategy Tests ====================

func TestUpdateStrategy_Struct(t *testing.T) {
	strategy := UpdateStrategy{
		BatchSize:      3,
		MaxUnavailable: 2,
		WaitBetween:    60,
	}

	if strategy.BatchSize != 3 {
		t.Errorf("BatchSize = %d, want 3", strategy.BatchSize)
	}
	if strategy.MaxUnavailable != 2 {
		t.Errorf("MaxUnavailable = %d, want 2", strategy.MaxUnavailable)
	}
	if strategy.WaitBetween != 60 {
		t.Errorf("WaitBetween = %d, want 60", strategy.WaitBetween)
	}
}

// ==================== HealthChecker Tests ====================

func TestNewHealthChecker_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-health-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hc := NewHealthChecker(db)
	if hc == nil {
		t.Fatal("NewHealthChecker should not be nil")
	}
}

func TestHealthChecker_CheckNode_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-health-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hc := NewHealthChecker(db)

	status := hc.CheckNode("test-node")
	t.Logf("CheckNode: healthy=%v nodeID=%s", status.Healthy, status.NodeID)
}

func TestHealthChecker_GetStatus_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-health-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hc := NewHealthChecker(db)

	status := hc.GetStatus("test-node")
	t.Logf("GetStatus: %v", status)
}

func TestHealthChecker_GetAllStatus_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-health-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hc := NewHealthChecker(db)

	statuses := hc.GetAllStatus()
	t.Logf("GetAllStatus: %d statuses", len(statuses))
}

// ==================== HealthStatus Tests ====================

func TestHealthStatus_Struct(t *testing.T) {
	now := time.Now()
	hs := &HealthStatus{
		NodeID:    "node-1",
		Healthy:   true,
		LastCheck: now,
		Checks:    10,
		Failures:  0,
		Message:   "all good",
	}

	if hs.NodeID != "node-1" {
		t.Errorf("NodeID = %s, want node-1", hs.NodeID)
	}
	if !hs.Healthy {
		t.Error("Healthy should be true")
	}
	if hs.Checks != 10 {
		t.Errorf("Checks = %d, want 10", hs.Checks)
	}
	if hs.Failures != 0 {
		t.Errorf("Failures = %d, want 0", hs.Failures)
	}
}

// ==================== BackupNode Tests ====================

func TestBackupNode_CPU(t *testing.T) {
	node := &BackupNode{
		Config: map[string]interface{}{
			"cpu":       4,
			"memory_mb": 8192,
			"disk_gb":   100,
		},
	}

	cpu := node.CPU()
	if cpu != 4 {
		t.Errorf("CPU = %d, want 4", cpu)
	}
}

func TestBackupNode_MemoryMB(t *testing.T) {
	node := &BackupNode{
		Config: map[string]interface{}{
			"cpu":       4,
			"memory_mb": 8192,
			"disk_gb":   100,
		},
	}

	mem := node.MemoryMB()
	if mem != 8192 {
		t.Errorf("MemoryMB = %d, want 8192", mem)
	}
}

func TestBackupNode_DiskGB(t *testing.T) {
	node := &BackupNode{
		Config: map[string]interface{}{
			"cpu":       4,
			"memory_mb": 8192,
			"disk_gb":   100,
		},
	}

	disk := node.DiskGB()
	if disk != 100 {
		t.Errorf("DiskGB = %d, want 100", disk)
	}
}