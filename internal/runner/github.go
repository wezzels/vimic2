// Package runner provides CI/CD runner orchestration
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
	"strings"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/pipeline"
)

// GitHubRunner manages GitHub Actions runners
type GitHubRunner struct {
	db           *pipeline.PipelineDB
	githubURL    string
	personalToken string
	runnerName   string
	labels       []string
	runnerUser   string
	stateFile    string
	runners      map[string]*GitHubRunnerInfo
	mu           sync.RWMutex
}

// GitHubRunnerInfo represents a registered GitHub runner
type GitHubRunnerInfo struct {
	ID             string       `json:"id"`
	VMID           string       `json:"vm_id"`
	PipelineID     string       `json:"pipeline_id"`
	Token          string       `json:"token"`
	Name           string       `json:"name"`
	Labels         []string     `json:"labels"`
	RunnerID       int64        `json:"runner_id"`
	Repository     string       `json:"repository"`
	Organization   string       `json:"organization"`
	IPAddress      string       `json:"ip_address"`
	Status         string       `json:"status"`
	WorkDir        string       `json:"work_dir"`
	CreatedAt      time.Time    `json:"created_at"`
	LastHeartbeat  *time.Time   `json:"last_heartbeat,omitempty"`
	DestroyedAt     *time.Time   `json:"destroyed_at,omitempty"`
}

// GitHubConfig represents GitHub runner configuration
type GitHubConfig struct {
	URL            string   `json:"url"`
	Token          string   `json:"token"`
	Labels         []string `json:"labels"`
	Username       string   `json:"username"`
	Repository     string   `json:"repository"`
	Organization   string   `json:"organization"`
}

// NewGitHubRunner creates a new GitHub runner manager
func NewGitHubRunner(db *pipeline.PipelineDB, config *GitHubConfig) (*GitHubRunner, error) {
	gr := &GitHubRunner{
		db:            db,
		githubURL:     config.URL,
		personalToken: config.Token,
		runnerName:    "vimic2-runner",
		labels:        config.Labels,
		runnerUser:    config.Username,
		runners:       make(map[string]*GitHubRunnerInfo),
	}

	if gr.githubURL == "" {
		gr.githubURL = "https://github.com"
	}

	if gr.runnerUser == "" {
		gr.runnerUser = "runner"
	}

	// Load state
	if err := gr.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return gr, nil
}

// loadState loads runner state from disk
func (gr *GitHubRunner) loadState() error {
	stateFile := gr.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var runners []*GitHubRunnerInfo
	if err := json.Unmarshal(data, &runners); err != nil {
		return err
	}

	for _, runner := range runners {
		gr.runners[runner.ID] = runner
	}

	return nil
}

// saveState saves runner state to disk
func (gr *GitHubRunner) saveState() error {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitHubRunnerInfo, 0, len(gr.runners))
	for _, runner := range gr.runners {
		runners = append(runners, runner)
	}

	data, err := json.MarshalIndent(runners, "", "  ")
	if err != nil {
		return err
	}

	stateFile := gr.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (gr *GitHubRunner) getStateFile() string {
	if gr.stateFile != "" {
		return gr.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "github-runners.json")
}

// SetStateFile sets the state file path
func (gr *GitHubRunner) SetStateFile(path string) {
	gr.stateFile = path
}

