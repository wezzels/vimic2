// Package database provides SQLite persistence for Vimic2
package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database handle
type DB struct {
	db *sql.DB
}

// NewDB creates a new database connection
func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &DB{db: db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return d, nil
}

// Close closes the database
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate creates the database schema
func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS hosts (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		address TEXT NOT NULL,
		port INTEGER DEFAULT 22,
		user TEXT NOT NULL,
		ssh_key_path TEXT,
		hv_type TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS clusters (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		config TEXT,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		cluster_id TEXT,
		host_id TEXT,
		name TEXT NOT NULL,
		role TEXT DEFAULT 'worker',
		state TEXT DEFAULT 'stopped',
		ip TEXT,
		config TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE,
		FOREIGN KEY (host_id) REFERENCES hosts(id)
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id TEXT,
		cpu REAL DEFAULT 0,
		memory REAL DEFAULT 0,
		disk REAL DEFAULT 0,
		network_rx REAL DEFAULT 0,
		network_tx REAL DEFAULT 0,
		recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_nodes_cluster ON nodes(cluster_id);
	CREATE INDEX IF NOT EXISTS idx_nodes_host ON nodes(host_id);
	CREATE INDEX IF NOT EXISTS idx_metrics_node ON metrics(node_id);
	CREATE INDEX IF NOT EXISTS idx_metrics_time ON metrics(recorded_at);

	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		rule_id TEXT,
		node_id TEXT,
		node_name TEXT,
		metric TEXT,
		value REAL,
		threshold REAL,
		message TEXT,
		fired_at TIMESTAMP,
		resolved INTEGER DEFAULT 0,
		resolved_at TIMESTAMP,
		FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_node ON alerts(node_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// ============== Host Operations ==============

// Host represents a hypervisor host
type Host struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Port       int       `json:"port"`
	User       string    `json:"user"`
	SSHKeyPath string    `json:"ssh_key_path"`
	HVType     string    `json:"hv_type"` // libvirt, hyperv, apple
	CreatedAt  time.Time `json:"created_at"`
}

// SaveHost saves a host to the database
func (d *DB) SaveHost(host *Host) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO hosts (id, name, address, port, user, ssh_key_path, hv_type)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, host.ID, host.Name, host.Address, host.Port, host.User, host.SSHKeyPath, host.HVType)
	return err
}

