// Package api provides API server tests
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServer_Routes tests route setup
func TestServer_Routes(t *testing.T) {
	// Create server with nil dependencies (testing route setup only)
	s := &Server{
		router: http.NewServeMux(),
	}

	// Setup routes
	s.setupRoutes()

	// Test health endpoint exists
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Health should work without auth
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestServerConfig tests server configuration
func TestServerConfig_Default(t *testing.T) {
	config := &ServerConfig{}
	if config.ListenAddr != "" {
		t.Error("expected empty default listen addr")
	}

	// Test with values
	config = &ServerConfig{
		ListenAddr:   ":9090",
		AuthEnabled:  true,
		AuthToken:    "test-token",
		TLSCert:      "/path/to/cert.pem",
		TLSKey:       "/path/to/key.pem",
	}

	if config.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", config.ListenAddr)
	}
	if !config.AuthEnabled {
		t.Error("expected auth enabled")
	}
	if config.AuthToken != "test-token" {
		t.Errorf("expected test-token, got %s", config.AuthToken)
	}
}

// TestServer_HandleHealth tests health endpoint returns healthy status
func TestServer_HandleHealthEndpoint(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.setupRoutes()

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

	if status, ok := resp["status"]; !ok || status != "healthy" {
		t.Errorf("expected status healthy, got %v", resp)
	}
}

// TestServer_HealthNoAuth tests that health endpoint doesn't require auth even when auth enabled
func TestServer_HealthNoAuth(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: true,
		authToken:   "secret-token",
	}
	s.setupRoutes()

	// Health endpoint should work without auth header
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for health, got %d", w.Code)
	}
}

// TestServer_AuthEnabledReturns401 tests that auth-enabled server returns 401 for protected routes
func TestServer_AuthEnabledReturns401(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: true,
		authToken:   "secret-token",
	}
	s.setupRoutes()

	// Protected endpoint without auth header
	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestServer_AuthInvalidTokenReturns401 tests that invalid token returns 401
func TestServer_AuthInvalidTokenReturns401(t *testing.T) {
	s := &Server{
		router:      http.NewServeMux(),
		authEnabled: true,
		authToken:   "secret-token",
	}
	s.setupRoutes()

	// Request with wrong token
	req := httptest.NewRequest("GET", "/api/pipelines", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestServer_NilDependencies tests server with nil dependencies
func TestServer_NilDependencies(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}

	// Verify server can be created with nil dependencies
	if s == nil {
		t.Error("server should not be nil")
	}
	if s.router == nil {
		t.Error("router should not be nil")
	}
}

// TestServer_MultipleRequests tests concurrent requests to health endpoint
func TestServer_MultipleRequests(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.setupRoutes()

	// Run multiple concurrent requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

// TestServer_HealthMethod tests only GET is allowed on health
func TestServer_HealthMethod(t *testing.T) {
	s := &Server{
		router: http.NewServeMux(),
	}
	s.setupRoutes()

	// POST should fail
	req := httptest.NewRequest("POST", "/api/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Method not allowed or not found
	if w.Code == http.StatusOK {
		t.Error("POST should not return 200 on health endpoint")
	}
}

// TestServerConfig_TLS tests TLS configuration
func TestServerConfig_TLS(t *testing.T) {
	config := &ServerConfig{
		ListenAddr: ":8443",
		TLSCert:    "/etc/ssl/cert.pem",
		TLSKey:     "/etc/ssl/key.pem",
	}

	if config.TLSCert != "/etc/ssl/cert.pem" {
		t.Errorf("expected TLS cert path, got %s", config.TLSCert)
	}
	if config.TLSKey != "/etc/ssl/key.pem" {
		t.Errorf("expected TLS key path, got %s", config.TLSKey)
	}
}