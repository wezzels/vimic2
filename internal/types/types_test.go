package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPipelineStatus_Constants(t *testing.T) {
	statuses := []PipelineStatus{
		PipelineStatusCreating,
		PipelineStatusRunning,
		PipelineStatusSuccess,
		PipelineStatusFailed,
		PipelineStatusCancelled,
	}

	for _, s := range statuses {
		if s == "" {
			t.Error("expected non-empty status")
		}
	}
}

func TestRunnerStatus_Constants(t *testing.T) {
	statuses := []RunnerStatus{
		RunnerStatusPending,
		RunnerStatusOnline,
		RunnerStatusOffline,
		RunnerStatusBusy,
		RunnerStatusCreating,
		RunnerStatusDestroyed,
	}

	for _, s := range statuses {
		if s == "" {
			t.Error("expected non-empty status")
		}
	}
}

func TestRunnerPlatform_Constants(t *testing.T) {
	platforms := []RunnerPlatform{
		PlatformGitLab,
		PlatformGitHub,
		PlatformJenkins,
		PlatformCircleCI,
		PlatformDrone,
		PlatformLocal,
	}

	for _, p := range platforms {
		if p == "" {
			t.Error("expected non-empty platform")
		}
	}
}

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
	if !config.Isolated {
		t.Error("expected isolated to be true")
	}
}

func TestNetworkState_Create(t *testing.T) {
	state := &NetworkState{
		ID:         "network-1",
		VLANID:     100,
		CIDR:       "10.100.1.0/24",
		BridgeName: "br100",
	}

	if state.ID != "network-1" {
		t.Errorf("expected network-1, got %s", state.ID)
	}
	if state.VLANID != 100 {
		t.Errorf("expected VLAN 100, got %d", state.VLANID)
	}
	if state.BridgeName != "br100" {
		t.Errorf("expected br100, got %s", state.BridgeName)
	}
}

func TestRunnerState_Create(t *testing.T) {
	now := time.Now()
	runner := &RunnerState{
		ID:         "runner-1",
		Platform:   PlatformGitLab,
		Status:     RunnerStatusOnline,
		Name:       "gitlab-runner-1",
		Labels:     []string{"docker", "linux"},
		PipelineID: "pipeline-1",
		VMID:       "vm-1",
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

func TestRunnerState_JSON(t *testing.T) {
	runner := &RunnerState{
		ID:         "runner-1",
		Platform:   PlatformGitHub,
		Status:     RunnerStatusBusy,
		Name:       "github-runner-1",
		Labels:     []string{"ubuntu", "docker"},
		PipelineID: "pipeline-1",
		VMID:       "vm-1",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(runner)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var runner2 RunnerState
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

func TestPipelineState_Create(t *testing.T) {
	now := time.Now()
	pipeline := &PipelineState{
		ID:         "pipeline-1",
		Platform:   PlatformGitLab,
		Repository: "https://gitlab.example.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		CommitMsg:  "Test commit",
		Author:     "test@example.com",
		Status:     PipelineStatusRunning,
		NetworkID:  "network-1",
		StartTime:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
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
}

func TestPipelineState_JSON(t *testing.T) {
	pipeline := &PipelineState{
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

	var pipeline2 PipelineState
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

func TestVMState_Create(t *testing.T) {
	now := time.Now()
	vm := &VMState{
		ID:          "vm-1",
		Name:        "test-vm",
		Status:      "running",
		IPAddress:   "10.100.1.10",
		MACAddress:  "02:42:0a:64:01:0a",
		PoolName:    "default",
		Template:    "ubuntu-22.04",
		CreatedAt:   now,
		DestroyedAt: nil,
	}

	if vm.ID != "vm-1" {
		t.Errorf("expected vm-1, got %s", vm.ID)
	}
	if vm.Status != "running" {
		t.Errorf("expected running, got %s", vm.Status)
	}
	if vm.IPAddress != "10.100.1.10" {
		t.Errorf("expected 10.100.1.10, got %s", vm.IPAddress)
	}
}

func TestPoolState_Create(t *testing.T) {
	pool := &PoolState{
		Name:         "ubuntu-pool",
		TemplatePath: "/templates/ubuntu-22.04.qcow2",
		Capacity:     10,
		Available:    7,
		Busy:         3,
	}

	if pool.Name != "ubuntu-pool" {
		t.Errorf("expected ubuntu-pool, got %s", pool.Name)
	}
	if pool.Available != 7 {
		t.Errorf("expected 7 available, got %d", pool.Available)
	}
	if pool.Busy != 3 {
		t.Errorf("expected 3 busy, got %d", pool.Busy)
	}
}

func TestArtifact_Create(t *testing.T) {
	a := &Artifact{
		ID:         "artifact-1",
		PipelineID: "pipeline-1",
		Name:       "build.tar.gz",
		Path:       "/artifacts/pipeline-1/build.tar.gz",
		Size:       1024000,
		Checksum:   "sha256:abc123",
		TTL:        30,
		CreatedAt:  time.Now(),
	}

	if a.ID != "artifact-1" {
		t.Errorf("expected artifact-1, got %s", a.ID)
	}
	if a.Size != 1024000 {
		t.Errorf("expected size 1024000, got %d", a.Size)
	}
}