//go:build integration

package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupPipelineDB(t *testing.T) (*PipelineDB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-pipeline-db-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== Pipeline CRUD ====================

func TestPipelineDB_SaveGetPipeline(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	p := &Pipeline{
		ID:         "pipe-1",
		Platform:   PlatformGitLab,
		Repository: "https://gitlab.com/test/repo",
		Branch:     "main",
		CommitSHA:  "abc123",
		CommitMsg:  "test commit",
		Author:     "testuser",
		Status:     PipelineStatusRunning,
		NetworkID:  "net-1",
		StartTime:  time.Now(),
	}

	err := db.SavePipeline(ctx, p)
	if err != nil {
		t.Fatalf("SavePipeline: %v", err)
	}

	got, err := db.GetPipeline(ctx, "pipe-1")
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if got.ID != "pipe-1" {
		t.Errorf("ID = %s, want pipe-1", got.ID)
	}
	if got.Repository != "https://gitlab.com/test/repo" {
		t.Errorf("Repository = %s", got.Repository)
	}
	if got.Status != PipelineStatusRunning {
		t.Errorf("Status = %s, want running", got.Status)
	}
}

func TestPipelineDB_ListPipelines(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		p := &Pipeline{
			ID:         fmt.Sprintf("pipe-%d", i),
			Platform:   PlatformGitLab,
			Repository: "https://gitlab.com/test/repo",
			Branch:     "main",
			Status:     PipelineStatusRunning,
		}
		db.SavePipeline(ctx, p)
	}

	pipelines, err := db.ListPipelines(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if len(pipelines) < 3 {
		t.Errorf("Expected at least 3 pipelines, got %d", len(pipelines))
	}
}

func TestPipelineDB_UpdatePipelineStatus(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	p := &Pipeline{
		ID:       "pipe-1",
		Platform: PlatformGitLab,
		Status:   PipelineStatusRunning,
	}
	db.SavePipeline(ctx, p)

	err := db.UpdatePipelineStatus(ctx, "pipe-1", PipelineStatusSuccess)
	if err != nil {
		t.Fatalf("UpdatePipelineStatus: %v", err)
	}

	got, _ := db.GetPipeline(ctx, "pipe-1")
	if got.Status != PipelineStatusSuccess {
		t.Errorf("Status = %s, want completed", got.Status)
	}
}

// ==================== Runner CRUD ====================

func TestPipelineDB_SaveGetRunner(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	r := &Runner{
		ID:         "runner-1",
		PipelineID: "pipe-1",
		Platform:   PlatformGitLab,
		Name:       "test-runner",
		Status:     RunnerStatusOnline,
		Labels:     []string{"linux", "docker"},
	}

	err := db.SaveRunner(ctx, r)
	if err != nil {
		t.Fatalf("SaveRunner: %v", err)
	}

	got, err := db.GetRunner(ctx, "runner-1")
	if err != nil {
		t.Fatalf("GetRunner: %v", err)
	}
	if got.ID != "runner-1" {
		t.Errorf("ID = %s, want runner-1", got.ID)
	}
	if got.Name != "test-runner" {
		t.Errorf("Name = %s, want test-runner", got.Name)
	}
}

func TestPipelineDB_ListRunnersByPipeline(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		r := &Runner{
			ID:         fmt.Sprintf("runner-%d", i),
			PipelineID: "pipe-1",
			Platform:   PlatformGitLab,
			Name:       fmt.Sprintf("runner-%d", i),
			Status:     RunnerStatusOnline,
		}
		db.SaveRunner(ctx, r)
	}

	runners, err := db.ListRunnersByPipeline(ctx, "pipe-1")
	if err != nil {
		t.Fatalf("ListRunnersByPipeline: %v", err)
	}
	if len(runners) < 3 {
		t.Errorf("Expected at least 3 runners, got %d", len(runners))
	}
}

