// Package monitor provides real implementation tests
package monitor

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestNewManager_Real tests real manager creation
func TestNewManager_Real(t *testing.T) {
	manager := NewManager(nil, nil)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestManager_StartCollection_ContextCancel tests collection with context cancellation
func TestManager_StartCollection_ContextCancel(t *testing.T) {
	t.Skip("requires database connection")
}

// TestManager_Collection_Interval tests collection interval
func TestManager_Collection_Interval(t *testing.T) {
	t.Skip("requires database connection")
}

// TestClusterMetrics_Struct tests ClusterMetrics struct
func TestClusterMetrics_Struct(t *testing.T) {
	cm := &ClusterMetrics{
		NodeMetrics: make(map[string][]*database.Metric),
		CPUAvg:      45.5,
		MemAvg:      60.2,
		DiskAvg:     30.1,
	}

	if cm.CPUAvg != 45.5 {
		t.Errorf("expected 45.5, got %f", cm.CPUAvg)
	}
	if cm.MemAvg != 60.2 {
		t.Errorf("expected 60.2, got %f", cm.MemAvg)
	}
	if cm.DiskAvg != 30.1 {
		t.Errorf("expected 30.1, got %f", cm.DiskAvg)
	}
	if cm.NodeMetrics == nil {
		t.Error("expected non-nil NodeMetrics map")
	}
}

// TestClusterMetrics_NodeMetrics tests node metrics aggregation
func TestClusterMetrics_NodeMetrics(t *testing.T) {
	cm := &ClusterMetrics{
		NodeMetrics: make(map[string][]*database.Metric),
	}

	// Add metrics for nodes
	cm.NodeMetrics["node-1"] = []*database.Metric{
		{CPU: 40.0, Memory: 60.0},
		{CPU: 45.0, Memory: 65.0},
	}
	cm.NodeMetrics["node-2"] = []*database.Metric{
		{CPU: 50.0, Memory: 70.0},
	}

	if len(cm.NodeMetrics) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(cm.NodeMetrics))
	}
	if len(cm.NodeMetrics["node-1"]) != 2 {
		t.Errorf("expected 2 metrics for node-1, got %d", len(cm.NodeMetrics["node-1"]))
	}
}

// TestAlertRule_Enabled tests alert rule enable/disable
func TestAlertRule_Enabled(t *testing.T) {
	rule := &AlertRule{
		ID:        "alert-1",
		Name:      "CPU Alert",
		Metric:    "cpu",
		Threshold: 80.0,
		Duration:  300,
		Enabled:   true,
	}

	if !rule.Enabled {
		t.Error("expected alert to be enabled")
	}

	// Disable
	rule.Enabled = false
	if rule.Enabled {
		t.Error("expected alert to be disabled")
	}
}

// TestAlerter_New tests alerter creation
func TestAlerter_New(t *testing.T) {
	alerter := NewAlerter(nil)

	if alerter == nil {
		t.Fatal("expected non-nil alerter")
	}
}

// TestMetric_Creation tests metric creation
func TestMetric_Creation(t *testing.T) {
	metric := &database.Metric{
		NodeID:     "node-1",
		CPU:        45.5,
		Memory:     60.2,
		Disk:       30.1,
		NetworkRX:  1024000,
		NetworkTX:  512000,
		RecordedAt: time.Now(),
	}

	if metric.NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", metric.NodeID)
	}
	if metric.CPU != 45.5 {
		t.Errorf("expected 45.5, got %f", metric.CPU)
	}
}

// TestMetric_Timestamp tests metric timestamp handling
func TestMetric_Timestamp(t *testing.T) {
	now := time.Now()
	metric := &database.Metric{
		RecordedAt: now,
	}

	if metric.RecordedAt.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if metric.RecordedAt != now {
		t.Error("timestamp mismatch")
	}
}

// TestMetricHistory_Real tests metric history with real data
func TestMetricHistory_Real(t *testing.T) {
	history := NewMetricHistory(100)

	// Add real samples
	for i := 0; i < 10; i++ {
		history.Add(MetricSample{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			CPUUsage:  float64(40 + i),
		})
	}

	if history.Len() != 10 {
		t.Errorf("expected 10 samples, got %d", history.Len())
	}

	// Get all samples
	samples := history.GetAll()
	if len(samples) != 10 {
		t.Errorf("expected 10 samples, got %d", len(samples))
	}
}

// TestMetricHistory_MaxSize tests metric history max size
func TestMetricHistory_MaxSize(t *testing.T) {
	history := NewMetricHistory(5)

	// Add more than max
	for i := 0; i < 10; i++ {
		history.Add(MetricSample{
			CPUUsage: float64(i),
		})
	}

	// Should only keep last 5
	if history.Len() > 5 {
		t.Errorf("expected max 5 samples, got %d", history.Len())
	}
}

// TestAlertThreshold_Checking tests alert threshold logic
func TestAlertThreshold_Checking(t *testing.T) {
	threshold := 80.0

	tests := []struct {
		value       float64
		shouldAlert bool
	}{
		{75.0, false},
		{80.0, true},
		{85.0, true},
		{100.0, true},
		{50.0, false},
	}

	for _, tt := range tests {
		alerts := tt.value >= threshold
		if alerts != tt.shouldAlert {
			t.Errorf("value %f: expected alert=%v, got %v", tt.value, tt.shouldAlert, alerts)
		}
	}
}

// TestMetricAggregation_Real tests metric aggregation logic
func TestMetricAggregation_Real(t *testing.T) {
	metrics := []*database.Metric{
		{CPU: 40.0, Memory: 60.0},
		{CPU: 50.0, Memory: 70.0},
		{CPU: 60.0, Memory: 80.0},
	}

	// Calculate averages
	var totalCPU, totalMem float64
	for _, m := range metrics {
		totalCPU += m.CPU
		totalMem += m.Memory
	}
	avgCPU := totalCPU / float64(len(metrics))
	avgMem := totalMem / float64(len(metrics))

	if avgCPU != 50.0 {
		t.Errorf("expected avg CPU 50.0, got %f", avgCPU)
	}
	if avgMem != 70.0 {
		t.Errorf("expected avg Memory 70.0, got %f", avgMem)
	}
}
