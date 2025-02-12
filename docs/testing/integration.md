# Integration Testing Guide

This document describes the integration testing strategy for the Kommon project. Integration tests ensure that different components of the system work together as expected under real-world conditions.

## Overview

Integration tests in Kommon focus on verifying the behavior of different executor types (Local, Docker, and Kubernetes) and their interaction with agents. These tests ensure that:

1. Executors can properly manage agent lifecycles
2. Resource management works correctly
3. Error conditions are handled appropriately
4. System cleanup occurs properly

## Environment Setup

### Prerequisites

To run the integration tests, you need the following components installed:

- Go 1.21 or later
- Docker (for Docker Executor tests)
- Kubernetes cluster access (for Kubernetes Executor tests)
- `kubectl` configured with appropriate cluster access
- `make` command-line tool

### Dependencies

The project uses the following testing libraries:
```go
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
```

### Mock Environment

Some tests use mock implementations of the agent interface for testing. These mocks allow us to:
- Simulate agent behavior without real network calls
- Test error conditions and edge cases
- Verify resource cleanup

## Test Specifications

### Local Executor Tests

The Local Executor tests verify the following functionality:

1. Multiple Agent Management
   - Creating multiple agents
   - Listing active agents
   - Verifying agent status
   - Proper agent cleanup

2. Resource Status Monitoring
   - CPU usage tracking
   - Memory usage tracking
   - Disk usage tracking

3. Error Handling
   - Duplicate agent creation attempts
   - Invalid agent destruction
   - Resource allocation failures

4. Cleanup Process
   - Proper resource release
   - Agent termination
   - System state cleanup

### Docker Executor Tests

The Docker Executor tests verify:

1. Container Lifecycle
   - Container creation
   - Container status monitoring
   - Container termination

2. Resource Management
   - CPU quota enforcement
   - Memory limit enforcement
   - Network allocation

3. Error Scenarios
   - Invalid container configurations
   - Resource constraint violations
   - Network failures

### Kubernetes Executor Tests

The Kubernetes Executor tests verify:

1. Pod Management
   - Pod creation and scheduling
   - Pod status monitoring
   - Pod termination

2. Resource Quotas
   - Namespace resource limits
   - Pod resource requests/limits
   - Storage allocation

3. Error Handling
   - Invalid pod specifications
   - Resource quota violations
   - Scheduling failures

## Running the Tests

### Full Test Suite

To run all integration tests:

```bash
make test
```

### Specific Executor Tests

To run tests for specific executors:

```bash
# Local Executor tests only
go test -v ./pkg/executor -run "TestLocalExecutor"

# Docker Executor tests only
go test -v ./pkg/executor -run "TestDockerExecutor"

# Kubernetes Executor tests only
go test -v ./pkg/executor -run "TestKubernetesExecutor"
```

### Test Tags

Some tests may require specific build tags. For example:

```bash
# Run tests that require Docker
go test -tags=docker -v ./pkg/executor

# Run tests that require Kubernetes
go test -tags=kubernetes -v ./pkg/executor
```

## Verifying Test Results

### Test Output

Test results include:
- Test execution status (PASS/FAIL)
- Execution time
- Resource usage statistics
- Error messages for failed tests

Example successful output:
```
=== RUN   TestLocalExecutorIntegration
--- PASS: TestLocalExecutorIntegration (0.00s)
    --- PASS: TestLocalExecutorIntegration/MultipleAgentsTest (0.00s)
    --- PASS: TestLocalExecutorIntegration/ResourceStatusTest (0.00s)
    --- PASS: TestLocalExecutorIntegration/ErrorHandlingTest (0.00s)
```

### Coverage Reports

To generate test coverage reports:

```bash
go test -coverprofile=coverage.out ./pkg/executor
go tool cover -html=coverage.out
```

## Troubleshooting

### Common Issues

1. Docker Tests Failing
   - Ensure Docker daemon is running
   - Check Docker permissions
   - Verify available system resources

2. Kubernetes Tests Failing
   - Verify cluster access
   - Check namespace permissions
   - Ensure proper KUBECONFIG setup

3. Resource-Related Failures
   - Check system resource availability
   - Verify resource quota configurations
   - Monitor system limits

### Debug Logs

To enable verbose logging during tests:

```bash
go test -v -debug ./pkg/executor
```

### Cleanup

If tests fail unexpectedly:

1. Clean up Docker resources:
```bash
docker system prune -f
```

2. Clean up Kubernetes resources:
```bash
kubectl delete pods,services -l app=kommon-test
```

3. Clean up local resources:
```bash
rm -rf /tmp/kommon-test*
```

## Contributing

When adding new integration tests:

1. Follow the existing test structure
2. Include appropriate cleanup mechanisms
3. Document test requirements
4. Add test cases to this document

## References

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Package](https://github.com/stretchr/testify)
- [Docker Testing Best Practices](https://docs.docker.com/develop/test-driven/)
- [Kubernetes Testing](https://kubernetes.io/docs/concepts/testing/)