# Testing Guide for vimic2/internal/realutil

This document provides a comprehensive guide to testing the `realutil` package, explaining each test category and its real-world use cases.

## Overview

The `realutil` package provides production-ready implementations of:
- **realdb**: SQLite database operations for cluster/host/node management
- **realfs**: Filesystem operations with atomic writes and file locking
- **realhv**: Hypervisor (libvirt) integration for VM management
- **realovs**: Open vSwitch client for network virtualization

**Current Test Coverage: 93.5%**

| Package | Coverage | Focus |
|---------|----------|-------|
| realovs | 96.9% | Network virtualization |
| realhv | 95.3% | VM lifecycle management |
| realfs | 95.1% | File I/O and locking |
| realdb | 89.5% | Persistent state management |

---

## Running Tests

### Unit Tests (Fast)

```bash
# Run all tests
go test ./internal/realutil/... -v

# Run with coverage
go test ./internal/realutil/... -cover

# Run specific package
go test ./internal/realutil/realdb/... -v
```

### Integration Tests (Requires External Services)

```bash
# Run with integration tags (requires libvirt, OVS)
go test -tags="integration,libvirt" ./internal/realutil/... -v

# Generate coverage report
go test -tags="integration,libvirt" ./internal/realutil/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## Package: realdb

### Purpose
SQLite-based persistent storage for vimic2 cluster state, including:
- Clusters (VM groups)
- Hosts (physical machines)
- Nodes (individual VMs)
- Metrics (performance data)
- Alerts (system notifications)
- Pools (resource pools)

### Test Categories

#### 1. Database Lifecycle Tests

**Tests:** `TestRealDatabase_NewDatabase_*`

```go
// Example: Creating database with custom config
cfg := &Config{
    Path:         "/var/lib/vimic2/state.db",
    MaxOpenConns: 25,
    MaxIdleConns: 10,
    BusyTimeout:  15 * time.Second,
    WALMode:      true,
}
db, err := NewDatabase(cfg)
```

**Real-World Use Case:**
- Production deployment with custom connection pooling
- High-concurrency environments needing WAL mode
- Embedded systems with limited resources (adjust connection limits)

**What We Test:**
- Database creation with various paths (`:memory:`, file paths)
- Connection pool configuration
- WAL mode enablement
- Busy timeout for concurrent access
- Migration execution

#### 2. CRUD Operations Tests

**Tests:** `TestRealDatabase_SaveCluster_*`, `TestRealDatabase_GetCluster_*`, `TestRealDatabase_DeleteCluster`

```go
// Create cluster with config
cluster := &Cluster{
    ID:     "prod-cluster",
    Name:   "production",
    Status: "running",
    Config: &ClusterConfig{
        MinNodes:  3,
        MaxNodes:  10,
        AutoScale: true,
    },
}
db.SaveCluster(cluster)

// Retrieve cluster
retrieved, err := db.GetCluster("prod-cluster")

// Update cluster (upsert)
cluster.Status = "maintenance"
db.SaveCluster(cluster)

// Delete cluster
db.DeleteCluster("prod-cluster")
```

**Real-World Use Case:**
- Infrastructure-as-Code tools storing desired state
- Cluster autoscaler tracking node counts
- Disaster recovery tracking cluster health

**What We Test:**
- Insert operations
- Upsert (INSERT OR REPLACE) for updates
- Query by ID
- Delete operations
- JSON serialization of nested configs

#### 3. Query Operations Tests

**Tests:** `TestRealDatabase_ListClusters_*`, `TestRealDatabase_ListClusterNodes_*`

```go
// List all clusters
clusters, err := db.ListClusters()

// List nodes in specific cluster
nodes, err := db.ListClusterNodes("prod-cluster")

// Filter metrics by time
metrics, err := db.GetNodeMetrics("node-1", time.Now().Add(-24*time.Hour))
```

**Real-World Use Case:**
- Dashboard displaying all clusters
- Monitoring specific cluster's VMs
- Historical metric analysis for capacity planning

**What We Test:**
- Empty result handling
- Multiple entity retrieval
- Time-based filtering
- JSON deserialization

#### 4. Metrics and Alerts Tests

**Tests:** `TestRealDatabase_SaveMetric_*`, `TestRealDatabase_GetNodeMetrics_*`

```go
// Record metric
metric := &Metric{
    NodeID:     "node-1",
    CPU:        75.5,
    Memory:     80.2,
    Disk:       45.0,
    NetworkRX:  1024000.0,
    NetworkTX:  512000.0,
    RecordedAt: time.Now(),
}
db.SaveMetric(metric)

