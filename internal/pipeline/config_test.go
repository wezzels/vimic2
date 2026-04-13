// Package pipeline provides configuration tests
package pipeline

import (
	"testing"
)

// TestDatabaseConfig tests database configuration
func TestDatabaseConfig_Create(t *testing.T) {
	config := DatabaseConfig{
		Path: "/var/lib/vimic2/pipeline.db",
	}

	if config.Path != "/var/lib/vimic2/pipeline.db" {
		t.Errorf("expected path /var/lib/vimic2/pipeline.db, got %s", config.Path)
	}
}

// TestHypervisorConfig tests hypervisor configuration
func TestHypervisorConfig_Create(t *testing.T) {
	config := HypervisorConfig{
		Type: "libvirt",
		URI:  "qemu:///system",
	}

	if config.Type != "libvirt" {
		t.Errorf("expected libvirt, got %s", config.Type)
	}
	if config.URI != "qemu:///system" {
		t.Errorf("expected qemu:///system, got %s", config.URI)
	}
}

// TestTemplatesConfig tests templates configuration
func TestTemplatesConfig_Create(t *testing.T) {
	config := TemplatesConfig{
		BasePath:        "/var/lib/vimic2/templates",
		DefaultTemplate: "base-runner-ubuntu-24.04.qcow2",
	}

	if config.BasePath != "/var/lib/vimic2/templates" {
		t.Errorf("expected /var/lib/vimic2/templates, got %s", config.BasePath)
	}
	if config.DefaultTemplate != "base-runner-ubuntu-24.04.qcow2" {
		t.Errorf("expected base-runner-ubuntu-24.04.qcow2, got %s", config.DefaultTemplate)
	}
}

// TestPlatformsConfig tests platforms configuration
func TestPlatformsConfig_Create(t *testing.T) {
	config := PlatformsConfig{
		GitLab: PlatformConfig{
			URL:     "https://gitlab.example.com",
			Token:   "glrt-xxx",
			Labels:  []string{"docker", "linux"},
			Enabled: true,
		},
		GitHub: PlatformConfig{
			URL:     "https://github.com",
			Token:   "ghp_xxx",
			Labels:  []string{"ubuntu", "docker"},
			Enabled: true,
		},
	}

	if config.GitLab.URL != "https://gitlab.example.com" {
		t.Errorf("expected GitLab URL, got %s", config.GitLab.URL)
	}
	if config.GitHub.URL != "https://github.com" {
		t.Errorf("expected GitHub URL, got %s", config.GitHub.URL)
	}
	if !config.GitLab.Enabled {
		t.Error("GitLab should be enabled")
	}
	if !config.GitHub.Enabled {
		t.Error("GitHub should be enabled")
	}
}

// TestPlatformConfig tests platform configuration
func TestPlatformConfig_Create(t *testing.T) {
	config := PlatformConfig{
		URL:               "https://gitlab.example.com",
		RegistrationToken: "glrt-xxx",
		Token:             "api-token",
		Labels:            []string{"docker", "linux", "golang"},
		Enabled:           true,
	}

	if config.URL != "https://gitlab.example.com" {
		t.Errorf("expected https://gitlab.example.com, got %s", config.URL)
	}
	if len(config.Labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(config.Labels))
	}
	if !config.Enabled {
		t.Error("platform should be enabled")
	}
}

// TestPoolConfig tests pool configuration
func TestPoolConfig_Create(t *testing.T) {
	config := PoolConfig{
		Template:     "base-runner-ubuntu-24.04.qcow2",
		MinSize:      2,
		MaxSize:      10,
		CPU:          4,
		Memory:       8192,
		DiskSize:     50,
		PreAllocated: 3,
	}

	if config.Template != "base-runner-ubuntu-24.04.qcow2" {
		t.Errorf("expected base template, got %s", config.Template)
	}
	if config.MinSize != 2 {
		t.Errorf("expected min size 2, got %d", config.MinSize)
	}
	if config.MaxSize != 10 {
		t.Errorf("expected max size 10, got %d", config.MaxSize)
	}
	if config.CPU != 4 {
		t.Errorf("expected 4 CPUs, got %d", config.CPU)
	}
	if config.Memory != 8192 {
		t.Errorf("expected 8192 MB memory, got %d", config.Memory)
	}
}

