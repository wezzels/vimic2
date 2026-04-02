// Package pipeline provides pipeline coordination
package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/network"
	"github.com/stsgym/vimic2/internal/pool"
	"github.com/stsgym/vimic2/internal/runner"
)

// Coordinator coordinates pipeline execution
type Coordinator struct {
	db             *PipelineDB
	poolManager    *pool.PoolManager
	networkManager *network.IsolationManager
	runnerManager  *runner.RunnerManager
	pipelines      map[string]*PipelineState
	mu             sync.RWMutex
	eventChan      chan PipelineEvent
}

// PipelineState represents the state of a pipeline
type PipelineState struct {
	ID           string           `json:"id"`
	Platform     RunnerPlatform   `json:"platform"`
	Repository   string           `json:"repository"`
	Branch       string           `json:"branch"`
	CommitSHA    string           `json:"commit_sha"`
	CommitMsg    string           `json:"commit_message"`
	Author       string           `json:"author"`
	Status       PipelineStatus   `json:"status"`
	NetworkID    string           `json:"network_id"`
	VMs          []string         `json:"vms"`
	Runners      []string         `json:"runners"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	Duration     int64            `json:"duration_seconds"`
	CurrentStage string           `json:"current_stage"`
	Stages       []StageState     `json:"stages"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// StageState represents the state of a pipeline stage
type StageState struct {
	Name      string          `json:"name"`
	Status    PipelineStatus   `json:"status"`
	StartTime *time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time      `json:"end_time,omitempty"`
	Jobs      []JobState      `json:"jobs"`
}

// JobState represents the state of a job
type JobState struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Stage       string          `json:"stage"`
	Status      PipelineStatus  `json:"status"`
	RunnerID    string          `json:"runner_id"`
	StartTime   *time.Time      `json:"start_time,omitempty"`
	EndTime     *time.Time      `json:"end_time,omitempty"`
	Duration    int64           `json:"duration_seconds"`
	Log         []string        `json:"log,omitempty"`
}

// PipelineEvent represents a pipeline state change event
type PipelineEvent struct {
	PipelineID   string          `json:"pipeline_id"`
	OldStatus    PipelineStatus  `json:"old_status"`
	NewStatus    PipelineStatus  `json:"new_status"`
	Stage        string          `json:"stage,omitempty"`
	JobID        string          `json:"job_id,omitempty"`
	Message      string          `json:"message"`
	Timestamp    time.Time        `json:"timestamp"`
}

// CoordinatorConfig represents coordinator configuration
type CoordinatorConfig struct {
	PoolSize      int              `json:"pool_size"`
	DefaultBranch string           `json:"default_branch"`
	Timeout       time.Duration    `json:"timeout"`
	MaxPipelines  int              `json:"max_pipelines"`
}

// NewCoordinator creates a new pipeline coordinator
func NewCoordinator(db *PipelineDB, poolMgr *pool.PoolManager, netMgr *network.IsolationManager, runnerMgr *runner.RunnerManager) (*Coordinator, error) {
	c := &Coordinator{
		db:             db,
		poolManager:    poolMgr,
		networkManager: netMgr,
		runnerManager:  runnerMgr,
		pipelines:      make(map[string]*PipelineState),
		eventChan:      make(chan PipelineEvent, 1000),
	}

	// Load existing pipelines from database
	if err := c.loadPipelines(); err != nil {
		return nil, fmt.Errorf("failed to load pipelines: %w", err)
	}

	// Start event processor
	go c.processEvents()

	return c, nil
}

// loadPipelines loads pipelines from database
func (c *Coordinator) loadPipelines() error {
	ctx := context.Background()
	pipelines, err := c.db.ListPipelines(ctx, 100, 0)
	if err != nil {
		return err
	}

	for _, p := range pipelines {
		state := &PipelineState{
			ID:         p.ID,
			Platform:   p.Platform,
			Repository: p.Repository,
			Branch:     p.Branch,
			CommitSHA: p.CommitSHA,
			CommitMsg:  p.CommitMsg,
			Author:     p.Author,
			Status:     p.Status,
			NetworkID:  p.NetworkID,
			StartTime:  p.StartTime,
			EndTime:    p.EndTime,
			Duration:   p.Duration,
			CreatedAt:  p.CreatedAt,
			UpdatedAt:  p.UpdatedAt,
		}
		c.pipelines[p.ID] = state
	}

	return nil
}

