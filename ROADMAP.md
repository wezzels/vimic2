# Vimic2 - Phase 1 Implementation Roadmap

**Focus:** Pure cluster management - no AI chat, no journaling. Just VM provisioning, monitoring, and orchestration.

---

## Phase 1: Core Infrastructure (Weeks 1-3)

### 1.1 Project Setup & Build System ✅
**Priority**: CRITICAL
**Time**: 1-2 days
**Files**: `Makefile`, `go.mod`, `cmd/vimic2/main.go`

**Status**: Complete - scaffold created, cross-platform Makefile, entry point

### 1.2 Hypervisor Abstraction Layer ✅
**Priority**: CRITICAL
**Time**: 2-3 days
**Files**: `pkg/hypervisor/hypervisor.go`, `pkg/hypervisor/libvirt.go`

**Status**: Complete - interface defined, libvirt implementation, stub for dev

### 1.3 Cluster Data Model ✅
**Priority**: CRITICAL
**Time**: 1-2 days
**Files**: `internal/cluster/manager.go`

**Status**: Complete - Manager with full lifecycle operations

### 1.4 Database Layer ✅
**Priority**: HIGH
**Time**: 1-2 days
**Files**: `internal/database/db.go`

**Status**: Complete - SQLite with full CRUD, hosts/clusters/nodes/metrics

### 1.5 Basic UI - Dashboard ✅
**Priority**: HIGH
**Time**: 2-3 days
**Files**: `internal/ui/app.go`

**Status**: Complete - Fyne v2.4 compatible UI with cluster view, node management, deploy wizard

**Instructions**:
```bash
# Initialize Go module
go mod init github.com/stsgym/vimic2

# Add dependencies
go get github.com/fyne-io/fyne/v2@latest
go get github.com/mattn/go-sqlite3@latest
go get github.com/spf13/viper@latest
go get go.uber.org/zap@latest
go get github.com/google/uuid@latest

# Create Makefile with cross-platform builds
cat > Makefile << 'EOF'
.PHONY: build build-all build-linux build-windows build-macos test clean

BUILD_DIR := ./build
VERSION := 0.1.0

build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(BUILD_DIR)/vimic2 ./cmd/vimic2

build-all: build-linux build-windows build-macos

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BUILD_DIR)/vimic2-linux-amd64 ./cmd/vimic2

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(BUILD_DIR)/vimic2-windows-amd64.exe ./cmd/vimic2

build-macos:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(BUILD_DIR)/vimic2-darwin-amd64 ./cmd/vimic2
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(BUILD_DIR)/vimic2-darwin-arm64 ./cmd/vimic2

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)

run:
	go run ./cmd/vimic2
EOF
```

**Verification**: `make build-all` produces binaries for all platforms

---

### 1.2 Hypervisor Abstraction Layer
**Priority**: CRITICAL
**Time**: 2-3 days
**Files**: `pkg/hypervisor/hypervisor.go`, `pkg/hypervisor/libvirt.go`

**Instructions**:
1. Define core interfaces:
```go
// pkg/hypervisor/hypervisor.go
package hypervisor

import "context"

type Hypervisor interface {
    // Node lifecycle
    CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error)
    DeleteNode(ctx context.Context, id string) error
    StartNode(ctx context.Context, id string) error
    StopNode(ctx context.Context, id string) error
    RestartNode(ctx context.Context, id string) error
    
    // Node queries
    ListNodes(ctx context.Context) ([]*Node, error)
    GetNode(ctx context.Context, id string) (*Node, error)
    GetNodeStatus(ctx context.Context, id string) (*NodeStatus, error)
    
    // Metrics
    GetMetrics(ctx context.Context, id string) (*Metrics, error)
    
    Close() error
}

type NodeConfig struct {
    Name       string
    CPU        int
    MemoryMB   uint64
    DiskGB     int
    Image      string
    Network    string
    SSHKey     string
}

type NodeState string
const (
    NodePending   NodeState = "pending"
    NodeRunning   NodeState = "running"
    NodeStopped   NodeState = "stopped"
    NodeError     NodeState = "error"
)

type Node struct {
    ID       string
    Name     string
    State    NodeState
    IP       string
    Host     string
    Config   *NodeConfig
    Created  time.Time
}

type NodeStatus struct {
    State       NodeState
    Uptime      time.Duration
    CPUPercent  float64
    MemUsed     uint64
    MemTotal    uint64
    DiskUsedGB  float64
    DiskTotalGB float64
}

type Metrics struct {
    CPU    float64
    Memory float64
    Disk   float64
    Network float64
    Timestamp time.Time
}
```

