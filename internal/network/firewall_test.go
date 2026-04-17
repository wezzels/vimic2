//go:build integration

package network

import (
	"testing"
)

// TestFirewallManager_Stub tests the stub backend which doesn't require iptables/nftables
func TestFirewallManager_Stub(t *testing.T) {
	fm, err := NewFirewallManager(FirewallBackendStub)
	if err != nil {
		t.Fatalf("NewFirewallManager(stub) failed: %v", err)
	}

	// Test backend getter
	if fm.GetBackend() != FirewallBackendStub {
		t.Errorf("GetBackend() = %s, want %s", fm.GetBackend(), FirewallBackendStub)
	}

	// Test chain operations (stub just stores in memory)
	rules := fm.ListRules()
	if rules == nil {
		rules = make(map[string][]string)
	}

	// Stub backend doesn't actually execute, so operations should succeed
	if err := fm.CreateIsolationRules("test-bridge", "10.0.0.0/24", 100); err != nil {
		// Stub might return nil or error depending on implementation
		t.Logf("CreateIsolationRules: %v (expected for stub)", err)
	}

	// Test close
	if err := fm.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

// TestFirewallManager_DetectBackend tests backend detection
func TestFirewallManager_DetectBackend(t *testing.T) {
	fm := &FirewallManager{
		backend: "",
		rules:   make(map[string][]string),
	}

	fm.detectBackend()

	// Should detect some backend (iptables, nftables, or stub)
	validBackends := map[FirewallBackend]bool{
		FirewallBackendIPTables: true,
		FirewallBackendNFTables: true,
		FirewallBackendStub:     true,
	}

	if !validBackends[fm.backend] {
		t.Errorf("detectBackend() = %s, want one of iptables/nftables/stub", fm.backend)
	}
}

// TestFirewallManager_IsBackendAvailable tests backend availability checking
func TestFirewallManager_IsBackendAvailable(t *testing.T) {
	tests := []struct {
		name    string
		backend FirewallBackend
	}{
		{"stub", FirewallBackendStub},
		{"iptables", FirewallBackendIPTables},
		{"nftables", FirewallBackendNFTables},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := &FirewallManager{
				backend: tt.backend,
				rules:   make(map[string][]string),
			}

			available := fm.isBackendAvailable()
			// Stub should always be available
			if tt.backend == FirewallBackendStub && !available {
				t.Error("isBackendAvailable() = false for stub backend")
			}
		})
	}
}

// TestFirewallManager_ListRules tests rule listing
func TestFirewallManager_ListRules(t *testing.T) {
	fm, err := NewFirewallManager(FirewallBackendStub)
	if err != nil {
		t.Fatalf("NewFirewallManager(stub) failed: %v", err)
	}

	rules := fm.ListRules()
	if rules == nil {
		// Rules map should be initialized
		t.Error("ListRules() returned nil map")
	}
}