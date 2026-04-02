// Package runner provides CI/CD runner orchestration
package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/stsgym/vimic2/internal/container"
)

// DockerRunner manages Docker-based ephemeral runners
type DockerRunner struct {
	containerMgr *container.Manager
	runnerMgr    *RunnerManager
	config      *DockerRunnerConfig
}

// DockerRunnerConfig represents Docker runner configuration
type DockerRunnerConfig struct {
	DefaultImage    string   `json:"default_image"`
	AllowedImages   []string `json:"allowed_images"`
	NetworkPrefix   string   `json:"network_prefix"`   // e.g., "vimic2-"
	DefaultPlatform string   `json:"default_platform"` // linux/amd64
	MaxContainers   int      `json:"max_containers"`  // per runner
	AutoRemove      bool     `json:"auto_remove"`     // remove after job
}

// DockerRunnerInfo represents a Docker container runner
type DockerRunnerInfo struct {
	ID           string                `json:"id"`
	ContainerID  string                `json:"container_id"`
	Image        string                `json:"image"`
	Platform     string                `json:"platform"`
	Network      string                `json:"network"`
	Status       RunnerStatus          `json:"status"`
	IPAddress    string                `json:"ip_address"`
	PipelineID   string                `json:"pipeline_id"`
	CurrentJobID string                `json:"current_job_id,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
}

// NewDockerRunner creates a new Docker runner manager
func NewDockerRunner(containerMgr *container.Manager, runnerMgr *RunnerManager, config *DockerRunnerConfig) (*DockerRunner, error) {
	if config == nil {
		config = &DockerRunnerConfig{
			DefaultImage:    "ubuntu:22.04",
			DefaultPlatform: "linux/amd64",
			MaxContainers:   10,
			AutoRemove:      true,
		}
	}

	return &DockerRunner{
		containerMgr: containerMgr,
		runnerMgr:    runnerMgr,
		config:       config,
	}, nil
}

// CreateRunner creates a new Docker-based runner for a pipeline
func (dr *DockerRunner) CreateRunner(ctx context.Context, pipelineID string, opts *DockerRunnerOptions) (*DockerRunnerInfo, error) {
	// Create isolated network for this runner
	network, err := dr.containerMgr.networkMgr.CreateNetwork(ctx, pipelineID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	// Determine image
	image := dr.config.DefaultImage
	if opts != nil && opts.Image != "" {
		image = opts.Image
	}

	// Pull image if needed
	if err := dr.containerMgr.PullImage(ctx, image); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Create container
	containerName := fmt.Sprintf("vimic2-%s-runner", pipelineID)
	containerInfo, err := dr.containerMgr.CreateContainer(ctx, &container.CreateOptions{
		Name:        containerName,
		Image:       image,
		Platform:    dr.config.DefaultPlatform,
		PipelineID:  pipelineID,
		NetworkMode: "container:" + network.ID,
		AutoRemove:  dr.config.AutoRemove,
		Resources: &container.Resources{
			CPU:    2.0,
			Memory: 4 * 1024 * 1024 * 1024, // 4GB
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := dr.containerMgr.StartContainer(ctx, containerInfo.ID); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Get container IP
	info, _ := dr.containerMgr.GetContainer(ctx, containerInfo.ID)

	return &DockerRunnerInfo{
		ID:          containerInfo.ID,
		ContainerID: containerInfo.ID,
		Image:       image,
		Platform:    dr.config.DefaultPlatform,
		Network:     network.Name,
		Status:      RunnerStatusOnline,
		IPAddress:   info.IPAddress,
		PipelineID:  pipelineID,
		CreatedAt:   time.Now(),
	}, nil
}

// StopRunner stops a Docker runner
func (dr *DockerRunner) StopRunner(ctx context.Context, runnerID string) error {
	return dr.containerMgr.StopContainer(ctx, runnerID, 30*time.Second)
}

// DestroyRunner removes a Docker runner
func (dr *DockerRunner) DestroyRunner(ctx context.Context, runnerID string) error {
	// Get container info first to find network
	info, err := dr.containerMgr.GetContainer(ctx, runnerID)
	if err != nil {
		return err
	}

	// Remove container
	if err := dr.containerMgr.RemoveContainer(ctx, runnerID, true); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// Remove network
	if info != nil && info.Name != "" {
		dr.containerMgr.networkMgr.RemoveNetwork(ctx, info.Name)
	}

	return nil
}

// ListRunners returns all Docker runners
func (dr *DockerRunner) ListRunners() []*DockerRunnerInfo {
	containers := dr.containerMgr.ListContainers()
	runners := make([]*DockerRunnerInfo, 0, len(containers))

	for _, c := range containers {
		runners = append(runners, &DockerRunnerInfo{
			ID:          c.ID,
			ContainerID: c.ID,
			Image:       c.Image,
			Platform:    c.Platform,
			Status:      RunnerStatus(c.Status),
			IPAddress:   c.IPAddress,
			PipelineID:  c.PipelineID,
			CreatedAt:   c.CreatedAt,
		})
	}

	return runners
}

// DockerRunnerOptions represents Docker runner creation options
type DockerRunnerOptions struct {
	Image     string
	Platform  string
	Resources *container.Resources
}
