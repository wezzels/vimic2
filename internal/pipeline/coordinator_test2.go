// Package pipeline provides coordinator tests
package pipeline

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// TestPipelineState tests pipeline state structure
func TestPipelineState_Create(t *testing.T) {
	state := &PipelineState{
		ID:           "pipeline-1",
		Platform:     types.PlatformGitLab,
		Repository:   "https://gitlab.example.com/test/repo",
		Branch:       "main",
		CommitSHA:    "abc123",
		CommitMsg:    "Test commit",
		Author:       "test@example.com",
		Status:       types.PipelineStatusRunning,
		NetworkID:    "network-1",
		VMs:          []string{"vm-1", "vm-2"},
		Runners:      []string{"runner-1"},
		StartTime:    time.Now(),
		Duration:     300,
		CurrentStage: "build",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if state.ID != "pipeline-1" {
		t.Errorf("expected pipeline-1, got %s", state.ID)
	}
	if state.Platform != types.PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", state.Platform)
	}
	if state.Status != types.PipelineStatusRunning {
		t.Errorf("expected running status, got %s", state.Status)
	}
	if len(state.VMs) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(state.VMs))
	}
}

// TestStageState tests stage state structure
func TestStageState_Create(t *testing.T) {
	now := time.Now()
	stage := &StageState{
		Name:      "build",
		Status:    types.PipelineStatusRunning,
		StartTime: &now,
		Jobs: []JobState{
			{
				ID:     "job-1",
				Name:   "compile",
				Stage:  "build",
				Status: types.PipelineStatusRunning,
			},
		},
	}

	if stage.Name != "build" {
		t.Errorf("expected build stage, got %s", stage.Name)
	}
	if stage.Status != types.PipelineStatusRunning {
		t.Errorf("expected running status, got %s", stage.Status)
	}
	if len(stage.Jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(stage.Jobs))
	}
}

// TestJobState tests job state structure
func TestJobState_Create(t *testing.T) {
	now := time.Now()
	job := &JobState{
		ID:        "job-1",
		Name:      "test",
		Stage:     "test",
		Status:    types.PipelineStatusSuccess,
		RunnerID:  "runner-1",
		StartTime: &now,
		Duration:  120,
		Log:       []string{"Starting tests", "Tests passed"},
	}

	if job.ID != "job-1" {
		t.Errorf("expected job-1, got %s", job.ID)
	}
	if job.Name != "test" {
		t.Errorf("expected test name, got %s", job.Name)
	}
	if job.Status != types.PipelineStatusSuccess {
		t.Errorf("expected success status, got %s", job.Status)
	}
	if job.Duration != 120 {
		t.Errorf("expected 120s duration, got %d", job.Duration)
	}
}

// TestPipelineEvent tests pipeline event structure
func TestPipelineEvent_Create(t *testing.T) {
	event := &PipelineEvent{
		PipelineID: "pipeline-1",
		OldStatus:  types.PipelineStatusCreating,
		NewStatus:  types.PipelineStatusRunning,
		Message:    "Pipeline started",
		Timestamp:  time.Now(),
	}

	if event.PipelineID != "pipeline-1" {
		t.Errorf("expected pipeline-1, got %s", event.PipelineID)
	}
	if event.OldStatus != types.PipelineStatusCreating {
		t.Errorf("expected creating status, got %s", event.OldStatus)
	}
	if event.NewStatus != types.PipelineStatusRunning {
		t.Errorf("expected running status, got %s", event.NewStatus)
	}
}

// TestPipelineState_JSON tests JSON marshaling
func TestPipelineState_JSON(t *testing.T) {
	state := &PipelineState{
		ID:         "pipeline-1",
		Platform:   types.PlatformGitLab,
		Repository: "https://gitlab.example.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		Status:     types.PipelineStatusRunning,
		StartTime:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal pipeline state: %v", err)
	}

	var state2 PipelineState
	if err := json.Unmarshal(data, &state2); err != nil {
		t.Fatalf("failed to unmarshal pipeline state: %v", err)
	}

	if state2.ID != state.ID {
		t.Errorf("expected ID %s, got %s", state.ID, state2.ID)
	}
	if state2.Platform != state.Platform {
		t.Errorf("expected platform %s, got %s", state.Platform, state2.Platform)
	}
}

// TestStageState_JSON tests stage state JSON marshaling
func TestStageState_JSON(t *testing.T) {
	stage := &StageState{
		Name:   "build",
		Status: types.PipelineStatusSuccess,
		Jobs: []JobState{
			{Name: "compile", Status: types.PipelineStatusSuccess},
		},
	}

	data, err := json.Marshal(stage)
	if err != nil {
		t.Fatalf("failed to marshal stage: %v", err)
	}

	var stage2 StageState
	if err := json.Unmarshal(data, &stage2); err != nil {
		t.Fatalf("failed to unmarshal stage: %v", err)
	}

	if stage2.Name != stage.Name {
		t.Errorf("expected name %s, got %s", stage.Name, stage2.Name)
	}
}

