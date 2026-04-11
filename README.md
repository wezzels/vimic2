# Vimic2 CI/CD Orchestration Platform

**A comprehensive VM-based CI/CD orchestration system with ephemeral build environments**

## Repositories

| Remote | URL |
|--------|-----|
| **IDM (origin)** | `git@idm.wezzel.com:crab-meat-repos/vimic2.git` |
| **GitHub** | `https://github.com/wezzels/vimic2` |

## Overview

Vimic2 is a complete CI/CD orchestration platform that creates isolated, ephemeral build environments using QEMU backing files, Open vSwitch network virtualization, and multi-platform runner support.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           VIMIC2 ARCHITECTURE                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │   GitLab     │    │   GitHub     │    │   Jenkins    │    │   CircleCI   │ │
│  │   Runner     │    │   Runner     │    │   Agent      │    │   Runner     │ │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘ │
│         │                   │                   │                   │         │
│  ┌──────┴───────────────────┴───────────────────┴───────────────────┴───────┐  │
│  │                        RUNNER MANAGER                                     │  │
│  │   CreateRunner • StartRunner • StopRunner • DestroyRunner               │  │
│  └──────────────────────────────┬────────────────────────────────────────────┘  │
│                                │                                                │
│  ┌─────────────────────────────┴─────────────────────────────────────────────┐ │
│  │                          POOL MANAGER                                       │ │
│  │   AcquireVM • ReleaseVM • PreAllocateVMs • StateTracker                   │ │
│  └──────────────────────────────┬────────────────────────────────────────────┘ │
│                                │                                                │
│  ┌─────────────────────────────┴─────────────────────────────────────────────┐ │
│  │                       TEMPLATE MANAGER                                      │ │
│  │   CreateTemplate • CreateOverlay • DeleteOverlay (QEMU backing files)     │ │
│  └──────────────────────────────┬────────────────────────────────────────────┘ │
│                                │                                                │
│  ┌─────────────────────────────┴─────────────────────────────────────────────┐ │
│  │                      NETWORK ISOLATION                                     │ │
│  │   OVS Bridge • VLAN Allocation • IPAM • Firewall Rules                    │ │
│  └──────────────────────────────┬────────────────────────────────────────────┘ │
│                                │                                                │
│  ┌─────────────────────────────┴─────────────────────────────────────────────┐ │
│  │                      DATABASE (SQLite)                                      │ │
│  │   Pipelines • Runners • VMs • Templates • Networks • Artifacts • Logs     │ │
│  └───────────────────────────────────────────────────────────────────────────┘ │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Features

### Core Features

- **Ephemeral Build Environments**: Isolated VMs per pipeline using QEMU backing files
- **Network Isolation**: Per-pipeline VLAN + CIDR allocation with firewall rules
- **Multi-Platform Runners**: GitLab, GitHub, Jenkins, CircleCI, Drone support
- **VM Pool Management**: Pre-allocated VMs with instant creation (~0.5s)
- **Copy-on-Write Overlays**: Minimal disk usage (~100MB per VM)
- **Real-Time Updates**: WebSocket-based live streaming
- **Artifact Management**: Checksum-verified storage with TTL

### Platform Support

| Platform | Runner Type | Status |
|----------|-------------|--------|
| GitLab | gitlab-runner | ✅ Implemented |
| GitHub | actions-runner | ✅ Implemented |
| Jenkins | agent.jar | ✅ Implemented |
| CircleCI | circleci-runner | ✅ Implemented |
| Drone | drone-runner-docker | ✅ Implemented |

### Build Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| Linux | amd64 | ✅ Supported |
| Linux | arm64 | ✅ Supported |
| Windows | amd64 | ✅ Supported |
| macOS | amd64 | ✅ Supported |
| macOS | arm64 | ✅ Supported |

## Project Structure

