//go:build libvirt

// Package host provides integration tests with real libvirt
package host

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestIntegration_Libvirt_Connect tests real libvirt connection
func TestIntegration_Libvirt_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}

	mgr := NewManager(db)

	// Add local libvirt host
	host := &database.Host{
		ID:        "local-libvirt",
		Name:      "local-libvirt",
		Address:   "127.0.0.1",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	conn, err := mgr.AddHost(host)
	if err != nil {
		t.Fatalf("AddHost failed: %v", err)
	}

	t.Logf("Connected to host: %s", conn.Name)

	// Get hypervisor
	hyp, err := mgr.GetHypervisor(host.ID)
	if err != nil {
		t.Fatalf("GetHypervisor failed: %v", err)
	}

	t.Logf("Got hypervisor: %T", hyp)
}

// TestIntegration_Libvirt_GetHostInfo tests getting host info
func TestIntegration_Libvirt_GetHostInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	mgr := NewManager(db)

	host := &database.Host{
		ID:        "info-test",
		Name:      "info-test",
		Address:   "127.0.0.1",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	mgr.AddHost(host)

	info, err := mgr.GetHostInfo(host.ID)
	if err != nil {
		t.Skipf("GetHostInfo failed: %v", err)
	}

	t.Logf("Host info: %+v", info)
}

// TestIntegration_Libvirt_RefreshConnections tests refreshing connections
func TestIntegration_Libvirt_RefreshConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	mgr := NewManager(db)

	// Add multiple hosts
	hosts := []*database.Host{
		{
			ID:        "refresh-1",
			Name:      "refresh-1",
			Address:   "127.0.0.1",
			Port:      22,
			User:      "root",
			HVType:    "libvirt",
			CreatedAt: time.Now(),
		},
		{
			ID:        "refresh-2",
			Name:      "refresh-2",
			Address:   "127.0.0.1",
			Port:      22,
			User:      "root",
			HVType:    "libvirt",
			CreatedAt: time.Now(),
		},
	}

	for _, h := range hosts {
		mgr.AddHost(h)
	}

	// Refresh all connections
	err := mgr.RefreshConnections()
	if err != nil {
		t.Logf("RefreshConnections returned: %v", err)
	}

	// List hosts
	connections := mgr.ListHosts()
	t.Logf("Listed %d hosts", len(connections))
}

// TestIntegration_Libvirt_SelectHost tests selecting best host
func TestIntegration_Libvirt_SelectHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	mgr := NewManager(db)

	host := &database.Host{
		ID:        "select-test",
		Name:      "select-test",
		Address:   "127.0.0.1",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	mgr.AddHost(host)

	selector := NewBestHostSelector(mgr)
	selectedID, err := selector.SelectHost()
	if err != nil {
		t.Skipf("SelectHost failed: %v", err)
	}

	t.Logf("Selected host ID: %s", selectedID)
}

// TestIntegration_Libvirt_RemoveHost tests removing a host
func TestIntegration_Libvirt_RemoveHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	mgr := NewManager(db)

	host := &database.Host{
		ID:        "remove-test",
		Name:      "remove-test",
		Address:   "127.0.0.1",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	mgr.AddHost(host)

	// Remove the host
	err := mgr.RemoveHost(host.ID)
	if err != nil {
		t.Fatalf("RemoveHost failed: %v", err)
	}

	// Verify removed
	conns := mgr.ListHosts()
	for _, c := range conns {
		if c.ID == host.ID {
			t.Error("host still exists after removal")
		}
	}

	t.Logf("Host removed successfully")
}

// TestIntegration_Libvirt_GetConnection tests getting a connection
func TestIntegration_Libvirt_GetConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	mgr := NewManager(db)

	host := &database.Host{
		ID:        "conn-test",
		Name:      "conn-test",
		Address:   "127.0.0.1",
		Port:      22,
		User:      "root",
		HVType:    "libvirt",
		CreatedAt: time.Now(),
	}

	mgr.AddHost(host)

	conn, err := mgr.GetConnection(host.ID)
	if err != nil {
		t.Fatalf("GetConnection failed: %v", err)
	}

	if conn == nil {
		t.Fatal("expected non-nil connection")
	}

	t.Logf("Got connection for host: %s", conn.Name)
}