2. Implement libvirt provider:
```go
// pkg/hypervisor/libvirt.go
package hypervisor

import (
    "libvirt.org/libvirt"
)

type LibvirtHypervisor struct {
    conn *libvirt.Connect
}

func NewLibvirt(addr string) (*LibvirtHypervisor, error) {
    conn, err := libvirt.NewConnect(addr)  // "qemu:///system" for local
    if err != nil {
        return nil, err
    }
    return &LibvirtHypervisor{conn: conn}, nil
}

func (h *LibvirtHypervisor) CreateNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
    // 1. Create qcow2 image from base
    // 2. Define domain XML
    // 3. Start domain
    // 4. Wait for DHCP lease
    // 5. Return Node info
}

func (h *LibvirtHypervisor) ListNodes(ctx context.Context) ([]*Node, error) {
    domains, err := h.conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE | libvirt.CONNECT_LIST_DOMAINS_INACTIVE)
    if err != nil {
        return nil, err
    }
    // Convert to Node structs
}
```

**Verification**: Can list, create, delete VMs on Linux host

---

### 1.3 Cluster Data Model
**Priority**: CRITICAL
**Time**: 1-2 days
**Files**: `internal/cluster/cluster.go`

**Instructions**:
```go
// internal/cluster/cluster.go
package cluster

type Cluster struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Nodes     []*NodeRef        `json:"nodes"`
    Config    *ClusterConfig    `json:"config"`
    Status    ClusterStatus     `json:"status"`
    Created   time.Time         `json:"created"`
    Updated   time.Time         `json:"updated"`
}

type ClusterConfig struct {
    MinNodes      int              `json:"min_nodes"`
    MaxNodes      int              `json:"max_nodes"`
    AutoScale     bool             `json:"autoscale"`
    ScaleOnCPU    float64          `json:"scale_on_cpu"`
    ScaleOnMemory float64          `json:"scale_on_memory"`
    Cooldown      time.Duration     `json:"cooldown"`
    Network       *NetworkConfig   `json:"network"`
    NodeDefaults  *NodeConfig      `json:"node_defaults"`
}

type NodeRef struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Role     string `json:"role"`
    HostID   string `json:"host_id"`
}

type ClusterStatus string
const (
    ClusterPending   ClusterStatus = "pending"
    ClusterDeploying ClusterStatus = "deploying"
    ClusterRunning   ClusterStatus = "running"
    ClusterDegraded  ClusterStatus = "degraded"
    ClusterError     ClusterStatus = "error"
}
```

**Verification**: Can create, serialize, deserialize cluster configs

---

### 1.4 Database Layer
**Priority**: HIGH
**Time**: 1-2 days
**Files**: `internal/database/db.go`

**Instructions**:
```bash
# Schema
CREATE TABLE hosts (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE,
    address TEXT,
    port INTEGER,
    user TEXT,
    ssh_key_path TEXT,
    hypervisor TEXT,
    created_at TIMESTAMP
);

CREATE TABLE clusters (
    id TEXT PRIMARY KEY,
    name TEXT,
    config TEXT,  -- JSON
    status TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    cluster_id TEXT,
    host_id TEXT,
    name TEXT,
    role TEXT,
    state TEXT,
    ip TEXT,
    config TEXT,  -- JSON
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

CREATE TABLE metrics (
    id INTEGER PRIMARY KEY,
    node_id TEXT,
    cpu REAL,
    memory REAL,
    disk REAL,
    network REAL,
    recorded_at TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES nodes(id)
);
```

**Verification**: CRUD operations work for hosts, clusters, nodes

---

### 1.5 Basic UI - Dashboard
**Priority**: HIGH
**Time**: 2-3 days
**Files**: `internal/ui/app.go`, `internal/ui/dashboard.go`

