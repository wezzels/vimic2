// Package runner provides runner manager tests
package runner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// TestRunnerInfo tests runner info structure
func TestRunnerInfo_Create(t *testing.T) {
	now := time.Now()
	runner := &RunnerInfo{
		ID:            "runner-1",
		VMID:          "vm-1",
		PoolName:      "pool-1",
		PipelineID:    "pipeline-1",
		Platform:      types.PlatformGitLab,
		PlatformID:    "gitlab-runner-123",
		Name:          "gitlab-runner-1",
		Labels:        []string{"docker", "linux"},
		Status:        types.RunnerStatusOnline,
		IPAddress:     "10.100.1.10",
		CurrentJob:    "job-1",
		HealthStatus:  "healthy",
		LastHeartbeat: &now,
		CreatedAt:     now,
	}

	if runner.ID != "runner-1" {
		t.Errorf("expected runner-1, got %s", runner.ID)
	}
	if runner.Platform != types.PlatformGitLab {
		t.Errorf("expected gitlab platform, got %s", runner.Platform)
	}
	if runner.Status != types.RunnerStatusOnline {
		t.Errorf("expected online status, got %s", runner.Status)
	}
	if len(runner.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(runner.Labels))
	}
}

// TestRunnerInfo_JSON tests runner info JSON marshaling
func TestRunnerInfo_JSON(t *testing.T) {
	now := time.Now()
	runner := &RunnerInfo{
		ID:         "runner-1",
		VMID:       "vm-1",
		PoolName:   "pool-1",
		PipelineID: "pipeline-1",
		Platform:   types.PlatformGitHub,
		Name:       "github-runner-1",
		Labels:     []string{"ubuntu", "docker"},
		Status:     types.RunnerStatusBusy,
		IPAddress:  "10.100.1.10",
		CreatedAt:  now,
	}

	data, err := json.Marshal(runner)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var runner2 RunnerInfo
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

// TestRunnerManagerConfig tests runner manager configuration
func TestRunnerManagerConfig_Create(t *testing.T) {
	config := &RunnerManagerConfig{
		GitLab: &GitLabConfig{
			URL:      "https://gitlab.example.com",
			Token:    "glrt-xxx",
			Executor: "docker",
		},
		GitHub: &GitHubConfig{
			Repo:   "owner/repo",
			Token:  "ghp_xxx",
			Labels: []string{"ubuntu", "docker"},
		},
		Jenkins: &JenkinsConfig{
			URL:   "https://jenkins.example.com",
			User:  "jenkins",
			Token: "jenkins-token",
		},
		CircleCI: &CircleCIConfig{
			Org:   "myorg",
			Token: "cci_xxx",
		},
		Drone: &DroneConfig{
			URL:   "https://drone.example.com",
			Token: "drone-token",
		},
	}

	if config.GitLab == nil {
		t.Error("GitLab config should not be nil")
	}
	if config.GitHub == nil {
		t.Error("GitHub config should not be nil")
	}
	if config.Jenkins == nil {
		t.Error("Jenkins config should not be nil")
	}
}

// TestGitLabConfig tests GitLab runner configuration
func TestGitLabConfig_Create(t *testing.T) {
	config := &GitLabConfig{
		URL:      "https://gitlab.example.com",
		Token:    "glrt-xxx",
		Executor: "docker",
	}

	if config.URL != "https://gitlab.example.com" {
		t.Errorf("expected gitlab URL, got %s", config.URL)
	}
	if config.Token != "glrt-xxx" {
		t.Errorf("expected token, got %s", config.Token)
	}
	if config.Executor != "docker" {
		t.Errorf("expected docker executor, got %s", config.Executor)
	}
}

// TestGitHubConfig tests GitHub runner configuration
func TestGitHubConfig_Create(t *testing.T) {
	config := &GitHubConfig{
		Repo:   "owner/repo",
		Token:  "ghp_xxx",
		Labels: []string{"ubuntu", "docker"},
	}

	if config.Repo != "owner/repo" {
		t.Errorf("expected owner/repo, got %s", config.Repo)
	}
	if config.Token != "ghp_xxx" {
		t.Errorf("expected token, got %s", config.Token)
	}
	if len(config.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(config.Labels))
	}
}

