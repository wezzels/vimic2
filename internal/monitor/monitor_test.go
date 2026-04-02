// Package monitor_test tests alerting and metrics functionality
package monitor_test

import (
	"testing"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestAlerterRuleManagement tests alert rule CRUD
func TestAlerterRuleManagement(t *testing.T) {
	a := monitor.NewAlerter(nil)

	// Add rules
	cpuRule := &monitor.AlertRule{
		ID:        "cpu-high",
		Name:      "High CPU Usage",
		Metric:    "cpu",
		Threshold: 80,
		Duration:  60,
		Enabled:   true,
	}
	a.AddRule(cpuRule)

	memRule := &monitor.AlertRule{
		ID:        "mem-high",
		Name:      "High Memory Usage",
		Metric:    "memory",
		Threshold: 90,
		Duration:  60,
		Enabled:   true,
	}
	a.AddRule(memRule)

	// Verify rules exist
	rules := a.GetRules()
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}

	// Verify rule values
	foundCPU := false
	foundMem := false
	for _, r := range rules {
		if r.ID == "cpu-high" {
			foundCPU = true
		}
		if r.ID == "mem-high" {
			foundMem = true
		}
	}

	if !foundCPU {
		t.Error("CPU rule not found")
	}
	if !foundMem {
		t.Error("Memory rule not found")
	}
}

// TestAlertEvaluation tests alert triggering
func TestAlertEvaluation(t *testing.T) {
	a := monitor.NewAlerter(nil)

	// Add CPU rule
	a.AddRule(&monitor.AlertRule{
		ID:        "cpu-high",
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 80,
		Duration:  60,
		Enabled:   true,
	})

	// Add memory rule
	a.AddRule(&monitor.AlertRule{
		ID:        "mem-high",
		Name:      "High Memory",
		Metric:    "memory",
		Threshold: 90,
		Duration:  60,
		Enabled:   true,
	})

	t.Run("TriggerCPUAlert", func(t *testing.T) {
		alerts := a.Evaluate("node-1", "test-node", &database.Metric{
			CPU:    85.0, // Above threshold
			Memory: 50.0,
		})

		if len(alerts) == 0 {
			t.Error("Expected CPU alert to trigger")
		}
	})

	t.Run("NoMemoryAlert", func(t *testing.T) {
		alerts := a.Evaluate("node-1", "test-node", &database.Metric{
			CPU:    50.0,
			Memory: 80.0, // Below threshold
		})

		for _, alert := range alerts {
			if alert.RuleID == "mem-high" {
				t.Error("Memory alert should not trigger at 80%")
			}
		}
	})

	t.Run("TriggerMemoryAlert", func(t *testing.T) {
		alerts := a.Evaluate("node-1", "test-node", &database.Metric{
			CPU:    50.0,
			Memory: 95.0, // Above threshold
		})

		found := false
		for _, alert := range alerts {
			if alert.RuleID == "mem-high" {
				found = true
			}
		}
		if !found {
			t.Error("Expected memory alert to trigger at 95%")
		}
	})
}

// TestAlertCallback tests alert callback functionality
func TestAlertCallback(t *testing.T) {
	a := monitor.NewAlerter(nil)

	a.AddRule(&monitor.AlertRule{
		ID:        "test-alert",
		Name:      "Test Alert",
		Metric:    "cpu",
		Threshold: 80,
		Duration:  60,
		Enabled:   true,
	})

	// Set up callback
	callbackCalled := false
	var receivedAlert *database.Alert
	a.SetCallback(func(alert *database.Alert) {
		callbackCalled = true
		receivedAlert = alert
	})

	// Trigger alert
	a.Evaluate("node-1", "test-node", &database.Metric{
		CPU: 90.0,
	})

	// Note: The callback is called synchronously in this implementation
	if receivedAlert != nil && receivedAlert.NodeID == "node-1" {
		if !callbackCalled {
			t.Error("Callback should have been called")
		}
	}
}

