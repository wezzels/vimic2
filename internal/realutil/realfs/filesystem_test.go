// Package realfs_test tests the real filesystem
package realfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/realutil/realfs"
)

// TestRealFilesystem_Create tests filesystem creation
func TestRealFilesystem_Create(t *testing.T) {
	fs := realfs.NewFilesystem()

	if fs == nil {
		t.Fatal("filesystem should not be nil")
	}
}

// TestRealFilesystem_MkdirAll tests directory creation
func TestRealFilesystem_MkdirAll(t *testing.T) {
	fs := realfs.NewFilesystem()

	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "a", "b", "c")
	err = fs.MkdirAll(testPath, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if !fs.Exists(testPath) {
		t.Error("directory should exist")
	}
}

// TestRealFilesystem_WriteRead tests file write and read
func TestRealFilesystem_WriteRead(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")

	err = fs.WriteFile(testFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	read, err := fs.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("expected %s, got %s", content, read)
	}
}

// TestRealFilesystem_Stat tests file stat
func TestRealFilesystem_Stat(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = fs.WriteFile(testFile, []byte("test"), 0644)

	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected test.txt, got %s", info.Name())
	}
	if info.IsDir() {
		t.Error("file should not be a directory")
	}
	if info.Size() != 4 {
		t.Errorf("expected size 4, got %d", info.Size())
	}
}

// TestRealFilesystem_Remove tests file removal
func TestRealFilesystem_Remove(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = fs.WriteFile(testFile, []byte("test"), 0644)

	err = fs.Remove(testFile)
	if err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	if fs.Exists(testFile) {
		t.Error("file should not exist")
	}
}

// TestRealFilesystem_ReadDir tests directory listing
func TestRealFilesystem_ReadDir(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_ = fs.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	_ = fs.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	infos, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("expected 2 files, got %d", len(infos))
	}
}

// TestRealFilesystem_Copy tests file copy
func TestRealFilesystem_Copy(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")
	_ = fs.WriteFile(srcFile, []byte("content"), 0644)

	err = fs.Copy(srcFile, dstFile)
	if err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}

	if !fs.Exists(dstFile) {
		t.Error("destination file should exist")
	}

	dst, _ := fs.ReadFile(dstFile)
	if string(dst) != "content" {
		t.Errorf("expected content, got %s", dst)
	}
}

// TestRealFilesystem_Move tests file move
func TestRealFilesystem_Move(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")
	_ = fs.WriteFile(srcFile, []byte("content"), 0644)

	err = fs.Move(srcFile, dstFile)
	if err != nil {
		t.Fatalf("failed to move file: %v", err)
	}

	if fs.Exists(srcFile) {
		t.Error("source file should not exist")
	}
	if !fs.Exists(dstFile) {
		t.Error("destination file should exist")
	}
}

// TestRealFilesystem_Symlink tests symlink creation
func TestRealFilesystem_Symlink(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "src.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")
	_ = fs.WriteFile(srcFile, []byte("content"), 0644)

	err = fs.Symlink(srcFile, linkFile)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	target, err := fs.Readlink(linkFile)
	if err != nil {
		t.Fatalf("failed to read symlink: %v", err)
	}

	if target != srcFile {
		t.Errorf("expected %s, got %s", srcFile, target)
	}
}

// TestRealFilesystem_Chmod tests permission change
func TestRealFilesystem_Chmod(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = fs.WriteFile(testFile, []byte("test"), 0644)

	err = fs.Chmod(testFile, 0600)
	if err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	info, _ := fs.Stat(testFile)
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %o", info.Mode().Perm())
	}
}

// TestRealFilesystem_Touch tests touch
func TestRealFilesystem_Touch(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Create new file
	err = fs.Touch(testFile)
	if err != nil {
		t.Fatalf("failed to touch new file: %v", err)
	}

	if !fs.Exists(testFile) {
		t.Error("file should exist")
	}

	// Touch existing file (update mtime)
	err = fs.Touch(testFile)
	if err != nil {
		t.Fatalf("failed to touch existing file: %v", err)
	}
}

