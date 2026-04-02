# Vimic2 CI/CD Orchestration - Future Enhancements

**Status: Core System Complete (7 Phases)**

This document outlines optional future enhancements organized as additional phases. Each phase includes detailed implementation plans, technical specifications, and expected outcomes.

---

## Phase 8: Docker Container Support

### Overview

Add Docker container-based runners as an alternative to full VM isolation. This provides faster startup times and lower resource overhead for workloads that don't require full VM isolation.

### Duration: Week 10-11

### Goals

| Goal | Description |
|------|-------------|
| Container Runners | Support Docker-based ephemeral runners |
| Hybrid Mode | Allow VM + Container runners in same pipeline |
| Resource Limits | CPU, memory, and I/O constraints per container |
| Network Namespaces | Isolate containers using CNI plugins |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    DOCKER RUNNER ARCHITECTURE                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     RUNNER MANAGER                          ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      ││
│  │  │  VM Runner   │  │ Docker Runner│  │  K8s Runner  │      ││
│  │  │   (Phase 4)  │  │  (Phase 8)   │  │  (Phase 9)   │      ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘      ││
│  └─────────────────────────────────────────────────────────────┘│
│                              │                                   │
│  ┌───────────────────────────┴───────────────────────────────┐  │
│  │                   CONTAINER MANAGER                         │  │
│  │   CreateContainer • DestroyContainer • ExecContainer       │  │
│  │   PullImage • PushImage • BuildImage                        │  │
│  └───────────────────────────────────────────────────────────────┘│
│                              │                                   │
│  ┌───────────────────────────┴───────────────────────────────┐  │
│  │                     DOCKER DAEMON                           │  │
│  │   Container Runtime + Image Cache + Volume Manager         │  │
│  └───────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/container/manager.go` | Container lifecycle management |
| `internal/container/docker.go` | Docker client wrapper |
| `internal/container/podman.go` | Podman client (rootless alternative) |
| `internal/container/network.go` | CNI network configuration |
| `internal/container/resource.go` | Resource limit management |
| `internal/container/image.go` | Image pull/build/push |
| `internal/container/container_test.go` | Container tests |

### Database Schema

```sql
-- Add container support to runners table
ALTER TABLE runners ADD COLUMN runtime_type TEXT DEFAULT 'vm';
ALTER TABLE runners ADD COLUMN container_image TEXT;
ALTER TABLE runners ADD COLUMN resource_limits TEXT; -- JSON

-- New containers table
CREATE TABLE containers (
    id TEXT PRIMARY KEY,
    runner_id TEXT NOT NULL,
    image TEXT NOT NULL,
    status TEXT NOT NULL,
    container_id TEXT,
    ip_address TEXT,
    resource_limits TEXT,
    volumes TEXT,
    environment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    FOREIGN KEY (runner_id) REFERENCES runners(id) ON DELETE CASCADE
);
```

### Configuration

```yaml
# config.yaml - Container settings
containers:
  enabled: true
  runtime: docker  # docker | podman
  
  networks:
    type: bridge
    subnet: 10.200.0.0/16
    mtu: 1500
    
  resource_defaults:
    cpu_quota: 50000    # 50%
    memory: 2g
    pids_limit: 100
    
  images:
    cache_dir: /var/lib/vimic2/containers/cache
    registry_mirror: https://mirror.gcr.io
    
  security:
    read_only_root: true
    no_new_privileges: true
    seccomp_profile: /etc/vimic2/seccomp.json
    apparmor_profile: vimic2-container
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/containers` | Create container |
| GET | `/api/containers/{id}` | Get container status |
| POST | `/api/containers/{id}/start` | Start container |
| POST | `/api/containers/{id}/stop` | Stop container |
| DELETE | `/api/containers/{id}` | Destroy container |
| POST | `/api/containers/{id}/exec` | Execute command |
| GET | `/api/containers/{id}/logs` | Stream logs |

### Implementation Steps

1. **Week 10**: Container Manager
   - [ ] Create `internal/container/manager.go`
   - [ ] Implement Docker client wrapper
   - [ ] Add Podman client support
   - [ ] Create database schema migrations
   - [ ] Write container tests

2. **Week 11**: Integration
   - [ ] Add `runtime_type` to runner creation
   - [ ] Update pool manager for containers
   - [ ] Add API endpoints
   - [ ] Update Web UI for container view
   - [ ] Integration tests

### Testing

```bash
# Unit tests
go test ./internal/container/... -v

