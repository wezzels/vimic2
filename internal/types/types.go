// Package types defines the interfaces used across vimic2.
// These interfaces bridge the concrete implementations in internal/pipeline,
// internal/pool, internal/runner, and internal/network.
package types

import (
	"context"
	"time"
)

// PipelineDB defines the database interface for pipeline operations.
// Updated to match the concrete PipelineDB implementation in internal/pipeline.
type PipelineDB interface {
	// Pipeline operations
	SavePipeline(ctx context.Context, p *PipelineState) error
	GetPipeline(ctx context.Context, id string) (*PipelineState, error)
	ListPipelines(ctx context.Context, limit, offset int) ([]*PipelineState, error)
	UpdatePipelineStatus(ctx context.Context, id string, status PipelineStatus) error
	DeletePipeline(ctx context.Context, id string) error

	// Runner operations
	SaveRunner(ctx context.Context, r *RunnerState) error
	GetRunner(ctx context.Context, id string) (*RunnerState, error)
	ListRunnersByPipeline(ctx context.Context, pipelineID string) ([]*RunnerState, error)
	DeleteRunner(ctx context.Context, id string) error

	// VM operations
	SaveVM(ctx context.Context, vm *VMState) error
	GetVM(ctx context.Context, id string) (*VMState, error)
	ListVMsByPool(ctx context.Context, poolID string) ([]*VMState, error)
	UpdateVMState(ctx context.Context, id string, state string) error
	DeleteVM(ctx context.Context, id string) error

	// Template operations
	SaveTemplate(ctx context.Context, t *TemplateState) error
	GetTemplate(ctx context.Context, id string) (*TemplateState, error)
	ListTemplates(ctx context.Context) ([]*TemplateState, error)
	DeleteTemplate(ctx context.Context, id string) error

	// Pool operations
	SavePool(ctx context.Context, p *PoolState) error
	GetPool(ctx context.Context, id string) (*PoolState, error)
	ListPools(ctx context.Context) ([]*PoolState, error)
	UpdatePoolSize(ctx context.Context, id string, delta int) error
	DeletePool(ctx context.Context, id string) error

	// Network operations
	SaveNetwork(ctx context.Context, n *NetworkState) error
	GetNetwork(ctx context.Context, id string) (*NetworkState, error)
	ListNetworks(ctx context.Context) ([]*NetworkState, error)
	DeleteNetwork(ctx context.Context, id string) error

	// Artifact operations
	SaveMetric(ctx context.Context, m *Metric) error
}

// PipelineStatus represents the status of a pipeline
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusCreating  PipelineStatus = "creating"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSuccess   PipelineStatus = "success"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCancelled PipelineStatus = "cancelled"
)

// RunnerStatus represents the status of a runner
type RunnerStatus string

const (
	RunnerStatusPending  RunnerStatus = "pending"
	RunnerStatusOnline   RunnerStatus = "online"
	RunnerStatusOffline  RunnerStatus = "offline"
	RunnerStatusError    RunnerStatus = "error"
	RunnerStatusBusy       RunnerStatus = "busy"
	RunnerStatusCreating  RunnerStatus = "creating"
	RunnerStatusDestroyed RunnerStatus = "destroyed"
)

// PipelineState represents the state of a pipeline (for the interface)
type PipelineState struct {
	ID         string         `json:"id"`
	Platform   RunnerPlatform `json:"platform"`
	Repository string         `json:"repository"`
	Branch     string         `json:"branch"`
	CommitSHA  string         `json:"commit_sha"`
	CommitMsg  string         `json:"commit_message"`
	Author     string         `json:"author"`
	Status     PipelineStatus `json:"status"`
	NetworkID  string         `json:"network_id"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Duration   int64          `json:"duration_seconds"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// RunnerState represents the state of a runner (for the interface)
type RunnerState struct {
	ID          string         `json:"id"`
	PipelineID  string         `json:"pipeline_id"`
	VMID        string         `json:"vm_id"`
	Platform    RunnerPlatform `json:"platform"`
	PlatformID  string         `json:"platform_runner_id"`
	Labels      []string       `json:"labels"`
	Name        string         `json:"name"`
	Status      RunnerStatus   `json:"status"`
	CurrentJob  string         `json:"current_job,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DestroyedAt *time.Time     `json:"destroyed_at,omitempty"`
}

// VMState represents the state of a VM (for the interface)
type VMState struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	IPAddress   string     `json:"ip_address"`
	MACAddress  string     `json:"mac_address"`
	PoolName    string     `json:"pool_name"`
	Template    string     `json:"template"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// PoolState represents the state of a pool (for the interface)
type PoolState struct {
	Name         string `json:"name"`
	TemplatePath string `json:"template_path"`
	Capacity     int    `json:"capacity"`
	Available    int    `json:"available"`
	Busy         int    `json:"busy"`
}

// NetworkState represents the state of a network (for the interface)
type NetworkState struct {
	ID          string     `json:"id"`
	PipelineID  string     `json:"pipeline_id"`
	BridgeName  string     `json:"bridge_name"`
	VLANID      int        `json:"vlan_id"`
	CIDR        string     `json:"cidr"`
	Gateway     string     `json:"gateway"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// TemplateState represents the state of a template (for the interface)
type TemplateState struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseImage string    `json:"base_image"`
	CPU       int       `json:"cpu"`
	Memory    int       `json:"memory"`
	CreatedAt time.Time `json:"created_at"`
}

// Metric represents a metric data point (for the interface)
type Metric struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Disk      float64   `json:"disk"`
	CreatedAt time.Time `json:"created_at"`
}

// Artifact represents a pipeline artifact
type Artifact struct {
	ID         string    `json:"id"`
	PipelineID string    `json:"pipeline_id"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Checksum   string    `json:"checksum"`
	TTL        int       `json:"ttl"`
	CreatedAt  time.Time `json:"created_at"`
}

// RunnerPlatform represents the CI/CD platform type
type RunnerPlatform string

const (
	PlatformGitLab   RunnerPlatform = "gitlab"
	PlatformGitHub   RunnerPlatform = "github"
	PlatformJenkins  RunnerPlatform = "jenkins"
	PlatformCircleCI RunnerPlatform = "circleci"
	PlatformDrone    RunnerPlatform = "drone"
	PlatformLocal    RunnerPlatform = "local"
)

// PoolManagerInterface is the interface for pool management
type PoolManagerInterface interface {
	AllocateVM(poolName string) (*VMState, error)
	ReleaseVM(vmID string) error
	GetPool(name string) (*PoolState, error)
	ListPools() ([]*PoolState, error)
}

// NetworkManagerInterface is the interface for network management
type NetworkManagerInterface interface {
	CreateNetwork(config *NetworkConfig) (string, error)
	DestroyNetwork(networkID string) error
	GetNetwork(networkID string) (*NetworkConfig, error)
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	VLAN     int      `json:"vlan"`
	CIDR     string   `json:"cidr"`
	Gateway  string   `json:"gateway"`
	DNS      []string `json:"dns"`
	Isolated bool     `json:"isolated"`
}

// RunnerManagerInterface is the interface for runner management
type RunnerManagerInterface interface {
	CreateRunner(platform RunnerPlatform, config map[string]interface{}) (string, error)
	DestroyRunner(runnerID string) error
	GetRunner(runnerID string) (map[string]interface{}, error)
}