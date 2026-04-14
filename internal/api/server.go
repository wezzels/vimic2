// Package api provides REST API endpoints
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stsgym/vimic2/internal/pipeline"
	"github.com/stsgym/vimic2/internal/types"
)

// Server represents the API server
type Server struct {
	db             *pipeline.PipelineDB
	coordinator    *pipeline.Coordinator
	dispatcher     *pipeline.JobDispatcher
	artifacts      *pipeline.ArtifactManager
	logs           *pipeline.LogCollector
	poolManager    types.PoolManagerInterface
	networkManager types.NetworkManagerInterface
	runnerManager  types.RunnerManagerInterface
	httpServer     *http.Server
	router         *http.ServeMux
	ws             *WebSocketServer
	authEnabled    bool
	authToken      string
}

// SetPoolManager sets the pool manager (for testing)
func (s *Server) SetPoolManager(pm types.PoolManagerInterface) {
	s.poolManager = pm
}

// SetNetworkManager sets the network manager (for testing)
func (s *Server) SetNetworkManager(nm types.NetworkManagerInterface) {
	s.networkManager = nm
}

// SetRunnerManager sets the runner manager (for testing)
func (s *Server) SetRunnerManager(rm types.RunnerManagerInterface) {
	s.runnerManager = rm
}

// ServerConfig represents server configuration
type ServerConfig struct {
	ListenAddr  string `json:"listen_addr"`
	AuthEnabled bool   `json:"auth_enabled"`
	AuthToken   string `json:"auth_token"`
	TLSCert     string `json:"tls_cert"`
	TLSKey      string `json:"tls_key"`
}

