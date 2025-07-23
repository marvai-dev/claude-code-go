package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/marvai-dev/claude-code-go/pkg/claude"
	"github.com/marvai-dev/claude-code-go/pkg/claude/buffer"
)

func main() {
	fmt.Println("=== Claude Code Go - Python SDK Aligned API ===\n")

	// Create client (same as before)
	client := &claude.ClaudeClient{
		BinPath: "claude",
	}

	// Example 1: Basic Query (matches Python SDK pattern)
	fmt.Println("1. Basic Query (Python SDK aligned):")
	basicQuery(client)

	// Example 2: Streaming Query with Options
	fmt.Println("\n2. Streaming Query with Full Options:")
	streamingQuery(client)

	// Example 3: Synchronous Query
	fmt.Println("\n3. Synchronous Query:")
	syncQuery(client)

	// Example 4: Python SDK Pattern Comparison
	fmt.Println("\n4. Python SDK Pattern Comparison:")
	pythonSDKComparison(client)
}

func basicQuery(client *claude.ClaudeClient) {
	ctx := context.Background()
	
	// This matches the Python SDK pattern:
	// async for message in query("Write hello world", ClaudeCodeOptions(max_turns=1)):
	messageCh, err := client.Query(ctx, "Write a simple hello world program", claude.QueryOptions{
		MaxTurns: 1,
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Iterate over messages (like Python async for)
	for message := range messageCh {
		if message.Content != "" {
			fmt.Printf("Message: %s\n", message.Content)
		}
		if message.Type == "result" {
			fmt.Printf("Result: %s\n", message.Content)
			break
		}
	}
}

func streamingQuery(client *claude.ClaudeClient) {
	ctx := context.Background()
	
	// Full QueryOptions equivalent to Python ClaudeCodeOptions
	messageCh, err := client.Query(ctx, "Create a simple Python calculator", claude.QueryOptions{
		MaxTurns:       3,
		SystemPrompt:   "You are a helpful coding assistant",
		AllowedTools:   []string{"Write", "Read", "Bash"},
		PermissionMode: "acceptEdits",
		Model:          "claude-3-5-sonnet-20241022",
		
		// Go-specific extensions (buffer management)
		BufferConfig: &buffer.Config{
			MaxStdoutSize:    5 * 1024 * 1024, // 5MB
			MaxStderrSize:    1 * 1024 * 1024, // 1MB
			BufferTimeout:    60 * time.Second,
		},
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	messageCount := 0
	for message := range messageCh {
		messageCount++
		fmt.Printf("Message %d [%s]: ", messageCount, message.Type)
		
		if message.Content != "" {
			// Truncate long content for display
			content := message.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Printf("%s\n", content)
		} else if len(message.Message) > 0 {
			fmt.Printf("Raw: %s\n", string(message.Message)[:min(100, len(message.Message))])
		}
		
		if message.Type == "result" {
			fmt.Printf("Final result received. Total messages: %d\n", messageCount)
			break
		}
	}
}

func syncQuery(client *claude.ClaudeClient) {
	ctx := context.Background()
	
	// Synchronous version for simple use cases
	result, err := client.QuerySync(ctx, "What is the capital of France?", claude.QueryOptions{
		MaxTurns:     1,
		SystemPrompt: "Give brief, direct answers",
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Answer: %s\n", result.Result)
	fmt.Printf("Cost: $%.4f\n", result.CostUSD)
	fmt.Printf("Duration: %dms\n", result.DurationMS)
}

func pythonSDKComparison(client *claude.ClaudeClient) {
	fmt.Println("Python SDK equivalent code:")
	fmt.Println(`
# Python SDK
async for message in query(
    "Write a hello world program",
    ClaudeCodeOptions(
        max_turns=3,
        system_prompt="You're a helpful coding assistant",
        allowed_tools=["Read", "Write", "Bash"],
        permission_mode="acceptEdits"
    )
):
    print(message.content)
`)

	fmt.Println("Go SDK equivalent:")
	fmt.Println(`
// Go SDK
messageCh, err := client.Query(ctx, "Write a hello world program", 
    claude.QueryOptions{
        MaxTurns:       3,
        SystemPrompt:   "You're a helpful coding assistant",
        AllowedTools:   []string{"Read", "Write", "Bash"},
        PermissionMode: "acceptEdits",
    })

for message := range messageCh {
    fmt.Println(message.Content)
}
`)

	fmt.Println("Running the Go equivalent:")
	
	ctx := context.Background()
	messageCh, err := client.Query(ctx, "Write a simple hello world program", claude.QueryOptions{
		MaxTurns:       3,
		SystemPrompt:   "You're a helpful coding assistant",
		AllowedTools:   []string{"Read", "Write", "Bash"},
		PermissionMode: "acceptEdits",
	})
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	for message := range messageCh {
		if message.Content != "" {
			// Show first part of content
			content := message.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			fmt.Printf("Content: %s\n", content)
		}
		if message.Type == "result" {
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}