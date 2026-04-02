// Package pipeline provides cloud-init generation
package pipeline

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloudInitGenerator generates cloud-init configuration
type CloudInitGenerator struct {
	sshKeyManager *SSHKeyManager
}

// NewCloudInitGenerator creates a new cloud-init generator
func NewCloudInitGenerator(sshKeyManager *SSHKeyManager) *CloudInitGenerator {
	return &CloudInitGenerator{
		sshKeyManager: sshKeyManager,
	}
}

// CloudInitConfig represents cloud-init configuration
type CloudInitConfig struct {
	Hostname      string
	Username      string
	SSHKeys       []string
	Packages      []string
	RunCommands   []string
	WriteFiles    []WriteFileConfig
	Environment   map[string]string
	Platform      RunnerPlatform
	PipelineID    string
	VMID          string
	WorkDir       string
}

// WriteFileConfig represents a file to write
type WriteFileConfig struct {
	Path        string
	Content     string
	Permissions string
	Owner       string
}

// GenerateUserData generates cloud-init user-data
func (g *CloudInitGenerator) GenerateUserData(config *CloudInitConfig) (string, error) {
	// Build user-data
	var sb strings.Builder

	sb.WriteString("#cloud-config\n\n")

	// Set hostname
	if config.Hostname != "" {
		sb.WriteString(fmt.Sprintf("hostname: %s\n", config.Hostname))
		sb.WriteString("manage_etc_hosts: true\n\n")
	}

	// Users
	sb.WriteString("users:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s\n", config.Username))
	sb.WriteString("    sudo: ALL=(ALL) NOPASSWD:ALL\n")
	sb.WriteString("    shell: /bin/bash\n")
	sb.WriteString("    groups: [sudo, docker]\n")
	sb.WriteString("    lock_passwd: false\n")

	// SSH keys
	sb.WriteString("    ssh_authorized_keys:\n")
	for _, key := range config.SSHKeys {
		sb.WriteString(fmt.Sprintf("      - %s\n", key))
	}
	if len(config.SSHKeys) == 0 && g.sshKeyManager != nil {
		sb.WriteString(fmt.Sprintf("      - %s\n", strings.TrimSpace(string(g.sshKeyManager.GetPublicKey()))))
	}
	sb.WriteString("\n")

	// Password authentication disabled
	sb.WriteString("ssh_pwauth: false\n\n")

	// Packages
	if len(config.Packages) > 0 {
		sb.WriteString("packages:\n")
		for _, pkg := range config.Packages {
			sb.WriteString(fmt.Sprintf("  - %s\n", pkg))
		}
		sb.WriteString("\n")
	}

	// Write files
	if len(config.WriteFiles) > 0 {
		sb.WriteString("write_files:\n")
		for _, wf := range config.WriteFiles {
			sb.WriteString(fmt.Sprintf("  - path: %s\n", wf.Path))
			if wf.Permissions != "" {
				sb.WriteString(fmt.Sprintf("    permissions: '%s'\n", wf.Permissions))
			}
			if wf.Owner != "" {
				sb.WriteString(fmt.Sprintf("    owner: %s\n", wf.Owner))
			}
			sb.WriteString(fmt.Sprintf("    content: |\n"))
			for _, line := range strings.Split(wf.Content, "\n") {
				sb.WriteString(fmt.Sprintf("      %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	// Environment
	if len(config.Environment) > 0 {
		sb.WriteString("environment:\n")
		for key, value := range config.Environment {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Run commands
	sb.WriteString("runcmd:\n")
	sb.WriteString("  - systemctl enable qemu-guest-agent\n")
	sb.WriteString("  - systemctl start qemu-guest-agent\n")

	// Platform-specific commands
	switch config.Platform {
	case PlatformGitLab:
		sb.WriteString("  - curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh | bash\n")
		sb.WriteString("  - apt-get install -y gitlab-runner\n")
	case PlatformGitHub:
		sb.WriteString("  - curl -o actions-runner.tar.gz -L https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-linux-x64-2.321.0.tar.gz\n")
		sb.WriteString("  - tar xzf actions-runner.tar.gz\n")
	case PlatformJenkins:
		sb.WriteString("  - curl -o agent.jar -L https://jenkins.example.com/jnlpJars/agent.jar\n")
	case PlatformDrone:
		sb.WriteString("  - docker run -d -v /var/run/docker.sock:/var/run/docker.sock drone/drone-runner-docker:1\n")
	}

	// Create work directory
	if config.WorkDir != "" {
		sb.WriteString(fmt.Sprintf("  - mkdir -p %s\n", config.WorkDir))
		sb.WriteString(fmt.Sprintf("  - chown %s:%s %s\n", config.Username, config.Username, config.WorkDir))
	}

	// Custom run commands
	for _, cmd := range config.RunCommands {
		sb.WriteString(fmt.Sprintf("  - %s\n", cmd))
	}

	sb.WriteString("\n")

	// Final message
	sb.WriteString("final_message: 'VM $VM_ID is ready after $UPTIME seconds'\n")

	return sb.String(), nil
}

// GenerateMetaData generates cloud-init meta-data
func (g *CloudInitGenerator) GenerateMetaData(config *CloudInitConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("instance-id: %s\n", config.VMID))
	sb.WriteString(fmt.Sprintf("local-hostname: %s\n", config.Hostname))

	return sb.String(), nil
}

// GenerateVendorData generates cloud-init vendor-data
func (g *CloudInitGenerator) GenerateVendorData(config *CloudInitConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("#cloud-config\n\n")

	// Runner configuration
	sb.WriteString("write_files:\n")
	sb.WriteString("  - path: /etc/runner/config.yaml\n")
	sb.WriteString("    content: |\n")
	sb.WriteString(fmt.Sprintf("      platform: %s\n", config.Platform))
	sb.WriteString(fmt.Sprintf("      pipeline_id: %s\n", config.PipelineID))
	sb.WriteString(fmt.Sprintf("      vm_id: %s\n", config.VMID))
	if config.WorkDir != "" {
		sb.WriteString(fmt.Sprintf("      work_dir: %s\n", config.WorkDir))
	}
	sb.WriteString("\n")

	return sb.String(), nil
}

// CreateCloudInitISO creates a cloud-init ISO
func (g *CloudInitGenerator) CreateCloudInitISO(config *CloudInitConfig) (string, error) {
	// Generate user-data and meta-data
	userData, err := g.GenerateUserData(config)
	if err != nil {
		return "", fmt.Errorf("failed to generate user-data: %w", err)
	}

	metaData, err := g.GenerateMetaData(config)
	if err != nil {
		return "", fmt.Errorf("failed to generate meta-data: %w", err)
	}

	vendorData, err := g.GenerateVendorData(config)
	if err != nil {
		return "", fmt.Errorf("failed to generate vendor-data: %w", err)
	}

	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write files
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "user-data"), []byte(userData), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "meta-data"), []byte(metaData), 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "vendor-data"), []byte(vendorData), 0644); err != nil {
		return "", fmt.Errorf("failed to write vendor-data: %w", err)
	}

	// Generate ISO
	isoPath := filepath.Join("/tmp", fmt.Sprintf("%s-cloud-init.iso", config.VMID))

	// Try genisoimage first
	cmd := exec.Command("genisoimage",
		"-output", isoPath,
		"-volid", "cidata",
		"-joliet", "-rock",
		filepath.Join(tmpDir, "user-data"),
		filepath.Join(tmpDir, "meta-data"),
		filepath.Join(tmpDir, "vendor-data"),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		// Fallback to cloud-localds
		cmd = exec.Command("cloud-localds", isoPath,
			filepath.Join(tmpDir, "user-data"),
			filepath.Join(tmpDir, "meta-data"),
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to create cloud-init ISO: %w: %s", err, output)
		}
	}

	return isoPath, nil
}

// GetDefaultPackages returns default packages for a platform
func (g *CloudInitGenerator) GetDefaultPackages(platform RunnerPlatform) []string {
	// Common packages
	common := []string{
		"qemu-guest-agent",
		"curl",
		"wget",
		"git",
		"build-essential",
		"apt-transport-https",
		"ca-certificates",
		"gnupg",
		"lsb-release",
		"software-properties-common",
	}

	switch platform {
	case PlatformGitLab:
		return append(common, []string{
			"gitlab-runner",
		}...)
	case PlatformGitHub:
		return append(common, []string{
			// GitHub Actions runner is installed via script
		}...)
	case PlatformJenkins:
		return append(common, []string{
			"openjdk-17-jre-headless",
		}...)
	case PlatformDrone:
		return append(common, []string{
			"docker.io",
			"docker-compose",
		}...)
	default:
		return common
	}
}

// GetDefaultRunCommands returns default run commands for a platform
func (g *CloudInitGenerator) GetDefaultRunCommands(platform RunnerPlatform, config *CloudInitConfig) []string {
	commands := []string{
		"systemctl enable docker",
		"systemctl start docker",
		"usermod -aG docker " + config.Username,
	}

	switch platform {
	case PlatformGitLab:
		commands = append(commands,
			"curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh | bash",
			"apt-get install -y gitlab-runner",
		)
	case PlatformGitHub:
		commands = append(commands,
			"mkdir -p /work",
			"chown "+config.Username+":"+config.Username+" /work",
		)
	case PlatformDrone:
		commands = append(commands,
			"docker run -d -v /var/run/docker.sock:/var/run/docker.sock -e DRONE_RPC_HOST=drone.example.com drone/drone-runner-docker:1",
		)
	}

	return commands
}

// GeneratePlatformConfig generates platform-specific configuration
func (g *CloudInitGenerator) GeneratePlatformConfig(platform RunnerPlatform, pipelineID, vmID string) (*CloudInitConfig, error) {
	hostname := fmt.Sprintf("%s-%s", platform, vmID[:8])

	config := &CloudInitConfig{
		Hostname:    hostname,
		Username:    "runner",
		Platform:    platform,
		PipelineID:  pipelineID,
		VMID:        vmID,
		WorkDir:     "/work",
		Packages:    g.GetDefaultPackages(platform),
		RunCommands: g.GetDefaultRunCommands(platform, &CloudInitConfig{Username: "runner"}),
		Environment: map[string]string{
			"PLATFORM":    string(platform),
			"PIPELINE_ID": pipelineID,
			"VM_ID":       vmID,
		},
	}

	// Add SSH key
	if g.sshKeyManager != nil {
		config.SSHKeys = []string{string(g.sshKeyManager.GetPublicKey())}
	}

	return config, nil
}