// RegisterRunner registers a new GitHub Actions runner
func (gr *GitHubRunner) RegisterRunner(ctx context.Context, vmID, pipelineID string) (*GitHubRunnerInfo, error) {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	// Generate runner name
	runnerName := fmt.Sprintf("%s-%s", gr.runnerName, vmID[:8])

	// Get registration token from GitHub
	registrationToken, err := gr.getRegistrationToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get registration token: %w", err)
	}

	// Download runner (if not already downloaded)
	runnerDir, err := gr.downloadRunner(ctx, vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to download runner: %w", err)
	}

	// Configure runner
	token, runnerID, err := gr.configureRunner(ctx, runnerDir, runnerName, registrationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to configure runner: %w", err)
	}

	// Create runner info
	runner := &GitHubRunnerInfo{
		ID:           generateRunnerID("gh"),
		VMID:         vmID,
		PipelineID:   pipelineID,
		Token:        token,
		Name:         runnerName,
		Labels:       gr.labels,
		RunnerID:     runnerID,
		WorkDir:      "/work",
		Status:       "registered",
		CreatedAt:    time.Now(),
	}

	gr.runners[runner.ID] = runner

	// Save to database
	dbRunner := &pipeline.Runner{
		ID:         runner.ID,
		PipelineID: pipelineID,
		VMID:       vmID,
		Platform:   pipeline.PlatformGitHub,
		PlatformID: fmt.Sprintf("%d", runnerID),
		Token:      token,
		Labels:     gr.labels,
		Name:       runnerName,
		Status:     pipeline.RunnerStatusCreating,
		CreatedAt:  time.Now(),
	}
	if err := gr.db.SaveRunner(ctx, dbRunner); err != nil {
		delete(gr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save runner: %w", err)
	}

	// Save state
	if err := gr.saveState(); err != nil {
		gr.db.DeleteRunner(ctx, runner.ID)
		delete(gr.runners, runner.ID)
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return runner, nil
}

// getRegistrationToken gets a registration token from GitHub API
func (gr *GitHubRunner) getRegistrationToken(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var url string
	if gr.runners[gr.runnerName] != nil && gr.runners[gr.runnerName].Organization != "" {
		// Organization runner
		url = fmt.Sprintf("%s/api/v3/orgs/%s/actions/runners/registration-token", gr.githubURL, gr.runners[gr.runnerName].Organization)
	} else if gr.runners[gr.runnerName] != nil && gr.runners[gr.runnerName].Repository != "" {
		// Repository runner
		url = fmt.Sprintf("%s/api/v3/repos/%s/actions/runners/registration-token", gr.githubURL, gr.runners[gr.runnerName].Repository)
	} else {
		// Default to repository runner
		url = fmt.Sprintf("%s/api/v3/repos/%s/actions/runners/registration-token", gr.githubURL, "owner/repo")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", gr.personalToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get registration token: %s: %s", resp.Status, body)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

// downloadRunner downloads the GitHub Actions runner
func (gr *GitHubRunner) downloadRunner(ctx context.Context, vmID string) (string, error) {
	runnerDir := filepath.Join(os.Getenv("HOME"), ".vimic2", "runners", vmID)

	// Check if already downloaded
	if _, err := os.Stat(filepath.Join(runnerDir, "config.sh")); err == nil {
		return runnerDir, nil
	}

	// Create directory
	if err := os.MkdirAll(runnerDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create runner directory: %w", err)
	}

	// Determine latest runner version
	version, err := gr.getLatestRunnerVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get runner version: %w", err)
	}

	// Download URL
	downloadURL := fmt.Sprintf("https://github.com/actions/runner/releases/download/%s/actions-runner-linux-x64-%s.tar.gz", version, version)
	archivePath := filepath.Join(runnerDir, "runner.tar.gz")

	// Download runner
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", archivePath, downloadURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to download runner: %w: %s", err, output)
	}

	// Extract runner
	cmd = exec.CommandContext(ctx, "tar", "xzf", archivePath, "-C", runnerDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to extract runner: %w: %s", err, output)
	}

	// Clean up archive
	os.Remove(archivePath)

	return runnerDir, nil
}

// getLatestRunnerVersion gets the latest GitHub Actions runner version
func (gr *GitHubRunner) getLatestRunnerVersion(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://api.github.com/repos/actions/runner/releases/latest"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get releases: %s", resp.Status)
	}

	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Remove 'v' prefix if present
	return strings.TrimPrefix(result.TagName, "v"), nil
}

// configureRunner configures the GitHub Actions runner
func (gr *GitHubRunner) configureRunner(ctx context.Context, runnerDir, runnerName, registrationToken string) (string, int64, error) {
	// Run config.sh
	cmd := exec.CommandContext(ctx, filepath.Join(runnerDir, "config.sh"),
		"--url", gr.githubURL,
		"--token", registrationToken,
		"--name", runnerName,
		"--labels", strings.Join(gr.labels, ","),
		"--work", "/work",
		"--unattended",
		"--replace",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("failed to configure runner: %w: %s", err, output)
	}

	// Get runner ID from .runner file
	runnerFile := filepath.Join(runnerDir, ".runner")
	data, err := ioutil.ReadFile(runnerFile)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read runner file: %w", err)
	}

	var runnerConfig struct {
		AgentId int64 `json:"agentId"`
	}
	if err := json.Unmarshal(data, &runnerConfig); err != nil {
		return "", 0, fmt.Errorf("failed to parse runner file: %w", err)
	}

	return registrationToken, runnerConfig.AgentId, nil
}

