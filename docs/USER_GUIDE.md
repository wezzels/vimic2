# Vimic2 - User Guide

## Table of Contents
1. [Dashboard](#dashboard)
2. [Hosts](#hosts)
3. [Clusters](#clusters)
4. [Nodes](#nodes)
5. [Monitoring](#monitoring)
6. [Alerts](#alerts)
7. [Auto-scaling](#auto-scaling)
8. [Settings](#settings)

---

## Dashboard

The dashboard provides an overview of your entire infrastructure.

### Stats Cards

| Card | Description |
|------|-------------|
| **Hosts** | Number of configured hypervisor hosts |
| **Clusters** | Total number of clusters |
| **Nodes** | Total VMs across all clusters |
| **Running** | Currently running VMs |
| **Stopped** | Stopped VMs |
| **Alerts** | Active alerts |

### Cluster Averages

Below the main stats:
- **Avg CPU**: Average CPU usage across running nodes
- **Avg Memory**: Average memory usage

---

## Hosts

Hosts are hypervisor servers that run your VMs.

### Adding a Host

1. Click **Add Host** (toolbar or sidebar)
2. Enter details:
   ```
   Name: hypervisor-1
   Address: 192.168.1.100
   SSH User: root
   ```
3. Click **Add**

### Host Connection

Vimic2 connects to hosts via:
- **Local**: Direct libvirt access (for localhost)
- **SSH**: Remote connection with key-based auth

### Removing a Host

> ⚠️ **Warning**: Removing a host will not delete its VMs. VMs will continue running but won't be managed by Vimic2.

1. Go to **Settings** → **Hosts**
2. Click **Remove** next to the host
3. Confirm removal

---

## Clusters

A cluster is a group of related VMs (nodes).

### Create Cluster

Use the deployment wizard:

1. Click **+ New Cluster**
2. **Step 1 - Name**: Enter cluster name
3. **Step 2 - Template**: Choose preset or custom
4. **Step 3 - Configure**: Set node specs
5. **Step 4 - Review**: Confirm and deploy

### Cluster Templates

| Template | Description | Nodes |
|----------|-------------|-------|
| **dev** | Development environment | 2 workers |
| **prod** | Production HA cluster | 3 masters, 5 workers |
| **db** | Database cluster | 1 primary, 2 replicas |

### Custom Cluster

Start from scratch and configure:

```yaml
Node Groups:
  - Name: web
    Role: worker
    Count: 3
    CPU: 2
    Memory: 2048 MB
    Disk: 20 GB
    
  - Name: api
    Role: worker
    Count: 2
    CPU: 4
    Memory: 4096 MB
    Disk: 30 GB
```

### Cluster Detail View

Click any cluster to see:
- Node list with status
- Resource usage per node
- Action buttons (Start/Stop/Restart/Delete)

### Delete Cluster

> ⚠️ **Warning**: This deletes ALL nodes in the cluster!

1. Select the cluster
2. Click **Delete**
3. Confirm deletion

---

## Nodes

Nodes are individual virtual machines.

### Node States

| State | Description |
|-------|-------------|
| 🟢 **Running** | VM is powered on |
| 🟡 **Stopped** | VM is powered off |
| 🔴 **Error** | VM encountered an error |

### Node Actions

| Action | Description |
|--------|-------------|
| **Start** | Power on the VM |
| **Stop** | Gracefully shutdown (or force if needed) |
| **Restart** | Reboot the VM |
| **Delete** | Remove the VM |

### Node Configuration

Each node has:

```yaml
Name: prod-worker-1
Role: worker
CPU: 2 cores
Memory: 2048 MB
Disk: 20 GB
Image: ubuntu-22.04
Network: nat
IP: 192.168.122.10
```

### Node Roles

| Role | Purpose |
|------|---------|
| **master** | Control plane node |
| **worker** | Application workload |
| **database** | Data storage node |

---

## Monitoring

### Real-time Metrics

Vimic2 collects metrics every 5 seconds:

- **CPU**: Percentage of CPU utilization
- **Memory**: Percentage of RAM in use
- **Disk**: Percentage of storage used
- **Network**: RX/TX bytes per second

### Node Detail Metrics

Click any node for:
- Current CPU, Memory, Disk usage
- Charts showing last hour of data
- Historical metrics table

### Cluster Overview

For cluster-level metrics:
- Average CPU across all nodes
- Average Memory across all nodes
- Per-node breakdown

### Metrics Retention

Metrics are stored in SQLite:
- Default retention: 24 hours
- Configurable in Settings
- Older metrics are automatically cleaned up

---

## Alerts

### Alert Rules

Configure conditions that trigger notifications:

| Rule | Metric | Threshold | Duration |
|------|--------|-----------|----------|
| High CPU | cpu | > 90% | 5 min |
| High Memory | memory | > 90% | 5 min |
| Disk Full | disk | > 95% | 1 min |
| Node Down | heartbeat | missing | 1 min |

### Adding Custom Rules

1. Go to **Alerts** → **Rules**
2. Click **Add Rule**
3. Configure:
   ```
   Name: Medium CPU Warning
   Metric: cpu
   Threshold: 80%
   Duration: 10 minutes
   ```

### Viewing Alerts

**Alerts** page shows:
- **Critical** (🔴): Immediate attention needed
- **Warning** (🟡): Monitor but not urgent
- **OK** (🟢): All clear

### Acknowledging Alerts

When an alert fires:
1. View the alert details
2. Click **Acknowledge**
3. Alert is marked as seen
4. Alert continues until condition resolves

### Alert Resolution

Alerts auto-resolve when:
- Metric drops below threshold
- Condition no longer met for duration
- Node comes back online

---

## Auto-scaling

Vimic2 can automatically adjust cluster size based on load.

### Enable Auto-scaling

1. Select cluster
2. Click **Scale**
3. Toggle **Auto-scale** ON

### Scaling Rules

| Setting | Description |
|---------|-------------|
| **Metric** | CPU or Memory |
| **Scale Up** | Threshold to add nodes |
| **Scale Down** | Threshold to remove nodes |
| **Cooldown** | Wait time between scaling |

### Example Configuration

```yaml
Scale Up Threshold: 70%
Scale Down Threshold: 30%
Scale Up Count: 1 node
Scale Down Count: 1 node
Cooldown: 5 minutes
```

When cluster average CPU > 70% for 5 minutes → Add 1 node

When cluster average CPU < 30% for 5 minutes → Remove 1 node

### Min/Max Limits

Prevent over/under scaling:

```yaml
Min Nodes: 2
Max Nodes: 10
```

### Manual Scaling

Override auto-scale temporarily:

1. Click **Scale**
2. Enter desired node count
3. Click **Scale Now**

---

## Settings

### General

| Setting | Default | Description |
|---------|---------|-------------|
| **Data Directory** | ~/.vimic2 | Where data is stored |
| **Log Level** | INFO | DEBUG, INFO, WARN, ERROR |

### Monitoring

| Setting | Default | Description |
|---------|---------|-------------|
| **Interval** | 5 seconds | How often to collect metrics |
| **Retention** | 24 hours | How long to keep metrics |

### Auto-scaling Defaults

| Setting | Default | Description |
|---------|---------|-------------|
| **Enabled** | Yes | Enable auto-scaling by default |
| **CPU Threshold** | 70% | Default scale-up threshold |
| **Memory Threshold** | 80% | Default scale-up threshold |
| **Cooldown** | 5 min | Default cooldown period |

### Hosts

View and manage configured hosts:
- Name and address
- Connection status
- Node count

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+R` | Refresh |
| `Ctrl+N` | New Cluster |
| `Ctrl+,` | Settings |
| `Ctrl+Q` | Quit |
| `F5` | Refresh dashboard |

---

## File Locations

```bash
~/.vimic2/
├── vimic2.db        # SQLite database
├── backups/         # Cluster backups
├── logs/           # Application logs
└── config.yaml     # Configuration
```

---

## Command Line

Vimic2 supports headless CLI mode:

```bash
# List clusters
vimic2 cluster list

# Start a node
vimic2 node start prod-worker-1

# View status
vimic2 status

# Import configuration
vimic2 import config.yaml
```

---

## Support

- **Documentation**: See docs/ folder
- **Issues**: Report bugs on GitLab
- **Email**: wlrobbi@stsgym.com