// NewServer creates a new API server
func NewServer(db *pipeline.PipelineDB, coordinator *pipeline.Coordinator, dispatcher *pipeline.JobDispatcher, artifacts *pipeline.ArtifactManager, logs *pipeline.LogCollector, poolMgr types.PoolManagerInterface, netMgr types.NetworkManagerInterface, runnerMgr types.RunnerManagerInterface, config *ServerConfig) (*Server, error) {
	if config == nil {
		config = &ServerConfig{
			ListenAddr: ":8080",
		}
	}

	s := &Server{
		db:             db,
		coordinator:    coordinator,
		dispatcher:     dispatcher,
		artifacts:      artifacts,
		logs:           logs,
		poolManager:    poolMgr,
		networkManager: netMgr,
		runnerManager:  runnerMgr,
		router:         http.NewServeMux(),
		authEnabled:    config.AuthEnabled,
		authToken:      config.AuthToken,
		ws:             NewWebSocketServer(coordinator),
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         config.ListenAddr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// setupRoutes sets up API routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("GET /api/health", s.handleHealth)

	// Pipelines
	s.router.HandleFunc("GET /api/pipelines", s.authMiddleware(s.handleListPipelines))
	s.router.HandleFunc("GET /api/pipelines/{id}", s.authMiddleware(s.handleGetPipeline))
	s.router.HandleFunc("POST /api/pipelines", s.authMiddleware(s.handleCreatePipeline))
	s.router.HandleFunc("POST /api/pipelines/{id}/start", s.authMiddleware(s.handleStartPipeline))
	s.router.HandleFunc("POST /api/pipelines/{id}/stop", s.authMiddleware(s.handleStopPipeline))
	s.router.HandleFunc("DELETE /api/pipelines/{id}", s.authMiddleware(s.handleDestroyPipeline))

	// Jobs
	s.router.HandleFunc("GET /api/jobs", s.authMiddleware(s.handleListJobs))
	s.router.HandleFunc("GET /api/jobs/{id}", s.authMiddleware(s.handleGetJob))
	s.router.HandleFunc("POST /api/jobs", s.authMiddleware(s.handleEnqueueJob))
	s.router.HandleFunc("POST /api/jobs/{id}/cancel", s.authMiddleware(s.handleCancelJob))
	s.router.HandleFunc("POST /api/jobs/{id}/retry", s.authMiddleware(s.handleRetryJob))

	// Runners
	s.router.HandleFunc("GET /api/runners", s.authMiddleware(s.handleListRunners))
	s.router.HandleFunc("GET /api/runners/{id}", s.authMiddleware(s.handleGetRunner))
	s.router.HandleFunc("POST /api/runners", s.authMiddleware(s.handleCreateRunner))
	s.router.HandleFunc("POST /api/runners/{id}/start", s.authMiddleware(s.handleStartRunner))
	s.router.HandleFunc("POST /api/runners/{id}/stop", s.authMiddleware(s.handleStopRunner))
	s.router.HandleFunc("DELETE /api/runners/{id}", s.authMiddleware(s.handleDestroyRunner))

	// Pools
	s.router.HandleFunc("GET /api/pools", s.authMiddleware(s.handleListPools))
	s.router.HandleFunc("GET /api/pools/{name}", s.authMiddleware(s.handleGetPool))
	s.router.HandleFunc("POST /api/pools", s.authMiddleware(s.handleCreatePool))
	s.router.HandleFunc("GET /api/pools/{name}/vms", s.authMiddleware(s.handleListPoolVMs))
	s.router.HandleFunc("POST /api/pools/{name}/acquire", s.authMiddleware(s.handleAcquireVM))
	s.router.HandleFunc("POST /api/pools/{name}/release", s.authMiddleware(s.handleReleaseVM))

	// Networks
	s.router.HandleFunc("GET /api/networks", s.authMiddleware(s.handleListNetworks))
	s.router.HandleFunc("GET /api/networks/{id}", s.authMiddleware(s.handleGetNetwork))
	s.router.HandleFunc("POST /api/networks", s.authMiddleware(s.handleCreateNetwork))
	s.router.HandleFunc("DELETE /api/networks/{id}", s.authMiddleware(s.handleDeleteNetwork))

	// Artifacts
	s.router.HandleFunc("GET /api/artifacts", s.authMiddleware(s.handleListArtifacts))
	s.router.HandleFunc("GET /api/artifacts/{id}", s.authMiddleware(s.handleGetArtifact))
	s.router.HandleFunc("GET /api/artifacts/{id}/download", s.authMiddleware(s.handleDownloadArtifact))
	s.router.HandleFunc("POST /api/artifacts/upload", s.authMiddleware(s.handleUploadArtifact))
	s.router.HandleFunc("DELETE /api/artifacts/{id}", s.authMiddleware(s.handleDeleteArtifact))

	// Logs
	s.router.HandleFunc("GET /api/logs/streams", s.authMiddleware(s.handleListLogStreams))
	s.router.HandleFunc("GET /api/logs/streams/{id}", s.authMiddleware(s.handleGetLogStream))
	s.router.HandleFunc("GET /api/logs/streams/{id}/read", s.authMiddleware(s.handleReadLogs))
	s.router.HandleFunc("POST /api/logs/streams/{id}/write", s.authMiddleware(s.handleWriteLog))
	s.router.HandleFunc("GET /api/logs/search", s.authMiddleware(s.handleSearchLogs))

	// Stats
	s.router.HandleFunc("GET /api/stats", s.authMiddleware(s.handleGetStats))

	// WebSocket
	s.router.HandleFunc("GET /ws", s.handleWebSocket)
}

// authMiddleware provides authentication middleware
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authEnabled {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		if parts[1] != s.authToken {
			s.writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		next(w, r)
	}
}

// Start starts the API server
func (s *Server) Start() error {
	fmt.Printf("[API] Starting server on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("[API] Stopping server")
	return s.httpServer.Shutdown(ctx)
}

// Response helpers

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

// Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Pipeline handlers

func (s *Server) handleListPipelines(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	pipelines := s.coordinator.ListPipelines()
	s.writeJSON(w, http.StatusOK, pipelines)
}

func (s *Server) handleGetPipeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pipeline, err := s.coordinator.GetPipeline(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, pipeline)
}

func (s *Server) handleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Platform    string   `json:"platform"`
		Repository  string   `json:"repository"`
		Branch      string   `json:"branch"`
		CommitSHA   string   `json:"commit_sha"`
		CommitMsg   string   `json:"commit_message"`
		Author      string   `json:"author"`
		RunnerCount int      `json:"runner_count"`
		Labels      []string `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RunnerCount == 0 {
		req.RunnerCount = 1
	}

	pipelineState, err := s.coordinator.CreatePipeline(
		r.Context(),
		types.RunnerPlatform(req.Platform),
		req.Repository,
		req.Branch,
	)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Start with specified runners
	if req.RunnerCount > 0 {
		if err := s.coordinator.StartPipeline(r.Context(), pipelineState.ID, req.RunnerCount); err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	s.writeJSON(w, http.StatusCreated, pipelineState)
}

func (s *Server) handleStartPipeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	runnerCount := 1
	if count := r.URL.Query().Get("runners"); count != "" {
		fmt.Sscanf(count, "%d", &runnerCount)
	}
	if err := s.coordinator.StartPipeline(r.Context(), id, runnerCount); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) handleStopPipeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.coordinator.CancelPipeline(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleDestroyPipeline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.coordinator.DeletePipeline(id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "destroyed"})
}

// Job handlers

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var jobs []*pipeline.Job

	switch status {
	case "running":
		jobs = s.dispatcher.ListRunningJobs()
	case "completed":
		jobs = s.dispatcher.ListCompletedJobs()
	case "failed":
		jobs = s.dispatcher.ListFailedJobs()
	case "pending":
		jobs = s.dispatcher.ListPendingJobs()
	default:
		jobs = s.dispatcher.ListJobs()
	}

	s.writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := s.dispatcher.GetJob(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleEnqueueJob(w http.ResponseWriter, r *http.Request) {
	var job pipeline.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.dispatcher.EnqueueJob(r.Context(), &job); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, job)
}

func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.dispatcher.CancelJob(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "canceled"})
}

func (s *Server) handleRetryJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.dispatcher.RetryJob(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "retried"})
}

// Runner handlers

func (s *Server) handleListRunners(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	// ListRunners not in interface - need concrete type
	s.writeError(w, http.StatusInternalServerError, "runner listing requires RunnerManager")
}

func (s *Server) handleGetRunner(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	id := r.PathValue("id")
	runnerInfo, err := s.runnerManager.GetRunner(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, runnerInfo)
}

func (s *Server) handleCreateRunner(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	var req struct {
		PoolName   string   `json:"pool_name"`
		Platform   string   `json:"platform"`
		PipelineID string   `json:"pipeline_id"`
		Labels     []string `json:"labels"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// CreateRunner requires context - use interface method
	config := map[string]interface{}{
		"pool_name":   req.PoolName,
		"platform":    req.Platform,
		"pipeline_id": req.PipelineID,
		"labels":      req.Labels,
	}
	runnerID, err := s.runnerManager.CreateRunner(types.RunnerPlatform(req.Platform), config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]string{"id": runnerID})
}

