// Package pipeline provides pipeline coordination
package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// Coordinator coordinates pipeline execution
type Coordinator struct {
	db             *PipelineDB
	poolManager    types.PoolManagerInterface
	networkManager types.NetworkManagerInterface
	runnerManager  types.RunnerManagerInterface
	pipelines      map[string]*PipelineState
	mu             sync.RWMutex
	eventChan      chan PipelineEvent
}

// PipelineState represents the state of a pipeline
type PipelineState struct {
	ID           string          `json:"id"`
	Platform     types.RunnerPlatform `json:"platform"`
	Repository   string          `json:"repository"`
	Branch       string          `json:"branch"`
	CommitSHA    string          `json:"commit_sha"`
	CommitMsg    string          `json:"commit_message"`
	Author       string          `json:"author"`
	Status       types.PipelineStatus `json:"status"`
	NetworkID    string          `json:"network_id"`
	VMs          []string        `json:"vms"`
	Runners      []string        `json:"runners"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      *time.Time      `json:"end_time,omitempty"`
	Duration     int64           `json:"duration_seconds"`
	CurrentStage string          `json:"current_stage"`
	Stages       []StageState    `json:"stages"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// StageState represents the state of a pipeline stage
type StageState struct {
	Name      string               `json:"name"`
	Status    types.PipelineStatus `json:"status"`
	StartTime *time.Time           `json:"start_time,omitempty"`
	EndTime   *time.Time           `json:"end_time,omitempty"`
	Jobs      []JobState           `json:"jobs"`
}

// JobState represents the state of a job
type JobState struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Stage     string               `json:"stage"`
	Status    types.PipelineStatus  `json:"status"`
	RunnerID  string               `json:"runner_id"`
	StartTime *time.Time           `json:"start_time,omitempty"`
	EndTime   *time.Time           `json:"end_time,omitempty"`
	Duration  int64                `json:"duration_seconds"`
	Log       []string             `json:"log,omitempty"`
}

// PipelineEvent represents a pipeline state change event
type PipelineEvent struct {
	PipelineID string               `json:"pipeline_id"`
	OldStatus  types.PipelineStatus  `json:"old_status"`
	NewStatus  types.PipelineStatus  `json:"new_status"`
	Stage      string               `json:"stage,omitempty"`
	JobID      string               `json:"job_id,omitempty"`
	Message    string               `json:"message"`
	Timestamp  time.Time            `json:"timestamp"`
}

// CoordinatorConfig represents coordinator configuration
type CoordinatorConfig struct {
	PoolSize      int           `json:"pool_size"`
	DefaultBranch string        `json:"default_branch"`
	Timeout       time.Duration `json:"timeout"`
	MaxPipelines  int           `json:"max_pipelines"`
}

// NewCoordinator creates a new pipeline coordinator
func NewCoordinator(db *PipelineDB, poolMgr types.PoolManagerInterface, netMgr types.NetworkManagerInterface, runnerMgr types.RunnerManagerInterface) (*Coordinator, error) {
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

	return c, nil
}

// loadPipelines loads existing pipelines from database
func (c *Coordinator) loadPipelines() error {
	ids, err := c.db.ListPipelines()
	if err != nil {
		return err
	}

	for _, id := range ids {
		state, err := c.db.LoadPipeline(id)
		if err != nil {
			continue
		}

		// Convert to PipelineState
		ps := &PipelineState{
			ID: id,
		}
		if v, ok := state["status"].(string); ok {
			ps.Status = types.PipelineStatus(v)
		}
		if v, ok := state["platform"].(string); ok {
			ps.Platform = types.RunnerPlatform(v)
		}
		if v, ok := state["repository"].(string); ok {
			ps.Repository = v
		}
		if v, ok := state["branch"].(string); ok {
			ps.Branch = v
		}

		c.pipelines[id] = ps
	}

	return nil
}

