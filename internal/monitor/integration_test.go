//go:build integration

// Package monitor provides integration tests for monitoring
package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/pkg/hypervisor"
)

// TestIntegration_Monitor_CreateManager tests creating a monitor manager
func TestIntegration_Monitor_CreateManager(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}

	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}

	t.Logf("Created monitor manager")
}

// TestIntegration_Monitor_GetNodeMetrics tests getting node metrics
func TestIntegration_Monitor_GetNodeMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	metrics, err := mgr.GetNodeMetrics("nonexistent-node", time.Hour)
	if err != nil {
		t.Logf("GetNodeMetrics failed (expected): %v", err)
	} else {
		t.Logf("Got %d metrics", len(metrics))
	}
}

// TestIntegration_Monitor_GetClusterMetrics tests getting cluster metrics
func TestIntegration_Monitor_GetClusterMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	metrics, err := mgr.GetClusterMetrics("nonexistent-cluster", time.Hour)
	if err != nil {
		t.Logf("GetClusterMetrics failed (expected): %v", err)
	} else {
		t.Logf("Got cluster metrics: %+v", metrics)
	}
}

// TestIntegration_Monitor_StartCollection tests starting collection
func TestIntegration_Monitor_StartCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	hosts := make(map[string]hypervisor.Hypervisor)
	mgr := NewManager(db, hosts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mgr.StartCollection(ctx, 1*time.Second)

	t.Logf("Collection started and stopped")
}

// TestIntegration_Alerter_Create tests creating an alerter
func TestIntegration_Alerter_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	alerter := NewAlerter(db)

	if alerter == nil {
		t.Fatal("expected non-nil alerter")
	}

	t.Logf("Created alerter")
}

// TestIntegration_Alerter_AddRule tests adding alert rules
func TestIntegration_Alerter_AddRule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	alerter := NewAlerter(db)

	rule := &AlertRule{
		ID:        "test-rule-1",
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 80,
		Duration:  300,
		Enabled:   true,
	}

	alerter.AddRule(rule)

	rules := alerter.GetRules()
	if len(rules) < 1 {
		t.Error("expected at least 1 rule")
	}

	t.Logf("Added rule: %s", rule.Name)
}

// TestIntegration_Alerter_DefaultRules tests default rules
func TestIntegration_Alerter_DefaultRules(t *testing.T) {
	rules := DefaultRules()

	if len(rules) == 0 {
		t.Error("expected default rules")
	}

	t.Logf("Got %d default rules", len(rules))
	for _, r := range rules {
		t.Logf("  - %s: %s > %.1f", r.Name, r.Metric, r.Threshold)
	}
}

// TestIntegration_Alerter_Evaluate tests alert evaluation
func TestIntegration_Alerter_Evaluate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	alerter := NewAlerter(db)

	rule := &AlertRule{
		ID:        "eval-rule",
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 50,
		Duration:  0,
		Enabled:   true,
	}
	alerter.AddRule(rule)

	metric := &database.Metric{
		NodeID:    "test-node",
		CPU:       85.0,
		Memory:    50.0,
		Disk:      40.0,
		NetworkRX: 1000,
		NetworkTX: 500,
	}

	alerts := alerter.Evaluate("test-node", "TestNode", metric)
	t.Logf("Generated %d alerts", len(alerts))
}

// TestIntegration_Alerter_Callback tests alert callbacks
func TestIntegration_Alerter_Callback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	alerter := NewAlerter(db)

	alertReceived := false
	alerter.SetCallback(func(alert *database.Alert) {
		alertReceived = true
		t.Logf("Alert received: %s", alert.Message)
	})

	rule := &AlertRule{
		ID:        "callback-rule",
		Name:      "High Memory",
		Metric:    "memory",
		Threshold: 50,
		Duration:  0,
		Enabled:   true,
	}
	alerter.AddRule(rule)

	metric := &database.Metric{
		NodeID:    "callback-node",
		CPU:       20.0,
		Memory:    90.0,
		Disk:      30.0,
	}

	alerter.Evaluate("callback-node", "CallbackNode", metric)

	if alertReceived {
		t.Logf("Callback was triggered")
	}
}

// TestIntegration_Alerter_Resolve tests resolving alerts
func TestIntegration_Alerter_Resolve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, _ := database.NewDB(":memory:")
	alerter := NewAlerter(db)

	alerter.ResolveAlert("nonexistent-node", "nonexistent-rule")

	t.Logf("ResolveAlert completed without error")
}