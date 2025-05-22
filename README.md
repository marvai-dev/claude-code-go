# Claude Code Go SDK

A comprehensive Go wrapper for the [Claude Code Command Line Interface](https://docs.anthropic.com/en/docs/claude-code), enabling seamless integration of Claude's AI coding capabilities into Go applications and AI agent frameworks.

This SDK provides a Go-native interface to all Claude Code CLI features, allowing you to programmatically integrate Claude's coding assistance into your Go applications.

## Features

- **Full Claude Code CLI Support**: Access all Claude Code features from your Go applications
- **Streaming Support**: Real-time streaming of Claude's responses
- **MCP Integration**: Model Context Protocol support for extending Claude with additional tools
- **Stdin Handling**: Process files and other input sources through Claude
- **Session Management**: Support for multi-turn conversations
- **Flexible Output Formats**: Text, JSON, and streaming JSON outputs
- **Go-Native CLI**: Optional Go implementation of the Claude Code CLI

## Installation

```bash
go get github.com/lancekrogers/claude-code-go
```

## Prerequisites

- **Claude Code CLI**: Must be installed and accessible in your PATH (or specified via the `BinPath` option)
  - Install from: <https://docs.anthropic.com/en/docs/claude-code/getting-started>
- **MCP Servers**: For MCP functionality, the necessary MCP servers must be available
  - See: <https://docs.anthropic.com/en/docs/claude-code/cli-usage#mcp-configuration>

## Quick Start

```go
package main

import (
 "fmt"
 "log"

 "github.com/lancekrogers/claude-code-go/pkg/claude"
)

func main() {
 // Create a new Claude client
 client := claude.NewClient("claude")

 // Run a simple prompt
 result, err := client.RunPrompt("Write a function to calculate Fibonacci numbers", nil)
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 fmt.Println(result.Result)
}
```

## Usage Examples

### Basic JSON Output

```go
client := claude.NewClient("claude")
result, err := client.RunPrompt("Generate a hello world function", &claude.RunOptions{
 Format: claude.JSONOutput,
})
if err != nil {
 log.Fatalf("Error: %v", err)
}

fmt.Printf("Cost: $%.6f\n", result.CostUSD)
fmt.Printf("Session ID: %s\n", result.SessionID)
fmt.Println(result.Result)
```

### Custom System Prompt

```go
client := claude.NewClient("claude")
result, err := client.RunPrompt("Create a database schema", &claude.RunOptions{
 SystemPrompt: "You are a database architect. Use PostgreSQL best practices and include proper indexing.",
})
if err != nil {
 log.Fatalf("Error: %v", err)
}

fmt.Println(result.Result)
```

### Processing Files

```go
client := claude.NewClient("claude")
file, err := os.Open("mycode.go")
if err != nil {
 log.Fatalf("Cannot open file: %v", err)
}
defer file.Close()

result, err := client.RunFromStdin(file, "Review this code for bugs", nil)
if err != nil {
 log.Fatalf("Error: %v", err)
}

fmt.Println(result.Result)
```

### Streaming Responses

```go
client := claude.NewClient("claude")
ctx := context.Background()

messageCh, errCh := client.StreamPrompt(ctx, "Build a React component", &claude.RunOptions{})

// Handle errors
go func() {
 for err := range errCh {
  log.Printf("Error: %v", err)
 }
}()

// Process messages
for msg := range messageCh {
 switch msg.Type {
 case "assistant":
  fmt.Println("Claude:", msg.Result)
 case "result":
  fmt.Printf("Done! Cost: $%.4f\n", msg.CostUSD)
 }
}
```

### Using MCP with Claude

```go
// Create MCP configuration file
mcpConfig := map[string]interface{}{
 "mcpServers": map[string]interface{}{
  "filesystem": map[string]interface{}{
   "command": "npx",
   "args": []string{"-y", "@modelcontextprotocol/server-filesystem", "./"},
  },
 },
}

// Write to a temporary file
mcpFile, _ := os.CreateTemp("", "mcp-*.json")
defer os.Remove(mcpFile.Name())
json.NewEncoder(mcpFile).Encode(mcpConfig)
mcpFile.Close()

// Run Claude with MCP configuration
client := claude.NewClient("claude")
result, err := client.RunPrompt(
 "List all files in the current directory",
 &claude.RunOptions{
  MCPConfigPath: mcpFile.Name(),
  AllowedTools:  []string{"mcp__filesystem__list_directory"},
 },
)
```

### Multi-turn Conversations

```go
client := claude.NewClient("claude")

// First turn
result, err := client.RunPrompt("Write a function to calculate fibonacci numbers", &claude.RunOptions{
 Format: claude.JSONOutput,
})
if err != nil {
 log.Fatalf("Error: %v", err)
}

sessionID := result.SessionID
fmt.Println("First response:", result.Result)

// Continue the conversation using convenience method
followup, err := client.ResumeConversation("Now optimize it for better performance", sessionID)
if err != nil {
 log.Fatalf("Error: %v", err)
}

fmt.Println("Follow-up response:", followup.Result)
```

### Convenience Methods

The SDK provides several convenience methods for common use cases:

```go
client := claude.NewClient("claude")

// Quick MCP integration
result, err := client.RunWithMCP(
 "List files in the project",
 "mcp-config.json",
 []string{"mcp__filesystem__list_directory"},
)

// Custom system prompt
result, err = client.RunWithSystemPrompt(
 "Create a REST API",
 "You are a senior backend engineer. Focus on security and performance.",
)

// Continue most recent conversation
result, err = client.ContinueConversation("Add error handling to the code")

// Resume specific conversation
result, err = client.ResumeConversation("Add tests", "session-id-123")
```

## API Reference

### Core Types

```go
// ClaudeClient is the main client for interacting with Claude Code
type ClaudeClient struct {
 // BinPath is the path to the Claude Code binary
 BinPath string
 // DefaultOptions are the default options to use for all requests
 DefaultOptions *RunOptions
}

// RunOptions configures how Claude Code is executed
type RunOptions struct {
 Format          OutputFormat
 SystemPrompt    string
 AppendPrompt    string
 MCPConfigPath   string
 AllowedTools    []string
 DisallowedTools []string
 PermissionTool  string
 ResumeID        string
 Continue        bool
 MaxTurns        int
 Verbose         bool
}

// Supported output formats
const (
 TextOutput       OutputFormat = "text"
 JSONOutput       OutputFormat = "json"
 StreamJSONOutput OutputFormat = "stream-json"
)
```

### Core Methods

```go
// NewClient creates a new Claude client with the specified binary path
func NewClient(binPath string) *ClaudeClient

// RunPrompt executes a prompt with Claude Code and returns the result
func (c *ClaudeClient) RunPrompt(prompt string, opts *RunOptions) (*ClaudeResult, error)

// StreamPrompt executes a prompt with Claude Code and streams the results through a channel
func (c *ClaudeClient) StreamPrompt(ctx context.Context, prompt string, opts *RunOptions) (<-chan Message, <-chan error)

// RunFromStdin runs Claude Code with input from stdin
func (c *ClaudeClient) RunFromStdin(stdin io.Reader, prompt string, opts *RunOptions) (*ClaudeResult, error)
```

### Convenience Methods

```go
// RunWithMCP runs Claude with MCP configuration
func (c *ClaudeClient) RunWithMCP(prompt string, mcpConfigPath string, allowedTools []string) (*ClaudeResult, error)

// RunWithSystemPrompt runs Claude with a custom system prompt
func (c *ClaudeClient) RunWithSystemPrompt(prompt string, systemPrompt string) (*ClaudeResult, error)

// ContinueConversation continues the most recent conversation
func (c *ClaudeClient) ContinueConversation(prompt string) (*ClaudeResult, error)

// ResumeConversation resumes a specific conversation by session ID
func (c *ClaudeClient) ResumeConversation(prompt string, sessionID string) (*ClaudeResult, error)
```

## Command-Line Interface

The package includes a Go-native CLI wrapper for Claude Code in the `cmd/claudecli` directory. Build it with:

```bash
go build -o claude-go ./cmd/claudecli
```

Usage:

```bash
# Simple prompt
./claude-go -p "Write a function to calculate Fibonacci numbers"

# JSON output
./claude-go -p "Generate a hello world function" --output-format json

# Process a file
cat mycode.go | ./claude-go -p "Review this code"

# MCP configuration
./claude-go -p "List files" --mcp-config config.json --allowedTools mcp__filesystem__list_directory

# Continue a conversation
./claude-go -p "How would you improve this?" --continue
```

## Integration with Agent Frameworks

This SDK is designed to be easily integrated with AI agent frameworks. Here's a simple example:

```go
package main

import (
 "context"
 "fmt"
 "log"

 "github.com/lancekrogers/claude-code-go/pkg/claude"
 "github.com/youragentframework/agent"
)

// ClaudeAgent implements the Agent interface for your framework
type ClaudeAgent struct {
 client *claude.ClaudeClient
 ctx    context.Context
}

// NewClaudeAgent creates a new Claude agent
func NewClaudeAgent(ctx context.Context, claudePath string) *ClaudeAgent {
 return &ClaudeAgent{
  client: claude.NewClient(claudePath),
  ctx:    ctx,
 }
}

// Execute runs a prompt through Claude Code
func (a *ClaudeAgent) Execute(prompt string, tools []string) (string, error) {
 result, err := a.client.RunPrompt(prompt, &claude.RunOptions{
  AllowedTools: tools,
  MaxTurns:     10,
 })
 if err != nil {
  return "", err
 }
 return result.Result, nil
}

// Main framework integration
func main() {
 ctx := context.Background()

 // Create Claude agent
 claudeAgent := NewClaudeAgent(ctx, "claude")

 // Register with your agent framework
 agent.Register("claude-code", claudeAgent)

 // Use the agent
 result, err := agent.Execute("claude-code", "Build a REST API", []string{"Bash"})
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 fmt.Println(result)
}
```

## Official Documentation

This Go SDK wraps the official Claude Code CLI. For comprehensive documentation on Claude Code features, configuration, and usage patterns, refer to the official documentation:

- **[Claude Code Overview](https://docs.anthropic.com/en/docs/claude-code/overview)** - Introduction and key concepts
- **[CLI Usage](https://docs.anthropic.com/en/docs/claude-code/cli-usage)** - Complete CLI reference
- **[SDK Documentation](https://docs.anthropic.com/en/docs/claude-code/sdk)** - Official SDK usage guide
- **[Getting Started](https://docs.anthropic.com/en/docs/claude-code/getting-started)** - Installation and setup
- **[Tutorials](https://docs.anthropic.com/en/docs/claude-code/tutorials)** - Step-by-step guides

## Development Tasks

We use [Task](https://taskfile.dev/#/installation) to automate common development workflows. Install Task and run these commands in the project root:

| Task         | Description                          |
| ------------ | ------------------------------------ |
| `task`       | Build the project and run all tests  |
| `task build` | Build the CLI binary and Go packages |
| `task test`  | Run the full test suite              |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Anthropic for creating Claude Code
- The Go community for the excellent testing and tooling support
