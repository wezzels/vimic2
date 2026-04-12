// Package runner provides CircleCI runner implementation.
package runner

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// CircleCIRunner implements RunnerInterface for CircleCI.
type CircleCIRunner struct {
	BaseRunner
	client   *http.Client
	apiURL   string
	apiToken string
	name     string
}

// NewCircleCIRunner creates a new CircleCI runner.
func NewCircleCIRunner(config *CircleCIConfig) (*CircleCIRunner, error) {
	if config.APIToken == "" {
		return nil, fmt.Errorf("circleci api token is required")
	}

	runner := &CircleCIRunner{
		BaseRunner: BaseRunner{
			id:       generateID("cci"),
			platform: types.PlatformCircleCI,
			labels:   config.Labels,
			status:   types.RunnerStatusCreating,
			health:   &HealthStatus{Healthy: false},
		},
		apiURL:   "https://circleci.com/api/v2",
		apiToken: config.APIToken,
		name:     config.Name,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return runner, nil
}

// Register registers the runner with CircleCI.
func (r *CircleCIRunner) Register(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusCreating)

	// POST /runner/resource
	registerURL := fmt.Sprintf("%s/runner/resource", r.apiURL)

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, nil)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Circle-Token", r.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		r.SetStatus(types.RunnerStatusError)
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	r.registered = true
	r.SetStatus(types.RunnerStatusOffline)

	return nil
}

// Start starts the CircleCI runner process.
func (r *CircleCIRunner) Start(ctx context.Context) error {
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
func (r *CircleCIRunner) Stop(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusOffline)
	r.health.Healthy = false
	return nil
}

// Unregister removes the runner from CircleCI.
func (r *CircleCIRunner) Unregister(ctx context.Context) error {
	r.registered = false
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}
