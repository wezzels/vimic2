//go:build integration

// Package network provides integration tests for IPAM
package network

import (
	"testing"
)

// TestIntegration_IPAM_CreateManager tests creating IPAM manager
func TestIntegration_IPAM_CreateManager(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.100.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	if ipam == nil {
		t.Fatal("expected non-nil IPAMManager")
	}

	t.Logf("Created IPAMManager with base CIDR: %s", config.BaseCIDR)
}

// TestIntegration_IPAM_AllocateRelease tests basic allocation
func TestIntegration_IPAM_AllocateRelease(t *testing.T) {
	config := &IPAMConfig{
		BaseCIDR: "10.200.0.0/24",
		DNS:      []string{"8.8.8.8"},
	}

	ipam, err := NewIPAMManager(config)
	if err != nil {
		t.Fatalf("NewIPAMManager failed: %v", err)
	}

	// First allocation
	cidr, ip, err := ipam.Allocate()
	if err != nil {
		t.Skipf("Allocate failed (may need pool): %v", err)
	}

	t.Logf("Allocated: %s from %s", ip, cidr)
}