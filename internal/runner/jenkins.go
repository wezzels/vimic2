// Package runner provides Jenkins agent implementation.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

// JenkinsRunner implements RunnerInterface for Jenkins.
type JenkinsRunner struct {
	BaseRunner
	client    *http.Client
	url       string
	username  string
	apiToken  string
	name      string
	agentName string
	secret    string
}

// NewJenkinsRunner creates a new Jenkins agent.
func NewJenkinsRunner(config *JenkinsConfig) (*JenkinsRunner, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("jenkins url is required")
	}
	if config.Username == "" || config.APIToken == "" {
		return nil, fmt.Errorf("jenkins credentials are required")
	}

	runner := &JenkinsRunner{
		BaseRunner: BaseRunner{
			id:       generateID("jenkins"),
			platform: types.PlatformJenkins,
			labels:   config.Labels,
			status:   types.RunnerStatusCreating,
			health:   &HealthStatus{Healthy: false},
		},
		url:       config.URL,
		username:  config.Username,
		apiToken:  config.APIToken,
		name:      config.Name,
		agentName: config.AgentName,
		secret:    config.Secret,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	return runner, nil
}

// Register registers the agent with Jenkins.
func (r *JenkinsRunner) Register(ctx context.Context) error {
	r.registered = true
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}

// Start starts the Jenkins agent process.
func (r *JenkinsRunner) Start(ctx context.Context) error {
	if !r.registered {
		return fmt.Errorf("agent must be registered before starting")
	}

	r.SetStatus(types.RunnerStatusOnline)
	r.health = &HealthStatus{
		Healthy:   true,
		LastCheck: time.Now(),
	}

	return nil
}

// Stop gracefully stops the agent.
func (r *JenkinsRunner) Stop(ctx context.Context) error {
	r.SetStatus(types.RunnerStatusOffline)
	r.health.Healthy = false
	return nil
}

// Unregister removes the agent from Jenkins.
func (r *JenkinsRunner) Unregister(ctx context.Context) error {
	r.registered = false
	r.SetStatus(types.RunnerStatusOffline)
	return nil
}

// ListJenkinsAgents lists all agents in Jenkins.
func (r *JenkinsRunner) ListJenkinsAgents(ctx context.Context) ([]RunnerInfo, error) {
	agentsURL := fmt.Sprintf("%s/computer/api/json", r.url)

	req, err := http.NewRequestWithContext(ctx, "GET", agentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list agents request: %w", err)
	}
	req.SetBasicAuth(r.username, r.apiToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer resp.Body.Close()

	var listResp struct {
		Computer []struct {
			DisplayName  string `json:"displayName"`
			Offline      bool   `json:"offline"`
			OfflineCause string `json:"offlineCause"`
		} `json:"computer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to parse agents response: %w", err)
	}

	agents := make([]RunnerInfo, 0, len(listResp.Computer))
	for _, c := range listResp.Computer {
		status := types.RunnerStatusOnline
		if c.Offline {
			status = types.RunnerStatusOffline
		}
		agents = append(agents, RunnerInfo{
			Name:   c.DisplayName,
			Status: status,
		})
	}

	return agents, nil
}
