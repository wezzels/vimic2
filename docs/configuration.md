# Vimic2 Configuration Reference

**Version**: 1.0

---

## Configuration File

Vimic2 uses a YAML configuration file at `~/.vimic2/config.yaml` or specified via `--config` flag.

### Full Configuration Example

```yaml
# Server Configuration
server:
  host: "0.0.0.0"
  port: 8080
  tls:
    enabled: false
    cert: "/path/to/cert.pem"
    key: "/path/to/key.pem"
  auth:
    token: "your-secure-token-here"
    token_file: "/path/to/token"

# Database Configuration
database:
  path: "~/.vimic2/vimic2.db"
  backup:
    enabled: true
    interval: 3600  # seconds
    path: "~/.vimic2/backups"

# Network Configuration
network:
  bridge: "vimicbr0"
  external_iface: "eth0"
  vlan_start: 100
  vlan_end: 1000
  ipam:
    start: "10.0.100.0"
    end: "10.0.200.0"
    gateway: "10.0.0.1"
    dns:
      - "8.8.8.8"
      - "8.8.4.4"

# VM Pool Configuration
pools:
  default:
    size: 10
    platform: "linux/amd64"
    memory: 2048
    cpus: 2
    template: "ubuntu-22.04"
    disk_size: 20971520  # 20GB in bytes
    
  highmem:
    size: 5
    platform: "linux/amd64"
    memory: 8192
    cpus: 4
    template: "ubuntu-22.04"
    
  windows:
    size: 3
    platform: "windows/amd64"
    memory: 4096
    cpus: 2
    template: "windows-2022"

# Runner Configuration
runners:
  gitlab:
    enabled: true
    url: "https://gitlab.com"
    registration_token: "your-gitlab-token"
    
  github:
    enabled: true
    url: "https://github.com"
    app_id: 12345
    private_key: "/path/to/private-key.pem"
    
  jenkins:
    enabled: false
    url: "http://jenkins:8080"
    credentials_id: "vimic2-agent"

# Hypervisor Configuration
hypervisor:
  type: "libvirt"  # libvirt, qemu, apple, windows, wsl
  connection: "qemu:///system"
  storage_pool: "default"
  network: "default"

# Monitoring Configuration
monitoring:
  enabled: true
  interval: 10  # seconds
  metrics:
    cpu: true
    memory: true
    disk: true
    network: true

# Alerting Configuration
alerts:
  enabled: true
  thresholds:
    cpu_percent: 80
    memory_percent: 85
    disk_percent: 90
  webhooks:
    - url: "http://alertmanager:9093/webhook"
      enabled: false

# Logging Configuration
logging:
  level: "info"  # debug, info, warn, error
  file: "~/.vimic2/logs/vimic2.log"
  max_size: 104857600  # 100MB
  max_backups: 5
```

---

## Environment Variables

Override configuration with environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `VIMIC2_CONFIG` | Config file path | `~/.vimic2/config.yaml` |
| `VIMIC2_DB_PATH` | Database file path | `~/.vimic2/vimic2.db` |
| `VIMIC2_TOKEN` | Auth token | - |
| `VIMIC2_PORT` | Server port | `8080` |
| `VIMIC2_HOST` | Server host | `0.0.0.0` |
| `VIMIC2_LOG_LEVEL` | Log level | `info` |
| `VIMIC2_BRIDGE` | OVS bridge name | `vimicbr0` |
| `VIMIC2_HYPERVISOR` | Hypervisor type | `libvirt` |

---

## Command Line Flags

```
vimic2 [flags]

Flags:
  -h, --help              Show help
  -v, --version           Show version
  -c, --config string     Config file path (default ~/.vimic2/config.yaml)
  -d, --debug             Enable debug logging
  -p, --port int          Server port (default 8080)
  -t, --token string      Auth token (overrides config)
```

---

## Pool Configuration

### Creating a Pool

```bash
vimic2 pool create --name build --size 10 --platform linux/amd64
```

### Pool Options

| Option | Description | Required |
|--------|-------------|----------|
| `--name` | Pool name | Yes |
| `--size` | Number of VMs | Yes |
| `--platform` | Platform (e.g., linux/amd64) | Yes |
| `--memory` | RAM in MB | No (default: 2048) |
| `--cpus` | Number of CPUs | No (default: 2) |
| `--template` | VM template name | No |
| `--disk-size` | Disk size in bytes | No (default: 20GB) |

---

## Network Configuration

### Creating a Network

```bash
vimic2 network create \
  --name my-network \
  --vlan 100 \
  --cidr 10.0.100.0/24 \
  --gateway 10.0.100.1
```

### Firewall Rules

```bash
# Allow all outbound
vimic2 network rule add my-network --action allow --dst 0.0.0.0/0 --outbound

# Allow HTTP/HTTPS inbound
vimic2 network rule add my-network --action allow --dst 0.0.0.0/0 --ports 80,443 --inbound

# Block specific IP
vimic2 network rule add my-network --action deny --src 192.168.1.100
```

---

## Runner Configuration

### GitLab Runner Registration

```bash
vimic2 runner register gitlab \
  --url https://gitlab.com \
  --token <registration-token> \
  --platform linux/amd64 \
  --labels docker,linux
```

### GitHub Actions Runner Registration

```bash
vimic2 runner register github \
  --url https://github.com/<org> \
  --app-id 12345 \
  --private-key /path/to/key.pem \
  --platform linux/amd64
```

---

## Hypervisor Configuration

### Libvirt (Linux)

```yaml
hypervisor:
  type: "libvirt"
  connection: "qemu:///system"
```

### QEMU (Standalone)

```yaml
hypervisor:
  type: "qemu"
  socket: "/tmp/qemu.sock"
```

### Apple HV (macOS)

```yaml
hypervisor:
  type: "apple"
```

### Windows Hyper-V

```yaml
hypervisor:
  type: "windows"
  connection: "hyperv://localhost"
```

### WSL (Windows Subsystem for Linux)

```yaml
hypervisor:
  type: "wsl"
```

---

## Template Configuration

VM templates are QEMU backing files stored in `/var/lib/vimic2/templates/`.

### Creating a Template

```bash
# Create from existing qcow2 image
vimic2 template create ubuntu-22.04 \
  --source /path/to/ubuntu-22.04.qcow2 \
  --size 20G

# Create blank template
vimic2 template create empty \
  --size 10G \
  --platform linux/amd64
```

---

## Monitoring Thresholds

Default alerting thresholds:

```yaml
alerts:
  thresholds:
    cpu_percent: 80      # Alert when CPU > 80%
    memory_percent: 85    # Alert when memory > 85%
    disk_percent: 90     # Alert when disk > 90%
    vm_failure_rate: 0.1 # Alert when > 10% of VMs fail
```

---

## Logging

### Log Levels

- `debug`: Detailed debugging info
- `info`: General operational info
- `warn`: Warning conditions
- `error`: Error conditions

### Log Locations

| Environment | Default Log Path |
|-------------|------------------|
| Linux | `~/.vimic2/logs/vimic2.log` |
| macOS | `~/.vimic2/logs/vimic2.log` |
| Windows | `%USERPROFILE%\.vimic2\logs\vimic2.log` |

---

## Full Minimal Config

For testing with stub hypervisor:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  auth:
    token: "test-token"

database:
  path: ":memory:"  # In-memory for testing

network:
  bridge: "vimicbr0"

hypervisor:
  type: "stub"  # No real VMs

monitoring:
  enabled: false
```
