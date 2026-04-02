// Package container provides Docker container lifecycle management
package container

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Manager manages Docker containers for ephemeral build environments
type Manager struct {
	client    *client.Client
	containers map[string]*ContainerInfo
	images     map[string]*ImageInfo
	networkMgr *NetworkManager
	mu         sync.RWMutex
	config     *Config
}

// ContainerInfo represents a managed container
type ContainerInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	Status     ContainerStatus   `json:"status"`
	Platform   string            `json:"platform"` // linux/amd64, linux/arm64
	IPAddress  string            `json:"ip_address"`
	Ports      nat.PortMap       `json:"ports"`
	Resources  *Resources        `json:"resources"`
	PipelineID string            `json:"pipeline_id"`
	CreatedAt  time.Time         `json:"created_at"`
	RemovedAt  *time.Time        `json:"removed_at,omitempty"`
}

// ContainerStatus represents container state
type ContainerStatus string

const (
	StatusCreating ContainerStatus = "creating"
	StatusRunning  ContainerStatus = "running"
	StatusExited  ContainerStatus = "exited"
	StatusFailed  ContainerStatus = "failed"
	StatusRemoved ContainerStatus = "removed"
)

// ImageInfo represents a cached Docker image
type ImageInfo struct {
	Repository string    `json:"repository"`
	Tag        string    `json:"tag"`
	Size       int64     `json:"size"`
	PulledAt   time.Time `json:"pulled_at"`
}

// Resources represents container resource limits
type Resources struct {
	CPU    float64 `json:"cpu"`    // CPU shares
	Memory int64   `json:"memory"` // bytes
	Disk   int64   `json:"disk"`   // bytes
}

// Config represents container manager configuration
type Config struct {
	Runtime        string   `json:"runtime"`         // docker, podman
	SocketPath    string   `json:"socket_path"`     // /var/run/docker.sock
	ImageCacheDir string   `json:"image_cache_dir"` // for pre-pulled images
	Networks      []string `json:"networks"`        // allowed networks
	DefaultImage  string   `json:"default_image"`   // default build image
}

// NewManager creates a new container manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = &Config{
			SocketPath:    "/var/run/docker.sock",
			DefaultImage:  "ubuntu:22.04",
			ImageCacheDir: "/var/lib/vimic2/images",
		}
	}

	cli, err := client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithUnixSocket(config.SocketPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	m := &Manager{
		client:     cli,
		containers: make(map[string]*ContainerInfo),
		images:     make(map[string]*ImageInfo),
		networkMgr: NewNetworkManager(cli),
		config:     config,
	}

	// Pre-populate with known images
	m.images["ubuntu:22.04"] = &ImageInfo{Repository: "ubuntu", Tag: "22.04"}
	m.images["ubuntu:24.04"] = &ImageInfo{Repository: "ubuntu", Tag: "24.04"}
	m.images["alpine:3.19"] = &ImageInfo{Repository: "alpine", Tag: "3.19"}
	m.images["debian:bookworm"] = &ImageInfo{Repository: "debian", Tag: "bookworm"}
	m.images["golang:1.23"] = &ImageInfo{Repository: "golang", Tag: "1.23"}
	m.images["node:20"] = &ImageInfo{Repository: "node", Tag: "20"}
	m.images["python:3.12"] = &ImageInfo{Repository: "python", Tag: "3.12"}
	m.images["rust:1.77"] = &ImageInfo{Repository: "rust", Tag: "1.77"}

	return m, nil
}

// ListContainers returns all managed containers
func (m *Manager) ListContainers() []*ContainerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	containers := make([]*ContainerInfo, 0, len(m.containers))
	for _, c := range m.containers {
		containers = append(containers, c)
	}
	return containers
}

// ListImages returns all cached images
func (m *Manager) ListImages() []*ImageInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	images := make([]*ImageInfo, 0, len(m.images))
	for _, img := range m.images {
		images = append(images, img)
	}
	return images
}

