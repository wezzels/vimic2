// Package orchestrator provides disaster recovery
package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/stsgym/vimic2/internal/database"
)

// RecoveryManager handles cluster backup and restore
type RecoveryManager struct {
	db        *database.DB
	backupDir string
	sugar     *zap.SugaredLogger
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManagerWithLogger(db *database.DB, backupDir string, sugar *zap.SugaredLogger) *RecoveryManager {
	if backupDir == "" {
		backupDir = "/var/lib/vimic2/backups"
	}
	return &RecoveryManager{
		db:        db,
		backupDir: backupDir,
		sugar:     sugar,
	}
}

// NewRecoveryManager creates a new recovery manager with default logger
func NewRecoveryManager(db *database.DB, backupDir string) *RecoveryManager {
	logger, _ := zap.NewProduction()
	return NewRecoveryManagerWithLogger(db, backupDir, logger.Sugar())
}

// Backup holds a cluster backup
type Backup struct {
	ID          string        `json:"id"`
	ClusterID   string        `json:"cluster_id"`
	ClusterName string        `json:"cluster_name"`
	Nodes       []*BackupNode `json:"nodes"`
	CreatedAt   time.Time     `json:"created_at"`
	Notes       string        `json:"notes"`
}

// BackupNode holds node backup info
type BackupNode struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Role   string                 `json:"role"`
	HostID string                 `json:"host_id"`
	Config map[string]interface{} `json:"config"`
	IP     string                 `json:"ip"`
}

// CPU returns the CPU count from config
func (n *BackupNode) CPU() int {
	if n.Config == nil {
		return 0
	}
	if v, ok := n.Config["cpu"]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}

// MemoryMB returns the memory in MB from config
func (n *BackupNode) MemoryMB() int {
	if n.Config == nil {
		return 0
	}
	if v, ok := n.Config["memory_mb"]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}

// DiskGB returns the disk size in GB from config
func (n *BackupNode) DiskGB() int {
	if n.Config == nil {
		return 0
	}
	if v, ok := n.Config["disk_gb"]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}

// CreateBackup creates a backup of a cluster
func (r *RecoveryManager) CreateBackup(ctx context.Context, clusterID, notes string) (*Backup, error) {
	cluster, err := r.db.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster not found")
	}

	nodes, err := r.db.ListClusterNodes(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	backup := &Backup{
		ID:          fmt.Sprintf("backup-%d", time.Now().UnixNano()),
		ClusterID:   clusterID,
		ClusterName: cluster.Name,
		Nodes:       make([]*BackupNode, 0, len(nodes)),
		CreatedAt:   time.Now(),
		Notes:       notes,
	}

	for _, node := range nodes {
		config := make(map[string]interface{})
		if node.Config != nil {
			config["cpu"] = node.Config.CPU
			config["memory_mb"] = node.Config.MemoryMB
			config["disk_gb"] = node.Config.DiskGB
			config["image"] = node.Config.Image
		}

		backup.Nodes = append(backup.Nodes, &BackupNode{
			ID:     node.ID,
			Name:   node.Name,
			Role:   node.Role,
			HostID: node.HostID,
			Config: config,
			IP:     node.IP,
		})
	}

	// Save backup to file
	if err := r.saveBackup(backup); err != nil {
		return nil, fmt.Errorf("failed to save backup: %w", err)
	}

	return backup, nil
}

// saveBackup saves backup to disk
func (r *RecoveryManager) saveBackup(backup *Backup) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(r.backupDir, 0755); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s.json", r.backupDir, backup.ID)
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ListBackups lists all backups for a cluster
func (r *RecoveryManager) ListBackups(clusterID string) ([]*Backup, error) {
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		return nil, err
	}

	var backups []*Backup
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) < 10 {
			continue
		}

		data, err := os.ReadFile(r.backupDir + "/" + entry.Name())
		if err != nil {
			continue
		}

		var backup Backup
		if err := json.Unmarshal(data, &backup); err != nil {
			continue
		}

		if clusterID == "" || backup.ClusterID == clusterID {
			backups = append(backups, &backup)
		}
	}

	return backups, nil
}

