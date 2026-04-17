package runner

import (
	"context"

	"github.com/stsgym/vimic2/internal/types"
)

// RunnerManagerAdapter wraps RunnerManager to implement types.RunnerManagerInterface.
type RunnerManagerAdapter struct {
	*RunnerManager
}

// NewRunnerManagerAdapter creates an adapter
func NewRunnerManagerAdapter(rm *RunnerManager) *RunnerManagerAdapter {
	return &RunnerManagerAdapter{RunnerManager: rm}
}

// CreateRunner creates a runner from platform and config
func (a *RunnerManagerAdapter) CreateRunner(platform types.RunnerPlatform, config map[string]interface{}) (string, error) {
	poolName := "default"
	if p, ok := config["pool_name"].(string); ok {
		poolName = p
	}
	pipelineID := ""
	if p, ok := config["pipeline_id"].(string); ok {
		pipelineID = p
	}
	var labels []string
	if l, ok := config["labels"].([]string); ok {
		labels = l
	}

	info, err := a.RunnerManager.CreateRunner(context.Background(), poolName, platform, pipelineID, labels)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

// DestroyRunner destroys a runner by ID
func (a *RunnerManagerAdapter) DestroyRunner(runnerID string) error {
	return a.RunnerManager.DestroyRunner(context.Background(), runnerID)
}

// GetRunner returns runner info as a map
func (a *RunnerManagerAdapter) GetRunner(runnerID string) (map[string]interface{}, error) {
	info, err := a.RunnerManager.GetRunner(runnerID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":         info.ID,
		"vm_id":      info.VMID,
		"pool_name":  info.PoolName,
		"platform":    string(info.Platform),
		"name":        info.Name,
		"status":      string(info.Status),
		"ip_address":  info.IPAddress,
		"created_at":  info.CreatedAt,
	}, nil
}

// Verify interface satisfaction
var _ types.RunnerManagerInterface = (*RunnerManagerAdapter)(nil)