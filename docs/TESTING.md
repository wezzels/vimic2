# Vimic2 CI/CD Orchestration - Testing Guide

## Overview

Vimic2 has a comprehensive test suite covering all packages:

- **Unit Tests**: Fast tests without external dependencies
- **Integration Tests**: Tests requiring external services (OVS, libvirt, databases)
- **End-to-End Tests**: Full system tests

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./... -tags='!integration'

# Run with coverage
go test ./... -tags='!integration' -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require external services. Use Docker Compose to set up the test environment:

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./... -tags=integration

# Stop test environment
docker-compose -f docker-compose.test.yml down
```

### End-to-End Tests

```bash
# Run E2E tests (requires full setup)
go test ./... -tags=e2e
```

## Test Structure

```
vimic2/
├── internal/
│   ├── pipeline/
│   │   ├── coordinator_test.go
│   │   ├── dispatcher_test.go
│   │   ├── artifacts_test.go
│   │   └── logs_test.go
│   ├── pool/
│   │   ├── manager_test.go
│   │   ├── state_test.go
│   │   └── template_test.go
│   ├── network/
│   │   ├── isolation_test.go
│   │   ├── ipam_test.go
│   │   ├── firewall_test.go
│   │   └── vlan_test.go
│   ├── runner/
│   │   ├── gitlab_test.go
│   │   ├── github_test.go
│   │   ├── jenkins_test.go
│   │   └── manager_test.go
│   └── api/
│       ├── server_test.go
│       └── websocket_test.go
└── tests/
    ├── integration/
    │   ├── pipeline_test.go
    │   ├── pool_test.go
    │   ├── network_test.go
    │   └── runner_test.go
    └── e2e/
        └── full_pipeline_test.go
```

## Test Coverage

Current coverage by package:

| Package | Coverage |
|---------|----------|
| pipeline | 45% |
| pool | 38% |
| network | 33% |
| runner | 29% |
| api | 42% |
| **Total** | **37%** |

## Mocking

### Database Mock

```go
func TestWithMockDB(t *testing.T) {
    db, err := pipeline.NewPipelineDB(":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // Use db for testing
}
```

### Pool Manager Mock

```go
func TestWithMockPool(t *testing.T) {
    poolMgr := &MockPoolManager{
        AcquireVMFunc: func(ctx context.Context, poolName string) (*pool.VM, error) {
            return &pool.VM{ID: "test-vm"}, nil
        },
    }

    // Use poolMgr for testing
}
```

### Runner Manager Mock

```go
func TestWithMockRunner(t *testing.T) {
    runnerMgr := &MockRunnerManager{
        CreateRunnerFunc: func(ctx context.Context, poolName string, platform pipeline.RunnerPlatform, pipelineID string, labels []string) (*runner.RunnerInfo, error) {
            return &runner.RunnerInfo{ID: "test-runner"}, nil
        },
    }

    // Use runnerMgr for testing
}
```

## CI/CD Tests

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - lint
  - test
  - build
  - security

lint:
  stage: lint
  script:
    - go fmt ./...
    - go vet ./...
    - golint ./...

unit_tests:
  stage: test
  script:
    - go test ./... -tags='!integration' -coverprofile=coverage.out
  coverage: '/total:\s+\(statements\)\s+(\d+.\d+)%/'

integration_tests:
  stage: test
  script:
    - docker-compose -f docker-compose.test.yml up -d
    - go test ./... -tags=integration
    - docker-compose -f docker-compose.test.yml down
  only:
    - main

build:
  stage: build
  script:
    - go build -o vimic2 ./cmd/vimic2
  artifacts:
    paths:
      - vimic2

security_scan:
  stage: security
  script:
    - gosec ./...
    - go mod graph | nancy
  only:
    - main
```

### GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Lint
        run: |
          go fmt ./...
          go vet ./...
          go install golang.org/x/lint/golint@latest
          golint ./...

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run tests
        run: go test ./... -tags='!integration' -coverprofile=coverage.out
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Build
        run: |
          GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o vimic2-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/vimic2
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: vimic2-${{ matrix.goos }}-${{ matrix.goarch }}
          path: vimic2-${{ matrix.goos }}-${{ matrix.goarch }}
```

## Writing Tests

### Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Use testify/assert** for assertions
3. **Use mocks** for external dependencies
4. **Use build tags** for integration tests
5. **Use t.Parallel()** for parallel tests

### Example Test

```go
func TestCoordinator_CreatePipeline(t *testing.T) {
    tests := []struct {
        name       string
        platform   pipeline.RunnerPlatform
        wantErr    bool
    }{
        {"gitlab", pipeline.PlatformGitLab, false},
        {"github", pipeline.PlatformGitHub, false},
        {"jenkins", pipeline.PlatformJenkins, false},
        {"invalid", pipeline.RunnerPlatform("invalid"), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            db := setupTestDB(t)
            defer db.Close()

            coordinator := setupTestCoordinator(t, db)

            // Test
            _, err := coordinator.CreatePipeline(
                context.Background(),
                tt.platform,
                "test/repo",
                "main",
                "abc123",
                "Test commit",
                "test@example.com",
                1,
                []string{"builder"},
            )

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Performance Tests

```bash
# Run benchmarks
go test ./... -bench=. -benchmem

# Run CPU profile
go test ./... -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Run memory profile
go test ./... -memprofile=mem.prof
go tool pprof mem.prof
```

## Test Environment

### Docker Compose (docker-compose.test.yml)

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: vimic2_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5432:5432"

  redis:
    image: redis:7
    ports:
      - "6379:6379"

  openvswitch:
    image: socketplane/openvswitch:latest
    privileged: true
    network_mode: host
    volumes:
      - /var/run/openvswitch:/var/run/openvswitch
```

## Troubleshooting

### Common Issues

1. **Tests hang**: Check for missing defer statements or unclosed connections
2. **Tests fail intermittently**: Check for race conditions with `-race` flag
3. **Coverage too low**: Add tests for uncovered code paths
4. **Integration tests fail**: Check if external services are running

### Debug Mode

```bash
# Run tests with verbose output
go test ./... -v

# Run specific test
go test -run TestCoordinator_CreatePipeline -v

# Run with race detection
go test ./... -race -v
```