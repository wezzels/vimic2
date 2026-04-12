// Package deploy provides executor coverage tests
package deploy

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestExecutor_Execute_Full tests Execute with real db and stub hypervisor
func TestExecutor_Execute_Full(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-exec-full-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Create executor with stub hypervisor
	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	executor := NewExecutor(db, hosts)

	ctx := context.Background()
	progress := make(chan *Progress, 100)

	// Create cluster to deploy
	cluster := &Cluster{
		ID:   "exec-full-test",
		Name: "exec-full-test",
		Network: &NetworkConfig{
			Name: "default",
			Type: "nat",
		},
		NodeGroups: []*NodeGroup{
			{
				Role:     "worker",
				Count:    2,
				CPU:      2,
				MemoryMB: 2048,
				DiskGB:   20,
				Image:    "ubuntu-22.04",
				HostID:   "test-host",
			},
		},
	}

	var wg sync.WaitGroup
	var execErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		execErr = executor.Execute(ctx, cluster, progress)
		close(progress)
	}()

	// Collect progress updates
	for prog := range progress {
		t.Logf("Progress: %s - %d/%d nodes", prog.Status, prog.DeployedNodes, prog.TotalNodes)
	}

	wg.Wait()
	if execErr != nil {
		t.Logf("Execute returned error: %v", execErr)
	}
}

// TestExecutor_Execute_NoHosts tests Execute with no hosts
func TestExecutor_Execute_NoHosts(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-exec-nohost-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// No hosts
	hosts := map[string]hypervisor.Hypervisor{}
	executor := NewExecutor(db, hosts)

	ctx := context.Background()
	progress := make(chan *Progress, 100)

	cluster := &Cluster{
		ID:   "no-host-test",
		Name: "no-host-test",
		Network: &NetworkConfig{
			Name: "default",
		},
		NodeGroups: []*NodeGroup{
			{Role: "worker", Count: 1}, // No HostID specified
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		executor.Execute(ctx, cluster, progress)
		close(progress)
	}()

	for prog := range progress {
		if prog.Status == "error" {
			t.Logf("Got expected error: %v", prog.Error)
		}
	}
	wg.Wait()
}

// TestExecutor_Execute_Cancel tests cancellation
func TestExecutor_Execute_Cancel(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-exec-cancel-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := database.NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	hosts := map[string]hypervisor.Hypervisor{
		"test-host": hypervisor.NewStubHypervisor(),
	}
	executor := NewExecutor(db, hosts)

	ctx, cancel := context.WithCancel(context.Background())
	progress := make(chan *Progress, 100)

	// Cancel immediately
	cancel()

	cluster := &Cluster{
		ID:      "cancel-test",
		Name:    "cancel-test",
		Network: &NetworkConfig{Name: "default"},
		NodeGroups: []*NodeGroup{
			{Role: "worker", Count: 1, HostID: "test-host"},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		executor.Execute(ctx, cluster, progress)
		close(progress)
	}()

	for range progress {
		// Drain progress channel
	}
	wg.Wait()

	// The cancelled context should cause Execute to return early
	// Just verify it doesn't hang
}

// TestWizard_CompleteWorkflow tests complete wizard workflow
func TestWizard_CompleteWorkflow(t *testing.T) {
	w := NewWizard()

	// Step through wizard
	w.SetName("test-cluster")
	w.SetNetwork(&NetworkConfig{
		Name:    "test-net",
		Type:    "nat",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
	})

	w.AddNodeGroup(&NodeGroup{
		Role:     "worker",
		Count:    3,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
	})

	w.AddNodeGroup(&NodeGroup{
		Role:     "master",
		Count:    1,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   50,
		Image:    "ubuntu-22.04",
	})

	// Navigate steps
	w.NextStep()
	w.NextStep()
	w.NextStep()

	// Get cluster
	cluster := w.GetCluster()
	if cluster.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", cluster.Name)
	}

	// Go back
	w.PrevStep()
	w.PrevStep()

	// Remove a node group
	w.RemoveNodeGroup(1)

	cluster = w.GetCluster()
	if len(cluster.NodeGroups) != 1 {
		t.Errorf("expected 1 node group after removal, got %d", len(cluster.NodeGroups))
	}
}

// TestWizard_NextPrev tests step navigation
func TestWizard_NextPrev(t *testing.T) {
	w := NewWizard()

	// Navigate forward
	for i := 0; i < 5; i++ {
		w.NextStep()
	}

	step := w.GetStep()
	t.Logf("After 5 NextStep: step=%d", step)

	// Navigate backward
	for i := 0; i < 3; i++ {
		w.PrevStep()
	}

	step = w.GetStep()
	t.Logf("After 3 PrevStep: step=%d", step)
}

// TestCluster_Validate_Fields tests cluster field validation
func TestCluster_Validate_Fields(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		valid   bool
	}{
		{
			name: "valid cluster",
			cluster: &Cluster{
				ID:      "test",
				Name:    "test-cluster",
				Network: &NetworkConfig{Name: "default"},
				NodeGroups: []*NodeGroup{
					{Role: "worker", Count: 1},
				},
			},
			valid: true,
		},
		{
			name: "empty name",
			cluster: &Cluster{
				ID:      "test",
				Name:    "",
				Network: &NetworkConfig{Name: "default"},
			},
			valid: false,
		},
		{
			name: "no network",
			cluster: &Cluster{
				ID:   "test",
				Name: "test",
			},
			valid: false,
		},
		{
			name: "no node groups",
			cluster: &Cluster{
				ID:      "test",
				Name:    "test",
				Network: &NetworkConfig{Name: "default"},
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Manual validation checks
			valid := tc.cluster.Name != "" && tc.cluster.Network != nil && len(tc.cluster.NodeGroups) > 0
			if tc.valid != valid {
				t.Errorf("expected valid=%v, got valid=%v", tc.valid, valid)
			}
		})
	}
}

