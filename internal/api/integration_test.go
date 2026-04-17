//go:build integration

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
)

func setupAPITest(t *testing.T) (*Server, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-api-test-")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	artifactConfig := &pipeline.ArtifactConfig{
		StoragePath:   filepath.Join(tmpDir, "artifacts"),
		RetentionDays: 30,
	}
	artifacts, err := pipeline.NewArtifactManager(db, artifactConfig)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	// Set state file to avoid leakage
	artifacts.SetStateFile(filepath.Join(tmpDir, "artifacts-state.json"))

	logConfig := &pipeline.LogConfig{
		StoragePath: filepath.Join(tmpDir, "logs"),
		BufferSize:  100,
	}
	logs, err := pipeline.NewLogCollector(db, logConfig)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	logs.SetStateFile(filepath.Join(tmpDir, "logs-state.json"))

	server, err := NewServer(db, nil, nil, artifacts, logs, nil, nil, nil, &ServerConfig{
		ListenAddr: ":0",
	})
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	return server, func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
}

func TestAPI_Health(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("Health response = %v, want healthy", resp)
	}
}

func TestAPI_ListPipelines(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pipelines", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListPipelines status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_ListArtifacts(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/artifacts", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListArtifacts status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_GetStats(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetStats status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["artifacts"] == nil {
		t.Error("GetStats should include artifacts")
	}
}

func TestAPI_UploadArtifact(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	content := []byte("test artifact content")
	req := httptest.NewRequest(http.MethodPost, "/api/artifacts/upload?pipeline_id=test-pipeline&type=log&name=test.log", bytes.NewReader(content))
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("UploadArtifact status = %d, want %d or %d", w.Code, http.StatusOK, http.StatusCreated)
	}
}

func TestAPI_ListLogStreams(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/logs/streams?pipeline_id=test-pipeline", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// May return 200 (empty list) or other codes
	t.Logf("ListLogStreams status: %d", w.Code)
}

func TestAPI_NotFound(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("NotFound status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAPI_WriteJSON_ContentType(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", w.Header().Get("Content-Type"))
	}
}

func TestAPI_Auth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vimic2-api-auth-test-")
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

	artifactConfig := &pipeline.ArtifactConfig{
		StoragePath:   filepath.Join(tmpDir, "artifacts"),
		RetentionDays: 30,
	}
	artifacts, _ := pipeline.NewArtifactManager(db, artifactConfig)

	logConfig := &pipeline.LogConfig{
		StoragePath: filepath.Join(tmpDir, "logs"),
		BufferSize:  100,
	}
	logs, _ := pipeline.NewLogCollector(db, logConfig)

	server, err := NewServer(db, nil, nil, artifacts, logs, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: true,
		AuthToken:   "test-token",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Without auth - /api/health doesn't require auth
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Health endpoint doesn't require auth, so it returns 200
	if w.Code != http.StatusOK {
		t.Errorf("Health without auth status = %d, want %d", w.Code, http.StatusOK)
	}

	// With auth - test a protected endpoint
	req = httptest.NewRequest(http.MethodGet, "/api/pipelines", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Pipelines without auth status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// With auth
	req = httptest.NewRequest(http.MethodGet, "/api/pipelines", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Pipelines with auth status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPI_CreateNetwork(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	networkData := map[string]interface{}{
		"name":    "test-network",
		"type":    "nat",
		"cidr":    "192.168.100.0/24",
		"gateway": "192.168.100.1",
	}
	body, _ := json.Marshal(networkData)

	req := httptest.NewRequest(http.MethodPost, "/api/networks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Without network manager, should return error
	t.Logf("CreateNetwork status: %d", w.Code)
}

func TestAPI_ListPools(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/pools", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	t.Logf("ListPools status: %d", w.Code)
}

func TestAPI_ListRunners(t *testing.T) {
	server, cleanup := setupAPITest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/runners", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	t.Logf("ListRunners status: %d", w.Code)
}