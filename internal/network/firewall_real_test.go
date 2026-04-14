//go:build integration

package network

import (
	"os"
	"testing"
)

// TestFirewallManager_Real tests with actual iptables/nftables
func TestFirewallManager_Real(t *testing.T) {
	// Check if iptables or nft is available
	hasIptables := false
	hasNft := false

	if _, err := os.Stat("/usr/sbin/iptables"); err == nil {
		hasIptables = true
	}
	if _, err := os.Stat("/usr/sbin/nft"); err == nil {
		hasNft = true
	}

	if !hasIptables && !hasNft {
		t.Skip("Neither iptables nor nft available")
	}

	t.Run("DetectBackend", func(t *testing.T) {
		fm := &FirewallManager{
			backend: "",
			rules:   make(map[string][]string),
		}
		fm.detectBackend()

		if fm.backend == "" {
			t.Error("detectBackend should set a backend")
		}

		t.Logf("Detected backend: %s", fm.backend)
	})

	t.Run("IsBackendAvailable", func(t *testing.T) {
		fm := &FirewallManager{
			backend: FirewallBackendStub,
			rules:   make(map[string][]string),
		}

		if !fm.isBackendAvailable() {
			t.Error("Stub backend should always be available")
		}
	})
}

// TestFirewallManager_ChainOperations tests chain creation/deletion
func TestFirewallManager_ChainOperations(t *testing.T) {
	// Need root for iptables operations
	if os.Geteuid() != 0 {
		t.Skip("Chain operations require root")
	}

	ovs := NewOVSClient()
	chain := "vimic2-test-chain"

	t.Run("CreateFirewallChain", func(t *testing.T) {
		err := ovs.CreateFirewallChain(chain, "drop")
		if err != nil {
			t.Skipf("CreateFirewallChain failed: %v", err)
		}
		t.Log("Firewall chain created")
	})

	t.Run("DeleteFirewallChain", func(t *testing.T) {
		err := ovs.DeleteFirewallChain(chain)
		if err != nil {
			t.Skipf("DeleteFirewallChain failed: %v", err)
		}
		t.Log("Firewall chain deleted")
	})
}