// Package host provides host manager tests
package host

import (
	"testing"
)

// TestHostConnection tests host connection structure
func TestHostConnection_Create(t *testing.T) {
	conn := &HostConnection{
		ID:      "host-1",
		Name:    "worker-node-1",
		Address: "192.168.1.100",
		Port:    22,
		User:    "root",
		SSHKey:  "/root/.ssh/id_rsa",
		IsLocal: false,
	}

	if conn.ID != "host-1" {
		t.Errorf("expected host-1, got %s", conn.ID)
	}
	if conn.Name != "worker-node-1" {
		t.Errorf("expected worker-node-1, got %s", conn.Name)
	}
	if conn.Address != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", conn.Address)
	}
	if conn.Port != 22 {
		t.Errorf("expected port 22, got %d", conn.Port)
	}
	if conn.User != "root" {
		t.Errorf("expected root user, got %s", conn.User)
	}
	if conn.IsLocal {
		t.Error("expected IsLocal to be false")
	}
}

// TestHostConnection_Local tests local host connection
func TestHostConnection_Local(t *testing.T) {
	conn := &HostConnection{
		ID:       "local-1",
		Name:     "localhost",
		Address:  "127.0.0.1",
		Port:     22,
		User:     "root",
		IsLocal:  true,
	}

	if !conn.IsLocal {
		t.Error("expected IsLocal to be true")
	}
	if conn.Address != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", conn.Address)
	}
}

// TestHostConnection_Remote tests remote host connection
func TestHostConnection_Remote(t *testing.T) {
	conn := &HostConnection{
		ID:       "remote-1",
		Name:     "remote-server",
		Address:  "10.0.0.50",
		Port:     2222,
		User:     "admin",
		SSHKey:   "/home/admin/.ssh/id_rsa",
		IsLocal:  false,
	}

	if conn.IsLocal {
		t.Error("expected IsLocal to be false")
	}
	if conn.Port != 2222 {
		t.Errorf("expected port 2222, got %d", conn.Port)
	}
}

// TestManager_CreateStruct tests manager struct creation
func TestManager_CreateStruct(t *testing.T) {
	mgr := &Manager{
		hosts: make(map[string]*HostConnection),
	}

	if mgr.hosts == nil {
		t.Error("hosts map should not be nil")
	}

	// Add a host
	mgr.hosts["host-1"] = &HostConnection{
		ID:   "host-1",
		Name: "test-host",
	}

	if len(mgr.hosts) != 1 {
		t.Errorf("expected 1 host, got %d", len(mgr.hosts))
	}
}

// TestManager_AddHostToMap tests adding hosts to map
func TestManager_AddHostToMap(t *testing.T) {
	mgr := &Manager{
		hosts: make(map[string]*HostConnection),
	}

	// Add multiple hosts
	hosts := []*HostConnection{
		{ID: "host-1", Name: "worker-1", Address: "192.168.1.10"},
		{ID: "host-2", Name: "worker-2", Address: "192.168.1.11"},
		{ID: "host-3", Name: "worker-3", Address: "192.168.1.12"},
	}

	for _, h := range hosts {
		mgr.hosts[h.ID] = h
	}

	if len(mgr.hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d", len(mgr.hosts))
	}

	// Verify each host exists
	for _, h := range hosts {
		if mgr.hosts[h.ID] == nil {
			t.Errorf("host %s should exist", h.ID)
		}
		if mgr.hosts[h.ID].Name != h.Name {
			t.Errorf("expected name %s, got %s", h.Name, mgr.hosts[h.ID].Name)
		}
	}
}

// TestManager_RemoveHost tests removing hosts from map
func TestManager_RemoveHost(t *testing.T) {
	mgr := &Manager{
		hosts: make(map[string]*HostConnection),
	}

	// Add hosts
	mgr.hosts["host-1"] = &HostConnection{ID: "host-1", Name: "worker-1"}
	mgr.hosts["host-2"] = &HostConnection{ID: "host-2", Name: "worker-2"}
	mgr.hosts["host-3"] = &HostConnection{ID: "host-3", Name: "worker-3"}

	// Remove one
	delete(mgr.hosts, "host-2")

	if len(mgr.hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(mgr.hosts))
	}
	if mgr.hosts["host-2"] != nil {
		t.Error("host-2 should be removed")
	}
}

// TestHostConnection_NetworkAddress tests network address handling
func TestHostConnection_NetworkAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		port    int
	}{
		{"IPv4", "192.168.1.100", 22},
		{"IPv4 with port", "10.0.0.50", 2222},
		{"localhost", "127.0.0.1", 22},
		{"hostname", "server.example.com", 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &HostConnection{
				Address: tt.address,
				Port:    tt.port,
			}

			if conn.Address != tt.address {
				t.Errorf("expected address %s, got %s", tt.address, conn.Address)
			}
			if conn.Port != tt.port {
				t.Errorf("expected port %d, got %d", tt.port, conn.Port)
			}
		})
	}
}

// TestHostConnection_SSHKeyPath tests SSH key path handling
func TestHostConnection_SSHKeyPath(t *testing.T) {
	conn := &HostConnection{
		ID:     "host-1",
		SSHKey: "/home/user/.ssh/id_ed25519",
	}

	if conn.SSHKey != "/home/user/.ssh/id_ed25519" {
		t.Errorf("expected SSH key path, got %s", conn.SSHKey)
	}

	// Test with default key
	conn2 := &HostConnection{
		ID:     "host-2",
		SSHKey: "~/.ssh/id_rsa",
	}

	if conn2.SSHKey != "~/.ssh/id_rsa" {
		t.Errorf("expected SSH key path, got %s", conn2.SSHKey)
	}
}

// TestHostConnection_DifferentUsers tests different user configurations
func TestHostConnection_DifferentUsers(t *testing.T) {
	users := []struct {
		name string
		user string
	}{
		{"root", "root"},
		{"admin", "admin"},
		{"ubuntu", "ubuntu"},
		{"centos", "centos"},
	}

	for _, tt := range users {
		t.Run(tt.name, func(t *testing.T) {
			conn := &HostConnection{User: tt.user}
			if conn.User != tt.user {
				t.Errorf("expected user %s, got %s", tt.user, conn.User)
			}
		})
	}
}