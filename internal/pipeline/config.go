// Package pipeline provides configuration management
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the complete pipeline configuration
type Config struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	Hypervisor  HypervisorConfig `mapstructure:"hypervisor"`
	Templates   TemplatesConfig  `mapstructure:"templates"`
	Platforms   PlatformsConfig  `mapstructure:"platforms"`
	Pools       map[string]PoolConfig `mapstructure:"pools"`
	Networks    NetworksConfig   `mapstructure:"networks"`
	Logging     LoggingConfig    `mapstructure:"logging"`
	SSH         SSHConfig        `mapstructure:"ssh"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// HypervisorConfig represents hypervisor configuration
type HypervisorConfig struct {
	Type string `mapstructure:"type"`
	URI  string `mapstructure:"uri"`
}

// TemplatesConfig represents template configuration
type TemplatesConfig struct {
	BasePath       string `mapstructure:"base_path"`
	DefaultTemplate string `mapstructure:"default_template"`
}

// PlatformsConfig represents CI/CD platform configurations
type PlatformsConfig struct {
	GitLab   PlatformConfig `mapstructure:"gitlab"`
	GitHub   PlatformConfig `mapstructure:"github"`
	Jenkins  PlatformConfig `mapstructure:"jenkins"`
	CircleCI PlatformConfig `mapstructure:"circleci"`
	Drone    PlatformConfig `mapstructure:"drone"`
}

// PlatformConfig represents a single platform configuration
type PlatformConfig struct {
	URL               string   `mapstructure:"url"`
	RegistrationToken string   `mapstructure:"registration_token"`
	Token             string   `mapstructure:"token"`
	Labels            []string `mapstructure:"labels"`
	Enabled           bool     `mapstructure:"enabled"`
}

// PoolConfig represents a VM pool configuration
type PoolConfig struct {
	Template   string `mapstructure:"template"`
	MinSize    int    `mapstructure:"min_size"`
	MaxSize    int    `mapstructure:"max_size"`
	CPU        int    `mapstructure:"cpu"`
	Memory     int    `mapstructure:"memory"`
	DiskSize   int64  `mapstructure:"disk_size"`
	PreAllocated int  `mapstructure:"pre_allocated"`
}

// NetworksConfig represents network configuration
type NetworksConfig struct {
	BaseCIDR  string `mapstructure:"base_cidr"`
	VLANStart int    `mapstructure:"vlan_start"`
	VLANEnd   int    `mapstructure:"vlan_end"`
	DNS       []string `mapstructure:"dns"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	ElasticsearchURL string `mapstructure:"elasticsearch_url"`
	RetentionDays   int    `mapstructure:"retention_days"`
	Level           string `mapstructure:"level"`
}

