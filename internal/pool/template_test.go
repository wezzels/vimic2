// Package pool provides template manager tests
package pool

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTemplate tests template structure
func TestTemplate_Create(t *testing.T) {
	template := &Template{
		ID:            "template-1",
		Name:          "ubuntu-22.04",
		Path:          "/var/lib/vimic2/templates/ubuntu-22.04.qcow2",
		BaseImage:     "ubuntu-22.04-server.qcow2",
		Size:          10737418240, // 10 GB
		OS:            "ubuntu",
		Arch:          "amd64",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}

	if template.ID != "template-1" {
		t.Errorf("expected template-1, got %s", template.ID)
	}
	if template.Name != "ubuntu-22.04" {
		t.Errorf("expected ubuntu-22.04, got %s", template.Name)
	}
	if template.OS != "ubuntu" {
		t.Errorf("expected ubuntu OS, got %s", template.OS)
	}
	if template.DefaultCPU != 2 {
		t.Errorf("expected 2 CPUs, got %d", template.DefaultCPU)
	}
	if template.DefaultMemory != 4096 {
		t.Errorf("expected 4096MB memory, got %d", template.DefaultMemory)
	}
}

// TestTemplate_JSON tests template JSON marshaling
func TestTemplate_JSON(t *testing.T) {
	template := &Template{
		ID:            "template-1",
		Name:          "ubuntu-22.04",
		Path:          "/var/lib/vimic2/templates/ubuntu-22.04.qcow2",
		BaseImage:     "ubuntu-22.04-server.qcow2",
		Size:          10737418240,
		OS:            "ubuntu",
		Arch:          "amd64",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}

	data, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("failed to marshal template: %v", err)
	}

	var template2 Template
	if err := json.Unmarshal(data, &template2); err != nil {
		t.Fatalf("failed to unmarshal template: %v", err)
	}

	if template2.ID != template.ID {
		t.Errorf("expected ID %s, got %s", template.ID, template2.ID)
	}
	if template2.Name != template.Name {
		t.Errorf("expected name %s, got %s", template.Name, template2.Name)
	}
	if template2.DefaultCPU != template.DefaultCPU {
		t.Errorf("expected CPU %d, got %d", template.DefaultCPU, template2.DefaultCPU)
	}
}

// TestOverlay_Structure tests overlay structure fields
func TestOverlay_Structure(t *testing.T) {
	now := time.Now()
	overlay := &Overlay{
		ID:         "overlay-2",
		TemplateID: "template-1",
		VMID:       "vm-1",
		Path:       "/var/lib/vimic2/overlays/vm-1.qcow2",
		CreatedAt:  now,
	}

	if overlay.ID != "overlay-2" {
		t.Errorf("expected overlay-2, got %s", overlay.ID)
	}
	if overlay.TemplateID != "template-1" {
		t.Errorf("expected template-1, got %s", overlay.TemplateID)
	}
	if overlay.VMID != "vm-1" {
		t.Errorf("expected vm-1, got %s", overlay.VMID)
	}
	if overlay.Path == "" {
		t.Error("overlay path should not be empty")
	}
}

// TestOverlay_JSONMarshaling tests overlay JSON marshaling
func TestOverlay_JSONMarshaling(t *testing.T) {
	now := time.Now()
	overlay := &Overlay{
		ID:         "overlay-2",
		TemplateID: "template-1",
		VMID:       "vm-1",
		Path:       "/var/lib/vimic2/overlays/vm-1.qcow2",
		CreatedAt:  now,
	}

	data, err := json.Marshal(overlay)
	if err != nil {
		t.Fatalf("failed to marshal overlay: %v", err)
	}

	var overlay2 Overlay
	if err := json.Unmarshal(data, &overlay2); err != nil {
		t.Fatalf("failed to unmarshal overlay: %v", err)
	}

	if overlay2.ID != overlay.ID {
		t.Errorf("expected ID %s, got %s", overlay.ID, overlay2.ID)
	}
	if overlay2.TemplateID != overlay.TemplateID {
		t.Errorf("expected TemplateID %s, got %s", overlay.TemplateID, overlay2.TemplateID)
	}
	if overlay2.VMID != overlay.VMID {
		t.Errorf("expected VMID %s, got %s", overlay.VMID, overlay2.VMID)
	}
}

