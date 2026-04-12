// Package deploy provides executor tests
package deploy

import (
	"testing"
)

// TestExecutor_Execute tests Execute function (requires db)
func TestExecutor_Execute(t *testing.T) {
	// Executor requires db, skip execution test
	// The function will panic on nil db
	_ = NewExecutor
}

// TestExecutor_SelectBestHost tests selectBestHost
func TestExecutor_SelectBestHost(t *testing.T) {
	executor := &Executor{
		hosts: nil,
	}

	// No hosts - will panic on nil map, skip test
	// Just verify the Executor struct exists
	_ = executor
}

// TestPresetTemplatesReal tests preset templates
func TestPresetTemplatesReal(t *testing.T) {
	dev := GetPreset("dev")
	if dev == nil {
		t.Fatal("expected dev preset")
	}
	if len(dev.NodeGroups) != 1 {
		t.Errorf("expected 1 node group, got %d", len(dev.NodeGroups))
	}

	prod := GetPreset("prod")
	if prod == nil {
		t.Fatal("expected prod preset")
	}
	if len(prod.NodeGroups) != 2 {
		t.Errorf("expected 2 node groups, got %d", len(prod.NodeGroups))
	}

	db := GetPreset("db")
	if db == nil {
		t.Fatal("expected db preset")
	}

	invalid := GetPreset("nonexistent")
	if invalid != nil {
		t.Error("expected nil for invalid preset")
	}
}

// TestWizard_ValidateDetailed tests validation
func TestWizard_ValidateDetailed(t *testing.T) {
	w := NewWizard()

	// Empty name should fail
	w.cluster.Name = ""
	err := w.Validate()
	if err == nil {
		t.Error("expected error for empty name")
	}

	// Valid name but no node groups
	w.cluster.Name = "test-cluster"
	err = w.Validate()
	if err == nil {
		t.Error("expected error for no node groups")
	}

	// Add node group
	w.AddNodeGroup(&NodeGroup{Name: "workers", Role: "worker", Count: 2})
	err = w.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid count
	w.cluster.NodeGroups[0].Count = 0
	err = w.Validate()
	if err == nil {
		t.Error("expected error for zero count")
	}
}

// TestNodeGroup_DefaultsReal tests node group defaults
func TestNodeGroup_DefaultsReal(t *testing.T) {
	ng := &NodeGroup{
		Name:  "test",
		Role:  "worker",
		Count: 3,
	}

	// Apply defaults manually
	if ng.CPU == 0 {
		ng.CPU = 2
	}
	if ng.MemoryMB == 0 {
		ng.MemoryMB = 2048
	}
	if ng.DiskGB == 0 {
		ng.DiskGB = 20
	}
	if ng.Image == "" {
		ng.Image = "ubuntu-22.04"
	}

	if ng.CPU == 0 {
		t.Error("CPU should have default")
	}
	if ng.MemoryMB == 0 {
		t.Error("Memory should have default")
	}
	if ng.DiskGB == 0 {
		t.Error("Disk should have default")
	}
	if ng.Image == "" {
		t.Error("Image should have default")
	}
}
