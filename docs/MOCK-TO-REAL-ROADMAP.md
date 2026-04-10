# Mock to Real Implementation Roadmap

## Overview

Convert 4 mock utilities to production implementations for integration testing and real workloads.

**Status: COMPLETE ✅**

| Mock | Real Implementation | Complexity | Status | Coverage |
|------|---------------------|------------|--------|---------|
| mockfs | Real filesystem (os package) | Low | ✅ Complete | 69.9% |
| mockdb | SQLite database | Medium | ✅ Complete | 85.1% |
| mockovs | Open vSwitch CLI (ovs-vsctl) | Medium | ✅ Complete | 43.8% |
| mockhv | libvirt/QEMU hypervisor | High | ✅ Complete | 37.6% |

---

## Phase 1: Real Filesystem (2-3 hours)

### Why Easiest First
- No external dependencies
- Go stdlib has everything needed
- Tests can run without setup
- Immediate value for template manager tests

### Tasks
- [x] Create `internal/realutil/realfs/filesystem.go`
- [x] Implement interface matching mockfs
- [x] Add atomic write support (write to temp, rename)
- [x] Add file locking for concurrent access
- [x] Add symlink support
- [x] Add permission management
- [x] Write tests (20 tests)
- [x] Verified: 69.9% coverage

### Key Methods
```go
type Filesystem interface {
    MkdirAll(path string, perm os.FileMode) error
    WriteFile(filename string, data []byte, perm os.FileMode) error
    ReadFile(filename string) ([]byte, error)
    Remove(path string) error
    Stat(filename string) (os.FileInfo, error)
    ReadDir(dirname string) ([]os.FileInfo, error)
    Exists(filename string) bool
    Copy(src, dst string) error
    Move(src, dst string) error
    Symlink(oldname, newname string) error
    Chmod(filename string, perm os.FileMode) error
    Chown(filename string, uid, gid int) error
}
```

### Estimated Lines: ~300 LOC

---

## Phase 2: Real Database (4-6 hours) ✅ COMPLETE

### Why Second
- Project already uses SQLite in internal/database
- Can leverage existing schema patterns
- Need connection pooling and migration support

### Tasks
- [x] Create `internal/realutil/realdb/database.go`
- [x] Implement interface matching mockdb
- [x] Add connection pooling (sqlx)
- [x] Add migration support
- [x] Add transaction support
- [x] Add query builder helpers
- [x] Add health check endpoint
- [x] Write tests (16 tests)
- [x] Add backup/restore utilities
- [x] Verified: 85.1% coverage

### Schema Design
```sql
CREATE TABLE clusters (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    config JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE hosts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    port INTEGER DEFAULT 22,
    user TEXT DEFAULT 'root',
    ssh_key_path TEXT,
    hv_type TEXT DEFAULT 'libvirt',
    created_at TIMESTAMP
);

CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    cluster_id TEXT REFERENCES clusters(id),
    host_id TEXT REFERENCES hosts(id),
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    state TEXT NOT NULL,
    ip TEXT,
    config JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id TEXT REFERENCES nodes(id),
    cpu REAL,
    memory REAL,
    disk REAL,
    network_rx REAL,
    network_tx REAL,
    recorded_at TIMESTAMP
);

CREATE INDEX idx_metrics_node_time ON metrics(node_id, recorded_at);
```

### Estimated Lines: ~800 LOC

---

## Phase 3: Real OVS Client (6-8 hours) ✅ COMPLETE

### Why Third
- Requires OVS installed for testing
- Command-line integration is straightforward
- Used by network isolation layer

### Tasks
- [x] Create `internal/realutil/realovs/ovs.go`
- [x] Implement interface matching mockovs
- [x] Add command execution with timeout
- [x] Add output parsing
- [x] Add error handling and retries
- [x] Add flow rule management
- [x] Add tunnel management (VXLAN, GRE, Geneve)
- [x] Add QoS configuration
- [x] Add port security (MAC/IP anti-spoofing)
- [x] Write integration tests (10 tests with real OVS)
- [x] Add health check
- [x] Verified: 43.8% coverage (25 unit + 10 integration tests)

### Key Commands
```go
// Bridge management
ovs-vsctl add-br <bridge>
ovs-vsctl del-br <bridge>
ovs-vsctl list-br

// Port management
ovs-vsctl add-port <bridge> <port>
ovs-vsctl del-port <bridge> <port>
ovs-vsctl list-ports <bridge>

// VLAN/Trunk
ovs-vsctl set port <port> tag=<vlan>
ovs-vsctl set port <port> trunks=<vlan1,vlan2>

// QoS
ovs-vsctl set interface <port> ingress_policing_rate=<rate>
ovs-vsctl set port <port> qos=@newqos -- --id=@newqos create qos type=linux-htb

// Flows
ovs-ofctl add-flow <bridge> <flow>
ovs-ofctl del-flows <bridge> <flow>
ovs-ofctl dump-flows <bridge>

// Tunnels
ovs-vsctl add-port <bridge> <name> -- set interface <name> type=vxlan options:remote_ip=<ip> options:key=<vni>
```

### Estimated Lines: ~600 LOC

---

## Phase 4: Real Hypervisor (8-12 hours) ✅ COMPLETE

