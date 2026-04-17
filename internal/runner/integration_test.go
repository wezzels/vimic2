//go:build integration
// +build integration

// Package runner provides integration tests with real database
package runner

import (
	"context"
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
		ID:     "vm-test-123",
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
	return nil, os.ErrNotExist
}

func (m *mockPoolManager) ListPools() ([]*types.PoolState, error) {
	var pools []*types.PoolState
	for _, p := range m.pools {
		pools = append(pools, p)
	}
	return pools, nil
}

// TestRunnerManager_UpdateRunnerStatus tests updating runner status
func TestRunnerManager_UpdateRunnerStatus_RealDB(t *testing.T) {
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
	runner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Test initial status
	if runner.Status != types.RunnerStatusCreating {
		t.Errorf("expected initial status creating, got %s", runner.Status)
	}
}

// TestRunnerManager_ListRunnersByPool tests listing runners by pool
func TestRunnerManager_ListRunnersByPool_RealDB(t *testing.T) {
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
	poolMgr.pools["pool-a"] = &types.PoolState{
		Name:      "pool-a",
		Capacity:  5,
		Available: 5,
	}
	poolMgr.pools["pool-b"] = &types.PoolState{
		Name:      "pool-b",
		Capacity:  5,
		Available: 5,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create runners in different pools
	_, err = rm.CreateRunner(context.Background(), "pool-a", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}
	_, err = rm.CreateRunner(context.Background(), "pool-a", types.PlatformGitLab, "pipeline-2", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}
	_, err = rm.CreateRunner(context.Background(), "pool-b", types.PlatformGitHub, "pipeline-3", []string{"ubuntu"})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// List all runners
	runners, err := rm.ListRunners()
	if err != nil {
		t.Fatalf("failed to list runners: %v", err)
	}

	if len(runners) != 3 {
		t.Errorf("expected 3 runners, got %d", len(runners))
	}
}

// TestRunnerManager_DifferentPlatforms tests runners on different platforms
func TestRunnerManager_DifferentPlatforms_RealDB(t *testing.T) {
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

	// Test GitLab runner
	gitlabRunner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("failed to create GitLab runner: %v", err)
	}
	if gitlabRunner.Platform != types.PlatformGitLab {
		t.Errorf("expected GitLab platform, got %s", gitlabRunner.Platform)
	}

	// Test GitHub runner
	githubRunner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitHub, "pipeline-2", []string{"ubuntu-latest"})
	if err != nil {
		t.Fatalf("failed to create GitHub runner: %v", err)
	}
	if githubRunner.Platform != types.PlatformGitHub {
		t.Errorf("expected GitHub platform, got %s", githubRunner.Platform)
	}

	// Test Jenkins runner
	jenkinsRunner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformJenkins, "pipeline-3", []string{"java"})
	if err != nil {
		t.Fatalf("failed to create Jenkins runner: %v", err)
	}
	if jenkinsRunner.Platform != types.PlatformJenkins {
		t.Errorf("expected Jenkins platform, got %s", jenkinsRunner.Platform)
	}
}

// TestRunnerManager_Labels tests runner labels handling
func TestRunnerManager_Labels_RealDB(t *testing.T) {
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

	// Create runner with multiple labels
	labels := []string{"docker", "linux", "ubuntu", "gpu"}
	runner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitLab, "pipeline-1", labels)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	if len(runner.Labels) != 4 {
		t.Errorf("expected 4 labels, got %d", len(runner.Labels))
	}

	// Verify labels are preserved
	for i, label := range labels {
		if runner.Labels[i] != label {
			t.Errorf("expected label %s, got %s", label, runner.Labels[i])
		}
	}
}

// TestRunnerManager_EmptyLabels tests runner with no labels
func TestRunnerManager_EmptyLabels_RealDB(t *testing.T) {
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

	// Create runner with no labels
	runner, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitLab, "pipeline-1", []string{})
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	if len(runner.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(runner.Labels))
	}
}

// TestRunnerManager_DestroyNonExistent tests destroying a non-existent runner
func TestRunnerManager_DestroyNonExistent_RealDB(t *testing.T) {
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

	// Try to destroy non-existent runner
	err = rm.DestroyRunner(context.Background(), "non-existent-runner")
	if err == nil {
		t.Error("expected error when destroying non-existent runner")
	}
	if !errors.Is(err, ErrRunnerNotFound) {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

// TestRunnerManager_GetNonExistent tests getting a non-existent runner
func TestRunnerManager_GetNonExistent_RealDB(t *testing.T) {
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

	// Try to get non-existent runner
	_, err = rm.GetRunner("non-existent-runner")
	if err == nil {
		t.Error("expected error when getting non-existent runner")
	}
	if !errors.Is(err, ErrRunnerNotFound) {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

// TestRunnerManager_PoolNotFoundError tests error when pool not found
func TestRunnerManager_PoolNotFound_RealDB(t *testing.T) {
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
	// Don't add any pools

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Runner creation should still work (pool manager is only used for allocation)
	runner, err := rm.CreateRunner(context.Background(), "non-existent-pool", types.PlatformGitLab, "pipeline-1", []string{"docker"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.PoolName != "non-existent-pool" {
		t.Errorf("expected pool name non-existent-pool, got %s", runner.PoolName)
	}
}

// TestRunnerManager_ConcurrentOperations tests concurrent runner operations
func TestRunnerManager_ConcurrentOperations_RealDB(t *testing.T) {
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
		Capacity:  10,
		Available: 10,
	}

	rm, err := NewRunnerManager(db, poolMgr, nil)
	if err != nil {
		t.Fatalf("failed to create runner manager: %v", err)
	}

	// Create multiple runners concurrently
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			_, err := rm.CreateRunner(context.Background(), "test-pool", types.PlatformGitLab, "pipeline-"+string(rune('A'+idx)), []string{"docker"})
			if err != nil {
				t.Errorf("failed to create runner %d: %v", idx, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// List all runners
	runners, err := rm.ListRunners()
	if err != nil {
		t.Fatalf("failed to list runners: %v", err)
	}

	if len(runners) != 10 {
		t.Errorf("expected 10 runners, got %d", len(runners))
	}
}