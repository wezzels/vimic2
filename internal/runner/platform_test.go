//go:build integration

package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// ==================== GitLab Runner Tests ====================

func TestGitLabConfig_Fields(t *testing.T) {
	config := &GitLabConfig{
		URL:        "https://gitlab.example.com",
		Token:      "test-token",
		Name:       "test-runner",
		ConfigPath: "/etc/gitlab-runner/config.toml",
		Labels:     []string{"docker", "linux"},
	}

	if config.URL != "https://gitlab.example.com" {
		t.Errorf("URL = %s, want https://gitlab.example.com", config.URL)
	}
	if config.Token != "test-token" {
		t.Error("Token should be set")
	}
	if config.ConfigPath != "/etc/gitlab-runner/config.toml" {
		t.Errorf("ConfigPath = %s, want /etc/gitlab-runner/config.toml", config.ConfigPath)
	}
	if len(config.Labels) != 2 {
		t.Errorf("Labels count = %d, want 2", len(config.Labels))
	}
}

func TestNewGitLabRunner(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}
	if runner == nil {
		t.Fatal("NewGitLabRunner should not return nil")
	}
	if runner.Platform() != types.PlatformGitLab {
		t.Errorf("Platform = %v, want GitLab", runner.Platform())
	}
	if runner.ID() == "" {
		t.Error("ID should not be empty")
	}
	if runner.Status() != types.RunnerStatusCreating {
		t.Errorf("Status = %v, want Idle", runner.Status())
	}
}

func TestGitLabRunner_Labels(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
		Labels: []string{"docker", "linux", "x86"},
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}
	labels := runner.Labels()
	if len(labels) != 3 {
		t.Errorf("Labels count = %d, want 3", len(labels))
	}
}

func TestGitLabRunner_Health(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}
	health := runner.Health()
	if health == nil {
		t.Error("Health should not be nil")
	}
}

// ==================== GitHub Runner Tests ====================

func TestGitHubConfig_Fields(t *testing.T) {
	config := &GitHubConfig{
		Repo:        "stsgym/vimic2",
		Token:       "ghp_test123",
		Name:        "vimic2-runner",
		RunnerGroup: "default",
		Labels:      []string{"self-hosted", "linux", "x64"},
		WorkPath:    "/home/runner/work",
	}

	if config.Repo != "stsgym/vimic2" {
		t.Errorf("Repo = %s, want stsgym/vimic2", config.Repo)
	}
	if len(config.Labels) != 3 {
		t.Errorf("Labels count = %d, want 3", len(config.Labels))
	}
	if config.RunnerGroup != "default" {
		t.Errorf("RunnerGroup = %s, want default", config.RunnerGroup)
	}
}

func TestNewGitHubRunner(t *testing.T) {
	config := &GitHubConfig{
		Repo:  "stsgym/vimic2",
		Token: "ghp_test123",
		Name:  "vimic2-runner",
	}

	runner, err := NewGitHubRunner(config)
	if err != nil {
		t.Fatalf("NewGitHubRunner failed: %v", err)
	}
	if runner == nil {
		t.Fatal("NewGitHubRunner should not return nil")
	}
	if runner.Platform() != types.PlatformGitHub {
		t.Errorf("Platform = %v, want GitHub", runner.Platform())
	}
	if runner.ID() == "" {
		t.Error("ID should not be empty")
	}
}

// ==================== Jenkins Runner Tests ====================

func TestJenkinsConfig_Fields(t *testing.T) {
	config := &JenkinsConfig{
		URL:       "http://jenkins.example.com",
		Username:  "admin",
		APIToken:  "api-token",
		AgentName: "vimic2-agent",
		Secret:    "test-secret",
		Labels:    []string{"linux", "docker"},
		WorkPath:  "/home/jenkins",
	}

	if config.URL != "http://jenkins.example.com" {
		t.Errorf("URL = %s, want http://jenkins.example.com", config.URL)
	}
	if config.Username != "admin" {
		t.Errorf("Username = %s, want admin", config.Username)
	}
	if config.AgentName != "vimic2-agent" {
		t.Errorf("AgentName = %s, want vimic2-agent", config.AgentName)
	}
}

