# Vimic2 Architecture

## Overview

Vimic2 is a cross-platform desktop application for managing VM clusters. It provides a unified interface for provisioning, monitoring, and orchestrating virtual machines across multiple hypervisor hosts.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                        UI Layer (Fyne)                        │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────────────┐  │
│  │Dashboard│ │ClusterMgr│ │ NodeView │ │ AlertManager   │  │
│  └─────────┘ └──────────┘ └──────────┘ └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                     Business Logic Layer                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────────┐  │
│  │ Cluster  │ │  Deploy  │ │ Monitor  │ │  Orchestrator │  │
│  │ Manager  │ │ Wizard   │ │ Manager  │ │   (Scaler)    │  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                      Data Layer                              │
│  ┌──────────────┐ ┌────────────────┐ ┌──────────────────┐   │
│  │   SQLite     │ │   Status       │ │    Provisioner   │   │
│  │  Database    │ │   Watcher      │ │    (Images)     │   │
│  └──────────────┘ └────────────────┘ └──────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                   Hypervisor Abstraction                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────────┐  │
│  │ Libvirt  │ │ Hyper-V  │ │ AppleVMM │ │    Stub       │  │
│  │ (Linux)  │ │(Windows) │ │ (macOS)  │ │  (Testing)   │  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Component Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                           App (main.go)                           │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                        App struct                           │ │
│  │  - window: fyne.Window                                      │ │
│  │  - db: *database.DB                                         │ │
│  │  - clusterMgr: *cluster.Manager                            │ │
│  │  - monitorMgr: *monitor.Manager                            │ │
│  │  - autoScaler: *orchestrator.AutoScaler                    │ │
│  └────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        ▼                           ▼                           ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│ cluster.Manager│          │ monitor.Manager│          │orchestrator   │
│               │          │               │          │               │
│ - CreateCluster│          │ - Collect()   │          │ - AutoScaler │
│ - DeleteCluster│          │ - GetMetrics │          │ - Updater    │
│ - ScaleCluster │          │ - Alerter    │          │ - Recovery   │
│ - DeployCluster│          │               │          │               │
└───────┬───────┘          └───────┬───────┘          └───────┬───────┘
        │                           │                           │
        ▼                           ▼                           │
┌───────────────┐          ┌───────────────┐                     │
│ database.DB   │          │  status.     │                     │
│               │          │  Watcher     │                     │
│ - hosts       │          │               │                     │
│ - clusters    │          │ - Subscribe   │                     │
│ - nodes       │          │ - Broadcast   │                     │
│ - metrics     │          │               │                     │
│ - alerts      │          └───────┬───────┘                     │
└───────────────┘                  │                             │
                                   ▼                             │
                    ┌──────────────────────────────┐              │
                    │       WebSocketHub          │              │
                    │  (real-time updates)        │              │
                    └──────────────────────────────┘              │
                                                                 │
        ┌───────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│                    Hypervisor Interface                         │
│                                                               │
│  type Hypervisor interface {                                  │
│      CreateNode(ctx, *NodeConfig) (*Node, error)             │
│      DeleteNode(ctx, id) error                                 │
│      StartNode(ctx, id) error                                 │
│      StopNode(ctx, id) error                                  │
│      RestartNode(ctx, id) error                              │
│      ListNodes(ctx) ([]*Node, error)                          │
│      GetNode(ctx, id) (*Node, error)                         │
│      GetNodeStatus(ctx, id) (*NodeStatus, error)             │
│      GetMetrics(ctx, id) (*Metrics, error)                    │
│      Close() error                                            │
│  }                                                           │
└───────────────────────────────────────────────────────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│                   Hypervisor Implementations                    │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐ │
│  │  Libvirt    │ │  Windows    │ │   Apple    │ │   Stub   │ │
│  │  (Linux)   │ │   (Hyper-V) │ │   (VMM)    │ │(Testing) │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘ │
└───────────────────────────────────────────────────────────────┘
```

## Data Model

### Database Schema

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│    hosts      │       │   clusters    │       │    nodes     │
├──────────────┤       ├──────────────┤       ├──────────────┤
│ id (PK)      │       │ id (PK)      │       │ id (PK)      │
│ name         │       │ name         │◄──┐   │ cluster_id   │──►
│ address      │       │ config (JSON)│   │   │ host_id      │──►
│ port         │       │ status       │   │   │ name         │
│ user         │       │ created_at   │   │   │ role         │
│ ssh_key_path │       │ updated_at   │   │   │ state        │
│ hv_type      │       └──────────────┘   │   │ ip           │
└──────────────┘                          │   │ config (JSON)│
                                          │   └──────────────┘
                                          │          │
┌──────────────┐                          │          │
│   metrics    │                          │          │
├──────────────┤                          │          │
│ id (PK)      │                          │          │
│ node_id (FK) │◄─────────────────────────┘          │
│ cpu          │                                    │
│ memory       │                                    │
│ disk         │                                    │
│ network_rx   │                                    │
│ network_tx   │                                    │
│ recorded_at  │                                    │
└──────────────┘                                    │

┌──────────────┐
│   alerts     │
├──────────────┤
│ id (PK)      │
│ rule_id      │
│ node_id (FK) │──►
│ node_name    │
│ metric       │
│ value        │
│ threshold    │
│ message      │
│ fired_at     │
│ resolved     │
│ resolved_at  │
└──────────────┘
```

## Key Flows

### 1. Create and Deploy Cluster

