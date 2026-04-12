// Package mockrunner provides a mock RunnerManagerInterface for testing
package mockrunner

import (
	"sync"

	"github.com/stsgym/vimic2/internal/types"
)

// MockRunnerManager implements types.RunnerManagerInterface for testing
type MockRunnerManager struct {
	mu      sync.RWMutex
	runners map[string]map[string]interface{}
	nextID  int
}

// NewMockRunnerManager creates a new mock runner manager
func NewMockRunnerManager() *MockRunnerManager {
	return &MockRunnerManager{
		runners: make(map[string]map[string]interface{}),
		nextID:  1,
	}
}

// CreateRunner creates a new runner
func (m *MockRunnerManager) CreateRunner(platform types.RunnerPlatform, config map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID
	id := "runner-" + string(platform) + "-" + randomID(8)
	m.nextID++

	// Store runner
	m.runners[id] = config

	return id, nil
}

// DestroyRunner destroys a runner
func (m *MockRunnerManager) DestroyRunner(runnerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runners, runnerID)
	return nil
}

// GetRunner gets a runner by ID
func (m *MockRunnerManager) GetRunner(runnerID string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, ok := m.runners[runnerID]
	if !ok {
		// Return default config
		return map[string]interface{}{
			"id":     runnerID,
			"status": "idle",
		}, nil
	}

	return config, nil
}

// AddRunner adds a runner to the manager (for testing)
func (m *MockRunnerManager) AddRunner(id string, config map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.runners[id] = config
}

// ListRunners lists all runner IDs
func (m *MockRunnerManager) ListRunners() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.runners))
	for id := range m.runners {
		ids = append(ids, id)
	}

	// If no runners, return default
	if len(ids) == 0 {
		ids = append(ids, "runner-default")
	}

	return ids, nil
}

// Verify MockRunnerManager implements types.RunnerManagerInterface
var _ types.RunnerManagerInterface = (*MockRunnerManager)(nil)

// randomID generates a random ID string
func randomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