```
vimic2/
├── internal/
│   ├── api/                    # REST API & WebSocket
│   │   ├── server.go           # API endpoints
│   │   ├── websocket.go        # Real-time updates
│   │   └── api_test.go         # API tests
│   │
│   ├── pipeline/               # Pipeline coordination
│   │   ├── coordinator.go      # Pipeline lifecycle
│   │   ├── dispatcher.go       # Job queue & workers
│   │   ├── artifacts.go        # Artifact storage
│   │   ├── logs.go             # Log collection
│   │   ├── database.go         # SQLite schema
│   │   ├── config.go           # YAML configuration
│   │   ├── ssh.go              # SSH key management
│   │   └── cloudinit.go        # Cloud-init generation
│   │
│   ├── pool/                   # VM pool management
│   │   ├── manager.go          # Pool operations
│   │   ├── state.go            # State tracking
│   │   └── template.go         # QEMU backing files
│   │
│   ├── network/                # Network isolation
│   │   ├── isolation.go        # OVS bridge management
│   │   ├── ipam.go             # IP address management
│   │   ├── firewall.go         # iptables/nftables rules
│   │   ├── vlan.go             # VLAN allocation
│   │   ├── ovs.go              # Open vSwitch client
│   │   └── topology.go         # Network topology
│   │
│   ├── runner/                 # Runner orchestration
│   │   ├── manager.go          # Multi-platform manager
│   │   ├── gitlab.go           # GitLab runner
│   │   ├── github.go           # GitHub runner
│   │   ├── jenkins.go          # Jenkins agent
│   │   ├── circleci.go         # CircleCI runner
│   │   └── drone.go            # Drone runner
│   │
│   └── web/                    # Web UI
│       ├── webui.go            # Template rendering
│       └── templates/          # HTML templates
│
├── scripts/
│   └── create-templates.sh    # VM template creation
│
├── docs/
│   ├── TESTING.md              # Testing guide
│   ├── ARCHITECTURE.md        # Architecture docs
│   └── QUICKSTART.md          # Quick start guide
│
├── .gitlab-ci.yml              # GitLab CI/CD
├── .github/workflows/ci.yml    # GitHub Actions
├── Makefile                     # Build targets
└── README.md                    # This file
```

## Quick Start

### Prerequisites

- Go 1.22+
- QEMU/KVM (for VM operations)
- Open vSwitch (for network isolation)
- libvirt-dev (optional, for libvirt integration)
- SQLite3

### Installation

```bash
# Clone repository
git clone https://github.com/wezzels/vimic2.git
cd vimic2

# Install dependencies
go mod download

# Build
make build
# or: go build -o vimic2 ./cmd/vimic2

# Verify build
./vimic2 --version
```

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
go test -short ./...

# Integration tests (requires OVS + libvirt)
sudo apt-get install openvswitch-switch libvirt-dev
go test ./internal/realutil/... -tags=integration

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Quick Demo

```bash
# 1. Start vimic2 with stub hypervisor (no VMs needed)
./vimic2 --config config.example.yaml &

# 2. Create a pipeline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "gitlab",
    "repository": "https://gitlab.example.com/demo/project",
    "branch": "main",
    "commit_sha": "abc123",
    "commit_message": "Demo commit",
    "author": "demo@example.com",
    "runner_count": 1
  }'

# 3. Check pipeline status
curl http://localhost:8080/api/pipelines

# 4. View stats
curl http://localhost:8080/api/stats
```

### Configuration

```bash
# Copy example config
cp config.example.yaml config.yaml

# Edit for your environment
vim config.yaml
```

### Minimal Configuration

Create `config.yaml` for quick testing:

```yaml
# Minimal config for testing (uses stub hypervisor)
database:
  path: ~/.vimic2/pipeline.db

hypervisor:
  type: stub

templates:
  base_path: ~/.vimic2/templates
  default_template: base.qcow2

platforms:
  gitlab:
    url: https://gitlab.example.com
    enabled: false

  github:
    url: https://github.com
    enabled: false

pools:
  default:
    template: base.qcow2
    min_size: 0
    max_size: 5

networks:
  base_cidr: 10.100.0.0/16
  vlan_start: 1000
  vlan_end: 2000
```

### Full Configuration

```yaml
# config.yaml - Full configuration
database:
  path: ~/.vimic2/pipeline.db

hypervisor:
  type: libvirt
  uri: qemu:///system

templates:
  base_path: /var/lib/vimic2/templates
  default_template: base-go-1.22.qcow2

platforms:
  gitlab:
    url: https://gitlab.example.com
    registration_token: ${GITLAB_RUNNER_TOKEN}
    labels: [builder, go, docker]
    enabled: true

  github:
    url: https://github.com
    token: ${GITHUB_RUNNER_TOKEN}
    labels: [builder, go]
    enabled: true

pools:
  builder:
    template: base-go-1.22.qcow2
    min_size: 2
    max_size: 10
    cpu: 4
    memory: 8192

networks:
  base_cidr: 10.100.0.0/16
  vlan_start: 1000
  vlan_end: 2000
  dns: [8.8.8.8, 8.8.4.4]
```

### API Usage

