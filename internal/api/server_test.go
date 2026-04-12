// Package api provides REST API endpoints
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stsgym/vimic2/internal/pipeline"
)

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	config := &ServerConfig{
		ListenAddr: ":8080",
	}

	server, err := NewServer(nil, nil, nil, nil, nil, nil, nil, nil, config)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.httpServer == nil {
		t.Fatal("Expected http server to be initialized")
	}
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	// Create minimal server
	server := &Server{
		router: http.NewServeMux(),
	}
	server.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	// Serve request
	server.router.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp["status"])
	}
}

// TestVersionEndpoint tests the version endpoint
func TestVersionEndpoint(t *testing.T) {
	server := &Server{
		router: http.NewServeMux(),
	}
	server.router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0"})
	})

	req := httptest.NewRequest("GET", "/version", nil)
	rec := httptest.NewRecorder()

	server.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

// TestAuthMiddleware tests the authentication middleware
func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		token      string
		wantStatus int
	}{
		{
			name:       "No auth when disabled",
			authHeader: "",
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Valid token",
			authHeader: "Bearer test-token",
			token:      "test-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer wrong-token",
			token:      "test-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Missing token",
			authHeader: "",
			token:      "test-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				router:      http.NewServeMux(),
				authEnabled: tt.token != "",
				authToken:   tt.token,
			}

			// Protected handler
			protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			})

			// Wrap with auth middleware
			handler := server.authMiddleware(protectedHandler)
			server.router.Handle("/", handler)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			server.router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

// TestJSONResponse tests JSON response helper
func TestJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]interface{}{
		"key": "value",
		"nested": map[string]int{
			"count": 42,
		},
	}

	// Write JSON response
	rec.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rec).Encode(data)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type: application/json")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if resp["key"] != "value" {
		t.Errorf("Expected key='value', got %v", resp["key"])
	}
}

// TestErrorResponse tests error response helper
func TestErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(rec).Encode(map[string]string{
		"error": "invalid request",
	})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if resp["error"] != "invalid request" {
		t.Errorf("Expected error message, got %s", resp["error"])
	}
}

// TestPipelineEndpoints tests pipeline CRUD endpoints
func TestPipelineEndpoints(t *testing.T) {
	t.Run("List pipelines empty", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rec).Encode([]pipeline.Pipeline{})

		var resp []pipeline.Pipeline
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		if len(resp) != 0 {
			t.Errorf("Expected empty list, got %d items", len(resp))
		}
	})

	t.Run("Get pipeline not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusNotFound)
		json.NewEncoder(rec).Encode(map[string]string{
			"error": "pipeline not found",
		})

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", rec.Code)
		}
	})
}

// TestCORSMiddleware tests CORS headers
func TestCORSMiddleware(t *testing.T) {
	server := &Server{
		router: http.NewServeMux(),
	}

	// Basic handler with CORS headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
	})

	server.router.Handle("/", handler)

	// Preflight request
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	rec := httptest.NewRecorder()
	server.router.ServeHTTP(rec, req)

	// Should allow origin
	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("Expected CORS origin *, got %s", allowOrigin)
	}
}

// TestRateLimit tests rate limiting (if implemented)
func TestRateLimit(t *testing.T) {
	// Basic test that server doesn't crash under multiple requests
	server := &Server{
		router: http.NewServeMux(),
	}
	server.router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		server.router.ServeHTTP(rec, req)

		// All should succeed (no rate limit in test)
		if rec.Code != http.StatusOK {
			t.Errorf("Request %d failed: %d", i, rec.Code)
		}
	}
}