// PullImage pulls a Docker image
func (m *Manager) PullImage(ctx context.Context, ref string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already cached
	if _, exists := m.images[ref]; exists {
		return nil
	}

	reader, err := m.client.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", ref, err)
	}
	defer reader.Close()

	// Wait for pull to complete
	buf := make([]byte, 1024)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
	}

	// Cache the image
	m.images[ref] = &ImageInfo{
		Repository: ref,
		Tag:        "latest",
		PulledAt:   time.Now(),
	}

	return nil
}

// CreateContainer creates a new ephemeral container
func (m *Manager) CreateContainer(ctx context.Context, opts *CreateOptions) (*ContainerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if opts.Image == "" {
		opts.Image = m.config.DefaultImage
	}

	// Create container config
	hostConfig := &container.HostConfig{
		NetworkMode:  opts.NetworkMode,
		Memory:       opts.Resources.Memory,
		NanoCPUs:     int64(opts.Resources.CPU * 1e9),
		Privileged:   opts.Privileged,
		CapAdd:       opts.CapAdd,
		PortBindings: opts.PortBindings,
		Binds:        opts.Volumes,
	}

	if opts.AutoRemove {
		hostConfig.AutoRemove = true
	}

	// Create container
	resp, err := m.client.ContainerCreate(ctx, &container.Config{
		Image:        opts.Image,
		Hostname:     opts.Hostname,
		Env:          opts.Env,
		Cmd:          opts.Cmd,
		WorkingDir:   opts.WorkingDir,
		Labels:       opts.Labels,
		ExposedPorts: opts.ExposedPorts,
	}, hostConfig, nil, nil, opts.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	info := &ContainerInfo{
		ID:         resp.ID,
		Name:       opts.Name,
		Image:      opts.Image,
		Status:     StatusCreating,
		Platform:   opts.Platform,
		PipelineID: opts.PipelineID,
		Resources:  opts.Resources,
		CreatedAt:  time.Now(),
	}

	m.containers[resp.ID] = info
	return info, nil
}

// StartContainer starts a created container
func (m *Manager) StartContainer(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container %s not found", id)
	}

	if err := m.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		info.Status = StatusFailed
		return fmt.Errorf("failed to start container: %w", err)
	}

	info.Status = StatusRunning
	return nil
}

// StopContainer stops a running container
func (m *Manager) StopContainer(ctx context.Context, id string, timeout time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container %s not found", id)
	}

	if err := m.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		info.Status = StatusFailed
		return fmt.Errorf("failed to stop container: %w", err)
	}

	info.Status = StatusExited
	return nil
}

// RemoveContainer removes a container
func (m *Manager) RemoveContainer(ctx context.Context, id string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.containers[id]
	if !exists {
		return fmt.Errorf("container %s not found", id)
	}

	opts := container.RemoveOptions{Force: force}
	if err := m.client.ContainerRemove(ctx, id, opts); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	now := time.Now()
	info.Status = StatusRemoved
	info.RemovedAt = &now
	delete(m.containers, id)

	return nil
}

// GetContainer returns container info
func (m *Manager) GetContainer(ctx context.Context, id string) (*ContainerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.containers[id]
	if !exists {
		// Try to get from Docker
		info2, err := m.client.ContainerInspect(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("container %s not found", id)
		}

		// Reconstruct info
		info = &ContainerInfo{
			ID:        info2.ID,
			Name:      info2.Name,
			Image:     info2.Config.Image,
			CreatedAt: info2.Created,
		}

		if info2.State.Running {
			info.Status = StatusRunning
		} else {
			info.Status = StatusExited
		}
	}

	return info, nil
}

// CreateOptions represents container creation options
type CreateOptions struct {
	Name        string
	Image       string
	Platform    string // linux/amd64, linux/arm64
	PipelineID  string
	Hostname    string
	Env         []string
	Cmd         []string
	WorkingDir  string
	Labels      map[string]string
	NetworkMode string
	Privileged  bool
	CapAdd      []string
	AutoRemove  bool
	Resources   *Resources
	PortBindings nat.PortMap
	Volumes     []string
	ExposedPorts nat.PortSet
}
