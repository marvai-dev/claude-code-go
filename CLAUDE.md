# Claude Code SDK Alignment Guide

This document ensures our Go SDK stays aligned with the official Python SDK structure and concepts. All future changes should maintain compatibility with the official SDK design patterns.

## Official Python SDK Reference

- **Documentation**: https://docs.anthropic.com/en/docs/claude-code/sdk
- **Core Concept**: Async `query()` method with streaming support
- **Configuration**: `ClaudeCodeOptions` class for all configuration

## Core API Alignment

### Python SDK Core Method
```python
async for message in query(
    prompt="Write a hello world program",
    options=ClaudeCodeOptions(
        max_turns=3,
        system_prompt="You're a helpful coding assistant",
        cwd=Path("/project/path"),
        allowed_tools=["Read", "Write", "Bash"],
        permission_mode="acceptEdits"
    )
):
    print(message.content)
```

### Go SDK Equivalent (Current Target)
```go
// Primary method should be Query() with streaming
messageCh, err := client.Query(ctx, "Write a hello world program", 
    QueryOptions{
        MaxTurns:       3,
        SystemPrompt:   "You're a helpful coding assistant",
        WorkingDir:     "/project/path",
        AllowedTools:   []string{"Read", "Write", "Bash"},
        PermissionMode: "acceptEdits",
    })

for message := range messageCh {
    fmt.Println(message.Content)
}
```

## Configuration Structure

### Python SDK Options
```python
ClaudeCodeOptions(
    max_turns=3,           # Conversation length limit
    system_prompt="...",   # Custom system prompt
    cwd=Path("..."),       # Working directory
    allowed_tools=[...],   # Tool permissions
    permission_mode="..."  # Permission handling mode
)
```

### Go SDK Options (Target Structure)
```go
type QueryOptions struct {
    // Core conversation control
    MaxTurns       int           `json:"max_turns,omitempty"`
    SystemPrompt   string        `json:"system_prompt,omitempty"`
    WorkingDir     string        `json:"cwd,omitempty"`
    
    // Tool and permission management
    AllowedTools   []string      `json:"allowed_tools,omitempty"`
    PermissionMode string        `json:"permission_mode,omitempty"`
    
    // Go-specific extensions (should be minimal)
    Context        context.Context
    BufferConfig   *BufferConfig  // Our buffer management extension
}
```

## Key Design Principles

### 1. Async-First Pattern
- **Python**: Uses `async for` iteration
- **Go**: Use channels for streaming, context for cancellation
- **Alignment**: Both provide streaming message iteration

### 2. Configuration Consolidation
- **Python**: Single `ClaudeCodeOptions` class
- **Go**: Single `QueryOptions` struct
- **Alignment**: All configuration in one place, not scattered across method parameters

### 3. Message Structure
- **Python**: Messages have `.content`, `.type`, etc.
- **Go**: Should mirror the same structure exactly
- **Alignment**: Field names and types should match

### 4. Tool Permission Model
- **Python**: `allowed_tools=["Read", "Write", "Bash"]`
- **Go**: `AllowedTools: []string{"Read", "Write", "Bash"}`
- **Alignment**: Same tool names, same permission concepts

## Required Refactoring

### Current Issues to Address

1. **Method Names**: We use `RunPrompt`, should be `Query`
2. **Configuration Spread**: We have scattered options, should consolidate
3. **Async Pattern**: We support streaming but not as primary interface
4. **Option Names**: Our field names don't match Python SDK

### Migration Plan

#### Phase 1: Add Aligned API (Maintain Backward Compatibility)
```go
// New aligned API
func (c *Client) Query(ctx context.Context, prompt string, opts QueryOptions) (<-chan Message, error)

// Keep existing API for backward compatibility
func (c *Client) RunPromptCtx(ctx context.Context, prompt string, opts *RunOptions) (*ClaudeResult, error)
```