# Integration tests
docker run -d --name vimic2-test-docker -v /var/run/docker.sock:/var/run/docker.sock vimic2:test
go test ./internal/container/... -tags=integration -v
```

---

## Phase 9: Kubernetes Integration

### Overview

Deploy Vimic2 runners on Kubernetes for cloud-native CI/CD. This enables horizontal scaling, fault tolerance, and integration with existing Kubernetes clusters.

### Duration: Week 12-14

### Goals

| Goal | Description |
|------|-------------|
| K8s Runners | Deploy runners as Kubernetes Pods |
| Auto-Scaling | HPA-based runner scaling |
| Helm Charts | Package for easy deployment |
| CRDs | Custom Resource Definitions for pipelines |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    KUBERNETES ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    KUBERNETES CLUSTER                        ││
│  │                                                             ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      ││
│  │  │    Namespace │  │    Namespace │  │    Namespace │      ││
│  │  │  vimic2-gitlab│  │  vimic2-github│  │vimic2-jenkins│     ││
│  │  │              │  │              │  │              │      ││
│  │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │      ││
│  │  │ │  Runner  │ │  │ │  Runner  │ │  │ │  Runner  │ │      ││
│  │  │ │  Pod     │ │  │ │  Pod     │ │  │ │  Pod     │ │      ││
│  │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │      ││
│  │  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────┐ │      ││
│  │  │ │  Runner  │ │  │ │  Runner  │ │  │ │  Runner  │ │      ││
│  │  │ │  Pod     │ │  │ │  Pod     │ │  │ │  Pod     │ │      ││
│  │  │ └──────────┘ │  │ └──────────┘ │  │ └──────────┘ │      ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘      ││
│  │                                                             ││
│  │  ┌─────────────────────────────────────────────────────┐   ││
│  │  │           VIMIC2 CONTROLLER                          │   ││
│  │  │   Pipeline CRD • Runner CRD • Pool CRD              │   ││
│  │  └─────────────────────────────────────────────────────┘   ││
│  │                                                             ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   VIMIC2 CONTROL PLANE                      ││
│  │   API Server • Web UI • Coordinator • Metrics               ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/kubernetes/controller.go` | Main controller |
| `internal/kubernetes/pipeline_crd.go` | Pipeline CRD handler |
| `internal/kubernetes/runner_crd.go` | Runner CRD handler |
| `internal/kubernetes/pool_crd.go` | Pool CRD handler |
| `internal/kubernetes/client.go` | Kubernetes client wrapper |
| `internal/kubernetes/scaler.go` | HPA management |
| `deploy/helm/vimic2/` | Helm chart |
| `deploy/k8s/crds/` | CRD definitions |

### CRD Definitions

```yaml
# deploy/k8s/crds/pipeline-crd.yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pipelines.vimic2.io
spec:
  group: vimic2.io
  names:
    kind: Pipeline
    plural: pipelines
    singular: pipeline
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              platform:
                type: string
                enum: [gitlab, github, jenkins, circleci, drone]
              repository:
                type: string
              branch:
                type: string
              runnerCount:
                type: integer
                minimum: 1
                maximum: 100
              labels:
                type: array
                items:
                  type: string
          status:
            type: object
            properties:
              phase:
                type: string
                enum: [Creating, Running, Stopped, Destroyed]
              runners:
                type: integer
              networkId:
                type: string
```

