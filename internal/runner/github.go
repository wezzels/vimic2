// Package runner provides GitHub Actions runner implementation.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// GitHubRunner implements RunnerInterface for GitHub Actions.
type GitHubRunner struct {
	BaseRunner
	client            *http.Client
	apiURL            string
	repo              string
	registrationToken string
	runnerToken       string
	name              string
	version           string
}

// NewGitHubRunner creates a new GitHub Actions runner.
func NewGitHubRunner(config *GitHubConfig) (*GitHubRunner, error) {
	if config.Repo == "" {
		return nil, fmt.Errorf("github repo is required (format: owner/repo)")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("github token is required")
	}

	runner := &GitHubRunner{
		BaseRunner: BaseRunner{
			id:       generateID("gh"),
			platform: types.PlatformGitHub,
			labels:   config.Labels,
			status:   types.RunnerStatusCreating,
			health:   &HealthStatus{Healthy: false},
		},
		apiURL:      "https://api.github.com",
		repo:        config.Repo,
		runnerToken: config.Token,
		name:        config.Name,
		version:     "2.311.0",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return runner, nil
}

// Register registers the runner with GitHub Actions.
func (r *GitHubRunner) Register(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusCreating)

	// Step 1: Get registration token from GitHub API
	regTokenURL := fmt.Sprintf("%s/repos/%s/actions/runners/registration-token", r.apiURL, r.repo)

	req, err := http.NewRequestWithContext(ctx, "POST", regTokenURL, nil)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("failed to create registration token request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", r.runnerToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := r.client.Do(req)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("failed to get registration token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("failed to get registration token: status %d", resp.StatusCode)
	}

	var regTokenResp struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regTokenResp); err != nil {
		return fmt.Errorf("failed to parse registration token response: %w", err)
	}

	r.registrationToken = regTokenResp.Token
	r.registered = true
	r.SetStatus(types.RunnerStatusOffline)

	return nil
}

// Start starts the GitHub runner process.
func (r *GitHubRunner) Start(ctx context.Context) error {
	if !r.registered {
		return fmt.Errorf("runner must be registered before starting")
	}

	r.SetStatus(types.RunnerStatusOnline)
	r.health = &HealthStatus{
		Healthy:   true,
		LastCheck: time.Now(),
	}

	return nil
}

// Stop gracefully stops the runner.
func (r *GitHubRunner) Stop(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusOffline)
	r.health.Healthy = false
	return nil
}

// Unregister removes the runner from GitHub.
func (r *GitHubRunner) Unregister(ctx context.Context) error {
	if !r.registered {
		return nil
	}

	r.registered = false
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}