// CreatePipeline creates a new pipeline
func (c *Coordinator) CreatePipeline(ctx context.Context, platform types.RunnerPlatform, repo, branch string) (*PipelineState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := generateID()
	now := time.Now()

	ps := &PipelineState{
		ID:        id,
		Platform:  platform,
		Repository: repo,
		Branch:    branch,
		Status:    types.StatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
		VMs:       make([]string, 0),
		Runners:   make([]string, 0),
		Stages:    make([]StageState, 0),
	}

	c.pipelines[id] = ps

	// Save to database
	state := map[string]interface{}{
		"status":    string(ps.Status),
		"platform":  string(ps.Platform),
		"repository": ps.Repository,
		"branch":    ps.Branch,
		"created":   ps.CreatedAt,
	}
	if err := c.db.SavePipeline(id, state); err != nil {
		delete(c.pipelines, id)
		return nil, fmt.Errorf("failed to save pipeline: %w", err)
	}

	c.eventChan <- PipelineEvent{
		PipelineID: id,
		NewStatus:  types.StatusCreated,
		Message:    "Pipeline created",
		Timestamp:  now,
	}

	return ps, nil
}

// StartPipeline starts pipeline execution
func (c *Coordinator) StartPipeline(ctx context.Context, id string, runners int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ps, ok := c.pipelines[id]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	if ps.Status != types.StatusCreated {
		return fmt.Errorf("pipeline already started: %s", id)
	}

	ps.Status = types.StatusRunning
	ps.StartTime = time.Now()
	ps.UpdatedAt = time.Now()

	// Allocate network
	networkID, err := c.networkManager.CreateNetwork(&types.NetworkConfig{
		VLAN:     1000 + len(c.pipelines),
		CIDR:     "10.100.0.0/24",
		Gateway:  "10.100.0.1",
		Isolated: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}
	ps.NetworkID = networkID

	// Allocate VMs
	for i := 0; i < runners; i++ {
		vm, err := c.poolManager.AllocateVM("default")
		if err != nil {
			// Cleanup
			c.cleanup(ps)
			return fmt.Errorf("failed to allocate VM: %w", err)
		}
		ps.VMs = append(ps.VMs, vm.ID)
	}

	// Update database
	state := map[string]interface{}{
		"status":    string(ps.Status),
		"network_id": ps.NetworkID,
		"vms":       ps.VMs,
		"started":   ps.StartTime,
	}
	if err := c.db.SavePipeline(id, state); err != nil {
		c.cleanup(ps)
		return fmt.Errorf("failed to save pipeline: %w", err)
	}

	c.eventChan <- PipelineEvent{
		PipelineID: id,
		OldStatus:  types.StatusCreated,
		NewStatus:  types.StatusRunning,
		Message:    "Pipeline started",
		Timestamp:  time.Now(),
	}

	return nil
}

// StopPipeline stops pipeline execution
func (c *Coordinator) StopPipeline(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ps, ok := c.pipelines[id]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	ps.Status = types.StatusCancelled
	ps.EndTime = &[]time.Time{time.Now()}[0]
	ps.UpdatedAt = time.Now()

	// Cleanup resources
	c.cleanup(ps)

	// Update database
	state := map[string]interface{}{
		"status":  string(ps.Status),
		"ended":   ps.EndTime,
	}
	if err := c.db.SavePipeline(id, state); err != nil {
		return fmt.Errorf("failed to save pipeline: %w", err)
	}

	c.eventChan <- PipelineEvent{
		PipelineID: id,
		OldStatus:  types.StatusRunning,
		NewStatus:  types.StatusCancelled,
		Message:    "Pipeline stopped",
		Timestamp:  time.Now(),
	}

	return nil
}

// cleanup releases all resources for a pipeline
func (c *Coordinator) cleanup(ps *PipelineState) {
	// Release VMs
	for _, vmID := range ps.VMs {
		c.poolManager.ReleaseVM(vmID)
	}

	// Destroy network
	if ps.NetworkID != "" {
		c.networkManager.DestroyNetwork(ps.NetworkID)
	}
}

// GetPipeline returns pipeline state
func (c *Coordinator) GetPipeline(id string) (*PipelineState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ps, ok := c.pipelines[id]
	if !ok {
		return nil, fmt.Errorf("pipeline not found: %s", id)
	}

	return ps, nil
}

// ListPipelines returns all pipelines
func (c *Coordinator) ListPipelines() []*PipelineState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*PipelineState, 0, len(c.pipelines))
	for _, ps := range c.pipelines {
		result = append(result, ps)
	}

	return result
}

// Events returns the event channel
func (c *Coordinator) Events() <-chan PipelineEvent {
	return c.eventChan
}

func generateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}