### Helm Chart Structure

```
deploy/helm/vimic2/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── ingress.yaml
│   ├── crds/
│   │   ├── pipeline-crd.yaml
│   │   ├── runner-crd.yaml
│   │   └── pool-crd.yaml
│   ├── rbac/
│   │   ├── role.yaml
│   │   ├── rolebinding.yaml
│   │   └── serviceaccount.yaml
│   └── notes.txt
└── values/
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

### Implementation Steps

1. **Week 12**: Controller Foundation
   - [ ] Create Kubernetes client wrapper
   - [ ] Implement CRD definitions
   - [ ] Create Helm chart skeleton
   - [ ] Set up RBAC manifests

2. **Week 13**: Controller Logic
   - [ ] Implement pipeline CRD handler
   - [ ] Implement runner CRD handler
   - [ ] Implement pool CRD handler
   - [ ] Add reconciliation loops

3. **Week 14**: Deployment
   - [ ] Create deployment manifests
   - [ ] Add auto-scaling support
   - [ ] Write integration tests
   - [ ] Document deployment guide

### Deployment Example

```bash
# Install via Helm
helm install vimic2 deploy/helm/vimic2 \
  --namespace vimic2 \
  --create-namespace \
  --set image.tag=v1.0.0 \
  --set platforms.gitlab.enabled=true \
  --set platforms.gitlab.token=${GITLAB_TOKEN}

# Create pipeline via CRD
kubectl apply -f - <<EOF
apiVersion: vimic2.io/v1
kind: Pipeline
metadata:
  name: build-pipeline
  namespace: vimic2-gitlab
spec:
  platform: gitlab
  repository: https://gitlab.example.com/project
  branch: main
  runnerCount: 5
  labels: [builder, go, docker]
EOF
```

---

## Phase 10: Auto-Scaling

### Overview

Implement dynamic pool sizing based on queue depth and workload patterns. This optimizes resource usage while maintaining pipeline responsiveness.

### Duration: Week 15-16

### Goals

| Goal | Description |
|------|-------------|
| Queue Monitoring | Track job queue depth per platform |
| Predictive Scaling | ML-based demand prediction |
| Cost Optimization | Scale down during low usage |
| SLA Management | Maintain target wait times |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    AUTO-SCALING ARCHITECTURE                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    METRICS COLLECTOR                         ││
│  │   QueueDepth • ActiveRunners • WaitTime • CPU/Memory        ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    SCALING ENGINE                           ││
│  │   ThresholdPolicy • PredictivePolicy • CostPolicy          ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    POOL MANAGER                             ││
│  │   ScaleUp • ScaleDown • PreAllocateVMs • Rebalance         ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/scaling/metrics.go` | Metrics collection |
| `internal/scaling/engine.go` | Scaling decision engine |
| `internal/scaling/policies.go` | Scaling policies |
| `internal/scaling/predictor.go` | ML-based prediction |
| `internal/scaling/cost.go` | Cost optimization |
| `internal/scaling/scaling_test.go` | Scaling tests |

### Scaling Policies

```go
// ScalingPolicy defines how to scale
type ScalingPolicy interface {
    // Calculate desired size based on metrics
    Calculate(currentSize int, metrics *Metrics) int
}

// ThresholdPolicy scales based on simple thresholds
type ThresholdPolicy struct {
    ScaleUpThreshold   float64 // Queue depth per runner
    ScaleDownThreshold float64
    MinSize            int
    MaxSize            int
    CooldownPeriod     time.Duration
}

// PredictivePolicy uses ML for demand prediction
type PredictivePolicy struct {
    Model              *MLModel
    LookaheadWindow    time.Duration
    MinSize            int
    MaxSize            int
}

// CostPolicy optimizes for cost efficiency
type CostPolicy struct {
    Budget             float64
    SpotInstanceRatio  float64
    ReservedRatio      float64
    MinSize            int
    MaxSize            int
}
```

