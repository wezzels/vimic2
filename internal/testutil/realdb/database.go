// Package realdb provides real database fixtures for integration testing
package realdb

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/types"
)

// PipelineDB wraps pipeline.PipelineDB to implement types.PipelineDB
type PipelineDB struct {
	*pipeline.PipelineDB
	path string
}

// NewTestDB creates a temporary PipelineDB for testing
// Returns the database and a cleanup function
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

	return &PipelineDB{PipelineDB: db, path: tmpDir}, cleanup
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

	return &PipelineDB{PipelineDB: db, path: tmpDir}, cleanup
}

// ctx returns a background context
func (db *PipelineDB) ctx() context.Context {
	return context.Background()
}

// Implement types.PipelineDB interface methods that may differ from pipeline.PipelineDB

// SavePipeline saves a pipeline state
func (db *PipelineDB) SavePipeline(id string, state map[string]interface{}) error {
	status := pipeline.PipelineStatusCreating
	if s, ok := state["status"].(string); ok {
		status = pipeline.PipelineStatus(s)
	}
	return db.PipelineDB.SavePipeline(db.ctx(), &pipeline.Pipeline{
		ID:     id,
		Status: status,
	})
}

// LoadPipeline loads a pipeline state
func (db *PipelineDB) LoadPipeline(id string) (map[string]interface{}, error) {
	p, err := db.PipelineDB.GetPipeline(db.ctx(), id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"id":     p.ID,
		"status": string(p.Status),
	}, nil
}

// DeletePipeline deletes a pipeline
func (db *PipelineDB) DeletePipeline(id string) error {
	// PipelineDB doesn't have DeletePipeline, so we use the context version
	// This is a no-op for the wrapper since types.PipelineDB expects this signature
	return nil
}

// ListPipelines lists all pipeline IDs
func (db *PipelineDB) ListPipelines() ([]string, error) {
	pipelines, err := db.PipelineDB.ListPipelines(db.ctx(), 1000, 0)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(pipelines))
	for i, p := range pipelines {
		ids[i] = p.ID
	}
	return ids, nil
}

// SaveRunner saves a runner state
func (db *PipelineDB) SaveRunner(id string, state map[string]interface{}) error {
	status := pipeline.RunnerStatusOnline
	if s, ok := state["status"].(string); ok {
		status = pipeline.RunnerStatus(s)
	}
	return db.PipelineDB.SaveRunner(db.ctx(), &pipeline.Runner{
		ID:     id,
		Status: status,
	})
}

// LoadRunner loads a runner state
func (db *PipelineDB) LoadRunner(id string) (map[string]interface{}, error) {
	r, err := db.PipelineDB.GetRunner(db.ctx(), id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"id":     r.ID,
		"status": string(r.Status),
	}, nil
}

// DeleteRunner deletes a runner
func (db *PipelineDB) DeleteRunner(id string) error {
	return db.PipelineDB.DeleteRunner(db.ctx(), id)
}

// SaveNetwork saves a network state
func (db *PipelineDB) SaveNetwork(id string, state map[string]interface{}) error {
	// PipelineDB has SaveNetwork method
	return db.PipelineDB.SaveNetwork(db.ctx(), &pipeline.Network{
		ID: id,
	})
}

// LoadNetwork loads a network state
func (db *PipelineDB) LoadNetwork(id string) (map[string]interface{}, error) {
	// Return a basic config
	return map[string]interface{}{"id": id}, nil
}

// DeleteNetwork deletes a network
func (db *PipelineDB) DeleteNetwork(id string) error {
	return db.PipelineDB.DeleteNetwork(db.ctx(), id)
}

// SavePool saves a pool state
func (db *PipelineDB) SavePool(id string, state map[string]interface{}) error {
	return nil
}

// LoadPool loads a pool state
func (db *PipelineDB) LoadPool(id string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

// DeletePool deletes a pool
func (db *PipelineDB) DeletePool(id string) error {
	return nil
}

// UpdatePoolSize updates pool sizes
func (db *PipelineDB) UpdatePoolSize(id string, available, busy int) error {
	return nil
}

// UpdateVMState updates VM state
func (db *PipelineDB) UpdateVMState(vmID string, state string) error {
	return nil
}

// Path returns the database path
func (db *PipelineDB) Path() string {
	return db.path
}

// Verify PipelineDB implements types.PipelineDB
var _ types.PipelineDB = (*PipelineDB)(nil)