// GetHost retrieves a host by ID
func (d *DB) GetHost(id string) (*Host, error) {
	row := d.db.QueryRow(`
		SELECT id, name, address, port, user, ssh_key_path, hv_type, created_at 
		FROM hosts WHERE id = ?
	`, id)
	h := &Host{}
	err := row.Scan(&h.ID, &h.Name, &h.Address, &h.Port, &h.User, &h.SSHKeyPath, &h.HVType, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return h, err
}

// ListHosts returns all hosts
func (d *DB) ListHosts() ([]*Host, error) {
	rows, err := d.db.Query(`
		SELECT id, name, address, port, user, ssh_key_path, hv_type, created_at 
		FROM hosts ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*Host
	for rows.Next() {
		h := &Host{}
		if err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.Port, &h.User, &h.SSHKeyPath, &h.HVType, &h.CreatedAt); err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// DeleteHost deletes a host
func (d *DB) DeleteHost(id string) error {
	_, err := d.db.Exec("DELETE FROM hosts WHERE id = ?", id)
	return err
}

// ============== Cluster Operations ==============

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	MinNodes      int            `json:"min_nodes"`
	MaxNodes      int            `json:"max_nodes"`
	AutoScale     bool           `json:"autoscale"`
	ScaleOnCPU    float64        `json:"scale_on_cpu"`
	ScaleOnMemory float64        `json:"scale_on_memory"`
	CooldownSec   int            `json:"cooldown_sec"`
	Network       *NetworkConfig `json:"network"`
	NodeDefaults  *NodeConfig    `json:"node_defaults"`
}

// NetworkConfig holds network configuration
type NetworkConfig struct {
	Type string `json:"type"` // nat, bridge
	CIDR string `json:"cidr"`
}

// NodeConfig holds default node configuration
type NodeConfig struct {
	CPU      int    `json:"cpu"`
	MemoryMB uint64 `json:"memory_mb"`
	DiskGB   int    `json:"disk_gb"`
	Image    string `json:"image"`
}

// Cluster represents a VM cluster
type Cluster struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Config    *ClusterConfig `json:"config"`
	Status    string         `json:"status"` // pending, deploying, running, degraded, error
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SaveCluster saves a cluster to the database
func (d *DB) SaveCluster(c *Cluster) error {
	configJSON, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		INSERT OR REPLACE INTO clusters (id, name, config, status)
		VALUES (?, ?, ?, ?)
	`, c.ID, c.Name, string(configJSON), c.Status)
	return err
}

// GetCluster retrieves a cluster by ID
func (d *DB) GetCluster(id string) (*Cluster, error) {
	row := d.db.QueryRow(`
		SELECT id, name, config, status, created_at, updated_at 
		FROM clusters WHERE id = ?
	`, id)

	c := &Cluster{}
	var configJSON string
	err := row.Scan(&c.ID, &c.Name, &configJSON, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if configJSON != "" {
		c.Config = &ClusterConfig{}
		if err := json.Unmarshal([]byte(configJSON), c.Config); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// ListClusters returns all clusters
func (d *DB) ListClusters() ([]*Cluster, error) {
	rows, err := d.db.Query(`
		SELECT id, name, config, status, created_at, updated_at 
		FROM clusters ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*Cluster
	for rows.Next() {
		c := &Cluster{}
		var configJSON string
		if err := rows.Scan(&c.ID, &c.Name, &configJSON, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if configJSON != "" {
			c.Config = &ClusterConfig{}
			json.Unmarshal([]byte(configJSON), c.Config)
		}
		clusters = append(clusters, c)
	}
	return clusters, nil
}

// UpdateClusterStatus updates a cluster's status
func (d *DB) UpdateClusterStatus(id, status string) error {
	_, err := d.db.Exec(`
		UPDATE clusters SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, status, id)
	return err
}

// DeleteCluster deletes a cluster
func (d *DB) DeleteCluster(id string) error {
	_, err := d.db.Exec("DELETE FROM clusters WHERE id = ?", id)
	return err
}

// ============== Node Operations ==============

// Node represents a virtual machine node
type Node struct {
	ID        string      `json:"id"`
	ClusterID string      `json:"cluster_id"`
	HostID    string      `json:"host_id"`
	Name      string      `json:"name"`
	Role      string      `json:"role"`  // worker, database, master
	State     string      `json:"state"` // pending, running, stopped, error
	IP        string      `json:"ip"`
	Config    *NodeConfig `json:"config"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// SaveNode saves a node to the database
func (d *DB) SaveNode(n *Node) error {
	configJSON, _ := json.Marshal(n.Config)

	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO nodes (id, cluster_id, host_id, name, role, state, ip, config)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, n.ID, n.ClusterID, n.HostID, n.Name, n.Role, n.State, n.IP, string(configJSON))
	return err
}

// GetNode retrieves a node by ID
func (d *DB) GetNode(id string) (*Node, error) {
	row := d.db.QueryRow(`
		SELECT id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at 
		FROM nodes WHERE id = ?
	`, id)

	n := &Node{}
	var configJSON sql.NullString
	err := row.Scan(&n.ID, &n.ClusterID, &n.HostID, &n.Name, &n.Role, &n.State, &n.IP, &configJSON, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if configJSON.Valid {
		n.Config = &NodeConfig{}
		json.Unmarshal([]byte(configJSON.String), n.Config)
	}
	return n, nil
}

// ListClusterNodes returns all nodes in a cluster
func (d *DB) ListClusterNodes(clusterID string) ([]*Node, error) {
	rows, err := d.db.Query(`
		SELECT id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at 
		FROM nodes WHERE cluster_id = ? ORDER BY name
	`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		n := &Node{}
		var configJSON sql.NullString
		if err := rows.Scan(&n.ID, &n.ClusterID, &n.HostID, &n.Name, &n.Role, &n.State, &n.IP, &configJSON, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		if configJSON.Valid {
			n.Config = &NodeConfig{}
			json.Unmarshal([]byte(configJSON.String), n.Config)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// ListAllNodes returns all nodes
func (d *DB) ListAllNodes() ([]*Node, error) {
	rows, err := d.db.Query(`
		SELECT id, cluster_id, host_id, name, role, state, ip, config, created_at, updated_at 
		FROM nodes ORDER BY cluster_id, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		n := &Node{}
		var configJSON sql.NullString
		if err := rows.Scan(&n.ID, &n.ClusterID, &n.HostID, &n.Name, &n.Role, &n.State, &n.IP, &configJSON, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		if configJSON.Valid {
			n.Config = &NodeConfig{}
			json.Unmarshal([]byte(configJSON.String), n.Config)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// UpdateNodeState updates a node's state
func (d *DB) UpdateNodeState(id, state, ip string) error {
	_, err := d.db.Exec(`
		UPDATE nodes SET state = ?, ip = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, state, ip, id)
	return err
}

// DeleteNode deletes a node
func (d *DB) DeleteNode(id string) error {
	_, err := d.db.Exec("DELETE FROM nodes WHERE id = ?", id)
	return err
}

// ============== Metrics Operations ==============

// Metric represents a metrics data point
type Metric struct {
	ID         int64     `json:"id"`
	NodeID     string    `json:"node_id"`
	CPU        float64   `json:"cpu"`
	Memory     float64   `json:"memory"`
	Disk       float64   `json:"disk"`
	NetworkRX  float64   `json:"network_rx"`
	NetworkTX  float64   `json:"network_tx"`
	RecordedAt time.Time `json:"recorded_at"`
}

// SaveMetric saves a metric
func (d *DB) SaveMetric(m *Metric) error {
	_, err := d.db.Exec(`
		INSERT INTO metrics (node_id, cpu, memory, disk, network_rx, network_tx)
		VALUES (?, ?, ?, ?, ?, ?)
	`, m.NodeID, m.CPU, m.Memory, m.Disk, m.NetworkRX, m.NetworkTX)
	return err
}

// GetLatestMetric returns the latest metric for a node
func (d *DB) GetLatestMetric(nodeID string) (*Metric, error) {
	row := d.db.QueryRow(`
		SELECT id, node_id, cpu, memory, disk, network_rx, network_tx, recorded_at
		FROM metrics WHERE node_id = ? ORDER BY recorded_at DESC LIMIT 1
	`, nodeID)

	m := &Metric{}
	err := row.Scan(&m.ID, &m.NodeID, &m.CPU, &m.Memory, &m.Disk, &m.NetworkRX, &m.NetworkTX, &m.RecordedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

// GetNodeMetrics returns metrics for a node within a time range
func (d *DB) GetNodeMetrics(nodeID string, since time.Time) ([]*Metric, error) {
	rows, err := d.db.Query(`
		SELECT id, node_id, cpu, memory, disk, network_rx, network_tx, recorded_at
		FROM metrics WHERE node_id = ? AND recorded_at >= ? ORDER BY recorded_at
	`, nodeID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		m := &Metric{}
		if err := rows.Scan(&m.ID, &m.NodeID, &m.CPU, &m.Memory, &m.Disk, &m.NetworkRX, &m.NetworkTX, &m.RecordedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// CleanupOldMetrics removes metrics older than the specified duration
func (d *DB) CleanupOldMetrics(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := d.db.Exec("DELETE FROM metrics WHERE recorded_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ============== Alert Operations ==============

// Alert represents an alert
type Alert struct {
	ID         string     `json:"id"`
	RuleID     string     `json:"rule_id"`
	NodeID     string     `json:"node_id"`
	NodeName   string     `json:"node_name"`
	Metric     string     `json:"metric"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	Message    string     `json:"message"`
	FiredAt    time.Time  `json:"fired_at"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// SaveAlert saves an alert to the database
func (d *DB) SaveAlert(alert *Alert) error {
	resolved := 0
	if alert.Resolved {
		resolved = 1
	}
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO alerts (id, rule_id, node_id, node_name, metric, value, threshold, message, fired_at, resolved, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, alert.ID, alert.RuleID, alert.NodeID, alert.NodeName, alert.Metric, alert.Value, alert.Threshold, alert.Message, alert.FiredAt, resolved, alert.ResolvedAt)
	return err
}

// GetActiveAlerts returns all unresolved alerts
func (d *DB) GetActiveAlerts() ([]*Alert, error) {
	rows, err := d.db.Query(`
		SELECT id, rule_id, node_id, node_name, metric, value, threshold, message, fired_at, resolved, resolved_at
		FROM alerts WHERE resolved = 0 ORDER BY fired_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*Alert
	for rows.Next() {
		a := &Alert{}
		var resolved int
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.RuleID, &a.NodeID, &a.NodeName, &a.Metric, &a.Value, &a.Threshold, &a.Message, &a.FiredAt, &resolved, &resolvedAt); err != nil {
			return nil, err
		}
		a.Resolved = resolved == 1
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

// GetNodeAlerts returns all alerts for a node
func (d *DB) GetNodeAlerts(nodeID string) ([]*Alert, error) {
	rows, err := d.db.Query(`
		SELECT id, rule_id, node_id, node_name, metric, value, threshold, message, fired_at, resolved, resolved_at
		FROM alerts WHERE node_id = ? ORDER BY fired_at DESC LIMIT 100
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*Alert
	for rows.Next() {
		a := &Alert{}
		var resolved int
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.RuleID, &a.NodeID, &a.NodeName, &a.Metric, &a.Value, &a.Threshold, &a.Message, &a.FiredAt, &resolved, &resolvedAt); err != nil {
			return nil, err
		}
		a.Resolved = resolved == 1
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}
