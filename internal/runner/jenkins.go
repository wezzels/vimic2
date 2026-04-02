// Package runner provides Jenkins runner orchestration
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

	"github.com/stsgym/vimic2/internal/pipeline"
)

// JenkinsRunner manages Jenkins agents
type JenkinsRunner struct {
	db          *pipeline.PipelineDB
	jenkinsURL  string
	username    string
	apiToken    string
	runnerUser  string
	stateFile   string
	runners     map[string]*JenkinsRunnerInfo
	mu          sync.RWMutex
}

// JenkinsRunnerInfo represents a Jenkins agent
type JenkinsRunnerInfo struct {
	ID            string       `json:"id"`
	VMID          string       `json:"vm_id"`
	PipelineID    string       `json:"pipeline_id"`
	AgentName     string       `json:"agent_name"`
	Secret        string       `json:"secret"`
	IPAddress     string       `json:"ip_address"`
	Status        string       `json:"status"`
	WorkDir       string        `json:"work_dir"`
	CreatedAt     time.Time    `json:"created_at"`
	LastHeartbeat *time.Time   `json:"last_heartbeat,omitempty"`
	DestroyedAt    *time.Time  `json:"destroyed_at,omitempty"`
}

// JenkinsConfig represents Jenkins runner configuration
type JenkinsConfig struct {
	URL       string `json:"url"`
	Username  string `json:"username"`
	APIToken  string `json:"api_token"`
	RunnerUser string `json:"runner_user"`
}

// NewJenkinsRunner creates a new Jenkins runner manager
func NewJenkinsRunner(db *pipeline.PipelineDB, config *JenkinsConfig) (*JenkinsRunner, error) {
	jr := &JenkinsRunner{
		db:         db,
		jenkinsURL: config.URL,
		username:   config.Username,
		apiToken:   config.APIToken,
		runnerUser: config.RunnerUser,
		runners:    make(map[string]*JenkinsRunnerInfo),
	}

	if jr.runnerUser == "" {
		jr.runnerUser = "jenkins"
	}

	// Load state
	if err := jr.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return jr, nil
}

func (jr *JenkinsRunner) loadState() error {
	stateFile := jr.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var runners []*JenkinsRunnerInfo
	if err := json.Unmarshal(data, &runners); err != nil {
		return err
	}

	for _, runner := range runners {
		jr.runners[runner.ID] = runner
	}

	return nil
}