### Configuration

```yaml
# config.yaml - Scaling settings
scaling:
  enabled: true
  policy: predictive  # threshold | predictive | cost
  
  threshold:
    scale_up_threshold: 2.0    # Queue depth per runner
    scale_down_threshold: 0.5
    cooldown_period: 5m
    evaluation_interval: 30s
    
  predictive:
    model_path: /var/lib/vimic2/models/scaling.model
    lookahead_window: 15m
    training_interval: 24h
    
  cost:
    budget: 1000.0  # USD per month
    spot_instance_ratio: 0.5
    reserved_ratio: 0.2
    
  limits:
    min_pool_size: 2
    max_pool_size: 50
    scale_up_step: 5
    scale_down_step: 2
```

### Metrics

```go
// Metrics collected for scaling decisions
type Metrics struct {
    // Queue metrics
    QueueDepth       int
    QueueWaitTime    time.Duration
    QueueArrivalRate float64
    
    // Runner metrics
    ActiveRunners    int
    IdleRunners      int
    BusyRunners      int
    RunnerUtilization float64
    
    // Job metrics
    JobsCompleted    int
    JobsFailed       int
    AvgJobDuration   time.Duration
    
    // Resource metrics
    CPUUtilization   float64
    MemoryUtilization float64
    NetworkThroughput float64
    
    // Time metrics
    Timestamp        time.Time
    TimeOfDay        time.Time
    DayOfWeek        time.Weekday
}
```

### Implementation Steps

1. **Week 15**: Metrics & Policies
   - [ ] Create metrics collector
   - [ ] Implement threshold policy
   - [ ] Add cost policy
   - [ ] Write scaling tests

2. **Week 16**: Predictive Scaling
   - [ ] Create ML model for prediction
   - [ ] Implement predictor
   - [ ] Add training pipeline
   - [ ] Integration tests

---

## Phase 11: Webhook & Notification System

### Overview

Add external notifications for pipeline events. Support multiple channels including Slack, email, PagerDuty, and generic webhooks.

### Duration: Week 17-18

### Goals

| Goal | Description |
|------|-------------|
| Multi-Channel | Slack, email, PagerDuty, generic |
| Event Filtering | Configure which events trigger notifications |
| Templates | Customizable notification templates |
| Rate Limiting | Prevent notification spam |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    NOTIFICATION ARCHITECTURE                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    EVENT BUS                                 ││
│  │   PipelineCreated • JobCompleted • JobFailed • ...          ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    NOTIFICATION MANAGER                     ││
│  │   EventFilter • RateLimiter • TemplateRenderer              ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┐          │
│  │  Slack  │  Email  │PagerDuty│ Webhook │  SMS    │          │
│  │ Channel │ Channel │ Channel │ Channel │ Channel │          │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴──────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/notification/manager.go` | Notification manager |
| `internal/notification/bus.go` | Event bus |
| `internal/notification/channels.go` | Channel interfaces |
| `internal/notification/slack.go` | Slack integration |
| `internal/notification/email.go` | Email integration |
| `internal/notification/pagerduty.go` | PagerDuty integration |
| `internal/notification/webhook.go` | Generic webhook |
| `internal/notification/templates.go` | Template rendering |
| `internal/notification/notification_test.go` | Tests |

### Configuration

```yaml
# config.yaml - Notification settings
notifications:
  enabled: true
  
  # Event subscriptions
  events:
    - name: pipeline_created
      channels: [slack, email]
      filter: "platform == 'gitlab'"
    - name: job_failed
      channels: [slack, pagerduty]
      filter: "severity == 'critical'"
    - name: pipeline_completed
      channels: [slack, email]
      filter: "duration > 10m"
  
  # Channel configurations
  channels:
    slack:
      enabled: true
      webhook_url: ${SLACK_WEBHOOK_URL}
      channel: "#ci-cd"
      username: "Vimic2 Bot"
      icon_emoji: ":robot:"
      
    email:
      enabled: true
      smtp_host: smtp.example.com
      smtp_port: 587
      smtp_user: ${SMTP_USER}
      smtp_pass: ${SMTP_PASS}
      from: vimic2@example.com
      to: [team@example.com]
      
    pagerduty:
      enabled: true
      service_key: ${PAGERDUTY_KEY}
      severity_map:
        critical: critical
        warning: warning
        info: info
        
    webhook:
      enabled: true
      url: https://example.com/webhook
      headers:
        Authorization: "Bearer ${WEBHOOK_TOKEN}"
      retry_count: 3
      retry_delay: 5s
      
  # Rate limiting
  rate_limit:
    enabled: true
    max_per_minute: 10
    max_per_hour: 100
    burst_size: 20
