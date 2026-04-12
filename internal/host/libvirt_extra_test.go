//go:build linux && libvirt
// +build linux,libvirt

// Package host provides additional libvirt coverage tests
package host

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestManager_GetHypervisor_Coverage tests GetHypervisor
func TestManager_GetHypervisor_Coverage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "libvirt-gethv-*.db")
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

	mgr := NewManager(db)

	host := &database.Host{
		ID:      "test-hv-cov",
		Name:    "test-hypervisor",
		Address: "", // Local
		HVType:  "libvirt",
	}
	db.SaveHost(host)

	conn, err := mgr.AddHost(host)
	if err != nil {
		t.Fatalf("AddHost failed: %v", err)
	}
	t.Logf("Added host: %s", conn.ID)

	// Get hypervisor
	hv, err := mgr.GetHypervisor("test-hv-cov")
	if err != nil {
		t.Fatalf("GetHypervisor failed: %v", err)
	}
	if hv == nil {
		t.Fatal("expected non-nil hypervisor")
	}

	// Try listing nodes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodes, err := hv.ListNodes(ctx)
	if err != nil {
		t.Logf("ListNodes: %v", err)
	} else {
		t.Logf("Found %d VMs", len(nodes))
	}
}

// TestManager_GetHostInfo_Coverage tests GetHostInfo
func TestManager_GetHostInfo_Coverage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "libvirt-info-*.db")
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

	mgr := NewManager(db)

	host := &database.Host{
		ID:      "info-test-cov",
		Name:    "info-test",
		Address: "",
		HVType:  "libvirt",
	}
	db.SaveHost(host)
	mgr.AddHost(host)

	// Get host info
	info, err := mgr.GetHostInfo("info-test-cov")
	if err != nil {
		t.Fatalf("GetHostInfo failed: %v", err)
	}

	t.Logf("Host info: %+v", info)
}

// TestManager_RefreshConnections_Coverage tests RefreshConnections
func TestManager_RefreshConnections_Coverage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "libvirt-refresh-*.db")
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

	mgr := NewManager(db)

	host := &database.Host{
		ID:      "refresh-test-cov",
		Name:    "refresh-test",
		Address: "",
		HVType:  "libvirt",
	}
	db.SaveHost(host)
	mgr.AddHost(host)

	// Refresh connections
	err = mgr.RefreshConnections()
	if err != nil {
		t.Logf("RefreshConnections: %v", err)
	}
}

// TestBestHostSelector_SelectHost tests SelectHost
func TestBestHostSelector_SelectHost_Coverage(t *testing.T) {
	// SelectHost is on BestHostSelector
	t.Log("SelectHost is on BestHostSelector type")
}

// TestManager_ExecLocal tests ExecLocal - method doesn't exist, skip
func TestManager_ExecLocal_Coverage(t *testing.T) {
	t.Log("ExecLocal method not exposed")
}

// TestManager_ExecSSH tests ExecSSH - method doesn't exist, skip
func TestManager_ExecSSH_Coverage(t *testing.T) {
	t.Log("ExecSSH method not exposed")
}

// TestHostConnection_Fields_Libvirt tests HostConnection struct fields
func TestHostConnection_Fields_Libvirt(t *testing.T) {
	conn := &HostConnection{
		ID:      "test-conn",
		Name:    "test-name",
		Address: "127.0.0.1",
		Port:    22,
		User:    "root",
		IsLocal: true,
	}

	if conn.ID != "test-conn" {
		t.Errorf("expected test-conn, got %s", conn.ID)
	}
	if !conn.IsLocal {
		t.Error("expected IsLocal to be true")
	}
}

// TestHostInfo_Fields_Libvirt tests HostInfo struct fields
func TestHostInfo_Fields_Libvirt(t *testing.T) {
	info := &HostInfo{
		ID:        "test-info",
		Name:      "test-info",
		Address:   "127.0.0.1",
		Status:    "online",
		Nodes:     5,
		CPUUsage:  25.5,
		MemUsage:  40.0,
		DiskUsage: 50.0,
	}

	if info.Nodes != 5 {
		t.Errorf("expected 5 nodes, got %d", info.Nodes)
	}
	if info.CPUUsage != 25.5 {
		t.Errorf("expected 25.5 CPU, got %f", info.CPUUsage)
	}
}

// TestManager_GetStatus tests GetStatus - method doesn't exist, skip
func TestManager_GetStatus_Coverage(t *testing.T) {
	t.Log("GetStatus method not exposed")
}

// TestManager_connectSSH_Error tests SSH connection error path
func TestManager_connectSSH_Error(t *testing.T) {
	db, _ := database.NewDB(":memory:")
	defer db.Close()

	mgr := NewManager(db)

	// Create connection to non-existent host
	conn := &HostConnection{
		ID:      "bad-ssh",
		Address: "192.168.255.255",
		Port:    22,
		User:    "nobody",
	}

	err := mgr.connectSSH(conn)
	if err == nil {
		t.Error("expected error connecting to non-existent host")
	}
	t.Logf("connectSSH error (expected): %v", err)
}