**Instructions**:
1. Create main application:
```go
// internal/ui/app.go
package ui

import (
    "github.com/stsgym/vimic2/internal/cluster"
    "github.com/stsgym/vimic2/internal/database"
    "github.com/stsgym/vimic2/pkg/hypervisor"
)

type App struct {
    db       *database.DB
    hosts    map[string]hypervisor.Hypervisor
    clusters *cluster.Manager
}

func NewApp(db *database.DB) *App {
    return &App{
        db:       db,
        hosts:    make(map[string]hypervisor.Hypervisor),
        clusters: cluster.NewManager(db),
    }
}
```

2. Create dashboard with Fyne:
```go
// internal/ui/dashboard.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

func (a *App) MakeDashboard() fyne.CanvasObject {
    // Left sidebar with hosts/clusters tree
    sidebar := widget.NewTreeList(
        // ... populate with hosts and clusters
    )
    
    // Main area with dashboard
    content := widget.NewVBox(
        widget.NewLabel("Dashboard"),
        a.statsGrid(),
    )
    
    return fyne.NewContainerWithLayout(
        layout.NewBorderLayout(sidebar, nil, nil, nil),
        sidebar,
        content,
    )
}
```

3. Add window with menu:
```go
// internal/ui/app.go
func (a *App) Run() error {
    w := fyneApp.NewWindow("Vimic2")
    w.SetContent(a.MakeDashboard())
    w.Resize(fyne.NewSize(1200, 800))
    w.ShowAndRun()
}
```

**Verification**: App launches, shows dashboard with sidebar

---

### 1.6 VM Lifecycle Operations ✅
**Priority**: CRITICAL
**Time**: 2-3 days
**Files**: `internal/cluster/manager.go`, `pkg/hypervisor/libvirt.go`

**Status**: Complete - CreateCluster, DeployCluster, ScaleCluster, StartNode, StopNode, RestartNode, DeleteNode all implemented

**Instructions**:
1. Implement ClusterManager:
```go
// internal/cluster/manager.go
package cluster

type Manager struct {
    db     *database.DB
    hosts  map[string]hypervisor.Hypervisor
}

func (m *Manager) CreateCluster(cfg *ClusterConfig) (*Cluster, error) {
    cluster := &Cluster{
        ID:     uuid.New().String(),
        Config: cfg,
        Status: ClusterPending,
    }
    if err := m.db.SaveCluster(cluster); err != nil {
        return nil, err
    }
    return cluster, nil
}

func (m *Manager) DeployCluster(id string) error {
    cluster, err := m.db.GetCluster(id)
    if err != nil {
        return err
    }
    
    cluster.Status = ClusterDeploying
    m.db.UpdateCluster(cluster)
    
    // Provision all nodes
    for _, nodeRef := range cluster.Nodes {
        err := m.provisionNode(cluster, nodeRef)
        if err != nil {
            return err
        }
    }
    
    cluster.Status = ClusterRunning
    return m.db.UpdateCluster(cluster)
}

func (m *Manager) provisionNode(cluster *Cluster, ref *NodeRef) error {
    // 1. Get host hypervisor
    hv := m.hosts[ref.HostID]
    
    // 2. Create VM
    node, err := hv.CreateNode(context.Background(), ref.Config)
    if err != nil {
        return err
    }
    
    // 3. Save to database
    return m.db.SaveNode(node)
}
```

2. Add UI buttons for operations:
```go
func (a *App) addNodeActions(node *cluster.NodeRef) fyne.CanvasObject {
    return widget.NewHBox(
        widget.NewButton("Start", func() { a.startNode(node.ID) }),
        widget.NewButton("Stop", func() { a.stopNode(node.ID) }),
        widget.NewButton("Delete", func() { a.deleteNode(node.ID) }),
    )
}
```

**Verification**: Can create cluster, deploy nodes, start/stop/delete nodes

---

## Phase 2: Cluster Management (Weeks 4-6) ✅ COMPLETE

### 2.1 Multi-Host Support ✅
**Priority**: HIGH
**Time**: 2-3 days
**Files**: `internal/host/manager.go`

