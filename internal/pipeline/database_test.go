// Package pipeline provides database tests
package pipeline

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

// TestDatabase_Open tests database opening
func TestDatabase_Open(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}

// TestPipeline_Create tests pipeline creation
func TestPipeline_Create(t *testing.T) {
	p := &Pipeline{
		ID:         "test-pipeline-1",
		Platform:   PlatformGitLab,
		Repository: "https://gitlab.example.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		CommitMsg:  "Test commit",
		Author:     "test@example.com",
		Status:     PipelineStatusCreating,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if p.ID != "test-pipeline-1" {
		t.Errorf("expected ID test-pipeline-1, got %s", p.ID)
	}
	if p.Platform != PlatformGitLab {
		t.Errorf("expected platform gitlab, got %s", p.Platform)
	}
	if p.Status != PipelineStatusCreating {
		t.Errorf("expected status creating, got %s", p.Status)
	}
}

// TestRunner_Create tests runner creation
func TestRunner_Create(t *testing.T) {
	r := &Runner{
		ID:         "runner-1",
		PipelineID: "pipeline-1",
		VMID:       "vm-1",
		Platform:   PlatformGitLab,
		PlatformID: "gitlab-runner-1",
		Labels:     []string{"docker", "linux"},
		Name:       "test-runner",
		Status:     RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}

	if r.ID != "runner-1" {
		t.Errorf("expected ID runner-1, got %s", r.ID)
	}
	if len(r.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(r.Labels))
	}
}

// TestVM_Create tests VM creation
func TestVM_Create(t *testing.T) {
	vm := &VM{
		ID:         "vm-1",
		PoolID:     "pool-1",
		TemplateID: "template-1",
		Name:       "test-vm-1",
		IP:         "10.100.1.10",
		State:      VMStateCreating,
		CPU:        4,
		Memory:     8192,
		CreatedAt:  time.Now(),
	}

	if vm.ID != "vm-1" {
		t.Errorf("expected ID vm-1, got %s", vm.ID)
	}
	if vm.CPU != 4 {
		t.Errorf("expected 4 CPUs, got %d", vm.CPU)
	}
	if vm.State != VMStateCreating {
		t.Errorf("expected state creating, got %s", vm.State)
	}
}

// TestNetwork_Create tests network creation
func TestNetwork_Create(t *testing.T) {
	n := &Network{
		ID:         "network-1",
		PipelineID: "pipeline-1",
		VLANID:     100,
		CIDR:       "10.100.1.0/24",
		Gateway:    "10.100.1.1",
		CreatedAt:  time.Now(),
	}

	if n.ID != "network-1" {
		t.Errorf("expected ID network-1, got %s", n.ID)
	}
	if n.VLANID != 100 {
		t.Errorf("expected VLAN 100, got %d", n.VLANID)
	}
	if n.CIDR != "10.100.1.0/24" {
		t.Errorf("expected CIDR 10.100.1.0/24, got %s", n.CIDR)
	}
}

// TestArtifact_Create tests artifact creation
func TestArtifact_Create(t *testing.T) {
	a := &types.Artifact{
		ID:         "artifact-1",
		PipelineID: "pipeline-1",
		Name:        "build.tar.gz",
		Path:        "/artifacts/pipeline-1/build.tar.gz",
		Size:        1024000,
		Checksum:    "sha256:abc123",
		TTL:         30,
		CreatedAt:   time.Now(),
	}

	if a.ID != "artifact-1" {
		t.Errorf("expected ID artifact-1, got %s", a.ID)
	}
	if a.Size != 1024000 {
		t.Errorf("expected size 1024000, got %d", a.Size)
	}
}

// TestPipelineStatus tests status transitions
func TestPipelineStatus_Transitions(t *testing.T) {
	statuses := []PipelineStatus{
		PipelineStatusCreating,
		PipelineStatusRunning,
		PipelineStatusSuccess,
	}

	for i, status := range statuses {
		if status == "" {
			t.Errorf("status %d is empty", i)
		}
	}

	// Test failure path
	p := &Pipeline{Status: PipelineStatusRunning}
	p.Status = PipelineStatusFailed
	if p.Status != PipelineStatusFailed {
		t.Error("failed to update status to failed")
	}

	// Test cancel path
	p.Status = PipelineStatusRunning
	p.Status = PipelineStatusCanceled
	if p.Status != PipelineStatusCanceled {
		t.Error("failed to update status to canceled")
	}
}

// TestRunnerStatus tests runner status
func TestRunnerStatus_Valid(t *testing.T) {
	validStatuses := []RunnerStatus{
		RunnerStatusCreating,
		RunnerStatusOnline,
		RunnerStatusBusy,
		RunnerStatusOffline,
		RunnerStatusError,
	}

	for _, status := range validStatuses {
		if status == "" {
			t.Errorf("empty status in valid list")
		}
	}
}

// TestVMState tests VM state transitions
func TestVMState_Transitions(t *testing.T) {
	vm := &VM{State: VMStateCreating}

	// Creating -> Running
	vm.State = VMStateRunning
	if vm.State != VMStateRunning {
		t.Error("failed to transition to running")
	}

	// Running -> Busy
	vm.State = VMStateBusy
	if vm.State != VMStateBusy {
		t.Error("failed to transition to busy")
	}

	// Busy -> Stopping
	vm.State = VMStateStopping
	if vm.State != VMStateStopping {
		t.Error("failed to transition to stopping")
	}

	// Stopping -> Stopped
	vm.State = VMStateStopped
	if vm.State != VMStateStopped {
		t.Error("failed to transition to stopped")
	}
}

// TestContext tests context usage
func TestContext_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		// Expected: context not expired yet
	case <-ctx.Done():
		t.Error("context expired too early")
	}

	// Wait for context to expire
	<-ctx.Done()
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
	}
}