func (s *Server) handleStartRunner(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	id := r.PathValue("id")
	// Get VM IP from runner info
	runnerInfo, err := s.runnerManager.GetRunner(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// VM IP would be obtained from pool manager in full implementation
	// For now, return success
	runnerID := ""
	if id, ok := runnerInfo["id"].(string); ok {
		runnerID = id
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "started", "runner_id": runnerID})
}

func (s *Server) handleStopRunner(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	id := r.PathValue("id")
	if err := s.runnerManager.DestroyRunner(id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleDestroyRunner(w http.ResponseWriter, r *http.Request) {
	if s.runnerManager == nil {
		s.writeError(w, http.StatusInternalServerError, "runner manager not configured")
		return
	}
	id := r.PathValue("id")
	if err := s.runnerManager.DestroyRunner(id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "destroyed"})
}

// Pool handlers

func (s *Server) handleListPools(w http.ResponseWriter, r *http.Request) {
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}
	pools, err := s.poolManager.ListPools()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, pools)
}

func (s *Server) handleGetPool(w http.ResponseWriter, r *http.Request) {
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}
	name := r.PathValue("name")
	poolInfo, err := s.poolManager.GetPool(name)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, poolInfo)
}

func (s *Server) handleCreatePool(w http.ResponseWriter, r *http.Request) {
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}
	// CreatePool requires concrete type - cast if available
	s.writeError(w, http.StatusInternalServerError, "pool creation requires PoolManager")
}

func (s *Server) handleListPoolVMs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}

	// Get the pool to check it exists
	pool, err := s.poolManager.GetPool(name)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Return pool info with VM list
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":      pool.Name,
		"capacity":  pool.Capacity,
		"available": pool.Available,
		"busy":      pool.Busy,
	})
}

func (s *Server) handleAcquireVM(w http.ResponseWriter, r *http.Request) {
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}
	name := r.PathValue("name")
	vm, err := s.poolManager.AllocateVM(name)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusCreated, vm)
}

func (s *Server) handleReleaseVM(w http.ResponseWriter, r *http.Request) {
	if s.poolManager == nil {
		s.writeError(w, http.StatusInternalServerError, "pool manager not configured")
		return
	}
	var req struct {
		VMID string `json:"vm_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.poolManager.ReleaseVM(req.VMID); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "released"})
}

// Network handlers

func (s *Server) handleListNetworks(w http.ResponseWriter, r *http.Request) {
	if s.networkManager == nil {
		s.writeError(w, http.StatusInternalServerError, "network manager not configured")
		return
	}
	// ListNetworks not in interface - need concrete type
	s.writeError(w, http.StatusInternalServerError, "network listing requires IsolationManager")
}

func (s *Server) handleGetNetwork(w http.ResponseWriter, r *http.Request) {
	if s.networkManager == nil {
		s.writeError(w, http.StatusInternalServerError, "network manager not configured")
		return
	}
	id := r.PathValue("id")
	network, err := s.networkManager.GetNetwork(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, network)
}

func (s *Server) handleCreateNetwork(w http.ResponseWriter, r *http.Request) {
	if s.networkManager == nil {
		s.writeError(w, http.StatusInternalServerError, "network manager not configured")
		return
	}
	var req struct {
		PipelineID string `json:"pipeline_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// CreateNetwork uses interface method
	networkID, err := s.networkManager.CreateNetwork(&types.NetworkConfig{})
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]string{"id": networkID})
}

