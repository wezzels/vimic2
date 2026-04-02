# Vimic2 - Quick Start Guide

A cross-platform desktop application for managing VM clusters with real-time monitoring, auto-scaling, and alerting.

## Installation

### Pre-built Binaries

Download from the releases page for your platform:

```bash
# Linux (x86_64)
curl -L -o vimic2 https://github.com/stsgym/vimic2/releases/latest/vimic2-linux-amd64
chmod +x vimic2
sudo mv vimic2 /usr/local/bin/

# macOS (Intel)
curl -L -o vimic2 https://github.com/stsgym/vimic2/releases/latest/vimic2-darwin-amd64
chmod +x vimic2
sudo mv vimic2 /usr/local/bin/

# macOS (Apple Silicon)
curl -L -o vimic2 https://github.com/stsgym/vimic2/releases/latest/vimic2-darwin-arm64
chmod +x vimic2
sudo mv vimic2 /usr/local/bin/

# Windows
# Download vimic2-windows-amd64.exe from releases
```

### Build from Source

```bash
# Clone the repository
git clone https://idm.wezzel.com/stsgym/vimic2.git
cd vimic2

# Install dependencies
go mod download

# Build for current platform
make build

# Or build for all platforms
make build-all
```

## Prerequisites

### Linux (libvirt)
```bash
# Install libvirt and qemu
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients

# Add your user to the libvirt group
sudo usermod -aG libvirt $USER
newgrp libvirt

# Verify libvirt is running
sudo systemctl status libvirtd
```

### macOS
```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install qemu
```

### Windows
- Windows 10/11 with Hyper-V enabled
- Or use WSL2 with Linux VM

## Getting Started

### 1. Launch Vimic2

```bash
vimic2
```

The dashboard will appear showing:
- Total hosts, clusters, and nodes
- Running/stopped status
- Recent clusters

### 2. Add a Host

1. Click **Add Host** in the sidebar or menu
2. Enter the host details:
   - **Name**: A friendly name (e.g., "hypervisor-1")
   - **Address**: IP address or hostname
   - **SSH User**: Username for SSH connection (e.g., "root")
3. Click **Add**

Vimic2 will:
- Test the SSH connection
- Detect the hypervisor type
- Add the host to your inventory

### 3. Create a Cluster

1. Click **+ New Cluster** in the sidebar
2. Follow the deployment wizard:

#### Step 1: Name
- Enter a name for your cluster (e.g., "production-cluster")

#### Step 2: Choose Template
Select a preset or start from scratch:
- **Development**: Small cluster (2 workers)
- **Production**: HA cluster (3 masters, 5 workers)
- **Database**: DB-focused (1 primary, 2 replicas)

#### Step 3: Configure Nodes
- Adjust CPU, memory, disk as needed
- Select which hosts to deploy to
- Set node roles (worker, master, database)

#### Step 4: Review & Deploy
- Review your configuration
- Click **Deploy**

### 4. Manage Nodes

From the cluster detail view:
- **Start**: Power on a stopped node
- **Stop**: Gracefully shutdown a running node
- **Restart**: Reboot a node
- **Delete**: Remove a node (with confirmation)

### 5. Monitor Resources

The dashboard shows:
- **CPU Usage**: Per-node and cluster average
- **Memory Usage**: Current utilization
- **Disk Usage**: Storage consumption
- **Network**: RX/TX traffic

Click any node for detailed metrics history.

### 6. Configure Alerts

1. Navigate to **Alerts** in the sidebar
2. View default alert rules:
   - High CPU (>90% for 5 minutes)
   - High Memory (>90% for 5 minutes)
   - Disk Full (>95%)
   - Node Down (no heartbeat for 1 minute)
3. Add custom rules:
   - Click **Add Rule**
   - Set metric, threshold, and duration
   - Enable/disable as needed

### 7. Enable Auto-scaling

1. Select a cluster
2. Click **Scale** in the toolbar
3. Configure:
   - **Min/Max Nodes**: Cluster size limits
   - **Scale Up Threshold**: CPU/memory % to trigger scale-up
   - **Scale Down Threshold**: CPU/memory % to trigger scale-down
   - **Cooldown**: Time between scaling operations

## Common Tasks

### Scale a Cluster Manually
```
Cluster Detail → Scale → Enter desired node count → Scale
```

### View Node Logs
```
Cluster Detail → Click node name → View Details
```

### Backup a Cluster
```
Cluster Detail → Settings → Create Backup
```

### Restore from Backup
```
Settings → Backups → Select backup → Restore
```

### Rolling Update
```
Cluster Detail → Update → Select new image → Update
```

## Troubleshooting

### "Connection refused" to host
1. Verify SSH is running on the host
2. Check firewall allows port 22
3. Verify SSH key authentication works

### "Permission denied" errors
1. Ensure user is in libvirt group (Linux)
2. Check SSH key has proper permissions
3. Verify sudo access for certain operations

### VMs not starting
1. Check libvirt daemon is running
2. Verify disk image storage has space
3. Check CPU/memory allocation is valid

### Dashboard not updating
1. Click **Refresh** in toolbar
2. Check monitoring interval in Settings
3. Verify database is writable

## Next Steps

- Read the [User Guide](./USER_GUIDE.md) for detailed documentation
- See [Architecture](./ARCHITECTURE.md) for technical details
- Check [API Reference](./API.md) for programmatic access