```
User Action: Click "New Cluster"
        │
        ▼
┌─────────────────┐
│ Deploy Wizard   │
│ Step 1-4       │
└────────┬────────┘
         │ On "Deploy"
         ▼
┌─────────────────┐
│ cluster.Manager │
│ .CreateCluster()│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Executor.Execute│
│ - For each node│
│   - Get host   │
│   - Create VM  │
│   - Save node  │
│   - Update DB  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  hypervisor     │
│ .CreateNode()   │
│ - Create disk   │
│ - Define XML    │
│ - Start VM     │
│ - Wait for IP  │
└────────┬────────┘
         │
         ▼
    ┌────────┐
    │ Status │
    │ Updated│
    └────────┘
```

### 2. Metrics Collection

```
┌─────────────────────────────────────────────────────────┐
│                  Background goroutine                     │
│                   (5 second tick)                        │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   monitor.Manager                        │
│  For each cluster                                       │
│    For each node (if running)                          │
│      - Collect metrics via hypervisor                   │
│      - Save to database                                │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   UI Updates                             │
│  - Status watcher notifies subscribers                  │
│  - Dashboard refreshes                                 │
│  - Charts update                                       │
└─────────────────────────────────────────────────────────┘
```

### 3. Auto-scaling

```
┌─────────────────────────────────────────────────────────┐
│                AutoScaler.runLoop()                     │
│                   (30 second tick)                      │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   Evaluate all rules                     │
│                                                         │
│  For each cluster with auto-scaling enabled:           │
│    1. Get current metrics (avg CPU/memory)             │
│    2. Check against thresholds                        │
│    3. If upper threshold exceeded + cooldown passed:   │
│       - Scale up (add nodes)                          │
│    4. If lower threshold exceeded + cooldown passed:   │
│       - Scale down (remove nodes)                      │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  cluster.ScaleCluster()                 │
│                                                         │
│  Scale Up:                                              │
│    - Create new nodes                                  │
│    - Add to cluster                                    │
│    - Provision on hosts                                │
│                                                         │
│  Scale Down:                                           │
│    - Select nodes to remove                            │
│    - Stop and delete VMs                              │
│    - Remove from cluster                              │
└─────────────────────────────────────────────────────────┘
```

## Package Structure

```
vimic2/
├── cmd/
│   └── vimic2/
│       └── main.go          # Entry point
│
├── internal/
│   ├── cluster/
│   │   └── manager.go       # Cluster lifecycle
│   │
│   ├── database/
│   │   └── db.go            # SQLite persistence
│   │
│   ├── deploy/
│   │   └── wizard.go        # Deployment orchestration
│   │
│   ├── host/
│   │   └── manager.go       # Multi-host management
│   │
│   ├── monitor/
│   │   ├── monitor.go       # Metrics collection
│   │   └── alerter.go      # Alert evaluation
│   │
│   ├── orchestrator/
│   │   ├── scaler.go        # Auto-scaling
│   │   ├── updater.go       # Rolling updates
│   │   └── recovery.go      # Backup/restore
│   │
│   ├── provisioner/
│   │   └── provisioner.go   # VM image management
│   │
│   ├── status/
│   │   └── watcher.go       # Real-time status
│   │
│   └── ui/
│       └── app.go           # Fyne UI
│
├── pkg/
│   └── hypervisor/
│       ├── hypervisor.go    # Interface definition
│       ├── libvirt.go       # Linux implementation
│       ├── apple.go         # macOS implementation
│       └── windows.go       # Windows implementation
│
├── docs/
│   ├── QUICKSTART.md
│   ├── USER_GUIDE.md
│   └── ARCHITECTURE.md      # This file
│
├── Makefile                 # Build targets
├── go.mod                   # Dependencies
└── README.md
```

## Dependencies

```go
require (
    github.com/fyne-io/fyne/v2    // Cross-platform UI
    github.com/mattn/go-sqlite3   // SQLite driver
    github.com/spf13/viper        // Configuration
    github.com/google/uuid         // UUID generation
    go.uber.org/zap               // Logging
    golang.org/x/crypto/ssh       // SSH connections
)
```

## Configuration

Vimic2 uses Viper for configuration:

```yaml
# ~/.vimic2/config.yaml
db_path: ~/.vimic2/vimic2.db

hosts:
  local:
    address: localhost
    port: 22
    user: root
    hv_type: libvirt

monitor:
  interval: 5s
  retention: 24h

autoscale:
  enabled: true
  cpu_threshold: 70
  memory_threshold: 80
  cooldown: 5m

logging:
  level: info
  file: ~/.vimic2/logs/vimic2.log
```

## Concurrency Model

- **Main goroutine**: UI event loop (Fyne)
- **Metrics collection**: Background goroutine with ticker
- **Auto-scaling**: Background goroutine with ticker
- **Status watcher**: Background goroutine with ticker
- **WebSocket hub**: Runs its own goroutine for client management

All shared state is protected by mutexes where needed.

## Error Handling

- Errors are logged via zap logger
- UI shows user-friendly error dialogs
- Database operations are wrapped with error checking
- Hypervisor operations have fallbacks where possible

## Testing Strategy

- **Unit tests**: Each package has `*_test.go`
- **Stub hypervisor**: For testing without real VMs
- **In-memory DB**: SQLite with temporary files for tests

## Future Considerations

- [ ] Web UI (browser-based)
- [ ] REST API server
- [ ] Multi-tenancy
- [ ] LDAP/SSO integration
- [ ] Distributed metrics (Prometheus)
- [ ] Container support (Docker, Kubernetes)
