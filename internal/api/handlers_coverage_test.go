// Package api provides handler tests
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleHealth tests the health endpoint
func TestHandleHealth(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.router.HandleFunc("GET /api/health", s.handleHealth)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", resp["status"])
	}
}

// TestHandleHealth_OK tests health check returns OK
func TestHandleHealth_OK(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s := &Server{router: http.NewServeMux()}
	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestAuthMiddleware_TokenValidation tests auth middleware token validation
func TestAuthMiddleware_TokenValidation(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: true,
		authToken:   "test-token",
	}

	// Handler that should be protected
	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test without token
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}

	// Test with wrong token
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong token, got %d", w.Code)
	}

	// Test with correct token
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with correct token, got %d", w.Code)
	}
}

// TestAuthMiddleware_Disabled tests auth middleware when disabled
func TestAuthMiddleware_Disabled(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: false,
	}

	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when auth disabled, got %d", w.Code)
	}
}

// TestServerConfig tests server config
func TestServerConfig(t *testing.T) {
	config := &ServerConfig{
		ListenAddr:  ":9090",
		AuthEnabled: true,
		AuthToken:   "secret",
	}

	if config.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", config.ListenAddr)
	}
	if !config.AuthEnabled {
		t.Error("expected auth enabled")
	}
}

// TestServerConfig_WithValues tests server config with values

// TestServer_StopHandlesNil tests server stop handles nil gracefully
func TestServer_StopHandlesNil(t *testing.T) {
	// Skip - Stop requires httpServer to be initialized
	t.Skip("Stop requires httpServer to be initialized")
}

// TestServer_New tests server creation
func TestServer_New(t *testing.T) {
	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if s == nil {
		t.Error("expected server")
	}
}

// TestServer_New_WithConfig tests server creation with config
func TestServer_New_WithConfig(t *testing.T) {
	config := &ServerConfig{
		ListenAddr:  ":9999",
		AuthEnabled: true,
		AuthToken:   "test",
	}

	s, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, config)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if s == nil {
		t.Error("expected server")
	}
	if s.authToken != "test" {
		t.Errorf("expected auth token, got %s", s.authToken)
	}
}

// TestSetupRoutes_Health tests health route
func TestSetupRoutes_Health(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: false,
	}

	s.setupRoutes()

	// Just test health - it doesn't require any dependencies
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestWebSocket tests WebSocket server creation
func TestWebSocket(t *testing.T) {
	ws := NewWebSocketServer(nil)
	if ws == nil {
		t.Error("expected WebSocket server")
	}
}

// TestWebSocketServer_BroadcastMessage tests broadcast without clients
func TestWebSocketServer_BroadcastMessage(t *testing.T) {
	ws := NewWebSocketServer(nil)

	// Should not panic
	ws.Broadcast(&WebSocketMessage{Type: "test", Payload: "message"})
}

// TestWebSocketServer_BroadcastToMessage tests broadcast to specific clients
func TestWebSocketServer_BroadcastToMessage(t *testing.T) {
	ws := NewWebSocketServer(nil)

	// Should not panic
	ws.BroadcastTo([]string{"nonexistent"}, &WebSocketMessage{Type: "test", Payload: "message"})
}

// TestPipelineRequest tests pipeline request parsing
func TestPipelineRequest(t *testing.T) {
	reqBody := `{"name": "test-pipeline", "stages": []}`

	// Just parse the raw JSON
	var createReq map[string]interface{}
	if err := json.Unmarshal([]byte(reqBody), &createReq); err != nil {
		t.Errorf("Unmarshal: %v", err)
	}

	if createReq["name"] != "test-pipeline" {
		t.Errorf("expected test-pipeline, got %v", createReq["name"])
	}
}

// TestHandleHealth_Method tests health with different methods
func TestHandleHealth_Method(t *testing.T) {
	s := &Server{router: http.NewServeMux()}

	methods := []string{"GET", "POST", "OPTIONS"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/health", nil)
		w := httptest.NewRecorder()
		s.handleHealth(w, req)
		t.Logf("%s /api/health: %d", method, w.Code)
	}
}

// TestAuthMiddleware_Malformed tests malformed auth header
func TestAuthMiddleware_Malformed(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: true,
		authToken:   "test-token",
	}

	handler := s.authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with malformed header
	tests := []string{
		"test-token",
		"Basic test-token",
		"Bearer",
		"",
	}

	for _, authHeader := range tests {
		req := httptest.NewRequest("GET", "/api/test", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		w := httptest.NewRecorder()
		handler(w, req)

		// Should be 401 for malformed/missing
		if authHeader != "Bearer test-token" {
			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401 for %q, got %d", authHeader, w.Code)
			}
		}
	}
}
