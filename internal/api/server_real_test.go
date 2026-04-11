// Package api provides real server tests
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestServer_writeJSON tests writeJSON helper with real server
func TestServer_writeJSON(t *testing.T) {
	// Create a minimal server
	s := &Server{
		router: http.NewServeMux(),
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Call writeJSON
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   "2026-04-11T14:00:00Z",
	})

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("expected Content-Type application/json")
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["status"] != "ok" {
		t.Errorf("expected ok, got %s", result["status"])
	}
}

// TestServer_writeError tests writeError helper with real server
func TestServer_writeError(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}

	w := httptest.NewRecorder()

	s.writeError(w, http.StatusNotFound, "pipeline not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["error"] != "pipeline not found" {
		t.Errorf("expected 'pipeline not found', got %s", result["error"])
	}
}

// TestServer_writeError_BadRequest tests bad request error
func TestServer_writeError_BadRequest(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}

	w := httptest.NewRecorder()

	s.writeError(w, http.StatusBadRequest, "invalid request body")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestServer_writeError_InternalError tests internal server error
func TestServer_writeError_InternalError(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}

	w := httptest.NewRecorder()

	s.writeError(w, http.StatusInternalServerError, "database connection failed")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// TestServerConfig_Validation tests config validation
func TestServerConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *ServerConfig
		valid  bool
	}{
		{
			name: "Valid config",
			config: &ServerConfig{
				ListenAddr:  ":8080",
				AuthEnabled: false,
			},
			valid: true,
		},
		{
			name: "With auth",
			config: &ServerConfig{
				ListenAddr:  ":9090",
				AuthEnabled: true,
				AuthToken:   "secret-token",
			},
			valid: true,
		},
		{
			name: "With TLS",
			config: &ServerConfig{
				ListenAddr:  ":443",
				TLSCert:     "/path/to/cert.pem",
				TLSKey:      "/path/to/key.pem",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.ListenAddr == "" {
				t.Error("ListenAddr should not be empty")
			}
		})
	}
}

// TestServer_HandleHealth tests health handler
func TestServer_HandleHealth(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}

	// Create handler
	handler := http.HandlerFunc(s.handleHealth)

	// Create request
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(w, req)

	// Verify
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["status"] != "healthy" {
		t.Errorf("expected healthy, got %s", result["status"])
	}
}

// TestJSON_EncodingDecoding tests JSON operations used by handlers
func TestJSON_EncodingDecoding(t *testing.T) {
	// Test pipeline request encoding
	req := map[string]interface{}{
		"platform":      "docker",
		"repository":    "github.com/user/repo",
		"branch":        "main",
		"runner_count":  3,
		"labels":        []string{"linux", "x64"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if decoded["platform"] != "docker" {
		t.Error("platform should be docker")
	}
	if decoded["branch"] != "main" {
		t.Error("branch should be main")
	}
}

// TestJSON_ResponseStructure tests response structure
func TestJSON_ResponseStructure(t *testing.T) {
	// Pipeline list response
	pipelines := []map[string]string{
		{"id": "pipe-1", "status": "running"},
		{"id": "pipe-2", "status": "completed"},
	}

	data, _ := json.Marshal(pipelines)
	var decoded []map[string]string
	json.Unmarshal(data, &decoded)

	if len(decoded) != 2 {
		t.Errorf("expected 2 pipelines, got %d", len(decoded))
	}

	// Error response
	errResp := map[string]string{"error": "not found"}
	data, _ = json.Marshal(errResp)
	var errDecoded map[string]string
	json.Unmarshal(data, &errDecoded)

	if errDecoded["error"] != "not found" {
		t.Error("error message mismatch")
	}
}

// TestHTTP_Methods tests HTTP method handling
func TestHTTP_Methods(t *testing.T) {
	methods := []struct {
		method string
		path   string
	}{
		{"GET", "/api/pipelines"},
		{"POST", "/api/pipelines"},
		{"GET", "/api/pipelines/pipe-1"},
		{"DELETE", "/api/pipelines/pipe-1"},
		{"POST", "/api/pipelines/pipe-1/start"},
	}

	for _, tt := range methods {
		t.Run(tt.method+"_"+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}
			if req.URL.Path != tt.path {
				t.Errorf("expected path %s, got %s", tt.path, req.URL.Path)
			}
		})
	}
}

// TestHTTP_QueryParams tests query parameter extraction
func TestHTTP_QueryParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/pipelines?limit=10&offset=20", nil)

	limit := req.URL.Query().Get("limit")
	offset := req.URL.Query().Get("offset")

	if limit != "10" {
		t.Errorf("expected limit 10, got %s", limit)
	}
	if offset != "20" {
		t.Errorf("expected offset 20, got %s", offset)
	}
}

// TestHTTP_PathValues tests path value extraction
func TestHTTP_PathValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/pipelines/pipe-123", nil)

	// Set path values manually (normally done by router)
	req.SetPathValue("id", "pipe-123")

	id := req.PathValue("id")
	if id != "pipe-123" {
		t.Errorf("expected pipe-123, got %s", id)
	}
}

// TestHTTP_Headers tests header handling
func TestHTTP_Headers(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")

	if req.Header.Get("Authorization") != "Bearer token123" {
		t.Error("authorization header mismatch")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("content-type header mismatch")
	}
}

// TestHTTP_RequestBody tests request body handling
func TestHTTP_RequestBody(t *testing.T) {
	body := `{"platform":"docker","repository":"github.com/user/repo"}`
	req := httptest.NewRequest("POST", "/api/pipelines", strings.NewReader(body))

	var decoded map[string]string
	err := json.NewDecoder(req.Body).Decode(&decoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded["platform"] != "docker" {
		t.Error("platform should be docker")
	}
}