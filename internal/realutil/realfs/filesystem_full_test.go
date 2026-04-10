package realfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realfs"
)

// TestRealFilesystem_WriteFile_CreatesDir tests WriteFile creating parent dirs
func TestRealFilesystem_WriteFile_CreatesDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "subdir", "nested", "test.txt")

	content := []byte("test content")
	err := fs.WriteFile(targetPath, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_WriteFile_Overwrite tests WriteFile overwriting existing file
func TestRealFilesystem_WriteFile_Overwrite(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	err := fs.WriteFile(targetPath, []byte("initial"), 0644)
	if err != nil {
		t.Fatalf("first WriteFile failed: %v", err)
	}

	// Overwrite
	err = fs.WriteFile(targetPath, []byte("overwritten"), 0644)
	if err != nil {
		t.Fatalf("second WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("file read failed: %v", err)
	}
	if string(data) != "overwritten" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_ReadDir_Empty tests ReadDir on empty directory
func TestRealFilesystem_ReadDir_Empty(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir, got %d entries", len(entries))
	}
}

// TestRealFilesystem_ReadDir_WithFiles tests ReadDir with multiple files
func TestRealFilesystem_ReadDir_WithFiles(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Create test files
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries, got %d", len(entries))
	}
}

// TestRealFilesystem_Copy_LargeFile tests Copy with larger file
func TestRealFilesystem_Copy_LargeFile(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "large.bin")
	dstPath := filepath.Join(tmpDir, "copy.bin")

	// Create 1MB file
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(srcPath, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	err := fs.Copy(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify copy
	dstData, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("copy read failed: %v", err)
	}
	if len(dstData) != len(largeContent) {
		t.Errorf("wrong size: got %d, want %d", len(dstData), len(largeContent))
	}
}

// TestRealFilesystem_Append_NewFile tests Append to new file
func TestRealFilesystem_Append_NewFile(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "append.txt")

	err := fs.Append(path, []byte("first line\n"))
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "first line\n" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_Append_Existing tests Append to existing file
func TestRealFilesystem_Append_Existing(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "append.txt")

	// First append
	if err := fs.Append(path, []byte("first line\n")); err != nil {
		t.Fatal(err)
	}

	// Second append
	err := fs.Append(path, []byte("second line\n"))
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	expected := "first line\nsecond line\n"
	if string(data) != expected {
		t.Errorf("wrong content: got %s, want %s", string(data), expected)
	}
}

// TestRealFilesystem_Lock_TryLock tests Lock and TryLock interaction
func TestRealFilesystem_Lock_TryLock(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	// Acquire lock
	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Release lock
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Try again - should succeed
	lock2, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock should succeed after unlock: %v", err)
	}
	lock2.Unlock()
}

// TestRealFilesystem_Lock_Write tests Lock Write method
func TestRealFilesystem_Lock_Write(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	defer lock.Unlock()

	// Write to lock file
	n, err := lock.Write([]byte("lock data"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 9 {
		t.Errorf("wrong bytes written: got %d, want 9", n)
	}

	// Read back
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "lock data" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_Lock_Read tests Lock Read method
func TestRealFilesystem_Lock_Read(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	// Write data first
	if err := os.WriteFile(lockPath, []byte("test data for read"), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	defer lock.Unlock()

	// Read from lock file
	buf := make([]byte, 100)
	n, err := lock.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != "test data for read" {
		t.Errorf("wrong content: got %s", string(buf[:n]))
	}
}

// TestRealFilesystem_TryLock_Basic tests basic TryLock functionality
func TestRealFilesystem_TryLock_Basic(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	// TryLock on non-existent file should work
	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// Unlock
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// Should be able to lock again
	lock2, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock should work after unlock: %v", err)
	}
	lock2.Unlock()
}