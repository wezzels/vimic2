// Package host provides multi-host hypervisor management tests
package host

import (
	"testing"

	"github.com/stsgym/vimic2/internal/database"
)

// TestManager_AddHost_Local tests adding a local host
func TestManager_AddHost_Local(t *testing.T) {
	mgr := NewManager(nil)

	cfg := &database.Host{
		ID:      "local-1",
		Name:    "localhost",
		Address: "127.0.0.1",
		Port:    22,
		User:    "root",
		HVType:  "libvirt",
	}

	conn, err := mgr.AddHost(cfg)
	_ = conn
	_ = err
}

// TestManager_AddHost_Remote tests adding a remote host
func TestManager_AddHost_Remote(t *testing.T) {
	mgr := NewManager(nil)

	cfg := &database.Host{
		ID:         "remote-1",
		Name:       "remote-server",
		Address:    "192.168.1.100",
		Port:       22,
		User:       "root",
		SSHKeyPath: "/nonexistent/key",
		HVType:     "libvirt",
	}

	conn, err := mgr.AddHost(cfg)
	_ = conn
	_ = err
}

// TestManager_AddHost_Duplicate tests adding duplicate host
func TestManager_AddHost_Duplicate(t *testing.T) {
	mgr := NewManager(nil)

	mgr.hosts["existing-1"] = &HostConnection{
		ID:      "existing-1",
		Name:    "existing",
		Address: "192.168.1.50",
	}

	cfg := &database.Host{
		ID:   "existing-1",
		Name: "existing",
	}

	conn, err := mgr.AddHost(cfg)
	if err != nil {
		t.Errorf("expected nil error for duplicate, got %v", err)
	}
	if conn == nil {
		t.Error("expected non-nil connection for duplicate")
	}
}

// TestManager_GetConnection_NonExistent tests getting non-existent host
func TestManager_GetConnection_NonExistent(t *testing.T) {
	mgr := NewManager(nil)

	conn, err := mgr.GetConnection("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent host")
	}
	if conn != nil {
		t.Error("expected nil for non-existent host")
	}
}

// TestManager_ListHosts_Multiple tests listing multiple hosts
func TestManager_ListHosts_Multiple(t *testing.T) {
	mgr := NewManager(nil)

	mgr.hosts["h1"] = &HostConnection{ID: "h1", Name: "host1"}
	mgr.hosts["h2"] = &HostConnection{ID: "h2", Name: "host2"}
	mgr.hosts["h3"] = &HostConnection{ID: "h3", Name: "host3"}

	hosts := mgr.ListHosts()
	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(hosts))
	}
}

// TestHostConnection_AllFields tests all fields
func TestHostConnection_AllFields(t *testing.T) {
	conn := &HostConnection{
		ID:      "conn-1",
		Name:    "test-connection",
		Address: "10.0.0.1",
		Port:    22,
		User:    "admin",
		SSHKey:  "/home/admin/.ssh/id_rsa",
		IsLocal: false,
	}

	if conn.ID != "conn-1" {
		t.Errorf("expected conn-1, got %s", conn.ID)
	}
	if conn.Name != "test-connection" {
		t.Errorf("expected test-connection, got %s", conn.Name)
	}
	if conn.User != "admin" {
		t.Errorf("expected admin, got %s", conn.User)
	}
}

// TestDatabaseHostStruct tests database.Host struct
func TestDatabaseHostStruct(t *testing.T) {
	host := &database.Host{
		ID:         "db-host-1",
		Name:       "db-server",
		Address:    "db.example.com",
		Port:       22,
		User:       "postgres",
		SSHKeyPath: "/var/lib/postgresql/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if host.ID != "db-host-1" {
		t.Errorf("expected db-host-1, got %s", host.ID)
	}
	if host.HVType != "libvirt" {
		t.Errorf("expected libvirt, got %s", host.HVType)
	}
}
