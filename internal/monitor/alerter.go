// Package monitor provides alerting functionality
package monitor

import (
	"fmt"
	"time"

	"github.com/stsgym/vimic2/internal/database"
)

// AlertRule defines conditions that trigger alerts
type AlertRule struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Metric    string  `json:"metric"` // cpu, memory, disk
	Threshold float64 `json:"threshold"`
	Duration  int     `json:"duration"` // seconds the condition must persist
	Enabled   bool    `json:"enabled"`
}

// Alerter handles alert evaluation and notification
type Alerter struct {
	rules    []*AlertRule
	alerts   map[string]*database.Alert // nodeID -> current alert
	db       *database.DB
	callback func(*database.Alert)
}

// NewAlerter creates a new alerter
func NewAlerter(db *database.DB) *Alerter {
	return &Alerter{
		rules:  make([]*AlertRule, 0),
		alerts: make(map[string]*database.Alert),
		db:     db,
	}
}

// AddRule adds an alert rule
func (a *Alerter) AddRule(rule *AlertRule) {
	a.rules = append(a.rules, rule)
}

// GetRules returns all alert rules
func (a *Alerter) GetRules() []*AlertRule {
	return a.rules
}

// SetCallback sets the alert callback function
func (a *Alerter) SetCallback(cb func(*database.Alert)) {
	a.callback = cb
}

// Evaluate checks metrics against rules and fires alerts
func (a *Alerter) Evaluate(nodeID, nodeName string, m *database.Metric) []*database.Alert {
	var fired []*database.Alert

	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		value := a.getMetricValue(m, rule.Metric)
		if value >= rule.Threshold {
			// Check if we already have an active alert for this node/rule
			key := fmt.Sprintf("%s-%s", nodeID, rule.ID)
			if existing, ok := a.alerts[key]; ok && !existing.Resolved {
				fired = append(fired, existing)
				continue
			}

			// Create new alert
			alert := &database.Alert{
				ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
				RuleID:    rule.ID,
				NodeID:    nodeID,
				NodeName:  nodeName,
				Metric:    rule.Metric,
				Value:     value,
				Threshold: rule.Threshold,
				Message:   fmt.Sprintf("High %s on %s (%.1f%%)", rule.Metric, nodeName, value),
				FiredAt:   time.Now(),
				Resolved:  false,
			}

			key = fmt.Sprintf("%s-%s", nodeID, rule.ID)
			a.alerts[key] = alert

			// Save to database if available
			if a.db != nil {
				a.db.SaveAlert(alert)
			}

			// Fire callback if set
			if a.callback != nil {
				a.callback(alert)
			}

			fired = append(fired, alert)
		}
	}

	return fired
}

// ResolveAlert marks an alert as resolved
func (a *Alerter) ResolveAlert(nodeID, ruleID string) {
	key := fmt.Sprintf("%s-%s", nodeID, ruleID)
	if alert, ok := a.alerts[key]; ok {
		alert.Resolved = true
		now := time.Now()
		alert.ResolvedAt = &now

		// Update in database
		if a.db != nil {
			a.db.SaveAlert(alert)
		}
	}
}

// getMetricValue extracts metric value by name
func (a *Alerter) getMetricValue(m *database.Metric, metric string) float64 {
	switch metric {
	case "cpu":
		return m.CPU
	case "memory":
		return m.Memory
	case "disk":
		return m.Disk
	case "network_rx":
		return m.NetworkRX
	case "network_tx":
		return m.NetworkTX
	default:
		return 0
	}
}

// DefaultRules returns the default alert rules
func DefaultRules() []*AlertRule {
	return []*AlertRule{
		{
			ID:        "high-cpu",
			Name:      "High CPU",
			Metric:    "cpu",
			Threshold: 90,
			Duration:  300, // 5 minutes
			Enabled:   true,
		},
		{
			ID:        "high-memory",
			Name:      "High Memory",
			Metric:    "memory",
			Threshold: 90,
			Duration:  300,
			Enabled:   true,
		},
		{
			ID:        "disk-full",
			Name:      "Disk Full",
			Metric:    "disk",
			Threshold: 95,
			Duration:  60,
			Enabled:   true,
		},
		{
			ID:        "node-down",
			Name:      "Node Down",
			Metric:    "heartbeat",
			Threshold: 1, // 1 = missing (binary indicator)
			Duration:  60,
			Enabled:   true,
		},
	}
}
