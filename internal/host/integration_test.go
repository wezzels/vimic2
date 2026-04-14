//go:build integration

package host

import (
	"context"
	"os"
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

func setupHostTest(t *testing.T) (*Manager, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-host-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := tmpDir + "/test.db"
	db, err := database.NewDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	mgr := NewManager(db)
	return mgr, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== ExecLocal Tests ====================

func TestExecLocal_Echo(t *testing.T) {
	out, err := ExecLocal("echo", "hello")
	if err != nil {
		t.Fatalf("ExecLocal echo failed: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("ExecLocal echo = %q, want %q", string(out), "hello\n")
	}
}

func TestExecLocal_Ls(t *testing.T) {
	out, err := ExecLocal("ls", "/")
	if err != nil {
		t.Fatalf("ExecLocal ls failed: %v", err)
	}
	if len(out) == 0 {
		t.Error("ExecLocal ls should return output")
	}
}

func TestExecLocal_InvalidCommand(t *testing.T) {
	_, err := ExecLocal("nonexistent_command_that_does_not_exist")
	if err == nil {
		t.Error("ExecLocal should fail for nonexistent command")
	}
}

func TestExecLocal_True(t *testing.T) {
	_, err := ExecLocal("true")
	if err != nil {
		t.Errorf("ExecLocal true failed: %v", err)
	}
}

func TestExecLocal_False(t *testing.T) {
	_, err := ExecLocal("false")
	if err == nil {
		t.Error("ExecLocal false should return error")
	}
}

// ==================== HostConnection Tests ====================

func TestHostConnection_GetStatus_Disconnected(t *testing.T) {
	conn := &HostConnection{
		ID:      "host-1",
		Name:    "test-host",
		Address: "10.0.0.1",
	}

	status := conn.GetStatus()
	if status != "disconnected" {
		t.Errorf("GetStatus = %s, want disconnected", status)
	}
}

func TestHostConnection_GetStatus_Active(t *testing.T) {
	conn := &HostConnection{
		ID:      "host-1",
		Name:    "test-host",
		Address: "10.0.0.1",
		hv:      hypervisor.NewStubHypervisor(),
	}

	status := conn.GetStatus()
	if status != "active" {
		t.Errorf("GetStatus = %s, want active", status)
	}
}

// ==================== HostInfo Tests ====================

func TestHostInfo_Fields(t *testing.T) {
	info := &HostInfo{
		ID:        "host-1",
		Name:      "test-host",
		Address:   "10.0.0.1",
		Status:    "connected",
		Nodes:     5,
		CPUUsage:  45.5,
		MemUsage:  60.2,
		DiskUsage: 30.1,
	}

	if info.ID != "host-1" {
		t.Errorf("ID = %s, want host-1", info.ID)
	}
	if info.Name != "test-host" {
		t.Errorf("Name = %s, want test-host", info.Name)
	}
	if info.Address != "10.0.0.1" {
		t.Errorf("Address = %s, want 10.0.0.1", info.Address)
	}
	if info.Status != "connected" {
		t.Errorf("Status = %s, want connected", info.Status)
	}
	if info.Nodes != 5 {
		t.Errorf("Nodes = %d, want 5", info.Nodes)
	}
}

// ==================== Manager Tests ====================

func TestManager_GetConnection_NotFound_Int(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	_, err := mgr.GetConnection("nonexistent")
	if err == nil {
		t.Error("GetConnection should return error for nonexistent host")
	}
}

func TestManager_ListHosts_Empty_Int(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	hosts := mgr.ListHosts()
	if len(hosts) != 0 {
		t.Errorf("ListHosts should return empty for new manager, got %d", len(hosts))
	}
}

func TestManager_GetHypervisor_NotFound_Int(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	_, err := mgr.GetHypervisor("nonexistent")
	if err == nil {
		t.Error("GetHypervisor should return error for nonexistent host")
	}
}

// ==================== BestHostSelector Tests ====================

func TestBestHostSelector_New(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	selector := NewBestHostSelector(mgr)
	if selector == nil {
		t.Error("NewBestHostSelector returned nil")
	}
}

func TestBestHostSelector_NoHosts(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	selector := NewBestHostSelector(mgr)
	_, err := selector.SelectHost()
	if err == nil {
		t.Error("SelectHost should return error with no hosts")
	}
}

func TestBestHostSelector_SingleHost_Int(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	conn := &HostConnection{
		ID:      "host-1",
		Name:    "test-host",
		Address: "10.0.0.1",
		hv:      hypervisor.NewStubHypervisor(),
	}
	mgr.hosts["host-1"] = conn

	selector := NewBestHostSelector(mgr)
	hostID, err := selector.SelectHost()
	if err != nil {
		t.Fatalf("SelectHost failed: %v", err)
	}
	if hostID != "host-1" {
		t.Errorf("SelectHost = %s, want host-1", hostID)
	}
}

func TestBestHostSelector_MultipleHosts_Int(t *testing.T) {
	mgr, cleanup := setupHostTest(t)
	defer cleanup()

	for i, id := range []string{"host-1", "host-2", "host-3"} {
		conn := &HostConnection{
			ID:      id,
			Name:    id,
			Address: "10.0.0." + string(rune('1'+i)),
			hv:      hypervisor.NewStubHypervisor(),
		}
		mgr.hosts[conn.ID] = conn
	}

	selector := NewBestHostSelector(mgr)
	hostID, err := selector.SelectHost()
	if err != nil {
		t.Fatalf("SelectHost failed: %v", err)
	}
	if hostID == "" {
		t.Error("SelectHost returned empty host ID")
	}
}

// ==================== IsLocalAddress Tests ====================

func TestIsLocalAddress(t *testing.T) {
	db, err := database.NewDB("")
	if err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(db)

	tests := []struct {
		addr string
		want bool
	}{
		{"127.0.0.1", true},
		{"localhost", true},
		{"::1", true},
		{"10.0.0.1", false},
		{"192.168.1.1", false},
		{"example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := mgr.isLocalAddress(tt.addr)
			if got != tt.want {
				t.Errorf("isLocalAddress(%s) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}

// ==================== Context Test ====================

func TestHostConnection_Context(t *testing.T) {
	conn := &HostConnection{
		ID:      "ctx-test",
		Name:    "context-host",
		Address: "10.0.0.1",
	}

	ctx := context.Background()
	_ = ctx // Just verifying we can use context

	if conn.ID != "ctx-test" {
		t.Errorf("ID = %s, want ctx-test", conn.ID)
	}
}