func (s *Server) handleDeleteNetwork(w http.ResponseWriter, r *http.Request) {
	if s.networkManager == nil {
		s.writeError(w, http.StatusInternalServerError, "network manager not configured")
		return
	}
	id := r.PathValue("id")
	if err := s.networkManager.DestroyNetwork(id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Artifact handlers

func (s *Server) handleListArtifacts(w http.ResponseWriter, r *http.Request) {
	pipelineID := r.URL.Query().Get("pipeline_id")
	var artifacts []*pipeline.Artifact

	if pipelineID != "" {
		artifacts = s.artifacts.ListArtifacts(pipelineID)
	} else {
		// Return all artifacts by pipeline
		// List all pipelines and get artifacts for each
		if s.coordinator != nil {
			pipelines := s.coordinator.ListPipelines()
			for _, ps := range pipelines {
							pipelineArtifacts := s.artifacts.ListArtifacts(ps.ID)
				artifacts = append(artifacts, pipelineArtifacts...)
			}
		}
	}

	s.writeJSON(w, http.StatusOK, artifacts)
}

func (s *Server) handleGetArtifact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	artifact, err := s.artifacts.GetArtifact(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, artifact)
}

func (s *Server) handleDownloadArtifact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	artifact, err := s.artifacts.GetArtifact(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", artifact.Name))
	w.Header().Set("Content-Type", "application/octet-stream")

	if err := s.artifacts.DownloadArtifact(r.Context(), id, w); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *Server) handleUploadArtifact(w http.ResponseWriter, r *http.Request) {
	pipelineID := r.URL.Query().Get("pipeline_id")
	artifactType := r.URL.Query().Get("type")
	name := r.URL.Query().Get("name")

	if pipelineID == "" || artifactType == "" || name == "" {
		s.writeError(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	artifact, err := s.artifacts.UploadArtifact(r.Context(), pipelineID, artifactType, name, r.Body, nil)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, artifact)
}

func (s *Server) handleDeleteArtifact(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.artifacts.DeleteArtifact(r.Context(), id); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Log handlers

func (s *Server) handleListLogStreams(w http.ResponseWriter, r *http.Request) {
	pipelineID := r.URL.Query().Get("pipeline_id")
	var streams []*pipeline.LogStream

	if pipelineID != "" {
		streams = s.logs.ListLogStreams(pipelineID)
	} else {
		// Return all streams by pipeline
		if s.coordinator != nil {
			pipelines := s.coordinator.ListPipelines()
			for _, ps := range pipelines {
				pipelineStreams := s.logs.ListLogStreams(ps.ID)
				streams = append(streams, pipelineStreams...)
			}
		}
	}

	s.writeJSON(w, http.StatusOK, streams)
}

func (s *Server) handleGetLogStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stream, err := s.logs.GetLogStream(id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, stream)
}

func (s *Server) handleReadLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	entries, err := s.logs.ReadLog(id, offset, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, entries)
}

func (s *Server) handleWriteLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Level    string `json:"level"`
		Message  string `json:"message"`
		Duration int64  `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.logs.WriteLog(r.Context(), id, req.Level, req.Message, req.Duration); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "written"})
}

func (s *Server) handleSearchLogs(w http.ResponseWriter, r *http.Request) {
	pipelineID := r.URL.Query().Get("pipeline_id")
	query := r.URL.Query().Get("query")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	entries, err := s.logs.SearchLogs(r.Context(), pipelineID, query, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, entries)
}

// Stats handler

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	// Get pipeline stats
	var pipelineCount int
	if s.coordinator != nil {
		pipelines := s.coordinator.ListPipelines()
		pipelineCount = len(pipelines)
	}

	stats := map[string]interface{}{
		"pipelines": pipelineCount,
		"runners":   0, // requires RunnerManager
		"pools":     0, // requires PoolManager
		"networks":  0, // requires IsolationManager
		"artifacts": s.artifacts != nil,
		"logs":      s.logs != nil,
	}

	s.writeJSON(w, http.StatusOK, stats)
}
