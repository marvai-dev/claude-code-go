# Claude Code Go SDK

A Go library for programmatically integrating the [Claude Code Command Line Interface](https://docs.anthropic.com/en/docs/claude-code) into Go applications. This SDK provides a Go-native interface to all Claude Code CLI features, enabling you to build AI-powered applications that leverage Claude's coding capabilities.

## Features

- **Full Claude Code CLI Wrapper**: Access all Claude Code features from your Go applications
- **Streaming Support**: Real-time streaming of Claude's responses with context cancellation
- **MCP Integration**: Model Context Protocol support for extending Claude with additional tools
- **Stdin Processing**: Process files and other input sources through Claude
- **Session Management**: Support for multi-turn conversations with automatic session handling
- **Multiple Output Formats**: Text, JSON, and streaming JSON outputs
- **Convenience Methods**: Simplified APIs for common use cases
- **Comprehensive Testing**: Unit and integration tests with mock server support

## Installation

```bash
go get github.com/lancekrogers/claude-code-go
```

## Prerequisites

- **Claude Max Subscription**: Claude Code requires a Claude Max subscription
  - **[Sign up for Claude Max](https://claude.ai/referral/UKHPp7nGJw)** to access Claude Code CLI
  - Claude Max provides unlimited usage of Claude Code with advanced features
- **Claude Code CLI**: Must be installed and accessible in your PATH
  - Install from: <https://docs.anthropic.com/en/docs/claude-code/getting-started>
  - The CLI handles authentication automatically when needed
- **MCP Servers** (optional): For MCP functionality, install the necessary MCP servers
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
result, err := client.RunPrompt("Create a database schema", &claude.RunOptions{
	SystemPrompt: "You are a database architect. Use PostgreSQL best practices.",
})
```

### Processing Files

```go
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

### MCP Integration

```go
// Create MCP configuration
mcpConfig := map[string]interface{}{
	"mcpServers": map[string]interface{}{
		"filesystem": map[string]interface{}{
			"command": "npx",
			"args": []string{"-y", "@modelcontextprotocol/server-filesystem", "./"},
		},
	},
}

// Write to temporary file
mcpFile, _ := os.CreateTemp("", "mcp-*.json")
defer os.Remove(mcpFile.Name())
json.NewEncoder(mcpFile).Encode(mcpConfig)
mcpFile.Close()

// Run with MCP tools
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
// First turn
result, err := client.RunPrompt("Write a fibonacci function", &claude.RunOptions{
	Format: claude.JSONOutput,
})

sessionID := result.SessionID

// Continue the conversation
followup, err := client.ResumeConversation("Now optimize it for performance", sessionID)
```

### Convenience Methods

```go
// Quick MCP integration
result, err := client.RunWithMCP(
	"List files in the project",
	"mcp-config.json",
	[]string{"mcp__filesystem__list_directory"},
)

// Custom system prompt
result, err = client.RunWithSystemPrompt(
	"You are a senior backend engineer",
	"Create a REST API",
)

// Continue most recent conversation
result, err = client.ContinueConversation("Add error handling to the code")
```

## API Reference

### Core Types

```go
// ClaudeClient is the main client for interacting with Claude Code
type ClaudeClient struct {
	BinPath        string
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

// Output formats
const (
	TextOutput       OutputFormat = "text"
	JSONOutput       OutputFormat = "json"
	StreamJSONOutput OutputFormat = "stream-json"
)
```

### Core Methods

```go
// Create new client
func NewClient(binPath string) *ClaudeClient

// Execute prompts
func (c *ClaudeClient) RunPrompt(prompt string, opts *RunOptions) (*ClaudeResult, error)
func (c *ClaudeClient) StreamPrompt(ctx context.Context, prompt string, opts *RunOptions) (<-chan Message, <-chan error)
func (c *ClaudeClient) RunFromStdin(stdin io.Reader, prompt string, opts *RunOptions) (*ClaudeResult, error)
```

### Convenience Methods

```go
// MCP integration
func (c *ClaudeClient) RunWithMCP(prompt, mcpConfigPath string, allowedTools []string) (*ClaudeResult, error)

// System prompts
func (c *ClaudeClient) RunWithSystemPrompt(systemPrompt, prompt string) (*ClaudeResult, error)

// Conversation management
func (c *ClaudeClient) ContinueConversation(prompt string) (*ClaudeResult, error)
func (c *ClaudeClient) ResumeConversation(prompt, sessionID string) (*ClaudeResult, error)
```

## Integration with Agent Frameworks

This SDK is designed for easy integration with AI agent frameworks:

```go
type ClaudeAgent struct {
	client *claude.ClaudeClient
	ctx    context.Context
}

func NewClaudeAgent(ctx context.Context) *ClaudeAgent {
	return &ClaudeAgent{
		client: claude.NewClient("claude"),
		ctx:    ctx,
	}
}

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
```

## Testing

The SDK includes comprehensive testing with both unit tests and integration tests:

```bash
# Run unit tests
task test

# Run integration tests with mock server
task test-integration

# Run integration tests with real Claude CLI
task test-integration-real

# Run all tests
task test-local
```

## Official Documentation

This Go SDK wraps the official Claude Code CLI. For comprehensive documentation:

- **[Claude Code Overview](https://docs.anthropic.com/en/docs/claude-code/overview)** - Introduction and concepts
- **[CLI Usage](https://docs.anthropic.com/en/docs/claude-code/cli-usage)** - Complete CLI reference
- **[SDK Guide](https://docs.anthropic.com/en/docs/claude-code/sdk)** - Official SDK patterns
- **[Getting Started](https://docs.anthropic.com/en/docs/claude-code/getting-started)** - Installation

## Development

We use [Task](https://taskfile.dev) for development automation:

| Command                    | Description                          |
| -------------------------- | ------------------------------------ |
| `task`                     | Build and test the SDK               |
| `task build-lib`           | Build the core library               |
| `task test`                | Run unit tests                       |
| `task test-integration`    | Run integration tests (mock)         |
| `task test-integration-real` | Run integration tests (real Claude) |
| `task test-coverage`       | Generate coverage report             |
| `task lint`                | Run linting tools                    |

## Project Architecture

This project is designed as a **Go SDK library** for wrapping the Claude Code CLI. It does not provide its own CLI - users should import the library into their Go applications.

See [CLAUDE.md](CLAUDE.md) for detailed architectural decisions and development guidelines.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Anthropic for creating Claude Code
- The Go community for excellent tooling and testing support