//go:build integration
// +build integration

// Package pipeline provides helper function tests
package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/types"
)

// TestPipelineState_GenerateID tests generateID function
func TestPipelineState_GenerateID(t *testing.T) {
	id := generateID()
	if len(id) == 0 {
		t.Error("expected non-empty ID")
	}

	// IDs should be unique
	id2 := generateID()
	if id == id2 {
		t.Error("expected unique IDs")
	}
}

// TestPipelineState_RandomString tests randomString function
func TestPipelineState_RandomString(t *testing.T) {
	str := randomString(10)
	if len(str) != 10 {
		t.Errorf("expected string of length 10, got %d", len(str))
	}

	// Strings should be unique
	str2 := randomString(10)
	if str == str2 {
		t.Error("expected unique strings")
	}
}

// TestDispatcher_GenerateJobID tests generateJobID function
func TestDispatcher_GenerateJobID(t *testing.T) {
	jobID := generateJobID()
	if len(jobID) == 0 {
		t.Error("expected non-empty job ID")
	}

	// Job IDs should be unique
	jobID2 := generateJobID()
	if jobID == jobID2 {
		t.Error("expected unique job IDs")
	}
}

// TestCoordinator_CreatePipelineWithRealDB tests creating pipeline with real database
func TestCoordinator_CreatePipelineWithRealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
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

	// Test pipeline creation with different platforms
	platforms := []struct {
		platform types.RunnerPlatform
		repo     string
	}{
		{types.PlatformGitLab, "https://gitlab.example.com/test/repo"},
		{types.PlatformGitHub, "https://github.com/test/repo"},
		{types.PlatformJenkins, "https://jenkins.example.com/job/test"},
	}

	for _, tc := range platforms {
		ps, err := coord.CreatePipeline(context.Background(), tc.platform, tc.repo, "main")
		if err != nil {
			t.Fatalf("CreatePipeline for %s failed: %v", tc.platform, err)
		}

		if ps.Platform != tc.platform {
			t.Errorf("expected platform %s, got %s", tc.platform, ps.Platform)
		}

		if ps.Repository != tc.repo {
			t.Errorf("expected repo %s, got %s", tc.repo, ps.Repository)
		}

		if ps.Branch != "main" {
			t.Errorf("expected branch main, got %s", ps.Branch)
		}
	}
}

// TestCoordinator_PipelineOperations tests basic pipeline operations
func TestCoordinator_PipelineOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
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

	// Create pipeline
	ps, err := coord.CreatePipeline(context.Background(), types.PlatformGitLab, "https://gitlab.example.com/test/repo", "main")
	if err != nil {
		t.Fatalf("CreatePipeline failed: %v", err)
	}

	// Get pipeline
	retrieved, err := coord.GetPipeline(ps.ID)
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}

	if retrieved.ID != ps.ID {
		t.Errorf("expected ID %s, got %s", ps.ID, retrieved.ID)
	}

	// List pipelines
	pipelines := coord.ListPipelines()
	if len(pipelines) == 0 {
		t.Error("expected at least one pipeline")
	}

	// Delete pipeline
	err = coord.DeletePipeline(ps.ID)
	if err != nil {
		t.Fatalf("DeletePipeline failed: %v", err)
	}

	// Verify deleted
	_, err = coord.GetPipeline(ps.ID)
	if err == nil {
		t.Error("expected error when getting deleted pipeline")
	}
}

// TestPipelineDB_GetStats tests PipelineDB statistics
func TestPipelineDB_GetStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	pipelineDB, err := NewPipelineDB(dbPath)
	if err != nil {
		t.Fatalf("NewPipelineDB failed: %v", err)
	}
	defer pipelineDB.Close()

	stats, err := pipelineDB.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats == nil {
		t.Error("expected non-nil stats")
	}
}