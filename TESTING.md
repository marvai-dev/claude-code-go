# Testing the Claude Code Go SDK

This document provides comprehensive guidance for testing the Claude Code Go SDK locally and setting up integration tests.

## Prerequisites

Before testing, ensure you have:

1. **Go 1.24.2+** installed
2. **Claude Code CLI** installed and accessible
3. **Task** installed for running build tasks

Note: Claude Code CLI will handle authentication automatically when you run it for the first time.

## Quick Start Testing

```bash
# Run all tests
task test

# Run with coverage
task test-coverage

# Run full pipeline
task all
```

## Local Testing Approaches

### 1. Unit Tests (Current)

Our unit tests use mocked commands and don't require actual Claude Code CLI:

```bash
# Run unit tests
go test ./pkg/claude -v

# Run specific test
go test ./pkg/claude -run TestRunPrompt -v

# Run all tests
go test ./...
```

### 2. Integration Tests with Mock Server

For testing without consuming API credits, we provide a mock Claude server:

```bash
# Run integration tests with mock
task test-integration-mock

# Run integration tests with real Claude (requires API key)
task test-integration-real
```

### 3. Manual Testing

Test the examples and demos:

```bash
# Build all examples (outputs to bin/ directory)
task build-examples

# Run the interactive demo
task demo

# Test basic example manually (requires Claude Code CLI)
./bin/basic-example

# Test advanced example manually
./bin/advanced-example
```

**⚠️ Important:** Always use `task build-examples` or `make build-examples` instead of manual `go build` commands to ensure binaries are placed in the `bin/` directory.

## Setting Up Integration Tests

### Option 1: Mock Claude Server

We provide a mock server that simulates Claude Code responses:

```bash
# Start mock server
task start-mock-server

# Run integration tests against mock
task test-integration
```

### Option 2: Real Claude Code CLI

For testing against actual Claude:

```bash
# Ensure Claude Code CLI is installed
claude --help

# Run integration tests (Claude will handle auth automatically)
task test-integration-real
```

## Test Categories

### Unit Tests (`pkg/claude/`)
- Test core SDK functionality
- Mock command execution
- Validate argument building
- Test convenience methods

### Integration Tests (`test/integration/`)
- Test full end-to-end workflows
- Validate CLI binary behavior
- Test MCP integration
- Test streaming functionality

### Example Tests (`examples/`)
- Validate example code works
- Test different usage patterns
- Verify documentation examples

## Testing Strategies

### 1. Isolated Testing (Recommended for CI)

```bash
# Uses mocked Claude responses
go test ./... -tags=mock
```

### 2. Local Integration Testing

```bash
# Uses mock server (no API costs)
task test-integration-mock

# Uses real Claude (costs API credits)
task test-integration-real
```

### 3. End-to-End Testing

```bash
# Full pipeline test with real Claude
task test-e2e
```

## Test Configuration

### Environment Variables

```bash
# For integration tests
export CLAUDE_CODE_PATH="/path/to/claude"        # Custom Claude CLI path
export TEST_TIMEOUT="30s"                       # Test timeout
export MOCK_SERVER_PORT="8080"                  # Mock server port
export USE_MOCK_SERVER="1"                      # Use mock server instead of real Claude CLI
```

### Test Flags

```bash
# Skip slow tests
go test -short ./...

# Run only integration tests
go test -tags=integration ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Creating New Integration Tests

### 1. Basic Integration Test

```go
// test/integration/basic_test.go
//go:build integration
// +build integration

package integration

import (
    "testing"
    "github.com/lancekrogers/claude-code-go/pkg/claude"
)

func TestBasicPrompt(t *testing.T) {
    client := claude.NewClient("claude")
    
    result, err := client.RunPrompt("What is 2+2?", &claude.RunOptions{
        Format: claude.JSONOutput,
    })
    
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if result.Result == "" {
        t.Error("Expected non-empty result")
    }
    
    if result.CostUSD <= 0 {
        t.Error("Expected positive cost")
    }
}
```

### 2. Mock Integration Test

```go
// test/integration/mock_test.go
func TestWithMockServer(t *testing.T) {
    // Start mock server
    mockServer := startMockServer(t)
    defer mockServer.Close()
    
    client := claude.NewClient(mockServer.URL + "/claude")
    
    result, err := client.RunPrompt("test prompt", nil)
    if err != nil {
        t.Fatalf("Mock test failed: %v", err)
    }
    
    // Validate mock response
    if !strings.Contains(result.Result, "mock response") {
        t.Error("Expected mock response")
    }
}
```

## Test Data Management

### Mock Responses

Store mock responses in `test/fixtures/`:

```bash
test/fixtures/
├── responses/
│   ├── simple_prompt.json
│   ├── streaming_response.jsonl
│   └── error_response.json
└── configs/
    ├── mcp_config.json
    └── test_config.json
