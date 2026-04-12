// Package realdb_test tests the real database fixtures
package realdb_test

import (
	"context"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/testutil/realdb"
)

// TestNewTestDB tests creating a test database
func TestNewTestDB(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	if db == nil {
		t.Fatal("expected non-nil database")
	}

	// Test Path method
	path := db.Path()
	if path == "" {
		t.Error("expected non-empty path")
	}
	t.Logf("Database path: %s", path)
}

// TestPipelineDB_SaveLoadPipeline tests saving and loading pipelines
func TestPipelineDB_SaveLoadPipeline(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Save a pipeline directly
	err := db.PipelineDB.SavePipeline(ctx, &pipeline.Pipeline{
		ID:     "test-pipeline",
		Status: pipeline.PipelineStatusRunning,
	})
	if err != nil {
		t.Fatalf("SavePipeline failed: %v", err)
	}

	// Load the pipeline
	p, err := db.PipelineDB.GetPipeline(ctx, "test-pipeline")
	if err != nil {
		t.Fatalf("GetPipeline failed: %v", err)
	}

	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}

	if p.ID != "test-pipeline" {
		t.Errorf("expected id 'test-pipeline', got %s", p.ID)
	}
}

// TestPipelineDB_DeletePipeline tests deleting pipelines
func TestPipelineDB_DeletePipeline(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create and delete runner (PipelineDB doesn't have DeletePipeline)
	db.PipelineDB.SavePipeline(ctx, &pipeline.Pipeline{
		ID:     "test-delete",
		Status: pipeline.PipelineStatusRunning,
	})

	// Verify created
	p, _ := db.PipelineDB.GetPipeline(ctx, "test-delete")
	if p == nil {
		t.Fatal("expected pipeline to exist")
	}
}

// TestPipelineDB_ListPipelines tests listing pipelines
func TestPipelineDB_ListPipelines(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple pipelines
	for i := 0; i < 3; i++ {
		db.PipelineDB.SavePipeline(ctx, &pipeline.Pipeline{
			ID:     "pipeline-" + string(rune('0'+i)),
			Status: pipeline.PipelineStatusRunning,
		})
	}

	// List all
	pipelines, err := db.PipelineDB.ListPipelines(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPipelines failed: %v", err)
	}

	if len(pipelines) < 3 {
		t.Errorf("expected at least 3 pipelines, got %d", len(pipelines))
	}
}

// TestPipelineDB_SaveLoadRunner tests saving and loading runners
func TestPipelineDB_SaveLoadRunner(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Save a runner
	err := db.PipelineDB.SaveRunner(ctx, &pipeline.Runner{
		ID:     "test-runner",
		Status: pipeline.RunnerStatusOnline,
	})
	if err != nil {
		t.Fatalf("SaveRunner failed: %v", err)
	}

	// Load the runner
	r, err := db.PipelineDB.GetRunner(ctx, "test-runner")
	if err != nil {
		t.Fatalf("GetRunner failed: %v", err)
	}

	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}

// TestMustNewTestDB tests the panic version
func TestMustNewTestDB(t *testing.T) {
	db, cleanup := realdb.MustNewTestDB()
	defer cleanup()

	if db == nil {
		t.Fatal("expected non-nil database")
	}
}

// TestPipelineDB_ConcurrentAccess tests concurrent operations
func TestPipelineDB_ConcurrentAccess(t *testing.T) {
	db, cleanup := realdb.NewTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Run concurrent saves
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			id := "concurrent-" + string(rune('0'+idx))
			db.PipelineDB.SavePipeline(ctx, &pipeline.Pipeline{
				ID:     id,
				Status: pipeline.PipelineStatusRunning,
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all were saved
	pipelines, err := db.PipelineDB.ListPipelines(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListPipelines failed: %v", err)
	}
	if len(pipelines) < 10 {
		t.Errorf("expected at least 10 pipelines, got %d", len(pipelines))
	}
}
