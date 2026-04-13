//go:build integration
// +build integration

// Package database provides integration tests for PipelineDB methods
package database

import (
	"os"
	"testing"
)

// TestDB_SaveLoadPipeline tests SavePipeline and LoadPipeline
func TestDB_SaveLoadPipeline(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pipeline-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Test SavePipeline
	state := map[string]interface{}{
		"status":     "running",
		"platform":   "gitlab",
		"repository": "https://gitlab.example.com/test/repo",
		"branch":     "main",
	}
	err = db.SavePipeline("pipeline-1", state)
	if err != nil {
		t.Fatalf("failed to save pipeline: %v", err)
	}

	// Test LoadPipeline
	loaded, err := db.LoadPipeline("pipeline-1")
	if err != nil {
		t.Fatalf("failed to load pipeline: %v", err)
	}

	if loaded["status"] != "running" {
		t.Errorf("expected status running, got %v", loaded["status"])
	}
	if loaded["platform"] != "gitlab" {
		t.Errorf("expected platform gitlab, got %v", loaded["platform"])
	}
}

// TestDB_DeletePipeline tests DeletePipeline
func TestDB_DeletePipeline(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pipeline-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create a pipeline
	state := map[string]interface{}{"status": "created"}
	err = db.SavePipeline("pipeline-2", state)
	if err != nil {
		t.Fatalf("failed to save pipeline: %v", err)
	}

	// Delete it
	err = db.DeletePipeline("pipeline-2")
	if err != nil {
		t.Fatalf("failed to delete pipeline: %v", err)
	}

	// Verify it's gone
	_, err = db.LoadPipeline("pipeline-2")
	if err == nil {
		t.Error("expected error when loading deleted pipeline")
	}
}

// TestDB_ListPipelines tests ListPipelines
func TestDB_ListPipelines(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pipeline-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create multiple pipelines
	for i := 0; i < 3; i++ {
		state := map[string]interface{}{"status": "running"}
		err = db.SavePipeline("pipeline-"+string(rune('a'+i)), state)
		if err != nil {
			t.Fatalf("failed to save pipeline: %v", err)
		}
	}

	// List pipelines
	ids, err := db.ListPipelines()
	if err != nil {
		t.Fatalf("failed to list pipelines: %v", err)
	}

	if len(ids) < 3 {
		t.Errorf("expected at least 3 pipelines, got %d", len(ids))
	}
}

// TestDB_SaveLoadRunner tests SaveRunner and LoadRunner
func TestDB_SaveLoadRunner(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-runner-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Test SaveRunner
	state := map[string]interface{}{
		"status":     "online",
		"platform":   "gitlab",
		"pool_name":  "test-pool",
		"vm_id":      "vm-1",
	}
	err = db.SaveRunner("runner-1", state)
	if err != nil {
		t.Fatalf("failed to save runner: %v", err)
	}

	// Test LoadRunner
	loaded, err := db.LoadRunner("runner-1")
	if err != nil {
		t.Fatalf("failed to load runner: %v", err)
	}

	if loaded["status"] != "online" {
		t.Errorf("expected status online, got %v", loaded["status"])
	}
	if loaded["platform"] != "gitlab" {
		t.Errorf("expected platform gitlab, got %v", loaded["platform"])
	}
}

// TestDB_DeleteRunner tests DeleteRunner
func TestDB_DeleteRunner(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-runner-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create a runner
	state := map[string]interface{}{"status": "online"}
	err = db.SaveRunner("runner-2", state)
	if err != nil {
		t.Fatalf("failed to save runner: %v", err)
	}

	// Delete it
	err = db.DeleteRunner("runner-2")
	if err != nil {
		t.Fatalf("failed to delete runner: %v", err)
	}

	// Verify it's gone
	_, err = db.LoadRunner("runner-2")
	if err == nil {
		t.Error("expected error when loading deleted runner")
	}
}