```bash
# Create pipeline
curl -X POST http://localhost:8080/api/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "gitlab",
    "repository": "https://gitlab.example.com/project",
    "branch": "main",
    "commit_sha": "abc123",
    "commit_message": "Test commit",
    "author": "user@example.com",
    "runner_count": 1,
    "labels": ["builder", "go"]
  }'

# List pipelines
curl http://localhost:8080/api/pipelines

# Get pipeline
curl http://localhost:8080/api/pipelines/{id}

# Start pipeline
curl -X POST http://localhost:8080/api/pipelines/{id}/start

# Stop pipeline
curl -X POST http://localhost:8080/api/pipelines/{id}/stop

# Get stats
curl http://localhost:8080/api/stats
```

### Web UI

Access the web UI at: http://localhost:3000

Features:
- Dashboard with real-time stats
- Pipeline management
- Runner monitoring
- Pool and network views
- Log streaming
- Artifact browser

## Development

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
go test -short ./...

# Integration tests (requires OVS + libvirt)
sudo apt-get install openvswitch-switch libvirt-dev
go test ./internal/realutil/... -tags=integration

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Test specific package
 go test ./internal/network/... -v
```

### Building

```bash
# Build for current platform
make build
# or
go build -o vimic2 ./cmd/vimic2

# Build for all platforms
make build-all

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o vimic2-linux-amd64 ./cmd/vimic2

# Create release
make release
```

### Code Structure

- **API Layer** (`internal/api/`): REST endpoints + WebSocket
- **Service Layer** (`internal/pipeline/`, `internal/runner/`): Business logic
- **Repository Layer** (`internal/database/`): Database operations
- **Infrastructure** (`internal/network/`, `internal/pool/`): External services

### Project Structure

```
vimic2/
├── cmd/vimic2/          # Main application entry point
├── internal/
│   ├── api/             # REST API + WebSocket server
│   ├── cluster/         # Cluster management
│   ├── container/       # Container runtime integration
│   ├── database/        # Database operations
│   ├── deploy/         # Deployment logic
│   ├── host/           # Host machine management
│   ├── monitor/        # Metrics + monitoring
│   ├── network/        # OVS, VLAN, IPAM, firewall
│   ├── orchestrator/   # Pipeline orchestration
│   ├── pipeline/       # Job dispatch, logs, artifacts
│   ├── pool/           # VM pool management
│   ├── realutil/       # Real integration tests
│   ├── runner/         # Multi-platform CI/CD runners
│   └── ui/             # Web UI components
├── pkg/
│   ├── hypervisor/     # Hypervisor abstraction
│   └── types/          # Shared types
├── docs/               # Documentation
├── scripts/            # Utility scripts
├── Makefile            # Build targets
└── README.md           # This file
```

## Performance

### VM Creation

| Operation | Traditional | Vimic2 (backing files) |
|-----------|-------------|-------------------------|
| Create VM | 30-60s | 0.5s |
| Disk usage | 2-5GB | 100MB (delta only) |
| Memory | 512MB+ | 512MB+ |

### Network Isolation

| Operation | Time |
|-----------|------|
| Create bridge | ~10ms |
| Allocate VLAN | ~1ms |
| Allocate CIDR | ~1ms |
| Create firewall rules | ~50ms |

### Test Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| realdb | 85.1% | ✅ Unit tests pass |
| realfs | 87.4% | ✅ Unit tests pass |
| realovs | 87.1% | ✅ Unit tests pass |
| realhv | 59.4% | ⚠️ Needs libvirt VMs |

## Troubleshooting

### Build Issues

```bash
# Go version mismatch
go version  # Should be 1.22+

# Missing dependencies
go mod tidy
go mod download

# Build errors
go build ./...  # Check for compile errors
```

### Test Issues

```bash
# Integration tests fail
sudo apt-get install openvswitch-switch libvirt-dev

# OVS tests fail
ovs-vsctl --version  # Verify OVS installed

# libvirt tests fail
virsh list  # Verify libvirt running
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `executable file not found: ovs-vsctl` | OVS not installed | `sudo apt-get install openvswitch-switch` |
| `no IP address available` | No qemu-guest-agent | Install agent in VM or use ARP fallback |
| `failed to connect to libvirt` | libvirt not running | `sudo systemctl start libvirtd` |
| `connection refused` | vimic2 not running | Start with `./vimic2 --config config.yaml` |

## License

MIT License - See LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Run tests: `go test -short ./...`
5. Submit pull request

## Support

- **Issues**: https://github.com/wezzels/vimic2/issues
- **Documentation**: `docs/`
- **Wiki**: https://github.com/wezzels/vimic2/wiki

## Authors

- Wesley Robbins (wez@stsgym.com)

---

*Built with ❤️ for CI/CD automation*