// TestTemplateManager_CreateStruct tests template manager struct fields
func TestTemplateManager_CreateStruct(t *testing.T) {
	tm := &TemplateManager{
		basePath:    "/var/lib/vimic2/templates",
		overlayPath: "/var/lib/vimic2/overlays",
		templates:   make(map[string]*Template),
		overlays:    make(map[string]*Overlay),
	}

	if tm.basePath != "/var/lib/vimic2/templates" {
		t.Errorf("unexpected base path: %s", tm.basePath)
	}
	if tm.overlayPath != "/var/lib/vimic2/overlays" {
		t.Errorf("unexpected overlay path: %s", tm.overlayPath)
	}
	if tm.templates == nil {
		t.Error("templates map should not be nil")
	}
	if tm.overlays == nil {
		t.Error("overlays map should not be nil")
	}
}

// TestTemplateManager_AddTemplateToMap tests adding templates to the map
func TestTemplateManager_AddTemplateToMap(t *testing.T) {
	tm := &TemplateManager{
		basePath:    "/var/lib/vimic2/templates",
		overlayPath: "/var/lib/vimic2/overlays",
		templates:   make(map[string]*Template),
		overlays:    make(map[string]*Overlay),
	}

	template := &Template{
		ID:            "template-1",
		Name:          "ubuntu-22.04",
		DefaultCPU:    2,
		DefaultMemory: 4096,
		CreatedAt:     time.Now(),
	}

	tm.templates[template.ID] = template

	if len(tm.templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(tm.templates))
	}
	if tm.templates["template-1"] == nil {
		t.Error("template should exist in map")
	}
	if tm.templates["template-1"].Name != "ubuntu-22.04" {
		t.Errorf("expected name ubuntu-22.04, got %s", tm.templates["template-1"].Name)
	}
}

// TestTemplateManager_AddOverlayToMap tests adding overlays to the map
func TestTemplateManager_AddOverlayToMap(t *testing.T) {
	tm := &TemplateManager{
		basePath:    "/var/lib/vimic2/templates",
		overlayPath: "/var/lib/vimic2/overlays",
		templates:   make(map[string]*Template),
		overlays:    make(map[string]*Overlay),
	}

	overlay := &Overlay{
		ID:         "overlay-1",
		TemplateID: "template-1",
		VMID:       "vm-1",
		Path:       "/var/lib/vimic2/overlays/vm-1.qcow2",
		CreatedAt:  time.Now(),
	}

	tm.overlays[overlay.ID] = overlay

	if len(tm.overlays) != 1 {
		t.Errorf("expected 1 overlay, got %d", len(tm.overlays))
	}
	if tm.overlays["overlay-1"] == nil {
		t.Error("overlay should exist in map")
	}
	if tm.overlays["overlay-1"].VMID != "vm-1" {
		t.Errorf("expected VMID vm-1, got %s", tm.overlays["overlay-1"].VMID)
	}
}

// TestTemplate_DefaultValues tests template default values
func TestTemplate_DefaultValues(t *testing.T) {
	template := &Template{
		ID:        "template-1",
		Name:      "minimal",
		CreatedAt: time.Now(),
	}

	// Test that defaults can be set
	if template.DefaultCPU == 0 {
		template.DefaultCPU = 2
	}
	if template.DefaultMemory == 0 {
		template.DefaultMemory = 2048
	}

	if template.DefaultCPU != 2 {
		t.Errorf("expected default CPU 2, got %d", template.DefaultCPU)
	}
	if template.DefaultMemory != 2048 {
		t.Errorf("expected default memory 2048, got %d", template.DefaultMemory)
	}
}

// TestOverlay_DestroyedAt tests overlay destruction
func TestOverlay_DestroyedAt(t *testing.T) {
	now := time.Now()
	destroyed := now.Add(1 * time.Hour)

	overlay := &Overlay{
		ID:          "overlay-3",
		TemplateID:  "template-1",
		VMID:        "vm-1",
		Path:        "/var/lib/vimic2/overlays/vm-1.qcow2",
		CreatedAt:   now,
		DestroyedAt: &destroyed,
	}

	if overlay.DestroyedAt == nil {
		t.Error("DestroyedAt should not be nil")
	}
	if overlay.DestroyedAt.Before(overlay.CreatedAt) {
		t.Error("DestroyedAt should be after CreatedAt")
	}
}
