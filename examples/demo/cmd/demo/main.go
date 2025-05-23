package main

import (
	"bufio"
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

const systemPrompt = `
You are a senior Go engineer with cryptocurrency experience interviewing
for a job. Create an it_works/ directory, then copy examples/demo/test-file.txt 
into it_works/test-file.txt. Create a Go program in it_works/keccac.go that prints
the Keccak hash of a file (go run keccak.go <file>). Use Go's built-in crypto/sha3 package. After writing the code, cd into it_works/ and test it by generating the hash of 
test-file.txt and ../README.md to demonstrate it works. Briefly explain your
approach in one short paragraph (≤3 sentences, no bullet points), then ask the
interviewer if they would like you to start coding.`

func main() {
	// Create Claude client
	client := claude.NewClient("claude")

	// First call with system prompt
	fmt.Println("Starting demo conversation...")
	result, err := client.RunWithSystemPrompt(
		"In <=3 sentences, describe your plan and ask if I want you to begin.",
		systemPrompt,
		&claude.RunOptions{
			Format: claude.JSONOutput,
			AllowedTools: []string{
				"Write",       // Create new files
				"Edit",        // Modify existing files
				"Read",        // Read file contents
				"Bash(mkdir)", // Create directories
				"Bash(chmod)", // Make files executable
				"Bash(go:*)",  // Run Go programs and build binaries
				"Bash(ls)",    // List directory contents
				"Bash(cat)",   // Display file contents
				"Bash(echo)",  // Create simple test content
				"Bash(pwd)",   // Show current directory
				"Bash(cp)",    // Copy files
				"Bash(cd)",    // Change directory
			},
		})
	if err != nil {
		log.Fatalf("Error running initial prompt: %v", err)
	}

	// Print Claude's first response
	fmt.Printf("Claude: %s\n\n", result.Result)

	// Capture session ID for conversation continuity
	sessionID := result.SessionID

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

		// Continue conversation with same session and permissions
		result, err := client.RunPrompt(input, &claude.RunOptions{
			Format:   claude.JSONOutput,
			ResumeID: sessionID,
			AllowedTools: []string{
				"Write",       // Create new files
				"Edit",        // Modify existing files
				"Read",        // Read file contents
				"Bash(mkdir)", // Create directories
				"Bash(chmod)", // Make files executable
				"Bash(go:*)",  // Run Go programs and build binaries
				"Bash(ls)",    // List directory contents
				"Bash(cat)",   // Display file contents
				"Bash(echo)",  // Create simple test content
				"Bash(pwd)",   // Show current directory
				"Bash(cp)",    // Copy files
				"Bash(cd)",    // Change directory
			},
		})
		if err != nil {
			log.Printf("Error continuing conversation: %v", err)
			continue
		}

		fmt.Printf("Claude: %s\n\n", result.Result)
	}

	fmt.Println("Demo completed!")
}
