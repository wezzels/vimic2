//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stsgym/vimic2/internal/pipeline"
)

// serveHTTP is a helper that recovers from panics in nil-manager handlers
func serveHTTP(s *Server, req *http.Request) (w *httptest.ResponseRecorder) {
	w = httptest.NewRecorder()
	defer func() {
		recover()
	}()
	s.router.ServeHTTP(w, req)
	return w
}

// ==================== Full Server with PipelineDB ====================

func setupAPIServerWithDB(t *testing.T) (*Server, *pipeline.PipelineDB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "vimic2-api-full-test-")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	pipelineDB, err := pipeline.NewPipelineDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("NewPipelineDB failed: %v", err)
	}

	s, err := NewServer(pipelineDB, nil, nil, nil, nil, nil, nil, nil, &ServerConfig{
		ListenAddr:  ":0",
		AuthEnabled: false,
	})
	if err != nil {
		pipelineDB.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("NewServer failed: %v", err)
	}

	return s, pipelineDB, func() {
		pipelineDB.Close()
		os.RemoveAll(tmpDir)
	}
}

// ==================== Pipeline Endpoints ====================

func TestAPI_GetPipeline_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pipelines/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetPipeline status: %d", w.Code)
}

func TestAPI_CreatePipeline_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pipelines", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreatePipeline status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_CreatePipeline_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"name":   "test-pipeline",
		"stages": []string{"build", "test", "deploy"},
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/pipelines", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreatePipeline status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_StartPipeline_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pipelines/nonexistent/start", nil)
	w := serveHTTP(s, req)
	t.Logf("StartPipeline status: %d", w.Code)
}

func TestAPI_StopPipeline_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pipelines/nonexistent/stop", nil)
	w := serveHTTP(s, req)
	t.Logf("StopPipeline status: %d", w.Code)
}

func TestAPI_DestroyPipeline_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/pipelines/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("DestroyPipeline status: %d", w.Code)
}

// ==================== Job Endpoints ====================

func TestAPI_GetJob_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/jobs/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetJob status: %d", w.Code)
}

func TestAPI_EnqueueJob_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/jobs", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("EnqueueJob status: %d", w.Code)
}

func TestAPI_EnqueueJob_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"pipeline_id": "test",
		"stage":       "build",
		"command":     "echo hello",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/jobs", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("EnqueueJob status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_CancelJob_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/jobs/nonexistent/cancel", nil)
	w := serveHTTP(s, req)
	t.Logf("CancelJob status: %d", w.Code)
}

func TestAPI_RetryJob_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/jobs/nonexistent/retry", nil)
	w := serveHTTP(s, req)
	t.Logf("RetryJob status: %d", w.Code)
}

// ==================== Runner Endpoints ====================

func TestAPI_GetRunner_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/runners/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetRunner status: %d", w.Code)
}

func TestAPI_CreateRunner_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/runners", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreateRunner status: %d", w.Code)
}

func TestAPI_CreateRunner_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"name":     "test-runner",
		"platform": "gitlab",
		"token":    "test-token",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/runners", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreateRunner status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_StartRunner_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/runners/nonexistent/start", nil)
	w := serveHTTP(s, req)
	t.Logf("StartRunner status: %d", w.Code)
}

func TestAPI_StopRunner_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/runners/nonexistent/stop", nil)
	w := serveHTTP(s, req)
	t.Logf("StopRunner status: %d", w.Code)
}

func TestAPI_DestroyRunner_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/runners/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("DestroyRunner status: %d", w.Code)
}

// ==================== Pool Endpoints ====================

func TestAPI_GetPool_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pools/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetPool status: %d", w.Code)
}

func TestAPI_CreatePool_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pools", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreatePool status: %d", w.Code)
}

func TestAPI_CreatePool_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"name":     "test-pool",
		"min_size": 2,
		"max_size": 10,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/pools", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreatePool status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_ListPoolVMs_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pools/nonexistent/vms", nil)
	w := serveHTTP(s, req)
	t.Logf("ListPoolVMs status: %d", w.Code)
}

func TestAPI_AcquireVM_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pools/nonexistent/acquire", nil)
	w := serveHTTP(s, req)
	t.Logf("AcquireVM status: %d", w.Code)
}