// UnregisterRunner unregisters a GitHub Actions runner
func (gr *GitHubRunner) UnregisterRunner(ctx context.Context, runnerID string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Remove runner from GitHub
	if err := gr.removeRunnerFromGitHub(ctx, runner.RunnerID); err != nil {
		// Log error but continue
		fmt.Printf("[GitHubRunner] Warning: failed to remove from GitHub: %v\n", err)
	}

	// Mark as destroyed
	now := time.Now()
	runner.DestroyedAt = &now

	// Delete from database
	if err := gr.db.DeleteRunner(ctx, runnerID); err != nil {
		return fmt.Errorf("failed to delete runner: %w", err)
	}

	// Delete from memory
	delete(gr.runners, runnerID)

	// Save state
	if err := gr.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// removeRunnerFromGitHub removes the runner from GitHub API
func (gr *GitHubRunner) removeRunnerFromGitHub(ctx context.Context, runnerID int64) error {
	client := &http.Client{Timeout: 30 * time.Second}

	url := fmt.Sprintf("%s/api/v3/repos/%s/actions/runners/%d", gr.githubURL, "owner/repo", runnerID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", gr.personalToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to remove runner: %s", resp.Status)
	}

	return nil
}

// StartRunner starts a GitHub Actions runner on a VM
func (gr *GitHubRunner) StartRunner(ctx context.Context, runnerID, vmIP string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Update runner status
	runner.IPAddress = vmIP
	runner.Status = "running"
	now := time.Now()
	runner.LastHeartbeat = &now

	// Get runner directory
	runnerDir := filepath.Join(os.Getenv("HOME"), ".vimic2", "runners", runner.VMID)

	// Copy runner to VM
	scpCmd := exec.CommandContext(ctx, "scp",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-r", runnerDir,
		fmt.Sprintf("%s@%s:/tmp/runner", gr.runnerUser, vmIP),
	)

	if output, err := scpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy runner: %w: %s", err, output)
	}

	// Start runner via SSH
	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", gr.runnerUser, vmIP),
		fmt.Sprintf("cd /tmp/runner && nohup ./run.sh > /tmp/runner.log 2>&1 &"),
	)

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start runner: %w: %s", err, output)
	}

	// Save state
	if err := gr.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// StopRunner stops a GitHub Actions runner
func (gr *GitHubRunner) StopRunner(ctx context.Context, runnerID string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	if runner.IPAddress == "" {
		return fmt.Errorf("runner not started")
	}

	// Stop runner via SSH
	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", gr.runnerUser, runner.IPAddress),
		"pkill -f Actions.Listener || true",
	)

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop runner: %w: %s", err, output)
	}

	runner.Status = "stopped"

	// Save state
	if err := gr.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetRunner returns a runner by ID
func (gr *GitHubRunner) GetRunner(runnerID string) (*GitHubRunnerInfo, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}

	return runner, nil
}

// ListRunners returns all GitHub runners
func (gr *GitHubRunner) ListRunners() []*GitHubRunnerInfo {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitHubRunnerInfo, 0, len(gr.runners))
	for _, runner := range gr.runners {
		runners = append(runners, runner)
	}
	return runners
}

// ListRunnersByPipeline returns runners for a pipeline
func (gr *GitHubRunner) ListRunnersByPipeline(pipelineID string) []*GitHubRunnerInfo {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitHubRunnerInfo, 0)
	for _, runner := range gr.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

// UpdateHeartbeat updates the runner heartbeat
func (gr *GitHubRunner) UpdateHeartbeat(runnerID string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	now := time.Now()
	runner.LastHeartbeat = &now

	// Save state
	if err := gr.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetStats returns runner statistics
func (gr *GitHubRunner) GetStats() map[string]int {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	stats := map[string]int{
		"total": len(gr.runners),
	}

	for _, runner := range gr.runners {
		stats[runner.Status]++
	}

	return stats
}