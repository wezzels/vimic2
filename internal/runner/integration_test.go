//go:build integration
// +build integration

// Package runner provides integration tests with real database
package runner

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/types"
)

// mockPoolManager implements types.PoolManagerInterface for testing
type mockPoolManager struct {
	pools map[string]*types.PoolState
	vms   map[string]*types.VMState
}

func newMockPoolManager() *mockPoolManager {
	return &mockPoolManager{
		pools: make(map[string]*types.PoolState),
		vms:   make(map[string]*types.VMState),
	}
}

func (m *mockPoolManager) AllocateVM(poolName string) (*types.VMState, error) {
	vm := &types.VMState{
		ID:     "vm-test-id-123",
		Status: "running",
	}
	m.vms[vm.ID] = vm
	return vm, nil
}

func (m *mockPoolManager) ReleaseVM(vmID string) error {
	delete(m.vms, vmID)
	return nil
}

func (m *mockPoolManager) GetPool(name string) (*types.PoolState, error) {
	if pool, ok := m.pools[name]; ok {
		return pool, nil
	}
	return nil, errors.New("pool not found")
}

func (m *mockPoolManager) ListPools() ([]*types.PoolState, error) {
	var pools []*types.PoolState
	for _, p := range m.pools {
		pools = append(pools, p)
	}
	return pools, nil
}

// TestRunnerManager_CreateRunner_RealDB tests runner creation with real database
func TestRunnerManager_CreateRunner_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-runner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	poolMgr := newMockPoolManager()
	poolMgr.pools["test-pool"] = &types.PoolState{
		Name:      "test-pool",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create a runner
	runner, err := rm.CreateRunner(nil, "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker", "linux"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	if runner.ID == "" {
		t.Error("expected runner ID to be set")
	}
	if runner.Platform != types.PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", runner.Platform)
	}
	if runner.PipelineID != "pipeline-1" {
		t.Errorf("expected pipeline-1, got %s", runner.PipelineID)
	}
	if len(runner.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(runner.Labels))
	}
}

// TestRunnerManager_GetRunner_RealDB tests runner retrieval with real database
func TestRunnerManager_GetRunner_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-runner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	poolMgr := newMockPoolManager()
	poolMgr.pools["test-pool"] = &types.PoolState{
		Name:      "test-pool",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create a runner first
	createdRunner, err := rm.CreateRunner(nil, "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Get the runner
	retrievedRunner, err := rm.GetRunner(createdRunner.ID)
	if err != nil {
		t.Fatalf("failed to get runner: %v", err)
	}

	if retrievedRunner.ID != createdRunner.ID {
		t.Errorf("expected runner ID %s, got %s", createdRunner.ID, retrievedRunner.ID)
	}
	if retrievedRunner.Platform != createdRunner.Platform {
		t.Errorf("expected platform %s, got %s", createdRunner.Platform, retrievedRunner.Platform)
	}
}

// TestRunnerManager_ListRunners_RealDB tests listing runners with real database
func TestRunnerManager_ListRunners_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-runner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	poolMgr := newMockPoolManager()
	poolMgr.pools["test-pool"] = &types.PoolState{
		Name:      "test-pool",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create multiple runners
	for i := 0; i < 3; i++ {
		_, err := rm.CreateRunner(nil, "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
		if err != nil {
			t.Fatalf("failed to create runner %d: %v", i, err)
		}
	}

	// List all runners
	runners, err := rm.ListRunners()
	if err != nil {
		t.Fatalf("failed to list runners: %v", err)
	}

	if len(runners) < 3 {
		t.Errorf("expected at least 3 runners, got %d", len(runners))
	}
}

// TestRunnerManager_DestroyRunner_RealDB tests runner deletion with real database
func TestRunnerManager_DestroyRunner_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-runner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	poolMgr := newMockPoolManager()
	poolMgr.pools["test-pool"] = &types.PoolState{
		Name:      "test-pool",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create a runner
	createdRunner, err := rm.CreateRunner(nil, "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Destroy the runner
	err = rm.DestroyRunner(nil, createdRunner.ID)
	if err != nil {
		t.Fatalf("failed to destroy runner: %v", err)
	}

	// Verify it's deleted
	_, err = rm.GetRunner(createdRunner.ID)
	if err == nil {
		t.Error("expected error when getting destroyed runner")
	}
}

// TestRunnerManager_SetRunnerStatus_RealDB tests runner status updates with real database
func TestRunnerManager_SetRunnerStatus_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-runner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	poolMgr := newMockPoolManager()
	poolMgr.pools["test-pool"] = &types.PoolState{
		Name:      "test-pool",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create a runner
	runner, err := rm.CreateRunner(nil, "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Update status
	runner.Status = types.RunnerStatusOnline

	// Verify status
	retrievedRunner, err := rm.GetRunner(runner.ID)
	if err != nil {
		t.Fatalf("failed to get runner: %v", err)
	}

	if retrievedRunner.Status != types.RunnerStatusOnline {
		t.Errorf("expected online status, got %s", retrievedRunner.Status)
	}
}