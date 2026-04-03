// Package runner provides CI/CD runner orchestration
package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// RunnerManager manages CI/CD runners across multiple platforms
type RunnerManager struct {
	db          types.PipelineDB
	poolManager types.PoolManagerInterface
	gitlab      *GitLabRunner
	github      *GitHubRunner
	jenkins     *JenkinsRunner
	circleci    *CircleCIRunner
	drone       *DroneRunner
	runners     map[string]*RunnerInfo
	mu          sync.RWMutex
}

// RunnerInfo represents a unified runner interface
type RunnerInfo struct {
	ID           string              `json:"id"`
	VMID         string              `json:"vm_id"`
	PoolName     string              `json:"pool_name"`
	PipelineID   string              `json:"pipeline_id"`
	Platform     types.RunnerPlatform `json:"platform"`
	PlatformID   string              `json:"platform_runner_id"`
	Name         string              `json:"name"`
	Labels       []string            `json:"labels"`
	Status       RunnerStatus        `json:"status"`
	IPAddress    string              `json:"ip_address"`
	CurrentJob   string              `json:"current_job,omitempty"`
	HealthStatus string              `json:"health_status"`
	LastHeartbeat *time.Time          `json:"last_heartbeat,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	DestroyedAt  *time.Time          `json:"destroyed_at,omitempty"`
}

// RunnerStatus represents the status of a runner
type RunnerStatus string

const (
	RunnerStatusCreating  RunnerStatus = "creating"
	RunnerStatusOnline    RunnerStatus = "online"
	RunnerStatusOffline   RunnerStatus = "offline"
	RunnerStatusBusy      RunnerStatus = "busy"
	RunnerStatusError     RunnerStatus = "error"
	RunnerStatusDestroyed RunnerStatus = "destroyed"
)

// RunnerManagerConfig represents runner manager configuration
type RunnerManagerConfig struct {
	GitLab   *GitLabConfig   `json:"gitlab"`
	GitHub   *GitHubConfig   `json:"github"`
	Jenkins  *JenkinsConfig  `json:"jenkins"`
	CircleCI *CircleCIConfig `json:"circleci"`
	Drone    *DroneConfig    `json:"drone"`
}

// NewRunnerManager creates a new runner manager
func NewRunnerManager(db *types.PipelineDB, poolManager *pool.PoolManager, config *RunnerManagerConfig) (*RunnerManager, error) {
	rm := &RunnerManager{
		db:          db,
		poolManager: poolManager,
		runners:     make(map[string]*RunnerInfo),
	}

	// Initialize platform-specific runners
	if config.GitLab != nil {
		gitlab, err := NewGitLabRunner(db, config.GitLab)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab runner: %w", err)
		}
		rm.gitlab = gitlab
	}

	if config.GitHub != nil {
		github, err := NewGitHubRunner(db, config.GitHub)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub runner: %w", err)
		}
		rm.github = github
	}

	if config.Jenkins != nil {
		jenkins, err := NewJenkinsRunner(db, config.Jenkins)
		if err != nil {
			return nil, fmt.Errorf("failed to create Jenkins runner: %w", err)
		}
		rm.jenkins = jenkins
	}

	if config.CircleCI != nil {
		circleci, err := NewCircleCIRunner(db, config.CircleCI)
		if err != nil {
			return nil, fmt.Errorf("failed to create CircleCI runner: %w", err)
		}
		rm.circleci = circleci
	}

	if config.Drone != nil {
		drone, err := NewDroneRunner(db, config.Drone)
		if err != nil {
			return nil, fmt.Errorf("failed to create Drone runner: %w", err)
		}
		rm.drone = drone
	}

	return rm, nil
}

// CreateRunner creates a new runner for a pipeline
func (rm *RunnerManager) CreateRunner(ctx context.Context, poolName string, platform types.RunnerPlatform, pipelineID string, labels []string) (*RunnerInfo, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Acquire VM from pool
	vm, err := rm.poolManager.AcquireVM(ctx, poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire VM: %w", err)
	}

	// Create platform-specific runner
	var runnerID, platformID string
	var err error

	switch platform {
	case pipeline.PlatformGitLab:
		if rm.gitlab == nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("GitLab runner not configured")
		}
		info, err := rm.gitlab.RegisterRunner(ctx, vm.ID, pipelineID)
		if err != nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("failed to register GitLab runner: %w", err)
		}
		runnerID = info.ID
		platformID = fmt.Sprintf("%d", info.PlatformID)

	case pipeline.PlatformGitHub:
		if rm.github == nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("GitHub runner not configured")
		}
		info, err := rm.github.RegisterRunner(ctx, vm.ID, pipelineID)
		if err != nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("failed to register GitHub runner: %w", err)
		}
		runnerID = info.ID
		platformID = fmt.Sprintf("%d", info.RunnerID)

	case pipeline.PlatformJenkins:
		if rm.jenkins == nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("Jenkins runner not configured")
		}
		info, err := rm.jenkins.RegisterRunner(ctx, vm.ID, pipelineID)
		if err != nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("failed to register Jenkins runner: %w", err)
		}
		runnerID = info.ID
		platformID = info.AgentName

	case pipeline.PlatformCircleCI:
		if rm.circleci == nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("CircleCI runner not configured")
		}
		info, err := rm.circleci.RegisterRunner(ctx, vm.ID, pipelineID)
		if err != nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("failed to register CircleCI runner: %w", err)
		}
		runnerID = info.ID
		platformID = info.ResourceClass

	case pipeline.PlatformDrone:
		if rm.drone == nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("Drone runner not configured")
		}
		info, err := rm.drone.RegisterRunner(ctx, vm.ID, pipelineID)
		if err != nil {
			rm.poolManager.ReleaseVM(ctx, vm.ID)
			return nil, fmt.Errorf("failed to register Drone runner: %w", err)
		}
		runnerID = info.ID
		platformID = info.Token[:8]

	default:
		rm.poolManager.ReleaseVM(ctx, vm.ID)
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	if err != nil {
		rm.poolManager.ReleaseVM(ctx, vm.ID)
		return nil, err
	}

	// Create runner info
	runner := &RunnerInfo{
		ID:         runnerID,
		VMID:       vm.ID,
		PoolName:   poolName,
		PipelineID: pipelineID,
		Platform:    platform,
		PlatformID: platformID,
		Name:       fmt.Sprintf("%s-runner-%s", platform, runnerID[:8]),
		Labels:     labels,
		Status:     RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}

	rm.runners[runnerID] = runner

	return runner, nil
}

// StartRunner starts a runner on its VM
func (rm *RunnerManager) StartRunner(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Get VM
	vm, err := rm.poolManager.GetVM(runner.VMID)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}

	// Start platform-specific runner
	switch runner.Platform {
	case pipeline.PlatformGitLab:
		if err := rm.gitlab.StartRunner(ctx, runnerID, vm.IP); err != nil {
			return fmt.Errorf("failed to start GitLab runner: %w", err)
		}

	case pipeline.PlatformGitHub:
		if err := rm.github.StartRunner(ctx, runnerID, vm.IP); err != nil {
			return fmt.Errorf("failed to start GitHub runner: %w", err)
		}

	case pipeline.PlatformJenkins:
		if err := rm.jenkins.StartRunner(ctx, runnerID, vm.IP); err != nil {
			return fmt.Errorf("failed to start Jenkins runner: %w", err)
		}

	case pipeline.PlatformCircleCI:
		if err := rm.circleci.StartRunner(ctx, runnerID, vm.IP); err != nil {
			return fmt.Errorf("failed to start CircleCI runner: %w", err)
		}

	case pipeline.PlatformDrone:
		if err := rm.drone.StartRunner(ctx, runnerID, vm.IP); err != nil {
			return fmt.Errorf("failed to start Drone runner: %w", err)
		}

	default:
		return fmt.Errorf("unsupported platform: %s", runner.Platform)
	}

	runner.Status = RunnerStatusOnline
	runner.IPAddress = vm.IP

	return nil
}

// StopRunner stops a runner
func (rm *RunnerManager) StopRunner(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Stop platform-specific runner
	switch runner.Platform {
	case pipeline.PlatformGitLab:
		if err := rm.gitlab.StopRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to stop GitLab runner: %w", err)
		}

	case pipeline.PlatformGitHub:
		if err := rm.github.StopRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to stop GitHub runner: %w", err)
		}

	case pipeline.PlatformJenkins:
		if err := rm.jenkins.StopRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to stop Jenkins runner: %w", err)
		}

	case pipeline.PlatformCircleCI:
		if err := rm.circleci.StopRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to stop CircleCI runner: %w", err)
		}

	case pipeline.PlatformDrone:
		if err := rm.drone.StopRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to stop Drone runner: %w", err)
		}

	default:
		return fmt.Errorf("unsupported platform: %s", runner.Platform)
	}

	runner.Status = RunnerStatusOffline

	return nil
}

// DestroyRunner destroys a runner and releases its VM
func (rm *RunnerManager) DestroyRunner(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Stop runner if running
	if runner.Status == RunnerStatusOnline || runner.Status == RunnerStatusBusy {
		rm.StopRunner(ctx, runnerID)
	}

	// Unregister from platform
	switch runner.Platform {
	case pipeline.PlatformGitLab:
		if err := rm.gitlab.UnregisterRunner(ctx, runnerID); err != nil {
			// Log but continue
			fmt.Printf("[RunnerManager] Warning: failed to unregister GitLab runner: %v\n", err)
		}

	case pipeline.PlatformGitHub:
		if err := rm.github.UnregisterRunner(ctx, runnerID); err != nil {
			fmt.Printf("[RunnerManager] Warning: failed to unregister GitHub runner: %v\n", err)
		}

	case pipeline.PlatformJenkins:
		if err := rm.jenkins.UnregisterRunner(ctx, runnerID); err != nil {
			fmt.Printf("[RunnerManager] Warning: failed to unregister Jenkins runner: %v\n", err)
		}

	case pipeline.PlatformCircleCI:
		if err := rm.circleci.UnregisterRunner(ctx, runnerID); err != nil {
			fmt.Printf("[RunnerManager] Warning: failed to unregister CircleCI runner: %v\n", err)
		}

	case pipeline.PlatformDrone:
		if err := rm.drone.UnregisterRunner(ctx, runnerID); err != nil {
			fmt.Printf("[RunnerManager] Warning: failed to unregister Drone runner: %v\n", err)
		}
	}

	// Release VM
	if err := rm.poolManager.ReleaseVM(ctx, runner.VMID); err != nil {
		fmt.Printf("[RunnerManager] Warning: failed to release VM: %v\n", err)
	}

	// Mark as destroyed
	now := time.Now()
	runner.DestroyedAt = &now
	runner.Status = RunnerStatusDestroyed

	// Delete from memory
	delete(rm.runners, runnerID)

	return nil
}

// GetRunner returns a runner by ID
func (rm *RunnerManager) GetRunner(runnerID string) (*RunnerInfo, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}

	return runner, nil
}

// ListRunners returns all runners
func (rm *RunnerManager) ListRunners() []*RunnerInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runners := make([]*RunnerInfo, 0, len(rm.runners))
	for _, runner := range rm.runners {
		runners = append(runners, runner)
	}
	return runners
}

// ListRunnersByPipeline returns runners for a pipeline
func (rm *RunnerManager) ListRunnersByPipeline(pipelineID string) []*RunnerInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runners := make([]*RunnerInfo, 0)
	for _, runner := range rm.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

// ListRunnersByPlatform returns runners for a platform
func (rm *RunnerManager) ListRunnersByPlatform(platform types.RunnerPlatform) []*RunnerInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runners := make([]*RunnerInfo, 0)
	for _, runner := range rm.runners {
		if runner.Platform == platform {
			runners = append(runners, runner)
		}
	}
	return runners
}

// UpdateRunnerStatus updates runner status
func (rm *RunnerManager) UpdateRunnerStatus(runnerID string, status RunnerStatus, currentJob string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	runner.Status = status
	runner.CurrentJob = currentJob
	now := time.Now()
	runner.LastHeartbeat = &now

	return nil
}

// CheckHealth checks runner health
func (rm *RunnerManager) CheckHealth(ctx context.Context, runnerID string) (string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	runner, ok := rm.runners[runnerID]
	if !ok {
		return "", fmt.Errorf("runner not found: %s", runnerID)
	}

	// Check if heartbeat is recent
	if runner.LastHeartbeat == nil {
		return "unknown", nil
	}

	timeSinceHeartbeat := time.Since(*runner.LastHeartbeat)
	if timeSinceHeartbeat > 5*time.Minute {
		return "unhealthy", nil
	}

	if timeSinceHeartbeat > 2*time.Minute {
		return "degraded", nil
	}

	return "healthy", nil
}

// GetStats returns runner statistics
func (rm *RunnerManager) GetStats() map[string]int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]int{
		"total": len(rm.runners),
	}

	for _, runner := range rm.runners {
		stats[string(runner.Status)]++
	}

	return stats
}

// Close closes the runner manager
func (rm *RunnerManager) Close() error {
	// Nothing to close
	return nil
}