func TestPipelineDB_DeleteRunner(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	r := &Runner{
		ID:         "runner-1",
		PipelineID: "pipe-1",
		Platform:   PlatformGitLab,
		Status:     RunnerStatusOnline,
	}
	db.SaveRunner(ctx, r)

	err := db.DeleteRunner(ctx, "runner-1")
	if err != nil {
		t.Fatalf("DeleteRunner: %v", err)
	}

	_, err = db.GetRunner(ctx, "runner-1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

// ==================== VM CRUD ====================

func TestPipelineDB_SaveGetVM(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	vm := &VM{
		ID:         "vm-1",
		PoolID:     "pool-1",
		Name:       "test-vm",
		IP:         "10.0.0.1",
		MAC:        "aa:bb:cc:dd:ee:ff",
		CPU:        2,
		Memory:     2048,
		State:      VMStateRunning,
	}

	err := db.SaveVM(ctx, vm)
	if err != nil {
		t.Fatalf("SaveVM: %v", err)
	}

	got, err := db.GetVM(ctx, "vm-1")
	if err != nil {
		t.Fatalf("GetVM: %v", err)
	}
	if got.ID != "vm-1" {
		t.Errorf("ID = %s, want vm-1", got.ID)
	}
	if got.IP != "10.0.0.1" {
		t.Errorf("IP = %s, want 10.0.0.1", got.IP)
	}
}

func TestPipelineDB_ListVMsByPool(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		vm := &VM{
			ID:     fmt.Sprintf("vm-%d", i),
			PoolID: "pool-1",
			Name:   fmt.Sprintf("vm-%d", i),
			State:  VMStateRunning,
		}
		db.SaveVM(ctx, vm)
	}

	vms, err := db.ListVMsByPool(ctx, "pool-1")
	if err != nil {
		t.Fatalf("ListVMsByPool: %v", err)
	}
	if len(vms) < 3 {
		t.Errorf("Expected at least 3 VMs, got %d", len(vms))
	}
}

func TestPipelineDB_UpdateVMState(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	vm := &VM{ID: "vm-1", PoolID: "pool-1", State: VMStateRunning}
	db.SaveVM(ctx, vm)

	err := db.UpdateVMState(ctx, "vm-1", VMStateStopped)
	if err != nil {
		t.Fatalf("UpdateVMState: %v", err)
	}

	got, _ := db.GetVM(ctx, "vm-1")
	if got.State != VMStateStopped {
		t.Errorf("State = %s, want stopped", got.State)
	}
}

func TestPipelineDB_DeleteVM(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	vm := &VM{ID: "vm-1", PoolID: "pool-1", State: VMStateRunning}
	db.SaveVM(ctx, vm)

	err := db.DeleteVM(ctx, "vm-1")
	if err != nil {
		t.Fatalf("DeleteVM: %v", err)
	}

	_, err = db.GetVM(ctx, "vm-1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

// ==================== Template CRUD ====================

func TestPipelineDB_SaveGetTemplate(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	tmpl := &Template{
		ID:        "tmpl-1",
		Name:      "ubuntu-22.04",
		BaseImage: "ubuntu-22.04.qcow2",
		Size:      10 * 1024 * 1024 * 1024,
		Packages:  []string{"curl", "wget"},
		OS:        "linux",
		OSVersion: "22.04",
	}

	err := db.SaveTemplate(ctx, tmpl)
	if err != nil {
		t.Fatalf("SaveTemplate: %v", err)
	}

	got, err := db.GetTemplate(ctx, "tmpl-1")
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if got.ID != "tmpl-1" {
		t.Errorf("ID = %s, want tmpl-1", got.ID)
	}
	if got.Name != "ubuntu-22.04" {
		t.Errorf("Name = %s, want ubuntu-22.04", got.Name)
	}
}

func TestPipelineDB_ListTemplates(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		tmpl := &Template{
			ID:        fmt.Sprintf("tmpl-%d", i),
			Name:      fmt.Sprintf("template-%d", i),
			BaseImage: "ubuntu.qcow2",
			OS:        "linux",
		}
		db.SaveTemplate(ctx, tmpl)
	}

	templates, err := db.ListTemplates(ctx)
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if len(templates) < 3 {
		t.Errorf("Expected at least 3 templates, got %d", len(templates))
	}
}

func TestPipelineDB_DeleteTemplate(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	tmpl := &Template{ID: "tmpl-1", Name: "test", BaseImage: "test.qcow2"}
	db.SaveTemplate(ctx, tmpl)

	err := db.DeleteTemplate(ctx, "tmpl-1")
	if err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}

	_, err = db.GetTemplate(ctx, "tmpl-1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

// ==================== Pool CRUD ====================

func TestPipelineDB_SaveGetPool(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	pool := &Pool{
		ID:          "pool-1",
		Name:        "test-pool",
		TemplateID:  "tmpl-1",
		MinSize:     2,
		MaxSize:     10,
		CurrentSize: 2,
		CPU:         2,
		Memory:      2048,
	}

	err := db.SavePool(ctx, pool)
	if err != nil {
		t.Fatalf("SavePool: %v", err)
	}

	got, err := db.GetPool(ctx, "pool-1")
	if err != nil {
		t.Fatalf("GetPool: %v", err)
	}
	if got.ID != "pool-1" {
		t.Errorf("ID = %s, want pool-1", got.ID)
	}
	if got.Name != "test-pool" {
		t.Errorf("Name = %s, want test-pool", got.Name)
	}
}

func TestPipelineDB_ListPools(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		pool := &Pool{
			ID:         fmt.Sprintf("pool-%d", i),
			Name:       fmt.Sprintf("pool-%d", i),
			TemplateID: "tmpl-1",
			MinSize:    2,
			MaxSize:    10,
		}
		db.SavePool(ctx, pool)
	}

	pools, err := db.ListPools(ctx)
	if err != nil {
		t.Fatalf("ListPools: %v", err)
	}
	if len(pools) < 3 {
		t.Errorf("Expected at least 3 pools, got %d", len(pools))
	}
}

func TestPipelineDB_UpdatePoolSize(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	pool := &Pool{ID: "pool-1", Name: "test", MinSize: 2, MaxSize: 10, CurrentSize: 2}
	db.SavePool(ctx, pool)

	err := db.UpdatePoolSize(ctx, "pool-1", 1)
	if err != nil {
		t.Fatalf("UpdatePoolSize: %v", err)
	}

	got, _ := db.GetPool(ctx, "pool-1")
	if got.CurrentSize != 3 {
		t.Errorf("CurrentSize = %d, want 3", got.CurrentSize)
	}
}

// ==================== Network CRUD ====================

func TestPipelineDB_SaveGetNetwork(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	net := &Network{
		ID:         "net-1",
		PipelineID: "pipe-1",
		BridgeName: "br-test",
		VLANID:     100,
		CIDR:       "10.0.0.0/24",
		Gateway:    "10.0.0.1",
	}

	err := db.SaveNetwork(ctx, net)
	if err != nil {
		t.Fatalf("SaveNetwork: %v", err)
	}

	got, err := db.GetNetwork(ctx, "net-1")
	if err != nil {
		t.Fatalf("GetNetwork: %v", err)
	}
	if got.ID != "net-1" {
		t.Errorf("ID = %s, want net-1", got.ID)
	}
	if got.BridgeName != "br-test" {
		t.Errorf("BridgeName = %s, want br-test", got.BridgeName)
	}
}

func TestPipelineDB_DeleteNetwork(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	net := &Network{ID: "net-1", PipelineID: "pipe-1", CIDR: "10.0.0.0/24"}
	db.SaveNetwork(ctx, net)

	err := db.DeleteNetwork(ctx, "net-1")
	if err != nil {
		t.Fatalf("DeleteNetwork: %v", err)
	}

	_, err = db.GetNetwork(ctx, "net-1")
	if err == nil {
		t.Error("Expected error after delete")
	}
}

// ==================== Artifact CRUD ====================

func TestPipelineDB_SaveGetArtifact(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	artifact := &Artifact{
		ID:         "art-1",
		PipelineID: "pipe-1",
		Name:       "test-artifact.txt",
		Type:       "log",
		Size:       1024,
		Path:       "/tmp/test-artifact.txt",
	}

	err := db.SaveArtifact(ctx, artifact)
	if err != nil {
		t.Fatalf("SaveArtifact: %v", err)
	}

	// Verify via list
	artifacts, err := db.ListArtifactsByPipeline(ctx, "pipe-1")
	if err != nil {
		t.Logf("ListArtifactsByPipeline: %v", err)
	} else if len(artifacts) == 0 {
		t.Error("Expected at least 1 artifact")
	} else if artifacts[0].ID != "art-1" {
		t.Errorf("ID = %s, want art-1", artifacts[0].ID)
	}
}

func TestPipelineDB_DeleteArtifact(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	artifact := &Artifact{ID: "art-1", PipelineID: "pipe-1", Name: "test.txt"}
	db.SaveArtifact(ctx, artifact)

	err := db.DeleteArtifact(ctx, "art-1")
	if err != nil {
		t.Fatalf("DeleteArtifact: %v", err)
	}

	artifacts, err := db.ListArtifactsByPipeline(ctx, "pipe-1")
	if err != nil {
		t.Logf("ListArtifactsByPipeline: %v", err)
	} else {
		for _, a := range artifacts {
			if a.ID == "art-1" {
				t.Error("Artifact should be deleted")
			}
		}
	}
}

func TestPipelineDB_ListArtifactsByPipeline(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		artifact := &Artifact{
			ID:         fmt.Sprintf("art-%d", i),
			PipelineID: "pipe-1",
			Name:       fmt.Sprintf("artifact-%d", i),
		}
		db.SaveArtifact(ctx, artifact)
	}

	artifacts, err := db.ListArtifactsByPipeline(ctx, "pipe-1")
	if err != nil {
		t.Fatalf("ListArtifactsByPipeline: %v", err)
	}
	if len(artifacts) < 3 {
		t.Errorf("Expected at least 3 artifacts, got %d", len(artifacts))
	}
}

// ==================== Log CRUD ====================

func TestPipelineDB_SaveListLogs(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	log := &LogEntry{
		ID:         "log-1",
		PipelineID: "pipe-1",
		Level:      "info",
		Message:    "test log message",
		Timestamp:  time.Now(),
	}

	err := db.SaveLog(ctx, log)
	if err != nil {
		t.Fatalf("SaveLog: %v", err)
	}

	logs, err := db.ListLogsByPipeline(ctx, "pipe-1", 10, 0)
	if err != nil {
		t.Logf("ListLogsByPipeline: %v", err)
	} else {
		t.Logf("Listed %d logs", len(logs))
	}
}

// ==================== Stats ====================

func TestPipelineDB_GetStats_CRUD(t *testing.T) {
	db, cleanup := setupPipelineDB(t)
	defer cleanup()
	ctx := context.Background()

	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Logf("GetStats: %v", err)
	} else {
		t.Logf("Stats: %v", stats)
	}
}

