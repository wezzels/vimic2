// Package runner provides CI/CD runner implementations for multiple platforms.
// This file defines the common RunnerInterface that all platform-specific runners must implement.
package runner

import (
	"context"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// RunnerInterface defines the common interface for all CI/CD runners.
// Each platform (GitLab, GitHub, Jenkins, etc.) must implement this interface.
type RunnerInterface interface {
	// Register registers the runner with the platform.
	Register(ctx context.Context) error
	
	// Start starts the runner process.
	Start(ctx context.Context) error
	
	// Stop gracefully stops the runner.
	Stop(ctx context.Context) error
	
	// Unregister removes the runner from the platform.
	Unregister(ctx context.Context) error
	
	// Status returns the current status of the runner.
	Status() types.RunnerStatus
	
	// ID returns the unique identifier of the runner.
	ID() string
	
	// Platform returns the runner platform type.
	Platform() types.RunnerPlatform
	
	// Labels returns the runner's labels/tags.
	Labels() []string
	
	// Health returns health status.
	Health() *HealthStatus
}

// HealthStatus represents runner health information.
type HealthStatus struct {
	Healthy       bool      `json:"healthy"`
	LastCheck     time.Time `json:"last_check"`
	LastError     string    `json:"last_error,omitempty"`
	ActiveJobs    int       `json:"active_jobs"`
	TotalJobs     int       `json:"total_jobs"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// BaseRunner provides common functionality for all runners.
type BaseRunner struct {
	id         string
	platform   types.RunnerPlatform
	labels     []string
	status     types.RunnerStatus
	health     *HealthStatus
	registered bool
}

// ID returns the runner's unique identifier.
func (r *BaseRunner) ID() string {
	return r.id
}

// Platform returns the runner's platform type.
func (r *BaseRunner) Platform() types.RunnerPlatform {
	return r.platform
}

// Labels returns the runner's labels.
func (r *BaseRunner) Labels() []string {
	return r.labels
}

// Status returns the current status.
func (r *BaseRunner) Status() types.RunnerStatus {
	return r.status
}

// Health returns health status.
func (r *BaseRunner) Health() *HealthStatus {
	return r.health
}

// SetStatus updates the runner status.
func (r *BaseRunner) SetStatus(status types.RunnerStatus) {
	r.status = status
}

// SetHealth updates the health status.
func (r *BaseRunner) SetHealth(health *HealthStatus) {
	r.health = health
}

// Config types for each platform

// GitLabConfig holds GitLab runner configuration.
type GitLabConfig struct {
	URL        string   `json:"url"`
	Token      string   `json:"token"`
	Name       string   `json:"name"`
	ConfigPath string   `json:"config_path"`
	Labels     []string `json:"labels"`
}

// GitHubConfig holds GitHub runner configuration.
type GitHubConfig struct {
	Repo        string   `json:"repo"`
	Token       string   `json:"token"`
	Name        string   `json:"name"`
	Labels      []string `json:"labels"`
	WorkPath    string   `json:"work_path"`
	RunnerGroup string   `json:"runner_group"`
}

// JenkinsConfig holds Jenkins runner configuration.
type JenkinsConfig struct {
	URL       string   `json:"url"`
	Username  string   `json:"username"`
	APIToken  string   `json:"api_token"`
	Name      string   `json:"name"`
	AgentName string   `json:"agent_name"`
	Secret    string   `json:"secret"`
	Labels    []string `json:"labels"`
	WorkPath  string   `json:"work_path"`
}

// CircleCIConfig holds CircleCI runner configuration.
type CircleCIConfig struct {
	APIToken string   `json:"api_token"`
	Name     string   `json:"name"`
	Labels   []string `json:"labels"`
	WorkPath string   `json:"work_path"`
}

// DroneConfig holds Drone runner configuration.
type DroneConfig struct {
	URL      string   `json:"url"`
	APIToken string   `json:"api_token"`
	Name     string   `json:"name"`
	Labels   []string `json:"labels"`
}