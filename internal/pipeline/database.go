// Package pipeline provides ephemeral build environment orchestration
package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// PipelineStatus represents the current state of a pipeline
type PipelineStatus string

const (
	PipelineStatusCreating PipelineStatus = "creating"
	PipelineStatusRunning  PipelineStatus = "running"
	PipelineStatusSuccess  PipelineStatus = "success"
	PipelineStatusFailed   PipelineStatus = "failed"
	PipelineStatusCanceled PipelineStatus = "canceled"
)

// RunnerPlatform represents the CI/CD platform type
type RunnerPlatform string

const (
	PlatformGitLab   RunnerPlatform = "gitlab"
	PlatformGitHub   RunnerPlatform = "github"
	PlatformJenkins  RunnerPlatform = "jenkins"
	PlatformCircleCI RunnerPlatform = "circleci"
	PlatformDrone    RunnerPlatform = "drone"
)

// RunnerStatus represents the current state of a runner
type RunnerStatus string

const (
	RunnerStatusCreating RunnerStatus = "creating"
	RunnerStatusOnline   RunnerStatus = "online"
	RunnerStatusBusy     RunnerStatus = "busy"
	RunnerStatusOffline  RunnerStatus = "offline"
	RunnerStatusError    RunnerStatus = "error"
)

// VMState represents the current state of a VM
type VMState string

const (
	VMStateCreating  VMState = "creating"
	VMStateRunning   VMState = "running"
	VMStateIdle      VMState = "idle"
	VMStateBusy      VMState = "busy"
	VMStateStopping  VMState = "stopping"
	VMStateStopped   VMState = "stopped"
	VMStateDestroyed VMState = "destroyed"
)

