// Package orchestrator provides rolling updates and health checks
package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"go.uber.org/zap"
)

// UpdateStrategy defines the strategy for rolling updates
type UpdateStrategy struct {
	BatchSize      int // Nodes to update simultaneously
	MaxUnavailable int // Max nodes that can be unavailable
	WaitBetween    int // Seconds to wait between batches
}

// IsValid validates the update strategy
func (s *UpdateStrategy) IsValid() bool {
	return s.BatchSize > 0 && s.MaxUnavailable >= 0 && s.WaitBetween >= 0
}

// CalculateBatches divides nodes into update batches
func (r *RollingUpdater) CalculateBatches(nodes []string, batchSize int) [][]string {
	if batchSize <= 0 || len(nodes) == 0 {
		return nil
	}
	
	var batches [][]string
	for i := 0; i < len(nodes); i += batchSize {
		end := i + batchSize
		if end > len(nodes) {
			end = len(nodes)
		}
		batches = append(batches, nodes[i:end])
	}
	return batches
}

// SetStrategy sets the update strategy
func (r *RollingUpdater) SetStrategy(strategy *UpdateStrategy) {
	r.strategy = strategy
}

// IsNodeHealthy checks if a node is healthy
func (r *RollingUpdater) IsNodeHealthy(ctx context.Context, nodeID string) (bool, error) {
	node, err := r.db.GetNode(nodeID)
	if err != nil {
		return false, err
	}
	if node == nil {
		return false, fmt.Errorf("node not found: %s", nodeID)
	}
	return node.State == "running", nil
}

// DrainNode cordons and drains a node
func (r *RollingUpdater) DrainNode(ctx context.Context, nodeID string) error {
	node, err := r.db.GetNode(nodeID)
	if err != nil {
		return err
	}
	// Mark as draining
	node.State = "draining"
	return r.db.SaveNode(node)
}

// UpgradeNode upgrades a node to a new version
func (r *RollingUpdater) UpgradeNode(ctx context.Context, nodeID string, version string) error {
	node, err := r.db.GetNode(nodeID)
	if err != nil {
		return err
	}
	node.Config.Version = version
	return r.db.SaveNode(node)
}

// RestoreNode restores a node to its previous state
func (r *RollingUpdater) RestoreNode(ctx context.Context, nodeID string) error {
	node, err := r.db.GetNode(nodeID)
	if err != nil {
		return err
	}
	node.State = "running"
	return r.db.SaveNode(node)
}

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"go.uber.org/zap"
)

// RollingUpdater performs zero-downtime rolling updates
type RollingUpdater struct {
	db       *database.DB
	nodes    map[string]*database.Node
	sugar    *zap.SugaredLogger
	ctx      context.Context
	cancel   context.CancelFunc
	strategy *UpdateStrategy
}

// UpdateConfig holds rolling update configuration
type UpdateConfig struct {
	BatchSize    int           // Nodes to update at once
	BatchPause   time.Duration // Pause between batches
	HealthCheck  bool          // Wait for health after each batch
	HealthTimeout time.Duration // Timeout for health check
	NewImage     string        // New image to use
}

// UpdateProgress tracks the progress of an update
type UpdateProgress struct {
	ClusterID     string
	TotalNodes    int
	UpdatedNodes  int
	CurrentNode   string
	Status        string // pending, updating, health-checking, complete, failed
	Error         error
	StartedAt     time.Time
	CompletedAt   time.Time
}