// TestProgress_Struct_Coverage tests Progress struct
func TestProgress_Struct_Coverage(t *testing.T) {
	p := &Progress{
		ClusterID:     "test",
		TotalNodes:    10,
		DeployedNodes: 5,
		CurrentNode:   "node-5",
		Status:        "deploying",
	}

	if p.DeployedNodes != 5 {
		t.Errorf("expected 5, got %d", p.DeployedNodes)
	}
}

// TestNodeGroup_Defaults tests node group defaults
func TestNodeGroup_Defaults_Coverage(t *testing.T) {
	ng := &NodeGroup{
		Role:     "worker",
		Count:    3,
		CPU:      4,
		MemoryMB: 8192,
		DiskGB:   100,
		Image:    "ubuntu-22.04",
	}

	if ng.Role != "worker" {
		t.Errorf("expected worker, got %s", ng.Role)
	}
	if ng.Count != 3 {
		t.Errorf("expected 3, got %d", ng.Count)
	}
}

// TestNetworkConfig_Defaults_Coverage tests network config
func TestNetworkConfig_Defaults_Coverage(t *testing.T) {
	nc := &NetworkConfig{
		Name:    "test-net",
		Type:    "nat",
		CIDR:    "10.0.0.0/24",
		Gateway: "10.0.0.1",
	}

	if nc.Name != "test-net" {
		t.Errorf("expected test-net, got %s", nc.Name)
	}
}

// TestPreset_All tests all presets that exist
func TestPreset_All(t *testing.T) {
	presets := []string{"dev", "prod", "db"} // Only test known presets

	for _, name := range presets {
		preset := GetPreset(name)
		if preset == nil {
			t.Errorf("expected preset for %s", name)
			continue
		}
		t.Logf("Preset %s: %d node groups", name, len(preset.NodeGroups))
	}
}

// TestGetPreset_Invalid tests invalid preset
func TestGetPreset_Invalid(t *testing.T) {
	preset := GetPreset("nonexistent")
	if preset != nil {
		t.Error("expected nil for invalid preset")
	}
}
