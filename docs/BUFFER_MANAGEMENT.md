# Buffer Management

This document describes the enhanced buffer management system in claude-code-go, designed to address buffer size issues and provide robust, configurable output handling.

## Overview

The claude-code-go library now includes a comprehensive buffer management system that provides:

- **Configurable buffer sizes** with automatic truncation
- **Timeout protection** for buffer operations
- **Memory usage monitoring** and metrics
- **Recovery mechanisms** for buffer failures
- **Health checking** and diagnostics

## Key Features

### 1. Limited Buffers

The `LimitedBuffer` provides automatic size limiting with configurable truncation:

```go
import "github.com/marvai-dev/claude-code-go/pkg/claude/buffer"

// Create a buffer with 1MB limit
buf := buffer.NewLimitedBuffer(1024*1024, "[TRUNCATED]")

// Write data - automatically truncates when limit reached
buf.Write([]byte("large data..."))

// Check if truncation occurred
if buf.Truncated() {
    fmt.Println("Buffer was truncated due to size limit")
}
```

### 2. Buffer Configuration

Configure buffer behavior through `RunOptions`:

```go
bufferConfig := &buffer.Config{
    MaxStdoutSize:     10 * 1024 * 1024, // 10MB
    MaxStderrSize:     1 * 1024 * 1024,  // 1MB
    BufferTimeout:     30 * time.Second,
    EnableTruncation:  true,
    TruncationSuffix:  "\n[... output truncated ...]",
}

opts := &claude.RunOptions{
    Format:       claude.JSONOutput,
    BufferConfig: bufferConfig,
}
```

### 3. Default Configuration

If no buffer configuration is provided, sensible defaults are used:

- **MaxStdoutSize**: 10MB
- **MaxStderrSize**: 1MB  
- **BufferTimeout**: 30 seconds
- **EnableTruncation**: true
- **TruncationSuffix**: "[... output truncated due to size limit ...]"

## Usage Examples

### Basic Usage with Custom Buffer Limits

```go
client := &claude.ClaudeClient{BinPath: "claude"}

// Configure smaller buffers for memory-constrained environments
bufferConfig := &buffer.Config{
    MaxStdoutSize:    100 * 1024, // 100KB
    MaxStderrSize:    10 * 1024,  // 10KB
    BufferTimeout:    15 * time.Second,
}

opts := &claude.RunOptions{
    Format:       claude.JSONOutput,
    BufferConfig: bufferConfig,
}

result, err := client.RunPromptCtx(ctx, "Your prompt", opts)
```

### Streaming with Buffer Management

```go
// Streaming with enhanced buffer management
bufferConfig := &buffer.Config{
    MaxStdoutSize:     2 * 1024 * 1024, // 2MB for streaming
    MaxStderrSize:     100 * 1024,      // 100KB for errors
    BufferTimeout:     45 * time.Second,
    EnableTruncation:  false, // Don't truncate streaming responses
}

opts := &claude.RunOptions{
    Format:       claude.StreamJSONOutput,
    BufferConfig: bufferConfig,
}

messageCh, errCh := client.StreamPrompt(ctx, "Your prompt", opts)
```

### Dangerous Operations with Larger Buffers

```go
// Dangerous operations automatically use larger default buffers
dangerous := dangerous.NewDangerousClient(&claude.ClaudeClient{BinPath: "claude"})

opts := &claude.RunOptions{
    // BufferConfig is automatically configured with larger limits:
    // MaxStdoutSize: 50MB, MaxStderrSize: 5MB
}

result, err := dangerous.RunWithDangerousFlags(ctx, "prompt", opts)
```

## Advanced Features

### 1. Monitored Buffers

Track buffer usage with built-in metrics:

```go
monitoredBuf := buffer.NewMonitoredBuffer(1024*1024, "[TRUNCATED]")

// Write data
monitoredBuf.Write(data)

// Get metrics
stats := monitoredBuf.GetMetrics()
fmt.Printf("Total bytes written: %d\n", stats.TotalBytesWritten)
fmt.Printf("Truncations: %d\n", stats.TotalTruncations)
fmt.Printf("Average write size: %.2f\n", stats.AverageWriteSize)
```

### 2. Resilient Buffers

Automatic recovery from buffer failures:

```go
recoveryConfig := &buffer.RecoveryConfig{
    MaxRetries:                3,
    RetryDelay:                time.Second,
    FallbackBufferSize:        1024 * 1024, // 1MB fallback
    EnableGracefulDegradation: true,
}

resilientBuf := buffer.NewResilientBuffer(10*1024*1024, "[TRUNCATED]", recoveryConfig)

// Automatically recovers from failures and falls back to smaller buffer
resilientBuf.Write(data)

if resilientBuf.IsUsingFallback() {
    fmt.Println("Using fallback buffer due to primary buffer failure")
}
```