// SSHConfig represents SSH key configuration
type SSHConfig struct {
	KeyPath     string `mapstructure:"key_path"`
	KeyType     string `mapstructure:"key_type"`
	KeySize     int    `mapstructure:"key_size"`
	Username    string `mapstructure:"username"`
	Port        int    `mapstructure:"port"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Set config file
	viper.SetConfigFile(configPath)
	
	// Set defaults
	setDefaults()
	
	// Read config
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	// Expand environment variables
	expandEnvironmentVariables()
	
	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.path", filepath.Join(os.Getenv("HOME"), ".vimic2", "pipeline.db"))
	
	// Hypervisor defaults
	viper.SetDefault("hypervisor.type", "libvirt")
	viper.SetDefault("hypervisor.uri", "qemu:///system")
	
	// Templates defaults
	viper.SetDefault("templates.base_path", "/var/lib/vimic2/templates")
	viper.SetDefault("templates.default_template", "base-runner-ubuntu-24.04.qcow2")
	
	// Network defaults
	viper.SetDefault("networks.base_cidr", "10.100.0.0/16")
	viper.SetDefault("networks.vlan_start", 1000)
	viper.SetDefault("networks.vlan_end", 2000)
	viper.SetDefault("networks.dns", []string{"8.8.8.8", "8.8.4.4"})
	
	// Logging defaults
	viper.SetDefault("logging.elasticsearch_url", "http://localhost:9200")
	viper.SetDefault("logging.retention_days", 30)
	viper.SetDefault("logging.level", "info")
	
	// SSH defaults
	viper.SetDefault("ssh.key_path", filepath.Join(os.Getenv("HOME"), ".vimic2", "keys"))
	viper.SetDefault("ssh.key_type", "ed25519")
	viper.SetDefault("ssh.key_size", 4096)
	viper.SetDefault("ssh.username", "root")
	viper.SetDefault("ssh.port", 22)
}

// expandEnvironmentVariables expands ${VAR} and $VAR in config
func expandEnvironmentVariables() {
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		if strings.Contains(val, "${") || strings.Contains(val, "$") {
			expanded := os.ExpandEnv(val)
			viper.Set(key, expanded)
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate database path
	if config.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}
	
	// Validate hypervisor
	if config.Hypervisor.Type != "libvirt" && config.Hypervisor.Type != "stub" {
		return fmt.Errorf("invalid hypervisor type: %s (must be libvirt or stub)", config.Hypervisor.Type)
	}
	
	// Validate templates path
	if config.Templates.BasePath == "" {
		return fmt.Errorf("templates base path is required")
	}
	
	// Validate at least one platform is enabled
	enabled := 0
	if config.Platforms.GitLab.Enabled {
		enabled++
	}
	if config.Platforms.GitHub.Enabled {
		enabled++
	}
	if config.Platforms.Jenkins.Enabled {
		enabled++
	}
	if config.Platforms.CircleCI.Enabled {
		enabled++
	}
	if config.Platforms.Drone.Enabled {
		enabled++
	}
	if enabled == 0 {
		return fmt.Errorf("at least one platform must be enabled")
	}
	
	// Validate pools
	if len(config.Pools) == 0 {
		return fmt.Errorf("at least one pool must be configured")
	}
	
	for name, pool := range config.Pools {
		if pool.Template == "" {
			return fmt.Errorf("pool %s: template is required", name)
		}
		if pool.MinSize < 0 {
			return fmt.Errorf("pool %s: min_size must be >= 0", name)
		}
		if pool.MaxSize < pool.MinSize {
			return fmt.Errorf("pool %s: max_size must be >= min_size", name)
		}
		if pool.CPU <= 0 {
			return fmt.Errorf("pool %s: cpu must be > 0", name)
		}
		if pool.Memory <= 0 {
			return fmt.Errorf("pool %s: memory must be > 0", name)
		}
	}
	
	// Validate networks
	if config.Networks.BaseCIDR == "" {
		return fmt.Errorf("networks base_cidr is required")
	}
	if config.Networks.VLANStart >= config.Networks.VLANEnd {
		return fmt.Errorf("networks vlan_start must be < vlan_end")
	}
	
	// Validate SSH
	if config.SSH.KeyPath == "" {
		return fmt.Errorf("ssh key_path is required")
	}
	if config.SSH.KeyType != "ed25519" && config.SSH.KeyType != "rsa" {
		return fmt.Errorf("ssh key_type must be ed25519 or rsa")
	}
	
	return nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Convert to viper
	viper.Set("database", config.Database)
	viper.Set("hypervisor", config.Hypervisor)
	viper.Set("templates", config.Templates)
	viper.Set("platforms", config.Platforms)
	viper.Set("pools", config.Pools)
	viper.Set("networks", config.Networks)
	viper.Set("logging", config.Logging)
	viper.Set("ssh", config.SSH)
	
	// Write config
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// GetPoolConfig returns configuration for a specific pool
func (c *Config) GetPoolConfig(poolName string) (*PoolConfig, error) {
	pool, ok := c.Pools[poolName]
	if !ok {
		return nil, fmt.Errorf("pool not found: %s", poolName)
	}
	return &pool, nil
}

// GetPlatformConfig returns configuration for a specific platform
func (c *Config) GetPlatformConfig(platformName string) (*PlatformConfig, error) {
	var config *PlatformConfig
	switch platformName {
	case "gitlab":
		config = &c.Platforms.GitLab
	case "github":
		config = &c.Platforms.GitHub
	case "jenkins":
		config = &c.Platforms.Jenkins
	case "circleci":
		config = &c.Platforms.CircleCI
	case "drone":
		config = &c.Platforms.Drone
	default:
		return nil, fmt.Errorf("unknown platform: %s", platformName)
	}
	
	if !config.Enabled {
		return nil, fmt.Errorf("platform %s is not enabled", platformName)
	}
	
	return config, nil
}

// ExampleConfig returns an example configuration
func ExampleConfig() string {
	return `# Vimic2 Pipeline Configuration

database:
  path: ~/.vimic2/pipeline.db

hypervisor:
  type: libvirt
  uri: qemu:///system

templates:
  base_path: /var/lib/vimic2/templates
  default_template: base-runner-ubuntu-24.04.qcow2

platforms:
  gitlab:
    url: https://gitlab.example.com
    registration_token: ${GITLAB_RUNNER_TOKEN}
    labels:
      - builder
      - go
      - docker
    enabled: true
  
  github:
    url: https://github.com
    token: ${GITHUB_RUNNER_TOKEN}
    labels:
      - builder
      - go
    enabled: false
  
  jenkins:
    url: https://jenkins.example.com
    token: ${JENKINS_TOKEN}
    labels:
      - builder
    enabled: false

pools:
  builder:
    template: base-go-1.23.qcow2
    min_size: 2
    max_size: 10
    cpu: 4
    memory: 8192
    disk_size: 50
    pre_allocated: 2
  
  tester:
    template: base-node-20.qcow2
    min_size: 1
    max_size: 5
    cpu: 2
    memory: 4096
    disk_size: 20
    pre_allocated: 1
  
  deployer:
    template: base-docker-27.qcow2
    min_size: 1
    max_size: 3
    cpu: 2
    memory: 4096
    disk_size: 30
    pre_allocated: 1

networks:
  base_cidr: 10.100.0.0/16
  vlan_start: 1000
  vlan_end: 2000
  dns:
    - 8.8.8.8
    - 8.8.4.4

logging:
  elasticsearch_url: http://localhost:9200
  retention_days: 30
  level: info

ssh:
  key_path: ~/.vimic2/keys
  key_type: ed25519
  key_size: 4096
  username: root
  port: 22
`
}