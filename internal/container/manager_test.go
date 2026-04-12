// Package container provides Docker container management tests
package container

import (
	"context"
	"testing"
)

// TestNewContainerManager tests manager creation
func TestNewContainerManager(t *testing.T) {
	mgr, err := NewContainerManager()
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestContainerConfig tests container configuration
func TestContainerConfig(t *testing.T) {
	config := &ContainerConfig{
		Image:       "nginx:latest",
		Name:        "web-server",
		Cmd:         []string{"nginx", "-g", "daemon off;"},
		Env:         []string{"ENV=production"},
		NetworkMode: "bridge",
		CPU:         1024,
		Memory:      512 * 1024 * 1024,
		Ports: []PortBinding{
			{HostPort: "8080", ContainerPort: "80", Protocol: "tcp"},
		},
	}

	if config.Image != "nginx:latest" {
		t.Errorf("expected nginx:latest, got %s", config.Image)
	}
	if config.Name != "web-server" {
		t.Errorf("expected web-server, got %s", config.Name)
	}
	if config.NetworkMode != "bridge" {
		t.Errorf("expected bridge, got %s", config.NetworkMode)
	}
	if len(config.Ports) != 1 {
		t.Errorf("expected 1 port binding, got %d", len(config.Ports))
	}
	if config.Ports[0].HostPort != "8080" {
		t.Errorf("expected host port 8080, got %s", config.Ports[0].HostPort)
	}
}

// TestPortBinding tests port binding struct
func TestPortBinding(t *testing.T) {
	pb := PortBinding{
		HostPort:      "8080",
		ContainerPort: "80",
		Protocol:      "tcp",
	}

	if pb.HostPort != "8080" {
		t.Errorf("expected 8080, got %s", pb.HostPort)
	}
	if pb.ContainerPort != "80" {
		t.Errorf("expected 80, got %s", pb.ContainerPort)
	}
	if pb.Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", pb.Protocol)
	}
}

// TestCreateContainer tests container creation
func TestCreateContainer(t *testing.T) {
	mgr, _ := NewContainerManager()

	config := &ContainerConfig{
		Image: "alpine:latest",
		Name:  "test-container",
		Cmd:   []string{"sh", "-c", "echo hello"},
	}

	info, err := mgr.CreateContainer(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil container info")
	}
	if info.Name != "test-container" {
		t.Errorf("expected name 'test-container', got '%s'", info.Name)
	}
	if info.Status != "created" {
		t.Errorf("expected status 'created', got '%s'", info.Status)
	}
}

// TestStartContainer tests container start
func TestStartContainer(t *testing.T) {
	mgr, _ := NewContainerManager()

	err := mgr.StartContainer(context.Background(), "container-123")
	if err != nil {
		t.Errorf("failed to start container: %v", err)
	}
}

// TestStopContainer tests container stop
func TestStopContainer(t *testing.T) {
	mgr, _ := NewContainerManager()

	err := mgr.StopContainer(context.Background(), "container-123")
	if err != nil {
		t.Errorf("failed to stop container: %v", err)
	}
}

// TestRemoveContainer tests container removal
func TestRemoveContainer(t *testing.T) {
	mgr, _ := NewContainerManager()

	err := mgr.RemoveContainer(context.Background(), "container-123")
	if err != nil {
		t.Errorf("failed to remove container: %v", err)
	}
}

// TestGetContainer tests container retrieval
func TestGetContainer(t *testing.T) {
	mgr, _ := NewContainerManager()

	info, err := mgr.GetContainer(context.Background(), "container-123")
	if err != nil {
		t.Fatalf("failed to get container: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil container info")
	}
	if info.ID != "container-123" {
		t.Errorf("expected ID 'container-123', got '%s'", info.ID)
	}
}

// TestListContainers tests container listing
func TestListContainers(t *testing.T) {
	mgr, _ := NewContainerManager()

	containers, err := mgr.ListContainers(context.Background())
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}
	if containers == nil {
		t.Fatal("expected non-nil container list")
	}
	// Empty list is valid for stub
}

// TestContainerInfo tests container info struct
func TestContainerInfo(t *testing.T) {
	info := &ContainerInfo{
		ID:        "abc123",
		Name:      "nginx-server",
		Status:    "running",
		IPAddress: "172.17.0.2",
		Ports: []PortBinding{
			{HostPort: "80", ContainerPort: "80"},
		},
	}

	if info.ID != "abc123" {
		t.Errorf("expected abc123, got %s", info.ID)
	}
	if info.Status != "running" {
		t.Errorf("expected running, got %s", info.Status)
	}
	if info.IPAddress != "172.17.0.2" {
		t.Errorf("expected 172.17.0.2, got %s", info.IPAddress)
	}
}