// GetBackup loads a backup
func (r *RecoveryManager) GetBackup(backupID string) (*Backup, error) {
	path := fmt.Sprintf("%s/%s.json", r.backupDir, backupID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, err
	}

	return &backup, nil
}

// RestoreCluster restores a cluster from backup
func (r *RecoveryManager) RestoreCluster(ctx context.Context, backupID string) error {
	backup, err := r.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	// Check if cluster exists
	existingCluster, _ := r.db.GetCluster(backup.ClusterID)
	if existingCluster == nil {
		// Recreate cluster if it doesn't exist
		existingCluster = &database.Cluster{
			ID:     backup.ClusterID,
			Name:   backup.ClusterName,
			Status: "restoring",
		}
		if err := r.db.SaveCluster(existingCluster); err != nil {
			return fmt.Errorf("failed to recreate cluster: %w", err)
		}
	}

	// Restore nodes from backup
	for _, backupNode := range backup.Nodes {
		// Check if node exists
		existingNode, _ := r.db.GetNode(backupNode.ID)

		if existingNode == nil {
			// Recreate node from backup config
			config := &database.NodeConfig{}
			if v, ok := backupNode.Config["cpu"]; ok {
				config.CPU = int(v.(float64))
			}
			if v, ok := backupNode.Config["memory_mb"]; ok {
				config.MemoryMB = uint64(v.(float64))
			}
			if v, ok := backupNode.Config["disk_gb"]; ok {
				config.DiskGB = int(v.(float64))
			}
			if v, ok := backupNode.Config["image"]; ok {
				config.Image = v.(string)
			}

			node := &database.Node{
				ID:        backupNode.ID,
				ClusterID: backup.ClusterID,
				Name:      backupNode.Name,
				Role:      backupNode.Role,
				HostID:    backupNode.HostID,
				IP:        backupNode.IP,
				State:     "pending",
				Config:    config,
			}
			if err := r.db.SaveNode(node); err != nil {
				r.sugar.Warnw("Failed to restore node", "node", backupNode.Name, "error", err)
			}
		} else {
			// Update existing node state
			if err := r.db.UpdateNodeState(backupNode.ID, "pending", backupNode.IP); err != nil {
				r.sugar.Warnw("Failed to update node", "node", backupNode.Name, "error", err)
			}
		}
	}

	// Update cluster status
	r.db.UpdateClusterStatus(backup.ClusterID, "running")

	return nil
}

// DeleteBackup deletes a backup
func (r *RecoveryManager) DeleteBackup(backupID string) error {
	path := fmt.Sprintf("%s/%s.json", r.backupDir, backupID)
	return os.Remove(path)
}

// Failover handles failover to standby nodes
func (r *RecoveryManager) Failover(ctx context.Context, clusterID, failedNodeID string) error {
	// Get the failed node
	failedNode, err := r.db.GetNode(failedNodeID)
	if err != nil {
		return fmt.Errorf("failed node not found: %w", err)
	}

	// Mark failed node as error state
	if err := r.db.UpdateNodeState(failedNodeID, "error", failedNode.IP); err != nil {
		r.sugar.Warnw("Failed to mark node as error", "node", failedNodeID, "error", err)
	}

	// Find available standby nodes in the cluster
	nodes, err := r.db.ListClusterNodes(clusterID)
	if err != nil {
		return fmt.Errorf("failed to list cluster nodes: %w", err)
	}

	// Look for a standby node to promote
	for _, node := range nodes {
		if node.Role == "standby" && node.State == "running" {
			// Promote standby to worker
			node.Role = "worker"
			if err := r.db.SaveNode(node); err != nil {
				r.sugar.Warnw("Failed to promote standby", "node", node.ID, "error", err)
				continue
			}
			r.sugar.Infow("Promoted standby node", "standby", node.ID, "replacing", failedNodeID)
			return nil
		}
	}

	// No standby available - log warning
	r.sugar.Warnw("No standby nodes available for failover", "cluster", clusterID, "failed", failedNodeID)

	return fmt.Errorf("no standby nodes available")
}