// TestJenkinsConfig tests Jenkins runner configuration
func TestJenkinsConfig_Create(t *testing.T) {
	config := &JenkinsConfig{
		URL:   "https://jenkins.example.com",
		User:  "jenkins",
		Token: "jenkins-token",
	}

	if config.URL != "https://jenkins.example.com" {
		t.Errorf("expected jenkins URL, got %s", config.URL)
	}
	if config.User != "jenkins" {
		t.Errorf("expected jenkins user, got %s", config.User)
	}
}

// TestCircleCIConfig tests CircleCI runner configuration
func TestCircleCIConfig_Create(t *testing.T) {
	config := &CircleCIConfig{
		Org:   "myorg",
		Token: "cci_xxx",
	}

	if config.Org != "myorg" {
		t.Errorf("expected org myorg, got %s", config.Org)
	}
	if config.Token != "cci_xxx" {
		t.Errorf("expected token, got %s", config.Token)
	}
}

// TestDroneConfig tests Drone runner configuration
func TestDroneConfig_Create(t *testing.T) {
	config := &DroneConfig{
		URL:   "https://drone.example.com",
		Token: "drone-token",
	}

	if config.URL != "https://drone.example.com" {
		t.Errorf("expected drone URL, got %s", config.URL)
	}
	if config.Token != "drone-token" {
		t.Errorf("expected token, got %s", config.Token)
	}
}

// TestRunnerStatus_Transitions tests runner status transitions
func TestRunnerStatus_Transitions(t *testing.T) {
	validTransitions := []struct {
		from types.RunnerStatus
		to   types.RunnerStatus
	}{
		{types.RunnerStatusCreating, types.RunnerStatusOnline},
		{types.RunnerStatusCreating, types.RunnerStatusError},
		{types.RunnerStatusOnline, types.RunnerStatusBusy},
		{types.RunnerStatusBusy, types.RunnerStatusOnline},
		{types.RunnerStatusOnline, types.RunnerStatusOffline},
		{types.RunnerStatusOffline, types.RunnerStatusOnline},
		{types.RunnerStatusOnline, types.RunnerStatusDestroyed},
	}

	for _, tt := range validTransitions {
		runner := &RunnerInfo{Status: tt.from}
		runner.Status = tt.to

		if runner.Status != tt.to {
			t.Errorf("failed to transition from %s to %s", tt.from, tt.to)
		}
	}
}

// TestRunnerInfo_Labels tests runner labels
func TestRunnerInfo_Labels(t *testing.T) {
	runner := &RunnerInfo{
		ID:     "runner-1",
		Labels: []string{"docker", "linux", "ubuntu"},
	}

	if len(runner.Labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(runner.Labels))
	}

	expected := map[string]bool{"docker": true, "linux": true, "ubuntu": true}
	for _, label := range runner.Labels {
		if !expected[label] {
			t.Errorf("unexpected label: %s", label)
		}
	}
}

// TestRunnerInfo_HealthStatus tests runner health status
func TestRunnerInfo_HealthStatus(t *testing.T) {
	runner := &RunnerInfo{
		ID:           "runner-1",
		HealthStatus: "healthy",
	}

	if runner.HealthStatus != "healthy" {
		t.Errorf("expected healthy status, got %s", runner.HealthStatus)
	}

	// Test updating health status
	runner.HealthStatus = "unhealthy"
	if runner.HealthStatus != "unhealthy" {
		t.Errorf("expected unhealthy status, got %s", runner.HealthStatus)
	}
}

