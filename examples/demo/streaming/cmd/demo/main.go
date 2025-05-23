package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lancekrogers/claude-code-go/pkg/claude"
)

// isExitCommand checks if the input is a command to exit the REPL
func isExitCommand(input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	exitCommands := []string{
		"exit", "quit", "bye", "goodbye", "q", ":q", ":quit", ":exit",
		"done", "finish", "end", "stop", "close", "leave", "logout",
		"ctrl+c", "^c", "cancel", "abort", "terminate",
	}

	for _, cmd := range exitCommands {
		if input == cmd {
			return true
		}
	}
	return false
}

// displayStreamingMessage processes and displays streaming messages from Claude
func displayStreamingMessage(msg claude.Message) {
	switch msg.Type {
	case "system":
		if msg.Subtype == "init" {
			fmt.Printf("🔄 Initializing Claude session %s...\n", msg.SessionID[:8])
		}
	case "user":
		// User messages aren't shown as they're the input we provided
	case "assistant":
		// Parse assistant message content
		var content map[string]interface{}
		if err := json.Unmarshal(msg.Message, &content); err == nil {
			if contentArray, ok := content["content"].([]interface{}); ok {
				for _, item := range contentArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if itemMap["type"] == "text" {
							if text, ok := itemMap["text"].(string); ok && strings.TrimSpace(text) != "" {
								fmt.Printf("💬 Claude: %s\n", text)
							}
						} else if itemMap["type"] == "tool_use" {
							if name, ok := itemMap["name"].(string); ok {
								if input, ok := itemMap["input"].(map[string]interface{}); ok {
									displayToolUse(name, input)
								}
							}
						}
					}
				}
			}
		}
	case "result":
		if msg.Subtype == "success" {
			fmt.Printf("\n📊 Response complete - Duration: %.1fs | Turns: %d\n", 
				float64(msg.DurationMS)/1000.0, msg.NumTurns)
		}
	}
}

// displayToolUse shows tool execution in a user-friendly format
func displayToolUse(toolName string, input map[string]interface{}) {
	switch toolName {
	case "Bash":
		if command, ok := input["command"].(string); ok {
			fmt.Printf("🔧 Running: %s\n", command)
		}
	case "Write":
		if filePath, ok := input["file_path"].(string); ok {
			fmt.Printf("📝 Creating file: %s\n", filePath)
		}
	case "Edit":
		if filePath, ok := input["file_path"].(string); ok {
			fmt.Printf("✏️  Editing file: %s\n", filePath)
		}
	case "Read":
		if filePath, ok := input["file_path"].(string); ok {
			fmt.Printf("📖 Reading file: %s\n", filePath)
		}
	default:
		fmt.Printf("🛠️  Using tool: %s\n", toolName)
	}
}

const systemPrompt = `
You are a senior Go engineer with cryptocurrency experience interviewing
for a job. Create an it_works/ directory, then copy examples/demo/streaming/test-file.txt 
into it_works/test-file.txt. Create a Go program in it_works/keccac.go that prints
the Keccak hash of a file (go run keccak.go <file>). Use Go's built-in crypto/sha3 package (import "crypto/sha3") with sha3.New256() - this is Keccac, not SHA-256. 
Do NOT use golang.org/x/crypto/sha3 or external libraries. IMPORTANT: Only work within the it_works/ 
directory - do NOT modify any files outside it_works/ including go.work, go.mod, or other project files. 
After writing the code, cd into it_works/ and test it by generating the hash of 
test-file.txt and ../README.md to demonstrate it works. Briefly explain your
approach in one short paragraph (≤3 sentences, no bullet points), then ask the
interviewer if they would like you to start coding.`

func main() {
	// Create Claude client
	client := claude.NewClient("claude")

	// First call with system prompt using streaming
	fmt.Println("🚀 Starting streaming demo conversation...")
	fmt.Println("📡 Using real-time tool execution display\n")
	
	ctx := context.Background()
	messageCh, errCh := client.StreamPrompt(ctx, 
		"In <=3 sentences, describe your plan and ask if I want you to begin.",
		&claude.RunOptions{
			Format: claude.StreamJSONOutput,
			SystemPrompt: systemPrompt,
			AllowedTools: []string{
				"Write(it_works/*)",           // Create files only in it_works/
				"Edit(it_works/*)",            // Edit files only in it_works/
				"Read(it_works/*)",            // Read files only in it_works/
				"Read(examples/demo/streaming/test-file.txt)", // Read source test file
				"Bash(mkdir it_works*)",       // Create it_works directory only
				"Bash(chmod it_works/*)",      // Make files executable in it_works/
				"Bash(cd it_works*)",          // Change to it_works directory only
				"Bash(cp examples/demo/streaming/test-file.txt it_works/)", // Copy test file to it_works/
				"Bash(go:* it_works/*)",       // Run Go commands in it_works/
				"Bash(ls it_works*)",          // List it_works contents
				"Bash(cat it_works/*)",        // Display it_works files
				"Bash(echo)",                  // Create simple test content
				"Bash(pwd)",                   // Show current directory
			},
		})

	var sessionID string
	
	// Process initial streaming messages
	for {
		select {
		case msg, ok := <-messageCh:
			if !ok {
				goto repl // Channel closed, start REPL
			}
			displayStreamingMessage(msg)
			if msg.SessionID != "" {
				sessionID = msg.SessionID
			}
		case err := <-errCh:
			if err != nil {
				log.Fatalf("Error in initial prompt: %v", err)
			}
			goto repl // Error channel closed, start REPL
		}
	}

repl:

	// Start REPL loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")
		if !scanner.Scan() {
			// EOF or error
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			// Blank line, exit
			break
		}

		// Check for exit commands
		if isExitCommand(input) {
			fmt.Println("Goodbye! Thanks for trying the Claude Code Go SDK demo.")
			break
		}

		// Continue conversation with streaming
		fmt.Printf("\n🔄 Processing: %s\n", input)
		
		messageCh, errCh := client.StreamPrompt(ctx, input, &claude.RunOptions{
			Format:   claude.StreamJSONOutput,
			ResumeID: sessionID,
			AllowedTools: []string{
				"Write(it_works/*)",           // Create files only in it_works/
				"Edit(it_works/*)",            // Edit files only in it_works/
				"Read(it_works/*)",            // Read files only in it_works/
				"Read(examples/demo/streaming/test-file.txt)", // Read source test file
				"Bash(mkdir it_works*)",       // Create it_works directory only
				"Bash(chmod it_works/*)",      // Make files executable in it_works/
				"Bash(cd it_works*)",          // Change to it_works directory only
				"Bash(cp examples/demo/streaming/test-file.txt it_works/)", // Copy test file to it_works/
				"Bash(go:* it_works/*)",       // Run Go commands in it_works/
				"Bash(ls it_works*)",          // List it_works contents
				"Bash(cat it_works/*)",        // Display it_works files
				"Bash(echo)",                  // Create simple test content
				"Bash(pwd)",                   // Show current directory
			},
		})

		// Process streaming messages for this interaction
		for {
			select {
			case msg, ok := <-messageCh:
				if !ok {
					goto nextInput // Channel closed, get next input
				}
				displayStreamingMessage(msg)
				if msg.SessionID != "" {
					sessionID = msg.SessionID
				}
			case err := <-errCh:
				if err != nil {
					log.Printf("Error continuing conversation: %v", err)
				}
				goto nextInput // Error or completion, get next input
			}
		}
		
	nextInput:
	}

	fmt.Println("Demo completed!")
}
