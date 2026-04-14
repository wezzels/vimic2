//go:build integration

package provisioner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// ==================== Manager Tests ====================

func TestManager_New(t *testing.T) {
	mgr := NewManager("/var/lib/libvirt/images")
	if mgr == nil {
		t.Fatal("manager should not be nil")
	}
	if mgr.imageDir != "/var/lib/libvirt/images" {
		t.Errorf("expected /var/lib/libvirt/images, got %s", mgr.imageDir)
	}
}

func TestManager_DefaultImageDir_Int(t *testing.T) {
	mgr := NewManager("")
	if mgr.imageDir != "/var/lib/libvirt/images" {
		t.Errorf("expected default image dir, got %s", mgr.imageDir)
	}
}

func TestManager_CustomImageDir_Int(t *testing.T) {
	mgr := NewManager("/custom/path/images")
	if mgr.imageDir != "/custom/path/images" {
		t.Errorf("expected /custom/path/images, got %s", mgr.imageDir)
	}
}

// TestManager_EnsureImage tests that EnsureImage returns an existing image path
func TestManager_EnsureImage_ExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake existing image
	fakeImage := filepath.Join(tmpDir, "existing.qcow2")
	if err := os.WriteFile(fakeImage, []byte("fake qcow2"), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(tmpDir)
	path, err := mgr.EnsureImage("existing", "ubuntu", "22.04")
	if err != nil {
		t.Fatalf("EnsureImage for existing file failed: %v", err)
	}
	if path != fakeImage {
		t.Errorf("Expected %s, got %s", fakeImage, path)
	}
}

// TestManager_EnsureImage tests image creation fallback (qemu-img not available)
func TestManager_EnsureImage_NoQemuImg(t *testing.T) {
	t.Skip("skipped: requires qemu-img and network access")
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)

	// This will fail without qemu-img but tests the code path
	_, err = mgr.EnsureImage("test-ubuntu", "ubuntu", "22.04")
	// Error is expected if qemu-img is not available
	if err != nil {
		t.Logf("EnsureImage with no qemu-img: %v (expected)", err)
	}
}

// TestManager_DeleteImage tests deleting a non-existent image
func TestManager_DeleteImage_NonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	err = mgr.DeleteImage("nonexistent")
	// Should not error on non-existent file
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("DeleteImage should handle non-existent file gracefully: %v", err)
	}
}

// TestManager_DeleteImage_Existing tests deleting an existing image
func TestManager_DeleteImage_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake image
	fakeImage := filepath.Join(tmpDir, "delete-me.qcow2")
	if err := os.WriteFile(fakeImage, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(tmpDir)
	err = mgr.DeleteImage("delete-me")
	if err != nil {
		t.Fatalf("DeleteImage failed: %v", err)
	}

	if _, err := os.Stat(fakeImage); !os.IsNotExist(err) {
		t.Error("Image should be deleted")
	}
}

// TestManager_CloneImage tests cloning a non-existent base image
func TestManager_CloneImage_NoBase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	_, err = mgr.CloneImage("/nonexistent/base.qcow2", "clone-node")
	// Expected to fail without qemu-img or base image
	if err != nil {
		t.Logf("CloneImage without base: %v (expected)", err)
	}
}

// TestManager_CloneImage tests cloning with an existing base
func TestManager_CloneImage_WithBase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake base image
	baseImage := filepath.Join(tmpDir, "base.qcow2")
	if err := os.WriteFile(baseImage, []byte("fake base qcow2"), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(tmpDir)
	path, err := mgr.CloneImage(baseImage, "clone-node")
	if err != nil {
		t.Logf("CloneImage: %v (qemu-img may not be available)", err)
	} else {
		t.Logf("Cloned to: %s", path)
		// Clean up cloned image
		os.Remove(path)
	}
}

// TestManager_ResizeImage tests resizing a non-existent image
func TestManager_ResizeImage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(tmpDir)
	err = mgr.ResizeImage(filepath.Join(tmpDir, "nonexistent.qcow2"), 50)
	// Expected to fail without qemu-img or file
	if err != nil {
		t.Logf("ResizeImage: %v (expected)", err)
	}
}

// ==================== NetworkManager Tests ====================

func TestNetworkManager_EnsureNetwork(t *testing.T) {
	nm := &NetworkManager{}
	ctx := context.Background()

	// This will fail without virsh but tests the code path
	err := nm.EnsureNetwork(ctx, "test-network", "nat", "192.168.122.0/24")
	if err != nil {
		t.Logf("EnsureNetwork: %v (virsh may not be available)", err)
	}
}

func TestNetworkManager_DeleteNetwork(t *testing.T) {
	nm := &NetworkManager{}
	ctx := context.Background()

	// This will fail without virsh but tests the code path
	err := nm.DeleteNetwork(ctx, "test-network")
	if err != nil {
		t.Logf("DeleteNetwork: %v (virsh may not be available)", err)
	}
}

func TestNetworkManager_GenerateNetworkXML_NAT(t *testing.T) {
	nm := &NetworkManager{}
	xml := nm.generateNetworkXML("test-nat", "nat", "192.168.122.0/24")

	if xml == "" {
		t.Error("generateNetworkXML should not return empty string")
	}

	// Verify XML structure
	checks := []string{"<network>", "</network>", "test-nat", "mode='nat'", "<dhcp>"}
	for _, check := range checks {
		if !contains(xml, check) {
			t.Errorf("NAT XML missing: %s", check)
		}
	}
}

func TestNetworkManager_GenerateNetworkXML_Bridge(t *testing.T) {
	nm := &NetworkManager{}
	xml := nm.generateNetworkXML("test-bridge", "bridge", "10.100.0.0/16")

	if xml == "" {
		t.Error("generateNetworkXML should not return empty string")
	}

	checks := []string{"<network>", "</network>", "test-bridge", "mode='bridge'", "brtest-bridge"}
	for _, check := range checks {
		if !contains(xml, check) {
			t.Errorf("Bridge XML missing: %s", check)
		}
	}
}

func TestNetworkManager_GenerateNetworkXML_EmptyName(t *testing.T) {
	nm := &NetworkManager{}
	xml := nm.generateNetworkXML("", "nat", "")
	if !contains(xml, "<network>") {
		t.Error("Should generate valid XML even with empty name")
	}
}

func TestHash_Basic_Int(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "test"},
		{"empty", ""},
		{"long", "very-long-network-name"},
		{"numbers", "network123"},
		{"special", "network-with-dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash(tt.input)
			if result < 0 || result > 255 {
				t.Errorf("hash(%s) = %d, expected range [0, 255]", tt.input, result)
			}
		})
	}
}

func TestHash_Consistency_Int(t *testing.T) {
	h1 := hash("test-network")
	h2 := hash("test-network")
	if h1 != h2 {
		t.Errorf("hash should be consistent: %d != %d", h1, h2)
	}
}

func TestNetworkConfig_Struct_Int(t *testing.T) {
	config := &NetworkConfig{
		Name:    "default",
		Type:    "nat",
		CIDR:    "192.168.122.0/24",
		Gateway: "192.168.122.1",
	}

	if config.Name != "default" {
		t.Errorf("expected default, got %s", config.Name)
	}
	if config.Type != "nat" {
		t.Errorf("expected nat type, got %s", config.Type)
	}
	if config.CIDR != "192.168.122.0/24" {
		t.Errorf("expected CIDR 192.168.122.0/24, got %s", config.CIDR)
	}
	if config.Gateway != "192.168.122.1" {
		t.Errorf("expected gateway 192.168.122.1, got %s", config.Gateway)
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}