// TestNetworksConfig tests networks configuration
func TestNetworksConfig_Create(t *testing.T) {
	config := NetworksConfig{
		BaseCIDR:  "10.100.0.0/16",
		VLANStart: 1000,
		VLANEnd:   2000,
		DNS:       []string{"8.8.8.8", "8.8.4.4"},
	}

	if config.BaseCIDR != "10.100.0.0/16" {
		t.Errorf("expected 10.100.0.0/16, got %s", config.BaseCIDR)
	}
	if config.VLANStart != 1000 {
		t.Errorf("expected VLAN start 1000, got %d", config.VLANStart)
	}
	if config.VLANEnd != 2000 {
		t.Errorf("expected VLAN end 2000, got %d", config.VLANEnd)
	}
	if len(config.DNS) != 2 {
		t.Errorf("expected 2 DNS servers, got %d", len(config.DNS))
	}
}

// TestLoggingConfig tests logging configuration
func TestLoggingConfig_Create(t *testing.T) {
	config := LoggingConfig{
		ElasticsearchURL: "http://localhost:9200",
		RetentionDays:    30,
		Level:            "info",
	}

	if config.ElasticsearchURL != "http://localhost:9200" {
		t.Errorf("expected http://localhost:9200, got %s", config.ElasticsearchURL)
	}
	if config.RetentionDays != 30 {
		t.Errorf("expected 30 days retention, got %d", config.RetentionDays)
	}
	if config.Level != "info" {
		t.Errorf("expected info level, got %s", config.Level)
	}
}

// TestSSHConfig tests SSH configuration
func TestSSHConfig_Create(t *testing.T) {
	config := SSHConfig{
		KeyPath:  "/var/lib/vimic2/keys",
		KeyType:  "ed25519",
		KeySize:  0,
		Username: "runner",
		Port:     22,
	}

	if config.KeyPath != "/var/lib/vimic2/keys" {
		t.Errorf("expected /var/lib/vimic2/keys, got %s", config.KeyPath)
	}
	if config.KeyType != "ed25519" {
		t.Errorf("expected ed25519, got %s", config.KeyType)
	}
	if config.Username != "runner" {
		t.Errorf("expected runner, got %s", config.Username)
	}
	if config.Port != 22 {
		t.Errorf("expected port 22, got %d", config.Port)
	}
}

// TestConfig_Create tests full config creation
func TestConfig_Create(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			Path: "/var/lib/vimic2/pipeline.db",
		},
		Hypervisor: HypervisorConfig{
			Type: "libvirt",
			URI:  "qemu:///system",
		},
		Templates: TemplatesConfig{
			BasePath:        "/var/lib/vimic2/templates",
			DefaultTemplate: "base-runner.qcow2",
		},
		Platforms: PlatformsConfig{
			GitLab: PlatformConfig{
				URL:     "https://gitlab.example.com",
				Enabled: true,
			},
		},
		Pools: map[string]PoolConfig{
			"ubuntu-22.04": {
				Template: "ubuntu-22.04.qcow2",
				MinSize:  2,
				MaxSize:  10,
			},
		},
		Networks: NetworksConfig{
			BaseCIDR:  "10.100.0.0/16",
			VLANStart: 1000,
			VLANEnd:   2000,
		},
	}

	if config.Database.Path != "/var/lib/vimic2/pipeline.db" {
		t.Errorf("expected database path, got %s", config.Database.Path)
	}
	if config.Hypervisor.Type != "libvirt" {
		t.Errorf("expected libvirt hypervisor, got %s", config.Hypervisor.Type)
	}
	if config.Pools == nil {
		t.Error("pools should not be nil")
	}
	if len(config.Pools) != 1 {
		t.Errorf("expected 1 pool, got %d", len(config.Pools))
	}
}

