// Package provisioner handles VM image and network management
package provisioner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Manager handles provisioning of VMs
type Manager struct {
	imageDir string
	networks map[string]*NetworkConfig
}

// NetworkConfig holds network configuration
type NetworkConfig struct {
	Name    string
	Type    string // nat, bridge
	CIDR    string
	Gateway string
}

// NewManager creates a new provisioner
func NewManager(imageDir string) *Manager {
	if imageDir == "" {
		imageDir = "/var/lib/libvirt/images"
	}
	return &Manager{
		imageDir: imageDir,
		networks: make(map[string]*NetworkConfig),
	}
}

// EnsureImage ensures a base image exists
func (m *Manager) EnsureImage(name, distro, version string) (string, error) {
	path := filepath.Join(m.imageDir, name+".qcow2")
	
	if _, err := os.Stat(path); err == nil {
		return path, nil // Exists
	}

	// Download or create image based on distro
	switch distro {
	case "ubuntu":
		return m.createUbuntuImage(path, version)
	case "debian":
		return m.createDebianImage(path, version)
	case "fedora":
		return m.createFedoraImage(path, version)
	case "centos":
		return m.createCentOSImage(path, version)
	default:
		return m.createGenericImage(path, distro, version)
	}
}

func (m *Manager) createUbuntuImage(path, version string) (string, error) {
	if version == "" {
		version = "22.04"
	}
	
	// Use cloud-init image
	url := fmt.Sprintf("https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img")
	
	// Download with curl
	cmd := exec.Command("curl", "-L", "-o", path, url)
	if err := cmd.Run(); err != nil {
		// Fallback: create empty qcow2
		cmd = exec.Command("qemu-img", "create", "-f", "qcow2", path, "20G")
		return path, cmd.Run()
	}
	return path, nil
}

func (m *Manager) createDebianImage(path, version string) (string, error) {
	if version == "" {
		version = "12"
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-b", 
		"/var/lib/libvirt/images/debian-base.qcow2", "-F", "qcow2", path)
	return path, cmd.Run()
}

func (m *Manager) createFedoraImage(path, version string) (string, error) {
	if version == "" {
		version = "39"
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-o", "size=20G", path)
	return path, cmd.Run()
}

func (m *Manager) createCentOSImage(path, version string) (string, error) {
	if version == "" {
		version = "9"
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-o", "size=20G", path)
	return path, cmd.Run()
}

func (m *Manager) createGenericImage(path, distro, version string) (string, error) {
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-o", "size=20G", path)
	return path, cmd.Run()
}

// CloneImage clones a base image for a new node
func (m *Manager) CloneImage(basePath, nodeName string) (string, error) {
	nodePath := filepath.Join(m.imageDir, nodeName+".qcow2")
	
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", "-b", basePath, "-F", "qcow2", nodePath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone image: %w", err)
	}
	
	return nodePath, nil
}

// ResizeImage resizes a disk image
func (m *Manager) ResizeImage(path string, sizeGB int) error {
	cmd := exec.Command("qemu-img", "resize", path, fmt.Sprintf("%dG", sizeGB))
	return cmd.Run()
}

// DeleteImage deletes a node's image
func (m *Manager) DeleteImage(nodeName string) error {
	path := filepath.Join(m.imageDir, nodeName+".qcow2")
	return os.Remove(path)
}

// NetworkManager handles libvirt network management
type NetworkManager struct{}

// EnsureNetwork ensures a libvirt network exists
func (nm *NetworkManager) EnsureNetwork(ctx context.Context, name, netType, cidr string) error {
	// Check if network exists
	cmd := exec.Command("virsh", "net-info", name)
	if err := cmd.Run(); err == nil {
		return nil // Network exists
	}

	// Create network XML
	xml := nm.generateNetworkXML(name, netType, cidr)
	
	// Define network
	cmd = exec.Command("virsh", "net-define", "--file", "-")
	cmd.Stdin = nil
	
	// Use virsh net-create for active creation
	cmd = exec.Command("virsh", "net-create", xml)
	return cmd.Run()
}

func (nm *NetworkManager) generateNetworkXML(name, netType, cidr string) string {
	if netType == "nat" {
		return fmt.Sprintf(`<network>
  <name>%s</name>
  <forward mode='nat'/>
  <bridge name='virbr%d' stp='on' delay='0'/>
  <ip address='192.168.%d.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.%d.2' end='192.168.%d.254'/>
    </dhcp>
  </ip>
</network>`, name, hash(name)%256, hash(name)%256, hash(name)%256, hash(name)%256)
	}
	// Bridge mode
	return fmt.Sprintf(`<network>
  <name>%s</name>
  <forward mode='bridge'/>
  <bridge name='br%s'/>
</network>`, name, name)
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % 256
}

// DeleteNetwork deletes a libvirt network
func (nm *NetworkManager) DeleteNetwork(ctx context.Context, name string) error {
	cmd := exec.Command("virsh", "net-destroy", name)
	cmd.Run() // Ignore error if doesn't exist
	cmd = exec.Command("virsh", "net-undefine", name)
	return cmd.Run()
}
