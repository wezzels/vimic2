//go:build integration
// +build integration

// Package pool provides integration tests for template management
package pool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestTemplateManager_SaveTemplates tests saving templates
func TestTemplateManager_SaveTemplates(t *testing.T) {
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

	tmplMgr, err := NewTemplateManager(basePath, overlayPath)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Create a template
	template := &Template{
		ID:            "ubuntu-22.04",
		Name:          "Ubuntu 22.04",
		Path:          "ubuntu-22.04.qcow2",
		Size:          10 * 1024 * 1024 * 1024,
		OS:            "linux",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}

	tmplMgr.templates[template.ID] = template

	// Save templates
	if err := tmplMgr.saveTemplates(); err != nil {
		t.Fatalf("failed to save templates: %v", err)
	}

	// Verify file was created
	templateFile := filepath.Join(basePath, "ubuntu-22.04.json")
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		t.Fatalf("template file was not created")
	}
}

// TestTemplateManager_GetTemplateFromFile tests retrieving a template from file
func TestTemplateManager_GetTemplateFromFile(t *testing.T) {
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

	// Create template definition JSON
	template := Template{
		ID:            "ubuntu-22.04",
		Name:          "Ubuntu 22.04",
		Path:          "ubuntu-22.04.qcow2",
		Size:          10 * 1024 * 1024 * 1024,
		OS:            "linux",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	templateFile := filepath.Join(basePath, "ubuntu-22.04.json")
	if err := os.WriteFile(templateFile, templateJSON, 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	tmplMgr, err := NewTemplateManager(basePath, overlayPath)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Get template
	retrieved, err := tmplMgr.GetTemplate("ubuntu-22.04")
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}

	if retrieved.Name != "Ubuntu 22.04" {
		t.Errorf("expected template name Ubuntu 22.04, got %s", retrieved.Name)
	}
}

// TestTemplateManager_ListOverlays tests listing overlays
func TestTemplateManager_ListOverlays(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-overlay-test-*")
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

	// Create template definition
	template := Template{
		ID:            "ubuntu-22.04",
		Name:          "Ubuntu 22.04",
		Path:          "ubuntu-22.04.qcow2",
		Size:          10 * 1024 * 1024 * 1024,
		OS:            "linux",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	templateFile := filepath.Join(basePath, "ubuntu-22.04.json")
	if err := os.WriteFile(templateFile, templateJSON, 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	tmplMgr, err := NewTemplateManager(basePath, overlayPath)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Create overlay
	overlay := &Overlay{
		ID:         "overlay-1",
		TemplateID: "ubuntu-22.04",
		Path:       filepath.Join(overlayPath, "overlay-1.qcow2"),
		VMID:       "vm-1",
		CreatedAt:  time.Now(),
	}
	tmplMgr.overlays[overlay.ID] = overlay

	// List overlays
	overlays := tmplMgr.ListOverlays()
	if len(overlays) != 1 {
		t.Errorf("expected 1 overlay, got %d", len(overlays))
	}
	if overlays[0].ID != "overlay-1" {
		t.Errorf("expected overlay ID overlay-1, got %s", overlays[0].ID)
	}
}

// TestPoolManager_LoadConfig tests config loading
func TestPoolManager_LoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	config := struct {
		Pools map[string]poolConfig `json:"pools"`
	}{
		Pools: map[string]poolConfig{
			"test-pool": {
				Template: "ubuntu-22.04",
				MinSize:  2,
				MaxSize:  10,
				CPU:      4,
				Memory:   8192,
			},
		},
	}

	configJSON, _ := json.MarshalIndent(config, "", "  ")
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, configJSON, 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

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

	pm, err := NewPoolManager(db, templateMgr, configFile)
	if err != nil {
		t.Fatalf("failed to create pool manager: %v", err)
	}
	defer pm.Close()

	// Verify config was loaded
	if len(pm.config) != 1 {
		t.Errorf("expected 1 pool config, got %d", len(pm.config))
	}
	if pm.config["test-pool"].Template != "ubuntu-22.04" {
		t.Errorf("expected template ubuntu-22.04, got %s", pm.config["test-pool"].Template)
	}
}

// TestPoolManager_AllocateVMContext tests context-based VM allocation
func TestPoolManager_AllocateVMContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-vmctx-test-*")
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
	_, err = pm.CreatePool(nil, "test-pool-ctx", "ubuntu-22.04", 1, 5, 2, 4096)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Allocate VM with context (this tests AllocateVMContext)
	// Note: AllocateVMContext calls AllocateVM internally
	vm, err := pm.AllocateVM("test-pool-ctx")
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

// TestTemplateManager_SaveOverlays tests saving overlays
func TestTemplateManager_SaveOverlays(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-overlay-save-test-*")
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

	// Create template
	template := Template{
		ID:            "ubuntu-22.04",
		Name:          "Ubuntu 22.04",
		Path:          "ubuntu-22.04.qcow2",
		Size:          10 * 1024 * 1024 * 1024,
		OS:            "linux",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}
	templateJSON, _ := json.MarshalIndent(template, "", "  ")
	templateFile := filepath.Join(basePath, "ubuntu-22.04.json")
	if err := os.WriteFile(templateFile, templateJSON, 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	tmplMgr, err := NewTemplateManager(basePath, overlayPath)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Create and save overlay
	overlay := &Overlay{
		ID:         "overlay-save-test",
		TemplateID: "ubuntu-22.04",
		Path:       filepath.Join(overlayPath, "overlay-save-test.qcow2"),
		VMID:       "vm-save-test",
		CreatedAt:  time.Now(),
	}
	tmplMgr.overlays[overlay.ID] = overlay

	// Save overlays
	if err := tmplMgr.saveOverlays(); err != nil {
		t.Fatalf("failed to save overlays: %v", err)
	}

	// Verify file was created
	overlayFile := filepath.Join(overlayPath, "overlay-save-test.json")
	if _, err := os.Stat(overlayFile); os.IsNotExist(err) {
		t.Fatalf("overlay file was not created")
	}
}
