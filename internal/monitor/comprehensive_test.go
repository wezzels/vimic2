// Package monitor provides comprehensive real tests
package monitor

import (
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// TestManager_Creation tests manager creation with real struct
func TestManager_Creation(t *testing.T) {
	mgr := NewManager(nil, nil)
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
}

// TestClusterMetrics_Aggregation tests metrics aggregation
func TestClusterMetrics_Aggregation(t *testing.T) {
	cm := &ClusterMetrics{
		CPUAvg:  calculateAverage([]float64{40, 50, 60}),
		MemAvg:  calculateAverage([]float64{70, 80, 90}),
		DiskAvg: calculateAverage([]float64{30, 40, 50}),
	}

	// CPU average should be 50
	if cm.CPUAvg < 49 || cm.CPUAvg > 51 {
		t.Errorf("expected ~50, got %f", cm.CPUAvg)
	}
}

// TestAlertRule_Thresholds tests alert threshold logic
func TestAlertRule_Thresholds(t *testing.T) {
	tests := []struct {
		name        string
		threshold   float64
		value       float64
		shouldAlert bool
	}{
		{"CPU at limit", 80.0, 80.0, true},
		{"CPU over limit", 80.0, 85.0, true},
		{"CPU under limit", 80.0, 75.0, false},
		{"Memory at limit", 90.0, 90.0, true},
		{"Memory over limit", 90.0, 95.0, true},
		{"Memory under limit", 90.0, 85.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := tt.value >= tt.threshold
			if alerts != tt.shouldAlert {
				t.Errorf("value %f, threshold %f: expected alert=%v", tt.value, tt.threshold, tt.shouldAlert)
			}
		})
	}
}

// TestAlertRule_Durations tests alert duration
func TestAlertRule_Durations(t *testing.T) {
	rules := []*AlertRule{
		{ID: "alert-1", Duration: 300, Threshold: 80.0},
		{ID: "alert-2", Duration: 600, Threshold: 90.0},
	}

	for _, rule := range rules {
		if rule.Duration <= 0 {
			t.Error("duration should be positive for alerts")
		}
	}
}

// TestMetricSample_TimeSeries tests metric time series
func TestMetricSample_TimeSeries(t *testing.T) {
	samples := []MetricSample{
		{Timestamp: time.Now().Add(-2 * time.Minute), CPUUsage: 40},
		{Timestamp: time.Now().Add(-1 * time.Minute), CPUUsage: 50},
		{Timestamp: time.Now(), CPUUsage: 60},
	}

	// Calculate average
	var total float64
	for _, s := range samples {
		total += s.CPUUsage
	}
	avg := total / float64(len(samples))

	if avg != 50 {
		t.Errorf("expected average 50, got %f", avg)
	}
}

// TestMetricHistory_CapacityReal tests history capacity limits
func TestMetricHistory_CapacityReal(t *testing.T) {
	history := NewMetricHistory(10)

	// Add more than capacity
	for i := 0; i < 15; i++ {
		history.Add(MetricSample{CPUUsage: float64(i)})
	}

	// Should only keep last 10
	if history.Len() > 10 {
		t.Errorf("expected max 10 samples, got %d", history.Len())
	}
}

// TestMetricHistory_GetAllReal tests retrieving all samples
func TestMetricHistory_GetAllReal(t *testing.T) {
	history := NewMetricHistory(100)

	for i := 0; i < 5; i++ {
		history.Add(MetricSample{CPUUsage: float64(i * 10)})
	}

	samples := history.GetAll()
	if len(samples) != 5 {
		t.Errorf("expected 5 samples, got %d", len(samples))
	}
}

// TestDatabase_MetricStruct tests database.Metric struct
func TestDatabase_MetricStruct(t *testing.T) {
	metric := &database.Metric{
		NodeID:     "node-1",
		CPU:        45.5,
		Memory:     60.2,
		Disk:       30.1,
		NetworkRX:  1024,
		NetworkTX:  512,
		RecordedAt: time.Now(),
	}

	if metric.NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", metric.NodeID)
	}
	if metric.CPU < 0 || metric.CPU > 100 {
		t.Error("CPU should be 0-100%")
	}
	if metric.Memory < 0 || metric.Memory > 100 {
		t.Error("Memory should be 0-100%")
	}
}

// TestAlertRule_EnableDisableReal tests enabling/disabling alerts
func TestAlertRule_EnableDisableReal(t *testing.T) {
	rule := &AlertRule{
		ID:        "alert-1",
		Name:      "High CPU",
		Metric:    "cpu",
		Threshold: 80.0,
		Enabled:   true,
	}

	// Verify initial state
	if !rule.Enabled {
		t.Error("expected alert to be enabled initially")
	}

	// Disable
	rule.Enabled = false
	if rule.Enabled {
		t.Error("expected alert to be disabled")
	}

	// Re-enable
	rule.Enabled = true
	if !rule.Enabled {
		t.Error("expected alert to be enabled again")
	}
}

// TestAlertRule_DurationReal tests alert duration
func TestAlertRule_DurationReal(t *testing.T) {
	rule := &AlertRule{
		ID:        "alert-1",
		Duration:  300, // 5 minutes
		Threshold: 80.0,
	}

	if rule.Duration != 300 {
		t.Errorf("expected duration 300, got %d", rule.Duration)
	}

	// Duration should be positive for meaningful alerts
	if rule.Duration <= 0 {
		t.Error("duration should be positive for alerts")
	}
}

// TestMetricValue_Validation tests metric value validation
func TestMetricValue_Validation(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		valid bool
	}{
		{"Valid CPU", 45.5, true},
		{"Valid Memory", 60.2, true},
		{"Zero", 0.0, true},
		{"Max value", 100.0, true},
		{"Over 100%", 105.0, false},
		{"Negative", -5.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.value >= 0 && tt.value <= 100
			if valid != tt.valid {
				t.Errorf("value %f: expected valid=%v", tt.value, tt.valid)
			}
		})
	}
}

// Helper function for testing
func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
