// Package types provides shared types to break import cycles
package types

import "time"

// PipelineDB interface for database operations
type PipelineDB interface {
	SavePipeline(id string, state map[string]interface{}) error
	LoadPipeline(id string) (map[string]interface{}, error)
	DeletePipeline(id string) error
	ListPipelines() ([]string, error)
	SaveRunner(id string, state map[string]interface{}) error
	LoadRunner(id string) (map[string]interface{}, error)
	DeleteRunner(id string) error
	SaveNetwork(id string, state map[string]interface{}) error
	LoadNetwork(id string) (map[string]interface{}, error)
	DeleteNetwork(id string) error
	SavePool(id string, state map[string]interface{}) error
	LoadPool(id string) (map[string]interface{}, error)
	DeletePool(id string) error
	UpdatePoolSize(id string, available, busy int) error
	UpdateVMState(vmID string, state string) error
}

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

// RunnerManagerInterface is the interface for runner management
type RunnerManagerInterface interface {
	CreateRunner(platform RunnerPlatform, config map[string]interface{}) (string, error)
	DestroyRunner(runnerID string) error
	GetRunner(runnerID string) (map[string]interface{}, error)
}

// PipelineStatus represents pipeline status
type PipelineStatus string

const (
	PipelineStatusCreating  PipelineStatus = "creating"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSuccess   PipelineStatus = "success"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCanceled  PipelineStatus = "canceled"
)

// RunnerPlatform represents runner platform type
type RunnerPlatform string

const (
	PlatformGitLab   RunnerPlatform = "gitlab"
	PlatformGitHub   RunnerPlatform = "github"
	PlatformJenkins  RunnerPlatform = "jenkins"
	PlatformCircleCI RunnerPlatform = "circleci"
	PlatformDrone    RunnerPlatform = "drone"
)

// RunnerStatus represents runner status
type RunnerStatus string

const (
	RunnerStatusCreating  RunnerStatus = "creating"
	RunnerStatusOnline    RunnerStatus = "online"
	RunnerStatusOffline   RunnerStatus = "offline"
	RunnerStatusBusy      RunnerStatus = "busy"
	RunnerStatusError     RunnerStatus = "error"
	RunnerStatusDestroyed RunnerStatus = "destroyed"
)

// VMState represents VM state
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

// PoolState represents pool state
type PoolState struct {
	Name         string `json:"name"`
	TemplatePath string `json:"template_path"`
	Capacity     int    `json:"capacity"`
	Available    int    `json:"available"`
	Busy         int    `json:"busy"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	VLAN     int      `json:"vlan"`
	CIDR     string   `json:"cidr"`
	Gateway  string   `json:"gateway"`
	DNS      []string `json:"dns"`
	Isolated bool     `json:"isolated"`
}

// NetworkState represents network state
type NetworkState struct {
	ID     string `json:"id"`
	VLAN   int    `json:"vlan"`
	CIDR   string `json:"cidr"`
	Status string `json:"status"`
}

// Runner represents a CI/CD runner
type Runner struct {
	ID           string         `json:"id"`
	Platform     RunnerPlatform `json:"platform"`
	Status       RunnerStatus   `json:"status"`
	Name         string         `json:"name"`
	Labels       []string       `json:"labels"`
	PipelineID   string         `json:"pipeline_id"`
	VMID         string         `json:"vm_id"`
	IPAddress    string         `json:"ip_address"`
	CurrentJob   string         `json:"current_job,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	DestroyedAt  *time.Time     `json:"destroyed_at,omitempty"`
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID           string         `json:"id"`
	Platform     RunnerPlatform `json:"platform"`
	Repository   string         `json:"repository"`
	Branch       string         `json:"branch"`
	CommitSHA    string         `json:"commit_sha"`
	CommitMsg    string         `json:"commit_message"`
	Author       string         `json:"author"`
	Status       PipelineStatus `json:"status"`
	NetworkID    string         `json:"network_id"`
	VMs          []string       `json:"vms"`
	Runners      []string       `json:"runners"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	Duration     int64          `json:"duration_seconds"`
	CurrentStage string         `json:"current_stage"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Stage represents a pipeline stage
type Stage struct {
	Name      string         `json:"name"`
	Status    PipelineStatus `json:"status"`
	Jobs      []Job          `json:"jobs"`
	StartTime *time.Time     `json:"start_time,omitempty"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
}

// Job represents a pipeline job
type Job struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Stage       string         `json:"stage"`
	Status      PipelineStatus `json:"status"`
	RunnerID    string         `json:"runner_id"`
	StartTime   *time.Time     `json:"start_time,omitempty"`
	EndTime     *time.Time     `json:"end_time,omitempty"`
	Duration    int64          `json:"duration_seconds"`
	Log         []string       `json:"log,omitempty"`
}

// Artifact represents a build artifact
type Artifact struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	Size        int64      `json:"size"`
	Checksum    string     `json:"checksum"`
	PipelineID  string     `json:"pipeline_id"`
	Downloads   int        `json:"downloads"`
	TTL         int        `json:"ttl_days"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}