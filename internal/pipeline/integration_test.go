//go:build integration
// +build integration

// Package pipeline provides integration tests with real database
package pipeline

import (
	"context"
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

// mockNetworkManager implements types.NetworkManagerInterface for testing
type mockNetworkManager struct {
	networks map[string]*types.NetworkConfig
}

func newMockNetworkManager() *mockNetworkManager {
	return &mockNetworkManager{
		networks: make(map[string]*types.NetworkConfig),
	}
}

func (m *mockNetworkManager) CreateNetwork(config *types.NetworkConfig) (string, error) {
	id := "net-123"
	m.networks[id] = config
	return id, nil
}

func (m *mockNetworkManager) DestroyNetwork(networkID string) error {
	delete(m.networks, networkID)
	return nil
}

func (m *mockNetworkManager) GetNetwork(networkID string) (*types.NetworkConfig, error) {
	if net, ok := m.networks[networkID]; ok {
		return net, nil
	}
	return nil, os.ErrNotExist
}

// mockRunnerManager implements types.RunnerManagerInterface for testing
type mockRunnerManager struct {
	runners map[string]map[string]interface{}
}

func newMockRunnerManager() *mockRunnerManager {
	return &mockRunnerManager{
		runners: make(map[string]map[string]interface{}),
	}
}

func (m *mockRunnerManager) CreateRunner(platform types.RunnerPlatform, config map[string]interface{}) (string, error) {
	id := string(platform) + "-runner-123"
	m.runners[id] = config
	return id, nil
}

func (m *mockRunnerManager) DestroyRunner(runnerID string) error {
	delete(m.runners, runnerID)
	return nil
}

func (m *mockRunnerManager) GetRunner(runnerID string) (map[string]interface{}, error) {
	if runner, ok := m.runners[runnerID]; ok {
		return runner, nil
	}
	return nil, os.ErrNotExist
}

// TestCoordinator_CreatePipeline_RealDB tests pipeline creation with real database
func TestCoordinator_CreatePipeline_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pipeline-test-*")
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

	netMgr := newMockNetworkManager()
	runnerMgr := newMockRunnerManager()

	coord, err := NewCoordinator(db, poolMgr, netMgr, runnerMgr)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Create a pipeline
	ps, err := coord.CreatePipeline(context.Background(), types.PlatformGitLab, "https://gitlab.example.com/test/repo", "main")
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	if ps.ID == "" {
		t.Error("expected pipeline ID to be set")
	}
	if ps.Platform != types.PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", ps.Platform)
	}
	if ps.Repository != "https://gitlab.example.com/test/repo" {
		t.Errorf("expected repository URL, got %s", ps.Repository)
	}
}

// TestCoordinator_GetPipeline_RealDB tests pipeline retrieval with real database
func TestCoordinator_GetPipeline_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pipeline-test-*")
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

	netMgr := newMockNetworkManager()
	runnerMgr := newMockRunnerManager()

	coord, err := NewCoordinator(db, poolMgr, netMgr, runnerMgr)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Create a pipeline first
	created, err := coord.CreatePipeline(context.Background(), types.PlatformGitLab, "https://gitlab.example.com/test/repo", "main")
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// Get the pipeline
	retrieved, err := coord.GetPipeline(created.ID)
	if err != nil {
		t.Fatalf("failed to get pipeline: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected pipeline ID %s, got %s", created.ID, retrieved.ID)
	}
	if retrieved.Repository != created.Repository {
		t.Errorf("expected repository %s, got %s", created.Repository, retrieved.Repository)
	}
}

// TestCoordinator_ListPipelines_RealDB tests listing pipelines with real database
func TestCoordinator_ListPipelines_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pipeline-test-*")
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

	netMgr := newMockNetworkManager()
	runnerMgr := newMockRunnerManager()

	coord, err := NewCoordinator(db, poolMgr, netMgr, runnerMgr)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Create multiple pipelines
	for i := 0; i < 3; i++ {
		_, err := coord.CreatePipeline(context.Background(), types.PlatformGitLab, "https://gitlab.example.com/test/repo", "main")
		if err != nil {
			t.Fatalf("failed to create pipeline %d: %v", i, err)
		}
	}

	// List all pipelines
	pipelines := coord.ListPipelines()

	if len(pipelines) < 3 {
		t.Errorf("expected at least 3 pipelines, got %d", len(pipelines))
	}
}

// TestCoordinator_DeletePipeline_RealDB tests pipeline deletion with real database
func TestCoordinator_DeletePipeline_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pipeline-test-*")
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

	netMgr := newMockNetworkManager()
	runnerMgr := newMockRunnerManager()

	coord, err := NewCoordinator(db, poolMgr, netMgr, runnerMgr)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Create a pipeline
	created, err := coord.CreatePipeline(context.Background(), types.PlatformGitLab, "https://gitlab.example.com/test/repo", "main")
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// Delete the pipeline
	err = coord.DeletePipeline(created.ID)
	if err != nil {
		t.Fatalf("failed to delete pipeline: %v", err)
	}

	// Verify it's deleted
	_, err = coord.GetPipeline(created.ID)
	if err == nil {
		t.Error("expected error when getting deleted pipeline")
	}
}