```

### Test Utilities

Common testing utilities in `test/utils/`:

```go
// test/utils/client.go
func NewTestClient(t *testing.T) *claude.ClaudeClient {
    return claude.NewClient(getTestClaudePath(t))
}

func getTestClaudePath(t *testing.T) string {
    if path := os.Getenv("CLAUDE_CODE_PATH"); path != "" {
        return path
    }
    return "claude" // Default
}
```

## Debugging Tests

### Verbose Output

```bash
# Show detailed test output
go test -v ./...

# Show test coverage
go test -v -cover ./...

# Show test with race detection
go test -v -race ./...
```

### Test Debugging

```go
func TestDebugExample(t *testing.T) {
    // Enable debug logging
    client := claude.NewClient("claude")
    
    // Log the command that would be executed
    t.Logf("Testing with client: %+v", client)
    
    result, err := client.RunPrompt("debug test", &claude.RunOptions{
        Verbose: true,
    })
    
    if err != nil {
        t.Logf("Error details: %v", err)
        t.Fail()
    }
    
    t.Logf("Result: %+v", result)
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24.2'
      
      # Unit tests (no external dependencies)
      - name: Run unit tests
        run: go test ./... -tags=mock
      
      # Integration tests with mock server
      - name: Run integration tests
        run: task test-integration-mock
      
      # Build validation
      - name: Build all
        run: task build-all
```

## Performance Testing

### Benchmark Tests

```go
// pkg/claude/benchmark_test.go
func BenchmarkRunPrompt(b *testing.B) {
    client := claude.NewClient("claude")
    
    for i := 0; i < b.N; i++ {
        _, err := client.RunPrompt("benchmark test", nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Load Testing

```bash
# Run benchmark tests
go test -bench=. ./pkg/claude

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./pkg/claude

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/claude
```

## Best Practices

### 1. Test Organization
- Unit tests in `*_test.go` files alongside source
- Integration tests in `test/integration/`
- Benchmarks in `*_benchmark_test.go` files

### 2. Test Naming
- Use descriptive test names: `TestRunPromptWithJSONOutput`
- Use table-driven tests for multiple scenarios
- Use subtests for complex test cases

### 3. Test Isolation
- Each test should be independent
- Use `t.Parallel()` for parallel execution
- Clean up resources in `defer` statements

### 4. Mocking Strategy
- Mock external dependencies (Claude CLI)
- Use dependency injection for testability
- Provide both mock and real test options

### 5. Error Testing
- Test both success and failure cases
- Validate error messages and types
- Test edge cases and boundary conditions

## Troubleshooting

### Common Issues

1. **"claude: command not found"**
   - Install Claude Code CLI from https://docs.anthropic.com/en/docs/claude-code/getting-started
   - Set `CLAUDE_CODE_PATH` environment variable

2. **"Claude Code CLI not working"**
   - Ensure Claude Code CLI is properly installed
   - Try running `claude --help` to verify installation

3. **Test timeouts**
   - Increase test timeout with `-timeout` flag
   - Use mock server for faster tests: `USE_MOCK_SERVER=1`

4. **Permission denied**
   - Check file permissions on test scripts
   - Ensure Claude CLI is executable

### Getting Help

- Check test logs: `go test -v ./...`
- Run with race detection: `go test -race ./...`
- Enable verbose output: `go test -v -cover ./...`
- Use debugger: `dlv test ./pkg/claude`

## Next Steps

1. **Set up your testing environment**
2. **Run the existing tests**
3. **Create integration tests for your use cases**
4. **Add performance benchmarks**
5. **Set up CI/CD pipeline**

For more detailed examples, see the `test/` directory and example files.