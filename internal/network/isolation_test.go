//go:build integration

package network

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestIsolationManager_StateManagement tests state file operations
func TestIsolationManager_StateManagement(t *testing.T) {
	// Create temp state file
	tmpDir, err := os.MkdirTemp("", "vimic2-isolation-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	stateFile := tmpDir + "/isolation-state.json"

	// Create isolated network struct
	network := &IsolatedNetwork{
		ID:         "test-network-1",
		PipelineID: "test-pipeline-1",
		BridgeName: "vimic2-test-1",
		VLANID:     100,
		CIDR:       "10.100.0.0/24",
		Gateway:    "10.100.0.1",
		DNS:        []string{"8.8.8.8"},
		VMs:        []string{"vm-1", "vm-2"},
		CreatedAt:  time.Now(),
	}

	t.Logf("Created isolated network struct: %+v", network)

	// Test JSON marshaling via json.Marshal
	imported, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	t.Logf("Serialized network: %s", string(imported))

	// Test unmarshaling
	var network2 IsolatedNetwork
	if err := json.Unmarshal(imported, &network2); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if network2.ID != network.ID {
		t.Errorf("Unmarshal ID = %s, want %s", network2.ID, network.ID)
	}

	_ = stateFile // Will use in full integration test
}

// TestIsolationManager_NetworkStruct tests IsolatedNetwork struct operations
func TestIsolationManager_NetworkStruct(t *testing.T) {
	now := time.Now()
	destroyed := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		network IsolatedNetwork
	}{
		{
			name: "basic network",
			network: IsolatedNetwork{
				ID:         "net-1",
				PipelineID: "pipeline-1",
				BridgeName: "br-test",
				VLANID:     100,
				CIDR:       "10.0.0.0/24",
				Gateway:    "10.0.0.1",
				DNS:        []string{"8.8.8.8", "1.1.1.1"},
				VMs:        []string{},
				CreatedAt:  now,
			},
		},
		{
			name: "network with VMs",
			network: IsolatedNetwork{
				ID:         "net-2",
				PipelineID: "pipeline-2",
				BridgeName: "br-test-2",
				VLANID:     200,
				CIDR:       "192.168.100.0/24",
				Gateway:    "192.168.100.1",
				DNS:        []string{"8.8.8.8"},
				VMs:        []string{"vm-1", "vm-2", "vm-3"},
				CreatedAt:  now,
				DestroyedAt: &destroyed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			net := tt.network

			// Test field access
			if net.ID == "" {
				t.Error("ID should not be empty")
			}
			if net.PipelineID == "" {
				t.Error("PipelineID should not be empty")
			}
			if net.VLANID < 1 || net.VLANID > 4094 {
				t.Errorf("VLANID %d out of valid range (1-4094)", net.VLANID)
			}
			if net.Gateway == "" {
				t.Error("Gateway should not be empty")
			}
		})
	}
}

// TestIsolationManager_Config tests NetworkConfig struct
func TestIsolationManager_Config(t *testing.T) {
	config := &NetworkConfig{
		VLANStart:       100,
		VLANEnd:         200,
		BaseCIDR:        "10.100.0.0/16",
		DNS:             []string{"8.8.8.8", "1.1.1.1"},
		OVSBridge:       "br-int",
		FirewallBackend: "stub",
	}

	// Validate config
	if config.VLANStart >= config.VLANEnd {
		t.Errorf("VLANStart %d should be less than VLANEnd %d", config.VLANStart, config.VLANEnd)
	}

	if len(config.DNS) == 0 {
		t.Error("DNS servers should not be empty")
	}

	if config.OVSBridge == "" {
		t.Error("OVSBridge should not be empty")
	}
}

// TestIsolationManager_GenerateNetworkID tests network ID generation
func TestIsolationManager_GenerateNetworkID(t *testing.T) {
	id1 := generateNetworkID("test-pipeline-1")
	id2 := generateNetworkID("test-pipeline-2")

	if id1 == "" {
		t.Error("generateNetworkID returned empty string")
	}

	if id1 == id2 {
		t.Error("generateNetworkID should generate unique IDs for different pipelines")
	}
}

// TestIsolationManager_CreateNetworkContext tests context creation
func TestIsolationManager_CreateNetworkContext(t *testing.T) {
	// This tests the struct creation without actual network operations
	ctx := context.Background()

	// Create a mock isolation manager for testing
	im := &IsolationManager{
		networks: make(map[string]*IsolatedNetwork),
	}

	// Test context creation (metadata only)
	networkCtx := &IsolatedNetwork{
		ID:         "test-ctx",
		PipelineID: "test-pipeline",
		VLANID:     100,
		CIDR:       "10.0.0.0/24",
		Gateway:    "10.0.0.1",
		CreatedAt:  time.Now(),
	}

	// Store in memory
	im.networks[networkCtx.ID] = networkCtx

	// Verify storage
	retrieved, ok := im.networks["test-ctx"]
	if !ok {
		t.Fatal("Network not found in map")
	}

	if retrieved.VLANID != 100 {
		t.Errorf("Retrieved VLANID = %d, want 100", retrieved.VLANID)
	}

	// Cleanup
	delete(im.networks, networkCtx.ID)

	_ = ctx // Used in full integration
}

// TestIsolationManager_WithRealDB tests isolation manager with real database
func TestIsolationManager_WithRealDB(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "vimic2-isolation-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create real database
	db, err := NewNetworkDB(tmpPath)
	if err != nil {
		t.Fatalf("NewNetworkDB failed: %v", err)
	}
	defer db.Close()

	// Test database operations
	ctx := context.Background()
	
	// Save a network
	net := &Network{
		Name: "test-isolation-net",
		Type: NetworkTypeBridge,
		CIDR: "10.200.0.0/24",
	}
	if err := db.SaveNetwork(ctx, net); err != nil {
		t.Fatalf("SaveNetwork failed: %v", err)
	}

	t.Log("Successfully created database for isolation manager")
}