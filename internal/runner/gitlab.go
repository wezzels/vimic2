// Package runner provides GitLab runner implementation.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// GitLabRunner implements RunnerInterface for GitLab CI/CD.
type GitLabRunner struct {
	BaseRunner
	client     *http.Client
	url        string
	token      string
	name       string
	configPath string
}

// NewGitLabRunner creates a new GitLab runner.
func NewGitLabRunner(config *GitLabConfig) (*GitLabRunner, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("gitlab url is required")
	}
	
	runner := &GitLabRunner{
		BaseRunner: BaseRunner{
			id:       generateID("gl"),
			platform: types.PlatformGitLab,
			labels:   config.Labels,
			status:   types.RunnerStatusCreating,
			health:   &HealthStatus{Healthy: false},
		},
		url:        config.URL,
		token:      config.Token,
		name:       config.Name,
		configPath: config.ConfigPath,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	return runner, nil
}

// Register registers the runner with GitLab.
func (r *GitLabRunner) Register(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusCreating)
	
	// Prepare registration request
	registerURL := fmt.Sprintf("%s/api/v4/runners", r.url)
	data := url.Values{}
	data.Set("token", r.token)
	data.Set("description", r.name)
	data.Set("tag_list", joinLabels(r.labels))
	data.Set("run_untagged", "true")
	
	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, nil)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("failed to create registration request: %w", err)
	}
	
	resp, err := r.client.Do(req)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}
	
	// Parse response to get runner token
	var result struct {
		Token string `json:"token"`
		ID    int    `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse registration response: %w", err)
	}
	
	r.token = result.Token
	r.registered = true
	r.SetStatus(types.RunnerStatusOffline)
	
	return nil
}

// Start starts the GitLab runner process.
func (r *GitLabRunner) Start(ctx context.Context) error {
	if !r.registered {
		return fmt.Errorf("runner must be registered before starting")
	}
	
	// In production, this would:
	// 1. Download gitlab-runner binary if not present
	// 2. Create config.toml
	// 3. Start gitlab-runner process
	// 4. Monitor process health
	
	r.SetStatus(types.RunnerStatusOnline)
	r.health = &HealthStatus{
		Healthy:   true,
		LastCheck: time.Now(),
	}
	
	return nil
}

// Stop gracefully stops the runner.
func (r *GitLabRunner) Stop(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusOffline)
	
	// In production, this would:
	// 1. Send SIGTERM to gitlab-runner process
	// 2. Wait for graceful shutdown
	// 3. Force kill if timeout
	
	r.SetStatus(types.RunnerStatusOffline)
	r.health.Healthy = false
	
	return nil
}

// Unregister removes the runner from GitLab.
func (r *GitLabRunner) Unregister(ctx context.Context) error {
	if !r.registered {
		return nil
	}
	
	// DELETE /api/v4/runners/:id
	deleteURL := fmt.Sprintf("%s/api/v4/runners/%s", r.url, r.id)
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create unregister request: %w", err)
	}
	
	req.Header.Set("Private-Token", r.token)
	
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("unregister request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unregister failed with status: %d", resp.StatusCode)
	}
	
	r.registered = false
	r.SetStatus(types.RunnerStatusOffline)
	
	return nil
}

// joinLabels joins labels with comma.
func joinLabels(labels []string) string {
	result := ""
	for i, label := range labels {
		if i > 0 {
			result += ","
		}
		result += label
	}
	return result
}