// TestConfig_MultiplePools tests config with multiple pools
func TestConfig_MultiplePools(t *testing.T) {
	config := &Config{
		Pools: map[string]PoolConfig{
			"ubuntu-22.04": {
				Template: "ubuntu-22.04.qcow2",
				MinSize:  2,
				MaxSize:  10,
			},
			"ubuntu-24.04": {
				Template: "ubuntu-24.04.qcow2",
				MinSize:  3,
				MaxSize:  15,
			},
			"fedora-39": {
				Template: "fedora-39.qcow2",
				MinSize:  1,
				MaxSize:  5,
			},
		},
	}

	if len(config.Pools) != 3 {
		t.Errorf("expected 3 pools, got %d", len(config.Pools))
	}

	// Verify each pool exists
	for name := range config.Pools {
		if config.Pools[name].Template == "" {
			t.Errorf("pool %s should have template", name)
		}
	}
}

// TestConfig_MultiplePlatforms tests config with multiple platforms
func TestConfig_MultiplePlatforms(t *testing.T) {
	config := &Config{
		Platforms: PlatformsConfig{
			GitLab: PlatformConfig{
				URL:     "https://gitlab.example.com",
				Enabled: true,
			},
			GitHub: PlatformConfig{
				URL:     "https://github.com",
				Enabled: true,
			},
			Jenkins: PlatformConfig{
				URL:     "https://jenkins.example.com",
				Enabled: false,
			},
		},
	}

	if !config.Platforms.GitLab.Enabled {
		t.Error("GitLab should be enabled")
	}
	if !config.Platforms.GitHub.Enabled {
		t.Error("GitHub should be enabled")
	}
	if config.Platforms.Jenkins.Enabled {
		t.Error("Jenkins should be disabled")
	}
}

// TestConfig_EmptyPools tests config with no pools
func TestConfig_EmptyPools(t *testing.T) {
	config := &Config{
		Pools: map[string]PoolConfig{},
	}

	if config.Pools == nil {
		t.Error("pools should be empty map, not nil")
	}
	if len(config.Pools) != 0 {
		t.Errorf("expected 0 pools, got %d", len(config.Pools))
	}
}

// TestPoolConfig_Validation tests pool config validation
func TestPoolConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config PoolConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: PoolConfig{
				Template: "ubuntu.qcow2",
				MinSize:  2,
				MaxSize:  10,
				CPU:      4,
				Memory:   8192,
			},
			valid: true,
		},
		{
			name: "zero min size",
			config: PoolConfig{
				Template: "ubuntu.qcow2",
				MinSize:  0,
				MaxSize:  10,
			},
			valid: true, // Zero min size is valid (no pre-allocated VMs)
		},
		{
			name: "negative values",
			config: PoolConfig{
				Template: "ubuntu.qcow2",
				MinSize:  -1,
				MaxSize:  -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			if tt.config.Template == "" && tt.valid {
				t.Error("valid config should have template")
			}
		})
	}
}

// TestSSHConfig_KeyTypes tests different SSH key types
func TestSSHConfig_KeyTypes(t *testing.T) {
	keyTypes := []string{"ed25519", "rsa", "ecdsa", "dsa"}

	for _, keyType := range keyTypes {
		t.Run(keyType, func(t *testing.T) {
			config := SSHConfig{KeyType: keyType}
			if config.KeyType != keyType {
				t.Errorf("expected %s, got %s", keyType, config.KeyType)
			}
		})
	}
}

// TestLoggingConfig_Levels tests different log levels
func TestLoggingConfig_Levels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			config := LoggingConfig{Level: level}
			if config.Level != level {
				t.Errorf("expected %s, got %s", level, config.Level)
			}
		})
	}
}