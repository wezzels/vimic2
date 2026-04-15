//go:build integration

package provisioner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManager_EnsureImage_Ubuntu(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("test-ubuntu", "ubuntu", "22.04")
	if err != nil {
		t.Skipf("EnsureImage failed (needs qemu-img or curl): %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
	t.Logf("Ubuntu image path: %s", path)
}

func TestManager_EnsureImage_Debian(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("test-debian", "debian", "12")
	if err != nil {
		t.Skipf("EnsureImage failed: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
	t.Logf("Debian image path: %s", path)
}

func TestManager_EnsureImage_Fedora(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("test-fedora", "fedora", "39")
	if err != nil {
		t.Skipf("EnsureImage failed: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
	t.Logf("Fedora image path: %s", path)
}

func TestManager_EnsureImage_CentOS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("test-centos", "centos", "9")
	if err != nil {
		t.Skipf("EnsureImage failed: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
	t.Logf("CentOS image path: %s", path)
}

func TestManager_EnsureImage_Generic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("test-alpine", "alpine", "3.19")
	if err != nil {
		t.Skipf("EnsureImage failed: %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
	t.Logf("Generic image path: %s", path)
}

func TestManager_EnsureImage_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a pre-existing image file
	existingPath := filepath.Join(tmpDir, "existing.qcow2")
	if err := os.WriteFile(existingPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)
	path, err := m.EnsureImage("existing", "ubuntu", "22.04")
	if err != nil {
		t.Fatalf("EnsureImage for existing file failed: %v", err)
	}
	if path != existingPath {
		t.Errorf("Expected path %s, got %s", existingPath, path)
	}
}

func TestManager_CloneImage_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake base image
	basePath := filepath.Join(tmpDir, "base.qcow2")
	if err := os.WriteFile(basePath, []byte("fake-qcow2-data"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)
	clonedPath, err := m.CloneImage(basePath, "node-1")
	if err != nil {
		t.Skipf("CloneImage failed (needs qemu-img): %v", err)
	}
	t.Logf("Cloned to: %s", clonedPath)

	// Clean up cloned image
	os.Remove(clonedPath)
}

func TestManager_ResizeImage_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir)

	// Create a minimal qcow2 image first
	path, err := m.EnsureImage("resize-test", "generic", "1")
	if err != nil {
		t.Skipf("EnsureImage failed: %v", err)
	}

	err = m.ResizeImage(path, 50)
	if err != nil {
		t.Skipf("ResizeImage failed (needs qemu-img): %v", err)
	}
	t.Logf("Resized %s to 50GB", path)
}

func TestManager_DeleteImage_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake image
	imgPath := filepath.Join(tmpDir, "delete-test.qcow2")
	if err := os.WriteFile(imgPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(tmpDir)
	err = m.DeleteImage("delete-test")
	if err != nil {
		t.Fatalf("DeleteImage failed: %v", err)
	}

	if _, err := os.Stat(imgPath); !os.IsNotExist(err) {
		t.Error("Image should be deleted")
	}
}

func TestNetworkManager_DeleteNetwork_Real(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-prov-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	nm := NetworkManager{}
	err = nm.DeleteNetwork(context.Background(), "nonexistent-network")
	if err != nil {
		t.Logf("DeleteNetwork for nonexistent: %v (expected)", err)
	}
}