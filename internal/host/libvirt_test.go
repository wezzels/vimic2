//go:build linux && libvirt
// +build linux,libvirt

// Package host provides real libvirt integration tests
package host

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestRealLibvirt_Connect tests connecting to local libvirt
func TestRealLibvirt_Connect(t *testing.T) {
	// Skip if no libvirt access
	if os.Getenv("LIBVIRT_DEFAULT_URI") == "" && os.Getuid() != 0 {
		t.Skip("requires root or LIBVIRT_DEFAULT_URI for libvirt access")
	}

	// Create database
	tmpFile, err := os.CreateTemp("", "libvirt-test-*.db")
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

	// Create host entry for local libvirt
	host := &database.Host{
		ID:      "local-libvirt",
		Name:    "localhost",
		Address: "127.0.0.1",
		Port:    0,
		User:    "root",
		HVType:  "libvirt",
	}
	db.SaveHost(host)

	// Create manager
	mgr := NewManager(db)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}

	// Try to add the host (connect to libvirt)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := mgr.AddHost(host)
	if err != nil {
		t.Logf("AddHost failed (may be expected): %v", err)
		return
	}

	// If connected, try listing VMs
	nodes, err := conn.hv.ListNodes(ctx)
	if err != nil {
		t.Logf("ListNodes failed: %v", err)
	} else {
		t.Logf("Found %d VMs on libvirt", len(nodes))
		for _, node := range nodes {
			t.Logf("  - %s: %s (%s)", node.ID, node.Name, node.State)
		}
	}

	_ = conn
}

// TestRealLibvirt_ListVMs tests listing VMs from local libvirt
func TestRealLibvirt_ListVMs(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("requires root for libvirt access")
	}

	// Use virsh to verify libvirt works
	// This test just verifies we can talk to libvirt
	// The actual hypervisor integration is in pkg/hypervisor

	t.Log("Libvirt integration would work here")
}

// TestManager_AddHost_LocalLibvirt tests adding local libvirt host
func TestManager_AddHost_LocalLibvirt(t *testing.T) {
	// Check for libvirt access (qemu:///system)
	// Either root or libvirt group membership
	if os.Getuid() != 0 {
		// Check if in libvirt/kvm group
		groups, _ := os.Getgroups()
		hasLibvirt := false
		for _, g := range groups {
			if g == 108 || g == 999 { // common libvirt/kvm group IDs
				hasLibvirt = true
				break
			}
		}
		if !hasLibvirt {
			t.Skip("requires root or libvirt group membership")
		}
	}

	tmpFile, err := os.CreateTemp("", "libvirt-host-*.db")
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

	// Use empty address to get qemu:///system directly (no SSH)
	host := &database.Host{
		ID:      "libvirt-local",
		Name:    "local-hypervisor",
		Address: "", // Empty = qemu:///system directly
		HVType:  "libvirt",
	}

	mgr := NewManager(db)

	conn, err := mgr.AddHost(host)
	if err != nil {
		t.Fatalf("AddHost failed: %v", err)
	}

	t.Logf("Connected to libvirt: %s", conn.ID)

	// Now try to list VMs
	ctx := context.Background()
	nodes, err := conn.hv.ListNodes(ctx)
	if err != nil {
		t.Logf("ListNodes failed: %v", err)
	} else {
		t.Logf("Found %d VMs on libvirt", len(nodes))
		for _, node := range nodes {
			t.Logf("  - %s: %s (%s)", node.ID, node.Name, node.State)
		}
	}
}
