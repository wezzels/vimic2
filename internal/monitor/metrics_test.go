// Package monitor provides metrics collection and alerting tests
package monitor

import (
	"context"
	"testing"
	"time"
)

// TestNewManager tests manager creation
func TestNewManager_Basic(t *testing.T) {
	mgr := NewManager(nil, nil)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestManager_StartCollection_Basic tests metrics collection start
func TestManager_StartCollection_Basic(t *testing.T) {
	mgr := NewManager(nil, nil)

	// Start with context that we cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should not panic - just returns immediately
	mgr.StartCollection(ctx, 1*time.Second)
}

// TestClusterMetricsStruct tests ClusterMetrics struct
func TestClusterMetricsStruct(t *testing.T) {
	metrics := &ClusterMetrics{
		CPUAvg:  45.5,
		MemAvg:  62.3,
		DiskAvg: 30.1,
	}

	if metrics.CPUAvg != 45.5 {
		t.Errorf("expected 45.5, got %f", metrics.CPUAvg)
	}
	if metrics.MemAvg != 62.3 {
		t.Errorf("expected 62.3, got %f", metrics.MemAvg)
	}
}

// TestAlertRuleStruct tests AlertRule struct
func TestAlertRuleStruct(t *testing.T) {
	rule := &AlertRule{
		ID:        "alert-1",
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 80.0,
		Duration:  300,
		Enabled:   true,
	}

	if rule.ID != "alert-1" {
		t.Errorf("expected alert-1, got %s", rule.ID)
	}
	if rule.Metric != "cpu" {
		t.Errorf("expected cpu, got %s", rule.Metric)
	}
	if rule.Threshold != 80.0 {
		t.Errorf("expected 80.0, got %f", rule.Threshold)
	}
	if !rule.Enabled {
		t.Error("expected enabled to be true")
	}
}

// TestNewAlerter tests alerter creation
func TestNewAlerter(t *testing.T) {
	alerter := NewAlerter(nil)
	if alerter == nil {
		t.Fatal("expected non-nil alerter")
	}
}

// TestAlertRule_EnableDisable tests enable/disable
func TestAlertRule_EnableDisable(t *testing.T) {
	rule := &AlertRule{Enabled: true}

	if !rule.Enabled {
		t.Error("expected enabled")
	}

	rule.Enabled = false
	if rule.Enabled {
		t.Error("expected disabled")
	}
}

// TestMetricHistory tests existing MetricHistory
func TestMetricHistory_Basic(t *testing.T) {
	history := NewMetricHistory(10)

	// Add samples
	for i := 0; i < 15; i++ {
		history.Add(MetricSample{
			Timestamp: time.Now(),
			CPUUsage:  float64(i),
		})
	}

	// Should only keep last 10
	if history.Len() != 10 {
		t.Errorf("expected 10 samples, got %d", history.Len())
	}
}

// TestMetricHistory_GetAll tests sample retrieval
func TestMetricHistory_GetAll(t *testing.T) {
	history := NewMetricHistory(100)

	for i := 0; i < 5; i++ {
		history.Add(MetricSample{
			CPUUsage: float64(i * 10),
		})
	}

	samples := history.GetAll()
	if len(samples) != 5 {
		t.Errorf("expected 5 samples, got %d", len(samples))
	}
}

// TestMetricSampleStruct tests MetricSample struct
func TestMetricSampleStruct(t *testing.T) {
	sample := MetricSample{
		Timestamp: time.Now(),
		CPUUsage:  75.5,
	}

	if sample.CPUUsage != 75.5 {
		t.Errorf("expected 75.5, got %f", sample.CPUUsage)
	}
}
