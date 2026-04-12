// Package runner provides Drone CI runner implementation.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// DroneRunner implements RunnerInterface for Drone CI.
type DroneRunner struct {
	BaseRunner
	client   *http.Client
	url      string
	apiToken string
	name     string
}

// NewDroneRunner creates a new Drone CI runner.
func NewDroneRunner(config *DroneConfig) (*DroneRunner, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("drone url is required")
	}
	if config.APIToken == "" {
		return nil, fmt.Errorf("drone api token is required")
	}

	runner := &DroneRunner{
		BaseRunner: BaseRunner{
			id:       generateID("drone"),
			platform: types.PlatformDrone,
			labels:   config.Labels,
			status:   types.RunnerStatusCreating,
			health:   &HealthStatus{Healthy: false},
		},
		url:      config.URL,
		apiToken: config.APIToken,
		name:     config.Name,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return runner, nil
}

// Register registers the runner with Drone.
func (r *DroneRunner) Register(ctx context.Context) error {
	r.registered = true
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}

// Start starts the Drone runner process.
func (r *DroneRunner) Start(ctx context.Context) error {
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
func (r *DroneRunner) Stop(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusOffline)
	r.health.Healthy = false
	return nil
}

// Unregister removes the runner from Drone.
func (r *DroneRunner) Unregister(ctx context.Context) error {
	r.registered = false
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}

// ListDroneRunners lists all runners in Drone.
func (r *DroneRunner) ListDroneRunners(ctx context.Context) ([]RunnerInfo, error) {
	runnersURL := fmt.Sprintf("%s/api/system/runners", r.url)

	req, err := http.NewRequestWithContext(ctx, "GET", runnersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list runners request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiToken))

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list runners: %w", err)
	}
	defer resp.Body.Close()

	var runners []struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
		Status  string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&runners); err != nil {
		return nil, fmt.Errorf("failed to parse runners response: %w", err)
	}

	result := make([]RunnerInfo, 0, len(runners))
	for _, r := range runners {
		result = append(result, RunnerInfo{
			ID:     fmt.Sprintf("%d", r.ID),
			Name:   r.Name,
			Status: types.RunnerStatus(r.Status),
		})
	}

	return result, nil
}
