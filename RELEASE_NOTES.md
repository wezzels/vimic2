# Vimic2 v0.1.0 Release Notes

## First Release - April 2, 2026

### Overview

Vimic2 is a cluster management and orchestration platform for VMs, built in Go with Fyne for cross-platform desktop UI.

### Features

- **Multi-Host Cluster Management**: Manage VMs across multiple hypervisors (libvirt, Hyper-V, Apple)
- **Real-Time Metrics**: CPU, memory, disk, and network monitoring
- **Auto-Scaling**: Automatic scale up/down based on CPU/Memory thresholds
- **Rolling Updates**: Zero-downtime batch updates with health checks
- **Disaster Recovery**: Backup and restore cluster configurations
- **SQLite Persistence**: Embedded database for all cluster state
- **Cross-Platform UI**: Fyne desktop application (Linux, Windows, macOS)

### All 4 Phases Complete

| Phase | Components | Status |
|-------|------------|--------|
| Phase 1 | Core Infrastructure | ✅ Complete |
| Phase 2 | Cluster Management | ✅ Complete |
| Phase 3 | Monitoring & Metrics | ✅ Complete |
| Phase 4 | Orchestration | ✅ Complete |

### Binaries

| Platform | Architecture | Size |
|----------|-------------|------|
| Linux | amd64 | 17MB |
| Windows | amd64 | 16MB |
| macOS | amd64 | 16MB |
| macOS | arm64 | 15MB |

### Installation

#### Linux

```bash
# Download
wget https://github.com/wez/stsgym-vimic2/releases/download/v0.1.0/vimic2-linux-amd64
chmod +x vimic2-linux-amd64
sudo mv vimic2-linux-amd64 /usr/local/bin/vimic2

# Run
vimic2
```

#### macOS

```bash
# Download (Intel)
wget https://github.com/wez/stsgym-vimic2/releases/download/v0.1.0/vimic2-darwin-amd64
chmod +x vimic2-darwin-amd64

# Download (Apple Silicon)
wget https://github.com/wez/stsgym-vimic2/releases/download/v0.1.0/vimic2-darwin-arm64
chmod +x vimic2-darwin-arm64

# Run
./vimic2-darwin-amd64
```

#### Windows

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/wez/stsgym-vimic2/releases/download/v0.1.0/vimic2-windows-amd64.exe" -OutFile "vimic2.exe"

# Run
.\vimic2.exe
```

### Configuration

Create `~/.vimic2/config.yaml`:

```yaml
hosts:
  local:
    name: Local Host
    address: localhost
    user: root
    port: 22
    type: libvirt

autoscale:
  enabled: true
  cpu_threshold: 70
  mem_threshold: 80
  cooldown: 5m

monitor:
  interval: 5s
  retention: 24h
```

### Test Coverage

| Package | Coverage | Tests |
|---------|----------|-------|
| internal/cluster | 60% | 7 |
| internal/deploy | 37% | 9 |
| internal/monitor | 50% | 6 |
| internal/orchestrator | 60% | 10 |
| internal/database | 80% | 3 |
| pkg/hypervisor | 33% | 4 |
| **Total** | **~50%** | **39** |

### Architecture

```
┌─────────────────────────────────────────────┐
│            Vimic2 UI (Fyne Desktop)          │
└───────────────────────┬─────────────────────┘
                        │
┌───────────────────────┼─────────────────────┐
│              Cluster Manager                  │
│  Create  │ Scale  │ Update  │ Backup        │
└───────────────────────┬─────────────────────┘
                        │
┌───────────────────────┼─────────────────────┐
│           Hypervisor Layer                   │
│  Libvirt │ Stub │ Hyper-V │ Apple          │
└───────────────────────┬─────────────────────┘
                        │
┌───────────────────────┼─────────────────────┐
│              VM Cluster                       │
│  Worker-1 │ Worker-2 │ Master              │
└─────────────────────────────────────────────┘
```

### Known Issues

1. **Fyne Display**: Requires X11 or Wayland display for GUI
2. **Libvirt**: Requires libvirt development headers on Linux
3. **Stub Hypervisor**: For testing only, does not create real VMs

### Roadmap

- [ ] Phase 5: Integration tests (80% coverage target)
- [ ] Release binaries with GitHub Actions CI/CD
- [ ] Web UI option (Fyne web backend)
- [ ] REST API for remote management
- [ ] Kubernetes cluster driver

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

### License

MIT License

### Authors

- Wesley Robbins
- OpenClaw AI Assistant

### Changelog

#### v0.1.0 (2026-04-02)

**Added**
- Multi-host cluster management
- VM lifecycle operations (start/stop/restart/delete)
- Real-time metrics collection
- Alerting with configurable thresholds
- Auto-scaling based on resource usage
- Rolling updates with health checks
- Disaster recovery (backup/restore/failover)
- SQLite persistence
- Fyne v2.4 desktop UI
- Viper YAML configuration
- Cross-platform builds

**Fixed**
- Fyne v2.4 API compatibility
- Libvirt import path
- HostConfig struct fields
- Recovery manager initialization

**Dependencies**
- Go 1.21+
- Fyne v2.4
- Viper v1.21
- SQLite3