// Create alert
alert := &Alert{
    NodeID:   "node-1",
    Type:     "cpu_high",
    Message:  "CPU exceeded 90%",
    Severity: "critical",
}
db.SaveAlert(alert)
```

**Real-World Use Case:**
- Prometheus-style metric collection
- Alerting system integration
- Capacity planning with historical data

**What We Test:**
- Auto-increment ID assignment
- Time-series data retrieval
- Alert lifecycle management

#### 5. Maintenance Tests

**Tests:** `TestRealDatabase_Backup`, `TestRealDatabase_Vacuum_*`, `TestRealDatabase_IntegrityCheck_*`

```go
// Backup database
db.Backup("/backup/vimic2-2024-04-11.db")

// Vacuum (optimize after deletions)
db.Vacuum()

// Check integrity
if err := db.IntegrityCheck(); err != nil {
    log.Fatal("Database corruption detected")
}
```

**Real-World Use Case:**
- Scheduled backup automation
- Post-deletion cleanup
- Startup integrity verification

**What We Test:**
- File-based backup creation
- SQLite VACUUM operation
- PRAGMA integrity_check

---

## Package: realfs

### Purpose
Production filesystem operations with:
- Atomic writes (write to temp, then rename)
- File locking for concurrent access
- Safe directory creation

### Test Categories

#### 1. Atomic Write Tests

**Tests:** `TestRealFilesystem_WriteFile_*`

```go
fs := NewFilesystem()

// Atomic write (creates parent dirs automatically)
err := fs.WriteFile("/var/lib/vimic2/config.json", []byte(data), 0644)
```

**Real-World Use Case:**
- Configuration file updates (no partial writes)
- State file persistence (crash-safe)
- Log rotation

**What We Test:**
- Temp file creation
- Atomic rename on POSIX
- Parent directory creation
- Permission handling

#### 2. File Locking Tests

**Tests:** `TestRealFilesystem_Lock_*`, `TestRealFilesystem_TryLock_*`

```go
fs := NewFilesystem()

// Exclusive lock (blocking)
lock, err := fs.Lock("/var/run/vimic2.pid")
defer lock.Unlock()

// Non-blocking lock attempt
lock, err := fs.TryLock("/var/run/vimic2.pid")
if err != nil {
    // Another instance is running
    return ErrAlreadyRunning
}
```

**Real-World Use Case:**
- Single-instance enforcement (PID files)
- Shared resource protection
- Distributed lock coordination

**What We Test:**
- Lock acquisition
- Double-lock prevention
- Unlock behavior
- Nested directory creation

#### 3. File Operations Tests

**Tests:** `TestRealFilesystem_Copy_*`, `TestRealFilesystem_ReadDir_*`

```go
// Copy file
err := fs.Copy("/source/config.yaml", "/dest/config.yaml")

// List directory
entries, err := fs.ReadDir("/var/lib/vimic2")

// Check existence
if fs.Exists("/var/run/vimic2.pid") {
    // Already running
}
```

**Real-World Use Case:**
- Configuration deployment
- Backup file creation
- Directory scanning for cleanup

**What We Test:**
- Large file copying
- Directory listing
- Existence checks

---

## Package: realhv

### Purpose
libvirt integration for VM lifecycle management:
- Connect to hypervisor (local/remote)
- Create, start, stop, delete VMs
- Query VM status and metrics

### Test Categories

#### 1. Connection Tests

**Tests:** `TestRealHypervisor_Connect_*`

```go
cfg := &Config{
    URI:         "qemu+ssh://10.0.0.117/system",
    Timeout:     30 * time.Second,
    AutoConnect: true,
}
hv := NewHypervisor(cfg)

ctx := context.Background()
if err := hv.Connect(ctx); err != nil {
    log.Fatal("Failed to connect to hypervisor")
}
defer hv.Disconnect()
```

**Real-World Use Case:**
- Remote VM management over SSH
- Local libvirt socket connection
- Cloud hypervisor integration

**What We Test:**
- Local connection (`qemu:///system`)
- Remote SSH connection (`qemu+ssh://host/system`)
- Apple virtualization framework
- Auto-connect behavior

