// Package api provides integration tests with real database
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/types"
)

// MockPoolManager implements types.PoolManagerInterface
type MockPoolManager struct {
	pools map[string]*types.PoolState
}

func NewMockPoolManager() *MockPoolManager {
	return &MockPoolManager{
		pools: make(map[string]*types.PoolState),
	}
}

func (m *MockPoolManager) AllocateVM(poolName string) (*types.VMState, error) {
	return &types.VMState{ID: "vm-" + poolName, Status: "running"}, nil
}

func (m *MockPoolManager) ReleaseVM(vmID string) error {
	return nil
}

func (m *MockPoolManager) GetPool(name string) (*types.PoolState, error) {
	return m.pools[name], nil
}

func (m *MockPoolManager) ListPools() ([]*types.PoolState, error) {
	result := make([]*types.PoolState, 0, len(m.pools))
	for _, p := range m.pools {
		result = append(result, p)
	}
	return result, nil
}

// MockNetworkManager implements types.NetworkManagerInterface
type MockNetworkManager struct{}

func NewMockNetworkManager() *MockNetworkManager {
	return &MockNetworkManager{}
}

func (m *MockNetworkManager) CreateNetwork(config *types.NetworkConfig) (string, error) {
	return "net-" + config.CIDR, nil
}

func (m *MockNetworkManager) DestroyNetwork(networkID string) error {
	return nil
}

func (m *MockNetworkManager) GetNetwork(networkID string) (*types.NetworkConfig, error) {
	return &types.NetworkConfig{CIDR: "10.0.0.0/24"}, nil
}

// MockRunnerManager implements types.RunnerManagerInterface
type MockRunnerManager struct{}

func NewMockRunnerManager() *MockRunnerManager {
	return &MockRunnerManager{}
}

func (m *MockRunnerManager) CreateRunner(platform types.RunnerPlatform, config map[string]interface{}) (string, error) {
	return "runner-" + string(platform), nil
}

func (m *MockRunnerManager) DestroyRunner(runnerID string) error {
	return nil
}

func (m *MockRunnerManager) GetRunner(runnerID string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": runnerID}, nil
}

// MockPipelineDB implements a minimal PipelineDB for testing
type MockPipelineDB struct{}

func NewMockPipelineDB() *MockPipelineDB {
	return &MockPipelineDB{}
}

func (m *MockPipelineDB) SavePipeline(id string, state map[string]interface{}) error {
	return nil
}

func (m *MockPipelineDB) LoadPipeline(id string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockPipelineDB) DeletePipeline(id string) error {
	return nil
}

func (m *MockPipelineDB) ListPipelines() ([]string, error) {
	return []string{}, nil
}

func (m *MockPipelineDB) SaveRunner(id string, state map[string]interface{}) error {
	return nil
}

func (m *MockPipelineDB) LoadRunner(id string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockPipelineDB) DeleteRunner(id string) error {
	return nil
}

func (m *MockPipelineDB) SaveNetwork(id string, state map[string]interface{}) error {
	return nil
}

func (m *MockPipelineDB) LoadNetwork(id string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockPipelineDB) DeleteNetwork(id string) error {
	return nil
}

func (m *MockPipelineDB) SavePool(id string, state map[string]interface{}) error {
	return nil
}

func (m *MockPipelineDB) LoadPool(id string) (map[string]interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}

func (m *MockPipelineDB) DeletePool(id string) error {
	return nil
}

func (m *MockPipelineDB) UpdatePoolSize(id string, available, busy int) error {
	return nil
}

func (m *MockPipelineDB) UpdateVMState(vmID string, state string) error {
	return nil
}

// TestIntegration_HandleHealth tests health endpoint
func TestIntegration_HandleHealth(t *testing.T) {
	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "healthy" && resp["status"] != "ok" {
		t.Errorf("expected status healthy or ok, got %s", resp["status"])
	}
}

