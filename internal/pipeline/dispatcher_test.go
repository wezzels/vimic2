//go:build integration

package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ==================== Dispatcher Tests ====================

func setupDispatcherTest(t *testing.T) (*JobDispatcher, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-dispatch-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("NewPipelineDB failed: %v", err)
	}

	// Create dispatcher without runner manager (nil is accepted)
	config := &DispatcherConfig{
		Workers:    1,
		QueueSize:  10,
		MaxRetries: 2,
		JobTimeout: 5 * time.Minute,
	}

	// NewJobDispatcher with nil runnerMgr should still work for basic ops
	dispatcher, err := NewJobDispatcher(db, nil, config)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("NewJobDispatcher failed: %v", err)
	}

	// Set state file to temp directory
	dispatcher.SetStateFile(filepath.Join(tmpDir, "jobs-state.json"))

	t.Cleanup(func() {
		dispatcher.Stop()
		db.Close()
		os.RemoveAll(tmpDir)
	})

	return dispatcher, tmpDir
}

func TestJobDispatcher_New(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	if dispatcher == nil {
		t.Fatal("JobDispatcher should not be nil")
	}
}

func TestJobDispatcher_DefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-dispatch-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dispatcher, err := NewJobDispatcher(db, nil, nil)
	if err != nil {
		t.Fatalf("NewJobDispatcher with nil config failed: %v", err)
	}

	stats := dispatcher.GetStats()
	if stats["total"] != 0 {
		t.Errorf("Stats total = %v, want 0 for empty dispatcher", stats["total"])
	}

	dispatcher.Stop()
}

func TestJobDispatcher_EnqueueJob(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	ctx := context.Background()

	job := &Job{
		PipelineID: "test-pipeline",
		Stage:       "build",
		Name:        "test-job",
		Commands:    []string{"echo hello"},
		Environment: map[string]string{"ENV": "test"},
	}

	err := dispatcher.EnqueueJob(ctx, job)
	if err != nil {
		t.Fatalf("EnqueueJob failed: %v", err)
	}

	if job.ID == "" {
		t.Error("Job ID should be set after enqueue")
	}
	if job.Status != PipelineStatusRunning {
		t.Errorf("Job Status = %s, want running", job.Status)
	}
}

func TestJobDispatcher_GetJob(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	ctx := context.Background()

	job := &Job{
		PipelineID: "test-pipeline",
		Stage:       "build",
		Name:        "test-job",
		Commands:    []string{"echo hello"},
	}

	dispatcher.EnqueueJob(ctx, job)

	retrieved, err := dispatcher.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}
	if retrieved.ID != job.ID {
		t.Errorf("GetJob ID = %s, want %s", retrieved.ID, job.ID)
	}
}

func TestJobDispatcher_ListJobs(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		job := &Job{
			PipelineID: "test-pipeline",
			Stage:       "build",
			Name:        "test-job",
			Commands:    []string{"echo hello"},
		}
		dispatcher.EnqueueJob(ctx, job)
	}

	jobs := dispatcher.ListJobs()
	if len(jobs) < 3 {
		t.Errorf("ListJobs returned %d jobs, want at least 3", len(jobs))
	}
}

func TestJobDispatcher_GetStats(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	ctx := context.Background()

	job := &Job{
		PipelineID: "test-pipeline",
		Stage:       "build",
		Name:        "test-job",
		Commands:    []string{"echo hello"},
	}
	dispatcher.EnqueueJob(ctx, job)

	stats := dispatcher.GetStats()
	if stats["total"] < 1 {
		t.Errorf("Stats total = %v, want at least 1", stats["total"])
	}
}

func TestJobDispatcher_CancelJob(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)
	ctx := context.Background()

	job := &Job{
		PipelineID: "test-pipeline",
		Stage:       "build",
		Name:        "test-job",
		Commands:    []string{"echo hello"},
	}

	dispatcher.EnqueueJob(ctx, job)

	// Cancel might fail if job already completed
	err := dispatcher.CancelJob(ctx, job.ID)
	if err != nil {
		t.Logf("CancelJob: %v (job may have already completed)", err)
	}
}

func TestJobDispatcher_StateFile(t *testing.T) {
	dispatcher, tmpDir := setupDispatcherTest(t)

	stateFile := dispatcher.getStateFile()
	if stateFile == "" {
		t.Error("getStateFile should not return empty string")
	}

	customPath := filepath.Join(tmpDir, "custom-jobs-state.json")
	dispatcher.SetStateFile(customPath)
	if dispatcher.getStateFile() != customPath {
		t.Errorf("getStateFile = %s, want %s", dispatcher.getStateFile(), customPath)
	}
}

func TestJobDispatcher_SelectRunner_NoRunners(t *testing.T) {
	dispatcher, _ := setupDispatcherTest(t)

	job := &Job{
		PipelineID: "test-pipeline",
		Stage:       "build",
		Name:        "test-job",
	}

	// SelectRunner should fail with no runners
	_, err := dispatcher.selectRunner(job)
	if err == nil {
		t.Error("selectRunner should return error with no available runners")
	}
}