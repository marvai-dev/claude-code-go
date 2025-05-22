# Claude Code Go SDK - Project Architecture

## Project Purpose

This project is a **Go SDK library** for programmatically integrating the Claude Code CLI into Go applications. It enables developers to use Claude Code as a subprocess from their Go programs.

## Architecture Decisions

### Core Focus: SDK Library Only

**Purpose**: Provide a Go library that wraps the existing `claude` CLI command
**Target Users**: Go developers who want to integrate Claude Code into their applications
**Import Path**: `github.com/lancekrogers/claude-code-go/pkg/claude`

### What This Project IS

- A Go SDK library (`pkg/claude/`)
- A subprocess wrapper around the official Claude Code CLI
- A tool for Go applications to programmatically use Claude Code
- JSON/streaming response parser and error handler
- Convenience methods for common Claude Code operations

### What This Project IS NOT

- A CLI distribution or alternative to the official `claude` command
- A replacement for the Claude Code CLI
- A standalone application for end users
- A command-line tool that users should invoke directly

## Key Architectural Principles

### 1. No Custom CLI (`cmd/` directory removed)

**Decision**: This project does not provide its own CLI interface.

**Rationale**:
- Users who want CLI access should use the official `claude` command
- Creating a CLI wrapper would cause user confusion
- Adds unnecessary maintenance burden
- Creates circular dependency: CLI → SDK → CLI

**Historical Context**: An initial `cmd/claudecli/` implementation was removed because it went against the core purpose of providing an SDK library.

### 2. SDK Library Focus (`pkg/claude/`)

**Decision**: The entire project focuses on the Go library in `pkg/claude/`.

**Core Functionality**:
- `claude.NewClient()` - Create client instances
- `client.RunPrompt()` - Execute single prompts
- `client.StreamPrompt()` - Handle streaming responses
- Convenience methods: `RunWithSystemPrompt()`, `ContinueConversation()`, etc.
- MCP tool validation and management
- Argument building and subprocess management

### 3. Subprocess Wrapper Pattern

**Decision**: The SDK executes the official `claude` CLI as a subprocess.

**Implementation**:
- Uses `exec.Command()` to spawn `claude` processes
- Builds command arguments programmatically
- Parses JSON/text responses
- Handles streaming JSON-LD output
- Manages error conditions and exit codes

## Usage Pattern

Go applications should use this library like:

```go
import "github.com/lancekrogers/claude-code-go/pkg/claude"

// Create client
client := claude.NewClient("claude")

// Basic usage
result, err := client.RunPrompt("Generate a function", &claude.RunOptions{
    Format: claude.JSONOutput,
})

// Streaming usage
ctx := context.Background()
messageCh, errCh := client.StreamPrompt(ctx, "Build an app", &claude.RunOptions{})

// Convenience methods
result, err := client.RunWithSystemPrompt(
    "You are a senior engineer",
    "Build a REST API",
    &claude.RunOptions{},
)
```

## Testing Strategy

### Unit Tests
- Test argument building logic
- Test response parsing
- Test error handling
- Mock subprocess execution

### Integration Tests
- Test against mock Claude server
- Validate full workflow: args → subprocess → parsing
- Test streaming responses
- Test all output formats (text, JSON, stream-json)

### Mock Server Approach
- HTTP server that simulates Claude Code responses
- Shell script mock binary that forwards to HTTP server
- Enables testing without real Claude Code CLI dependency

## Development Guidelines

### Do Not Add CLI Components
- Never create `cmd/` directories
- Do not build standalone executables for end users
- Focus exclusively on the library interface

### Examples and Documentation
- Examples should show Go library usage
- Documentation should focus on `import` and API usage
- Avoid CLI-style examples or documentation

### Dependencies
- Minimize external dependencies
- Only add deps that enhance SDK functionality
- No CLI frameworks or command-line parsing libraries

## Integration Tests

Run integration tests with:
```bash
task test-integration     # With mock server
task test-integration-real # With real Claude CLI (if available)
```

The mock server enables testing the full SDK workflow without requiring Claude Code CLI installation.

## Related Documentation

- [SDK Documentation](ai_docs/claude_docs/sdk.md) - Official Claude Code SDK patterns
- [README.md](README.md) - Project overview and usage examples
- [TESTING.md](TESTING.md) - Comprehensive testing guide

---

**Remember**: This is an SDK library, not a CLI. Users import the Go package; they don't run a binary.