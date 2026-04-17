//go:build integration

package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/types"
)

// ==================== Mock Interface Implementations ====================

type mockPoolManager struct {
	pools []*types.PoolState
	vms   []*types.VMState
	err   error
}

func (m *mockPoolManager) AllocateVM(poolName string) (*types.VMState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &types.VMState{ID: "vm-1", PoolName: poolName, Status: "running"}, nil
}

func (m *mockPoolManager) ReleaseVM(vmID string) error {
	return m.err
}

func (m *mockPoolManager) GetPool(name string) (*types.PoolState, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, p := range m.pools {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, os.ErrNotExist
}

func (m *mockPoolManager) ListPools() ([]*types.PoolState, error) {
	return m.pools, m.err
}

type mockNetworkManager struct {
	err error
}

func (m *mockNetworkManager) CreateNetwork(config *types.NetworkConfig) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "net-1", nil
}

func (m *mockNetworkManager) DestroyNetwork(networkID string) error {
	return m.err
}

func (m *mockNetworkManager) GetNetwork(networkID string) (*types.NetworkConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &types.NetworkConfig{CIDR: networkID}, nil
}

type mockRunnerManager struct {
	err error
}

func (m *mockRunnerManager) CreateRunner(platform types.RunnerPlatform, config map[string]interface{}) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "runner-1", nil
}

func (m *mockRunnerManager) DestroyRunner(runnerID string) error {
	return m.err
}

func (m *mockRunnerManager) GetRunner(runnerID string) (map[string]interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return map[string]interface{}{"id": runnerID, "status": "running"}, nil
}

// ==================== Helper ====================

func setupAPIServerWithMocks(t *testing.T) (*Server, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-api-mock-test-")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	s, err := NewServer(db, nil, nil, nil, nil, &mockPoolManager{}, &mockNetworkManager{}, &mockRunnerManager{}, &ServerConfig{})
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	return s, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== Pool Handler Tests ====================

func TestAPI_ListPools_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pools", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListPools: expected 200, got %d", w.Code)
	}
}

func TestAPI_GetPool_NotFound_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pools/nonexistent", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("GetPool response: %d", w.Code)
}

func TestAPI_AcquireVM_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	body := `{"pool": "default"}`
	req := httptest.NewRequest(http.MethodPost, "/api/pools/default/vms/acquire", strings.NewReader(body))
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("AcquireVM response: %d", w.Code)
}

func TestAPI_ReleaseVM_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	body := `{"vm_id": "vm-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/pools/vms/release", strings.NewReader(body))
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ReleaseVM response: %d", w.Code)
}

// ==================== Network Handler Tests ====================

func TestAPI_ListNetworks_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ListNetworks response: %d", w.Code)
}

func TestAPI_GetNetwork_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/networks/net-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("GetNetwork response: %d", w.Code)
}

func TestAPI_CreateNetwork_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	body := `{"name": "test-net", "cidr": "10.0.0.0/24"}`
	req := httptest.NewRequest(http.MethodPost, "/api/networks", strings.NewReader(body))
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("CreateNetwork response: %d", w.Code)
}

func TestAPI_DeleteNetwork_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/networks/net-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("DeleteNetwork response: %d", w.Code)
}

// ==================== Runner Handler Tests ====================

func TestAPI_CreateRunner_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	body := `{"platform": "local", "config": {"name": "test-runner"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/runners", strings.NewReader(body))
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("CreateRunner response: %d", w.Code)
}

func TestAPI_GetRunner_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/runners/runner-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("GetRunner response: %d", w.Code)
}

func TestAPI_DestroyRunner_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/runners/runner-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("DestroyRunner response: %d", w.Code)
}

// ==================== Stats Handler ====================

func TestAPI_GetStats_Mock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("GetStats response: %d", w.Code)
}

// ==================== Log Handler Tests ====================

func TestAPI_ListLogStreams_Mock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pipelines/pipe-1/logs", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ListLogStreams response: %d", w.Code)
}

func TestAPI_GetLogStream(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/logs/stream-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("GetLogStream response: %d", w.Code)
}

func TestAPI_ReadLogs(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/logs/log-1/entries?offset=0&limit=100", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ReadLogs response: %d", w.Code)
}

func TestAPI_WriteLog(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	body := `{"level": "info", "message": "test log", "duration": 100}`
	req := httptest.NewRequest(http.MethodPost, "/api/logs/log-1/entries", strings.NewReader(body))
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("WriteLog response: %d", w.Code)
}

func TestAPI_SearchLogs(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/logs/search?q=test", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("SearchLogs response: %d", w.Code)
}

// ==================== Artifact Handler Tests ====================

func TestAPI_ListArtifacts_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pipelines/pipe-1/artifacts", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ListArtifacts response: %d", w.Code)
}

func TestAPI_UploadArtifact_Mock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/pipelines/pipe-1/artifacts", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("UploadArtifact response: %d", w.Code)
}

func TestAPI_DownloadArtifact_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/artifacts/art-1/download", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("DownloadArtifact response: %d", w.Code)
}

func TestAPI_DeleteArtifact_WithMock(t *testing.T) {
	s, cleanup := setupAPIServerWithMocks(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/artifacts/art-1", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("DeleteArtifact response: %d", w.Code)
}

// ==================== Pool Manager Error Tests ====================

func TestAPI_ListPools_Error(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-api-mock-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	s, err := NewServer(db, nil, nil, nil, nil, &mockPoolManager{err: os.ErrNotExist}, &mockNetworkManager{}, &mockRunnerManager{}, &ServerConfig{})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pools", nil)
	w := httptest.NewRecorder()
	w = serveHTTP(s, req)

	t.Logf("ListPools error response: %d", w.Code)
}

func jsonBody(s string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(s))
}