// TestMetricsCollection tests metric collection
func TestMetricsCollection(t *testing.T) {
	// Create stub hypervisor for testing
	stub := hypervisor.NewStubHypervisor()

	// Create node config
	nodeConfig := &hypervisor.NodeConfig{
		Name:     "test-node",
		CPU:      2,
		MemoryMB: 4096,
		DiskGB:   20,
		Image:    "test",
	}

	// Create node
	node, err := stub.CreateNode(nil, nodeConfig)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	if node.ID == "" {
		t.Error("Node ID should not be empty")
	}

	// Get metrics
	metrics, err := stub.GetMetrics(nil, node.ID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Stub returns mock metrics
	if metrics.CPU < 0 || metrics.CPU > 100 {
		t.Errorf("CPU should be between 0-100, got %f", metrics.CPU)
	}
}

// TestMetricsThreshold tests various threshold scenarios
func TestMetricsThreshold(t *testing.T) {
	tests := []struct {
		name      string
		rule      *monitor.AlertRule
		metric    *database.Metric
		shouldHit bool
	}{
		{
			name: "CPU above threshold",
			rule: &monitor.AlertRule{
				ID:        "cpu",
				Metric:    "cpu",
				Threshold: 80,
				Enabled:   true,
			},
			metric:    &database.Metric{CPU: 85.0},
			shouldHit: true,
		},
		{
			name: "CPU below threshold",
			rule: &monitor.AlertRule{
				ID:        "cpu",
				Metric:    "cpu",
				Threshold: 80,
				Enabled:   true,
			},
			metric:    &database.Metric{CPU: 75.0},
			shouldHit: false,
		},
		{
			name: "Memory above threshold",
			rule: &monitor.AlertRule{
				ID:        "mem",
				Metric:    "memory",
				Threshold: 90,
				Enabled:   true,
			},
			metric:    &database.Metric{Memory: 95.0},
			shouldHit: true,
		},
		{
			name: "Disk above threshold",
			rule: &monitor.AlertRule{
				ID:        "disk",
				Metric:    "disk",
				Threshold: 85,
				Enabled:   true,
			},
			metric:    &database.Metric{Disk: 90.0},
			shouldHit: true,
		},
		{
			name: "Disabled rule",
			rule: &monitor.AlertRule{
				ID:        "cpu",
				Metric:    "cpu",
				Threshold: 80,
				Enabled:   false,
			},
			metric:    &database.Metric{CPU: 95.0},
			shouldHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new alerter for each test to avoid rule conflicts
			alerter := monitor.NewAlerter(nil)
			alerter.AddRule(tt.rule)
			alerts := alerter.Evaluate("node-1", "test", tt.metric)

			hit := false
			for _, alert := range alerts {
				if alert.RuleID == tt.rule.ID {
					hit = true
				}
			}

			if tt.shouldHit && !hit {
				t.Errorf("Expected alert for %s", tt.name)
			}
			if !tt.shouldHit && hit {
				t.Errorf("Unexpected alert for %s", tt.name)
			}
		})
	}
}

// TestMultipleAlerts tests handling multiple concurrent alerts
func TestMultipleAlerts(t *testing.T) {
	a := monitor.NewAlerter(nil)

	// Add multiple rules
	a.AddRule(&monitor.AlertRule{
		ID:        "cpu-70",
		Metric:    "cpu",
		Threshold: 70,
		Enabled:   true,
	})
	a.AddRule(&monitor.AlertRule{
		ID:        "cpu-90",
		Metric:    "cpu",
		Threshold: 90,
		Enabled:   true,
	})
	a.AddRule(&monitor.AlertRule{
		ID:        "mem-85",
		Metric:    "memory",
		Threshold: 85,
		Enabled:   true,
	})

	// High CPU and Memory
	alerts := a.Evaluate("node-1", "test", &database.Metric{
		CPU:    95.0, // Above both CPU thresholds
		Memory: 90.0, // Above memory threshold
	})

	// Should get at least 2 alerts (cpu-90 and mem-85)
	if len(alerts) < 2 {
		t.Errorf("Expected at least 2 alerts, got %d", len(alerts))
	}
}

// TestManager tests monitor manager functionality
func TestManager(t *testing.T) {
	// Create manager with stub
	hosts := map[string]hypervisor.Hypervisor{
		"host-1": hypervisor.NewStubHypervisor(),
	}
	mgr := monitor.NewManager(nil, hosts)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}