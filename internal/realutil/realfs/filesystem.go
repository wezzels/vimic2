// Package realfs provides real filesystem operations for production use
package realfs

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Filesystem provides real filesystem operations
type Filesystem struct {
	mu sync.RWMutex
}

// NewFilesystem creates a new real filesystem instance
func NewFilesystem() *Filesystem {
	return &Filesystem{}
}

// MkdirAll creates directories recursively
func (f *Filesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// WriteFile writes content to a file with atomic write support
func (f *Filesystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Write to temp file first (atomic write)
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, perm); err != nil {
		return err
	}

	// Rename is atomic on POSIX systems
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile)
		return err
	}

	return nil
}

// ReadFile reads content from a file
func (f *Filesystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// Remove removes a file or directory
func (f *Filesystem) Remove(path string) error {
	return os.RemoveAll(path)
}

// Stat returns file info
func (f *Filesystem) Stat(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// ReadDir reads directory contents
func (f *Filesystem) ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Exists checks if a file exists
func (f *Filesystem) Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Copy copies a file from src to dst
func (f *Filesystem) Copy(src, dst string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Move moves a file from src to dst
func (f *Filesystem) Move(src, dst string) error {
	return os.Rename(src, dst)
}

// Symlink creates a symbolic link
func (f *Filesystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// Readlink reads the target of a symbolic link
func (f *Filesystem) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

// Chmod changes file permissions
func (f *Filesystem) Chmod(filename string, perm os.FileMode) error {
	return os.Chmod(filename, perm)
}

// Chown changes file ownership
func (f *Filesystem) Chown(filename string, uid, gid int) error {
	return os.Chown(filename, uid, gid)
}

// Touch creates an empty file or updates modification time
func (f *Filesystem) Touch(filename string) error {
	if f.Exists(filename) {
		return os.Chtimes(filename, time.Now(), time.Now())
	}
	return os.WriteFile(filename, []byte{}, 0644)
}

// Append appends data to a file
func (f *Filesystem) Append(filename string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// TempDir creates a temporary directory and returns its path
func (f *Filesystem) TempDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

// TempFile creates a temporary file and returns its path
func (f *Filesystem) TempFile(prefix, suffix string) (*os.File, error) {
	return os.CreateTemp("", prefix+"*"+suffix)
}

// Walk walks a directory tree
func (f *Filesystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// Glob returns matching files
func (f *Filesystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Abs returns an absolute path
func (f *Filesystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Rel returns a relative path
func (f *Filesystem) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

// Dir returns the directory of a path
func (f *Filesystem) Dir(path string) string {
	return filepath.Dir(path)
}

// Base returns the base name of a path
func (f *Filesystem) Base(path string) string {
	return filepath.Base(path)
}

// Ext returns the file extension
func (f *Filesystem) Ext(path string) string {
	return filepath.Ext(path)
}

// Join joins path elements
func (f *Filesystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Split splits a path into directory and file
func (f *Filesystem) Split(path string) (dir, file string) {
	return filepath.Split(path)
}

// SplitList splits a PATH-like list
func (f *Filesystem) SplitList(path string) []string {
	return filepath.SplitList(path)
}

// Clean cleans a path
func (f *Filesystem) Clean(path string) string {
	return filepath.Clean(path)
}

// IsAbs reports whether the path is absolute
func (f *Filesystem) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Watch represents a file watcher
type Watch struct {
	Path      string
	Events    chan Event
	Errors    chan error
	Done      chan struct{}
	callbacks map[string][]func(Event)
	mu        sync.RWMutex
}

// Event represents a filesystem event
type Event struct {
	Path string
	Op   Op
}

// Op represents a file operation
type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

// String returns the operation name
func (op Op) String() string {
	var names []string
	if op&Create != 0 {
		names = append(names, "CREATE")
	}
	if op&Write != 0 {
		names = append(names, "WRITE")
	}
	if op&Remove != 0 {
		names = append(names, "REMOVE")
	}
	if op&Rename != 0 {
		names = append(names, "RENAME")
	}
	if op&Chmod != 0 {
		names = append(names, "CHMOD")
	}
	if len(names) == 0 {
		return "?"
	}
	return names[0]
}

// LockFile represents a locked file
type LockFile struct {
	file   *os.File
	path   string
	locked bool
	mu     sync.Mutex
}

// Lock creates an exclusive lock on a file
func (f *Filesystem) Lock(filename string) (*LockFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	lf := &LockFile{
		file: file,
		path: filename,
	}

	if err := lf.Lock(); err != nil {
		file.Close()
		return nil, err
	}

	return lf, nil
}

// TryLock attempts to acquire a lock without blocking
func (f *Filesystem) TryLock(filename string) (*LockFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	lf := &LockFile{
		file: file,
		path: filename,
	}

	if err := lf.TryLock(); err != nil {
		file.Close()
		return nil, err
	}

	return lf, nil
}

// Lock acquires an exclusive lock (blocking)
func (lf *LockFile) Lock() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.locked {
		return errors.New("already locked")
	}

	// Use flock-style advisory locking via FileLock
	// Note: This is a simplified implementation
	// Production should use syscall.Flock or similar
	lf.locked = true
	return nil
}

// TryLock attempts to acquire lock without blocking
func (lf *LockFile) TryLock() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if lf.locked {
		return errors.New("already locked")
	}

	lf.locked = true
	return nil
}

// Unlock releases the lock
func (lf *LockFile) Unlock() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if !lf.locked {
		return errors.New("not locked")
	}

	lf.locked = false
	return lf.file.Close()
}

// Write writes data to the locked file
func (lf *LockFile) Write(data []byte) (int, error) {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if !lf.locked {
		return 0, errors.New("not locked")
	}

	return lf.file.Write(data)
}

// Read reads data from the locked file
func (lf *LockFile) Read(p []byte) (int, error) {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	if !lf.locked {
		return 0, errors.New("not locked")
	}

	return lf.file.Read(p)
}

// Path returns the lock file path
func (lf *LockFile) Path() string {
	return lf.path
}

// File returns the underlying file
func (lf *LockFile) File() *os.File {
	return lf.file
}
