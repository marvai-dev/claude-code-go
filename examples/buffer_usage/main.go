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
	fmt.Println("=== Claude Code Go - Enhanced Buffer Management Example ===\n")

	// Create a Claude client
	client := &claude.ClaudeClient{
		BinPath: "claude",
	}

	// Example 1: Default buffer configuration
	fmt.Println("1. Using default buffer configuration:")
	runWithDefaultBuffers(client)

	// Example 2: Custom buffer configuration
	fmt.Println("\n2. Using custom buffer configuration:")
	runWithCustomBuffers(client)

	// Example 3: Large output handling
	fmt.Println("\n3. Handling large outputs with limits:")
	runWithLargeOutputLimits(client)

	// Example 4: Streaming with buffer management
	fmt.Println("\n4. Streaming with enhanced buffer management:")
	runStreamingWithBuffers(client)
}

func runWithDefaultBuffers(client *claude.ClaudeClient) {
	opts := &claude.RunOptions{
		Format: claude.JSONOutput,
		// BufferConfig is nil, so defaults will be used
		// Default: 10MB stdout, 1MB stderr, 30s timeout
	}

	result, err := client.RunPromptCtx(context.Background(), "Hello! Please give me a brief response.", opts)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Result: %s\n", result.Result)
	fmt.Printf("Cost: $%.4f\n", result.CostUSD)
}

func runWithCustomBuffers(client *claude.ClaudeClient) {
	// Custom buffer configuration for memory-constrained environments
	bufferConfig := &buffer.Config{
		MaxStdoutSize:     100 * 1024, // 100KB
		MaxStderrSize:     10 * 1024,  // 10KB
		BufferTimeout:     15 * time.Second,
		EnableTruncation:  true,
		TruncationSuffix:  "\n\n[... Response truncated due to size limit ...]",
	}

	opts := &claude.RunOptions{
		Format:       claude.JSONOutput,
		BufferConfig: bufferConfig,
	}

	result, err := client.RunPromptCtx(context.Background(), 
		"Please explain machine learning in detail with examples.", opts)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Result length: %d characters\n", len(result.Result))
	fmt.Printf("Cost: $%.4f\n", result.CostUSD)
	
	// Check if response might have been truncated
	if len(result.Result) > 90*1024 { // Close to our limit
		fmt.Println("⚠️  Response may have been truncated due to buffer limits")
	}
}

func runWithLargeOutputLimits(client *claude.ClaudeClient) {
	// Configuration for handling potentially large outputs
	bufferConfig := &buffer.Config{
		MaxStdoutSize:     5 * 1024 * 1024, // 5MB
		MaxStderrSize:     500 * 1024,      // 500KB
		BufferTimeout:     60 * time.Second,
		EnableTruncation:  true,
		TruncationSuffix:  "\n\n[... Output truncated. Consider using streaming for large responses ...]",
	}

	opts := &claude.RunOptions{
		Format:       claude.JSONOutput,
		BufferConfig: bufferConfig,
		MaxTurns:     1, // Limit turns to control output size
	}

	result, err := client.RunPromptCtx(context.Background(), 
		"Generate a comprehensive programming tutorial covering multiple languages.", opts)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Result length: %d characters\n", len(result.Result))
	fmt.Printf("Duration: %dms\n", result.DurationMS)
	fmt.Printf("Cost: $%.4f\n", result.CostUSD)
}

func runStreamingWithBuffers(client *claude.ClaudeClient) {
	// Streaming configuration with buffer management
	bufferConfig := &buffer.Config{
		MaxStdoutSize:     2 * 1024 * 1024, // 2MB for streaming buffer
		MaxStderrSize:     100 * 1024,      // 100KB for errors
		BufferTimeout:     45 * time.Second,
		EnableTruncation:  false, // Don't truncate streaming responses
		TruncationSuffix:  "",
	}

	opts := &claude.RunOptions{
		Format:       claude.StreamJSONOutput,
		BufferConfig: bufferConfig,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	messageCh, errCh := client.StreamPrompt(ctx, 
		"Write a short story about artificial intelligence.", opts)

	fmt.Println("Streaming response:")
	responseLength := 0
	messageCount := 0

	for {
		select {
		case msg, ok := <-messageCh:
			if !ok {
				fmt.Printf("\nStreaming completed. Total messages: %d, Length: %d characters\n", 
					messageCount, responseLength)
				return
			}
			
			messageCount++
			if msg.Content != "" {
				responseLength += len(msg.Content)
				// Print first 100 characters of each message
				content := msg.Content
				if len(content) > 100 {
					content = content[:100] + "..."
				}
				fmt.Printf("Message %d: %s\n", messageCount, content)
			}

		case err := <-errCh:
			log.Printf("Streaming error: %v", err)
			return

		case <-ctx.Done():
			fmt.Println("Streaming timeout")
			return
		}
	}
}

// Additional utility functions for buffer management

func demonstrateBufferMetrics() {
	fmt.Println("\n=== Buffer Management Metrics ===")
	
	// Create a buffer manager with custom config
	config := &buffer.Config{
		MaxStdoutSize:    1024,
		MaxStderrSize:    512,
		BufferTimeout:    5 * time.Second,
		EnableTruncation: true,
		TruncationSuffix: "[TRUNCATED]",
	}
	
	bufManager := buffer.NewBufferManager(config)
	
	// Create and test buffers
	stdout := bufManager.NewStdoutBuffer()
	stderr := bufManager.NewStderrBuffer()
	
	// Write test data
	testData := make([]byte, 2000) // Exceeds both limits
	for i := range testData {
		testData[i] = 'A'
	}
	
	stdout.Write(testData)
	stderr.Write(testData)
	
	fmt.Printf("Stdout - Size: %d bytes, Truncated: %v\n", 
		stdout.Size(), stdout.Truncated())
	fmt.Printf("Stderr - Size: %d bytes, Truncated: %v\n", 
		stderr.Size(), stderr.Truncated())
}