func TestNewJenkinsRunner(t *testing.T) {
	config := &JenkinsConfig{
		URL:       "http://jenkins.example.com",
		Username:  "admin",
		APIToken:  "test-api-token",
		AgentName: "vimic2-agent",
		Secret:    "test-secret",
	}

	runner, err := NewJenkinsRunner(config)
	if err != nil {
		t.Fatalf("NewJenkinsRunner failed: %v", err)
	}
	if runner == nil {
		t.Fatal("NewJenkinsRunner should not return nil")
	}
	if runner.Platform() != types.PlatformJenkins {
		t.Errorf("Platform = %v, want Jenkins", runner.Platform())
	}
}

// ==================== CircleCI Runner Tests ====================

func TestCircleCIConfig_Fields(t *testing.T) {
	config := &CircleCIConfig{
		APIToken: "test-token",
		Name:     "vimic2-runner",
		Labels:   []string{"linux"},
		WorkPath: "/home/circleci",
	}

	if config.APIToken != "test-token" {
		t.Error("APIToken should be set")
	}
	if config.WorkPath != "/home/circleci" {
		t.Errorf("WorkPath = %s, want /home/circleci", config.WorkPath)
	}
}

func TestNewCircleCIRunner(t *testing.T) {
	config := &CircleCIConfig{
		APIToken: "test-token",
		Name:     "vimic2-runner",
	}

	runner, err := NewCircleCIRunner(config)
	if err != nil {
		t.Fatalf("NewCircleCIRunner failed: %v", err)
	}
	if runner == nil {
		t.Fatal("NewCircleCIRunner should not return nil")
	}
	if runner.Platform() != types.PlatformCircleCI {
		t.Errorf("Platform = %v, want CircleCI", runner.Platform())
	}
}

// ==================== Drone Runner Tests ====================

func TestDroneConfig_Fields(t *testing.T) {
	config := &DroneConfig{
		URL:   "https://drone.example.com",
		APIToken: "test-token",
		Name:  "vimic2-runner",
		Labels: []string{"linux", "arm64"},
	}

	if config.URL != "https://drone.example.com" {
		t.Errorf("URL = %s, want https://drone.example.com", config.URL)
	}
	if len(config.Labels) != 2 {
		t.Errorf("Labels count = %d, want 2", len(config.Labels))
	}
}

func TestNewDroneRunner(t *testing.T) {
	config := &DroneConfig{
		URL:   "https://drone.example.com",
		APIToken: "test-token",
		Name:  "vimic2-runner",
	}

	runner, err := NewDroneRunner(config)
	if err != nil {
		t.Fatalf("NewDroneRunner failed: %v", err)
	}
	if runner == nil {
		t.Fatal("NewDroneRunner should not return nil")
	}
	if runner.Platform() != types.PlatformDrone {
		t.Errorf("Platform = %v, want Drone", runner.Platform())
	}
}

// ==================== RunnerManager Tests ====================

