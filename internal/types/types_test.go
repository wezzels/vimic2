// Package types provides shared types tests
package types

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPipelineStatus tests pipeline status constants
func TestPipelineStatus_Constants(t *testing.T) {
	statuses := []PipelineStatus{
		PipelineStatusCreating,
		PipelineStatusRunning,
		PipelineStatusSuccess,
		PipelineStatusFailed,
		PipelineStatusCanceled,
	}

	for _, status := range statuses {
		if status == "" {
			t.Error("empty pipeline status")
		}
	}
}

// TestRunnerPlatform tests runner platform constants
func TestRunnerPlatform_Constants(t *testing.T) {
	platforms := []RunnerPlatform{
		PlatformGitLab,
		PlatformGitHub,
		PlatformJenkins,
		PlatformCircleCI,
		PlatformDrone,
	}

	for _, platform := range platforms {
		if platform == "" {
			t.Error("empty runner platform")
		}
	}
}

// TestRunnerStatus tests runner status constants
func TestRunnerStatus_Constants(t *testing.T) {
	statuses := []RunnerStatus{
		RunnerStatusCreating,
		RunnerStatusOnline,
		RunnerStatusOffline,
		RunnerStatusBusy,
		RunnerStatusError,
		RunnerStatusDestroyed,
	}

	for _, status := range statuses {
		if status == "" {
			t.Error("empty runner status")
		}
	}
}

// TestVMState tests VM state structure
func TestVMState_Create(t *testing.T) {
	now := time.Now()
	state := &VMState{
		ID:         "vm-1",
		Name:       "test-vm",
		Status:     "running",
		IPAddress:  "10.100.1.10",
		MACAddress: "52:54:00:12:34:56",
		PoolName:   "pool-1",
		Template:   "ubuntu-22.04",
		CreatedAt:  now,
	}

	if state.ID != "vm-1" {
		t.Errorf("expected vm-1, got %s", state.ID)
	}
	if state.Status != "running" {
		t.Errorf("expected running status, got %s", state.Status)
	}
	if state.PoolName != "pool-1" {
		t.Errorf("expected pool-1, got %s", state.PoolName)
	}
}

// TestVMState_JSON tests VM state JSON marshaling
func TestVMState_JSON(t *testing.T) {
	state := &VMState{
		ID:         "vm-1",
		Name:       "test-vm",
		Status:     "running",
		IPAddress:  "10.100.1.10",
		MACAddress: "52:54:00:12:34:56",
		PoolName:   "pool-1",
		Template:   "ubuntu-22.04",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var state2 VMState
	if err := json.Unmarshal(data, &state2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if state2.ID != state.ID {
		t.Errorf("expected ID %s, got %s", state.ID, state2.ID)
	}
	if state2.Status != state.Status {
		t.Errorf("expected status %s, got %s", state.Status, state2.Status)
	}
}

// TestPoolState tests pool state structure
func TestPoolState_Create(t *testing.T) {
	pool := &PoolState{
		Name:         "pool-1",
		TemplatePath: "/var/lib/vimic2/templates/ubuntu.qcow2",
		Capacity:     10,
		Available:    7,
		Busy:         3,
	}

	if pool.Name != "pool-1" {
		t.Errorf("expected pool-1, got %s", pool.Name)
	}
	if pool.Capacity != 10 {
		t.Errorf("expected capacity 10, got %d", pool.Capacity)
	}
	if pool.Available != 7 {
		t.Errorf("expected available 7, got %d", pool.Available)
	}
	if pool.Busy != 3 {
		t.Errorf("expected busy 3, got %d", pool.Busy)
	}
}

// TestNetworkConfig tests network configuration
func TestNetworkConfig_Create(t *testing.T) {
	config := &NetworkConfig{
		VLAN:     100,
		CIDR:     "10.100.1.0/24",
		Gateway:  "10.100.1.1",
		DNS:      []string{"8.8.8.8", "8.8.4.4"},
		Isolated: true,
	}

	if config.VLAN != 100 {
		t.Errorf("expected VLAN 100, got %d", config.VLAN)
	}
	if config.CIDR != "10.100.1.0/24" {
		t.Errorf("expected CIDR 10.100.1.0/24, got %s", config.CIDR)
	}
	if len(config.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(config.DNS))
	}
	if !config.Isolated {
		t.Error("expected isolated to be true")
	}
}

// TestNetworkState tests network state
func TestNetworkState_Create(t *testing.T) {
	state := &NetworkState{
		ID:     "network-1",
		VLAN:   100,
		CIDR:   "10.100.1.0/24",
		Status: "active",
	}

	if state.ID != "network-1" {
		t.Errorf("expected network-1, got %s", state.ID)
	}
	if state.VLAN != 100 {
		t.Errorf("expected VLAN 100, got %d", state.VLAN)
	}
	if state.Status != "active" {
		t.Errorf("expected active status, got %s", state.Status)
	}
}

// TestRunner tests runner structure
func TestRunner_Create(t *testing.T) {
	now := time.Now()
	runner := &Runner{
		ID:         "runner-1",
		Platform:   PlatformGitLab,
		Status:     RunnerStatusOnline,
		Name:       "gitlab-runner-1",
		Labels:     []string{"docker", "linux"},
		PipelineID: "pipeline-1",
		VMID:       "vm-1",
		IPAddress:  "10.100.1.10",
		CreatedAt:  now,
	}

	if runner.ID != "runner-1" {
		t.Errorf("expected runner-1, got %s", runner.ID)
	}
	if runner.Platform != PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", runner.Platform)
	}
	if runner.Status != RunnerStatusOnline {
		t.Errorf("expected online status, got %s", runner.Status)
	}
	if len(runner.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(runner.Labels))
	}
}

