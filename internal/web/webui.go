// Package web provides web UI
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/stsgym/vimic2/internal/types"
)

//go:embed templates/*
var templatesFS embed.FS

// WebUI represents the web UI
type WebUI struct {
	db             types.PipelineDB
	poolManager    types.PoolManagerInterface
	networkManager types.NetworkManagerInterface
	httpServer     *http.Server
	templates      *template.Template
}

// WebUIConfig represents web UI configuration
type WebUIConfig struct {
	ListenAddr string `json:"listen_addr"`
	BasePath   string `json:"base_path"`
}

// NewWebUI creates a new web UI
func NewWebUI(db types.PipelineDB, poolMgr types.PoolManagerInterface, netMgr types.NetworkManagerInterface, config *WebUIConfig) (*WebUI, error) {
	if config == nil {
		config = &WebUIConfig{
			ListenAddr: ":3000",
			BasePath:   "/",
		}
	}

	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	ui := &WebUI{
		db:             db,
		poolManager:    poolMgr,
		networkManager: netMgr,
		templates:      tmpl,
	}

	return ui, nil
}

// Start starts the web UI server
func (ui *WebUI) Start() error {
	// Setup routes
	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		return fmt.Errorf("failed to get static files: %w", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// API routes
	mux.HandleFunc("/api/health", ui.handleHealth)
	mux.HandleFunc("/api/stats", ui.handleStats)
	mux.HandleFunc("/api/pipelines", ui.handlePipelines)
	mux.HandleFunc("/api/pools", ui.handlePools)
	mux.HandleFunc("/api/networks", ui.handleNetworks)
	mux.HandleFunc("/api/runners", ui.handleRunners)

	// Dashboard
	mux.HandleFunc("/", ui.handleDashboard)

	// Create server
	ui.httpServer = &http.Server{
		Addr:         ui.getListenAddr(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return ui.httpServer.ListenAndServe()
}

// Stop stops the web UI server
func (ui *WebUI) Stop() error {
	if ui.httpServer != nil {
		return ui.httpServer.Close()
	}
	return nil
}

func (ui *WebUI) getListenAddr() string {
	return ":3000"
}

// API Handlers

func (ui *WebUI) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func (ui *WebUI) handleStats(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pipelines": 0,
		"runners":   0,
		"vms":       0,
		"networks":  0,
	})
}

func (ui *WebUI) handlePipelines(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]interface{}{})
}

func (ui *WebUI) handlePools(w http.ResponseWriter, r *http.Request) {
	pools, err := ui.poolManager.ListPools()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(pools)
}

func (ui *WebUI) handleNetworks(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]interface{}{})
}

func (ui *WebUI) handleRunners(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]interface{}{})
}

func (ui *WebUI) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ui.templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Title":   "Vimic2 Dashboard",
		"Version": "0.1.0",
	})
}