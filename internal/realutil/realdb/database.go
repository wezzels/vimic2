// Package realdb provides real database operations for production use
package realdb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database provides real database operations
type Database struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// Config holds database configuration
type Config struct {
	Path         string
	MaxOpenConns int
	MaxIdleConns int
	BusyTimeout  time.Duration
	WALMode      bool
}

// Cluster represents a cluster record
type Cluster struct {
	ID        string
	Name      string
	Status    string
	Config    *ClusterConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	MinNodes  int `json:"min_nodes"`
	MaxNodes  int `json:"max_nodes"`
	AutoScale bool `json:"auto_scale"`
}

// Host represents a host record
type Host struct {
	ID         string
	Name       string
	Address    string
	Port       int
	User       string
	SSHKeyPath string
	HVType     string
	CreatedAt  time.Time
}

// Node represents a node record
type Node struct {
	ID        string
	ClusterID string
	HostID    string
	Name      string
	Role      string
	State     string
	IP        string
	Config    *NodeConfig
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NodeConfig represents node configuration
type NodeConfig struct {
	CPU    int    `json:"cpu"`
	Memory int    `json:"memory"`
	Disk   int    `json:"disk"`
	Image  string `json:"image"`
}

// Pool represents a pool record
type Pool struct {
	ID        string
	Name      string
	Available int
	Busy      int
}

// Metric represents a metric record
type Metric struct {
	ID         int64
	NodeID     string
	CPU        float64
	Memory     float64
	Disk       float64
	NetworkRX  float64
	NetworkTX  float64
	RecordedAt time.Time
}

// Alert represents an alert record
type Alert struct {
	ID        string
	NodeID    string
	Type      string
	Message   string
	Severity  string
	CreatedAt time.Time
}

// NewDatabase creates a new database instance
func NewDatabase(cfg *Config) (*Database, error) {
	if cfg.Path == "" {
		cfg.Path = ":memory:"
	}

	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// Set busy timeout for SQLite
	if cfg.BusyTimeout > 0 {
		_, err = db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", cfg.BusyTimeout.Milliseconds()))
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set busy timeout: %w", err)
		}
	}

	// Enable WAL mode for better concurrent access
	if cfg.WALMode {
		_, err = db.Exec("PRAGMA journal_mode = WAL")
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
		}
	}

	database := &Database{
		db:   db,
		path: cfg.Path,
	}

	// Run migrations
	if err := database.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return database, nil
}

// NewDatabaseWithDefaults creates a database with sensible defaults
func NewDatabaseWithDefaults(path string) (*Database, error) {
	return NewDatabase(&Config{
		Path:         path,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		BusyTimeout:  5 * time.Second,
		WALMode:      true,
	})
}

// migrate runs database migrations
func (d *Database) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS clusters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'created',
			config TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS hosts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			address TEXT NOT NULL,
			port INTEGER DEFAULT 22,
			user TEXT DEFAULT 'root',
			ssh_key_path TEXT,
			hv_type TEXT DEFAULT 'libvirt',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			cluster_id TEXT REFERENCES clusters(id),
			host_id TEXT REFERENCES hosts(id),
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'worker',
			state TEXT NOT NULL DEFAULT 'created',
			ip TEXT,
			config TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pools (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			available INTEGER DEFAULT 0,
			busy INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			node_id TEXT REFERENCES nodes(id),
			cpu REAL,
			memory REAL,
			disk REAL,
			network_rx REAL,
			network_tx REAL,
			recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			node_id TEXT REFERENCES nodes(id),
			type TEXT NOT NULL,
			message TEXT,
			severity TEXT DEFAULT 'info',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_cluster ON nodes(cluster_id)`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_host ON nodes(host_id)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_node_time ON metrics(node_id, recorded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_node ON alerts(node_id)`,
	}

	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping checks database connectivity
func (d *Database) Ping() error {
	return d.db.Ping()
}

// Stats returns database statistics
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// Begin starts a transaction
func (d *Database) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// Cluster operations