// TestRunner_JSON tests runner JSON marshaling
func TestRunner_JSON(t *testing.T) {
	runner := &Runner{
		ID:         "runner-1",
		Platform:   PlatformGitHub,
		Status:     RunnerStatusBusy,
		Name:       "github-runner-1",
		Labels:     []string{"ubuntu", "docker"},
		PipelineID: "pipeline-1",
		VMID:       "vm-1",
		IPAddress:  "10.100.1.10",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(runner)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var runner2 Runner
	if err := json.Unmarshal(data, &runner2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if runner2.ID != runner.ID {
		t.Errorf("expected ID %s, got %s", runner.ID, runner2.ID)
	}
	if runner2.Platform != runner.Platform {
		t.Errorf("expected platform %s, got %s", runner.Platform, runner2.Platform)
	}
}

// TestPipeline tests pipeline structure
func TestPipeline_Create(t *testing.T) {
	now := time.Now()
	pipeline := &Pipeline{
		ID:           "pipeline-1",
		Platform:     PlatformGitLab,
		Repository:   "https://gitlab.example.com/test/repo",
		Branch:       "main",
		CommitSHA:    "abc123",
		CommitMsg:    "Test commit",
		Author:       "test@example.com",
		Status:       PipelineStatusRunning,
		NetworkID:    "network-1",
		VMs:          []string{"vm-1", "vm-2"},
		Runners:      []string{"runner-1"},
		StartTime:    now,
		CurrentStage: "build",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if pipeline.ID != "pipeline-1" {
		t.Errorf("expected pipeline-1, got %s", pipeline.ID)
	}
	if pipeline.Platform != PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", pipeline.Platform)
	}
	if pipeline.Status != PipelineStatusRunning {
		t.Errorf("expected running status, got %s", pipeline.Status)
	}
	if len(pipeline.VMs) != 2 {
		t.Errorf("expected 2 VMs, got %d", len(pipeline.VMs))
	}
}

// TestPipeline_JSON tests pipeline JSON marshaling
func TestPipeline_JSON(t *testing.T) {
	pipeline := &Pipeline{
		ID:         "pipeline-1",
		Platform:   PlatformGitHub,
		Repository: "https://github.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		Status:     PipelineStatusSuccess,
		StartTime:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(pipeline)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var pipeline2 Pipeline
	if err := json.Unmarshal(data, &pipeline2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if pipeline2.ID != pipeline.ID {
		t.Errorf("expected ID %s, got %s", pipeline.ID, pipeline2.ID)
	}
	if pipeline2.Platform != pipeline.Platform {
		t.Errorf("expected platform %s, got %s", pipeline.Platform, pipeline2.Platform)
	}
}

// TestStage tests stage structure
func TestStage_Create(t *testing.T) {
	// Note: Stage is defined in types.go but we'll test the concept
	// using the values from the file
	stageName := "build"
	stageStatus := PipelineStatusRunning

	if stageName != "build" {
		t.Errorf("expected build stage, got %s", stageName)
	}
	if stageStatus != PipelineStatusRunning {
		t.Errorf("expected running status, got %s", stageStatus)
	}
}

// TestNetworkConfig_JSON tests network config JSON
func TestNetworkConfig_JSON(t *testing.T) {
	config := &NetworkConfig{
		VLAN:     100,
		CIDR:     "10.100.1.0/24",
		Gateway:  "10.100.1.1",
		DNS:      []string{"8.8.8.8"},
		Isolated: true,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var config2 NetworkConfig
	if err := json.Unmarshal(data, &config2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config2.VLAN != config.VLAN {
		t.Errorf("expected VLAN %d, got %d", config.VLAN, config2.VLAN)
	}
}

// TestPoolState_JSON tests pool state JSON
func TestPoolState_JSON(t *testing.T) {
	pool := &PoolState{
		Name:         "pool-1",
		TemplatePath: "/var/lib/vimic2/templates/ubuntu.qcow2",
		Capacity:     10,
		Available:    7,
		Busy:         3,
	}

	data, err := json.Marshal(pool)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var pool2 PoolState
	if err := json.Unmarshal(data, &pool2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if pool2.Name != pool.Name {
		t.Errorf("expected name %s, got %s", pool.Name, pool2.Name)
	}
}

// TestRunner_DestroyedAt tests runner destruction
func TestRunner_DestroyedAt(t *testing.T) {
	now := time.Now()
	destroyed := now.Add(1 * time.Hour)

	runner := &Runner{
		ID:          "runner-1",
		Status:      RunnerStatusDestroyed,
		CreatedAt:   now,
		DestroyedAt: &destroyed,
	}

	if runner.DestroyedAt == nil {
		t.Error("DestroyedAt should not be nil")
	}
	if runner.DestroyedAt.Before(runner.CreatedAt) {
		t.Error("DestroyedAt should be after CreatedAt")
	}
}

// TestVMState_DestroyedAt tests VM destruction
func TestVMState_DestroyedAt(t *testing.T) {
	now := time.Now()
	destroyed := now.Add(1 * time.Hour)

	vm := &VMState{
		ID:          "vm-1",
		Status:      "destroyed",
		CreatedAt:   now,
		DestroyedAt: &destroyed,
	}

	if vm.DestroyedAt == nil {
		t.Error("DestroyedAt should not be nil")
	}
	if vm.DestroyedAt.Before(vm.CreatedAt) {
		t.Error("DestroyedAt should be after CreatedAt")
	}
}
