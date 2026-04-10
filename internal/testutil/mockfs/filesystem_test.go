// Package mockfs_test tests the mock filesystem
package mockfs_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/testutil/mockfs"
)

// TestMockFilesystem_Create tests filesystem creation
func TestMockFilesystem_Create(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	if fs == nil {
		t.Fatal("filesystem should not be nil")
	}
}

// TestMockFilesystem_MkdirAll tests directory creation
func TestMockFilesystem_MkdirAll(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	err := fs.MkdirAll("/var/lib/test", 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if !fs.Exists("/var/lib/test") {
		t.Error("directory should exist")
	}
}

// TestMockFilesystem_WriteRead tests file write and read
func TestMockFilesystem_WriteRead(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	// Create parent directory
	_ = fs.MkdirAll("/var/lib/test", 0755)

	// Write file
	content := []byte("test content")
	err := fs.WriteFile("/var/lib/test/file.txt", content, 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Read file
	read, err := fs.ReadFile("/var/lib/test/file.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("expected %s, got %s", content, read)
	}
}

// TestMockFilesystem_Stat tests file stat
func TestMockFilesystem_Stat(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	_ = fs.MkdirAll("/var/lib/test", 0755)
	_ = fs.WriteFile("/var/lib/test/file.txt", []byte("test"), 0644)

	info, err := fs.Stat("/var/lib/test/file.txt")
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Name() != "file.txt" {
		t.Errorf("expected file.txt, got %s", info.Name())
	}
	if info.IsDir() {
		t.Error("file should not be a directory")
	}
}

// TestMockFilesystem_Remove tests file removal
func TestMockFilesystem_Remove(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	_ = fs.MkdirAll("/var/lib/test", 0755)
	_ = fs.WriteFile("/var/lib/test/file.txt", []byte("test"), 0644)

	err := fs.Remove("/var/lib/test/file.txt")
	if err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	if fs.Exists("/var/lib/test/file.txt") {
		t.Error("file should not exist")
	}
}

// TestMockFilesystem_ReadDir tests directory listing
func TestMockFilesystem_ReadDir(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	_ = fs.MkdirAll("/var/lib/test", 0755)
	_ = fs.WriteFile("/var/lib/test/file1.txt", []byte("test"), 0644)
	_ = fs.WriteFile("/var/lib/test/file2.txt", []byte("test"), 0644)

	infos, err := fs.ReadDir("/var/lib/test")
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("expected 2 files, got %d", len(infos))
	}
}

// TestMockFilesystem_ErrorMode tests error mode
func TestMockFilesystem_ErrorMode(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	// Enable error mode
	fs.SetErrorMode(true)

	// Operations should fail
	err := fs.MkdirAll("/test", 0755)
	if err == nil {
		t.Error("expected error in error mode")
	}

	// Disable error mode
	fs.SetErrorMode(false)

	// Operations should succeed
	err = fs.MkdirAll("/test", 0755)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockFilesystem_NestedDirectories tests nested directory creation
func TestMockFilesystem_NestedDirectories(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	err := fs.MkdirAll("/a/b/c/d/e", 0755)
	if err != nil {
		t.Fatalf("failed to create nested directories: %v", err)
	}

	if !fs.Exists("/a/b/c/d/e") {
		t.Error("nested directory should exist")
	}
}

// TestMockFilesystem_MultipleFiles tests multiple file operations
func TestMockFilesystem_MultipleFiles(t *testing.T) {
	fs := mockfs.NewMockFilesystem()

	_ = fs.MkdirAll("/test", 0755)

	for i := 0; i < 10; i++ {
		content := []byte("test")
		err := fs.WriteFile("/test/file"+string(rune('0'+i))+".txt", content, 0644)
		if err != nil {
			t.Errorf("failed to write file %d: %v", i, err)
		}
	}

	infos, err := fs.ReadDir("/test")
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(infos) != 10 {
		t.Errorf("expected 10 files, got %d", len(infos))
	}
}