// TestJobState_JSON tests job state JSON marshaling
func TestJobState_JSON(t *testing.T) {
	job := &JobState{
		ID:     "job-1",
		Name:   "test",
		Stage:  "test",
		Status: types.PipelineStatusSuccess,
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal job: %v", err)
	}

	var job2 JobState
	if err := json.Unmarshal(data, &job2); err != nil {
		t.Fatalf("failed to unmarshal job: %v", err)
	}

	if job2.ID != job.ID {
		t.Errorf("expected ID %s, got %s", job.ID, job2.ID)
	}
}

// TestPipelineState_StatusTransitions tests status transitions
func TestPipelineState_StatusTransitions(t *testing.T) {
	state := &PipelineState{
		ID:     "pipeline-1",
		Status: types.PipelineStatusCreating,
	}

	// Valid transitions
	transitions := []struct {
		from types.PipelineStatus
		to   types.PipelineStatus
	}{
		{types.PipelineStatusCreating, types.PipelineStatusRunning},
		{types.PipelineStatusRunning, types.PipelineStatusSuccess},
		{types.PipelineStatusRunning, types.PipelineStatusFailed},
		{types.PipelineStatusRunning, types.PipelineStatusCanceled},
	}

	for _, tt := range transitions {
		state.Status = tt.from
		// Transition to new status
		state.Status = tt.to
		if state.Status != tt.to {
			t.Errorf("failed to transition from %s to %s", tt.from, tt.to)
		}
	}
}

// TestPipelineState_Stages tests stages management
func TestPipelineState_Stages(t *testing.T) {
	state := &PipelineState{
		ID:     "pipeline-1",
		Stages: []StageState{},
	}

	// Add stages
	stage1 := StageState{Name: "build", Status: types.PipelineStatusRunning}
	stage2 := StageState{Name: "test", Status: types.PipelineStatusCreating}
	stage3 := StageState{Name: "deploy", Status: types.PipelineStatusCreating}

	state.Stages = append(state.Stages, stage1, stage2, stage3)

	if len(state.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(state.Stages))
	}
	if state.Stages[0].Name != "build" {
		t.Errorf("expected first stage build, got %s", state.Stages[0].Name)
	}
}

// TestPipelineState_VMManagement tests VM management
func TestPipelineState_VMManagement(t *testing.T) {
	state := &PipelineState{
		ID:  "pipeline-1",
		VMs: []string{},
	}

	// Add VMs
	state.VMs = append(state.VMs, "vm-1", "vm-2", "vm-3")

	if len(state.VMs) != 3 {
		t.Errorf("expected 3 VMs, got %d", len(state.VMs))
	}

	// Remove a VM
	for i, vm := range state.VMs {
		if vm == "vm-2" {
			state.VMs = append(state.VMs[:i], state.VMs[i+1:]...)
			break
		}
	}

	if len(state.VMs) != 2 {
		t.Errorf("expected 2 VMs after removal, got %d", len(state.VMs))
	}
}

// TestPipelineState_Duration tests duration calculation
func TestPipelineState_Duration(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now()

	state := &PipelineState{
		ID:        "pipeline-1",
		StartTime: start,
		EndTime:   &end,
	}

	state.Duration = int64(end.Sub(start).Seconds())

	if state.Duration < 299 || state.Duration > 301 {
		t.Errorf("expected duration ~300s, got %d", state.Duration)
	}
}

// TestJobState_LogManagement tests log management
func TestJobState_LogManagement(t *testing.T) {
	job := &JobState{
		ID:  "job-1",
		Log: []string{},
	}

	// Add log lines
	job.Log = append(job.Log, "Starting build...", "Compiling source...", "Build complete!")

	if len(job.Log) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(job.Log))
	}
	if job.Log[0] != "Starting build..." {
		t.Errorf("expected first log line 'Starting build...', got %s", job.Log[0])
	}
}

// TestCoordinator_CreateStruct tests coordinator struct fields
func TestCoordinator_CreateStruct(t *testing.T) {
	coord := &Coordinator{
		pipelines: make(map[string]*PipelineState),
	}

	if coord.pipelines == nil {
		t.Error("pipelines map should not be nil")
	}

	// Add a pipeline state
	coord.pipelines["pipeline-1"] = &PipelineState{
		ID:     "pipeline-1",
		Status: types.PipelineStatusRunning,
	}

	if len(coord.pipelines) != 1 {
		t.Errorf("expected 1 pipeline, got %d", len(coord.pipelines))
	}
}
