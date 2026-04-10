// Package orchestrator provides unit tests for orchestrator types
package orchestrator

import (
	"testing"
)

// TestUpdateStrategy_Valid tests update strategy validation
func TestUpdateStrategy_Valid(t *testing.T) {
	tests := []struct {
		name     string
		strategy UpdateStrategy
		expected bool
	}{
		{
			name: "valid strategy",
			strategy: UpdateStrategy{
				BatchSize:      3,
				MaxUnavailable: 2,
				WaitBetween:    30,
			},
			expected: true,
		},
		{
			name: "zero batch size",
			strategy: UpdateStrategy{
				BatchSize:      0,
				MaxUnavailable: 2,
				WaitBetween:    30,
			},
			expected: false,
		},
		{
			name: "negative batch size",
			strategy: UpdateStrategy{
				BatchSize:      -1,
				MaxUnavailable: 2,
				WaitBetween:    30,
			},
			expected: false,
		},
		{
			name: "zero max unavailable (valid)",
			strategy: UpdateStrategy{
				BatchSize:      3,
				MaxUnavailable: 0,
				WaitBetween:    30,
			},
			expected: true,
		},
		{
			name: "zero wait between (valid)",
			strategy: UpdateStrategy{
				BatchSize:      3,
				MaxUnavailable: 2,
				WaitBetween:    0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestRollingUpdater_CalculateBatches tests batch calculation
func TestRollingUpdater_CalculateBatches(t *testing.T) {
	updater := &RollingUpdater{}

	tests := []struct {
		name      string
		nodes     []string
		batchSize int
		expected  int
	}{
		{
			name:      "10 nodes, batch 3",
			nodes:     []string{"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10"},
			batchSize: 3,
			expected:  4, // 3+3+3+1
		},
		{
			name:      "5 nodes, batch 2",
			nodes:     []string{"n1", "n2", "n3", "n4", "n5"},
			batchSize: 2,
			expected:  3, // 2+2+1
		},
		{
			name:      "3 nodes, batch 3",
			nodes:     []string{"n1", "n2", "n3"},
			batchSize: 3,
			expected:  1, // 3
		},
		{
			name:      "empty nodes",
			nodes:     []string{},
			batchSize: 3,
			expected:  0,
		},
		{
			name:      "zero batch size",
			nodes:     []string{"n1", "n2", "n3"},
			batchSize: 0,
			expected:  0,
		},
		{
			name:      "negative batch size",
			nodes:     []string{"n1", "n2", "n3"},
			batchSize: -1,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := updater.CalculateBatches(tt.nodes, tt.batchSize)
			if tt.expected == 0 {
				if batches != nil {
					t.Errorf("expected nil, got %d batches", len(batches))
				}
				return
			}
			if len(batches) != tt.expected {
				t.Errorf("expected %d batches, got %d", tt.expected, len(batches))
			}
		})
	}
}

// TestRollingUpdater_BatchContents tests batch contents
func TestRollingUpdater_BatchContents(t *testing.T) {
	updater := &RollingUpdater{}

	nodes := []string{"n1", "n2", "n3", "n4", "n5"}
	batches := updater.CalculateBatches(nodes, 2)

	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}

	// First batch: n1, n2
	if len(batches[0]) != 2 || batches[0][0] != "n1" || batches[0][1] != "n2" {
		t.Errorf("first batch incorrect: %v", batches[0])
	}

	// Second batch: n3, n4
	if len(batches[1]) != 2 || batches[1][0] != "n3" || batches[1][1] != "n4" {
		t.Errorf("second batch incorrect: %v", batches[1])
	}

	// Third batch: n5
	if len(batches[2]) != 1 || batches[2][0] != "n5" {
		t.Errorf("third batch incorrect: %v", batches[2])
	}
}

// TestRollingUpdater_SetStrategy tests strategy setting
func TestRollingUpdater_SetStrategy(t *testing.T) {
	updater := &RollingUpdater{}

	strategy := &UpdateStrategy{
		BatchSize:      5,
		MaxUnavailable: 2,
		WaitBetween:    60,
	}

	updater.SetStrategy(strategy)

	if updater.strategy == nil {
		t.Error("strategy not set")
	}
	if updater.strategy.BatchSize != 5 {
		t.Errorf("expected batch size 5, got %d", updater.strategy.BatchSize)
	}
}