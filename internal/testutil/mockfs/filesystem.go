// Package mockfs provides mock filesystem for testing
package mockfs

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

// MockFile represents a mock file
type MockFile struct {
	Name    string
	Content []byte
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
	Children map[string]*MockFile
}

// MockFilesystem provides a mock filesystem for testing
type MockFilesystem struct {
	root    *MockFile
	mu      sync.RWMutex
	errMode bool // If true, return errors
}

// NewMockFilesystem creates a new mock filesystem
func NewMockFilesystem() *MockFilesystem {
	return &MockFilesystem{
		root: &MockFile{
			Name:     "/",
			IsDir:    true,
			Children: make(map[string]*MockFile),
			ModTime:  time.Now(),
		},
	}
}

// MkdirAll creates directories recursively
func (m *MockFilesystem) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode {
		return errors.New("mock error: mkdir")
	}

	parts := splitPath(path)
	current := m.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			if !child.IsDir {
				return errors.New("not a directory")
			}
			current = child
		} else {
			newDir := &MockFile{
				Name:     part,
				IsDir:    true,
				Mode:     perm,
				ModTime:  time.Now(),
				Children: make(map[string]*MockFile),
			}
			current.Children[part] = newDir
			current = newDir
		}
	}

	return nil
}

// WriteFile writes content to a file
func (m *MockFilesystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode {
		return errors.New("mock error: write")
	}

	parts := splitPath(filename)
	if len(parts) == 0 {
		return errors.New("invalid path")
	}

	// Navigate to parent directory
	current := m.root
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return errors.New("parent directory not found")
		}
	}

	// Create or update file
	filename = parts[len(parts)-1]
	current.Children[filename] = &MockFile{
		Name:     filename,
		Content:  data,
		Mode:     perm,
		ModTime:  time.Now(),
		IsDir:    false,
	}

	return nil
}

// ReadFile reads content from a file
func (m *MockFilesystem) ReadFile(filename string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode {
		return nil, errors.New("mock error: read")
	}

	parts := splitPath(filename)
	current := m.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return nil, os.ErrNotExist
		}
	}

	if current.IsDir {
		return nil, errors.New("is a directory")
	}

	return current.Content, nil
}

// Remove removes a file or directory
func (m *MockFilesystem) Remove(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errMode {
		return errors.New("mock error: remove")
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return errors.New("invalid path")
	}

	current := m.root
	for i := 0; i < len(parts)-1; i++ {
		if child, ok := current.Children[parts[i]]; ok {
			current = child
		} else {
			return os.ErrNotExist
		}
	}

	delete(current.Children, parts[len(parts)-1])
	return nil
}

// Stat returns file info
func (m *MockFilesystem) Stat(filename string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode {
		return nil, errors.New("mock error: stat")
	}

	parts := splitPath(filename)
	current := m.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return nil, os.ErrNotExist
		}
	}

	return &mockFileInfo{
		name:    current.Name,
		size:    int64(len(current.Content)),
		mode:    current.Mode,
		modTime: current.ModTime,
		isDir:   current.IsDir,
	}, nil
}

// ReadDir reads directory contents
func (m *MockFilesystem) ReadDir(dirname string) ([]os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.errMode {
		return nil, errors.New("mock error: readdir")
	}

	parts := splitPath(dirname)
	current := m.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return nil, os.ErrNotExist
		}
	}

	if !current.IsDir {
		return nil, errors.New("not a directory")
	}

	var infos []os.FileInfo
	for _, child := range current.Children {
		infos = append(infos, &mockFileInfo{
			name:    child.Name,
			size:    int64(len(child.Content)),
			mode:    child.Mode,
			modTime: child.ModTime,
			isDir:   child.IsDir,
		})
	}

	return infos, nil
}

// Exists checks if a file exists
func (m *MockFilesystem) Exists(filename string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	parts := splitPath(filename)
	current := m.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return false
		}
	}

	return true
}

// SetErrorMode enables or disables error mode
func (m *MockFilesystem) SetErrorMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errMode = enabled
}

// CreateFile creates a file with content
func (m *MockFilesystem) CreateFile(path string, content []byte) error {
	return m.WriteFile(path, content, 0644)
}

// helper functions

func splitPath(path string) []string {
	if path == "/" || path == "" {
		return nil
	}

	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

// mockFileInfo implements os.FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64         { return m.size }
func (m *mockFileInfo) Mode() os.FileMode   { return m.mode }
func (m *mockFileInfo) ModTime() time.Time  { return m.modTime }
func (m *mockFileInfo) IsDir() bool         { return m.isDir }
func (m *mockFileInfo) Sys() interface{}    { return nil }

// MockFile implements io.Reader and io.Writer
func (m *MockFile) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	m.Content = append(m.Content, p...)
	return len(p), nil
}