# Vimic2 Troubleshooting Guide

**Version**: 1.0

---

## Common Issues

### Cannot Connect to Vimic2 Server

**Symptom:** `connection refused` or `timeout` when accessing Vimic2 API.

**Diagnosis:**
```bash
# Check if server is running
ps aux | grep vimic2

# Check port is listening
netstat -tlnp | grep 8080
# or
ss -tlnp | grep 8080
```

**Solutions:**
1. Start the server: `vimic2 server start`
2. Check port is not blocked by firewall: `sudo ufw allow 8080`
3. Verify correct host/port in config: `vimic2 server status`

---

### Authentication Failed

**Symptom:** `401 Unauthorized` when calling API.

**Diagnosis:**
```bash
# Check current token
cat ~/.vimic2/config.yaml | grep token
```

**Solutions:**
1. Verify token matches in config and API call
2. Reset token: `vimic2 auth regenerate-token`
3. Check for trailing spaces in token

---

### Database Locked

**Symptom:** `database is locked` errors.

**Diagnosis:**
```bash
# Check for multiple vimic2 processes
ps aux | grep vimic2 | wc -l
```

**Solutions:**
1. Stop all vimic2 instances: `pkill vimic2`
2. Wait 30 seconds for SQLite to release locks
3. Check disk space: `df -h`
4. If corruption suspected: restore from backup

---

### Network Creation Failed

**Symptom:** `Failed to create network` or `OVS bridge not found`.

**Diagnosis:**
```bash
# Check if OVS is installed
ovs-vsctl show

# Check bridge exists
ip link show vimicbr0
```

**Solutions:**
1. Install Open vSwitch: `sudo apt install openvswitch-switch`
2. Create bridge manually:
   ```bash
   sudo ovs-vsctl add-br vimicbr0
   sudo ip link set vimicbr0 up
   ```
3. Run as root or with correct permissions

---

### VM Pool Empty

**Symptom:** `No available VMs in pool` when acquiring.

**Diagnosis:**
```bash
# List pools and their status
vimic2 pool list

# Check pool details
vimic2 pool status default
```

**Solutions:**
1. Increase pool size: `vimic2 pool resize --name default --size 20`
2. Check hypervisor connection: `vimic2 hypervisor status`
3. Verify templates exist: `vimic2 template list`
4. Check VM creation logs: `vimic2 logs --component pool`

---

### Hypervisor Not Responding

**Symptom:** `hypervisor connection failed` or VMs won't start.

**Diagnosis:**
```bash
# Check hypervisor status
vimic2 hypervisor status

# For libvirt
virsh list

# For QEMU
ps aux | grep qemu
```

**Solutions:**

**Libvirt:**
```bash
# Restart libvirt
sudo systemctl restart libvirtd

# Check connection
virsh --connect qemu:///system list
```

**QEMU:**
```bash
# Check QEMU processes
ps aux | grep qemu

# Restart QEMU if needed
sudo systemctl restart qemu
```

**Apple HV (macOS):**
- Ensure Hypervisor.framework entitlement is granted
- Check System Preferences → Security & Privacy

**Windows Hyper-V:**
```powershell
# Check Hyper-V service
Get-Service vmcompute

# Restart if stopped
Restart-Service vmcompute
```

---

### WebSocket Connection Drops

**Symptom:** Real-time updates stop or connection refused.

**Diagnosis:**
```bash
# Check WebSocket endpoint
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  http://localhost:8080/api/ws
```

**Solutions:**
1. Check token is valid and passed as query param: `ws://host/api/ws?token=<token>`
2. Verify no proxy blocking WebSocket
3. Check server logs for errors
4. Restart server: `vimic2 server restart`

---

### High Memory Usage

**Symptom:** System memory exhausted, vimic2 process using too much RAM.

**Diagnosis:**
```bash
# Check vimic2 memory usage
ps aux | grep vimic2

# Check pool sizes
vimic2 pool list
```

**Solutions:**
1. Reduce pool sizes: `vimic2 pool resize --name default --size 5`
2. Reduce monitoring interval: set `monitoring.interval: 30` in config
3. Disable unused runners in config
4. Enable memory profiling: add `--profile` flag and analyze

---

### Pipeline Hangs

**Symptom:** Pipeline stuck in `running` state, never completes.

**Diagnosis:**
```bash
# List pipelines
vimic2 pipeline list

# Get pipeline details
vimic2 pipeline status <pipeline-id>

# Check job logs
vimic2 job logs <job-id>
```

**Solutions:**
1. Force stop: `vimic2 pipeline stop <pipeline-id>`
2. Cancel stuck job: `vimic2 job cancel <job-id>`
3. Check runner is responsive: `vimic2 runner status <runner-id>`
4. Review timeout settings in pipeline config

---

### Artifact Upload Fails

**Symptom:** `Failed to store artifact` or checksum mismatch.

**Diagnosis:**
```bash
# Check artifact storage location
df -h ~/.vimic2/artifacts

# Check disk space
df -h
```

**Solutions:**
1. Free up disk space or configure alternate storage path
2. Check file permissions: `ls -la ~/.vimic2/artifacts`
3. Verify checksum algorithm matches (SHA256)
4. Increase `max_file_size` in config if artifacts are large

---

### Backup/Restore Issues

**Symptom:** Backup fails or restore doesn't work.

**Diagnosis:**
```bash
# Check backup directory
ls -la ~/.vimic2/backups/

# Check backup schedule
vimic2 backup status
```

**Solutions:**
1. Ensure backup directory exists and is writable
2. Check disk space for backups
3. Verify backup file integrity:
   ```bash
   tar -tzf ~/.vimic2/backups/backup-2026-04-03.tar.gz
   ```
4. For restore: `vimic2 backup restore --file backup-2026-04-03.tar.gz`

---

## Debug Mode

Enable debug logging for detailed diagnostics:

```bash
# Start with debug logging
vimic2 server start --debug

# Or via config
logging:
  level: "debug"
```

Debug logs show:
- All API requests/responses
- Hypervisor commands
- Network operations
- Pool management details

---

## Health Check Commands

```bash
# Overall health
vimic2 health

# Component status
vimic2 status

# Database integrity
vimic2 db check

# Network connectivity
vimic2 network diag
```

---

## Getting Help

### View Logs

```bash
# All logs
vimic2 logs

# Follow in real-time
vimic2 logs -f

# Filter by component
vimic2 logs --component api
vimic2 logs --component pool
vimic2 logs --component runner
```

### Generate Debug Report

```bash
vimic2 debug report --output debug.txt
```

This collects:
- Configuration
- Recent logs
- Database state
- System info

---

## Known Limitations

1. **SQLite not cluster-aware**: Do not run multiple vimic2 instances against same database
2. **OVS required for network isolation**: Without OVS, network features disabled
3. **macOS Apple HV**: Requires entitlements and macOS 13+
4. **Windows WSL**: Limited VM support, recommend Hyper-V instead
5. **ARM64 on Windows**: Not supported

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Authentication failed |
| 4 | Resource not found |
| 5 | Database error |
| 6 | Hypervisor error |
| 7 | Network error |
| 8 | Permission denied |
