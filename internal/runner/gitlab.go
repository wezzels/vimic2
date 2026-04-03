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
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// GitLabRunner manages GitLab CI runners
type GitLabRunner struct {
	db           *types.PipelineDB
	gitlabURL    string
	registrationToken string
	runnerToken  string
	runnerName   string
	tags         []string
	runnerUser   string
	stateFile    string
	runners      map[string]*GitLabRunnerInfo
	mu           sync.RWMutex
}

// GitLabRunnerInfo represents a registered GitLab runner
type GitLabRunnerInfo struct {
	ID             string       `json:"id"`
	VMID           string       `json:"vm_id"`
	PipelineID     string       `json:"pipeline_id"`
	Token          string       `json:"token"`
	Name           string       `json:"name"`
	Tags           []string     `json:"tags"`
	PlatformID     int          `json:"platform_runner_id"`
	Configuration  string       `json:"configuration"`
	IPAddress      string       `json:"ip_address"`
	Status         string       `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
	LastHeartbeat  *time.Time   `json:"last_heartbeat,omitempty"`
	DestroyedAt     *time.Time   `json:"destroyed_at,omitempty"`
}

// GitLabConfig represents GitLab runner configuration
type GitLabConfig struct {
	URL               string   `json:"url"`
	RegistrationToken string   `json:"registration_token"`
	Token             string   `json:"token"`
	Labels            []string `json:"labels"`
	Username          string   `json:"username"`
}

// NewGitLabRunner creates a new GitLab runner manager
func NewGitLabRunner(db *types.PipelineDB, config *GitLabConfig) (*GitLabRunner, error) {
	gr := &GitLabRunner{
		db:               db,
		gitlabURL:        config.URL,
		registrationToken: config.RegistrationToken,
		runnerToken:      config.Token,
		runnerName:       "vimic2-runner",
		tags:             config.Labels,
		runnerUser:        config.Username,
		runners:          make(map[string]*GitLabRunnerInfo),
	}

	if gr.runnerUser == "" {
		gr.runnerUser = "gitlab-runner"
	}

	// Load state
	if err := gr.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return gr, nil
}

// loadState loads runner state from disk
func (gr *GitLabRunner) loadState() error {
	stateFile := gr.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var runners []*GitLabRunnerInfo
	if err := json.Unmarshal(data, &runners); err != nil {
		return err
	}

	for _, runner := range runners {
		gr.runners[runner.ID] = runner
	}

	return nil
}

// saveState saves runner state to disk
func (gr *GitLabRunner) saveState() error {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitLabRunnerInfo, 0, len(gr.runners))
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
func (gr *GitLabRunner) getStateFile() string {
	if gr.stateFile != "" {
		return gr.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "gitlab-runners.json")
}

// SetStateFile sets the state file path
func (gr *GitLabRunner) SetStateFile(path string) {
	gr.stateFile = path
}

// RegisterRunner registers a new GitLab runner
func (gr *GitLabRunner) RegisterRunner(ctx context.Context, vmID, pipelineID string) (*GitLabRunnerInfo, error) {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	// Generate runner name
	runnerName := fmt.Sprintf("%s-%s", gr.runnerName, vmID[:8])

	// Register with GitLab
	token, platformID, err := gr.registerWithGitLab(ctx, runnerName)
	if err != nil {
		return nil, fmt.Errorf("failed to register with GitLab: %w", err)
	}

	// Generate configuration
	config := gr.generateConfiguration(token)

	// Create runner info
	runner := &GitLabRunnerInfo{
		ID:            generateRunnerID("gl"),
		VMID:          vmID,
		PipelineID:    pipelineID,
		Token:         token,
		Name:          runnerName,
		Tags:          gr.tags,
		PlatformID:    platformID,
		Configuration: config,
		Status:        "registered",
		CreatedAt:     time.Now(),
	}

	gr.runners[runner.ID] = runner

	// Save to database
	dbRunner := &pipeline.Runner{
		ID:         runner.ID,
		PipelineID: pipelineID,
		VMID:       vmID,
		Platform:   pipeline.PlatformGitLab,
		PlatformID: fmt.Sprintf("%d", platformID),
		Token:      token,
		Labels:     gr.tags,
		Name:       runnerName,
		Status:     types.RunnerStatusCreating,
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

// registerWithGitLab registers the runner with GitLab API
func (gr *GitLabRunner) registerWithGitLab(ctx context.Context, runnerName string) (string, int, error) {
	// Use gitlab-runner register command
	configPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-config.toml", runnerName))

	// Run gitlab-runner register
	cmd := exec.CommandContext(ctx, "gitlab-runner", "register",
		"--non-interactive",
		"--url", gr.gitlabURL,
		"--registration-token", gr.registrationToken,
		"--executor", "shell",
		"--description", runnerName,
		"--tag-list", joinTags(gr.tags),
		"--config", configPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", 0, fmt.Errorf("registration failed: %w: %s", err, output)
	}

	// Read token from config
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse token from config
	token := gr.parseTokenFromConfig(string(configData))

	// Get runner ID from GitLab API
	platformID, err := gr.getRunnerIDFromGitLab(ctx, token)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get runner ID: %w", err)
	}

	// Clean up temp config
	os.Remove(configPath)

	return token, platformID, nil
}

// parseTokenFromConfig extracts token from GitLab runner config
func (gr *GitLabRunner) parseTokenFromConfig(config string) string {
	// Parse TOML-like config for token
	lines := []string{}
	for _, line := range splitLines(config) {
		lines = append(lines, line)
	}

	for i, line := range lines {
		if len(line) > 6 && line[:6] == "[[runners]]" {
			// Look for token in subsequent lines
			for j := i + 1; j < len(lines) && len(lines[j]) > 0 && lines[j][0] != '['; j++ {
				if len(lines[j]) > 8 && lines[j][:8] == "  token =" {
					return trimQuotes(lines[j][9:])
				}
			}
		}
	}

	return ""
}

// getRunnerIDFromGitLab gets the runner ID from GitLab API
func (gr *GitLabRunner) getRunnerIDFromGitLab(ctx context.Context, token string) (int, error) {
	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Get runners list
	url := fmt.Sprintf("%s/api/v4/runners", gr.gitlabURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("PRIVATE-TOKEN", gr.runnerToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get runners: %s", resp.Status)
	}

	var runners []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Token       string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runners); err != nil {
		return 0, err
	}

	// Find runner by token
	for _, runner := range runners {
		if runner.Token == token {
			return runner.ID, nil
		}
	}

	// If not found by token, get latest runner
	if len(runners) > 0 {
		return runners[0].ID, nil
	}

	return 0, fmt.Errorf("runner not found")
}

// generateConfiguration generates GitLab runner configuration
func (gr *GitLabRunner) generateConfiguration(token string) string {
	return fmt.Sprintf(`concurrent = 4
check_interval = 3

[[runners]]
  name = "%s"
  url = "%s"
  token = "%s"
  executor = "shell"
  shell = "bash"
  builds_dir = "/work/builds"
  cache_dir = "/work/cache"
  [runners.custom_build_dir]
    enabled = true
  [runners.cache]
    [runners.cache.s3]
    [runners.cache.gcs]
`, gr.runnerName, gr.gitlabURL, token)
}

// UnregisterRunner unregisters a GitLab runner
func (gr *GitLabRunner) UnregisterRunner(ctx context.Context, runnerID string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return fmt.Errorf("runner not found: %s", runnerID)
	}

	// Unregister from GitLab
	if err := gr.unregisterFromGitLab(ctx, runner.Token); err != nil {
		// Log error but continue
		fmt.Printf("[GitLabRunner] Warning: failed to unregister from GitLab: %v\n", err)
	}

	// Stop runner process
	cmd := exec.CommandContext(ctx, "gitlab-runner", "unregister",
		"--url", gr.gitlabURL,
		"--token", runner.Token,
	)
	cmd.Run() // Ignore error

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

// unregisterFromGitLab unregisters the runner from GitLab API
func (gr *GitLabRunner) unregisterFromGitLab(ctx context.Context, token string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/api/v4/runners", gr.gitlabURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", gr.runnerToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to unregister: %s", resp.Status)
	}

	return nil
}

// StartRunner starts a GitLab runner on a VM
func (gr *GitLabRunner) StartRunner(ctx context.Context, runnerID, vmIP string) error {
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

	// Generate config for this VM
	configPath := filepath.Join(os.Getenv("HOME"), ".vimic2", "runners", fmt.Sprintf("%s-config.toml", runnerID))
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	config := gr.generateConfiguration(runner.Token)
	if err := ioutil.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Start runner via SSH
	sshCmd := exec.CommandContext(ctx, "ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", gr.runnerUser, vmIP),
		"nohup gitlab-runner run --config /tmp/gitlab-runner-config.toml > /tmp/gitlab-runner.log 2>&1 &",
	)

	// Copy config to VM
	scpCmd := exec.CommandContext(ctx, "scp",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		configPath,
		fmt.Sprintf("%s@%s:/tmp/gitlab-runner-config.toml", gr.runnerUser, vmIP),
	)

	if output, err := scpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy config: %w: %s", err, output)
	}

	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start runner: %w: %s", err, output)
	}

	// Save state
	if err := gr.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// StopRunner stops a GitLab runner
func (gr *GitLabRunner) StopRunner(ctx context.Context, runnerID string) error {
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
		"pkill -f gitlab-runner || true",
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
func (gr *GitLabRunner) GetRunner(runnerID string) (*GitLabRunnerInfo, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runner, ok := gr.runners[runnerID]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", runnerID)
	}

	return runner, nil
}

// ListRunners returns all GitLab runners
func (gr *GitLabRunner) ListRunners() []*GitLabRunnerInfo {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitLabRunnerInfo, 0, len(gr.runners))
	for _, runner := range gr.runners {
		runners = append(runners, runner)
	}
	return runners
}

// ListRunnersByPipeline returns runners for a pipeline
func (gr *GitLabRunner) ListRunnersByPipeline(pipelineID string) []*GitLabRunnerInfo {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	runners := make([]*GitLabRunnerInfo, 0)
	for _, runner := range gr.runners {
		if runner.PipelineID == pipelineID {
			runners = append(runners, runner)
		}
	}
	return runners
}

// UpdateHeartbeat updates the runner heartbeat
func (gr *GitLabRunner) UpdateHeartbeat(runnerID string) error {
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
func (gr *GitLabRunner) GetStats() map[string]int {
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

// Helper functions

func generateRunnerID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, randomString(8))
}

func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}