// Package runner provides Drone runner orchestration
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

// DroneRunner manages Drone CI runners
type DroneRunner struct {
	db          *types.PipelineDB
	droneURL    string
	rpcHost     string
	rpcSecret   string
	runnerUser  string
	stateFile   string
	runners     map[string]*DroneRunnerInfo
	mu          sync.RWMutex
}

// DroneRunnerInfo represents a Drone runner
type DroneRunnerInfo struct {
	ID            string       `json:"id"`
	VMID          string       `json:"vm_id"`
	PipelineID    string       `json:"pipeline_id"`
	Token         string       `json:"token"`
	IPAddress     string       `json:"ip_address"`
	Status        string       `json:"status"`
	WorkDir       string        `json:"work_dir"`
	CreatedAt     time.Time    `json:"created_at"`
	LastHeartbeat *time.Time   `json:"last_heartbeat,omitempty"`
	DestroyedAt    *time.Time  `json:"destroyed_at,omitempty"`
}

// DroneConfig represents Drone runner configuration
type DroneConfig struct {
	URL       string `json:"url"`
	RPCHost   string `json:"rpc_host"`
	RPCSecret string `json:"rpc_secret"`
	RunnerUser string `json:"runner_user"`
}

// NewDroneRunner creates a new Drone runner manager
func NewDroneRunner(db *types.PipelineDB, config *DroneConfig) (*DroneRunner, error) {
	dr := &DroneRunner{
		db:         db,
		droneURL:   config.URL,
		rpcHost:    config.RPCHost,
		rpcSecret:  config.RPCSecret,
		runnerUser: config.RunnerUser,
		runners:    make(map[string]*DroneRunnerInfo),
	}

	if dr.runnerUser == "" {
		dr.runnerUser = "drone"
	}

	if err := dr.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return dr, nil
}

func (dr *DroneRunner) loadState() error {
	stateFile := dr.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var runners []*DroneRunnerInfo
	if err := json.Unmarshal(data, &runners); err != nil {
		return err
	}

	for _, runner := range runners {
		dr.runners[runner.ID] = runner
	}

	return nil
}

func (dr *DroneRunner) saveState() error {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	runners := make([]*DroneRunnerInfo, 0, len(dr.runners))
	for _, runner := range dr.runners {
		runners = append(runners, runner)
	}

	data, err := json.MarshalIndent(runners, "", "  ")
	if err != nil {
		return err
	}

	stateFile := dr.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

func (dr *DroneRunner) getStateFile() string {
	if dr.stateFile != "" {
		return dr.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "drone-runners.json")
}

func (dr *DroneRunner) SetStateFile(path string) {
	dr.stateFile = path
}

func (dr *DroneRunner) RegisterRunner(ctx context.Context, vmID, pipelineID string) (*DroneRunnerInfo, error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	token := fmt.Sprintf("%s-%s", dr.rpcSecret, generateRunnerID("drone"))

	runner := &DroneRunnerInfo{
		ID:         generateRunnerID("drone"),
		VMID:       vmID,
		PipelineID: pipelineID,
		Token:      token,
		Status:     "registered",
		WorkDir:    "/work",
		CreatedAt:  time.Now(),
	}

	dr.runners[runner.ID] = runner

	dbRunner := &pipeline.Runner{
		ID:         runner.ID,
		PipelineID: pipelineID,
		VMID:       vmID,
		Platform:   pipeline.PlatformDrone,
		PlatformID: token[:8],
		Token:      token,
		Name:       fmt.Sprintf("drone-runner-%s", vmID[:8]),
		Status:     types.RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}

	if err := dr.db.SaveRunner(ctx, dbRunner); err != nil {
		delete(dr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save runner: %w", err)
	}

	if err := dr.saveState(); err != nil {
		dr.db.DeleteRunner(ctx, runner.ID)
		delete(dr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return runner, nil
}

func (dr *DroneRunner) UnregisterRunner(ctx context.Context, runnerID string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	runner, ok := dr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	now := time.Time{}
	runner.DestroyedAt = &now

	dr.db.DeleteRunner(ctx, runnerID)
	delete(dr.runners, runnerID)

	return dr.saveState()
}

func (dr *DroneRunner) StartRunner(ctx context.Context, runnerID, vmIP string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	runner, ok := dr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", dr.runnerUser, vmIP),
		fmt.Sprintf("docker run -d -v /var/run/docker.sock:/var/run/docker.sock "+
			"-e DRONE_RPC_HOST=%s -e DRONE_RPC_SECRET=%s -e DRONE_RUNNER_CAPACITY=2 "+
			"drone/drone-runner-docker:1",
			dr.rpcHost, runner.Token),
	)

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start runner: %w: %s", err, output)
	}

	runner.IPAddress = vmIP
	runner.Status = "running"
	now := time.Time{}
	runner.LastHeartbeat = &now

	return dr.saveState()
}

func (dr *DroneRunner) StopRunner(ctx context.Context, runnerID string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	runner, ok := dr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	if runner.IPAddress == "" {
		return fmt.Errorf("runner not started")
	}

	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", dr.runnerUser, runner.IPAddress),
		"docker stop $(docker ps -q -f ancestor=drone/drone-runner-docker:1) || true",
	)

	sshCmd.Run()
	runner.Status = "stopped"
	return dr.saveState()
}

func (dr *DroneRunner) GetRunner(runnerID string) (*DroneRunnerInfo, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	runner, ok := dr.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}
	return runner, nil
}

func (dr *DroneRunner) ListRunners() []*DroneRunnerInfo {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	runners := make([]*DroneRunnerInfo, 0, len(dr.runners))
	for _, runner := range dr.runners {
		runners = append(runners, runner)
	}
	return runners
}

func (dr *DroneRunner) ListRunnersByPipeline(pipelineID string) []*DroneRunnerInfo {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	runners := make([]*DroneRunnerInfo, 0)
	for _, runner := range dr.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

func (dr *DroneRunner) UpdateHeartbeat(runnerID string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	runner, ok := dr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	now := time.Time{}
	runner.LastHeartbeat = &now
	return dr.saveState()
}

func (dr *DroneRunner) GetStats() map[string]int {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	stats := map[string]int{"total": len(dr.runners)}
	for _, runner := range dr.runners {
		stats[runner.Status]++
	}
	return stats
}