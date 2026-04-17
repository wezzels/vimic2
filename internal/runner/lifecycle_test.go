//go:build integration

package runner

import (
	"context"
	"testing"

	"github.com/stsgym/vimic2/internal/types"
)

// ==================== CircleCI Lifecycle ====================

func TestCircleCIRunner_Lifecycle(t *testing.T) {
	r, err := NewCircleCIRunner(&CircleCIConfig{
		APIToken: "test-token",
		Name:     "test-circleci",
		Labels:   []string{"docker"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Register(context.Background()); err != nil {
		t.Skipf("Register needs real CircleCI API: %v", err)
	}
	if r.Status() != types.RunnerStatusOffline {
		t.Errorf("Status after register = %s, want offline", r.Status())
	}

	if err := r.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOnline {
		t.Errorf("Status after start = %s, want online", r.Status())
	}
	if h := r.Health(); !h.Healthy {
		t.Error("Expected healthy after start")
	}

	if err := r.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOffline {
		t.Errorf("Status after stop = %s, want offline", r.Status())
	}
}

func TestCircleCIRunner_StartWithoutRegister(t *testing.T) {
	r, err := NewCircleCIRunner(&CircleCIConfig{
		APIToken: "test-token",
		Name:     "test-circleci",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r.Start(context.Background())
	if err == nil {
		t.Error("Expected error when starting unregistered runner")
	}
}

// ==================== Jenkins Lifecycle ====================

func TestJenkinsRunner_Lifecycle(t *testing.T) {
	r, err := NewJenkinsRunner(&JenkinsConfig{
		URL:      "http://jenkins.example.com",
		Username: "admin",
		APIToken: "test-token",
		Name:     "test-agent",
		Labels:   []string{"linux"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Register(context.Background()); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOffline {
		t.Errorf("Status after register = %s, want offline", r.Status())
	}

	if err := r.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOnline {
		t.Errorf("Status after start = %s, want online", r.Status())
	}

	if err := r.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestJenkinsRunner_StartWithoutRegister(t *testing.T) {
	r, err := NewJenkinsRunner(&JenkinsConfig{
		URL:      "http://jenkins.example.com",
		Username: "admin",
		APIToken: "test-token",
		Name:     "test-agent",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r.Start(context.Background())
	if err == nil {
		t.Error("Expected error when starting unregistered agent")
	}
}

// ==================== Drone Lifecycle ====================

func TestDroneRunner_Lifecycle(t *testing.T) {
	r, err := NewDroneRunner(&DroneConfig{
		URL:      "http://drone.example.com",
		APIToken: "test-token",
		Name:     "test-drone",
		Labels:   []string{"linux"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Register(context.Background()); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOffline {
		t.Errorf("Status after register = %s, want offline", r.Status())
	}

	if err := r.Start(context.Background()); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if r.Status() != types.RunnerStatusOnline {
		t.Errorf("Status after start = %s, want online", r.Status())
	}

	if err := r.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if err := r.Unregister(context.Background()); err != nil {
		t.Logf("Unregister: %v", err)
	}
}

func TestDroneRunner_StartWithoutRegister(t *testing.T) {
	r, err := NewDroneRunner(&DroneConfig{
		URL:      "http://drone.example.com",
		APIToken: "test-token",
		Name:     "test-drone",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = r.Start(context.Background())
	if err == nil {
		t.Error("Expected error when starting unregistered runner")
	}
}