```

### Notification Template

```yaml
# Template for Slack notification
templates:
  slack:
    pipeline_created: |
      :rocket: *Pipeline Created*
      • Platform: {{.Platform}}
      • Repository: {{.Repository}}
      • Branch: {{.Branch}}
      • Commit: `{{.CommitSHA | truncate 8}}`
      • Author: {{.Author}}
      
    job_failed: |
      :x: *Job Failed*
      • Pipeline: {{.PipelineID}}
      • Job: {{.JobName}}
      • Error: {{.ErrorMessage}}
      • Duration: {{.Duration}}
      • <{{.LogURL}}|View Logs>
```

### Implementation Steps

1. **Week 17**: Core System
   - [ ] Create event bus
   - [ ] Implement notification manager
   - [ ] Add channel interfaces
   - [ ] Create Slack integration

2. **Week 18**: Channels
   - [ ] Add email integration
   - [ ] Add PagerDuty integration
   - [ ] Add generic webhook
   - [ ] Template rendering
   - [ ] Integration tests

---

## Phase 12: Metrics & Observability

### Overview

Add Prometheus/OpenMetrics export for comprehensive monitoring. Include Grafana dashboards and alerting rules.

### Duration: Week 19-20

### Goals

| Goal | Description |
|------|-------------|
| Prometheus Export | OpenMetrics-compatible metrics |
| Grafana Dashboards | Pre-built monitoring dashboards |
| Alerting Rules | AlertManager configuration |
| Trace Integration | OpenTelemetry support |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    OBSERVABILITY ARCHITECTURE                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    VIMIC2 COMPONENTS                         ││
│  │                                                             ││
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    ││
│  │  │Coordinator│  │Dispatcher│  │PoolMgr  │  │ RunnerMgr │    ││
│  │  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘    ││
│  │        │            │            │            │            ││
│  │        └────────────┴────────────┴────────────┘            ││
│  │                          │                                  ││
│  │  ┌───────────────────────┴───────────────────────────────┐ ││
│  │  │                   METRICS EXPORTER                     │ ││
│  │  │   Counters • Gauges • Histograms • Summaries         │ ││
│  │  └───────────────────────────────────────────────────────┘ ││
│  └─────────────────────────────────────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    PROMETHEUS                                ││
│  │   Scrape • Store • Query • Alert                            ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    GRAFANA                                   ││
│  │   Dashboards • Visualization • Alerting                    ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/metrics/exporter.go` | Prometheus exporter |
| `internal/metrics/collector.go` | Metrics collection |
| `internal/metrics/pipeline_metrics.go` | Pipeline metrics |
| `internal/metrics/pool_metrics.go` | Pool metrics |
| `internal/metrics/runner_metrics.go` | Runner metrics |
| `internal/metrics/network_metrics.go` | Network metrics |
| `internal/tracing/tracer.go` | OpenTelemetry tracer |
| `deploy/grafana/dashboards/` | Grafana dashboards |
| `deploy/prometheus/rules/` | Alerting rules |

### Metrics Exposed

