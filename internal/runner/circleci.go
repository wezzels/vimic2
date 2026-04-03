// Package runner provides CircleCI runner orchestration
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// CircleCIRunner manages CircleCI runners
type CircleCIRunner struct {
	db           *types.PipelineDB
	circleciURL  string
	apiToken     string
	resourceClass string
	runnerUser   string
	stateFile    string
	runners      map[string]*CircleCIRunnerInfo
	mu           sync.RWMutex
}

// CircleCIRunnerInfo represents a CircleCI runner
type CircleCIRunnerInfo struct {
	ID            string       `json:"id"`
	VMID          string       `json:"vm_id"`
	PipelineID    string       `json:"pipeline_id"`
	ResourceClass string       `json:"resource_class"`
	Token         string       `json:"token"`
	IPAddress     string       `json:"ip_address"`
	Status         string       `json:"status"`
	WorkDir       string        `json:"work_dir"`
	CreatedAt     time.Time    `json:"created_at"`
	LastHeartbeat *time.Time   `json:"last_heartbeat,omitempty"`
	DestroyedAt    *time.Time  `json:"destroyed_at,omitempty"`
}

// CircleCIConfig represents CircleCI runner configuration
type CircleCIConfig struct {
	URL           string `json:"url"`
	Token         string `json:"token"`
	ResourceClass string `json:"resource_class"`
	RunnerUser    string `json:"runner_user"`
}

// NewCircleCIRunner creates a new CircleCI runner manager
func NewCircleCIRunner(db *types.PipelineDB, config *CircleCIConfig) (*CircleCIRunner, error) {
	cr := &CircleCIRunner{
		db:            db,
		circleciURL:   config.URL,
		apiToken:      config.Token,
		resourceClass: config.ResourceClass,
		runnerUser:    config.RunnerUser,
		runners:       make(map[string]*CircleCIRunnerInfo),
	}

	if cr.circleciURL == "" {
		cr.circleciURL = "https://circleci.com"
	}

	if cr.runnerUser == "" {
		cr.runnerUser = "circleci"
	}

	if err := cr.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return cr, nil
}

func (cr *CircleCIRunner) loadState() error {
	stateFile := cr.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var runners []*CircleCIRunnerInfo
	if err := json.Unmarshal(data, &runners); err != nil {
		return err
	}

	for _, runner := range runners {
		cr.runners[runner.ID] = runner
	}

	return nil
}

func (cr *CircleCIRunner) saveState() error {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	runners := make([]*CircleCIRunnerInfo, 0, len(cr.runners))
	for _, runner := range cr.runners {
		runners = append(runners, runner)
	}

	data, err := json.MarshalIndent(runners, "", "  ")
	if err != nil {
		return err
	}

	stateFile := cr.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

func (cr *CircleCIRunner) getStateFile() string {
	if cr.stateFile != "" {
		return cr.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "circleci-runners.json")
}

func (cr *CircleCIRunner) SetStateFile(path string) {
	cr.stateFile = path
}

func (cr *CircleCIRunner) RegisterRunner(ctx context.Context, vmID, pipelineID string) (*CircleCIRunnerInfo, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	resourceClass := fmt.Sprintf("%s-%s", cr.resourceClass, vmID[:8])

	// Create resource class in CircleCI
	token, err := cr.createResourceClass(ctx, resourceClass)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource class: %w", err)
	}

	runner := &CircleCIRunnerInfo{
		ID:            generateRunnerID("cci"),
		VMID:          vmID,
		PipelineID:    pipelineID,
		ResourceClass: resourceClass,
		Token:         token,
		Status:        "registered",
		WorkDir:       "/work",
		CreatedAt:     time.Now(),
	}

	cr.runners[runner.ID] = runner

	dbRunner := &pipeline.Runner{
		ID:         runner.ID,
		PipelineID: pipelineID,
		VMID:       vmID,
		Platform:   pipeline.PlatformCircleCI,
		PlatformID: resourceClass,
		Token:      token,
		Name:       resourceClass,
		Status:     types.RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}

	if err := cr.db.SaveRunner(ctx, dbRunner); err != nil {
		delete(cr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save runner: %w", err)
	}

	if err := cr.saveState(); err != nil {
		cr.db.DeleteRunner(ctx, runner.ID)
		delete(cr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return runner, nil
}

func (cr *CircleCIRunner) createResourceClass(ctx context.Context, resourceClass string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("%s/api/v2/runner/resource-class", cr.circleciURL)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	req.Header.Set("Circle-Token", cr.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create resource class: %s: %s", resp.Status, body)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func (cr *CircleCIRunner) UnregisterRunner(ctx context.Context, runnerID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	runner, ok := cr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	cr.deleteResourceClass(ctx, runner.ResourceClass)

	now := time.Now()
	runner.DestroyedAt = &now

	cr.db.DeleteRunner(ctx, runnerID)
	delete(cr.runners, runnerID)

	return cr.saveState()
}

func (cr *CircleCIRunner) deleteResourceClass(ctx context.Context, resourceClass string) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/api/v2/runner/resource-class/%s", cr.circleciURL, resourceClass)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	req.Header.Set("Circle-Token", cr.apiToken)

	client.Do(req) // Ignore error
}

func (cr *CircleCIRunner) StartRunner(ctx context.Context, runnerID, vmIP string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	runner, ok := cr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Download CircleCI runner
	downloadURL := fmt.Sprintf("%s/api/v2/runner/download", cr.circleciURL)

	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", cr.runnerUser, vmIP),
		fmt.Sprintf("curl -o /tmp/circleci-runner '%s' && chmod +x /tmp/circleci-runner && "+
			"CIRCLECI_RESOURCE_CLASS=%s CIRCLECI_RUNNER_TOKEN=%s nohup /tmp/circleci-runner &",
			downloadURL, runner.ResourceClass, runner.Token),
	)

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start runner: %w: %s", err, output)
	}

	runner.IPAddress = vmIP
	runner.Status = "running"
	now := time.Now()
	runner.LastHeartbeat = &now

	return cr.saveState()
}

func (cr *CircleCIRunner) StopRunner(ctx context.Context, runnerID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	runner, ok := cr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	if runner.IPAddress == "" {
		return fmt.Errorf("runner not started")
	}

	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", cr.runnerUser, runner.IPAddress),
		"pkill -f circleci-runner || true",
	)

	sshCmd.Run()
	runner.Status = "stopped"
	return cr.saveState()
}

func (cr *CircleCIRunner) GetRunner(runnerID string) (*CircleCIRunnerInfo, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	runner, ok := cr.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}
	return runner, nil
}

func (cr *CircleCIRunner) ListRunners() []*CircleCIRunnerInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	runners := make([]*CircleCIRunnerInfo, 0, len(cr.runners))
	for _, runner := range cr.runners {
		runners = append(runners, runner)
	}
	return runners
}

func (cr *CircleCIRunner) ListRunnersByPipeline(pipelineID string) []*CircleCIRunnerInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	runners := make([]*CircleCIRunnerInfo, 0)
	for _, runner := range cr.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

func (cr *CircleCIRunner) UpdateHeartbeat(runnerID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	runner, ok := cr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	now := time.Time{}
	runner.LastHeartbeat = &now
	return cr.saveState()
}

func (cr *CircleCIRunner) GetStats() map[string]int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	stats := map[string]int{"total": len(cr.runners)}
	for _, runner := range cr.runners {
		stats[runner.Status]++
	}
	return stats
}