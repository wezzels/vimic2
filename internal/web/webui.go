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
	runnerManager  *runner.RunnerManager
	httpServer     *http.Server
	templates      *template.Template
}

// WebUIConfig represents web UI configuration
type WebUIConfig struct {
	ListenAddr string `json:"listen_addr"`
	BasePath   string `json:"base_path"`
}

// NewWebUI creates a new web UI
func NewWebUI(db *pipeline.PipelineDB, coordinator *pipeline.Coordinator, dispatcher *pipeline.JobDispatcher, artifacts *pipeline.ArtifactManager, logs *pipeline.LogCollector, poolMgr *pool.PoolManager, netMgr *network.IsolationManager, runnerMgr *runner.RunnerManager, config *WebUIConfig) (*WebUI, error) {
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
		coordinator:    coordinator,
		dispatcher:     dispatcher,
		artifacts:      artifacts,
		logs:           logs,
		poolManager:    poolMgr,
		networkManager: netMgr,
		runnerManager:  runnerMgr,
		templates:      tmpl,
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Static files
	staticFS, _ := fs.Sub(templatesFS, "templates/static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Pages
	mux.HandleFunc("/", ui.handleIndex)
	mux.HandleFunc("/pipelines", ui.handlePipelines)
	mux.HandleFunc("/pipelines/", ui.handlePipelineDetail)
	mux.HandleFunc("/runners", ui.handleRunners)
	mux.HandleFunc("/pools", ui.handlePools)
	mux.HandleFunc("/networks", ui.handleNetworks)
	mux.HandleFunc("/artifacts", ui.handleArtifacts)
	mux.HandleFunc("/logs", ui.handleLogs)

	// API endpoints for HTMX
	mux.HandleFunc("/api/htmx/pipelines", ui.handleHtmxPipelines)
	mux.HandleFunc("/api/htmx/runners", ui.handleHtmxRunners)
	mux.HandleFunc("/api/htmx/pools", ui.handleHtmxPools)
	mux.HandleFunc("/api/htmx/stats", ui.handleHtmxStats)

	ui.httpServer = &http.Server{
		Addr:         config.ListenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return ui, nil
}

// Start starts the web UI
func (ui *WebUI) Start() error {
	fmt.Printf("[WebUI] Starting on %s\n", ui.httpServer.Addr)
	return ui.httpServer.ListenAndServe()
}

// Stop stops the web UI
func (ui *WebUI) Stop() error {
	return ui.httpServer.Close()
}

// Template data

type IndexData struct {
	Title        string
	PipelineStats map[string]int
	JobStats     map[string]int
	RunnerStats  map[string]int
	PoolStats    map[string]int
	NetworkStats map[string]int
}

type PipelinesData struct {
	Title     string
	Pipelines []*pipeline.PipelineState
}

type PipelineDetailData struct {
	Title    string
	Pipeline *pipeline.PipelineState
}

type RunnersData struct {
	Title   string
	Runners interface{}
}

type PoolsData struct {
	Title string
	Pools []*pool.Pool
}

type NetworksData struct {
	Title    string
	Networks []*network.IsolatedNetwork
}

// Handlers

func (ui *WebUI) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := IndexData{
		Title:         "Vimic2 Dashboard",
		PipelineStats: ui.coordinator.GetStats(),
		JobStats:      ui.dispatcher.GetStats(),
		RunnerStats:   ui.runnerManager.GetStats(),
		PoolStats:     ui.poolManager.GetVMStats(),
		NetworkStats:  ui.networkManager.GetNetworkStats(),
	}

	ui.renderTemplate(w, "index.html", data)
}

func (ui *WebUI) handlePipelines(w http.ResponseWriter, r *http.Request) {
	pipelines := ui.coordinator.ListPipelines(100, 0)

	data := PipelinesData{
		Title:     "Pipelines",
		Pipelines: pipelines,
	}

	ui.renderTemplate(w, "pipelines.html", data)
}

func (ui *WebUI) handlePipelineDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pipeline, err := ui.coordinator.GetPipeline(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data := PipelineDetailData{
		Title:    fmt.Sprintf("Pipeline %s", id[:8]),
		Pipeline: pipeline,
	}

	ui.renderTemplate(w, "pipeline-detail.html", data)
}

func (ui *WebUI) handleRunners(w http.ResponseWriter, r *http.Request) {
	runners := ui.runnerManager.ListRunners()

	data := RunnersData{
		Title:   "Runners",
		Runners: runners,
	}

	ui.renderTemplate(w, "runners.html", data)
}

func (ui *WebUI) handlePools(w http.ResponseWriter, r *http.Request) {
	pools := ui.poolManager.ListPools()

	data := PoolsData{
		Title: "Pools",
		Pools: pools,
	}

	ui.renderTemplate(w, "pools.html", data)
}

func (ui *WebUI) handleNetworks(w http.ResponseWriter, r *http.Request) {
	networks := ui.networkManager.ListNetworks()

	data := NetworksData{
		Title:    "Networks",
		Networks: networks,
	}

	ui.renderTemplate(w, "networks.html", data)
}

func (ui *WebUI) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement artifacts page
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (ui *WebUI) handleLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logs page
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HTMX handlers

func (ui *WebUI) handleHtmxPipelines(w http.ResponseWriter, r *http.Request) {
	pipelines := ui.coordinator.ListPipelines(50, 0)
	ui.renderTemplate(w, "htmx/pipelines.html", pipelines)
}

func (ui *WebUI) handleHtmxRunners(w http.ResponseWriter, r *http.Request) {
	runners := ui.runnerManager.ListRunners()
	ui.renderTemplate(w, "htmx/runners.html", runners)
}

func (ui *WebUI) handleHtmxPools(w http.ResponseWriter, r *http.Request) {
	pools := ui.poolManager.ListPools()
	ui.renderTemplate(w, "htmx/pools.html", pools)
}

func (ui *WebUI) handleHtmxStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"pipelines": ui.coordinator.GetStats(),
		"jobs":      ui.dispatcher.GetStats(),
		"runners":   ui.runnerManager.GetStats(),
		"pools":     ui.poolManager.GetVMStats(),
		"networks":  ui.networkManager.GetNetworkStats(),
		"artifacts": ui.artifacts.GetStats(),
		"logs":      ui.logs.GetStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Helper functions

func (ui *WebUI) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := ui.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}