func (d *Database) SaveCluster(cluster *Cluster) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	configJSON, err := json.Marshal(cluster.Config)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		INSERT INTO clusters (id, name, status, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			status = excluded.status,
			config = excluded.config,
			updated_at = excluded.updated_at
	`, cluster.ID, cluster.Name, cluster.Status, configJSON, cluster.CreatedAt, cluster.UpdatedAt)

	return err
}

func (d *Database) GetCluster(id string) (*Cluster, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var configJSON sql.NullString
	var createdAt, updatedAt sql.NullTime
	cluster := &Cluster{}

	err := d.db.QueryRow(`
		SELECT id, name, status, config, created_at, updated_at
		FROM clusters WHERE id = ?
	`, id).Scan(&cluster.ID, &cluster.Name, &cluster.Status, &configJSON,
		&createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cluster not found: %s", id)
		}
		return nil, err
	}

	if configJSON.Valid && configJSON.String != "" {
		cluster.Config = &ClusterConfig{}
		json.Unmarshal([]byte(configJSON.String), cluster.Config)
	}
	if createdAt.Valid {
		cluster.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		cluster.UpdatedAt = updatedAt.Time
	}

	return cluster, nil
}

func (d *Database) ListClusters() ([]*Cluster, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT id, name, status, config, created_at, updated_at
		FROM clusters ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*Cluster
	for rows.Next() {
		var configJSON string
		cluster := &Cluster{}
		err := rows.Scan(&cluster.ID, &cluster.Name, &cluster.Status, &configJSON,
			&cluster.CreatedAt, &cluster.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if configJSON != "" {
			cluster.Config = &ClusterConfig{}
			json.Unmarshal([]byte(configJSON), cluster.Config)
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (d *Database) DeleteCluster(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM clusters WHERE id = ?", id)
	return err
}

// Host operations

func (d *Database) SaveHost(host *Host) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		INSERT INTO hosts (id, name, address, port, user, ssh_key_path, hv_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			address = excluded.address,
			port = excluded.port,
			user = excluded.user,
			ssh_key_path = excluded.ssh_key_path,
			hv_type = excluded.hv_type
	`, host.ID, host.Name, host.Address, host.Port, host.User, host.SSHKeyPath, host.HVType, host.CreatedAt)

	return err
}

func (d *Database) GetHost(id string) (*Host, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var sshKeyPath sql.NullString
	host := &Host{}
	err := d.db.QueryRow(`
		SELECT id, name, address, port, user, ssh_key_path, hv_type, created_at
		FROM hosts WHERE id = ?
	`, id).Scan(&host.ID, &host.Name, &host.Address, &host.Port, &host.User,
		&sshKeyPath, &host.HVType, &host.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("host not found: %s", id)
		}
		return nil, err
	}

	if sshKeyPath.Valid {
		host.SSHKeyPath = sshKeyPath.String
	}

	return host, nil
}

func (d *Database) ListHosts() ([]*Host, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT id, name, address, port, user, ssh_key_path, hv_type, created_at
		FROM hosts ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*Host
	for rows.Next() {
		host := &Host{}
		err := rows.Scan(&host.ID, &host.Name, &host.Address, &host.Port, &host.User,
			&host.SSHKeyPath, &host.HVType, &host.CreatedAt)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}

	return hosts, nil
}

func (d *Database) DeleteHost(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM hosts WHERE id = ?", id)
	return err
}

// Node operations

