// Package kubernetes provides Kubernetes integration tests
package kubernetes

import (
	"context"
	"testing"
)

// TestNewManager tests manager creation
func TestNewManager(t *testing.T) {
	mgr, err := NewManager("", "default")
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if mgr.namespace != "default" {
		t.Errorf("expected namespace 'default', got '%s'", mgr.namespace)
	}
}

// TestPodConfig tests pod configuration
func TestPodConfig(t *testing.T) {
	config := &PodConfig{
		Name:          "test-pod",
		Image:         "nginx:latest",
		CPULimit:      "500m",
		MemoryLimit:   "512Mi",
		CPURequest:    "100m",
		MemoryRequest: "128Mi",
		Env:           []string{"KEY=value"},
		Command:       []string{"/bin/sh"},
		Args:          []string{"-c", "echo hello"},
	}

	if config.Name != "test-pod" {
		t.Errorf("expected test-pod, got %s", config.Name)
	}
	if config.Image != "nginx:latest" {
		t.Errorf("expected nginx:latest, got %s", config.Image)
	}
	if config.CPULimit != "500m" {
		t.Errorf("expected 500m, got %s", config.CPULimit)
	}
	if len(config.Env) != 1 {
		t.Errorf("expected 1 env var, got %d", len(config.Env))
	}
}

// TestCreatePod tests pod creation
func TestCreatePod(t *testing.T) {
	mgr, _ := NewManager("", "test-ns")

	config := &PodConfig{
		Name:  "nginx",
		Image: "nginx:latest",
	}

	pod, err := mgr.CreatePod(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}
	if pod == nil {
		t.Fatal("expected non-nil pod")
	}
	if pod.Name != "nginx" {
		t.Errorf("expected name 'nginx', got '%s'", pod.Name)
	}
	if pod.Namespace != "test-ns" {
		t.Errorf("expected namespace 'test-ns', got '%s'", pod.Namespace)
	}
}

// TestGetPod tests getting pod info
func TestGetPod(t *testing.T) {
	mgr, _ := NewManager("", "default")

	pod, err := mgr.GetPod(context.Background(), "test-pod")
	if err != nil {
		t.Fatalf("failed to get pod: %v", err)
	}
	if pod == nil {
		t.Fatal("expected non-nil pod")
	}
	if pod.Name != "test-pod" {
		t.Errorf("expected name 'test-pod', got '%s'", pod.Name)
	}
}

// TestDeletePod tests pod deletion
func TestDeletePod(t *testing.T) {
	mgr, _ := NewManager("", "default")

	err := mgr.DeletePod(context.Background(), "test-pod")
	if err != nil {
		t.Errorf("failed to delete pod: %v", err)
	}
}

// TestListPods tests listing pods
func TestListPods(t *testing.T) {
	mgr, _ := NewManager("", "default")

	pods, err := mgr.ListPods(context.Background())
	if err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
	if pods == nil {
		t.Fatal("expected non-nil pod list")
	}
	// Empty list is valid for stub
}

// TestPodInfo tests pod info struct
func TestPodInfo(t *testing.T) {
	info := &PodInfo{
		Name:      "web-server",
		Namespace: "production",
		Status:    "Running",
		IP:        "10.0.0.50",
	}

	if info.Name != "web-server" {
		t.Errorf("expected web-server, got %s", info.Name)
	}
	if info.Status != "Running" {
		t.Errorf("expected Running, got %s", info.Status)
	}
	if info.IP != "10.0.0.50" {
		t.Errorf("expected 10.0.0.50, got %s", info.IP)
	}
}

// TestManager_Namespace tests namespace getter
func TestManager_Namespace(t *testing.T) {
	tests := []string{"default", "kube-system", "production", "test"}

	for _, ns := range tests {
		t.Run(ns, func(t *testing.T) {
			mgr, _ := NewManager("", ns)
			if mgr.namespace != ns {
				t.Errorf("expected namespace %s, got %s", ns, mgr.namespace)
			}
		})
	}
}