// TestRunnerInfo_CurrentJob tests runner job tracking
func TestRunnerInfo_CurrentJob(t *testing.T) {
	runner := &RunnerInfo{
		ID:         "runner-1",
		Status:     types.RunnerStatusBusy,
		CurrentJob: "job-123",
	}

	if runner.CurrentJob != "job-123" {
		t.Errorf("expected job-123, got %s", runner.CurrentJob)
	}

	// Clear job when done
	runner.CurrentJob = ""
	runner.Status = types.RunnerStatusOnline

	if runner.CurrentJob != "" {
		t.Errorf("expected empty job, got %s", runner.CurrentJob)
	}
	if runner.Status != types.RunnerStatusOnline {
		t.Errorf("expected online status, got %s", runner.Status)
	}
}

// TestRunnerInfo_LastHeartbeat tests runner heartbeat tracking
func TestRunnerInfo_LastHeartbeat(t *testing.T) {
	now := time.Now()
	runner := &RunnerInfo{
		ID:            "runner-1",
		LastHeartbeat: &now,
	}

	if runner.LastHeartbeat == nil {
		t.Error("LastHeartbeat should not be nil")
	}

	// Check heartbeat freshness
	heartbeatAge := time.Since(*runner.LastHeartbeat)
	if heartbeatAge > time.Minute {
		t.Errorf("heartbeat too old: %v", heartbeatAge)
	}
}

// TestRunnerManager_StructFields tests runner manager struct fields
func TestRunnerManager_StructFields(t *testing.T) {
	rm := &RunnerManager{
		runners: make(map[string]*RunnerInfo),
	}

	if rm.runners == nil {
		t.Error("runners map should not be nil")
	}

	// Add a runner
	rm.runners["runner-1"] = &RunnerInfo{
		ID:     "runner-1",
		Status: types.RunnerStatusOnline,
	}

	if len(rm.runners) != 1 {
		t.Errorf("expected 1 runner, got %d", len(rm.runners))
	}
}

// TestRunnerInfo_DestroyedAt tests runner destruction
func TestRunnerInfo_DestroyedAt(t *testing.T) {
	now := time.Now()
	destroyed := now.Add(1 * time.Hour)

	runner := &RunnerInfo{
		ID:          "runner-1",
		Status:      types.RunnerStatusDestroyed,
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

// TestRunnerInfo_JSON_Fields tests JSON field names
func TestRunnerInfo_JSON_Fields(t *testing.T) {
	runner := &RunnerInfo{
		ID:         "runner-1",
		Platform:   types.PlatformGitLab,
		Name:       "test-runner",
		Labels:     []string{"docker"},
		Status:     types.RunnerStatusOnline,
		IPAddress:  "10.100.1.10",
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(runner)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Check JSON field names
	jsonStr := string(data)
	expectedFields := []string{`"id"`, `"platform"`, `"name"`, `"labels"`, `"status"`, `"ip_address"`}
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("expected field %s in JSON", field)
		}
	}
}

// TestGitLabConfig_JSON tests GitLab config JSON
func TestGitLabConfig_JSON(t *testing.T) {
	config := &GitLabConfig{
		URL:      "https://gitlab.example.com",
		Token:    "glrt-xxx",
		Executor: "docker",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var config2 GitLabConfig
	if err := json.Unmarshal(data, &config2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config2.URL != config.URL {
		t.Errorf("expected URL %s, got %s", config.URL, config2.URL)
	}
}

// TestGitHubConfig_JSON tests GitHub config JSON
func TestGitHubConfig_JSON(t *testing.T) {
	config := &GitHubConfig{
		Repo:   "owner/repo",
		Token:  "ghp-xxx",
		Labels: []string{"docker"},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var config2 GitHubConfig
	if err := json.Unmarshal(data, &config2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config2.Repo != config.Repo {
		t.Errorf("expected repo %s, got %s", config.Repo, config2.Repo)
	}
}

// TestJenkinsConfig_JSON tests Jenkins config JSON
func TestJenkinsConfig_JSON(t *testing.T) {
	config := &JenkinsConfig{
		URL:   "https://jenkins.example.com",
		User:  "jenkins",
		Token: "token",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var config2 JenkinsConfig
	if err := json.Unmarshal(data, &config2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if config2.URL != config.URL {
		t.Errorf("expected URL %s, got %s", config.URL, config2.URL)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}