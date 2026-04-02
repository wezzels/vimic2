// Package orchestrator provides auto-scaling and orchestration
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/cluster"
	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/monitor"
	"go.uber.org/zap"
)

// AutoScaler handles automatic cluster scaling
type AutoScaler struct {
	clusterMgr *cluster.Manager
	monitorMgr *monitor.Manager
	db         *database.DB
	rules      map[string]*ScaleRule
	sugar      *zap.SugaredLogger
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// ScaleRule defines when to scale
type ScaleRule struct {
	ClusterID       string
	Metric          string  // cpu, memory
	UpperThreshold  float64 // Scale up when above
	LowerThreshold  float64 // Scale down when below
	ScaleUpCount    int     // Nodes to add
	ScaleDownCount  int     // Nodes to remove
	Cooldown        time.Duration
	LastScaleUp     time.Time
	LastScaleDown   time.Time
	Enabled         bool
}

// NewAutoScaler creates a new auto-scaler
func NewAutoScaler(clusterMgr *cluster.Manager, monitorMgr *monitor.Manager, sugar *zap.SugaredLogger) *AutoScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &AutoScaler{
		clusterMgr: clusterMgr,
		monitorMgr: monitorMgr,
		db:         nil,
		rules:      make(map[string]*ScaleRule),
		sugar:      sugar,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetDB sets the database
func (a *AutoScaler) SetDB(db *database.DB) {
	a.db = db
}

// AddRule adds a scale rule for a cluster
func (a *AutoScaler) AddRule(rule *ScaleRule) {
	a.mu.Lock()
	defer a.mu.Unlock()
	rule.LastScaleUp = time.Time{}
	rule.LastScaleDown = time.Time{}
	a.rules[rule.ClusterID] = rule
}

// RemoveRule removes a scale rule
func (a *AutoScaler) RemoveRule(clusterID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.rules, clusterID)
}

// Start begins the auto-scaling loop
func (a *AutoScaler) Start(interval time.Duration) {
	go a.runLoop(interval)
}

// Stop stops the auto-scaling loop
func (a *AutoScaler) Stop() {
	a.cancel()
}

func (a *AutoScaler) runLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.evaluateAll()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *AutoScaler) evaluateAll() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for clusterID, rule := range a.rules {
		if !rule.Enabled {
			continue
		}
		if err := a.evaluate(clusterID, rule); err != nil {
			a.sugar.Warnw("AutoScale evaluation failed", "cluster", clusterID, "error", err)
		}
	}
}

func (a *AutoScaler) evaluate(clusterID string, rule *ScaleRule) error {
	// Get cluster metrics
	metrics, err := a.monitorMgr.GetClusterMetrics(clusterID, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	var currentValue float64
	switch rule.Metric {
	case "cpu":
		currentValue = metrics.CPUAvg
	case "memory":
		currentValue = metrics.MemAvg
	case "disk":
		currentValue = metrics.DiskAvg
	default:
		return fmt.Errorf("unknown metric: %s", rule.Metric)
	}

	// Check if we can scale
	canScaleUp := time.Since(rule.LastScaleUp) > rule.Cooldown
	canScaleDown := time.Since(rule.LastScaleDown) > rule.Cooldown

	// Get current node count
	cluster, err := a.clusterMgr.GetCluster(clusterID)
	if err != nil {
		return err
	}
	
	maxNodes := 100
	minNodes := 1
	if cluster.Config != nil {
		maxNodes = cluster.Config.MaxNodes
		minNodes = cluster.Config.MinNodes
	}

	nodes, _ := a.db.ListClusterNodes(clusterID)
	currentNodes := len(nodes)

	// Scale up
	if currentValue >= rule.UpperThreshold && canScaleUp && currentNodes < maxNodes {
		a.sugar.Infow("Scaling up cluster", "cluster", clusterID, "metric", rule.Metric, "value", currentValue)
		
		if err := a.clusterMgr.ScaleCluster(context.Background(), clusterID, currentNodes+rule.ScaleUpCount); err != nil {
			return fmt.Errorf("failed to scale up: %w", err)
		}
		rule.LastScaleUp = time.Now()
	}

	// Scale down
	if currentValue <= rule.LowerThreshold && canScaleDown && currentNodes > minNodes {
		a.sugar.Infow("Scaling down cluster", "cluster", clusterID, "metric", rule.Metric, "value", currentValue)
		
		newCount := currentNodes - rule.ScaleDownCount
		if newCount < minNodes {
			newCount = minNodes
		}
		
		if err := a.clusterMgr.ScaleCluster(context.Background(), clusterID, newCount); err != nil {
			return fmt.Errorf("failed to scale down: %w", err)
		}
		rule.LastScaleDown = time.Now()
	}

	return nil
}

// GetRule returns the scale rule for a cluster
func (a *AutoScaler) GetRule(clusterID string) *ScaleRule {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.rules[clusterID]
}
