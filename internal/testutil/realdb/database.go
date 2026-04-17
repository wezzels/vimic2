// Package realdb provides real database fixtures for integration testing
package realdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/types"
)

// PipelineDB wraps pipeline.PipelineDBAdapter to implement types.PipelineDB
type PipelineDB struct {
	*pipeline.PipelineDBAdapter
	path string
}

// NewTestDB creates a temporary PipelineDB for testing
func NewTestDB(t *testing.T) (*PipelineDB, func()) {
	tmpDir, err := os.MkdirTemp("", "vimic2-test-*.d")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create PipelineDB: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	adapter := pipeline.NewPipelineDBAdapter(db)
	return &PipelineDB{PipelineDBAdapter: adapter, path: tmpDir}, cleanup
}

// MustNewTestDB creates a test DB or panics
func MustNewTestDB() (*PipelineDB, func()) {
	tmpDir, err := os.MkdirTemp("", "vimic2-test-*.d")
	if err != nil {
		panic(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		panic(err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	adapter := pipeline.NewPipelineDBAdapter(db)
	return &PipelineDB{PipelineDBAdapter: adapter, path: tmpDir}, cleanup
}

// Path returns the database path
func (db *PipelineDB) Path() string {
	return db.path
}

// Verify PipelineDB implements types.PipelineDB
var _ types.PipelineDB = (*PipelineDB)(nil)
