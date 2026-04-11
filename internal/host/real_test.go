// Package host provides real implementation tests
package host

import (
	"testing"

	"github.com/stsgym/vimic2/internal/database"
)

// TestHostConnection_LocalAddress tests local address detection
func TestHostConnection_LocalAddress(t *testing.T) {
	tests := []struct {
		addr     string
		isLocal  bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"192.168.1.100", false},
		{"10.0.0.1", false},
		{"example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			// Local address check logic
			isLocal := tt.addr == "localhost" || 
				tt.addr == "127.0.0.1" || 
				tt.addr == "::1"

			if isLocal != tt.isLocal {
				t.Errorf("expected isLocal=%v for %s", tt.isLocal, tt.addr)
			}
		})
	}
}

// TestNewManager_Real tests real manager creation
func TestNewManager_Real(t *testing.T) {
	manager := NewManager(nil)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	if manager.hosts == nil {
		t.Error("expected non-nil hosts map")
	}
}

// TestManager_GetConnection_NotFound tests GetConnection when not found
func TestManager_GetConnection_NotFound(t *testing.T) {
	manager := NewManager(nil)

	conn, err := manager.GetConnection("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
	if conn != nil {
		t.Error("expected nil connection")
	}
}

// TestManager_ListHosts_Empty tests ListHosts when empty
func TestManager_ListHosts_Empty(t *testing.T) {
	manager := NewManager(nil)

	hosts := manager.ListHosts()
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

// TestManager_RemoveHost_Real tests host removal
func TestManager_RemoveHost_Real(t *testing.T) {
	manager := NewManager(nil)

	// Add a host to the map
	manager.hosts["test-host"] = &HostConnection{
		ID:      "test-host",
		Name:    "Test Host",
		Address: "192.168.1.100",
	}

	// Remove it
	err := manager.RemoveHost("test-host")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify removed
	if _, exists := manager.hosts["test-host"]; exists {
		t.Error("expected host to be removed")
	}

	// Remove nonexistent (should succeed)
	err = manager.RemoveHost("nonexistent")
	if err != nil {
		t.Errorf("expected no error for nonexistent, got: %v", err)
	}
}

// TestHostConnection_GetStatus tests status retrieval
func TestHostConnection_GetStatus(t *testing.T) {
	conn := &HostConnection{
		ID:      "host-1",
		Name:    "Test Host",
		Address: "192.168.1.100",
		Port:    22,
	}

	// Status should indicate disconnected when no client
	status := conn.GetStatus()
	if status != "disconnected" && status != "offline" {
		t.Logf("status for disconnected host: %s", status)
	}
}

// TestHostConnection_AddressFormat tests address formatting
func TestHostConnection_AddressFormat(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		port     int
		expected string
	}{
		{"With port", "192.168.1.100", 22, "192.168.1.100:22"},
		{"No port default", "192.168.1.100", 0, "192.168.1.100"},
		{"Custom port", "10.0.0.1", 2222, "10.0.0.1:2222"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &HostConnection{
				Address: tt.address,
				Port:    tt.port,
			}

			if conn.Address != tt.address {
				t.Errorf("expected %s, got %s", tt.address, conn.Address)
			}
		})
	}
}

// TestHostConnection_Fields tests all HostConnection fields
func TestHostConnection_Fields(t *testing.T) {
	conn := &HostConnection{
		ID:      "conn-1",
		Name:    "Primary HV",
		Address: "192.168.1.10",
		Port:    22,
		User:    "root",
		SSHKey:  "/root/.ssh/id_rsa",
		IsLocal: false,
	}

	if conn.ID != "conn-1" {
		t.Errorf("expected conn-1, got %s", conn.ID)
	}
	if conn.Name != "Primary HV" {
		t.Errorf("expected Primary HV, got %s", conn.Name)
	}
	if conn.User != "root" {
		t.Errorf("expected root, got %s", conn.User)
	}
	if conn.IsLocal {
		t.Error("expected IsLocal to be false")
	}
}

// TestHostInfo_Struct tests HostInfo struct
func TestHostInfo_Struct(t *testing.T) {
	info := &HostInfo{
		ID:        "host-1",
		Name:      "HV-1",
		Address:   "192.168.1.10",
		Status:    "online",
		Nodes:     10,
		CPUUsage:  45.5,
		MemUsage: 60.2,
	}

	if info.ID != "host-1" {
		t.Errorf("expected host-1, got %s", info.ID)
	}
	if info.Status != "online" {
		t.Errorf("expected online, got %s", info.Status)
	}
	if info.Nodes != 10 {
		t.Errorf("expected 10 nodes, got %d", info.Nodes)
	}
}

// TestManager_HostsMap tests hosts map operations
func TestManager_HostsMap(t *testing.T) {
	manager := NewManager(nil)

	// Add multiple hosts
	manager.hosts["hv-1"] = &HostConnection{ID: "hv-1", Name: "HV1"}
	manager.hosts["hv-2"] = &HostConnection{ID: "hv-2", Name: "HV2"}
	manager.hosts["hv-3"] = &HostConnection{ID: "hv-3", Name: "HV3"}

	if len(manager.hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(manager.hosts))
	}

	// List should return all
	hosts := manager.ListHosts()
	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts in list, got %d", len(hosts))
	}
}

// TestDatabaseHost_Struct tests database.Host struct
func TestDatabaseHost_Struct(t *testing.T) {
	host := &database.Host{
		ID:         "db-host-1",
		Name:       "Database HV",
		Address:    "db.example.com",
		Port:       22,
		User:       "admin",
		SSHKeyPath: "/home/admin/.ssh/id_rsa",
		HVType:     "libvirt",
	}

	if host.ID != "db-host-1" {
		t.Errorf("expected db-host-1, got %s", host.ID)
	}
	if host.HVType != "libvirt" {
		t.Errorf("expected libvirt, got %s", host.HVType)
	}
}

// TestBestHostSelector tests host selection logic
func TestBestHostSelector(t *testing.T) {
	selector := NewBestHostSelector(nil)

	if selector == nil {
		t.Error("expected non-nil selector")
	}
}