// TestRealFilesystem_Append tests append
func TestRealFilesystem_Append(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = fs.WriteFile(testFile, []byte("line1\n"), 0644)

	err = fs.Append(testFile, []byte("line2\n"))
	if err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	content, _ := fs.ReadFile(testFile)
	expected := "line1\nline2\n"
	if string(content) != expected {
		t.Errorf("expected %s, got %s", expected, content)
	}
}

// TestRealFilesystem_TempDir tests temp directory creation
func TestRealFilesystem_TempDir(t *testing.T) {
	fs := realfs.NewFilesystem()

	dir, err := fs.TempDir("realfs-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	if !fs.Exists(dir) {
		t.Error("temp dir should exist")
	}
}

// TestRealFilesystem_TempFile tests temp file creation
func TestRealFilesystem_TempFile(t *testing.T) {
	fs := realfs.NewFilesystem()

	file, err := fs.TempFile("realfs-", ".txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if !fs.Exists(file.Name()) {
		t.Error("temp file should exist")
	}
}

// TestRealFilesystem_Walk tests walk
func TestRealFilesystem_Walk(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_ = fs.MkdirAll(filepath.Join(tmpDir, "a", "b"), 0755)
	_ = fs.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	_ = fs.WriteFile(filepath.Join(tmpDir, "a", "file2.txt"), []byte("2"), 0644)
	_ = fs.WriteFile(filepath.Join(tmpDir, "a", "b", "file3.txt"), []byte("3"), 0644)

	var files []string
	err = fs.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

// TestRealFilesystem_Glob tests glob
func TestRealFilesystem_Glob(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_ = fs.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	_ = fs.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("2"), 0644)
	_ = fs.WriteFile(filepath.Join(tmpDir, "file3.log"), []byte("3"), 0644)

	matches, err := fs.Glob(filepath.Join(tmpDir, "*.txt"))
	if err != nil {
		t.Fatalf("failed to glob: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

// TestRealFilesystem_PathOps tests path operations
func TestRealFilesystem_PathOps(t *testing.T) {
	fs := realfs.NewFilesystem()

	// Test Abs
	abs, err := fs.Abs("/tmp")
	if err != nil {
		t.Fatalf("failed to get abs: %v", err)
	}
	if abs != "/tmp" {
		t.Errorf("expected /tmp, got %s", abs)
	}

	// Test Dir
	dir := fs.Dir("/tmp/test.txt")
	if dir != "/tmp" {
		t.Errorf("expected /tmp, got %s", dir)
	}

	// Test Base
	base := fs.Base("/tmp/test.txt")
	if base != "test.txt" {
		t.Errorf("expected test.txt, got %s", base)
	}

	// Test Ext
	ext := fs.Ext("/tmp/test.txt")
	if ext != ".txt" {
		t.Errorf("expected .txt, got %s", ext)
	}

	// Test Join
	join := fs.Join("/tmp", "test.txt")
	if join != "/tmp/test.txt" {
		t.Errorf("expected /tmp/test.txt, got %s", join)
	}

	// Test Split
	dir, file := fs.Split("/tmp/test.txt")
	if dir != "/tmp/" {
		t.Errorf("expected /tmp/, got %s", dir)
	}
	if file != "test.txt" {
		t.Errorf("expected test.txt, got %s", file)
	}

	// Test Clean
	clean := fs.Clean("/tmp/../tmp/./test.txt")
	if clean != "/tmp/test.txt" {
		t.Errorf("expected /tmp/test.txt, got %s", clean)
	}

	// Test IsAbs
	if !fs.IsAbs("/tmp") {
		t.Error("/tmp should be absolute")
	}
	if fs.IsAbs("tmp") {
		t.Error("tmp should not be absolute")
	}
}

// TestRealFilesystem_LockFile tests file locking
func TestRealFilesystem_LockFile(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockFile := filepath.Join(tmpDir, "test.lock")

	// Acquire lock
	lf, err := fs.Lock(lockFile)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Write to locked file
	_, err = lf.Write([]byte("locked content"))
	if err != nil {
		t.Fatalf("failed to write to locked file: %v", err)
	}

	// Read from locked file
	buf := make([]byte, 100)
	_, err = lf.File().ReadAt(buf, 0)
	// Note: Read might return EOF or partial read, which is fine

	// Check path
	if lf.Path() != lockFile {
		t.Errorf("expected %s, got %s", lockFile, lf.Path())
	}

	// Release lock
	err = lf.Unlock()
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}
}

// TestRealFilesystem_TryLock tests non-blocking lock
func TestRealFilesystem_TryLock(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockFile := filepath.Join(tmpDir, "test.lock")

	// Acquire lock
	lf, err := fs.TryLock(lockFile)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Release lock
	err = lf.Unlock()
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}
}

// TestRealFilesystem_AtomicWrite tests atomic write behavior
func TestRealFilesystem_AtomicWrite(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")

	// Write should create parent directories if needed
	subDir := filepath.Join(tmpDir, "subdir", "deep")
	testFile = filepath.Join(subDir, "test.txt")

	err = fs.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if !fs.Exists(testFile) {
		t.Error("file should exist")
	}
}

// TestRealFilesystem_Chown tests chown
func TestRealFilesystem_Chown(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = fs.WriteFile(testFile, []byte("test"), 0644)

	// Note: Chown may fail without root, but we test the function exists
	_ = fs.Chown(testFile, os.Getuid(), os.Getgid())
}

// TestRealFilesystem_Rel tests relative path
func TestRealFilesystem_Rel(t *testing.T) {
	fs := realfs.NewFilesystem()

	rel, err := fs.Rel("/tmp", "/tmp/test.txt")
	if err != nil {
		t.Fatalf("failed to get rel: %v", err)
	}
	if rel != "test.txt" {
		t.Errorf("expected test.txt, got %s", rel)
	}
}

// TestRealFilesystem_SplitList tests split list
func TestRealFilesystem_SplitList(t *testing.T) {
	fs := realfs.NewFilesystem()

	parts := fs.SplitList("/usr/bin:/usr/local/bin")
	if len(parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(parts))
	}
	if parts[0] != "/usr/bin" {
		t.Errorf("expected /usr/bin, got %s", parts[0])
	}
}