func (jr *JenkinsRunner) saveState() error {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	runners := make([]*JenkinsRunnerInfo, 0, len(jr.runners))
	for _, runner := range jr.runners {
		runners = append(runners, runner)
	}

	data, err := json.MarshalIndent(runners, "", "  ")
	if err != nil {
		return err
	}

	stateFile := jr.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

func (jr *JenkinsRunner) getStateFile() string {
	if jr.stateFile != "" {
		return jr.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "jenkins-runners.json")
}

func (jr *JenkinsRunner) SetStateFile(path string) {
	jr.stateFile = path
}

func (jr *JenkinsRunner) RegisterRunner(ctx context.Context, vmID, pipelineID string) (*JenkinsRunnerInfo, error) {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	agentName := fmt.Sprintf("vimic2-agent-%s", vmID[:8])

	// Create agent in Jenkins
	secret, err := jr.createAgentInJenkins(ctx, agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent in Jenkins: %w", err)
	}

	runner := &JenkinsRunnerInfo{
		ID:         generateRunnerID("jenkins"),
		VMID:       vmID,
		PipelineID: pipelineID,
		AgentName:  agentName,
		Secret:     secret,
		Status:     "registered",
		WorkDir:    "/work",
		CreatedAt:  time.Now(),
	}

	jr.runners[runner.ID] = runner

	// Save to database
	dbRunner := &pipeline.Runner{
		ID:         runner.ID,
		PipelineID: pipelineID,
		VMID:       vmID,
		Platform:   pipeline.PlatformJenkins,
		PlatformID: agentName,
		Token:      secret,
		Name:       agentName,
		Status:     pipeline.RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}
	if err := jr.db.SaveRunner(ctx, dbRunner); err != nil {
		delete(jr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save runner: %w", err)
	}

	if err := jr.saveState(); err != nil {
		jr.db.DeleteRunner(ctx, runner.ID)
		delete(jr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return runner, nil
}

func (jr *JenkinsRunner) createAgentInJenkins(ctx context.Context, agentName string) (string, error) {
	// Download agent JAR
	jarURL := fmt.Sprintf("%s/jnlpJars/agent.jar", jr.jenkinsURL)
	jarPath := filepath.Join(os.TempDir(), "jenkins-agent.jar")

	cmd := exec.CommandContext(ctx, "curl", "-o", jarPath, jarURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to download agent JAR: %w: %s", err, output)
	}

	// Create agent via Jenkins CLI (simplified)
	// In production, would use Jenkins API or CLI
	return fmt.Sprintf("%s-secret", agentName), nil
}

func (jr *JenkinsRunner) UnregisterRunner(ctx context.Context, runnerID string) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	runner, ok := jr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Delete agent from Jenkins
	jr.deleteAgentFromJenkins(ctx, runner.AgentName)

	now := time.Now()
	runner.DestroyedAt = &now

	jr.db.DeleteRunner(ctx, runnerID)
	delete(jr.runners, runnerID)

	return jr.saveState()
}

func (jr *JenkinsRunner) deleteAgentFromJenkins(ctx context.Context, agentName string) {
	// Delete agent via Jenkins API
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/computer/%s/doDelete", jr.jenkinsURL, agentName)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	req.SetBasicAuth(jr.username, jr.apiToken)

	client.Do(req) // Ignore error
}

func (jr *JenkinsRunner) StartRunner(ctx context.Context, runnerID, vmIP string) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	runner, ok := jr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Download agent JAR
	jarURL := fmt.Sprintf("%s/jnlpJars/agent.jar", jr.jenkinsURL)

	// Start agent via SSH
	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", jr.runnerUser, vmIP),
		fmt.Sprintf("curl -o /tmp/agent.jar %s && java -jar /tmp/agent.jar -jnlpUrl %s/computer/%s/slave-agent.jnlp -secret %s &",
			jarURL, jr.jenkinsURL, runner.AgentName, runner.Secret),
	)

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start agent: %w: %s", err, output)
	}

	runner.IPAddress = vmIP
	runner.Status = "running"
	now := time.Now()
	runner.LastHeartbeat = &now

	return jr.saveState()
}

func (jr *JenkinsRunner) StopRunner(ctx context.Context, runnerID string) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	runner, ok := jr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	if runner.IPAddress == "" {
		return fmt.Errorf("runner not started")
	}

	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", jr.runnerUser, runner.IPAddress),
		"pkill -f 'agent.jar' || true",
	)

	sshCmd.Run() // Ignore error

	runner.Status = "stopped"
	return jr.saveState()
}

func (jr *JenkinsRunner) GetRunner(runnerID string) (*JenkinsRunnerInfo, error) {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	runner, ok := jr.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}
	return runner, nil
}

func (jr *JenkinsRunner) ListRunners() []*JenkinsRunnerInfo {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	runners := make([]*JenkinsRunnerInfo, 0, len(jr.runners))
	for _, runner := range jr.runners {
		runners = append(runners, runner)
	}
	return runners
}

func (jr *JenkinsRunner) ListRunnersByPipeline(pipelineID string) []*JenkinsRunnerInfo {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	runners := make([]*JenkinsRunnerInfo, 0)
	for _, runner := range jr.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

func (jr *JenkinsRunner) UpdateHeartbeat(runnerID string) error {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	runner, ok := jr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	now := time.Now()
	runner.LastHeartbeat = &now
	return jr.saveState()
}

func (jr *JenkinsRunner) GetStats() map[string]int {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	stats := map[string]int{"total": len(jr.runners)}
	for _, runner := range jr.runners {
		stats[runner.Status]++
	}
	return stats
}