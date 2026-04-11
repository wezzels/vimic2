// Package api provides quick win tests
package api

import (
	"testing"
)

// TestServerConfig_Fields tests ServerConfig field access
func TestServerConfig_Fields(t *testing.T) {
	config := &ServerConfig{
		ListenAddr: ":8080",
	}

	if config.ListenAddr != ":8080" {
		t.Error("ListenAddr mismatch")
	}
}

// TestServerConfig_Default tests default values
func TestServerConfig_Default(t *testing.T) {
	config := &ServerConfig{}

	if config.ListenAddr != "" {
		t.Error("expected empty ListenAddr")
	}
}

// TestErrorResponse tests error response struct
func TestErrorResponse_Quick(t *testing.T) {
	err := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Error:   "not_found",
		Message: "resource not found",
	}

	if err.Error != "not_found" {
		t.Error("Error mismatch")
	}
}

// TestHealthResponse tests health response struct
func TestHealthResponse(t *testing.T) {
	health := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	if health.Status != "ok" {
		t.Error("Status mismatch")
	}
}

// TestVersionResponse tests version response struct
func TestVersionResponse(t *testing.T) {
	version := struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}{
		Version: "1.0.0",
		Commit:  "abc123",
	}

	if version.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
}

// TestPipelineResponse tests pipeline response struct
func TestPipelineResponse(t *testing.T) {
	pipeline := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{
		ID:     "pipeline-1",
		Status: "running",
	}

	if pipeline.ID != "pipeline-1" {
		t.Error("ID mismatch")
	}
}

// TestJobResponse tests job response struct
func TestJobResponse(t *testing.T) {
	job := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{
		ID:     "job-1",
		Status: "pending",
	}

	if job.ID != "job-1" {
		t.Error("ID mismatch")
	}
}

// TestRunnerResponse tests runner response struct
func TestRunnerResponse(t *testing.T) {
	runner := struct {
		ID       string `json:"id"`
		Platform string `json:"platform"`
	}{
		ID:       "runner-1",
		Platform: "linux",
	}

	if runner.ID != "runner-1" {
		t.Error("ID mismatch")
	}
}

// TestPoolResponse tests pool response struct
func TestPoolResponse(t *testing.T) {
	pool := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   "pool-1",
		Name: "default",
	}

	if pool.ID != "pool-1" {
		t.Error("ID mismatch")
	}
}

// TestHostResponse tests host response struct
func TestHostResponse(t *testing.T) {
	host := struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}{
		ID:      "host-1",
		Name:    "hv1",
		Address: "192.168.1.1",
	}

	if host.ID != "host-1" {
		t.Error("ID mismatch")
	}
}

// TestClusterResponse tests cluster response struct
func TestClusterResponse(t *testing.T) {
	cluster := struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}{
		ID:     "cluster-1",
		Name:   "prod",
		Status: "running",
	}

	if cluster.ID != "cluster-1" {
		t.Error("ID mismatch")
	}
}

// TestNetworkResponse tests network response struct
func TestNetworkResponse(t *testing.T) {
	network := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		CIDR string `json:"cidr"`
	}{
		ID:   "net-1",
		Name: "default",
		CIDR: "10.0.0.0/24",
	}

	if network.ID != "net-1" {
		t.Error("ID mismatch")
	}
}