**Status**: Complete - SSH connections, local/remote detection, host management
1. Add host registration:
```go
type HostManager struct {
    hosts map[string]hypervisor.Hypervisor
}

func (m *HostManager) AddHost(cfg *HostConfig) error {
    hv, err := hypervisor.NewHypervisor(cfg)
    if err != nil {
        return err
    }
    m.hosts[cfg.ID] = hv
    return nil
}
```

2. Implement host connection via SSH:
```go
// SSH tunnel for remote libvirt
type SSHConnection struct {
    Client  *ssh.Client
    Tunnel  net.Listener
    Remote  net.Conn
}
```

**Verification**: Can connect to multiple hosts, list nodes on each

---

### 2.2 Cluster Deployment Wizard ✅
**Priority**: HIGH
**Time**: 2-3 days
**Files**: `internal/deploy/wizard.go`

**Status**: Complete - 4-step wizard (Name, Template, Nodes, Review), preset templates, validation

### 2.3 Real-Time Status Updates ✅
**Priority**: MEDIUM
**Time**: 1-2 days
**Files**: `internal/status/watcher.go`

**Status**: Complete - Watcher with Subscribe/Unsubscribe, checkLoop, node/cluster updates, metrics collection

### 2.4 Node Detail View ✅
**Priority**: MEDIUM
**Time**: 1-2 days
**Files**: `internal/ui/app.go`

**Status**: Complete - Node detail view with status cards, metrics history, action buttons

---

## Phase 3: Monitoring & Metrics (Weeks 7-9) ✅ COMPLETE

### 3.1 Metrics Dashboard ✅
**Priority**: MEDIUM
**Time**: 2-3 days
**Files**: `internal/monitor/monitor.go`

**Status**: Complete - Manager with StartCollection, metrics collection loop, node metrics

### 3.2 Alerting System ✅
**Priority**: MEDIUM
**Time**: 2-3 days
**Files**: `internal/monitor/alerter.go`

**Status**: Complete - AlertRule, Alerter with Evaluate/ResolveAlert, DefaultRules (high-cpu, high-memory, disk-full, node-down)

### 3.3 Historical Data View ✅
**Priority**: LOW
**Time**: 1-2 days
**Files**: `internal/database/db.go`

**Status**: Complete - GetNodeMetrics(nodeID, since) for time-range queries, metrics stored in SQLite
```

**Verification**: Alert fires when CPU > 90% for 5 minutes

---

### 3.3 Historical Data View
**Priority**: LOW
**Time**: 2-3 days
**Files**: `internal/ui/history.go`

**Instructions**:
1. Query historical metrics from database
2. Render time-series charts
3. Add date range picker

**Verification**: Can view 24h/7d/30d metrics history

---

## Phase 4: Orchestration (Weeks 10-12) ✅ COMPLETE

### 4.1 Auto-Scaling
**Priority**: HIGH
**Time**: 3-4 days
**Files**: `internal/orchestrator/scaler.go`

**Instructions**:
```go
type AutoScaler struct {
    cluster    *Cluster
    rule       *ScaleRule
    cooldown   time.Time
}

type ScaleRule struct {
    Metric       string  // cpu, memory
    UpperThreshold float64   // Scale up when above
    LowerThreshold float64   // Scale down when below
    ScaleBy      int     // Number of nodes to add/remove
}

func (s *AutoScaler) Evaluate() error {
    if time.Since(s.cooldown) < s.cluster.Config.Cooldown {
        return nil  // In cooldown
    }
    
    avgCPU := s.getAverageCPU()
    
    if avgCPU > s.rule.UpperThreshold {
        return s.scaleUp()
    } else if avgCPU < s.rule.LowerThreshold {
        return s.scaleDown()
    }
    return nil
}

func (s *AutoScaler) scaleUp() error {
    if len(s.cluster.Nodes) >= s.cluster.Config.MaxNodes {
        return nil  // At max
    }
    // Add nodes
}
```

**Verification**: Cluster auto-scales when CPU > 70% for 5 minutes

---

### 4.2 Rolling Updates
**Priority**: MEDIUM
**Time**: 2-3 days
**Files**: `internal/orchestrator/updater.go`

**Instructions**:
```go
type RollingUpdater struct {
    Cluster    *Cluster
    NewImage   string
    BatchSize  int
}