#### 2. VM Lifecycle Tests

**Tests:** `TestRealHypervisor_CreateNode_*`, `TestRealHypervisor_StartNode_*`

```go
// Create VM
vmCfg := &VMConfig{
    Name:     "vm-001",
    CPU:      4,
    MemoryMB: 8192,
    DiskGB:   100,
    Image:    "ubuntu-22.04.qcow2",
    Network:  "default",
}
vm, err := hv.CreateNode(ctx, vmCfg)

// Start VM
err := hv.StartNode(ctx, vm.ID)

// Stop VM
err := hv.StopNode(ctx, vm.ID)

// Delete VM
err := hv.DeleteNode(ctx, vm.ID)
```

**Real-World Use Case:**
- CI/CD pipeline VM provisioning
- Development environment creation
- Autoscaler node management

**What We Test:**
- VM creation with various configs
- Start/stop/restart operations
- Deletion with cleanup

#### 3. Query Operations Tests

**Tests:** `TestRealHypervisor_ListNodes_*`, `TestRealHypervisor_GetNodeStatus_*`

```go
// List all VMs
vms, err := hv.ListNodes(ctx)

// Get specific VM
vm, err := hv.GetNode(ctx, "vm-001")

// Get VM status
status, err := hv.GetNodeStatus(ctx, "vm-001")

// Get VM metrics
metrics, err := hv.GetMetrics(ctx, "vm-001")
```

**Real-World Use Case:**
- Dashboard VM inventory
- Health monitoring
- Performance analysis

**What We Test:**
- List operations
- Status retrieval
- Metrics collection

---

## Package: realovs

### Purpose
Open vSwitch client for:
- Bridge/port management
- Flow rule configuration
- VXLAN/GRE tunnel setup

### Test Categories

#### 1. Bridge Management Tests

**Tests:** `TestRealOVS_CreateBridge_*`, `TestRealOVS_ListBridges_*`

```go
c := NewClientWithDefaults()

// Create bridge
err := c.CreateBridge("br-int")

// Set VLAN
err := c.SetBridgeVLAN("br-int", 100)

// List bridges
bridges, err := c.ListBridges()
```

**Real-World Use Case:**
- SDN controller integration
- Network isolation setup
- Multi-tenant network creation

**What We Test:**
- Bridge creation/deletion
- VLAN configuration
- Trunk port setup

#### 2. Port Management Tests

**Tests:** `TestRealOVS_AddPort_*`, `TestRealOVS_SetPortVLAN_*`

```go
// Add port to bridge
err := c.AddPort("br-int", "vnet0")

// Set port VLAN
err := c.SetPortVLAN("vnet0", 100)

// Set QoS
err := c.SetPortQoS("vnet0", 1000) // 1Gbps limit

// Set port security
err := c.SetPortSecurity("vnet0", "00:11:22:33:44:55", "10.0.0.100")
```

**Real-World Use Case:**
- VM network attachment
- Tenant isolation
- Bandwidth limiting
- Anti-spoofing protection

**What We Test:**
- Port add/delete
- VLAN tagging
- QoS configuration
- Port security

#### 3. Flow Rule Tests

**Tests:** `TestRealOVS_AddFlow_*`, `TestRealOVS_ListFlows_*`

```go
// Add flow rule
err := c.AddFlow("br-int", 100, "in_port=1", "output:2")

// List flows
flows, err := c.ListFlows("br-int")

// Clear all flows
err := c.ClearFlows("br-int")
```

**Real-World Use Case:**
- Custom routing rules
- Traffic shaping
- Security policy enforcement

**What We Test:**
- Flow rule creation
- Flow listing/parsing
- Flow deletion

---

## Testing Patterns

### 1. Table-Driven Tests

Use for testing multiple inputs:

```go
func TestRealDatabase_NewDatabase_PathVariations(t *testing.T) {
    tests := []struct {
        name string
        path string
    }{
        {"memory", ":memory:"},
        {"empty", ""},
        {"relative", "test.db"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            cfg := &Config{Path: tc.path}
            db, err := NewDatabase(cfg)
            // ...
        })
    }
}
```

### 2. Cleanup Pattern

Always defer cleanup:

