// Package pipeline provides job dispatching
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/stsgym/vimic2/internal/runner"
	"github.com/stsgym/vimic2/internal/types"
)

// JobDispatcher dispatches jobs to runners
type JobDispatcher struct {
	db            *PipelineDB
	runnerManager *runner.RunnerManager
	queues        map[string]*JobQueue
	jobs          map[string]*Job
	pending       chan *Job
	running       map[string]*Job
	completed     map[string]*Job
	failed        map[string]*Job
	mu            sync.RWMutex
	stateFile     string
	workers       int
	stopChan      chan struct{}
}

// Job represents a job to be executed
type Job struct {
	ID           string          `json:"id"`
	PipelineID   string          `json:"pipeline_id"`
	Stage        string          `json:"stage"`
	Name         string          `json:"name"`
	Commands     []string        `json:"commands"`
	Environment  map[string]string `json:"environment"`
	Artifacts    []string        `json:"artifacts"`
	Cache        []string        `json:"cache"`
	Timeout      time.Duration    `json:"timeout"`
	Status       PipelineStatus  `json:"status"`
	RunnerID     string          `json:"runner_id"`
	StartTime    *time.Time      `json:"start_time,omitempty"`
	EndTime      *time.Time      `json:"end_time,omitempty"`
	Duration     int64           `json:"duration_seconds"`
	Retries      int             `json:"retries"`
	MaxRetries   int             `json:"max_retries"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// JobQueue represents a job queue
type JobQueue struct {
	Name      string  `json:"name"`
	Priority int     `json:"priority"`
	Jobs      []*Job  `json:"jobs"`
}

// DispatcherConfig represents dispatcher configuration
type DispatcherConfig struct {
	Workers     int    `json:"workers"`
	QueueSize   int    `json:"queue_size"`
	MaxRetries  int    `json:"max_retries"`
	JobTimeout time.Duration `json:"job_timeout"`
}

// NewJobDispatcher creates a new job dispatcher
func NewJobDispatcher(db *PipelineDB, runnerMgr *runner.RunnerManager, config *DispatcherConfig) (*JobDispatcher, error) {
	if config == nil {
		config = &DispatcherConfig{
			Workers:     10,
			QueueSize:   100,
			MaxRetries:  3,
			JobTimeout: 30 * time.Minute,
		}
	}

	d := &JobDispatcher{
		db:            db,
		runnerManager: runnerMgr,
		queues:        make(map[string]*JobQueue),
		jobs:          make(map[string]*Job),
		pending:       make(chan *Job, config.QueueSize),
		running:       make(map[string]*Job),
		completed:     make(map[string]*Job),
		failed:        make(map[string]*Job),
		workers:       config.Workers,
		stopChan:      make(chan struct{}),
	}

	// Load state
	if err := d.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Start workers
	for i := 0; i < d.workers; i++ {
		go d.worker(i)
	}

	return d, nil
}

// loadState loads job state from disk
func (d *JobDispatcher) loadState() error {
	stateFile := d.getStateFile()
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return err
	}

	var state struct {
		Jobs      []*Job `json:"jobs"`
		Pending   []*Job `json:"pending"`
		Running   []*Job `json:"running"`
		Completed []*Job `json:"completed"`
		Failed    []*Job `json:"failed"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	for _, job := range state.Jobs {
		d.jobs[job.ID] = job
	}
	for _, job := range state.Pending {
		d.jobs[job.ID] = job
	}
	for _, job := range state.Running {
		d.jobs[job.ID] = job
		d.running[job.ID] = job
	}
	for _, job := range state.Completed {
		d.jobs[job.ID] = job
		d.completed[job.ID] = job
	}
	for _, job := range state.Failed {
		d.jobs[job.ID] = job
		d.failed[job.ID] = job
	}

	return nil
}

// saveState saves job state to disk
func (d *JobDispatcher) saveState() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	state := struct {
		Jobs      []*Job `json:"jobs"`
		Pending   []*Job `json:"pending"`
		Running   []*Job `json:"running"`
		Completed []*Job `json:"completed"`
		Failed    []*Job `json:"failed"`
	}{
		Jobs: make([]*Job, 0, len(d.jobs)),
	}

	for _, job := range d.jobs {
		state.Jobs = append(state.Jobs, job)
	}
	for _, job := range d.running {
		state.Running = append(state.Running, job)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	stateFile := d.getStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(stateFile, data, 0644)
}