func (u *RollingUpdater) Execute() error {
    nodes := u.Cluster.Nodes
    
    for i := 0; i < len(nodes); i += u.BatchSize {
        batch := nodes[i:i+u.BatchSize]
        
        // Upgrade batch
        for _, node := range batch {
            if err := u.upgradeNode(node); err != nil {
                return err
            }
        }
        
        // Wait for health
        if err := u.waitForHealthy(batch); err != nil {
            return err
        }
    }
    return nil
}
```

**Verification**: Can rolling-update nodes with zero downtime

---

### 4.3 Health Checks ✅
**Priority**: MEDIUM
**Time**: 1-2 days
**Files**: `internal/orchestrator/health.go`, `internal/status/watcher.go`

**Status**: Complete - HealthChecker with CheckAllNodes, CheckNode, node status tracking

### 4.4 Disaster Recovery ✅
**Priority**: LOW
**Time**: 2-3 days
**Files**: `internal/orchestrator/recovery.go`

**Status**: Complete - BackupNode struct, RecoveryManager with CreateBackup/RestoreBackup, Backup struct with JSON serialization

---

## 📋 Task Summary

| Phase | Task | Priority | Time | Status |
|-------|------|----------|------|--------|
| 1 | Project Setup | 🔴 CRITICAL | 1-2d | ✅ |
| 1 | Hypervisor Abstraction | 🔴 CRITICAL | 2-3d | ✅ |
| 1 | Cluster Model | 🔴 CRITICAL | 1-2d | ✅ |
| 1 | Database | 🟡 HIGH | 1-2d | ✅ |
| 1 | Dashboard UI | 🟡 HIGH | 2-3d | ✅ |
| 1 | VM Lifecycle | 🔴 CRITICAL | 2-3d | ✅ |
| 2 | Multi-Host | 🟡 HIGH | 2-3d | ✅ |
| 2 | Deploy Wizard | 🟡 HIGH | 2-3d | ✅ |
| 2 | Real-Time Updates | 🟡 MEDIUM | 1-2d | ✅ |
| 2 | Node Detail View | 🟡 MEDIUM | 1-2d | ✅ |
| 3 | Metrics Dashboard | 🟡 MEDIUM | 2-3d | ✅ |
| 3 | Alerting | 🟡 MEDIUM | 2-3d | ✅ |
| 3 | Historical Data | 🔵 LOW | 2-3d | ✅ |
| 4 | Auto-Scaling | 🟡 HIGH | 3-4d | ✅ |
| 4 | Rolling Updates | 🟡 MEDIUM | 2-3d | ✅ |
| 4 | Health Checks | 🟡 MEDIUM | 1-2d | ✅ |
| 4 | Disaster Recovery | 🔵 LOW | 2-3d | ✅ |

---

## 🏗️ Build Commands

```bash
# Development
make run          # Run from source
make test         # Run tests
make fmt          # Format code

# Production
make build             # Current platform
make build-all        # All platforms
make build-linux      # Linux only
make build-windows    # Windows only
make build-macos      # macOS only
```

---

## 📁 Key Files

```
vimic2/
├── cmd/vimic2/main.go           # Entry point
├── internal/
│   ├── cluster/
│   │   ├── cluster.go          # Cluster model
│   │   └── manager.go         # Cluster operations
│   ├── orchestrator/
│   │   ├── scaler.go          # Auto-scaling
│   │   └── updater.go         # Rolling updates
│   ├── monitor/
│   │   ├── collector.go       # Metrics collection
│   │   └── alerter.go         # Alerting
│   ├── database/
│   │   └── db.go              # SQLite persistence
│   └── ui/
│       ├── app.go             # Main app
│       ├── dashboard.go       # Dashboard view
│       ├── cluster.go         # Cluster view
│       ├── node.go            # Node detail
│       └── deploy.go          # Deploy wizard
├── pkg/
│   └── hypervisor/
│       ├── hypervisor.go      # Interface
│       └── libvirt.go         # Linux impl
├── Makefile
└── go.mod
```

---

*Vimic2 Phase 1 Roadmap - Cluster Management Focus*
