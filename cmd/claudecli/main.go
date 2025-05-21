package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/lancekrogers/claude-code-go/internal/claude"
)

func main() {
	// Parse command-line flags
	printMode := flag.Bool("p", false, "Run in non-interactive print mode")
	outputFormat := flag.String("output-format", "", "Output format (text, json, stream-json)")
	resumeID := flag.String("resume", "", "Resume a conversation by session ID")
	continueSession := flag.Bool("continue", false, "Continue the most recent conversation")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	maxTurns := flag.Int("max-turns", 0, "Maximum number of agentic turns")
	systemPrompt := flag.String("system-prompt", "", "Override system prompt")
	appendPrompt := flag.String("append-system-prompt", "", "Append to system prompt")
	allowedTools := flag.String("allowedTools", "", "Comma-separated list of allowed tools")
	disallowedTools := flag.String("disallowedTools", "", "Comma-separated list of disallowed tools")
	mcpConfig := flag.String("mcp-config", "", "Path to MCP configuration file")
	permissionTool := flag.String("permission-prompt-tool", "", "MCP tool for handling permission prompts")
	claudePath := flag.String("claude-path", "claude", "Path to the Claude Code binary")

	flag.Parse()

	// Get prompt from command-line arguments or stdin
	var prompt string
	args := flag.Args()

	if len(args) > 0 {
		prompt = strings.Join(args, " ")
	} else {
		// Check if there's data on stdin
		stdinStat, _ := os.Stdin.Stat()
		if (stdinStat.Mode() & os.ModeCharDevice) == 0 {
			// Data is being piped to stdin
			promptBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			prompt = string(promptBytes)
		}
	}

	// Create client options
	opts := &claude.RunOptions{
		ResumeID:       *resumeID,
		Continue:       *continueSession,
		Verbose:        *verbose,
		MaxTurns:       *maxTurns,
		SystemPrompt:   *systemPrompt,
		AppendPrompt:   *appendPrompt,
		MCPConfigPath:  *mcpConfig,
		PermissionTool: *permissionTool,
	}

	// Parse allowed/disallowed tools
	if *allowedTools != "" {
		opts.AllowedTools = strings.Split(*allowedTools, ",")
	}

	if *disallowedTools != "" {
		opts.DisallowedTools = strings.Split(*disallowedTools, ",")
	}

	// Set output format
	if *outputFormat != "" {
		switch *outputFormat {
		case "text":
			opts.Format = claude.TextOutput
		case "json":
			opts.Format = claude.JSONOutput
		case "stream-json":
			opts.Format = claude.StreamJSONOutput
		default:
			fmt.Fprintf(os.Stderr, "Invalid output format: %s\n", *outputFormat)
			os.Exit(1)
		}
	}

	// Create Claude client
	client := claude.NewClient(*claudePath)

	// Handle SIGINT (Ctrl+C)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		fmt.Fprintln(os.Stderr, "Interrupted")
		cancel()
	}()

	// Execute Claude command
	if *printMode {
		// Run in non-interactive print mode
		if opts.Format == claude.StreamJSONOutput {
			// Handle streaming
			messageCh, errCh := client.StreamPrompt(ctx, prompt, opts)

			// Handle errors in a separate goroutine
			go func() {
				for err := range errCh {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}()

			// Print messages as they arrive
			for msg := range messageCh {
				if opts.Format == claude.StreamJSONOutput {
					// For JSON streaming, output each message as JSON
					msgJSON, _ := json.Marshal(msg)
					fmt.Println(string(msgJSON))
				} else if msg.Type == "assistant" {
					// For text streaming, just output assistant messages
					fmt.Print(msg.Result)
				}
			}
		} else {
			// Handle non-streaming
			var result *claude.ClaudeResult
			var err error

			// Check if we're reading from stdin
			stdinStat, _ := os.Stdin.Stat()
			if (stdinStat.Mode()&os.ModeCharDevice) == 0 && prompt == "" {
				// We have data on stdin but no prompt, so use RunFromStdin
				result, err = client.RunFromStdin(os.Stdin, "", opts)
			} else {
				// Normal prompt execution
				result, err = client.RunPrompt(prompt, opts)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if opts.Format == claude.JSONOutput {
				// Output as JSON
				resultJSON, _ := json.Marshal(result)
				fmt.Println(string(resultJSON))
			} else {
				// Output as text
				fmt.Println(result.Result)
			}
		}
	} else {
		// Run in interactive mode (not implemented in this example)
		fmt.Println("Interactive mode not implemented in this Go wrapper.")
		fmt.Println("Use the original Claude Code CLI for interactive mode.")
		os.Exit(1)
	}
}