// Pipeline represents a build pipeline
type Pipeline struct {
	ID         string         `json:"id"`
	Platform   RunnerPlatform `json:"platform"`
	Repository string         `json:"repository"`
	Branch     string         `json:"branch"`
	CommitSHA  string         `json:"commit_sha"`
	CommitMsg  string         `json:"commit_message"`
	Author     string         `json:"author"`
	Status     PipelineStatus `json:"status"`
	NetworkID  string         `json:"network_id"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Duration   int64          `json:"duration_seconds"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Runner represents a CI/CD runner
type Runner struct {
	ID          string         `json:"id"`
	PipelineID  string         `json:"pipeline_id"`
	VMID        string         `json:"vm_id"`
	Platform    RunnerPlatform `json:"platform"`
	PlatformID  string         `json:"platform_runner_id"`
	Token       string         `json:"-"` // Sensitive, not serialized
	Labels      []string       `json:"labels"`
	Name        string         `json:"name"`
	Status      RunnerStatus   `json:"status"`
	CurrentJob  string         `json:"current_job,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DestroyedAt *time.Time     `json:"destroyed_at,omitempty"`
}

// VM represents a virtual machine
type VM struct {
	ID          string     `json:"id"`
	PoolID      string     `json:"pool_id"`
	TemplateID  string     `json:"template_id"`
	Name        string     `json:"name"`
	IP          string     `json:"ip"`
	MAC         string     `json:"mac"`
	CPU         int        `json:"cpu"`
	Memory      int        `json:"memory"`
	State       VMState    `json:"state"`
	OverlayID   string     `json:"overlay_id"`
	RunnerToken string     `json:"-"` // Sensitive
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// Template represents a VM template
type Template struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BaseImage string    `json:"base_image"`
	Size      int64     `json:"size"`
	Packages  []string  `json:"packages"`
	OS        string    `json:"os"`
	OSVersion string    `json:"os_version"`
	Checksum  string    `json:"checksum"`
	ReadOnly  bool      `json:"read_only"`
	CreatedAt time.Time `json:"created_at"`
}

// Pool represents a VM pool
type Pool struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TemplateID  string    `json:"template_id"`
	MinSize     int       `json:"min_size"`
	MaxSize     int       `json:"max_size"`
	CurrentSize int       `json:"current_size"`
	CPU         int       `json:"cpu"`
	Memory      int       `json:"memory"`
	CreatedAt   time.Time `json:"created_at"`
}

// Network represents an isolated network
type Network struct {
	ID          string     `json:"id"`
	PipelineID  string     `json:"pipeline_id"`
	BridgeName  string     `json:"bridge_name"`
	VLANID      int        `json:"vlan_id"`
	CIDR        string     `json:"cidr"`
	Gateway     string     `json:"gateway"`
	CreatedAt   time.Time  `json:"created_at"`
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`
}

// PipelineDB provides database operations for pipeline management
type PipelineDB struct {
	db *sql.DB
}

// NewPipelineDB creates a new pipeline database
func NewPipelineDB(dbPath string) (*PipelineDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	pdb := &PipelineDB{db: db}
	if err := pdb.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return pdb, nil
}

// initialize creates the database schema
func (db *PipelineDB) initialize() error {
	schema := `
	-- Pipelines table
	CREATE TABLE IF NOT EXISTS pipelines (
		id TEXT PRIMARY KEY,
		platform TEXT NOT NULL,
		repository TEXT NOT NULL,
		branch TEXT NOT NULL,
		commit_sha TEXT NOT NULL,
		commit_message TEXT,
		author TEXT,
		status TEXT NOT NULL DEFAULT 'creating',
		network_id TEXT,
		start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		end_time DATETIME,
		duration_seconds INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (network_id) REFERENCES networks(id)
	);

	-- Runners table
	CREATE TABLE IF NOT EXISTS runners (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		vm_id TEXT NOT NULL,
		platform TEXT NOT NULL,
		platform_runner_id TEXT,
		token TEXT,
		labels TEXT,
		name TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'creating',
		current_job TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		destroyed_at DATETIME,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id),
		FOREIGN KEY (vm_id) REFERENCES vms(id)
	);

	-- VMs table
	CREATE TABLE IF NOT EXISTS vms (
		id TEXT PRIMARY KEY,
		pool_id TEXT NOT NULL,
		template_id TEXT NOT NULL,
		name TEXT NOT NULL,
		ip TEXT,
		mac TEXT,
		cpu INTEGER DEFAULT 2,
		memory INTEGER DEFAULT 4096,
		state TEXT NOT NULL DEFAULT 'creating',
		overlay_id TEXT,
		runner_token TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		destroyed_at DATETIME,
		FOREIGN KEY (pool_id) REFERENCES pools(id),
		FOREIGN KEY (template_id) REFERENCES templates(id)
	);

	-- Templates table
	CREATE TABLE IF NOT EXISTS templates (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		base_image TEXT NOT NULL,
		size INTEGER NOT NULL,
		packages TEXT,
		os TEXT DEFAULT 'ubuntu',
		os_version TEXT DEFAULT '24.04',
		checksum TEXT,
		read_only INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Pools table
	CREATE TABLE IF NOT EXISTS pools (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		template_id TEXT NOT NULL,
		min_size INTEGER DEFAULT 1,
		max_size INTEGER DEFAULT 10,
		current_size INTEGER DEFAULT 0,
		cpu INTEGER DEFAULT 2,
		memory INTEGER DEFAULT 4096,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (template_id) REFERENCES templates(id)
	);

	-- Networks table
	CREATE TABLE IF NOT EXISTS networks (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		bridge_name TEXT NOT NULL,
		vlan_id INTEGER NOT NULL,
		cidr TEXT NOT NULL,
		gateway TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		destroyed_at DATETIME,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
	);

	-- Artifacts table
	CREATE TABLE IF NOT EXISTS artifacts (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		size INTEGER NOT NULL,
		checksum TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
	);

	-- Logs table
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		pipeline_id TEXT NOT NULL,
		runner_id TEXT NOT NULL,
		stage TEXT,
		job TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		level TEXT DEFAULT 'info',
		message TEXT NOT NULL,
		duration_ms INTEGER DEFAULT 0,
		FOREIGN KEY (pipeline_id) REFERENCES pipelines(id),
		FOREIGN KEY (runner_id) REFERENCES runners(id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_pipelines_status ON pipelines(status);
	CREATE INDEX IF NOT EXISTS idx_pipelines_created ON pipelines(created_at);
	CREATE INDEX IF NOT EXISTS idx_runners_pipeline ON runners(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_runners_status ON runners(status);
	CREATE INDEX IF NOT EXISTS idx_vms_pool ON vms(pool_id);
	CREATE INDEX IF NOT EXISTS idx_vms_state ON vms(state);
	CREATE INDEX IF NOT EXISTS idx_networks_pipeline ON networks(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_logs_pipeline ON logs(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	`

	_, err := db.db.Exec(schema)
	return err
}

// Close closes the database connection
func (db *PipelineDB) Close() error {
	return db.db.Close()
}

// Pipeline operations

func (db *PipelineDB) SavePipeline(ctx context.Context, p *Pipeline) error {
	query := `
		INSERT OR REPLACE INTO pipelines 
		(id, platform, repository, branch, commit_sha, commit_message, author, status, 
		 network_id, start_time, end_time, duration_seconds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		p.ID, p.Platform, p.Repository, p.Branch, p.CommitSHA, p.CommitMsg, p.Author,
		p.Status, p.NetworkID, p.StartTime, p.EndTime, p.Duration, p.CreatedAt, p.UpdatedAt)
	return err
}

func (db *PipelineDB) GetPipeline(ctx context.Context, id string) (*Pipeline, error) {
	query := `SELECT id, platform, repository, branch, commit_sha, commit_message, author, status,
		network_id, start_time, end_time, duration_seconds, created_at, updated_at
		FROM pipelines WHERE id = ?`

	p := &Pipeline{}
	var endTime sql.NullTime
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Platform, &p.Repository, &p.Branch, &p.CommitSHA, &p.CommitMsg, &p.Author,
		&p.Status, &p.NetworkID, &p.StartTime, &endTime, &p.Duration, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		p.EndTime = &endTime.Time
	}

	return p, nil
}

func (db *PipelineDB) ListPipelines(ctx context.Context, limit, offset int) ([]*Pipeline, error) {
	query := `SELECT id, platform, repository, branch, commit_sha, commit_message, author, status,
		network_id, start_time, end_time, duration_seconds, created_at, updated_at
		FROM pipelines ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pipelines []*Pipeline
	for rows.Next() {
		p := &Pipeline{}
		var endTime sql.NullTime
		err := rows.Scan(
			&p.ID, &p.Platform, &p.Repository, &p.Branch, &p.CommitSHA, &p.CommitMsg, &p.Author,
			&p.Status, &p.NetworkID, &p.StartTime, &endTime, &p.Duration, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if endTime.Valid {
			p.EndTime = &endTime.Time
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, nil
}

func (db *PipelineDB) UpdatePipelineStatus(ctx context.Context, id string, status PipelineStatus) error {
	query := `UPDATE pipelines SET status = ?, updated_at = ? WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// Runner operations

func (db *PipelineDB) SaveRunner(ctx context.Context, r *Runner) error {
	labelsJSON, _ := json.Marshal(r.Labels)
	query := `INSERT OR REPLACE INTO runners 
		(id, pipeline_id, vm_id, platform, platform_runner_id, token, labels, name, status, 
		 current_job, created_at, destroyed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		r.ID, r.PipelineID, r.VMID, r.Platform, r.PlatformID, r.Token, string(labelsJSON),
		r.Name, r.Status, r.CurrentJob, r.CreatedAt, r.DestroyedAt)
	return err
}

func (db *PipelineDB) GetRunner(ctx context.Context, id string) (*Runner, error) {
	query := `SELECT id, pipeline_id, vm_id, platform, platform_runner_id, token, labels, name,
		status, current_job, created_at, destroyed_at FROM runners WHERE id = ?`

	r := &Runner{}
	var labelsJSON string
	var destroyedAt sql.NullTime
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.PipelineID, &r.VMID, &r.Platform, &r.PlatformID, &r.Token, &labelsJSON,
		&r.Name, &r.Status, &r.CurrentJob, &r.CreatedAt, &destroyedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(labelsJSON), &r.Labels)
	if destroyedAt.Valid {
		r.DestroyedAt = &destroyedAt.Time
	}

	return r, nil
}

func (db *PipelineDB) ListRunnersByPipeline(ctx context.Context, pipelineID string) ([]*Runner, error) {
	query := `SELECT id, pipeline_id, vm_id, platform, platform_runner_id, token, labels, name,
		status, current_job, created_at, destroyed_at FROM runners WHERE pipeline_id = ?`

	rows, err := db.db.QueryContext(ctx, query, pipelineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []*Runner
	for rows.Next() {
		r := &Runner{}
		var labelsJSON string
		var destroyedAt sql.NullTime
		err := rows.Scan(
			&r.ID, &r.PipelineID, &r.VMID, &r.Platform, &r.PlatformID, &r.Token, &labelsJSON,
			&r.Name, &r.Status, &r.CurrentJob, &r.CreatedAt, &destroyedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(labelsJSON), &r.Labels)
		if destroyedAt.Valid {
			r.DestroyedAt = &destroyedAt.Time
		}
		runners = append(runners, r)
	}

	return runners, nil
}

func (db *PipelineDB) DeleteRunner(ctx context.Context, id string) error {
	query := `DELETE FROM runners WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, id)
	return err
}

// VM operations

func (db *PipelineDB) SaveVM(ctx context.Context, vm *VM) error {
	query := `INSERT OR REPLACE INTO vms 
		(id, pool_id, template_id, name, ip, mac, cpu, memory, state, overlay_id, 
		 runner_token, created_at, destroyed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var destroyedAt interface{}
	if vm.DestroyedAt != nil {
		destroyedAt = vm.DestroyedAt
	}

	_, err := db.db.ExecContext(ctx, query,
		vm.ID, vm.PoolID, vm.TemplateID, vm.Name, vm.IP, vm.MAC, vm.CPU, vm.Memory,
		vm.State, vm.OverlayID, vm.RunnerToken, vm.CreatedAt, destroyedAt)
	return err
}

func (db *PipelineDB) GetVM(ctx context.Context, id string) (*VM, error) {
	query := `SELECT id, pool_id, template_id, name, ip, mac, cpu, memory, state, overlay_id,
		runner_token, created_at, destroyed_at FROM vms WHERE id = ?`

	vm := &VM{}
	var destroyedAt sql.NullTime
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&vm.ID, &vm.PoolID, &vm.TemplateID, &vm.Name, &vm.IP, &vm.MAC, &vm.CPU, &vm.Memory,
		&vm.State, &vm.OverlayID, &vm.RunnerToken, &vm.CreatedAt, &destroyedAt)
	if err != nil {
		return nil, err
	}

	if destroyedAt.Valid {
		vm.DestroyedAt = &destroyedAt.Time
	}

	return vm, nil
}

func (db *PipelineDB) ListVMsByPool(ctx context.Context, poolID string) ([]*VM, error) {
	query := `SELECT id, pool_id, template_id, name, ip, mac, cpu, memory, state, overlay_id,
		runner_token, created_at, destroyed_at FROM vms WHERE pool_id = ?`

	rows, err := db.db.QueryContext(ctx, query, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []*VM
	for rows.Next() {
		vm := &VM{}
		var destroyedAt sql.NullTime
		err := rows.Scan(
			&vm.ID, &vm.PoolID, &vm.TemplateID, &vm.Name, &vm.IP, &vm.MAC, &vm.CPU, &vm.Memory,
			&vm.State, &vm.OverlayID, &vm.RunnerToken, &vm.CreatedAt, &destroyedAt)
		if err != nil {
			return nil, err
		}
		if destroyedAt.Valid {
			vm.DestroyedAt = &destroyedAt.Time
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

func (db *PipelineDB) UpdateVMState(ctx context.Context, id string, state VMState) error {
	query := `UPDATE vms SET state = ? WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, state, id)
	return err
}

func (db *PipelineDB) DeleteVM(ctx context.Context, id string) error {
	query := `DELETE FROM vms WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, id)
	return err
}

// Template operations

func (db *PipelineDB) SaveTemplate(ctx context.Context, t *Template) error {
	packagesJSON, _ := json.Marshal(t.Packages)
	query := `INSERT OR REPLACE INTO templates 
		(id, name, base_image, size, packages, os, os_version, checksum, read_only, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		t.ID, t.Name, t.BaseImage, t.Size, string(packagesJSON),
		t.OS, t.OSVersion, t.Checksum, t.ReadOnly, t.CreatedAt)
	return err
}

func (db *PipelineDB) GetTemplate(ctx context.Context, id string) (*Template, error) {
	query := `SELECT id, name, base_image, size, packages, os, os_version, checksum, read_only, created_at
		FROM templates WHERE id = ?`

	t := &Template{}
	var packagesJSON string
	var readOnly int
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.BaseImage, &t.Size, &packagesJSON, &t.OS, &t.OSVersion,
		&t.Checksum, &readOnly, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(packagesJSON), &t.Packages)
	t.ReadOnly = readOnly == 1
	return t, nil
}

func (db *PipelineDB) ListTemplates(ctx context.Context) ([]*Template, error) {
	query := `SELECT id, name, base_image, size, packages, os, os_version, checksum, read_only, created_at
		FROM templates ORDER BY name`

	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*Template
	for rows.Next() {
		t := &Template{}
		var packagesJSON string
		var readOnly int
		err := rows.Scan(
			&t.ID, &t.Name, &t.BaseImage, &t.Size, &packagesJSON, &t.OS, &t.OSVersion,
			&t.Checksum, &readOnly, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(packagesJSON), &t.Packages)
		t.ReadOnly = readOnly == 1
		templates = append(templates, t)
	}

	return templates, nil
}

func (db *PipelineDB) DeleteTemplate(ctx context.Context, id string) error {
	query := `DELETE FROM templates WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, id)
	return err
}

// Pool operations

func (db *PipelineDB) SavePool(ctx context.Context, p *Pool) error {
	query := `INSERT OR REPLACE INTO pools 
		(id, name, template_id, min_size, max_size, current_size, cpu, memory, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		p.ID, p.Name, p.TemplateID, p.MinSize, p.MaxSize, p.CurrentSize, p.CPU, p.Memory, p.CreatedAt)
	return err
}

func (db *PipelineDB) GetPool(ctx context.Context, id string) (*Pool, error) {
	query := `SELECT id, name, template_id, min_size, max_size, current_size, cpu, memory, created_at
		FROM pools WHERE id = ?`

	p := &Pool{}
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.TemplateID, &p.MinSize, &p.MaxSize, &p.CurrentSize,
		&p.CPU, &p.Memory, &p.CreatedAt)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (db *PipelineDB) ListPools(ctx context.Context) ([]*Pool, error) {
	query := `SELECT id, name, template_id, min_size, max_size, current_size, cpu, memory, created_at
		FROM pools ORDER BY name`

	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []*Pool
	for rows.Next() {
		p := &Pool{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.TemplateID, &p.MinSize, &p.MaxSize, &p.CurrentSize,
			&p.CPU, &p.Memory, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		pools = append(pools, p)
	}

	return pools, nil
}

func (db *PipelineDB) UpdatePoolSize(ctx context.Context, id string, delta int) error {
	query := `UPDATE pools SET current_size = current_size + ? WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, delta, id)
	return err
}

// Network operations

func (db *PipelineDB) SaveNetwork(ctx context.Context, n *Network) error {
	query := `INSERT OR REPLACE INTO networks 
		(id, pipeline_id, bridge_name, vlan_id, cidr, gateway, created_at, destroyed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	var destroyedAt interface{}
	if n.DestroyedAt != nil {
		destroyedAt = n.DestroyedAt
	}

	_, err := db.db.ExecContext(ctx, query,
		n.ID, n.PipelineID, n.BridgeName, n.VLANID, n.CIDR, n.Gateway, n.CreatedAt, destroyedAt)
	return err
}

func (db *PipelineDB) GetNetwork(ctx context.Context, id string) (*Network, error) {
	query := `SELECT id, pipeline_id, bridge_name, vlan_id, cidr, gateway, created_at, destroyed_at
		FROM networks WHERE id = ?`

	n := &Network{}
	var destroyedAt sql.NullTime
	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.PipelineID, &n.BridgeName, &n.VLANID, &n.CIDR, &n.Gateway, &n.CreatedAt, &destroyedAt)
	if err != nil {
		return nil, err
	}

	if destroyedAt.Valid {
		n.DestroyedAt = &destroyedAt.Time
	}

	return n, nil
}

func (db *PipelineDB) DeleteNetwork(ctx context.Context, id string) error {
	query := `DELETE FROM networks WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, id)
	return err
}

// Artifact operations

func (db *PipelineDB) SaveArtifact(ctx context.Context, a *Artifact) error {
	query := `INSERT OR REPLACE INTO artifacts 
		(id, pipeline_id, type, name, path, size, checksum, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		a.ID, a.PipelineID, a.Type, a.Name, a.Path, a.Size, a.Checksum, a.CreatedAt)
	return err
}

func (db *PipelineDB) DeleteArtifact(ctx context.Context, id string) error {
	query := `DELETE FROM artifacts WHERE id = ?`
	_, err := db.db.ExecContext(ctx, query, id)
	return err
}

func (db *PipelineDB) ListArtifactsByPipeline(ctx context.Context, pipelineID string) ([]*Artifact, error) {
	query := `SELECT id, pipeline_id, type, name, path, size, checksum, created_at
		FROM artifacts WHERE pipeline_id = ?`

	rows, err := db.db.QueryContext(ctx, query, pipelineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []*Artifact
	for rows.Next() {
		a := &Artifact{}
		err := rows.Scan(
			&a.ID, &a.PipelineID, &a.Type, &a.Name, &a.Path, &a.Size, &a.Checksum, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}

	return artifacts, nil
}

// Log operations

func (db *PipelineDB) SaveLog(ctx context.Context, l *LogEntry) error {
	query := `INSERT INTO logs 
		(id, pipeline_id, runner_id, stage, job_id, timestamp, level, message, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query,
		l.ID, l.PipelineID, l.RunnerID, l.Stage, l.JobID, l.Timestamp, l.Level, l.Message, l.Duration)
	return err
}

func (db *PipelineDB) ListLogsByPipeline(ctx context.Context, pipelineID string, limit, offset int) ([]*LogEntry, error) {
	query := `SELECT id, pipeline_id, runner_id, stage, job_id, timestamp, level, message, duration_ms
		FROM logs WHERE pipeline_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := db.db.QueryContext(ctx, query, pipelineID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogEntry
	for rows.Next() {
		l := &LogEntry{}
		err := rows.Scan(
			&l.ID, &l.PipelineID, &l.RunnerID, &l.Stage, &l.JobID, &l.Timestamp,
			&l.Level, &l.Message, &l.Duration)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

// Stats operations

func (db *PipelineDB) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	counts := []struct {
		name  string
		table string
	}{
		{"pipelines", "pipelines"},
		{"runners", "runners"},
		{"vms", "vms"},
		{"templates", "templates"},
		{"pools", "pools"},
		{"networks", "networks"},
		{"artifacts", "artifacts"},
		{"logs", "logs"},
	}

	for _, c := range counts {
		var count int
		err := db.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", c.table)).Scan(&count)
		if err != nil {
			return nil, err
		}
		stats[c.name] = count
	}

	return stats, nil
}
