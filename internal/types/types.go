// Package types provides shared types to break import cycles
package types

import "time"

// PipelineDB is a reference to the pipeline database
type PipelineDB interface {
	SavePipeline(id string, state map[string]interface{}) error
	LoadPipeline(id string) (map[string]interface{}, error)
	DeletePipeline(id string) error
	ListPipelines() ([]string, error)
}

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

// RunnerPlatform represents runner platform type
type RunnerPlatform string

const (
	PlatformGitLab   RunnerPlatform = "gitlab"
	PlatformGitHub   RunnerPlatform = "github"
	PlatformJenkins  RunnerPlatform = "jenkins"
	PlatformCircleCI RunnerPlatform = "circleci"
	PlatformDrone    RunnerPlatform = "drone"
)

// PipelineStatus represents pipeline status
type PipelineStatus string

const (
	StatusCreated   PipelineStatus = "created"
	StatusRunning   PipelineStatus = "running"
	StatusSuccess   PipelineStatus = "success"
	StatusFailed    PipelineStatus = "failed"
	StatusCancelled PipelineStatus = "cancelled"
)

// NetworkConfig represents network configuration
type NetworkConfig struct {
	VLAN     int      `json:"vlan"`
	CIDR     string   `json:"cidr"`
	Gateway  string   `json:"gateway"`
	DNS      []string `json:"dns"`
	Isolated bool     `json:"isolated"`
}

// PoolManagerInterface is the interface for pool management (breaks import cycle)
type PoolManagerInterface interface {
	AllocateVM(poolName string) (*VMState, error)
	ReleaseVM(vmID string) error
	GetPool(name string) (*PoolState, error)
	ListPools() ([]*PoolState, error)
}

// NetworkManagerInterface is the interface for network management (breaks import cycle)
type NetworkManagerInterface interface {
	CreateNetwork(config *NetworkConfig) (string, error)
	DestroyNetwork(networkID string) error
	GetNetwork(networkID string) (*NetworkConfig, error)
}

// RunnerManagerInterface is the interface for runner management (breaks import cycle)
type RunnerManagerInterface interface {
	CreateRunner(platform RunnerPlatform, config map[string]interface{}) (string, error)
	DestroyRunner(runnerID string) error
	GetRunner(runnerID string) (map[string]interface{}, error)
	ListRunners() ([]map[string]interface{}, error)
}

// RunnerState represents runner state
type RunnerState struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Platform      RunnerPlatform  `json:"platform"`
	Status        string          `json:"status"`
	Version       string          `json:"version"`
	Tags          []string        `json:"tags"`
	JobsCompleted int             `json:"jobs_completed"`
	SuccessRate   float64         `json:"success_rate"`
	LastJob       *time.Time      `json:"last_job,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// NetworkState represents network state
type NetworkState struct {
	ID     string `json:"id"`
	VLAN   int    `json:"vlan"`
	CIDR   string `json:"cidr"`
	Status string `json:"status"`
}

// StageConfig represents stage configuration
type StageConfig struct {
	Name     string   `json:"name"`
	Jobs     []JobConfig `json:"jobs"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// JobConfig represents job configuration
type JobConfig struct {
	Name        string            `json:"name"`
	Image       string            `json:"image,omitempty"`
	Commands    []string          `json:"commands"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
}