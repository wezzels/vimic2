// Package runner provides tests for runner interfaces
package runner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// TestBaseRunner_ID tests BaseRunner ID method
func TestBaseRunner_ID(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		labels:   []string{"docker", "linux"},
	}

	if runner.ID() != "runner-123" {
		t.Errorf("expected runner-123, got %s", runner.ID())
	}
}

// TestBaseRunner_Platform tests BaseRunner Platform method
func TestBaseRunner_Platform(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitHub,
	}

	if runner.Platform() != types.PlatformGitHub {
		t.Errorf("expected PlatformGitHub, got %s", runner.Platform())
	}
}

// TestBaseRunner_Labels tests BaseRunner Labels method
func TestBaseRunner_Labels(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		labels:   []string{"docker", "linux", "golang"},
	}

	labels := runner.Labels()
	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}
	if labels[0] != "docker" {
		t.Errorf("expected docker, got %s", labels[0])
	}
}

// TestBaseRunner_Status tests BaseRunner Status method
func TestBaseRunner_Status(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		status:   types.RunnerStatusOnline,
	}

	if runner.Status() != types.RunnerStatusOnline {
		t.Errorf("expected RunnerStatusOnline, got %s", runner.Status())
	}
}

// TestBaseRunner_Health tests BaseRunner Health method
func TestBaseRunner_Health(t *testing.T) {
	now := time.Now()
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		health: &HealthStatus{
			Healthy:       true,
			LastCheck:     now,
			ActiveJobs:    2,
			TotalJobs:     10,
			UptimeSeconds: 3600,
		},
	}

	health := runner.Health()
	if !health.Healthy {
		t.Error("expected healthy runner")
	}
	if health.ActiveJobs != 2 {
		t.Errorf("expected 2 active jobs, got %d", health.ActiveJobs)
	}
}

// TestBaseRunner_SetStatus tests BaseRunner SetStatus method
func TestBaseRunner_SetStatus(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		status:   types.RunnerStatusOffline,
	}

	runner.SetStatus(types.RunnerStatusOnline)

	if runner.Status() != types.RunnerStatusOnline {
		t.Errorf("expected RunnerStatusOnline, got %s", runner.Status())
	}
}

// TestBaseRunner_SetHealth tests BaseRunner SetHealth method
func TestBaseRunner_SetHealth(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
	}

	newHealth := &HealthStatus{
		Healthy:       true,
		LastCheck:     time.Now(),
		ActiveJobs:    5,
		TotalJobs:     100,
		UptimeSeconds: 7200,
	}

	runner.SetHealth(newHealth)

	health := runner.Health()
	if health.TotalJobs != 100 {
		t.Errorf("expected 100 total jobs, got %d", health.TotalJobs)
	}
}

// TestHealthStatus_JSON tests HealthStatus JSON marshaling
func TestHealthStatus_JSON(t *testing.T) {
	now := time.Now()
	health := &HealthStatus{
		Healthy:       true,
		LastCheck:     now,
		LastError:     "",
		ActiveJobs:    3,
		TotalJobs:     50,
		UptimeSeconds: 86400,
	}

	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var health2 HealthStatus
	if err := json.Unmarshal(data, &health2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if health2.Healthy != true {
		t.Error("expected healthy true")
	}
	if health2.ActiveJobs != 3 {
		t.Errorf("expected 3 active jobs, got %d", health2.ActiveJobs)
	}
}

// TestRunnerStatus_Constants tests runner status constants
func TestRunnerStatus_Constants(t *testing.T) {
	statuses := []types.RunnerStatus{
		types.RunnerStatusOffline,
		types.RunnerStatusOnline,
		types.RunnerStatusBusy,
		types.RunnerStatusError,
		types.RunnerStatusCreating,
		types.RunnerStatusCreating,
		types.RunnerStatusDestroyed,
	}

	for _, s := range statuses {
		if s == "" {
			t.Error("empty status")
		}
	}
}

// TestRunnerPlatform_Constants tests runner platform constants
func TestRunnerPlatform_Constants(t *testing.T) {
	platforms := []types.RunnerPlatform{
		types.PlatformGitLab,
		types.PlatformGitHub,
		types.PlatformJenkins,
		types.PlatformCircleCI,
		types.PlatformDrone,
	}

	for _, p := range platforms {
		if p == "" {
			t.Error("empty platform")
		}
	}
}

// TestGitLabConfig_Defaults tests GitLab config defaults
func TestGitLabConfig_Defaults(t *testing.T) {
	config := &GitLabConfig{
		URL:    "https://gitlab.example.com",
		Token:  "glrt-xxx",
		Name:   "runner-1",
		Labels: []string{"docker"},
	}

	if config.URL != "https://gitlab.example.com" {
		t.Errorf("unexpected URL: %s", config.URL)
	}
	if len(config.Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(config.Labels))
	}
}

