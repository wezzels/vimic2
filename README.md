# Vimic2 CI/CD Orchestration Platform

**A comprehensive VM-based CI/CD orchestration system with ephemeral build environments**

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

- Go 1.23+
- QEMU/KVM
- Open vSwitch
- libvirt (optional)
- SQLite3

### Installation

```bash
# Clone repository
git clone https://idm.wezzel.com/crab-meat-repos/stsgym-work.git
cd stsgym-work/vimic2

# Install dependencies
go mod download

# Build
make build

# Create VM templates
./scripts/create-templates.sh all

# Run
./vimic2 --config config.yaml
```

### Configuration

```yaml
# config.yaml
database:
  path: ~/.vimic2/pipeline.db

hypervisor:
  type: libvirt
  uri: qemu:///system

templates:
  base_path: /var/lib/vimic2/templates
  default_template: base-go-1.23.qcow2

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
    template: base-go-1.23.qcow2
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
# Unit tests
go test ./... -tags='!integration'

# Integration tests (requires OVS/libvirt)
go test ./... -tags=integration

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release
make release
```

### Code Structure

- **API Layer**: REST endpoints + WebSocket
- **Service Layer**: Business logic (coordinator, dispatcher)
- **Repository Layer**: Database operations
- **Infrastructure**: External services (QEMU, OVS, CI/CD platforms)

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

### Runner Creation

| Platform | Registration | Start |
|----------|-------------|-------|
| GitLab | ~2s | ~5s |
| GitHub | ~3s | ~10s |
| Jenkins | ~1s | ~3s |
| CircleCI | ~2s | ~5s |
| Drone | ~1s | ~2s |

## License

MIT License - See LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes
4. Run tests
5. Submit pull request

## Support

- **Issues**: https://idm.wezzel.com/crab-meat-repos/stsgym-work/-/issues
- **Documentation**: `docs/`
- **Wiki**: https://idm.wezzel.com/crab-meat-repos/stsgym-work/-/wikis

## Authors

- Wesley Robbins (wez@stsgym.com)

## Acknowledgments

- Open vSwitch for network virtualization
- QEMU/KVM for VM management
- GitLab Runner for GitLab CI integration
- GitHub Actions Runner for GitHub integration
- Jenkins Agent for Jenkins integration
- CircleCI Runner for CircleCI integration
- Drone Runner for Drone integration

---

*Built with ❤️ for CI/CD automation*