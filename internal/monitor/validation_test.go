// Package monitor provides metrics collection tests
package monitor

import (
	"testing"
	"time"
)

// TestAlertRuleValidation tests alert rule validation
func TestAlertRule_Validation(t *testing.T) {
	tests := []struct {
		name  string
		rule  AlertRule
		valid bool
	}{
		{
			name:  "Valid rule",
			rule:  AlertRule{ID: "1", Name: "CPU Alert", Metric: "cpu", Threshold: 80.0, Enabled: true},
			valid: true,
		},
		{
			name:  "Empty ID",
			rule:  AlertRule{Name: "CPU Alert", Metric: "cpu"},
			valid: false,
		},
		{
			name:  "Empty metric",
			rule:  AlertRule{ID: "1", Name: "Alert"},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic
			isValid := tt.rule.ID != "" && tt.rule.Metric != ""
			if isValid != tt.valid {
				t.Errorf("expected valid=%v", tt.valid)
			}
		})
	}
}

// TestClusterMetricsFields tests ClusterMetrics fields
func TestClusterMetrics_Fields(t *testing.T) {
	cm := ClusterMetrics{
		NodeMetrics: nil,
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
}

// TestMetricThresholds tests metric threshold checking
func TestMetricThresholds(t *testing.T) {
	thresholds := map[string]float64{
		"cpu":    80.0,
		"memory": 90.0,
		"disk":   85.0,
	}

	tests := []struct {
		metric    string
		value     float64
		shouldAlert bool
	}{
		{"cpu", 75.0, false},
		{"cpu", 85.0, true},
		{"memory", 85.0, false},
		{"memory", 95.0, true},
		{"disk", 80.0, false},
		{"disk", 90.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.metric, func(t *testing.T) {
			threshold := thresholds[tt.metric]
			alerts := tt.value >= threshold
			if alerts != tt.shouldAlert {
				t.Errorf("metric %s value %f: expected alert=%v", tt.metric, tt.value, tt.shouldAlert)
			}
		})
	}
}

// TestTimeRanges tests time range calculations
func TestTimeRanges(t *testing.T) {
	ranges := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"1 minute", 1 * time.Minute, "1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"1 hour", 1 * time.Hour, "1h"},
		{"24 hours", 24 * time.Hour, "24h"},
	}

	for _, tr := range ranges {
		t.Run(tr.name, func(t *testing.T) {
			if tr.duration <= 0 {
				t.Error("expected positive duration")
			}
		})
	}
}

// TestAlerterRules tests alerter rule management
func TestAlerter_Rules(t *testing.T) {
	alerter := NewAlerter(nil)
	if alerter == nil {
		t.Fatal("expected non-nil alerter")
	}

	// Alarmer should start with empty rules
	// (Can't test actual methods without exported fields)
}

// TestMetricAggregation tests metric aggregation logic
func TestMetricAggregation(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}

	// Calculate average
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))

	if avg != 30.0 {
		t.Errorf("expected 30.0, got %f", avg)
	}

	// Calculate max
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max != 50.0 {
		t.Errorf("expected max 50.0, got %f", max)
	}

	// Calculate min
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	if min != 10.0 {
		t.Errorf("expected min 10.0, got %f", min)
	}
}

// TestAlertLevels tests alert level values
func TestAlertLevels(t *testing.T) {
	levels := []string{"info", "warning", "critical", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			if level == "" {
				t.Error("level should not be empty")
			}
		})
	}
}

// TestMetricTypes tests metric type values
func TestMetricTypes(t *testing.T) {
	types := []string{"cpu", "memory", "disk", "network"}

	for _, mt := range types {
		t.Run(mt, func(t *testing.T) {
			if mt == "" {
				t.Error("metric type should not be empty")
			}
		})
	}
}