// getStateFile returns the state file path
func (d *JobDispatcher) getStateFile() string {
	if d.stateFile != "" {
		return d.stateFile
	}
	return filepath.Join(os.Getenv("HOME"), ".vimic2", "jobs-state.json")
}

// SetStateFile sets the state file path
func (d *JobDispatcher) SetStateFile(path string) {
	d.stateFile = path
}

// EnqueueJob adds a job to the queue
func (d *JobDispatcher) EnqueueJob(ctx context.Context, job *Job) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Validate job
	if job.ID == "" {
		job.ID = generateJobID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}
	if job.Status == "" {
		job.Status = PipelineStatusCreating
	}

	// Store job
	d.jobs[job.ID] = job

	// Add to queue
	queueName := d.getQueueName(job)
	if _, ok := d.queues[queueName]; !ok {
		d.queues[queueName] = &JobQueue{
			Name: queueName,
			Jobs: []*Job{},
		}
	}
	d.queues[queueName].Jobs = append(d.queues[queueName].Jobs, job)

	// Update status
	job.Status = PipelineStatusRunning
	job.UpdatedAt = time.Now()

	// Add to pending channel
	select {
	case d.pending <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// getQueueName returns the queue name for a job
func (d *JobDispatcher) getQueueName(job *Job) string {
	// Use stage name as queue name
	if job.Stage != "" {
		return job.Stage
	}
	return "default"
}

// worker processes jobs from the queue
func (d *JobDispatcher) worker(id int) {
	for {
		select {
		case <-d.stopChan:
			return
		case job := <-d.pending:
			d.processJob(id, job)
		}
	}
}

// processJob processes a single job
func (d *JobDispatcher) processJob(workerID int, job *Job) {
	d.mu.Lock()
	d.running[job.ID] = job
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.running, job.ID)
		d.mu.Unlock()
	}()

	// Mark as running
	now := time.Time{}
	job.StartTime = &now
	job.Status = PipelineStatusRunning
	job.UpdatedAt = now

	// Select runner
	runner, err := d.selectRunner(job)
	if err != nil {
		d.failJob(job, fmt.Errorf("failed to select runner: %w", err))
		return
	}

	job.RunnerID = runner.ID

	// Execute job
	if err := d.executeJob(job, runner); err != nil {
		d.failJob(job, err)
		return
	}

	// Mark as completed
	d.completeJob(job)
}

// selectRunner selects a runner for a job
func (d *JobDispatcher) selectRunner(job *Job) (*runner.RunnerInfo, error) {
	// Get all runners
	allRunners, err := d.runnerManager.ListRunners()
	if err != nil {
		return nil, fmt.Errorf("failed to list runners: %w", err)
	}

	// Filter by status (only use idle runners)
	var availableRunners []*runner.RunnerInfo
	for _, r := range allRunners {
		if r.Status == types.RunnerStatusOnline || r.Status == types.RunnerStatusBusy {
			availableRunners = append(availableRunners, r)
		}
	}

	if len(availableRunners) == 0 {
		return nil, fmt.Errorf("no available runners")
	}

	// TODO: Implement more sophisticated runner selection
	// - Match labels
	// - Consider load balancing
	// - Consider affinity/anti-affinity

	// For now, just return the first available runner
	return availableRunners[0], nil
}

// executeJob executes a job on a runner
func (d *JobDispatcher) executeJob(job *Job, runner *runner.RunnerInfo) error {
	// TODO: Implement actual job execution
	// - SSH into runner
	// - Execute commands
	// - Collect output
	// - Upload artifacts

	// For now, simulate execution
	fmt.Printf("[JobDispatcher] Executing job %s on runner %s\n", job.ID, runner.ID)

	// Simulate job duration
	time.Sleep(1 * time.Second)

	return nil
}

