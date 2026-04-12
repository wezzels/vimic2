// Package monitor provides alerting functionality
package monitor

import (
	"os"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

func TestAlerter(t *testing.T) {
	// Create a real temporary database for testing
	db, err := database.NewDB("/tmp/vimic2_test_alerts.db")
	if err != nil {
		t.Skip("Could not create test database")
	}
	defer db.Close()
	defer os.Remove("/tmp/vimic2_test_alerts.db")

	alerter := NewAlerter(db)

	if alerter == nil {
		t.Fatal("Alerter should not be nil")
	}

	// Test adding rules
	alerter.AddRule(&AlertRule{
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 90,
		Duration:  300,
		Enabled:   true,
	})

	alerter.AddRule(&AlertRule{
		Name:      "High Memory",
		Metric:    "memory",
		Threshold: 90,
		Duration:  300,
		Enabled:   true,
	})

	// Test default rules
	defaults := DefaultRules()
	if len(defaults) != 4 {
		t.Errorf("Expected 4 default rules, got %d", len(defaults))
	}

	// Test evaluate
	metric := &database.Metric{
		NodeID: "test-node",
		CPU:    95.0,
		Memory: 50.0,
		Disk:   30.0,
	}

	alerts := alerter.Evaluate("test-node", "test-node", metric)

	// Should have at least the high CPU alert
	if len(alerts) == 0 {
		t.Error("Expected at least 1 alert for high CPU")
	}

	// Test getMetricValue
	value := alerter.getMetricValue(metric, "cpu")
	if value != 95.0 {
		t.Errorf("Expected 95.0, got %.1f", value)
	}

	value = alerter.getMetricValue(metric, "memory")
	if value != 50.0 {
		t.Errorf("Expected 50.0, got %.1f", value)
	}
}

func TestAlertRule(t *testing.T) {
	rule := &AlertRule{
		ID:        "test-rule",
		Name:      "Test Rule",
		Metric:    "cpu",
		Threshold: 85.0,
		Duration:  600,
		Enabled:   true,
	}

	if rule.ID != "test-rule" {
		t.Errorf("Expected ID 'test-rule', got '%s'", rule.ID)
	}
	if rule.Threshold != 85.0 {
		t.Errorf("Expected 85.0, got %.1f", rule.Threshold)
	}
	if !rule.Enabled {
		t.Error("Expected rule to be enabled")
	}
}

func TestAlert(t *testing.T) {
	now := time.Now()
	alert := &database.Alert{
		ID:        "alert-1",
		RuleID:    "rule-1",
		NodeID:    "node-1",
		NodeName:  "test-node",
		Metric:    "cpu",
		Value:     95.0,
		Threshold: 90.0,
		Message:   "High CPU on test-node (95%)",
		FiredAt:   now,
		Resolved:  false,
	}

	if alert.Value != 95.0 {
		t.Errorf("Expected 95.0, got %.1f", alert.Value)
	}
	if alert.Resolved {
		t.Error("Alert should not be resolved")
	}

	// Test resolution
	alert.Resolved = true
	resolvedAt := time.Now()
	alert.ResolvedAt = &resolvedAt

	if !alert.Resolved {
		t.Error("Alert should be resolved")
	}
	if alert.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()

	if len(rules) != 4 {
		t.Errorf("Expected 4 rules, got %d", len(rules))
	}

	expectedMetrics := map[string]bool{
		"cpu":    false,
		"memory": false,
		"disk":   false,
	}

	for _, rule := range rules {
		if rule.Name == "" {
			t.Error("Rule name should not be empty")
		}
		if rule.Duration <= 0 {
			t.Errorf("Rule %s duration should be positive", rule.Name)
		}
		// Threshold can be 0 for some metrics like heartbeat (node down = 0)
		if rule.Metric != "heartbeat" && rule.Threshold <= 0 {
			t.Errorf("Rule %s threshold should be positive for metric %s", rule.Name, rule.Metric)
		}
		expectedMetrics[rule.Metric] = true
	}

	for metric, found := range expectedMetrics {
		if !found {
			t.Errorf("Missing rule for metric: %s", metric)
		}
	}
}

func TestAlerterCallback(t *testing.T) {
	db, err := database.NewDB("/tmp/vimic2_test_callback.db")
	if err != nil {
		t.Skip("Could not create test database")
	}
	defer db.Close()
	defer os.Remove("/tmp/vimic2_test_callback.db")

	alerter := NewAlerter(db)

	callbackCalled := false
	alerter.SetCallback(func(a *database.Alert) {
		callbackCalled = true
	})

	// Add a rule that will fire
	alerter.AddRule(&AlertRule{
		Name:      "Test",
		Metric:    "cpu",
		Threshold: 10, // Low threshold to trigger
		Duration:  0,  // Immediate
		Enabled:   true,
	})

	metric := &database.Metric{CPU: 95}
	alerter.Evaluate("node", "node", metric)

	if !callbackCalled {
		t.Error("Callback should have been called")
	}
}

func TestAlerterNoDB(t *testing.T) {
	// Test with nil DB
	alerter := NewAlerter(nil)

	alerter.AddRule(&AlertRule{
		Name:      "Test",
		Metric:    "cpu",
		Threshold: 50,
		Enabled:   true,
	})

	metric := &database.Metric{CPU: 95}
	alerts := alerter.Evaluate("node", "node", metric)

	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
}