```promql
# Pipeline metrics
vimic2_pipelines_total{platform="gitlab"} 1234
vimic2_pipelines_running{platform="gitlab"} 5
vimic2_pipelines_duration_seconds{platform="gitlab",status="success"} histogram
vimic2_pipelines_queue_depth{platform="gitlab"} 10

# Job metrics
vimic2_jobs_total{platform="gitlab",status="success"} 5678
vimic2_jobs_duration_seconds{platform="gitlab"} histogram
vimic2_jobs_queued{platform="gitlab"} 15

# Runner metrics
vimic2_runners_total{platform="gitlab",status="online"} 12
vimic2_runners_jobs_completed_total{runner_id="abc123"} 100

# Pool metrics
vimic2_pool_size{pool="builder"} 10
vimic2_pool_vms_available{pool="builder"} 3
vimic2_pool_vms_busy{pool="builder"} 7

# Network metrics
vimic2_networks_total 5
vimic2_network_vlans_allocated 5
vimic2_network_ips_allocated 50

# System metrics
vimic2_api_requests_total{endpoint="/api/pipelines",method="GET"} 10000
vimic2_api_request_duration_seconds{endpoint="/api/pipelines"} histogram
vimic2_database_queries_total{query="list_pipelines"} 5000
vimic2_database_query_duration_seconds{query="list_pipelines"} histogram
```

### Grafana Dashboards

```json
// deploy/grafana/dashboards/vimic2-overview.json
{
  "title": "Vimic2 Overview",
  "panels": [
    {
      "title": "Pipelines",
      "type": "graph",
      "targets": [
        {
          "expr": "vimic2_pipelines_running",
          "legendFormat": "{{platform}}"
        }
      ]
    },
    {
      "title": "Job Duration",
      "type": "heatmap",
      "targets": [
        {
          "expr": "rate(vimic2_jobs_duration_seconds_bucket[5m])",
          "format": "heatmap"
        }
      ]
    },
    {
      "title": "Pool Utilization",
      "type": "gauge",
      "targets": [
        {
          "expr": "vimic2_pool_vms_busy / vimic2_pool_size * 100",
          "legendFormat": "{{pool}}"
        }
      ]
    }
  ]
}
```

### Alerting Rules

```yaml
# deploy/prometheus/rules/vimic2-alerts.yaml
groups:
  - name: vimic2
    rules:
      - alert: HighQueueDepth
        expr: vimic2_pipelines_queue_depth > 20
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High pipeline queue depth"
          description: "Queue depth is {{ $value }} for platform {{ $labels.platform }}"
          
      - alert: NoAvailableRunners
        expr: vimic2_runners_total{status="online"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "No runners available"
          description: "All runners are offline for platform {{ $labels.platform }}"
          
      - alert: PoolExhausted
        expr: vimic2_pool_vms_available == 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "VM pool exhausted"
          description: "No VMs available in pool {{ $labels.pool }}"
```

### Implementation Steps

1. **Week 19**: Metrics Export
   - [ ] Create metrics exporter
   - [ ] Implement collectors for each component
   - [ ] Add OpenTelemetry tracing
   - [ ] Test metrics endpoint

2. **Week 20**: Dashboards & Alerts
   - [ ] Create Grafana dashboards
   - [ ] Add Prometheus alert rules
   - [ ] Document monitoring setup
   - [ ] Integration tests

---

## Phase 13: Distributed Storage Backend

### Overview

Add S3/GCS/Azure Blob storage backends for artifacts. Enable distributed artifact storage across regions.

### Duration: Week 21-22

### Goals

| Goal | Description |
|------|-------------|
| Multi-Backend | S3, GCS, Azure Blob, local filesystem |
| Encryption | Server-side and client-side encryption |
| CDN Integration | CloudFront, Cloud CDN for distribution |
| Retention Policies | TTL-based artifact cleanup |

