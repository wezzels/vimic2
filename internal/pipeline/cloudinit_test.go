//go:build integration

package pipeline

import (
	"os"
	"testing"
)

// ==================== Cloud-Init Generator Tests ====================

func TestCloudInitGenerator_New(t *testing.T) {
	generator := NewCloudInitGenerator(nil)
	if generator == nil {
		t.Fatal("NewCloudInitGenerator returned nil")
	}
}

func TestCloudInitGenerator_DefaultPackages(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	platforms := []RunnerPlatform{PlatformGitLab, PlatformGitHub, PlatformJenkins, PlatformCircleCI, PlatformDrone}
	for _, platform := range platforms {
		packages := generator.GetDefaultPackages(platform)
		if len(packages) == 0 {
			t.Errorf("GetDefaultPackages(%s) returned empty list", platform)
		}
	}
}

func TestCloudInitGenerator_DefaultRunCommands(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	config := &CloudInitConfig{
		Username: "runner",
	}

	for _, platform := range []RunnerPlatform{PlatformGitLab, PlatformGitHub, PlatformJenkins} {
		commands := generator.GetDefaultRunCommands(platform, config)
		if len(commands) == 0 {
			t.Errorf("GetDefaultRunCommands(%s) returned empty list", platform)
		}
	}
}

func TestCloudInitGenerator_GeneratePlatformConfig(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	tests := []struct {
		name     string
		platform RunnerPlatform
		pipeline string
		vmID     string
	}{
	{"gitlab", PlatformGitLab, "pipeline-1-abc", "vm-1-defghijk"},
	{"github", PlatformGitHub, "pipeline-2-abc", "vm-2-defghijk"},
	{"jenkins", PlatformJenkins, "pipeline-3-abc", "vm-3-defghijk"},
	{"circleci", PlatformCircleCI, "pipeline-4-abc", "vm-4-defghijk"},
	{"drone", PlatformDrone, "pipeline-5-abc", "vm-5-defghijk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := generator.GeneratePlatformConfig(tt.platform, tt.pipeline, tt.vmID)
			if err != nil {
				t.Fatalf("GeneratePlatformConfig(%s) failed: %v", tt.platform, err)
			}
			if config == nil {
				t.Error("GeneratePlatformConfig should not return nil")
			}
		})
	}
}

func TestCloudInitGenerator_GenerateUserData(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	config := &CloudInitConfig{
		Hostname:    "test-runner-1",
		Username:    "runner",
		SSHKeys:     []string{"ssh-rsa AAAAB3NzaC1yc2E..."},
		Packages:    []string{"curl", "git"},
		RunCommands: []string{"echo hello", "apt-get update"},
	}

	userData, err := generator.GenerateUserData(config)
	if err != nil {
		t.Fatalf("GenerateUserData failed: %v", err)
	}

	if len(userData) == 0 {
		t.Error("GenerateUserData returned empty data")
	}
}

func TestCloudInitGenerator_GenerateMetaData(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	config := &CloudInitConfig{
		Hostname: "test-runner-1",
	}

	metaData, err := generator.GenerateMetaData(config)
	if err != nil {
		t.Fatalf("GenerateMetaData failed: %v", err)
	}

	if len(metaData) == 0 {
		t.Error("GenerateMetaData returned empty data")
	}
}

func TestCloudInitGenerator_CreateCloudInitISO(t *testing.T) {
	generator := NewCloudInitGenerator(nil)

	config := &CloudInitConfig{
		Hostname:  "test-runner-1",
		Username:  "runner",
		SSHKeys:   []string{"ssh-rsa AAAAB3NzaC1yc2E..."},
		Packages:  []string{"curl"},
	}

	isoPath, err := generator.CreateCloudInitISO(config)
	if err != nil {
		t.Fatalf("CreateCloudInitISO failed: %v", err)
	}

	if isoPath == "" {
		t.Error("CreateCloudInitISO should return a non-empty path")
	}

	// Verify the ISO was created
	if _, err := os.Stat(isoPath); os.IsNotExist(err) {
		t.Error("ISO file was not created at returned path")
	}
}