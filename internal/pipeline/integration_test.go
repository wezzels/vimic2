//go:build integration
// +build integration

// Package pipeline provides integration tests with real database
package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		ID:     "vm-test-" + time.Now().Format("20060102150405"),
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
	id := "net-" + time.Now().Format("20060102150405")
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
	id := string(platform) + "-runner-" + time.Now().Format("20060102150405")
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

// TestPipelineStatus_Constants tests that pipeline status constants are valid
func TestPipelineStatus_Constants(t *testing.T) {
	statuses := []types.PipelineStatus{
		types.PipelineStatusCreating,
		types.PipelineStatusRunning,
		types.PipelineStatusSuccess,
		types.PipelineStatusFailed,
		types.PipelineStatusCanceled,
	}

	expected := []string{"creating", "running", "success", "failed", "canceled"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("expected status %s, got %s", expected[i], status)
		}
	}
}

// TestRunnerPlatform_Constants tests that runner platform constants are valid
func TestRunnerPlatform_Constants(t *testing.T) {
	platforms := []types.RunnerPlatform{
		types.PlatformGitLab,
		types.PlatformGitHub,
		types.PlatformJenkins,
		types.PlatformCircleCI,
		types.PlatformDrone,
	}

	expected := []string{"gitlab", "github", "jenkins", "circleci", "drone"}

	for i, platform := range platforms {
		if string(platform) != expected[i] {
			t.Errorf("expected platform %s, got %s", expected[i], platform)
		}
	}
}

// TestPipelineState_Fields tests PipelineState struct fields
func TestPipelineState_Fields(t *testing.T) {
	now := time.Now()
	ps := &PipelineState{
		ID:         "pipeline-1",
		Platform:   types.PlatformGitLab,
		Repository: "https://gitlab.example.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		CommitMsg:  "Test commit",
		Author:     "test@example.com",
		Status:     types.PipelineStatusRunning,
		NetworkID:  "network-1",
		VMs:        []string{"vm-1", "vm-2"},
		Runners:    []string{"runner-1"},
		StartTime:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if ps.ID != "pipeline-1" {
		t.Errorf("expected ID pipeline-1, got %s", ps.ID)
	}
	if ps.Platform != types.PlatformGitLab {
		t.Errorf("expected platform gitlab, got %s", ps.Platform)
	}
	if len(ps.VMs) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(ps.VMs))
	}
	if len(ps.Runners) != 1 {
		t.Errorf("expected 1 runner, got %d", len(ps.Runners))
	}
}

// TestStageState_Fields tests StageState struct fields
func TestStageState_Fields(t *testing.T) {
	now := time.Now()
	stage := &StageState{
		Name:      "build",
		Status:    types.PipelineStatusRunning,
		StartTime: &now,
		Jobs: []JobState{
			{
				ID:     "job-1",
				Name:   "build-job",
				Stage:  "build",
				Status: types.PipelineStatusRunning,
			},
		},
	}

	if stage.Name != "build" {
		t.Errorf("expected stage name build, got %s", stage.Name)
	}
	if stage.Status != types.PipelineStatusRunning {
		t.Errorf("expected status running, got %s", stage.Status)
	}
	if len(stage.Jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(stage.Jobs))
	}
}

// TestJobState_Fields tests JobState struct fields
func TestJobState_Fields(t *testing.T) {
	now := time.Now()
	job := &JobState{
		ID:        "job-1",
		Name:      "test-job",
		Stage:     "test",
		Status:    types.PipelineStatusSuccess,
		RunnerID:  "runner-1",
		StartTime: &now,
		EndTime:   &now,
		Duration:  60,
		Log:       []string{"line 1", "line 2"},
	}

	if job.ID != "job-1" {
		t.Errorf("expected job ID job-1, got %s", job.ID)
	}
	if job.Status != types.PipelineStatusSuccess {
		t.Errorf("expected status success, got %s", job.Status)
	}
	if len(job.Log) != 2 {
		t.Errorf("expected 2 log lines, got %d", len(job.Log))
	}
}

// TestPipelineEvent_Fields tests PipelineEvent struct fields
func TestPipelineEvent_Fields(t *testing.T) {
	now := time.Now()
	event := &PipelineEvent{
		PipelineID: "pipeline-1",
		OldStatus:  types.PipelineStatusCreating,
		NewStatus:  types.PipelineStatusRunning,
		Message:    "Pipeline started",
		Timestamp:  now,
	}

	if event.PipelineID != "pipeline-1" {
		t.Errorf("expected pipeline ID pipeline-1, got %s", event.PipelineID)
	}
	if event.OldStatus != types.PipelineStatusCreating {
		t.Errorf("expected old status creating, got %s", event.OldStatus)
	}
	if event.NewStatus != types.PipelineStatusRunning {
		t.Errorf("expected new status running, got %s", event.NewStatus)
	}
}