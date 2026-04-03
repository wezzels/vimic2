# Vimic2 API Reference

**Version**: 1.0  
**Base URL**: `http://localhost:8080/api`  
**Authentication**: Bearer token (see `Authorization` header)

---

## Authentication

All endpoints except `/health` require authentication via Bearer token:

```
Authorization: Bearer <token>
```

---

## Health Check

### GET /api/health

Health check endpoint (no authentication required).

**Response 200:**
```json
{
  "status": "ok",
  "service": "vimic2",
  "version": "1.0.0"
}
```

---

## Pipelines

### GET /api/pipelines

List all pipelines.

**Response 200:**
```json
{
  "pipelines": [
    {
      "id": "pipe_abc123",
      "name": "my-pipeline",
      "status": "running",
      "created_at": "2026-04-03T10:00:00Z"
    }
  ]
}
```

### GET /api/pipelines/{id}

Get pipeline details.

**Response 200:**
```json
{
  "id": "pipe_abc123",
  "name": "my-pipeline",
  "status": "running",
  "jobs": [
    {
      "id": "job_xyz",
      "name": "build",
      "status": "running",
      "runner_id": "runner_1"
    }
  ],
  "created_at": "2026-04-03T10:00:00Z"
}
```

### POST /api/pipelines

Create a new pipeline.

**Request:**
```json
{
  "name": "my-pipeline",
  "config": {
    "stages": ["build", "test", "deploy"],
    "timeout": 3600
  }
}
```

**Response 201:**
```json
{
  "id": "pipe_abc123",
  "name": "my-pipeline",
  "status": "created"
}
```

### POST /api/pipelines/{id}/start

Start a pipeline.

**Response 200:**
```json
{
  "id": "pipe_abc123",
  "status": "running"
}
```

### POST /api/pipelines/{id}/stop

Stop a pipeline.

**Response 200:**
```json
{
  "id": "pipe_abc123",
  "status": "stopped"
}
```

### DELETE /api/pipelines/{id}

Delete/destroy a pipeline.

**Response 200:**
```json
{
  "id": "pipe_abc123",
  "status": "destroyed"
}
```

---

## Jobs

### GET /api/jobs

List all jobs.

**Response 200:**
```json
{
  "jobs": [
    {
      "id": "job_xyz",
      "pipeline_id": "pipe_abc123",
      "name": "build",
      "status": "running"
    }
  ]
}
```

### GET /api/jobs/{id}

Get job details.

**Response 200:**
```json
{
  "id": "job_xyz",
  "pipeline_id": "pipe_abc123",
  "name": "build",
  "status": "running",
  "started_at": "2026-04-03T10:00:00Z",
  "logs": "..."
}
```

### POST /api/jobs

Enqueue a new job.

**Request:**
```json
{
  "pipeline_id": "pipe_abc123",
  "name": "build",
  "runner_id": "runner_1"
}
```

**Response 201:**
```json
{
  "id": "job_xyz",
  "status": "queued"
}
```

### POST /api/jobs/{id}/cancel

Cancel a job.

**Response 200:**
```json
{
  "id": "job_xyz",
  "status": "cancelled"
}
```

### POST /api/jobs/{id}/retry

Retry a failed job.

**Response 200:**
```json
{
  "id": "job_xyz",
  "status": "queued"
}
```

---

## Runners

### GET /api/runners

List all runners.

**Response 200:**
```json
{
  "runners": [
    {
      "id": "runner_1",
      "type": "gitlab",
      "status": "online",
      "platform": "linux/amd64"
    }
  ]
}
```

### GET /api/runners/{id}

Get runner details.

**Response 200:**
```json
{
  "id": "runner_1",
  "type": "gitlab",
  "status": "online",
  "platform": "linux/amd64",
  "current_job": "job_xyz",
  "labels": ["docker", "linux"]
}
```

### POST /api/runners

Register a new runner.

**Request:**
```json
{
  "type": "gitlab",
  "name": "gitlab-runner-1",
  "platform": "linux/amd64",
  "labels": ["docker", "linux"]
}
```

**Response 201:**
```json
{
  "id": "runner_1",
  "token": "runner_token_xxx"
}
```

### POST /api/runners/{id}/start

Start a runner.

**Response 200:**
```json
{
  "id": "runner_1",
  "status": "online"
}
```

### POST /api/runners/{id}/stop

Stop a runner.