```go
func TestRealDatabase_SaveCluster(t *testing.T) {
    db, err := NewDatabaseWithDefaults(":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()  // Always cleanup

    // Test code...
}
```

### 3. Subtests for Organization

Group related tests:

```go
func TestRealFilesystem_Lock(t *testing.T) {
    t.Run("acquire", func(t *testing.T) { /* ... */ })
    t.Run("double_lock", func(t *testing.T) { /* ... */ })
    t.Run("unlock", func(t *testing.T) { /* ... */ })
}
```

---

## Integration Testing

### Prerequisites

```bash
# Install libvirt
sudo apt install libvirt-daemon-system libvirt-dev

# Install OVS
sudo apt install openvswitch-switch

# Add user to groups
sudo usermod -aG libvirt,libvirt-qemu $USER
```

### Running Integration Tests

```bash
# Requires libvirt running
go test -tags="integration,libvirt" ./internal/realutil/realhv/... -v

# Requires OVS running
go test -tags="integration" ./internal/realutil/realovs/... -v
```

---

## Coverage Gaps and Limitations

### Uncovered Error Paths (~6-7%)

These require mocking external dependencies:

| Package | Gap | Reason |
|---------|-----|--------|
| realdb | `sql.Open` errors | SQLite driver mocking needed |
| realdb | `db.Exec` failures | Database corruption simulation |
| realhv | libvirt connection failures | libvirt mock needed |
| realovs | `exec.Command` failures | OVS service mocking |

### Adding Mock Tests

To reach 100%, add these mock frameworks:

```go
import (
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/golang/mock/gomock"
)

// Example: Mocking SQLite
func TestRealDatabase_NewDatabase_SQLiteError(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    mock.ExpectExec("PRAGMA").WillReturnError(errors.New("disk full"))
    // ...
}
```

---

## Best Practices

### 1. Test Real Behavior

```go
// Good: Tests actual behavior
func TestRealDatabase_SaveCluster(t *testing.T) {
    db, _ := NewDatabaseWithDefaults(":memory:")
    defer db.Close()
    
    cluster := &Cluster{ID: "test"}
    err := db.SaveCluster(cluster)
    if err != nil {
        t.Errorf("SaveCluster failed: %v", err)
    }
}

// Bad: Tests implementation details
func TestRealDatabase_internalFields(t *testing.T) {
    db, _ := NewDatabaseWithDefaults(":memory:")
    if db.path != ":memory:" {  // Don't test private fields
        t.Error("wrong path")
    }
}
```

### 2. Test Edge Cases

```go
func TestRealDatabase_SaveMetric_BoundaryValues(t *testing.T) {
    tests := []struct {
        name   string
        metric *Metric
    }{
        {"zero", &Metric{CPU: 0}},
        {"negative", &Metric{CPU: -1}},
        {"max", &Metric{CPU: 1e10}},
    }
    // ...
}
```

### 3. Document Why, Not What

```go
// Good: Explains why
// TestRealDatabase_IntegrityCheck_AfterOperations verifies
// database integrity after CRUD operations, catching potential
// corruption from transaction issues.

// Bad: Just repeats the code
// TestRealDatabase_IntegrityCheck tests IntegrityCheck.
```

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Install dependencies
        run: |
          sudo apt update
          sudo apt install -y libvirt-dev openvswitch-switch
      
      - name: Run unit tests
        run: go test ./internal/realutil/... -cover
      
      - name: Run integration tests
        run: go test -tags="integration,libvirt" ./internal/realutil/... -v
```

---

## Troubleshooting

### Common Issues

**Q: Tests fail with "database is locked"**
```
A: SQLite has locking issues in tests. Use :memory: for tests,
   or increase BusyTimeout.
```

**Q: OVS tests fail with "permission denied"**
```
A: Run tests with sudo, or add user to ovs group:
   sudo usermod -aG ovs $USER
```

**Q: libvirt tests fail with "connection refused"**
```
A: Ensure libvirtd is running:
   sudo systemctl start libvirtd
```

---

## Summary

This testing guide covers:
- **Unit tests** for all packages (fast, no dependencies)
- **Integration tests** for external services (libvirt, OVS)
- **Real-world use cases** for each test category
- **Best practices** for Go testing
- **CI/CD integration** examples

The 93.5% coverage achieved provides confidence in production behavior, with remaining gaps documented as requiring external mocking frameworks.