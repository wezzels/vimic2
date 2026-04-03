// Package runner provides CI/CD runner orchestration
package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// RunnerManager manages CI/CD runners across multiple platforms
type RunnerManager struct {
	db          types.PipelineDB
	poolManager types.PoolManagerInterface
	runners     map[string]*RunnerInfo
	mu          sync.RWMutex
}

// RunnerInfo represents a unified runner interface
type RunnerInfo struct {
	ID           string            `json:"id"`
	VMID         string            `json:"vm_id"`
	PoolName     string            `json:"pool_name"`
	PipelineID   string            `json:"pipeline_id"`
	Platform     types.RunnerPlatform `json:"platform"`
	PlatformID   string            `json:"platform_runner_id"`
	Name         string            `json:"name"`
	Labels       []string          `json:"labels"`
	Status       types.RunnerStatus `json:"status"`
	IPAddress    string            `json:"ip_address"`
	CurrentJob   string            `json:"current_job,omitempty"`
	HealthStatus string            `json:"health_status"`
	LastHeartbeat *time.Time       `json:"last_heartbeat,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	DestroyedAt  *time.Time        `json:"destroyed_at,omitempty"`
}

// RunnerManagerConfig represents runner manager configuration
type RunnerManagerConfig struct {
	GitLab    *GitLabConfig   `json:"gitlab"`
	GitHub    *GitHubConfig   `json:"github"`
	Jenkins   *JenkinsConfig  `json:"jenkins"`
	CircleCI  *CircleCIConfig `json:"circleci"`
	Drone     *DroneConfig    `json:"drone"`
}

// NewRunnerManager creates a new runner manager
func NewRunnerManager(db types.PipelineDB, poolManager types.PoolManagerInterface, config *RunnerManagerConfig) (*RunnerManager, error) {
	rm := &RunnerManager{
		db:       db,
		poolManager: poolManager,
		runners:  make(map[string]*RunnerInfo),
	}

	return rm, nil
}

// CreateRunner creates a new runner for a pipeline
func (rm *RunnerManager) CreateRunner(ctx context.Context, poolName string, platform types.RunnerPlatform, pipelineID string, labels []string) (*RunnerInfo, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner := &RunnerInfo{
		ID:         generateID("runner"),
		PoolName:   poolName,
		Platform:   platform,
		PipelineID: pipelineID,
		Labels:     labels,
		Status:     types.RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}

	rm.runners[runner.ID] = runner
	return runner, nil
}

// GetRunner returns a runner by ID
func (rm *RunnerManager) GetRunner(runnerID string) (*RunnerInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runner, exists := rm.runners[runnerID]
	if !exists {
		return nil, ErrRunnerNotFound
	}

	return runner, nil
}

// ListRunners returns all runners
func (rm *RunnerManager) ListRunners() ([]*RunnerInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runners := make([]*RunnerInfo, 0, len(rm.runners))
	for _, runner := range rm.runners {
		runners = append(runners, runner)
	}

	return runners, nil
}

// DestroyRunner destroys a runner
func (rm *RunnerManager) DestroyRunner(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner, exists := rm.runners[runnerID]
	if !exists {
		return ErrRunnerNotFound
	}

	now := time.Now()
	runner.DestroyedAt = &now
	runner.Status = types.RunnerStatusDestroyed

	delete(rm.runners, runnerID)
	return nil
}

// ErrRunnerNotFound is returned when a runner is not found
var ErrRunnerNotFound = errors.New("runner not found")

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, randomString(4), time.Now().UnixNano())
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Placeholder implementations for platform-specific runners
type GitLabConfig struct {
	Token    string `json:"token"`
	URL      string `json:"url"`
	Executor string `json:"executor"`
}

type GitHubConfig struct {
	Token    string `json:"token"`
	Repo     string `json:"repo"`
	Labels   []string `json:"labels"`
}

type JenkinsConfig struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Token    string `json:"token"`
}

type CircleCIConfig struct {
	Token    string `json:"token"`
	Org      string `json:"org"`
}

type DroneConfig struct {
	Token    string `json:"token"`
	URL      string `json:"url"`
}