**Response 200:**
```json
{
  "id": "runner_1",
  "status": "offline"
}
```

### DELETE /api/runners/{id}

Deregister a runner.

**Response 200:**
```json
{
  "id": "runner_1",
  "status": "deregistered"
}
```

---

## VM Pools

### GET /api/pools

List all VM pools.

**Response 200:**
```json
{
  "pools": [
    {
      "name": "default",
      "size": 10,
      "available": 7,
      "platform": "linux/amd64"
    }
  ]
}
```

### GET /api/pools/{name}

Get pool details.

**Response 200:**
```json
{
  "name": "default",
  "size": 10,
  "available": 7,
  "platform": "linux/amd64",
  "vm_template": "ubuntu-22.04"
}
```

### POST /api/pools

Create a new VM pool.

**Request:**
```json
{
  "name": "default",
  "size": 10,
  "platform": "linux/amd64",
  "vm_template": "ubuntu-22.04"
}
```

**Response 201:**
```json
{
  "name": "default",
  "size": 10,
  "status": "created"
}
```

### GET /api/pools/{name}/vms

List VMs in a pool.

**Response 200:**
```json
{
  "vms": [
    {
      "id": "vm_abc",
      "pool": "default",
      "status": "available",
      "ip": "10.0.0.5"
    }
  ]
}
```

### POST /api/pools/{name}/acquire

Acquire a VM from a pool.

**Response 200:**
```json
{
  "vm_id": "vm_abc",
  "ip": "10.0.0.5",
  "mac": "52:54:00:12:34:56"
}
```

### POST /api/pools/{name}/release

Release a VM back to pool.

**Request:**
```json
{
  "vm_id": "vm_abc"
}
```

**Response 200:**
```json
{
  "vm_id": "vm_abc",
  "status": "available"
}
```

---

## Networks

### GET /api/networks

List all networks.

**Response 200:**
```json
{
  "networks": [
    {
      "id": "net_abc",
      "name": "pipeline-123",
      "vlan": 100,
      "cidr": "10.0.100.0/24",
      "gateway": "10.0.100.1"
    }
  ]
}
```

### GET /api/networks/{id}

Get network details.

**Response 200:**
```json
{
  "id": "net_abc",
  "name": "pipeline-123",
  "vlan": 100,
  "cidr": "10.0.100.0/24",
  "gateway": "10.0.100.1",
  "dns": ["8.8.8.8"],
  "firewall_rules": [
    {
      "action": "allow",
      "src": "10.0.100.0/24",
      "dst": "0.0.0.0/0",
      "ports": "80,443"
    }
  ]
}
```

### POST /api/networks

Create a new network.

**Request:**
```json
{
  "name": "pipeline-123",
  "vlan": 100,
  "cidr": "10.0.100.0/24"
}
```

**Response 201:**
```json
{
  "id": "net_abc",
  "name": "pipeline-123",
  "status": "created"
}
```

### DELETE /api/networks/{id}

Delete a network.

**Response 200:**
```json
{
  "id": "net_abc",
  "status": "destroyed"
}
```

---

## Artifacts

### GET /api/artifacts

List all artifacts.

**Response 200:**
```json
{
  "artifacts": [
    {
      "id": "art_xyz",
      "name": "build-output",
      "size": 1048576,
      "checksum": "sha256:abc123..."
    }
  ]
}
```

### GET /api/artifacts/{id}

Download/get artifact.

**Response 200:**
```json
{
  "id": "art_xyz",
  "name": "build-output",
  "size": 1048576,
  "checksum": "sha256:abc123...",
  "url": "/api/artifacts/art_xyz/download"
}
```

---

## Error Responses

All endpoints may return these error codes:

| Code | Meaning |
|------|---------|
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |

**Error Response Format:**
```json
{
  "error": "resource_not_found",
  "message": "Pipeline pipe_abc123 not found"
}
```

---

## WebSocket Events

Connect to `/api/ws` for real-time updates.

**Authentication:**
```
ws://localhost:8080/api/ws?token=<bearer_token>
```

**Event Types:**
- `pipeline.started`
- `pipeline.stopped`
- `pipeline.completed`
- `job.started`
- `job.completed`
- `job.failed`
- `runner.online`
- `runner.offline`

**Event Format:**
```json
{
  "type": "job.completed",
  "data": {
    "id": "job_xyz",
    "status": "success"
  }
}
```
