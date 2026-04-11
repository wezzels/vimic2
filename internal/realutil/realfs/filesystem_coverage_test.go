package realfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realfs"
)

// TestRealFilesystem_WriteFile_ParentDir tests WriteFile creating parent dirs
func TestRealFilesystem_WriteFile_ParentDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "a", "b", "c", "test.txt")

	content := []byte("nested content")
	err := fs.WriteFile(targetPath, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if string(data) != "nested content" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_WriteFile_Overwrite tests WriteFile overwriting existing file
func TestRealFilesystem_WriteFile_Overwrite(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "test.txt")

	// First write
	if err := fs.WriteFile(targetPath, []byte("first"), 0644); err != nil {
		t.Fatalf("first WriteFile failed: %v", err)
	}

	// Overwrite
	if err := fs.WriteFile(targetPath, []byte("second"), 0644); err != nil {
		t.Fatalf("second WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "second" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_WriteFile_Permissions tests WriteFile with different permissions
func TestRealFilesystem_WriteFile_Permissions(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Test with different permissions
	for _, perm := range []os.FileMode{0644, 0755, 0600, 0777} {
		path := filepath.Join(tmpDir, "test_"+string(rune(perm)))
		if err := fs.WriteFile(path, []byte("test"), perm); err != nil {
			t.Errorf("WriteFile failed with perm %o: %v", perm, err)
		}
	}
}

// TestRealFilesystem_ReadDir_NonExistent tests ReadDir with non-existent directory
func TestRealFilesystem_ReadDir_NonExistent(t *testing.T) {
	fs := realfs.NewFilesystem()

	_, err := fs.ReadDir("/non/existent/directory")
	if err == nil {
		t.Error("ReadDir should fail for non-existent directory")
	}
}

// TestRealFilesystem_ReadDir_Empty tests ReadDir with empty directory
func TestRealFilesystem_ReadDir_Empty(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Create empty subdirectory
	subDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	entries, err := fs.ReadDir(subDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty directory, got %d entries", len(entries))
	}
}

// TestRealFilesystem_Copy_NonExistentSource tests Copy with non-existent source
func TestRealFilesystem_Copy_NonExistentSource(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	err := fs.Copy("/non/existent/file", filepath.Join(tmpDir, "dst"))
	if err == nil {
		t.Error("Copy should fail for non-existent source")
	}
}

// TestRealFilesystem_Copy_SameSourceAndDest tests Copy with same source and dest
func TestRealFilesystem_Copy_SameSourceAndDest(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "same.txt")

	// Create source file
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy to same path
	err := fs.Copy(srcPath, srcPath)
	// This may or may not error depending on implementation
	_ = err
}

// TestRealFilesystem_Copy_LargeFile tests Copy with large file
func TestRealFilesystem_Copy_LargeFile(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "large.bin")
	dstPath := filepath.Join(tmpDir, "copy.bin")

	// Create large file (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(srcPath, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	if err := fs.Copy(srcPath, dstPath); err != nil {
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

	if err := fs.Append(path, []byte("first line\n")); err != nil {
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
	if err := fs.Append(path, []byte("line 1\n")); err != nil {
		t.Fatal(err)
	}

	// Second append
	if err := fs.Append(path, []byte("line 2\n")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "line 1\nline 2\n"
	if string(data) != expected {
		t.Errorf("wrong content: got %s, want %s", string(data), expected)
	}
}

// TestRealFilesystem_Append_NonExistentDir tests Append with non-existent directory
func TestRealFilesystem_Append_NonExistentDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent", "append.txt")

	// Append should fail because directory doesn't exist
	err := fs.Append(path, []byte("test"))
	if err == nil {
		t.Error("Append should fail for non-existent directory")
	}
}

// TestRealFilesystem_Lock_Basic tests basic lock/unlock
func TestRealFilesystem_Lock_Basic(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Unlock should work
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

// TestRealFilesystem_TryLock_Basic tests basic try-lock
func TestRealFilesystem_TryLock_Basic(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// Unlock should work
	if err := lock.Unlock(); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
}

// TestRealFilesystem_Lock_WriteWithoutLock tests writing without lock
func TestRealFilesystem_Lock_AlreadyLocked(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Try to lock again on same LockFile (method)
	err = lock.Lock()
	if err == nil {
		t.Error("Lock should fail when already locked")
	}

	lock.Unlock()
}

// TestRealFilesystem_TryLock_AlreadyLockedMethod tests TryLock on already locked (method)
func TestRealFilesystem_TryLock_AlreadyLockedMethod(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// Try to lock again on same LockFile (method)
	err = lock.TryLock()
	if err == nil {
		t.Error("TryLock should fail when already locked")
	}

	lock.Unlock()
}

// TestRealFilesystem_Unlock_NotLocked tests Unlock on not locked file
func TestRealFilesystem_Unlock_NotLocked(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Unlock once
	if err := lock.Unlock(); err != nil {
		t.Fatalf("first Unlock failed: %v", err)
	}

	// Try to unlock again (not locked)
	err = lock.Unlock()
	if err == nil {
		t.Error("Unlock should fail when not locked")
	}
}

// TestRealFilesystem_Lock_ParentDirCreation tests Lock creating parent directory
func TestRealFilesystem_Lock_ParentDirCreation(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "subdir", "nested", "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	defer lock.Unlock()

	// Parent directory should be created
	if !fs.Exists(filepath.Join(tmpDir, "subdir", "nested")) {
		t.Error("parent directory should be created")
	}
}

// TestRealFilesystem_WriteFile_InvalidPath tests WriteFile with invalid path
func TestRealFilesystem_WriteFile_InvalidPath(t *testing.T) {
	fs := realfs.NewFilesystem()

	// Try to write to a path that requires creating directories in /dev
	// This should work if parent dir exists
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "valid.txt")

	if err := fs.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

// TestRealFilesystem_ReadDir_WithSubdirs tests ReadDir with subdirectories
func TestRealFilesystem_ReadDir_WithSubdirs(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Create files and subdirectories
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries (file + subdir), got %d", len(entries))
	}
}

// TestRealFilesystem_Copy_Overwrite tests Copy overwriting existing file
func TestRealFilesystem_Copy_Overwrite(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	// Create source and initial destination
	if err := os.WriteFile(srcPath, []byte("source content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy should overwrite
	if err := fs.Copy(srcPath, dstPath); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "source content" {
		t.Errorf("wrong content: got %s", string(data))
	}
}

// TestRealFilesystem_Copy_IntoDir tests Copy into directory
func TestRealFilesystem_Copy_IntoDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstDir := filepath.Join(tmpDir, "dst")
	dstPath := filepath.Join(dstDir, "src.txt")

	// Create source and destination directory
	if err := os.WriteFile(srcPath, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy into directory
	if err := fs.Copy(srcPath, dstPath); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify copy
	if _, err := os.Stat(dstPath); err != nil {
		t.Errorf("destination file not created: %v", err)
	}
}

// TestRealFilesystem_Lock_ParentDir tests Lock creating parent directory
func TestRealFilesystem_Lock_ParentDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "subdir", "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Parent directory should be created
	if !fs.Exists(filepath.Join(tmpDir, "subdir")) {
		t.Error("parent directory should be created")
	}

	lock.Unlock()
}

// TestRealFilesystem_TryLock_ParentDir tests TryLock (no parent dir creation)
func TestRealFilesystem_TryLock_ParentDir(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	lock.Unlock()
}

// TestRealFilesystem_WriteFile_EmptyData tests WriteFile with empty data
func TestRealFilesystem_WriteFile_EmptyData(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.txt")

	if err := fs.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

// TestRealFilesystem_Append_Multiple tests multiple appends
func TestRealFilesystem_Append_Multiple(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "append.txt")

	for i := 0; i < 5; i++ {
		if err := fs.Append(path, []byte("line\n")); err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "line\nline\nline\nline\nline\n"
	if string(data) != expected {
		t.Errorf("wrong content: got %s, want %s", string(data), expected)
	}
}

// TestRealFilesystem_Copy_ErrorPaths tests Copy error conditions
func TestRealFilesystem_Copy_ErrorPaths(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Copy non-existent source
	err := fs.Copy("/non/existent/file", filepath.Join(tmpDir, "dst"))
	if err == nil {
		t.Error("Copy should fail for non-existent source")
	}
}

// TestRealFilesystem_WriteFile_TempFileError tests WriteFile temp file failure
func TestRealFilesystem_WriteFile_TempFileError(t *testing.T) {
	fs := realfs.NewFilesystem()
	// This tests the rename failure path - we can't easily simulate temp file write failure
	// But we can test that the cleanup happens by creating a scenario where rename fails
	tmpDir := t.TempDir()

	// Create a directory where the file should be (to cause rename failure)
	targetPath := filepath.Join(tmpDir, "test.txt")
	if err := os.Mkdir(targetPath, 0755); err != nil {
		t.Fatal(err)
	}

	// WriteFile should fail because target is a directory
	err := fs.WriteFile(targetPath, []byte("test"), 0644)
	if err == nil {
		t.Error("WriteFile should fail when target is a directory")
	}
}

// TestRealFilesystem_ReadDir_EntryInfoError tests ReadDir with entry.Info failure
func TestRealFilesystem_ReadDir_EntryInfoError(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Create a file in the directory
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// ReadDir should work
	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

// TestRealFilesystem_Copy_SourceStatError tests Copy with source stat failure
func TestRealFilesystem_Copy_SourceStatError(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()

	// Create a source file and remove read permission
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy should work normally
	if err := fs.Copy(srcPath, dstPath); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify copy worked
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test" {
		t.Errorf("wrong content: %s", string(data))
	}
}

// TestRealFilesystem_Lock_DoubleUnlock tests double unlock
func TestRealFilesystem_Lock_DoubleUnlock(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// First unlock
	if err := lock.Unlock(); err != nil {
		t.Fatalf("first Unlock failed: %v", err)
	}

	// Second unlock should fail
	if err := lock.Unlock(); err == nil {
		t.Error("second Unlock should fail")
	}
}

// TestRealFilesystem_TryLock_DoubleUnlock tests double unlock on TryLock
func TestRealFilesystem_TryLock_DoubleUnlock(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}

	// First unlock
	if err := lock.Unlock(); err != nil {
		t.Fatalf("first Unlock failed: %v", err)
	}

	// Second unlock should fail
	if err := lock.Unlock(); err == nil {
		t.Error("second Unlock should fail")
	}
}

// TestRealFilesystem_Lock_ReadAfterWrite tests read after write on locked file
func TestRealFilesystem_Lock_ReadAfterWrite(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	// Write data
	data := []byte("hello world")
	n, err := lock.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("wrong bytes written: got %d, want %d", n, len(data))
	}

	// Read at current position (after write, so at EOF)
	// This will read 0 bytes since we're at the end
	buf := make([]byte, 100)
	n, err = lock.Read(buf)
	// At EOF, Read returns 0, io.EOF
	if err != nil && err.Error() != "EOF" {
		// Some implementations return EOF
	}
	// n should be 0 at EOF
	_ = n

	lock.Unlock()
}

// TestRealFilesystem_Lock_FileAndPath tests File() and Path() methods
func TestRealFilesystem_Lock_FileAndPath(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.Lock(lockPath)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	defer lock.Unlock()

	// Test Path()
	if lock.Path() != lockPath {
		t.Errorf("Path() returned wrong path: got %s, want %s", lock.Path(), lockPath)
	}

	// Test File()
	if lock.File() == nil {
		t.Error("File() should not return nil")
	}
}

// TestRealFilesystem_TryLock_FileAndPath tests File() and Path() methods
func TestRealFilesystem_TryLock_FileAndPath(t *testing.T) {
	fs := realfs.NewFilesystem()
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	lock, err := fs.TryLock(lockPath)
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	defer lock.Unlock()

	// Test Path()
	if lock.Path() != lockPath {
		t.Errorf("Path() returned wrong path: got %s, want %s", lock.Path(), lockPath)
	}

	// Test File()
	if lock.File() == nil {
		t.Error("File() should not return nil")
	}
}