// CreatePipeline creates a new pipeline with network and runners
func (c *Coordinator) CreatePipeline(ctx context.Context, platform RunnerPlatform, repo, branch, commitSHA, commitMsg, author string, runnerCount int, labels []string) (*PipelineState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate pipeline ID
	pipelineID := generatePipelineID()

	// Create network for isolation
	net, err := c.networkManager.CreateNetwork(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	// Create pipeline state
	state := &PipelineState{
		ID:         pipelineID,
		Platform:   platform,
		Repository: repo,
		Branch:     branch,
		CommitSHA:  commitSHA,
		CommitMsg:  commitMsg,
		Author:     author,
		Status:     PipelineStatusCreating,
		NetworkID:  net.ID,
		VMs:        []string{},
		Runners:    []string{},
		StartTime:  time.Now(),
		Stages:     []StageState{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to database
	p := &Pipeline{
		ID:         state.ID,
		Platform:   state.Platform,
		Repository: state.Repository,
		Branch:     state.Branch,
		CommitSHA:  state.CommitSHA,
		CommitMsg:  state.CommitMsg,
		Author:     state.Author,
		Status:     state.Status,
		NetworkID:  state.NetworkID,
		StartTime:  state.StartTime,
		CreatedAt:  state.CreatedAt,
		UpdatedAt:  state.UpdatedAt,
	}

	if err := c.db.SavePipeline(ctx, p); err != nil {
		c.networkManager.DeleteNetwork(ctx, net.ID)
		return nil, fmt.Errorf("failed to save pipeline: %w", err)
	}

	// Create runners
	for i := 0; i < runnerCount; i++ {
		poolName := c.getPoolName(platform)
		runnerInfo, err := c.runnerManager.CreateRunner(ctx, poolName, platform, pipelineID, labels)
		if err != nil {
			// Cleanup on error
			c.cleanupPipeline(ctx, state)
			return nil, fmt.Errorf("failed to create runner %d: %w", i, err)
		}

		state.Runners = append(state.Runners, runnerInfo.ID)
		state.VMs = append(state.VMs, runnerInfo.VMID)
	}

	// Update status
	state.Status = PipelineStatusRunning
	state.UpdatedAt = time.Now()

	c.db.UpdatePipelineStatus(ctx, pipelineID, PipelineStatusRunning)

	// Store in memory
	c.pipelines[pipelineID] = state

	// Emit event
	c.emitEvent(PipelineEvent{
		PipelineID:   pipelineID,
		OldStatus:    PipelineStatusCreating,
		NewStatus:    PipelineStatusRunning,
		Message:      "Pipeline created successfully",
		Timestamp:    time.Now(),
	})

	return state, nil
}

// StartPipeline starts all runners for a pipeline
func (c *Coordinator) StartPipeline(ctx context.Context, pipelineID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Start each runner
	for _, runnerID := range state.Runners {
		if err := c.runnerManager.StartRunner(ctx, runnerID); err != nil {
			return fmt.Errorf("failed to start runner %s: %w", runnerID, err)
		}
	}

	// Update status
	state.Status = PipelineStatusRunning
	state.UpdatedAt = time.Now()
	c.db.UpdatePipelineStatus(ctx, pipelineID, PipelineStatusRunning)

	// Emit event
	c.emitEvent(PipelineEvent{
		PipelineID:   pipelineID,
		OldStatus:    PipelineStatusCreating,
		NewStatus:    PipelineStatusRunning,
		Message:      "Pipeline started",
		Timestamp:    time.Now(),
	})

	return nil
}

// StopPipeline stops all runners for a pipeline
func (c *Coordinator) StopPipeline(ctx context.Context, pipelineID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Stop each runner
	for _, runnerID := range state.Runners {
		if err := c.runnerManager.StopRunner(ctx, runnerID); err != nil {
			// Log but continue
			fmt.Printf("[Coordinator] Warning: failed to stop runner %s: %v\n", runnerID, err)
		}
	}

	// Update status
	state.Status = PipelineStatusCanceled
	now := time.Now()
	state.EndTime = &now
	state.UpdatedAt = now

	c.db.UpdatePipelineStatus(ctx, pipelineID, PipelineStatusCanceled)

	// Emit event
	c.emitEvent(PipelineEvent{
		PipelineID:   pipelineID,
		OldStatus:    PipelineStatusRunning,
		NewStatus:    PipelineStatusCanceled,
		Message:      "Pipeline stopped",
		Timestamp:    now,
	})

	return nil
}

// DestroyPipeline destroys a pipeline and cleans up resources
func (c *Coordinator) DestroyPipeline(ctx context.Context, pipelineID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Cleanup pipeline resources
	if err := c.cleanupPipeline(ctx, state); err != nil {
		return fmt.Errorf("failed to cleanup pipeline: %w", err)
	}

	// Delete from memory
	delete(c.pipelines, pipelineID)

	return nil
}

// cleanupPipeline cleans up all pipeline resources
func (c *Coordinator) cleanupPipeline(ctx context.Context, state *PipelineState) error {
	// Stop and destroy runners
	for _, runnerID := range state.Runners {
		if err := c.runnerManager.DestroyRunner(ctx, runnerID); err != nil {
			fmt.Printf("[Coordinator] Warning: failed to destroy runner %s: %v\n", runnerID, err)
		}
	}

	// Delete network
	if state.NetworkID != "" {
		if err := c.networkManager.DeleteNetwork(ctx, state.NetworkID); err != nil {
			fmt.Printf("[Coordinator] Warning: failed to delete network %s: %v\n", state.NetworkID, err)
		}
	}

	// Update status
	state.Status = PipelineStatusFailed
	now := time.Now()
	state.EndTime = &now
	state.UpdatedAt = now

	c.db.UpdatePipelineStatus(ctx, state.ID, PipelineStatusFailed)

	// Emit event
	c.emitEvent(PipelineEvent{
		PipelineID:   state.ID,
		OldStatus:    state.Status,
		NewStatus:    PipelineStatusFailed,
		Message:      "Pipeline destroyed",
		Timestamp:    now,
	})

	return nil
}

// GetPipeline returns a pipeline by ID
func (c *Coordinator) GetPipeline(pipelineID string) (*PipelineState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	return state, nil
}

// ListPipelines returns all pipelines
func (c *Coordinator) ListPipelines(limit, offset int) []*PipelineState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pipelines := make([]*PipelineState, 0, len(c.pipelines))
	for _, state := range c.pipelines {
		pipelines = append(pipelines, state)
	}

	// Sort by created_at descending
	// TODO: implement proper sorting with limit/offset

	if limit > 0 && len(pipelines) > limit {
		pipelines = pipelines[:limit]
	}

	return pipelines
}

// ListPipelinesByStatus returns pipelines by status
func (c *Coordinator) ListPipelinesByStatus(status PipelineStatus) []*PipelineState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pipelines := make([]*PipelineState, 0)
	for _, state := range c.pipelines {
		if state.Status == status {
			pipelines = append(pipelines, state)
		}
	}

	return pipelines
}

// UpdateStageStatus updates the status of a stage
func (c *Coordinator) UpdateStageStatus(ctx context.Context, pipelineID, stageName string, status PipelineStatus) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Find or create stage
	var stage *StageState
	for i := range state.Stages {
		if state.Stages[i].Name == stageName {
			stage = &state.Stages[i]
			break
		}
	}

	if stage == nil {
		state.Stages = append(state.Stages, StageState{
			Name:   stageName,
			Status: status,
		})
		stage = &state.Stages[len(state.Stages)-1]
	}

	// Update stage status
	stage.Status = status
	now := time.Time{}
	if status == PipelineStatusRunning && stage.StartTime == nil {
		stage.StartTime = &now
	} else if status == PipelineStatusSuccess || status == PipelineStatusFailed {
		stage.EndTime = &now
	}

	state.CurrentStage = stageName
	state.UpdatedAt = now

	// Check if all stages are complete
	allComplete := true
	for _, s := range state.Stages {
		if s.Status != PipelineStatusSuccess && s.Status != PipelineStatusFailed {
			allComplete = false
			break
		}
	}

	if allComplete {
		state.Status = PipelineStatusSuccess
		state.EndTime = &now
		c.db.UpdatePipelineStatus(ctx, pipelineID, PipelineStatusSuccess)
	}

	// Emit event
	c.emitEvent(PipelineEvent{
		PipelineID:   pipelineID,
		OldStatus:    state.Status,
		NewStatus:    status,
		Stage:        stageName,
		Message:      fmt.Sprintf("Stage %s updated to %s", stageName, status),
		Timestamp:    now,
	})

	return nil
}