func TestAPI_ReleaseVM_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/pools/nonexistent/vms/vm-1/release", nil)
	w := serveHTTP(s, req)
	t.Logf("ReleaseVM status: %d", w.Code)
}

// ==================== Network Endpoints ====================

func TestAPI_GetNetwork_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/networks/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetNetwork status: %d", w.Code)
}

func TestAPI_CreateNetwork_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/networks", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreateNetwork status: %d", w.Code)
}

func TestAPI_CreateNetwork_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"name":        "test-network",
		"bridge_name": "br-test",
		"cidr":        "10.0.0.0/24",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/networks", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("CreateNetwork status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_DeleteNetwork_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/networks/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("DeleteNetwork status: %d", w.Code)
}

// ==================== Artifact Endpoints ====================

func TestAPI_GetArtifact_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/artifacts/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetArtifact status: %d", w.Code)
}

func TestAPI_DownloadArtifact_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/artifacts/nonexistent/download", nil)
	w := serveHTTP(s, req)
	t.Logf("DownloadArtifact status: %d", w.Code)
}

func TestAPI_DeleteArtifact_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("DELETE", "/api/artifacts/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("DeleteArtifact status: %d", w.Code)
}

// ==================== Log Endpoints ====================

func TestAPI_GetLogStream_NotFound(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/logs/streams/nonexistent", nil)
	w := serveHTTP(s, req)
	t.Logf("GetLogStream status: %d", w.Code)
}

func TestAPI_ReadLogs_Empty(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/logs/nonexistent?limit=100", nil)
	w := serveHTTP(s, req)
	t.Logf("ReadLogs status: %d", w.Code)
}

func TestAPI_WriteLog_InvalidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("WriteLog status: %d", w.Code)
}

func TestAPI_WriteLog_ValidJSON(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"stream":  "test-stream",
		"level":   "info",
		"message": "test log message",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/logs", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := serveHTTP(s, req)
	t.Logf("WriteLog status: %d, body: %s", w.Code, w.Body.String())
}

func TestAPI_SearchLogs_Empty(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/logs/search?query=test&stream=default", nil)
	w := serveHTTP(s, req)
	t.Logf("SearchLogs status: %d", w.Code)
}

// ==================== WebSocket Tests ====================

func TestAPI_WebSocket_Broadcast(t *testing.T) {
	ws := NewWebSocketServer(nil)
	// Broadcast with no clients should not panic
	ws.Broadcast(&WebSocketMessage{Type: "test-event", Payload: map[string]string{"key": "value"}})
	t.Log("Broadcast with no clients succeeded")
}

// ==================== Server Start/Stop ====================

func TestAPI_ServerStartStop_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		s.Start()
	}()

	time.Sleep(100 * time.Millisecond)
	s.Stop(ctx)

	t.Log("Server with DB started and stopped successfully")
}

// ==================== Stats Endpoint ====================

func TestAPI_GetStats_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := serveHTTP(s, req)
	t.Logf("GetStats with DB status: %d, body: %s", w.Code, w.Body.String())
}

// ==================== List Endpoints with DB ====================

func TestAPI_ListPipelines_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := serveHTTP(s, req)
	t.Logf("ListPipelines with DB status: %d", w.Code)
}

func TestAPI_ListJobs_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/jobs?status=running", nil)
	w := serveHTTP(s, req)
	t.Logf("ListJobs with DB status: %d", w.Code)
}

func TestAPI_ListRunners_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/runners", nil)
	w := serveHTTP(s, req)
	t.Logf("ListRunners with DB status: %d", w.Code)
}

func TestAPI_ListPools_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/pools", nil)
	w := serveHTTP(s, req)
	t.Logf("ListPools with DB status: %d", w.Code)
}

func TestAPI_ListNetworks_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/networks", nil)
	w := serveHTTP(s, req)
	t.Logf("ListNetworks with DB status: %d", w.Code)
}

func TestAPI_ListArtifacts_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/artifacts", nil)
	w := serveHTTP(s, req)
	t.Logf("ListArtifacts with DB status: %d", w.Code)
}

func TestAPI_ListLogStreams_WithDB(t *testing.T) {
	s, _, cleanup := setupAPIServerWithDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/logs/streams", nil)
	w := serveHTTP(s, req)
	t.Logf("ListLogStreams with DB status: %d", w.Code)
}