#### Phase 2: Unified Configuration
```go
// Align configuration structure
type QueryOptions struct {
    MaxTurns       int      `json:"max_turns,omitempty"`
    SystemPrompt   string   `json:"system_prompt,omitempty"`
    WorkingDir     string   `json:"cwd,omitempty"`
    AllowedTools   []string `json:"allowed_tools,omitempty"`
    PermissionMode string   `json:"permission_mode,omitempty"`
    
    // Buffer management (Go-specific extension)
    BufferConfig   *BufferConfig `json:"-"`
}
```

#### Phase 3: Message Structure Alignment
```go
// Align message structure with Python SDK
type Message struct {
    Content   string                 `json:"content"`
    Type      string                 `json:"type"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    // ... other fields matching Python SDK
}
```

## Buffer Management Integration

### Principle: Extend, Don't Replace
Our buffer management should be a **Go-specific extension** that doesn't break SDK alignment:

```go
type QueryOptions struct {
    // Standard SDK fields (match Python exactly)
    MaxTurns       int      `json:"max_turns,omitempty"`
    SystemPrompt   string   `json:"system_prompt,omitempty"`
    // ... other standard fields
    
    // Go-specific extensions (not serialized to match Python)
    BufferConfig   *BufferConfig `json:"-"`
    Context        context.Context `json:"-"`
}
```

### Implementation Strategy
1. Keep buffer management as internal implementation detail
2. Expose configuration through Go-idiomatic options
3. Don't let buffer concepts leak into the main API surface
4. Maintain robustness without breaking SDK alignment

## Compatibility Checklist

When making changes, ensure:

- [ ] Method names match Python SDK (`query` → `Query`)
- [ ] Configuration options match Python field names exactly
- [ ] Message structures align with Python SDK
- [ ] Tool names and permissions work identically
- [ ] Streaming patterns provide equivalent functionality
- [ ] Error handling follows similar patterns
- [ ] Documentation examples can be easily translated between SDKs

## Testing Alignment

### Cross-SDK Test Cases
Maintain test cases that can be directly translated:

```python
# Python test
messages = []
async for message in query(
    "Write hello world", 
    ClaudeCodeOptions(max_turns=1)
):
    messages.append(message)
assert len(messages) > 0
```

```go
// Go equivalent test
var messages []Message
messageCh, err := client.Query(ctx, "Write hello world", 
    QueryOptions{MaxTurns: 1})
require.NoError(t, err)

for message := range messageCh {
    messages = append(messages, message)
}
assert.True(t, len(messages) > 0)
```

## Deprecation Strategy

### Backward Compatibility Plan
1. **Keep existing APIs** with deprecation warnings
2. **Add new aligned APIs** as primary interface
3. **Gradually migrate examples** to new API
4. **Update documentation** to show new patterns first
5. **Eventually remove old APIs** in major version bump

### Example Migration
```go
// OLD (deprecated but maintained)
func (c *Client) RunPromptCtx(ctx context.Context, prompt string, opts *RunOptions) (*ClaudeResult, error) {
    // Internally convert to new API
    return c.runLegacy(ctx, prompt, opts)
}

// NEW (aligned with Python SDK)
func (c *Client) Query(ctx context.Context, prompt string, opts QueryOptions) (<-chan Message, error) {
    // Primary implementation
}
```

## Future Changes Protocol

### Before Making Changes
1. **Check Python SDK** for equivalent functionality
2. **Ensure field names match** Python SDK exactly
3. **Test cross-compatibility** with Python examples
4. **Update this document** with any new alignments

### When Python SDK Changes
1. **Update Go SDK** to match new patterns
2. **Maintain backward compatibility** for existing Go users
3. **Update examples** to show new patterns
4. **Test migration path** works smoothly

## Extension Guidelines

### Acceptable Go-Specific Extensions
- Context support for cancellation
- Channel-based streaming (Go idiom for async iteration)
- Buffer management for memory safety
- Go-specific error types
- Interface implementations for Go patterns

### Unacceptable Deviations
- Different configuration field names
- Different tool permission models
- Different message structures
- Different core method names
- Breaking the async/streaming paradigm

This alignment ensures our Go SDK can evolve alongside the official Python SDK while providing Go-idiomatic patterns and maintaining our robustness improvements.