// TestGitHubConfig_Defaults tests GitHub config defaults
func TestGitHubConfig_Defaults(t *testing.T) {
	config := &GitHubConfig{
		Repo:        "owner/repo",
		Token:       "ghp_xxx",
		Name:        "runner-1",
		Labels:      []string{"ubuntu", "docker"},
		WorkPath:    "/tmp/runner",
		RunnerGroup: "default",
	}

	if config.Repo != "owner/repo" {
		t.Errorf("unexpected repo: %s", config.Repo)
	}
	if len(config.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(config.Labels))
	}
}

// TestJenkinsConfig_Defaults tests Jenkins config defaults
func TestJenkinsConfig_Defaults(t *testing.T) {
	config := &JenkinsConfig{
		URL:       "https://jenkins.example.com",
		Username:  "jenkins",
		APIToken:  "api-token",
		Name:      "agent-1",
		AgentName: "agent-1",
		Secret:    "secret",
		Labels:    []string{"linux"},
		WorkPath:  "/tmp/agent",
	}

	if config.URL != "https://jenkins.example.com" {
		t.Errorf("unexpected URL: %s", config.URL)
	}
	if config.Username != "jenkins" {
		t.Errorf("unexpected username: %s", config.Username)
	}
}

// TestCircleCIConfig_Defaults tests CircleCI config defaults
func TestCircleCIConfig_Defaults(t *testing.T) {
	config := &CircleCIConfig{
		APIToken: "cci_xxx",
		Name:     "runner-1",
		Labels:   []string{"docker"},
		WorkPath: "/tmp/runner",
	}

	if config.APIToken != "cci_xxx" {
		t.Errorf("unexpected token: %s", config.APIToken)
	}
}

// TestDroneConfig_Defaults tests Drone config defaults
func TestDroneConfig_Defaults(t *testing.T) {
	config := &DroneConfig{
		URL:      "https://drone.example.com",
		APIToken: "drone-token",
		Name:     "runner-1",
		Labels:   []string{"docker"},
	}

	if config.URL != "https://drone.example.com" {
		t.Errorf("unexpected URL: %s", config.URL)
	}
}

// TestBaseRunner_EmptyLabels tests BaseRunner with empty labels
func TestBaseRunner_EmptyLabels(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		labels:   []string{},
	}

	labels := runner.Labels()
	if labels == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}
}

// TestHealthStatus_ZeroValues tests HealthStatus with zero values
func TestHealthStatus_ZeroValues(t *testing.T) {
	health := &HealthStatus{}

	if health.Healthy != false {
		t.Error("expected false")
	}
	if health.ActiveJobs != 0 {
		t.Errorf("expected 0 active jobs, got %d", health.ActiveJobs)
	}
}

// TestBaseRunner_NilHealth tests BaseRunner with nil health
func TestBaseRunner_NilHealth(t *testing.T) {
	runner := &BaseRunner{
		id:       "runner-123",
		platform: types.PlatformGitLab,
		health:   nil,
	}

	health := runner.Health()
	if health != nil {
		t.Error("expected nil health")
	}
}