// NewRollingUpdater creates a new rolling updater
func NewRollingUpdater(db *database.DB, sugar *zap.SugaredLogger) *RollingUpdater {
	ctx, cancel := context.WithCancel(context.Background())
	return &RollingUpdater{
		db:     db,
		nodes:  make(map[string]*database.Node),
		sugar:  sugar,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Update performs a rolling update on a cluster
func (r *RollingUpdater) Update(clusterID string, config *UpdateConfig, progress chan<- *UpdateProgress) error {
	nodes, err := r.db.ListClusterNodes(clusterID)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes in cluster")
	}

	totalNodes := len(nodes)
	update := &UpdateProgress{
		ClusterID:    clusterID,
		TotalNodes:   totalNodes,
		Status:       "pending",
		StartedAt:    time.Now(),
	}

	r.sugar.Infow("Starting rolling update", "cluster", clusterID, "nodes", totalNodes)

	// Process in batches
	for i := 0; i < len(nodes); i += config.BatchSize {
		select {
		case <-r.ctx.Done():
			update.Status = "cancelled"
			progress <- update
			return fmt.Errorf("update cancelled")
		default:
		}

		batch := nodes[i:]
		if len(batch) > config.BatchSize {
			batch = batch[:config.BatchSize]
		}

		update.Status = "updating"
		update.CurrentNode = batch[0].Name
		progress <- update

		// Update batch
		for _, node := range batch {
			if err := r.updateNode(node, config); err != nil {
				update.Error = err
				update.Status = "failed"
				progress <- update
				return fmt.Errorf("failed to update node %s: %w", node.Name, err)
			}
			update.UpdatedNodes++
			progress <- update
		}

		// Health check after batch
		if config.HealthCheck && len(nodes) > config.BatchSize {
			update.Status = "health-checking"
			progress <- update

			if err := r.waitForHealthy(batch, config.HealthTimeout); err != nil {
				update.Error = err
				update.Status = "failed"
				progress <- update
				return fmt.Errorf("health check failed: %w", err)
			}
		}

		// Pause between batches
		if len(nodes) > config.BatchSize && i+config.BatchSize < len(nodes) {
			r.sugar.Info("Pausing between batches", "pause", config.BatchPause)
			time.Sleep(config.BatchPause)
		}
	}

	update.Status = "complete"
	update.CompletedAt = time.Now()
	progress <- update

	r.sugar.Info("Rolling update complete", "cluster", clusterID)
	return nil
}

func (r *RollingUpdater) updateNode(node *database.Node, config *UpdateConfig) error {
	r.sugar.Infow("Updating node", "node", node.Name, "image", config.NewImage)

	// Get the hypervisor for this node
	// Note: In production, we'd get this from a host manager
	// For now, we simulate the update by updating the node's image config
	if node.Config != nil {
		node.Config.Image = config.NewImage
	}

	// Update would involve:
	// 1. Stop node via hypervisor.StopNode(ctx, node.ID)
	// 2. Replace disk image or resize
	// 3. Start node via hypervisor.StartNode(ctx, node.ID)

	return nil
}

func (r *RollingUpdater) waitForHealthy(nodes []*database.Node, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for _, node := range nodes {
		for time.Now().Before(deadline) {
			select {
			case <-r.ctx.Done():
				return fmt.Errorf("cancelled")
			default:
			}

			// Check node health
			// 1. Check if node is running in database
			currentNode, err := r.db.GetNode(node.ID)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			// 2. Verify node is in running state
			if currentNode.State == "running" {
				// 3. Could add additional checks:
				// - SSH connectivity check
				// - HTTP health endpoint check
				// - Process/service checks
				break // Node is healthy, move to next
			}

			time.Sleep(2 * time.Second)
		}
	}

	return nil
}

// Cancel cancels the running update
func (r *RollingUpdater) Cancel() {
	r.cancel()
}

// HealthChecker performs health checks on nodes
type HealthChecker struct {
	db    *database.DB
	nodes map[string]*HealthStatus
	mu    sync.RWMutex
}

// HealthStatus holds the health status of a node
type HealthStatus struct {
	NodeID      string
	Healthy     bool
	LastCheck   time.Time
	Checks      int
	Failures    int
	Message     string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *database.DB) *HealthChecker {
	return &HealthChecker{
		db:    db,
		nodes: make(map[string]*HealthStatus),
	}
}

// CheckNode checks the health of a single node
func (h *HealthChecker) CheckNode(nodeID string) *HealthStatus {
	h.mu.Lock()
	defer h.mu.Unlock()

	status := &HealthStatus{
		NodeID:    nodeID,
		LastCheck: time.Now(),
		Checks:    1,
	}

	node, err := h.db.GetNode(nodeID)
	if err != nil || node == nil {
		status.Healthy = false
		status.Message = "Node not found"
		return status
	}

	// Check if node is running
	if node.State != "running" {
		status.Healthy = false
		status.Message = fmt.Sprintf("Node not running (state: %s)", node.State)
		return status
	}

	// TODO: Additional health checks
	// - SSH connectivity
	// - Response time
	// - Resource usage

	status.Healthy = true
	status.Message = "OK"

	// Update stored status
	if existing, ok := h.nodes[nodeID]; ok {
		status.Failures = existing.Failures
		if !status.Healthy {
			status.Failures++
		}
	}

	h.nodes[nodeID] = status
	return status
}

// GetStatus returns the health status of a node
func (h *HealthChecker) GetStatus(nodeID string) *HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.nodes[nodeID]
}

// GetAllStatus returns health status for all known nodes
func (h *HealthChecker) GetAllStatus() []*HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var list []*HealthStatus
	for _, status := range h.nodes {
		list = append(list, status)
	}
	return list
}
