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
	db             types.PipelineDB
	poolManager    types.PoolManagerInterface
	networkManager types.NetworkManagerInterface
	runnerManager  types.RunnerManagerInterface
	pipelines      map[string]*PipelineState
	mu             sync.RWMutex
	eventChan      chan PipelineEvent
}

// PipelineState represents the in-memory state of a pipeline (coordinator-level)
type PipelineState struct {
	ID           string               `json:"id"`
	Platform     types.RunnerPlatform `json:"platform"`
	Repository   string               `json:"repository"`
	Branch       string               `json:"branch"`
	CommitSHA    string               `json:"commit_sha"`
	CommitMsg    string               `json:"commit_message"`
	Author       string               `json:"author"`
	Status       types.PipelineStatus `json:"status"`
	NetworkID    string               `json:"network_id"`
	VMs          []string             `json:"vms"`
	Runners      []string             `json:"runners"`
	StartTime    time.Time            `json:"start_time"`
	EndTime      *time.Time           `json:"end_time,omitempty"`
	Duration     int64                `json:"duration_seconds"`
	CurrentStage string               `json:"current_stage"`
	Stages       []StageState         `json:"stages"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
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
	Status    types.PipelineStatus `json:"status"`
	RunnerID  string               `json:"runner_id"`
	StartTime *time.Time           `json:"start_time,omitempty"`
	EndTime   *time.Time           `json:"end_time,omitempty"`
	Duration  int64                `json:"duration_seconds"`
	Log       []string             `json:"log,omitempty"`
}

// PipelineEvent represents a pipeline state change event
type PipelineEvent struct {
	PipelineID string               `json:"pipeline_id"`
	OldStatus  types.PipelineStatus `json:"old_status"`
	NewStatus  types.PipelineStatus `json:"new_status"`
	Message    string               `json:"message"`
	Timestamp  time.Time            `json:"timestamp"`
}

// NewCoordinator creates a new coordinator
func NewCoordinator(db types.PipelineDB, poolMgr types.PoolManagerInterface, netMgr types.NetworkManagerInterface, runnerMgr types.RunnerManagerInterface) (*Coordinator, error) {
	c := &Coordinator{
		db:             db,
		poolManager:    poolMgr,
		networkManager: netMgr,
		runnerManager:  runnerMgr,
		pipelines:      make(map[string]*PipelineState),
		eventChan:      make(chan PipelineEvent, 1000),
	}

	if err := c.loadPipelines(); err != nil {
		return nil, fmt.Errorf("failed to load pipelines: %w", err)
	}

	return c, nil
}

// loadPipelines loads existing pipelines from database
func (c *Coordinator) loadPipelines() error {
	ctx := context.Background()
	dbPipelines, err := c.db.ListPipelines(ctx, 100, 0)
	if err != nil {
		return err
	}

	for _, p := range dbPipelines {
		c.pipelines[p.ID] = pipelineStateToCoordinatorState(p)
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
		ID:         id,
		Platform:   platform,
		Repository: repo,
		Branch:     branch,
		Status:     types.PipelineStatusCreating,
		CreatedAt:  now,
		UpdatedAt:  now,
		VMs:        make([]string, 0),
		Runners:    make([]string, 0),
		Stages:     make([]StageState, 0),
	}

	c.pipelines[id] = ps

	// Save to database using the typed interface
	dbState := coordinatorStateToPipelineState(ps)
	if err := c.db.SavePipeline(ctx, dbState); err != nil {
		delete(c.pipelines, id)
		return nil, fmt.Errorf("failed to save pipeline: %w", err)
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

	if ps.Status != types.PipelineStatusCreating {
		return fmt.Errorf("pipeline already started")
	}

	ps.Status = types.PipelineStatusRunning
	ps.StartTime = time.Now()
	ps.UpdatedAt = time.Now()

	if err := c.db.UpdatePipelineStatus(ctx, id, types.PipelineStatusRunning); err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}

	return nil
}

// GetPipeline returns a pipeline by ID
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

	pipelines := make([]*PipelineState, 0, len(c.pipelines))
	for _, ps := range c.pipelines {
		pipelines = append(pipelines, ps)
	}

	return pipelines
}

// CancelPipeline cancels a pipeline
func (c *Coordinator) CancelPipeline(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ps, ok := c.pipelines[id]
	if !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	ps.Status = types.PipelineStatusCancelled
	now := time.Now()
	ps.EndTime = &now
	ps.UpdatedAt = now

	if err := c.db.UpdatePipelineStatus(ctx, id, types.PipelineStatusCancelled); err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}

	return nil
}

// DeletePipeline deletes a pipeline
func (c *Coordinator) DeletePipeline(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.pipelines[id]; !ok {
		return fmt.Errorf("pipeline not found: %s", id)
	}

	delete(c.pipelines, id)
	return c.db.DeletePipeline(context.Background(), id)
}

// Subscribe subscribes to pipeline events
func (c *Coordinator) Subscribe() chan PipelineEvent {
	return c.eventChan
}

// Conversion helpers between coordinator PipelineState and types.PipelineState
func coordinatorStateToPipelineState(ps *PipelineState) *types.PipelineState {
	if ps == nil {
		return nil
	}
	return &types.PipelineState{
		ID:         ps.ID,
		Platform:   ps.Platform,
		Repository: ps.Repository,
		Branch:     ps.Branch,
		CommitSHA:  ps.CommitSHA,
		CommitMsg:  ps.CommitMsg,
		Author:     ps.Author,
		Status:     ps.Status,
		NetworkID:  ps.NetworkID,
		StartTime:  ps.StartTime,
		EndTime:    ps.EndTime,
		Duration:   ps.Duration,
		CreatedAt:  ps.CreatedAt,
		UpdatedAt:  ps.UpdatedAt,
	}
}

func pipelineStateToCoordinatorState(ps *types.PipelineState) *PipelineState {
	if ps == nil {
		return nil
	}
	return &PipelineState{
		ID:         ps.ID,
		Platform:   ps.Platform,
		Repository: ps.Repository,
		Branch:     ps.Branch,
		CommitSHA:  ps.CommitSHA,
		CommitMsg:  ps.CommitMsg,
		Author:     ps.Author,
		Status:     ps.Status,
		NetworkID:  ps.NetworkID,
		StartTime:  ps.StartTime,
		EndTime:    ps.EndTime,
		Duration:   ps.Duration,
		CreatedAt:  ps.CreatedAt,
		UpdatedAt:  ps.UpdatedAt,
		VMs:        make([]string, 0),
		Runners:    make([]string, 0),
		Stages:     make([]StageState, 0),
	}
}

func generateID() string {
	return fmt.Sprintf("pipeline-%s-%d", randomString(8), time.Now().UnixNano())
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}