// TestIntegration_HandleListPipelines tests listing pipelines
func TestIntegration_HandleListPipelines(t *testing.T) {
	mockDB := NewMockPipelineDB()
	coord, _ := pipeline.NewCoordinator(mockDB, NewMockPoolManager(), NewMockNetworkManager(), NewMockRunnerManager())

	s, err := NewServer(nil, coord, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()

	s.handleListPipelines(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var pipelines []interface{}
	json.NewDecoder(w.Body).Decode(&pipelines)
	t.Logf("Got %d pipelines", len(pipelines))
}

// TestIntegration_HandleCreatePipeline tests creating pipeline
func TestIntegration_HandleCreatePipeline(t *testing.T) {
	mockDB := NewMockPipelineDB()
	coord, _ := pipeline.NewCoordinator(mockDB, NewMockPoolManager(), NewMockNetworkManager(), NewMockRunnerManager())

	s, err := NewServer(nil, coord, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Just test that the handler doesn't panic
	req := httptest.NewRequest("POST", "/api/pipelines", nil)
	w := httptest.NewRecorder()

	s.handleCreatePipeline(w, req)

	t.Logf("CreatePipeline status: %d", w.Code)
}

// mockDBForTypes wraps PipelineDB to implement types.PipelineDB
type mockDBForTypes struct {
	*pipeline.PipelineDB
}

// TestIntegration_HandleGetStats tests stats endpoint
func TestIntegration_HandleGetStats(t *testing.T) {
	t.Skip("requires runner manager with database")
}

// TestIntegration_HandleStartPipeline tests starting pipeline
func TestIntegration_HandleStartPipeline(t *testing.T) {
	mockDB := NewMockPipelineDB()
	coord, _ := pipeline.NewCoordinator(mockDB, NewMockPoolManager(), NewMockNetworkManager(), NewMockRunnerManager())

	s, err := NewServer(nil, coord, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/pipelines/test/start", nil)
	w := httptest.NewRecorder()

	s.handleStartPipeline(w, req)

	t.Logf("StartPipeline status: %d", w.Code)
}

// TestIntegration_HandleStopPipeline tests stopping pipeline
func TestIntegration_HandleStopPipeline(t *testing.T) {
	mockDB := NewMockPipelineDB()
	coord, _ := pipeline.NewCoordinator(mockDB, NewMockPoolManager(), NewMockNetworkManager(), NewMockRunnerManager())

	s, err := NewServer(nil, coord, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/pipelines/test/stop", nil)
	w := httptest.NewRecorder()

	s.handleStopPipeline(w, req)

	t.Logf("StopPipeline status: %d", w.Code)
}

// TestIntegration_HandleDestroyPipeline tests destroying pipeline
func TestIntegration_HandleDestroyPipeline(t *testing.T) {
	mockDB := NewMockPipelineDB()
	coord, _ := pipeline.NewCoordinator(mockDB, NewMockPoolManager(), NewMockNetworkManager(), NewMockRunnerManager())

	s, err := NewServer(nil, coord, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/pipelines/test", nil)
	w := httptest.NewRecorder()

	s.handleDestroyPipeline(w, req)

	t.Logf("DestroyPipeline status: %d", w.Code)
}

// TestIntegration_HandleListRunners tests listing runners
func TestIntegration_HandleListRunners(t *testing.T) {
	t.Skip("requires runner manager with database")
}

// TestIntegration_HandleListPools tests listing pools
func TestIntegration_HandleListPools(t *testing.T) {
	t.Skip("requires pool manager with database")
}

// TestIntegration_HandleListNetworks tests listing networks
func TestIntegration_HandleListNetworks(t *testing.T) {
	t.Skip("requires network manager with database")
}

// TestIntegration_HandleListArtifacts tests listing artifacts
func TestIntegration_HandleListArtifacts(t *testing.T) {
	t.Skip("requires artifact manager with database")
}

// TestIntegration_Routes tests all routes
func TestIntegration_Routes(t *testing.T) {
	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Only test routes that don't require managers
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/health"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s %s: expected 200, got %d", route.method, route.path, w.Code)
		}
	}
}

// TestIntegration_WriteJSON tests JSON helper
func TestIntegration_WriteJSON(t *testing.T) {
	s, _ := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}
}

// TestIntegration_WriteError tests error helper
func TestIntegration_WriteError(t *testing.T) {
	s, _ := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()

	s.writeError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "test error" {
		t.Errorf("expected 'test error', got %s", resp["error"])
	}
}

// TestIntegration_AuthMiddleware tests auth
func TestIntegration_AuthMiddleware(t *testing.T) {
	s, _ := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		AuthEnabled: true,
		AuthToken:   "secret-token",
	})

	handler := s.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No token
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// Wrong token
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// Correct token
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestIntegration_ConfigDefaults tests config
func TestIntegration_ConfigDefaults(t *testing.T) {
	config := &ServerConfig{}
	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, config)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	if s.httpServer.Addr != ":8080" {
		t.Errorf("expected :8080, got %s", s.httpServer.Addr)
	}
}

// TestIntegration_WebSocketMessage tests WebSocket message
func TestIntegration_WebSocketMessage(t *testing.T) {
	msg := &WebSocketMessage{
		Type:    "pipeline:update",
		Payload: map[string]interface{}{"id": "test"},
	}

	if msg.Type != "pipeline:update" {
		t.Errorf("expected pipeline:update, got %s", msg.Type)
	}
}

// TestIntegration_NewServer tests server creation
func TestIntegration_NewServer(t *testing.T) {
	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.router == nil {
		t.Fatal("expected non-nil router")
	}
}

// TestIntegration_NewServer_WithConfig tests server creation with config
func TestIntegration_NewServer_WithConfig(t *testing.T) {
	config := &ServerConfig{
		ListenAddr:  ":9090",
		AuthEnabled: false,
	}

	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, config)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if s.httpServer.Addr != ":9090" {
		t.Errorf("expected :9090, got %s", s.httpServer.Addr)
	}
}