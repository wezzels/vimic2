//go:build integration
// +build integration

// Package pool provides integration tests with real database
package pool

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
)

// TestPoolManager_CreatePool_RealDB tests pool creation with real database
func TestPoolManager_CreatePool_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pool-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create real database
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Test creating a pool
	pool, err := pm.CreatePool(nil, "test-pool", "ubuntu-22.04", 2, 10, 4, 8192)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	if pool.ID == "" {
		t.Error("expected pool ID to be set")
	}
	if pool.Name != "test-pool" {
		t.Errorf("expected pool name test-pool, got %s", pool.Name)
	}
	if pool.MinSize != 2 {
		t.Errorf("expected min size 2, got %d", pool.MinSize)
	}
	if pool.MaxSize != 10 {
		t.Errorf("expected max size 10, got %d", pool.MaxSize)
	}
}

// TestPoolManager_GetPool_RealDB tests pool retrieval with real database
func TestPoolManager_GetPool_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pool-test-*")
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

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Create a pool first
	createdPool, err := pm.CreatePool(nil, "test-pool-get", "ubuntu-22.04", 1, 5, 2, 4096)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Test getting the pool
	retrievedPool, err := pm.GetPool(createdPool.Name)
	if err != nil {
		t.Fatalf("failed to get pool: %v", err)
	}

	if retrievedPool.ID != createdPool.ID {
		t.Errorf("expected pool ID %s, got %s", createdPool.ID, retrievedPool.ID)
	}
	if retrievedPool.Name != createdPool.Name {
		t.Errorf("expected pool name %s, got %s", createdPool.Name, retrievedPool.Name)
	}
}

// TestPoolManager_ListPools_RealDB tests listing pools with real database
func TestPoolManager_ListPools_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pool-test-*")
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

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Create multiple pools
	poolNames := []string{"test-pool-a", "test-pool-b", "test-pool-c"}
	for _, name := range poolNames {
		_, err := pm.CreatePool(nil, name, "ubuntu-22.04", 1, 5, 2, 4096)
		if err != nil {
			t.Fatalf("failed to create pool %s: %v", name, err)
		}
	}

	// List all pools
	pools, err := pm.ListPools()
	if err != nil {
		t.Fatalf("failed to list pools: %v", err)
	}

	if len(pools) < 3 {
		t.Errorf("expected at least 3 pools, got %d", len(pools))
	}
}

// TestPoolManager_DeletePool_RealDB tests pool deletion with real database
func TestPoolManager_DeletePool_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-pool-test-*")
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

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Create a pool
	createdPool, err := pm.CreatePool(nil, "test-pool-delete", "ubuntu-22.04", 1, 5, 2, 4096)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Delete the pool
	err = pm.DeletePool(nil, createdPool.Name)
	if err != nil {
		t.Fatalf("failed to delete pool: %v", err)
	}

	// Verify it's deleted
	_, err = pm.GetPool(createdPool.Name)
	if err == nil {
		t.Error("expected error when getting deleted pool")
	}
}

// TestPoolManager_AllocateVM_RealDB tests VM allocation with real database
func TestPoolManager_AllocateVM_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-vm-test-*")
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

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Create a pool
	_, err = pm.CreatePool(nil, "test-pool-vm", "ubuntu-22.04", 1, 5, 2, 4096)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Allocate a VM
	vm, err := pm.AllocateVM("test-pool-vm")
	if err != nil {
		t.Fatalf("failed to allocate VM: %v", err)
	}

	if vm.ID == "" {
		t.Error("expected VM ID to be set")
	}
	if vm.Status == "" {
		t.Error("expected VM status to be set")
	}
}

// TestPoolManager_ReleaseVM_RealDB tests VM release with real database
func TestPoolManager_ReleaseVM_RealDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-vm-test-*")
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

	templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	pm, err := NewPoolManager(db, templateMgr, "")
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	pm.SetStateFile(filepath.Join(tmpDir, "pool-state.json"))
	defer pm.Close()

	// Create a pool and allocate a VM
	_, err = pm.CreatePool(nil, "test-pool-release", "ubuntu-22.04", 1, 5, 2, 4096)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	vm, err := pm.AllocateVM("test-pool-release")
	if err != nil {
		t.Fatalf("failed to allocate VM: %v", err)
	}

	// Release the VM
	err = pm.ReleaseVM(vm.ID)
	if err != nil {
		t.Fatalf("failed to release VM: %v", err)
	}
}

// TestTemplateManager_RealFS tests template management with real filesystem
func TestTemplateManager_RealFS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-template-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	basePath := filepath.Join(tmpDir, "base")
	overlayPath := filepath.Join(tmpDir, "overlay")

	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatalf("failed to create base path: %v", err)
	}
	if err := os.MkdirAll(overlayPath, 0755); err != nil {
		t.Fatalf("failed to create overlay path: %v", err)
	}

	// Create a template definition JSON file (as the manager expects)
	templateDef := `{"id":"ubuntu-22.04","name":"ubuntu-22.04","path":"ubuntu-22.04.qcow2","size":10737418240}`
	templateJSON := filepath.Join(basePath, "ubuntu-22.04.json")
	if err := os.WriteFile(templateJSON, []byte(templateDef), 0644); err != nil {
		t.Fatalf("failed to create template definition: %v", err)
	}

	// Also create the actual disk image file for completeness
	baseImage := filepath.Join(basePath, "ubuntu-22.04.qcow2")
	if err := os.WriteFile(baseImage, []byte("mock qcow2 image"), 0644); err != nil {
		t.Fatalf("failed to create mock base image: %v", err)
	}

	tmplMgr, err := NewTemplateManager(basePath, overlayPath)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Test listing templates
	templates := tmplMgr.ListTemplates()

	if len(templates) == 0 {
		t.Error("expected at least one template")
	}

	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "ubuntu-22.04" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find ubuntu-22.04 template")
	}
}

// TestDatabase_Persistence tests that pool state persists across manager restarts
func TestDatabase_Persistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-persistence-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	statePath := filepath.Join(tmpDir, "pool-state.json")

	// First session: create pool
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}

		templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		pm, err := NewPoolManager(db, templateMgr, "")
		if err != nil {
			t.Fatalf("failed to create pool manager: %v", err)
		}
		pm.SetStateFile(statePath)

		_, err = pm.CreatePool(nil, "persistent-pool", "ubuntu-22.04", 1, 5, 2, 4096)
		if err != nil {
			t.Fatalf("failed to create pool: %v", err)
		}

		// Force save state before closing
		pm.SaveState()
		pm.Close()
		db.Close()
	}

	// Verify state file was created
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatalf("state file was not created at %s", statePath)
	}

	// Second session: verify pool still exists
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer db.Close()

		templateMgr, err := NewTemplateManager(tmpDir, tmpDir)
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		pm, err := NewPoolManager(db, templateMgr, "")
		if err != nil {
			t.Fatalf("failed to create pool manager: %v", err)
		}
		pm.SetStateFile(statePath)
		defer pm.Close()

		// Force reload state
		if err := pm.LoadState(); err != nil {
			t.Fatalf("failed to load state: %v", err)
		}

		pool, err := pm.GetPool("persistent-pool")
		if err != nil {
			t.Fatalf("failed to get pool: %v", err)
		}

		if pool.Name != "persistent-pool" {
			t.Errorf("expected pool name persistent-pool, got %s", pool.Name)
		}
	}
}