### Why Last
- Most complex, requires libvirt
- Needs QEMU/KVM for full functionality
- Integration tests require VM infrastructure

### Tasks
- [x] Create `internal/realutil/realhv/hypervisor.go`
- [x] Implement interface matching mockhv
- [x] Add libvirt connection management
- [x] Add VM lifecycle (create, start, stop, delete)
- [x] Add resource monitoring (CPU, memory, disk, network)
- [x] Add connection pooling
- [x] Write tests (12 tests)
- [x] Verified: 37.6% coverage

### Key Libvirt Operations
```go
// Connection
conn, err := libvirt.NewConnect("qemu:///system")

// VM Creation
dom, err := conn.DomainDefineXML(xml)
err = dom.Create()

// VM Operations
dom.Shutdown()
dom.Reboot()
dom.Destroy()
dom.Undefine()

// Metrics
info, _ := dom.GetInfo()
cpuStats, _ := dom.GetCPUStats()
memStats, _ := dom.GetMemoryStats()

// Snapshots
snap, _ := domain.SnapshotCreateXML(xml, flags)
snaps, _ := domain.ListAllSnapshots()

// Migration
dom.MigrateToURI3(destURI, params, flags)
```

### Estimated Lines: ~1000 LOC

---

## Testing Strategy

### Unit Tests (No External Deps)
- All 4 real utilities have unit tests using interfaces
- Test against mock implementations first
- Validate interface compliance

### Integration Tests (Requires Infrastructure)
| Utility | Requirement | Setup |
|---------|-------------|-------|
| realfs | None | N/A |
| realdb | SQLite | N/A (embedded) |
| realovs | Open vSwitch | `apt install openvswitch-switch` |
| realhv | libvirt + QEMU | `apt install libvirt-daemon qemu-kvm` |

### Test Containers
```yaml
# docker-compose.test.yml
services:
  ovs:
    image: openvswitch/ovs
    privileged: true
  
  libvirt:
    image: libvirt/libvirt
    privileged: true
    volumes:
      - /var/run/libvirt:/var/run/libvirt
```

---

## Timeline

**Completed: April 10, 2026**

| Phase | Duration | Status |
|-------|----------|--------|
| 1. Real Filesystem | 3 hours | ✅ Complete |
| 2. Real Database | 6 hours | ✅ Complete |
| 3. Real OVS Client | 8 hours | ✅ Complete |
| 4. Real Hypervisor | 12 hours | ✅ Complete |
| **Total** | **29 hours** | **Complete** |

---

## Dependency Analysis

### Current vimic2 Dependencies
```
github.com/mattn/go-sqlite3      # Already present
libvirt.org/libvirt-go          # Already present (pkg/hypervisor)
github.com/stsgym/vimic2/pkg/hypervisor
```

### New Dependencies Required
```
# None - all use existing project packages
```

---

## Interface Compatibility

Each real implementation must satisfy the same interface as its mock:

```go
// Both mock and real implement these interfaces:

type Filesystem interface { ... }  // mockfs.MockFilesystem, realfs.RealFilesystem
type Database interface { ... }    // mockdb.MockDB, realdb.RealDB  
type Hypervisor interface { ... } // mockhv.MockHypervisor, realhv.RealHypervisor
type OVSClient interface { ... }   // mockovs.MockOVSClient, realovs.RealOVSClient
```

This allows:
- Easy switching between mock and real in tests
- Dependency injection in production code
- Graceful degradation if services unavailable

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-------------|--------|------------|
| OVS not installed | Medium | Test failures | Detect and skip integration tests |
| libvirt connection issues | Medium | VM ops fail | Connection pooling + retry |
| SQLite locking | Low | DB contention | WAL mode + proper pooling |
| File permissions | Low | FS ops fail | Proper error handling |

---

## Success Criteria

- [x] All 4 real utilities implemented
- [x] 100% interface compatibility with mocks
- [x] Unit tests pass without external dependencies
- [x] Integration tests pass with infrastructure (OVS)
- [x] Documentation complete
- [x] Roadmap updated with final status

## Final Results

| Utility | Tests | Coverage | Commit |
|---------|-------|----------|--------|
| realfs | 20 | 69.9% | `be2e7f9` |
| realdb | 16 | 85.1% | `72644d4` |
| realovs | 35 (25 unit + 10 integration) | 43.8% | `c3aeec3`, `aec89bb` |
| realhv | 12 | 37.6% | `9a1ba87` |

## Usage

```go
import (
    "github.com/stsgym/vimic2/internal/realutil/realfs"
    "github.com/stsgym/vimic2/internal/realutil/realdb"
    "github.com/stsgym/vimic2/internal/realutil/realovs"
    "github.com/stsgym/vimic2/internal/realutil/realhv"
    "github.com/stsgym/vimic2/internal/testutil/mockfs"
    "github.com/stsgym/vimic2/internal/testutil/mockdb"
    "github.com/stsgym/vimic2/internal/testutil/mockovs"
    "github.com/stsgym/vimic2/internal/testutil/mockhv"
)

// Use mocks for unit tests
testFS := mockfs.NewMockFilesystem()

// Use real for integration/production
prodFS := realfs.NewFilesystem()
```