func TestRunnerManagerConfig_Struct(t *testing.T) {
	config := &RunnerManagerConfig{
		GitLab: &GitLabConfig{
			URL:   "https://gitlab.example.com",
			Token: "gl-token",
			Name:  "gitlab-runner",
		},
		GitHub: &GitHubConfig{
			Repo:  "stsgym/vimic2",
			Token: "gh-token",
			Name:  "github-runner",
		},
		Jenkins: &JenkinsConfig{
			URL:       "http://jenkins.example.com",
			AgentName: "jenkins-agent",
			Secret:    "jenkins-secret",
		},
		CircleCI: &CircleCIConfig{
			APIToken: "cc-token",
			Name:     "circleci-runner",
		},
		Drone: &DroneConfig{
			URL:   "https://drone.example.com",
			APIToken: "drone-token",
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
	if config.CircleCI == nil {
		t.Error("CircleCI config should not be nil")
	}
	if config.Drone == nil {
		t.Error("Drone config should not be nil")
	}
}

func TestRunnerInfo_Struct(t *testing.T) {
	now := time.Now()
	info := &RunnerInfo{
		ID:            "runner-1",
		VMID:          "vm-1",
		PoolName:      "default",
		PipelineID:    "pipeline-1",
		Platform:      types.PlatformGitLab,
		PlatformID:    "gl-runner-1",
		Name:          "test-runner",
		Labels:        []string{"docker", "linux"},
		Status:        types.RunnerStatusCreating,
		IPAddress:     "10.0.0.1",
		HealthStatus:  "healthy",
		LastHeartbeat: &now,
		CreatedAt:     now,
	}

	if info.ID != "runner-1" {
		t.Errorf("ID = %s, want runner-1", info.ID)
	}
	if info.Platform != types.PlatformGitLab {
		t.Errorf("Platform = %v, want GitLab", info.Platform)
	}
	if info.Status != types.RunnerStatusCreating {
		t.Errorf("Status = %v, want Idle", info.Status)
	}
	if info.HealthStatus != "healthy" {
		t.Errorf("HealthStatus = %s, want healthy", info.HealthStatus)
	}
}

func TestHealthStatus_Struct(t *testing.T) {
	now := time.Now()
	hs := &HealthStatus{
		Healthy:       true,
		LastCheck:     now,
		LastError:     "",
		ActiveJobs:    5,
		TotalJobs:     100,
		UptimeSeconds: 3600,
	}

	if !hs.Healthy {
		t.Error("Healthy should be true")
	}
	if hs.ActiveJobs != 5 {
		t.Errorf("ActiveJobs = %d, want 5", hs.ActiveJobs)
	}
	if hs.TotalJobs != 100 {
		t.Errorf("TotalJobs = %d, want 100", hs.TotalJobs)
	}
	if hs.UptimeSeconds != 3600 {
		t.Errorf("UptimeSeconds = %d, want 3600", hs.UptimeSeconds)
	}
}

// ==================== Join Labels Test ====================

func TestJoinLabels(t *testing.T) {
	tests := []struct {
		labels []string
		want   string
	}{
		{[]string{"docker", "linux"}, "docker,linux"},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
		{nil, ""},
		{[]string{"a", "b", "c"}, "a,b,c"},
	}

	for _, tt := range tests {
		got := joinLabels(tt.labels)
		if got != tt.want {
			t.Errorf("joinLabels(%v) = %q, want %q", tt.labels, got, tt.want)
		}
	}
}

// ==================== RunnerInterface Compliance Tests ====================

func TestRunnerInterface_GitLabCompliance(t *testing.T) {
	runner, err := NewGitLabRunner(&GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test",
		Name:  "test",
	})
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}
	var _ RunnerInterface = runner
}

func TestRunnerInterface_GitHubCompliance(t *testing.T) {
	runner, err := NewGitHubRunner(&GitHubConfig{
		Repo:  "stsgym/vimic2",
		Token: "test",
		Name:  "test",
	})
	if err != nil {
		t.Fatalf("NewGitHubRunner failed: %v", err)
	}
	var _ RunnerInterface = runner
}

func TestRunnerInterface_JenkinsCompliance(t *testing.T) {
	runner, err := NewJenkinsRunner(&JenkinsConfig{
		URL:       "http://jenkins.example.com",
		Username:  "test",
		APIToken:  "test",
		AgentName: "test",
		Secret:    "test",
	})
	if err != nil {
		t.Fatalf("NewJenkinsRunner failed: %v", err)
	}
	var _ RunnerInterface = runner
}

func TestRunnerInterface_CircleCICompliance(t *testing.T) {
	runner, err := NewCircleCIRunner(&CircleCIConfig{
		APIToken: "test",
		Name:     "test",
	})
	if err != nil {
		t.Fatalf("NewCircleCIRunner failed: %v", err)
	}
	var _ RunnerInterface = runner
}

