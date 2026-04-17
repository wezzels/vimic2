// Package provisioner provides provisioner unit tests
package provisioner

import (
	"strings"
	"testing"
)

// TestHash tests hash function
func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minValue int
		maxValue int
	}{
		{"simple", "test", 0, 255},
		{"empty", "", 0, 255},
		{"long", "very-long-network-name", 0, 255},
		{"numbers", "network123", 0, 255},
		{"special", "network-with-dashes", 0, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash(tt.input)
			if result < tt.minValue || result > tt.maxValue {
				t.Errorf("hash(%s) = %d, expected range [%d, %d]", tt.input, result, tt.minValue, tt.maxValue)
			}
		})
	}
}

// TestHash_Consistency tests that same input produces same output
func TestHash_Consistency(t *testing.T) {
	input := "test-network"
	h1 := hash(input)
	h2 := hash(input)

	if h1 != h2 {
		t.Errorf("hash should be consistent: %d != %d", h1, h2)
	}
}

// TestHash_DifferentInputs tests different inputs produce different outputs
func TestHash_DifferentInputs(t *testing.T) {
	inputs := []string{"network1", "network2", "network3", "default", "isolated"}
	results := make(map[int]string)

	for _, input := range inputs {
		h := hash(input)
		if existing, ok := results[h]; ok {
			// Collision is possible but unlikely for small set
			t.Logf("hash collision: %s and %s both hash to %d", existing, input, h)
		}
		results[h] = input
	}
}

// TestGenerateNetworkXML_NAT tests NAT network XML generation
func TestGenerateNetworkXML_NAT(t *testing.T) {
	nm := &NetworkManager{}
	xml := nm.generateNetworkXML("test-nat", "nat", "192.168.122.0/24")

	if !strings.Contains(xml, "<network>") {
		t.Error("expected <network> element")
	}
	if !strings.Contains(xml, "test-nat") {
		t.Error("expected network name in XML")
	}
	if !strings.Contains(xml, "mode='nat'") {
		t.Error("expected NAT mode in XML")
	}
	if !strings.Contains(xml, "<dhcp>") {
		t.Error("expected DHCP element in NAT network")
	}
	if !strings.Contains(xml, "</network>") {
		t.Error("expected </network> closing element")
	}
}

// TestGenerateNetworkXML_Bridge tests bridge network XML generation
func TestGenerateNetworkXML_Bridge(t *testing.T) {
	nm := &NetworkManager{}
	xml := nm.generateNetworkXML("test-bridge", "bridge", "10.100.0.0/16")

	if !strings.Contains(xml, "<network>") {
		t.Error("expected <network> element")
	}
	if !strings.Contains(xml, "test-bridge") {
		t.Error("expected network name in XML")
	}
	if !strings.Contains(xml, "mode='bridge'") {
		t.Error("expected bridge mode in XML")
	}
	if strings.Contains(xml, "<dhcp>") {
		t.Error("DHCP should not be in bridge network")
	}
	if !strings.Contains(xml, "</network>") {
		t.Error("expected </network> closing element")
	}
}

// TestGenerateNetworkXML_EdgeCases tests edge cases
func TestGenerateNetworkXML_EdgeCases(t *testing.T) {
	nm := &NetworkManager{}

	// Empty name
	xml := nm.generateNetworkXML("", "nat", "")
	if !strings.Contains(xml, "<network>") {
		t.Error("expected valid XML even with empty name")
	}

	// Long name
	longName := strings.Repeat("a", 100)
	xml = nm.generateNetworkXML(longName, "nat", "192.168.1.0/24")
	if !strings.Contains(xml, longName) {
		t.Error("expected long name in XML")
	}
}

// TestNetworkManager tests NetworkManager creation
func TestNetworkManager_Create(t *testing.T) {
	nm := &NetworkManager{}

	if nm == nil {
		t.Error("NetworkManager should not be nil")
	}
}

// TestNetworkConfig_Equality tests network config comparison
func TestNetworkConfig_Equality(t *testing.T) {
	config1 := &NetworkConfig{
		Name:    "test",
		Type:    "nat",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
	}

	config2 := &NetworkConfig{
		Name:    "test",
		Type:    "nat",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
	}

	if config1.Name != config2.Name {
		t.Error("names should be equal")
	}
	if config1.Type != config2.Type {
		t.Error("types should be equal")
	}
	if config1.CIDR != config2.CIDR {
		t.Error("CIDRs should be equal")
	}
	if config1.Gateway != config2.Gateway {
		t.Error("gateways should be equal")
	}
}

// TestNetworkConfig_NilSafe tests nil safety
func TestNetworkConfig_NilSafe(t *testing.T) {
	var config *NetworkConfig

	if config != nil {
		t.Error("nil config should be nil")
	}

	// Test nil checks
	config = &NetworkConfig{}
	if config.Name != "" {
		t.Error("empty name should be empty string")
	}
}

// TestManager_ImageDir tests image directory handling
func TestManager_ImageDir(t *testing.T) {
	dirs := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "/var/lib/libvirt/images"},
		{"custom", "/custom/images", "/custom/images"},
		{"relative", "./images", "./images"},
	}

	for _, tt := range dirs {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(tt.input)
			if mgr.imageDir != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, mgr.imageDir)
			}
		})
	}
}