// failJob marks a job as failed
func (d *JobDispatcher) failJob(job *Job, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Time{}
	job.EndTime = &now
	job.Status = PipelineStatusFailed
	job.Duration = int64(now.Sub(*job.StartTime).Seconds())
	job.UpdatedAt = now

	// Check for retries
	if job.Retries < job.MaxRetries {
		job.Retries++
		job.Status = PipelineStatusRunning
		job.StartTime = nil
		job.EndTime = nil

		// Re-enqueue
		d.pending <- job
		return
	}

	d.failed[job.ID] = job

	// Save state
	d.saveState()

	fmt.Printf("[JobDispatcher] Job %s failed: %v\n", job.ID, err)
}

// completeJob marks a job as completed
func (d *JobDispatcher) completeJob(job *Job) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Time{}
	job.EndTime = &now
	job.Status = PipelineStatusSuccess
	job.Duration = int64(now.Sub(*job.StartTime).Seconds())
	job.UpdatedAt = now

	d.completed[job.ID] = job

	// Save state
	d.saveState()

	fmt.Printf("[JobDispatcher] Job %s completed in %d seconds\n", job.ID, job.Duration)
}

// GetJob returns a job by ID
func (d *JobDispatcher) GetJob(jobID string) (*Job, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	job, ok := d.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all jobs
func (d *JobDispatcher) ListJobs() []*Job {
	d.mu.RLock()
	defer d.mu.RUnlock()

	jobs := make([]*Job, 0, len(d.jobs))
	for _, job := range d.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListPendingJobs returns pending jobs
func (d *JobDispatcher) ListPendingJobs() []*Job {
	d.mu.RLock()
	defer d.mu.RUnlock()

	jobs := make([]*Job, 0)
	for _, job := range d.jobs {
		if job.Status == PipelineStatusRunning {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// ListRunningJobs returns running jobs
func (d *JobDispatcher) ListRunningJobs() []*Job {
	d.mu.RLock()
	defer d.mu.RUnlock()

	jobs := make([]*Job, 0, len(d.running))
	for _, job := range d.running {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListCompletedJobs returns completed jobs
func (d *JobDispatcher) ListCompletedJobs() []*Job {
	d.mu.RLock()
	defer d.mu.RUnlock()

	jobs := make([]*Job, 0, len(d.completed))
	for _, job := range d.completed {
		jobs = append(jobs, job)
	}
	return jobs
}

// ListFailedJobs returns failed jobs
func (d *JobDispatcher) ListFailedJobs() []*Job {
	d.mu.RLock()
	defer d.mu.RUnlock()

	jobs := make([]*Job, 0, len(d.failed))
	for _, job := range d.failed {
		jobs = append(jobs, job)
	}
	return jobs
}

// CancelJob cancels a job
func (d *JobDispatcher) CancelJob(ctx context.Context, jobID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	job, ok := d.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != PipelineStatusRunning {
		return fmt.Errorf("job is not running")
	}

	// TODO: Send cancel signal to runner

	now := time.Time{}
	job.EndTime = &now
	job.Status = PipelineStatusCanceled
	job.Duration = int64(now.Sub(*job.StartTime).Seconds())
	job.UpdatedAt = now

	delete(d.running, job.ID)

	// Save state
	d.saveState()

	return nil
}

// RetryJob retries a failed job
func (d *JobDispatcher) RetryJob(ctx context.Context, jobID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	job, ok := d.failed[jobID]
	if !ok {
		return fmt.Errorf("job not found or not failed: %s", jobID)
	}

	// Reset job state
	job.Retries = 0
	job.Status = PipelineStatusRunning
	job.StartTime = nil
	job.EndTime = nil
	job.UpdatedAt = time.Time{}

	// Remove from failed
	delete(d.failed, jobID)

	// Re-enqueue
	d.pending <- job

	return nil
}

// GetStats returns job statistics
func (d *JobDispatcher) GetStats() map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]int{
		"total":     len(d.jobs),
		"pending":   len(d.pending),
		"running":  len(d.running),
		"completed": len(d.completed),
		"failed":    len(d.failed),
	}
}

// Stop stops the dispatcher
func (d *JobDispatcher) Stop() error {
	close(d.stopChan)

	// Save state
	return d.saveState()
}

// Helper functions

func generateJobID() string {
	return fmt.Sprintf("job-%s-%d", randomString(8), time.Now().Unix())
}