// TestDB_SaveLoadNetwork tests SaveNetwork and LoadNetwork
func TestDB_SaveLoadNetwork(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Test SaveNetwork
	state := map[string]interface{}{
		"vlan":   100,
		"cidr":   "10.100.0.0/24",
		"status": "active",
	}
	err = db.SaveNetwork("network-1", state)
	if err != nil {
		t.Fatalf("failed to save network: %v", err)
	}

	// Test LoadNetwork
	loaded, err := db.LoadNetwork("network-1")
	if err != nil {
		t.Fatalf("failed to load network: %v", err)
	}

	// JSON unmarshals numbers as float64
	if loaded["vlan"] != float64(100) {
		t.Errorf("expected vlan 100, got %v", loaded["vlan"])
	}
	if loaded["cidr"] != "10.100.0.0/24" {
		t.Errorf("expected cidr 10.100.0.0/24, got %v", loaded["cidr"])
	}
}

// TestDB_DeleteNetwork tests DeleteNetwork
func TestDB_DeleteNetwork(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-network-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create a network
	state := map[string]interface{}{"vlan": 200}
	err = db.SaveNetwork("network-2", state)
	if err != nil {
		t.Fatalf("failed to save network: %v", err)
	}

	// Delete it
	err = db.DeleteNetwork("network-2")
	if err != nil {
		t.Fatalf("failed to delete network: %v", err)
	}

	// Verify it's gone
	_, err = db.LoadNetwork("network-2")
	if err == nil {
		t.Error("expected error when loading deleted network")
	}
}

// TestDB_SaveLoadPool tests SavePool and LoadPool
func TestDB_SaveLoadPool(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pool-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Test SavePool
	state := map[string]interface{}{
		"name":      "test-pool",
		"template":  "ubuntu-22.04",
		"min_size":  2,
		"max_size":  10,
		"available": 5,
		"busy":      3,
	}
	err = db.SavePool("pool-1", state)
	if err != nil {
		t.Fatalf("failed to save pool: %v", err)
	}

	// Test LoadPool
	loaded, err := db.LoadPool("pool-1")
	if err != nil {
		t.Fatalf("failed to load pool: %v", err)
	}

	if loaded["name"] != "test-pool" {
		t.Errorf("expected name test-pool, got %v", loaded["name"])
	}
	if loaded["template"] != "ubuntu-22.04" {
		t.Errorf("expected template ubuntu-22.04, got %v", loaded["template"])
	}
}

// TestDB_DeletePool tests DeletePool
func TestDB_DeletePool(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pool-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create a pool
	state := map[string]interface{}{"name": "test-pool"}
	err = db.SavePool("pool-2", state)
	if err != nil {
		t.Fatalf("failed to save pool: %v", err)
	}

	// Delete it
	err = db.DeletePool("pool-2")
	if err != nil {
		t.Fatalf("failed to delete pool: %v", err)
	}

	// Verify it's gone
	_, err = db.LoadPool("pool-2")
	if err == nil {
		t.Error("expected error when loading deleted pool")
	}
}

// TestDB_UpdatePoolSize tests UpdatePoolSize
func TestDB_UpdatePoolSize(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-pool-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create a pool
	state := map[string]interface{}{"name": "test-pool"}
	err = db.SavePool("pool-3", state)
	if err != nil {
		t.Fatalf("failed to save pool: %v", err)
	}

	// Update pool size
	err = db.UpdatePoolSize("pool-3", 8, 2)
	if err != nil {
		t.Fatalf("failed to update pool size: %v", err)
	}

	// The update should succeed
}

// TestDB_UpdateVMState tests UpdateVMState
func TestDB_UpdateVMState(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-vm-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// UpdateVMState should succeed even if VM doesn't exist
	// (it's an update operation)
	err = db.UpdateVMState("vm-1", "running")
	if err != nil {
		t.Fatalf("failed to update VM state: %v", err)
	}
}