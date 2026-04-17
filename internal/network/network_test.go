//go:build integration

package network

import (
	"context"
	"os"
	"testing"
)

// TestNetworkManager_Database tests network database operations
func TestNetworkManager_Database(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vimic2-net-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create database directly
	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("NewNetworkDB failed: %v", err)
	}
	defer db.Close()

	// Test network database operations
	ctx := context.Background()

	// Create network
	network := &Network{
		ID:      "test-net-1",
		Name:    "test-network",
		Type:    NetworkTypeBridge,
		CIDR:    "10.100.0.0/24",
		Gateway: "10.100.0.1",
	}

	err = db.SaveNetwork(ctx, network)
	if err != nil {
		t.Errorf("SaveNetwork failed: %v", err)
	}

	if network.ID == "" {
		t.Error("Network ID should be set after creation")
	}

	// Get network
	retrieved, err := db.GetNetwork(ctx, network.ID)
	if err != nil {
		t.Errorf("GetNetwork failed: %v", err)
	}

	if retrieved.Name != network.Name {
		t.Errorf("Retrieved network name = %s, want %s", retrieved.Name, network.Name)
	}
}

// TestNetworkManager_CreateRouter tests router creation
func TestNetworkManager_CreateRouter(t *testing.T) {
	t.Skip("Requires OVS/network namespaces")
}

// TestNetworkManager_CreateFirewall tests firewall creation
func TestNetworkManager_CreateFirewall(t *testing.T) {
	t.Skip("Requires iptables")
}

// TestNetworkManager_CreateTunnel tests tunnel creation
func TestNetworkManager_CreateTunnel(t *testing.T) {
	t.Skip("Requires OVS")
}

// TestNetworkManager_GenerateID tests ID generation
func TestNetworkManager_GenerateID(t *testing.T) {
	id1 := generateID("net")
	id2 := generateID("net")

	if id1 == "" {
		t.Error("generateID returned empty string")
	}

	if id1 == id2 {
		t.Error("generateID should generate unique IDs")
	}

	// Check format (should be prefix-random)
	if len(id1) < 10 {
		t.Errorf("ID %s seems too short", id1)
	}
}

// TestNetworkStructs tests network struct field validation
func TestNetworkStructs(t *testing.T) {
	t.Run("Network", func(t *testing.T) {
		net := Network{
			ID:      "net-1",
			Name:    "test",
			Type:    NetworkTypeBridge,
			CIDR:    "10.0.0.0/24",
			Gateway: "10.0.0.1",
		}

		if net.Type != NetworkTypeBridge {
			t.Errorf("Network type = %s, want %s", net.Type, NetworkTypeBridge)
		}
	})

	t.Run("Router", func(t *testing.T) {
		router := Router{
			ID:      "router-1",
			Name:    "test-router",
			Enabled: true,
		}

		if router.Name == "" {
			t.Error("Router name should not be empty")
		}
	})

	t.Run("Firewall", func(t *testing.T) {
		fw := Firewall{
			ID:            "fw-1",
			Name:          "test-fw",
			DefaultPolicy: "drop",
			Rules:         []FirewallRule{},
			Enabled:       true,
		}

		if fw.DefaultPolicy != "drop" && fw.DefaultPolicy != "accept" {
			t.Errorf("Invalid default policy: %s", fw.DefaultPolicy)
		}
	})

	t.Run("Tunnel", func(t *testing.T) {
		tunnel := Tunnel{
			ID:       "tunnel-1",
			Name:     "test-tunnel",
			Protocol: TunnelVXLAN,
			VNI:      100,
			DestPort: 4789,
		}

		if tunnel.VNI < 1 || tunnel.VNI > 16777215 {
			t.Errorf("VNI %d out of valid range (1-16777215)", tunnel.VNI)
		}
	})

	t.Run("FirewallRule", func(t *testing.T) {
		rule := FirewallRule{
			ID:        "rule-1",
			Name:      "allow-ssh",
			Direction: "ingress",
			Protocol:  "tcp",
			DestPort:  22,
			Action:    "accept",
			Priority:  100,
			Enabled:   true,
		}

		if rule.Protocol != "tcp" && rule.Protocol != "udp" && rule.Protocol != "icmp" {
			t.Errorf("Invalid protocol: %s", rule.Protocol)
		}

		if rule.Action != "accept" && rule.Action != "drop" && rule.Action != "reject" {
			t.Errorf("Invalid action: %s", rule.Action)
		}
	})
}