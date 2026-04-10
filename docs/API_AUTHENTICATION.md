# API Documentation - Vimic2

**Version**: 1.0  
**Base URL**: `http://localhost:8080/api` (local) or `https://vimic2.stsgym.com/api` (production)

---

## Overview

The Vimic2 REST API provides programmatic access to all Vimic2 functionality. The API is organized around REST and uses standard HTTP response codes.

## Authentication

### Bearer Token

All authenticated endpoints require a Bearer token in the Authorization header:

```http
Authorization: Bearer <your-token>
```

### Obtaining a Token

Tokens are issued via the `/auth/login` endpoint:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-04-05T00:00:00Z"
}
```

### Token Refresh

Refresh an expiring token:

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer <current-token>"
```

### Token Storage

Store the token securely:
- **CLI**: Use `vimic2 config set token <token>`
- **Code**: Store in environment variable or secure vault

---

## Swagger UI

A interactive API explorer is available at:

- **Local**: `http://localhost:8080/api/docs/`
- **Online**: `https://vimic2.stsgym.com/api/docs/`

The Swagger UI allows you to:
- Explore all endpoints
- Execute requests directly
- View request/response schemas

---

## Rate Limiting

| Tier | Requests/Minute | Burst |
|------|-----------------|-------|
| Anonymous | 10 | 20 |
| Authenticated | 100 | 200 |
| Admin | 1000 | 2000 |

Rate limit headers are included in every response:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1712236800
```

---

## Pagination

List endpoints support pagination:

```bash
# Get first page
GET /pipelines?limit=20

# Get next page
GET /pipelines?limit=20&cursor=eyJpZCI6MTJ9
```

**Response with pagination:**
```json
{
  "items": [...],
  "next_cursor": "eyJpZCI6MTJ9",
  "total": 42,
  "has_more": true
}
```

---

## Filtering & Sorting

### Filtering

Filter resources using query parameters:

```bash
# Filter by status
GET /pipelines?status=running

# Filter by date range
GET /pipelines?created_after=2026-04-01T00:00:00Z

# Multiple filters
GET /nodes?cluster=prod&status=running
```

### Sorting

Sort using `sort` parameter:

```bash
# Ascending
GET /pipelines?sort=created_at

# Descending
GET /pipelines?sort=-created_at

# Multiple fields
GET /pipelines?sort=status,-created_at
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "resource_not_found",
  "message": "Pipeline with ID 'pipe_abc' not found",
  "details": {
    "resource": "pipeline",
    "id": "pipe_abc"
  },
  "request_id": "req_xyz789"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

### Common Errors

**401 Unauthorized:**
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired token"
}
```

**400 Bad Request:**
```json
{
  "error": "validation_error",
  "message": "Invalid request body",
  "details": {
    "field": "name",
    "error": "must be 3-50 characters"
  }
}
```

---

## WebSocket API

For real-time updates, connect to the WebSocket endpoint:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/ws?token=' + token);

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Event:', data.type);
    console.log('Data:', data.payload);
};
```

### Authentication

Pass token as query parameter for WebSocket authentication:
```
ws://localhost:8080/api/ws?token=<bearer-token>
```

### Event Types

| Event | Description | Payload |
|-------|-------------|---------|
| `pipeline.started` | Pipeline started | `{pipeline_id, name, status}` |
| `pipeline.stopped` | Pipeline stopped | `{pipeline_id, name, status}` |
| `pipeline.completed` | Pipeline finished | `{pipeline_id, name, result}` |
| `pipeline.failed` | Pipeline failed | `{pipeline_id, error}` |
| `job.started` | Job started | `{job_id, pipeline_id, name}` |
| `job.completed` | Job finished | `{job_id, result, duration}` |
| `job.failed` | Job failed | `{job_id, error}` |
| `runner.online` | Runner connected | `{runner_id, name}` |
| `runner.offline` | Runner disconnected | `{runner_id}` |

### Subscribe to Events

Subscribe to specific event types:

```javascript
ws.send(JSON.stringify({
    action: 'subscribe',
    events: ['pipeline.started', 'pipeline.completed']
}));
```

---

## SDK Examples

### Go SDK

```go
package main

import (
    "fmt"
    vimic "github.com/stsgym/vimic-sdk-go"
)

func main() {
    client := vimic.NewClient("http://localhost:8080/api")
    client.SetToken("your-token-here")

    // List pipelines
    pipelines, err := client.Pipelines.List()
    if err != nil {
        panic(err)
    }
    
    for _, p := range pipelines {
        fmt.Printf("%s: %s\n", p.ID, p.Name)
    }
    
    // Start a pipeline
    job, err := client.Pipelines.Start(pipelines[0].ID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Started job: %s\n", job.ID)
}
```

### Python SDK

```python
from vimic import VimicClient

client = VimicClient("http://localhost:8080/api", token="your-token")

# List pipelines
pipelines = client.pipelines.list()
for p in pipelines:
    print(f"{p.id}: {p.name}")

# Start a pipeline
job = client.pipelines.start(pipelines[0].id)
print(f"Started job: {job.id}")
```

### JavaScript SDK

```javascript
import { VimicClient } from '@stsgym/vimic-sdk';

const client = new VimicClient('http://localhost:8080/api', { token: 'your-token' });

// List pipelines
const pipelines = await client.pipelines.list();
for (const p of pipelines) {
    console.log(`${p.id}: ${p.name}`);
}

// Start a pipeline
const job = await client.pipelines.start(pipelines[0].id);
console.log(`Started job: ${job.id}`);
```

---

## Postman Collection

Import our Postman collection for easy API exploration:

1. Download `vimic2-api.postman_collection.json` from `/docs/postman/`
2. Import into Postman
3. Set environment variable `baseUrl` to your Vimic2 server
4. Set `token` from login response

---

## curl Examples

### Health Check

```bash
curl http://localhost:8080/api/health
```

### List Pipelines

```bash
curl http://localhost:8080/api/pipelines \
  -H "Authorization: Bearer $TOKEN"
```

### Create Pipeline

```bash
curl -X POST http://localhost:8080/api/pipelines \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-pipeline",
    "config": {
      "stages": ["build", "test", "deploy"],
      "timeout": 3600
    }
  }'
```

### Get Pipeline

```bash
curl http://localhost:8080/api/pipelines/pipe_abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Start Pipeline

```bash
curl -X POST http://localhost:8080/api/pipelines/pipe_abc123/start \
  -H "Authorization: Bearer $TOKEN"
```

### List Runners

```bash
curl http://localhost:8080/api/runners \
  -H "Authorization: Bearer $TOKEN"
```

### Register Runner

```bash
curl -X POST http://localhost:8080/api/runners \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "gitlab",
    "name": "prod-runner-1",
    "platform": "linux/amd64",
    "labels": ["docker", "production"]
  }'
```

### Acquire VM from Pool

```bash
curl -X POST http://localhost:8080/api/pools/linux-pool/acquire \
  -H "Authorization: Bearer $TOKEN"
```

---

## OpenAPI Specification

The full OpenAPI 3.0 specification is available at:
- Local: `http://localhost:8080/api/openapi.yaml`
- Raw: `https://vimic2.stsgym.com/api/openapi.yaml`

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-04-03 | Initial API release |
