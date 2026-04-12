// Package database provides DB tests with real implementations
package database

import (
	"os"
	"testing"
)

// TestNewDB_Real tests real database creation
func TestNewDB_Real(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("expected non-nil DB")
	}
}

// TestDB_SaveAndGetHost tests host save/get operations
func TestDB_SaveAndGetHost(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Save a host
	host := &Host{
		ID:         "host-1",
		Name:       "test-host",
		Address:    "192.168.1.10",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/root/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	err = db.SaveHost(host)
	if err != nil {
		t.Fatalf("failed to save host: %v", err)
	}

	// Get it back
	retrieved, err := db.GetHost("host-1")
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	if retrieved.Name != "test-host" {
		t.Errorf("expected test-host, got %s", retrieved.Name)
	}
	if retrieved.Address != "192.168.1.10" {
		t.Errorf("expected 192.168.1.10, got %s", retrieved.Address)
	}
}

// TestDB_ListHosts tests listing hosts
func TestDB_ListHosts(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Save multiple hosts
	for i := 1; i <= 3; i++ {
		host := &Host{
			ID:      string(rune('a' + i)),
			Name:    "host-" + string(rune('0'+i)),
			Address: "192.168.1." + string(rune('0'+i)),
			HVType:  "libvirt",
		}
		db.SaveHost(host)
	}

	hosts, err := db.ListHosts()
	if err != nil {
		t.Fatalf("failed to list hosts: %v", err)
	}

	if len(hosts) < 3 {
		t.Errorf("expected at least 3 hosts, got %d", len(hosts))
	}
}

// TestDB_DeleteHost tests host deletion
func TestDB_DeleteHost(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	// Save and delete
	host := &Host{ID: "host-to-delete", Name: "delete-me"}
	db.SaveHost(host)

	err = db.DeleteHost("host-to-delete")
	if err != nil {
		t.Fatalf("failed to delete host: %v", err)
	}

	// Verify deletion - GetHost may return nil or error depending on implementation
	retrieved, _ := db.GetHost("host-to-delete")
	if retrieved != nil {
		t.Error("expected host to be deleted")
	}
}

// TestDB_SaveAndGetCluster tests cluster operations
func TestDB_SaveAndGetCluster(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	cluster := &Cluster{
		ID:     "cluster-1",
		Name:   "test-cluster",
		Status: "running",
	}

	err = db.SaveCluster(cluster)
	if err != nil {
		t.Fatalf("failed to save cluster: %v", err)
	}

	retrieved, err := db.GetCluster("cluster-1")
	if err != nil {
		t.Fatalf("failed to get cluster: %v", err)
	}

	if retrieved.Name != "test-cluster" {
		t.Errorf("expected test-cluster, got %s", retrieved.Name)
	}
}

// TestDB_UpdateClusterStatus tests status updates
func TestDB_UpdateClusterStatus(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	cluster := &Cluster{ID: "cluster-1", Name: "test", Status: "pending"}
	db.SaveCluster(cluster)

	err = db.UpdateClusterStatus("cluster-1", "running")
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	retrieved, _ := db.GetCluster("cluster-1")
	if retrieved.Status != "running" {
		t.Errorf("expected running, got %s", retrieved.Status)
	}
}

// TestDB_SaveAndGetNode tests node operations
func TestDB_SaveAndGetNode(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	node := &Node{
		ID:        "node-1",
		Name:      "worker-1",
		ClusterID: "cluster-1",
		HostID:    "host-1",
		State:     "running",
		Role:      "worker",
	}

	err = db.SaveNode(node)
	if err != nil {
		t.Fatalf("failed to save node: %v", err)
	}

	retrieved, err := db.GetNode("node-1")
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	if retrieved.Name != "worker-1" {
		t.Errorf("expected worker-1, got %s", retrieved.Name)
	}
}

// TestDB_SaveAndGetMetric tests metric operations
func TestDB_SaveAndGetMetric(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	db, err := NewDB(tmpPath)
	if err != nil {
		t.Fatalf("failed to create DB: %v", err)
	}
	defer db.Close()

	metric := &Metric{
		NodeID: "node-1",
		CPU:    45.5,
		Memory: 60.2,
	}

	err = db.SaveMetric(metric)
	if err != nil {
		t.Fatalf("failed to save metric: %v", err)
	}

	retrieved, err := db.GetLatestMetric("node-1")
	if err != nil {
		t.Fatalf("failed to get metric: %v", err)
	}

	if retrieved.CPU != 45.5 {
		t.Errorf("expected CPU 45.5, got %f", retrieved.CPU)
	}
}
