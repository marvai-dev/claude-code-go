package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lancekrogers/claude-code-go/pkg/claude"
)

const systemPrompt = "You are a senior Python engineer with cryptocurrency experience interviewing for a job. Create a `it_works/` directory containing a Python script that prints the Keccak hash of a file (`python keccak.py <file>`). Explain your approach and ask the interviewer if you should start coding."

func main() {
	// Create Claude client
	client := claude.NewClient("claude")

	// First call with system prompt
	fmt.Println("Starting demo conversation...")
	result, err := client.RunWithSystemPrompt("", systemPrompt, &claude.RunOptions{
		Format: claude.JSONOutput,
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

		// Continue conversation with same session
		result, err := client.ResumeConversation(input, sessionID)
		if err != nil {
			log.Printf("Error continuing conversation: %v", err)
			continue
		}

		fmt.Printf("Claude: %s\n\n", result.Result)
	}

	fmt.Println("Demo completed!")
}