// UpdateJobStatus updates the status of a job
func (c *Coordinator) UpdateJobStatus(ctx context.Context, pipelineID, stageName, jobID string, status PipelineStatus) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Find stage
	var stage *StageState
	for i := range state.Stages {
		if state.Stages[i].Name == stageName {
			stage = &state.Stages[i]
			break
		}
	}

	if stage == nil {
		return fmt.Errorf("stage not found: %s", stageName)
	}

	// Find job
	var job *JobState
	for i := range stage.Jobs {
		if stage.Jobs[i].ID == jobID {
			job = &stage.Jobs[i]
			break
		}
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update job status
	job.Status = status
	now := time.Time{}
	if status == PipelineStatusRunning && job.StartTime == nil {
		job.StartTime = &now
	} else if status == PipelineStatusSuccess || status == PipelineStatusFailed {
		job.EndTime = &now
		if job.StartTime != nil {
			job.Duration = int64(now.Sub(*job.StartTime).Seconds())
		}
	}

	state.UpdatedAt = now

	return nil
}

// AddLogLine adds a log line to a job
func (c *Coordinator) AddLogLine(ctx context.Context, pipelineID, stageName, jobID, line string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.pipelines[pipelineID]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Find stage
	var stage *StageState
	for i := range state.Stages {
		if state.Stages[i].Name == stageName {
			stage = &state.Stages[i]
			break
		}
	}

	if stage == nil {
		return fmt.Errorf("stage not found: %s", stageName)
	}

	// Find job
	var job *JobState
	for i := range stage.Jobs {
		if stage.Jobs[i].ID == jobID {
			job = &stage.Jobs[i]
			break
		}
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Add log line
	job.Log = append(job.Log, line)

	return nil
}

// GetStats returns pipeline statistics
func (c *Coordinator) GetStats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]int{
		"total": len(c.pipelines),
	}

	for _, state := range c.pipelines {
		stats[string(state.Status)]++
	}

	return stats
}

