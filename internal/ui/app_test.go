// Package ui provides the Fyne UI implementation tests
package ui

import (
	"testing"

	"github.com/stsgym/vimic2/internal/database"
)

// TestAppCreation tests app struct creation
func TestAppCreation(t *testing.T) {
	app := &App{}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

// TestApp_Fields tests app field defaults
func TestApp_Fields(t *testing.T) {
	app := &App{}

	// Fields should be nil by default
	if app.db != nil {
		t.Error("expected db to be nil")
	}
	if app.window != nil {
		t.Error("expected window to be nil")
	}
	if app.clusters != nil {
		t.Error("expected clusters to be nil")
	}
}

// TestApp_SelectedCluster tests selected cluster management
func TestApp_SelectedCluster(t *testing.T) {
	app := &App{}

	// No selection initially
	if app.selectedCluster != nil {
		t.Error("expected no selected cluster")
	}

	// Set selection
	app.selectedCluster = &database.Cluster{ID: "cluster-1", Name: "test"}

	if app.selectedCluster == nil {
		t.Fatal("expected selected cluster")
	}
	if app.selectedCluster.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", app.selectedCluster.Name)
	}
}

// TestApp_SelectedNode tests selected node management
func TestApp_SelectedNode(t *testing.T) {
	app := &App{}

	// No selection initially
	if app.selectedNode != nil {
		t.Error("expected no selected node")
	}

	// Set selection
	app.selectedNode = &database.Node{ID: "node-1", Name: "worker-1"}

	if app.selectedNode == nil {
		t.Fatal("expected selected node")
	}
	if app.selectedNode.Name != "worker-1" {
		t.Errorf("expected name 'worker-1', got '%s'", app.selectedNode.Name)
	}
}

// TestApp_Clusters tests clusters list management
func TestApp_Clusters(t *testing.T) {
	app := &App{}

	// Initialize clusters list
	app.clusters = []*database.Cluster{
		{ID: "c1", Name: "cluster-1"},
		{ID: "c2", Name: "cluster-2"},
		{ID: "c3", Name: "cluster-3"},
	}

	if len(app.clusters) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(app.clusters))
	}

	// Verify cluster names
	expected := []string{"cluster-1", "cluster-2", "cluster-3"}
	for i, c := range app.clusters {
		if c.Name != expected[i] {
			t.Errorf("expected cluster %d name '%s', got '%s'", i, expected[i], c.Name)
		}
	}
}

// TestClusterStruct tests Cluster struct fields
func TestClusterStruct(t *testing.T) {
	cluster := &database.Cluster{
		ID:     "cluster-1",
		Name:   "production",
		Status: "running",
	}

	if cluster.ID != "cluster-1" {
		t.Errorf("expected cluster-1, got %s", cluster.ID)
	}
	if cluster.Name != "production" {
		t.Errorf("expected production, got %s", cluster.Name)
	}
	if cluster.Status != "running" {
		t.Errorf("expected running, got %s", cluster.Status)
	}
}

// TestNodeStruct tests Node struct fields
func TestNodeStruct(t *testing.T) {
	node := &database.Node{
		ID:        "node-1",
		Name:      "worker-1",
		ClusterID: "cluster-1",
		State:     "running",
		IP:        "192.168.1.100",
	}

	if node.ID != "node-1" {
		t.Errorf("expected node-1, got %s", node.ID)
	}
	if node.Name != "worker-1" {
		t.Errorf("expected worker-1, got %s", node.Name)
	}
	if node.State != "running" {
		t.Errorf("expected running, got %s", node.State)
	}
	if node.IP != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", node.IP)
	}
}