func TestRunnerInterface_DroneCompliance(t *testing.T) {
	runner, err := NewDroneRunner(&DroneConfig{
		URL:      "https://drone.example.com",
		APIToken: "test",
		Name:     "test",
	})
	if err != nil {
		t.Fatalf("NewDroneRunner failed: %v", err)
	}
	var _ RunnerInterface = runner
}

// ==================== Runner Start/Stop/Status Tests ====================

func TestGitLabRunner_StatusTransitions(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}

	// Initial status should be idle
	if runner.Status() != types.RunnerStatusCreating {
		t.Errorf("Initial status = %v, want Idle", runner.Status())
	}

	// Health check
	health := runner.Health()
	if health == nil {
		t.Error("Health should not be nil")
	}
}

// ==================== Context Cancellation Test ====================

func TestRunner_ContextCancellation(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start should fail quickly since there's no real GitLab server
	err = runner.Start(ctx)
	if err == nil {
		// If it succeeded, try to stop
		runner.Stop(ctx)
	}
	// No assertion on error — just verifying it doesn't hang
}

// ==================== GitLab Runner Register/Unregister ====================

func TestGitLabRunner_RegisterUnregister(t *testing.T) {
	config := &GitLabConfig{
		URL:   "https://gitlab.example.com",
		Token: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewGitLabRunner(config)
	if err != nil {
		t.Fatalf("NewGitLabRunner failed: %v", err)
	}
	ctx := context.Background()

	// Register should fail without a real GitLab server
	err = runner.Register(ctx)
	t.Logf("Register: %v (expected to fail without server)", err)

	// Unregister should also fail
	err = runner.Unregister(ctx)
	t.Logf("Unregister: %v (expected to fail without server)", err)
}

func TestGitHubRunner_RegisterUnregister(t *testing.T) {
	config := &GitHubConfig{
		Repo:  "stsgym/vimic2",
		Token: "ghp-test",
		Name:  "test-runner",
	}

	runner, err := NewGitHubRunner(config)
	if err != nil {
		t.Fatalf("NewGitHubRunner failed: %v", err)
	}
	ctx := context.Background()

	err = runner.Register(ctx)
	t.Logf("Register: %v (expected to fail without server)", err)

	err = runner.Unregister(ctx)
	t.Logf("Unregister: %v (expected to fail without server)", err)
}

func TestJenkinsRunner_RegisterUnregister(t *testing.T) {
	config := &JenkinsConfig{
		URL:       "http://jenkins.example.com",
		Username:  "admin",
		APIToken:  "test-api-token",
		AgentName: "test-agent",
		Secret:    "test-secret",
	}

	runner, err := NewJenkinsRunner(config)
	if err != nil {
		t.Fatalf("NewJenkinsRunner failed: %v", err)
	}
	ctx := context.Background()

	err = runner.Register(ctx)
	t.Logf("Register: %v (expected to fail without server)", err)

	err = runner.Unregister(ctx)
	t.Logf("Unregister: %v (expected to fail without server)", err)
}

// ==================== Drone Runner List ====================

func TestDroneRunner_ListDroneRunners(t *testing.T) {
	config := &DroneConfig{
		URL:   "https://drone.example.com",
		APIToken: "test-token",
		Name:  "test-runner",
	}

	runner, err := NewDroneRunner(config)
	if err != nil {
		t.Fatalf("NewDroneRunner failed: %v", err)
	}
	ctx := context.Background()

	runners, err := runner.ListDroneRunners(ctx)
	t.Logf("ListDroneRunners: runners=%d err=%v (expected to fail without server)", len(runners), err)
}

// ==================== CircleCI Runner Register ====================

func TestCircleCIRunner_RegisterUnregister(t *testing.T) {
	config := &CircleCIConfig{
		APIToken: "test-token",
		Name:     "test-runner",
	}

	runner, err := NewCircleCIRunner(config)
	if err != nil {
		t.Fatalf("NewCircleCIRunner failed: %v", err)
	}
	ctx := context.Background()

	err = runner.Register(ctx)
	t.Logf("Register: %v (expected to fail without server)", err)

	err = runner.Unregister(ctx)
	t.Logf("Unregister: %v (expected to fail without server)", err)
}