// Events returns the event channel
func (c *Coordinator) Events() <-chan PipelineEvent {
	return c.eventChan
}

// emitEvent emits a pipeline event
func (c *Coordinator) emitEvent(event PipelineEvent) {
	select {
	case c.eventChan <- event:
	default:
		// Channel full, drop event
	}
}

// processEvents processes pipeline events
func (c *Coordinator) processEvents() {
	for event := range c.eventChan {
		// Log event
		fmt.Printf("[Coordinator] Pipeline %s: %s -> %s (%s)\n",
			event.PipelineID, event.OldStatus, event.NewStatus, event.Message)

		// Could emit to external systems here (webhooks, etc.)
	}
}

// getPoolName returns the pool name for a platform
func (c *Coordinator) getPoolName(platform RunnerPlatform) string {
	switch platform {
	case PlatformGitLab:
		return "gitlab-runner"
	case PlatformGitHub:
		return "github-runner"
	case PlatformJenkins:
		return "jenkins-agent"
	case PlatformCircleCI:
		return "circleci-runner"
	case PlatformDrone:
		return "drone-runner"
	default:
		return "default"
	}
}

// Close closes the coordinator
func (c *Coordinator) Close() error {
	close(c.eventChan)
	return nil
}

// Helper functions

func generatePipelineID() string {
	return fmt.Sprintf("pipeline-%s-%d", randomString(8), time.Now().Unix())
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}