### 3. Health Monitoring

Monitor buffer system health:

```go
enhanced := buffer.NewEnhancedBufferManager(bufferConfig)

// Use monitored buffers
stdout := enhanced.NewMonitoredStdoutBuffer()
stderr := enhanced.NewMonitoredStderrBuffer()

// Check health
healthChecker := buffer.NewHealthChecker()
stats := enhanced.GetGlobalMetrics()
health := healthChecker.CheckHealth(stats)

if !health.IsHealthy {
    fmt.Printf("Buffer system issues detected: %v\n", health.Issues)
    fmt.Printf("Suggestions: %v\n", health.Suggestions)
}
```

## Configuration Best Practices

### Memory-Constrained Environments

```go
bufferConfig := &buffer.Config{
    MaxStdoutSize:    500 * 1024,  // 500KB
    MaxStderrSize:    50 * 1024,   // 50KB
    BufferTimeout:    10 * time.Second,
    EnableTruncation: true,
}
```

### High-Volume Processing

```go
bufferConfig := &buffer.Config{
    MaxStdoutSize:    50 * 1024 * 1024, // 50MB
    MaxStderrSize:    5 * 1024 * 1024,  // 5MB
    BufferTimeout:    120 * time.Second,
    EnableTruncation: true,
}
```

### Development/Testing

```go
bufferConfig := &buffer.Config{
    MaxStdoutSize:    100 * 1024 * 1024, // 100MB - generous for testing
    MaxStderrSize:    10 * 1024 * 1024,  // 10MB
    BufferTimeout:    300 * time.Second,  // 5 minutes
    EnableTruncation: false, // Keep all output for debugging
}
```

## Migration Guide

### From Unbounded Buffers

Before (vulnerable to memory issues):
```go
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
```

After (protected with limits):
```go
bufManager := buffer.NewBufferManager(buffer.DefaultConfig())
stdout := bufManager.NewStdoutBuffer()
stderr := bufManager.NewStderrBuffer()
cmd.Stdout = stdout
cmd.Stderr = stderr
```

### Adding Buffer Configuration to Existing Code

```go
// Old code
opts := &claude.RunOptions{
    Format: claude.JSONOutput,
}

// New code with buffer management
opts := &claude.RunOptions{
    Format: claude.JSONOutput,
    BufferConfig: &buffer.Config{
        MaxStdoutSize: 5 * 1024 * 1024, // 5MB
        MaxStderrSize: 512 * 1024,      // 512KB
        BufferTimeout: 30 * time.Second,
    },
}
```

## Troubleshooting

### Common Issues

1. **Buffer Truncation Warnings**
   - Increase `MaxStdoutSize` or `MaxStderrSize`
   - Consider using streaming for large outputs
   - Use `EnableTruncation: false` for development

2. **Timeout Errors**
   - Increase `BufferTimeout`
   - Check for hanging operations
   - Use context cancellation

3. **Memory Usage**
   - Monitor with `MonitoredBuffer`
   - Use health checking
   - Adjust buffer sizes based on actual usage

### Debugging

Enable detailed logging and monitoring:

```go
// Use monitored buffers for detailed metrics
enhanced := buffer.NewEnhancedBufferManager(config)
stdout := enhanced.NewMonitoredStdoutBuffer()

// After operations, check metrics
stats := stdout.GetMetrics()
fmt.Printf("Buffer metrics: %+v\n", stats)

// Check system health
health := enhanced.healthCheck.CheckHealth(stats)
if !health.IsHealthy {
    fmt.Printf("Health issues: %v\n", health.Issues)
}
```

## Performance Considerations

- **Buffer Size**: Larger buffers use more memory but reduce truncation
- **Timeout Values**: Shorter timeouts prevent hanging but may interrupt legitimate operations
- **Monitoring Overhead**: Monitored buffers add slight performance cost but provide valuable insights
- **Fallback Buffers**: Resilient buffers use additional memory for fallback storage

## Compatibility

The buffer management system is fully backward compatible:

- If no `BufferConfig` is specified, default configuration is used
- Existing code continues to work without modification
- New features are opt-in through configuration

## API Reference

See the [Go documentation](https://pkg.go.dev/github.com/marvai-dev/claude-code-go/pkg/claude/buffer) for complete API reference.