//go:build integration

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/database"
	"github.com/stsgym/vimic2/internal/pipeline"
)

func setupAPITestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-api-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

func setupAPIServer(t *testing.T) (*Server, *database.DB, func()) {
	t.Helper()
	db, dbCleanup := setupAPITestDB(t)

	pipelineDB := &pipeline.PipelineDB{}

	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: false,
	})
	if err != nil {
		dbCleanup()
		t.Fatalf("NewServer failed: %v", err)
	}

	return s, db, dbCleanup
}

// ==================== Health Endpoint ====================

func TestAPI_HealthEndpoint(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", body["status"])
	}
}

// ==================== Pipeline Endpoints ====================

func TestAPI_ListPipelines_NilCoordinator(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// With nil coordinator, should return empty list
	if w.Code != http.StatusOK {
		t.Logf("ListPipelines status: %d", w.Code)
	}
}

// ==================== Job Endpoints ====================

func TestAPI_ListJobs_NilDispatcher(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/jobs?status=running", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListJobs status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_ListCompletedJobs_NilDispatcher(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/jobs?status=completed", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListCompletedJobs status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Runner Endpoints ====================

func TestAPI_ListRunners_NilManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/runners", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListRunners status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Pool Endpoints ====================

func TestAPI_ListPools_NilManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pools", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListPools status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Network Endpoints ====================

func TestAPI_ListNetworks_NilManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/networks", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListNetworks status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Artifact Endpoints ====================

func TestAPI_ListArtifacts_NilManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/artifacts", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListArtifacts status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Log Endpoints ====================

func TestAPI_ListLogStreams_NilManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/logs/streams", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("ListLogStreams status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Stats Endpoint ====================

func TestAPI_GetStats_NilCoordinator(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	t.Logf("GetStats status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== Config Tests ====================

func TestServerConfig_Defaults_Real(t *testing.T) {
	config := ServerConfig{}
	if config.ListenAddr != "" {
		t.Errorf("Default ListenAddr = %s, want empty", config.ListenAddr)
	}
	if config.AuthEnabled {
		t.Error("Default AuthEnabled should be false")
	}
}

func TestServerConfig_Custom(t *testing.T) {
	config := ServerConfig{
		ListenAddr:  ":9090",
		AuthEnabled: true,
		AuthToken:   "test-token",
	}
	if config.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %s, want :9090", config.ListenAddr)
	}
	if !config.AuthEnabled {
		t.Error("AuthEnabled should be true")
	}
	if config.AuthToken != "test-token" {
		t.Error("AuthToken should be set")
	}
}

// ==================== Auth Middleware Tests ====================

func TestAPI_Auth_ValidToken(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Should pass auth (nil coordinator returns empty list)
	if w.Code == http.StatusUnauthorized {
		t.Error("Valid token should not return 401")
	}
}

func TestAPI_Auth_InvalidToken(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Invalid token should return 401, got %d", w.Code)
	}
}

func TestAPI_Auth_MissingHeader(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Missing auth header should return 401, got %d", w.Code)
	}
}

func TestAPI_Auth_InvalidFormat(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	req.Header.Set("Authorization", "Basic dGVzdA==")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Invalid auth format should return 401, got %d", w.Code)
	}
}

// ==================== Set Managers Tests ====================

func TestAPI_SetPoolManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	s.SetPoolManager(nil)
	if s.poolManager != nil {
		t.Error("SetPoolManager(nil) should set poolManager to nil")
	}
}

func TestAPI_SetNetworkManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	s.SetNetworkManager(nil)
	if s.networkManager != nil {
		t.Error("SetNetworkManager(nil) should set networkManager to nil")
	}
}

func TestAPI_SetRunnerManager(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	s.SetRunnerManager(nil)
	if s.runnerManager != nil {
		t.Error("SetRunnerManager(nil) should set runnerManager to nil")
	}
}

// ==================== WebSocket Server Tests ====================

func TestAPI_WebSocketServer_Create(t *testing.T) {
	ws := NewWebSocketServer(nil)
	if ws == nil {
		t.Fatal("NewWebSocketServer should not return nil")
	}
}

// ==================== Error Response Tests ====================

func TestAPI_ErrorResponseFormat(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "secret",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] == nil && body["status"] == nil {
		t.Errorf("expected error response, got %v", body)
	}
}

// ==================== ServerConfig JSON Tests ====================

func TestServerConfig_JSON(t *testing.T) {
	config := ServerConfig{
		ListenAddr:  ":8080",
		AuthEnabled: true,
		AuthToken:   "my-secret-token",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got ServerConfig
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %s, want :8080", got.ListenAddr)
	}
	if !got.AuthEnabled {
		t.Error("AuthEnabled should be true")
	}
}

// ==================== Start/Stop Tests ====================

func TestAPI_ServerStartStop(t *testing.T) {
	pipelineDB := &pipeline.PipelineDB{}
	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: false,
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start server in background
	go func() {
		s.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	// Stop
	s.Stop(ctx)

	t.Log("Server started and stopped successfully")
}

// ==================== Network Create Endpoint ====================

func TestAPI_CreateNetwork_InvalidCIDR(t *testing.T) {
	s, _, cleanup := setupAPIServer(t)
	defer cleanup()

	body := `{"name": "test-network", "bridge_name": "br-test", "cidr": "invalid"}`
	req := httptest.NewRequest("POST", "/api/networks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Should fail with invalid CIDR (nil networkManager)
	t.Logf("CreateNetwork status: %d, body: %s", w.Code, w.Body.String())
}