### Technical Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    STORAGE ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    ARTIFACT MANAGER                          ││
│  │   Upload • Download • Delete • List                         ││
│  └──────────────────────────────┬──────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    STORAGE BACKEND                          ││
│  │                                                             ││
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    ││
│  │  │   Local  │  │    S3    │  │   GCS    │  │Azure Blob│    ││
│  │  │  (base)  │  │  (new)   │  │  (new)   │  │  (new)   │    ││
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘    ││
│  │                                                             ││
│  └─────────────────────────────────────────────────────────────┘│
│                                │                                │
│  ┌─────────────────────────────┴──────────────────────────────┐│
│  │                    CDN LAYER                                ││
│  │   CloudFront • Cloud CDN • Azure CDN                       ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Files to Create

| File | Purpose |
|------|---------|
| `internal/storage/backend.go` | Backend interface |
| `internal/storage/local.go` | Local filesystem (existing) |
| `internal/storage/s3.go` | S3 backend |
| `internal/storage/gcs.go` | GCS backend |
| `internal/storage/azure.go` | Azure Blob backend |
| `internal/storage/encryption.go` | Client-side encryption |
| `internal/storage/cdn.go` | CDN integration |
| `internal/storage/storage_test.go` | Storage tests |

### Configuration

```yaml
# config.yaml - Storage settings
storage:
  backend: s3  # local | s3 | gcs | azure
  
  s3:
    bucket: vimic2-artifacts
    region: us-west-2
    endpoint: https://s3.amazonaws.com
    access_key: ${AWS_ACCESS_KEY}
    secret_key: ${AWS_SECRET_KEY}
    encryption: AES256
    storage_class: STANDARD_IA
    
  gcs:
    bucket: vimic2-artifacts
    credentials_file: /etc/vimic2/gcs-credentials.json
    encryption_key: ${GCS_ENCRYPTION_KEY}
    storage_class: NEARLINE
    
  azure:
    account: vimic2artifacts
    container: artifacts
    credentials_file: /etc/vimic2/azure-credentials.json
    encryption_scope: vimic2-encryption
    
  cdn:
    enabled: true
    provider: cloudfront
    distribution_id: ${CLOUDFRONT_DISTRIBUTION}
    cache_ttl: 3600
    
  retention:
    enabled: true
    default_ttl: 30d
    max_ttl: 90d
    cleanup_interval: 6h
```

### Implementation Steps

1. **Week 21**: Backends
   - [ ] Create storage backend interface
   - [ ] Implement S3 backend
   - [ ] Implement GCS backend
   - [ ] Implement Azure backend
   - [ ] Add encryption support

2. **Week 22**: CDN & Retention
   - [ ] Add CDN integration
   - [ ] Implement retention policies
   - [ ] Add cleanup job
   - [ ] Integration tests

---

## Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| 8 | Week 10-11 | Docker container support |
| 9 | Week 12-14 | Kubernetes integration |
| 10 | Week 15-16 | Auto-scaling |
| 11 | Week 17-18 | Notifications |
| 12 | Week 19-20 | Metrics & observability |
| 13 | Week 21-22 | Distributed storage |

### Dependencies

```
Phase 8 ──┐
          ├──▶ Phase 9 ──┐
Phase 4 ──┘              │
                         ├──▶ Phase 10
Phase 5 ─────────────────┘
                    │
                    ├──▶ Phase 11
                    │
                    ├──▶ Phase 12
                    │
Phase 5 ────────────┴──▶ Phase 13
```

### Prioritization

| Priority | Phase | Reason |
|----------|-------|--------|
| **P1** | 10 | Auto-scaling is critical for production |
| **P1** | 12 | Observability is essential for operations |
| **P2** | 9 | Kubernetes enables cloud-native deployments |
| **P2** | 13 | Distributed storage enables multi-region |
| **P3** | 8 | Docker provides lighter-weight alternative |
| **P3** | 11 | Notifications are nice-to-have |

---

*Last Updated: 2026-04-02*
*Status: Core system complete, future phases documented*