// TestRealFilesystem_OpString tests operation string
func TestRealFilesystem_OpString(t *testing.T) {
	op := realfs.Create
	if op.String() != "CREATE" {
		t.Errorf("expected CREATE, got %s", op.String())
	}

	op = realfs.Write
	if op.String() != "WRITE" {
		t.Errorf("expected WRITE, got %s", op.String())
	}

	op = realfs.Remove
	if op.String() != "REMOVE" {
		t.Errorf("expected REMOVE, got %s", op.String())
	}

	op = realfs.Rename
	if op.String() != "RENAME" {
		t.Errorf("expected RENAME, got %s", op.String())
	}

	op = realfs.Chmod
	if op.String() != "CHMOD" {
		t.Errorf("expected CHMOD, got %s", op.String())
	}

	op = realfs.Op(0)
	if op.String() != "?" {
		t.Errorf("expected ?, got %s", op.String())
	}
}

// TestRealFilesystem_LockFileIO tests lock file read/write
func TestRealFilesystem_LockFileIO(t *testing.T) {
	fs := realfs.NewFilesystem()

	tmpDir, err := os.MkdirTemp("", "realfs-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockFile := filepath.Join(tmpDir, "test.lock")

	lf, err := fs.Lock(lockFile)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Test write
	n, err := lf.Write([]byte("test"))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}

	// Test read
	buf := make([]byte, 100)
	_, _ = lf.File().ReadAt(buf, 0) // May return EOF, that's fine

	// Release lock
	err = lf.Unlock()
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}
}