func (d *Database) SaveNode(node *Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	configJSON, err := json.Marshal(node.Config)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		INSERT INTO nodes (id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			cluster_id = excluded.cluster_id,
			host_id = excluded.host_id,
			name = excluded.name,
			role = excluded.role,
			state = excluded.state,
			ip = excluded.ip,
			config = excluded.config,
			updated_at = excluded.updated_at
	`, node.ID, node.ClusterID, node.HostID, node.Name, node.Role, node.State, node.IP,
		configJSON, node.CreatedAt, node.UpdatedAt)

	return err
}

func (d *Database) GetNode(id string) (*Node, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var configJSON sql.NullString
	var createdAt, updatedAt sql.NullTime
	node := &Node{}

	err := d.db.QueryRow(`
		SELECT id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at
		FROM nodes WHERE id = ?
	`, id).Scan(&node.ID, &node.ClusterID, &node.HostID, &node.Name, &node.Role,
		&node.State, &node.IP, &configJSON, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("node not found: %s", id)
		}
		return nil, err
	}

	if configJSON.Valid && configJSON.String != "" {
		node.Config = &NodeConfig{}
		json.Unmarshal([]byte(configJSON.String), node.Config)
	}
	if createdAt.Valid {
		node.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		node.UpdatedAt = updatedAt.Time
	}

	return node, nil
}

func (d *Database) ListClusterNodes(clusterID string) ([]*Node, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at
		FROM nodes WHERE cluster_id = ? ORDER BY created_at DESC
	`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var configJSON string
		node := &Node{}
		err := rows.Scan(&node.ID, &node.ClusterID, &node.HostID, &node.Name, &node.Role,
			&node.State, &node.IP, &configJSON, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if configJSON != "" {
			node.Config = &NodeConfig{}
			json.Unmarshal([]byte(configJSON), node.Config)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (d *Database) DeleteNode(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM nodes WHERE id = ?", id)
	return err
}

// Metric operations

func (d *Database) SaveMetric(metric *Metric) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	result, err := d.db.Exec(`
		INSERT INTO metrics (node_id, cpu, memory, disk, network_rx, network_tx, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, metric.NodeID, metric.CPU, metric.Memory, metric.Disk, metric.NetworkRX,
		metric.NetworkTX, metric.RecordedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	metric.ID = id
	return nil
}

func (d *Database) GetNodeMetrics(nodeID string, since time.Time) ([]*Metric, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT id, node_id, cpu, memory, disk, network_rx, network_tx, recorded_at
		FROM metrics WHERE node_id = ? AND recorded_at >= ?
		ORDER BY recorded_at DESC
	`, nodeID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		metric := &Metric{}
		err := rows.Scan(&metric.ID, &metric.NodeID, &metric.CPU, &metric.Memory,
			&metric.Disk, &metric.NetworkRX, &metric.NetworkTX, &metric.RecordedAt)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// Alert operations

func (d *Database) SaveAlert(alert *Alert) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		INSERT INTO alerts (id, node_id, type, message, severity, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			node_id = excluded.node_id,
			type = excluded.type,
			message = excluded.message,
			severity = excluded.severity
	`, alert.ID, alert.NodeID, alert.Type, alert.Message, alert.Severity, alert.CreatedAt)

	return err
}

func (d *Database) GetAlert(id string) (*Alert, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	alert := &Alert{}
	err := d.db.QueryRow(`
		SELECT id, node_id, type, message, severity, created_at
		FROM alerts WHERE id = ?
	`, id).Scan(&alert.ID, &alert.NodeID, &alert.Type, &alert.Message,
		&alert.Severity, &alert.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("alert not found: %s", id)
		}
		return nil, err
	}

	return alert, nil
}

func (d *Database) ListAlerts() ([]*Alert, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT id, node_id, type, message, severity, created_at
		FROM alerts ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*Alert
	for rows.Next() {
		alert := &Alert{}
		err := rows.Scan(&alert.ID, &alert.NodeID, &alert.Type, &alert.Message,
			&alert.Severity, &alert.CreatedAt)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (d *Database) DeleteAlert(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM alerts WHERE id = ?", id)
	return err
}

// Pool operations

func (d *Database) SavePool(pool *Pool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec(`
		INSERT INTO pools (id, name, available, busy)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			available = excluded.available,
			busy = excluded.busy
	`, pool.ID, pool.Name, pool.Available, pool.Busy)

	return err
}

func (d *Database) GetPool(id string) (*Pool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pool := &Pool{}
	err := d.db.QueryRow(`
		SELECT id, name, available, busy
		FROM pools WHERE id = ?
	`, id).Scan(&pool.ID, &pool.Name, &pool.Available, &pool.Busy)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("pool not found: %s", id)
		}
		return nil, err
	}

	return pool, nil
}

func (d *Database) ListPools() ([]*Pool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`SELECT id, name, available, busy FROM pools`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []*Pool
	for rows.Next() {
		pool := &Pool{}
		err := rows.Scan(&pool.ID, &pool.Name, &pool.Available, &pool.Busy)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

func (d *Database) DeletePool(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("DELETE FROM pools WHERE id = ?", id)
	return err
}

// Count returns total counts for all tables
func (d *Database) Count() map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	counts := make(map[string]int)

	tables := []string{"clusters", "hosts", "nodes", "alerts", "pools"}
	for _, table := range tables {
		var count int
		err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			count = 0
		}
		counts[table] = count
	}

	return counts
}

// Backup creates a backup of the database
func (d *Database) Backup(backupPath string) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, err := d.db.Exec("VACUUM INTO ?", backupPath)
	return err
}

// Vacuum optimizes the database
func (d *Database) Vacuum() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.db.Exec("VACUUM")
	return err
}

// IntegrityCheck runs an integrity check
func (d *Database) IntegrityCheck() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var result string
	err := d.db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return err
	}

	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}

	return nil
}