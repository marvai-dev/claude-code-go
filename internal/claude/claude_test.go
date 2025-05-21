package claude

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// mockExecCommand returns a function that creates a mock command
func mockExecCommand(t *testing.T, expectedArgs []string, output string, exitCode int) func(name string, arg ...string) *exec.Cmd {
	return func(name string, arg ...string) *exec.Cmd {
		// Verify correct arguments were passed
		if len(arg) != len(expectedArgs) {
			t.Errorf("Expected %d arguments, got %d", len(expectedArgs), len(arg))
		}

		for i, a := range arg {
			if i < len(expectedArgs) && a != expectedArgs[i] {
				t.Errorf("Expected arg[%d] to be %q, got %q", i, expectedArgs[i], a)
			}
		}

		// Create a fake command that outputs our desired text and exits with the given code
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, arg...)

		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"GO_HELPER_OUTPUT=" + output,
			"GO_HELPER_EXIT_CODE=" + string(rune(exitCode)+'0'),
		}
		return cmd
	}
}

// TestHelperProcess isn't a real test - it's used to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	output := os.Getenv("GO_HELPER_OUTPUT")
	exitCode := int(os.Getenv("GO_HELPER_EXIT_CODE")[0] - '0')

	if output != "" {
		os.Stdout.Write([]byte(output))
	}

	os.Exit(exitCode)
}

func TestRunPrompt(t *testing.T) {
	// Save the original exec.Command and restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Test with text output
	execCommand = mockExecCommand(t, []string{"-p", "Hello, Claude"}, "Hello, human!", 0)

	client := &ClaudeClient{BinPath: "claude"}
	result, err := client.RunPrompt("Hello, Claude", &RunOptions{Format: TextOutput})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Result != "Hello, human!" {
		t.Errorf("Expected %q, got %q", "Hello, human!", result.Result)
	}

	// Test with JSON output
	jsonOutput := `{"type":"result","subtype":"success","cost_usd":0.001,"duration_ms":1234,"duration_api_ms":1000,"is_error":false,"num_turns":1,"result":"JSON response","session_id":"abc123"}`
	execCommand = mockExecCommand(t, []string{"-p", "JSON test", "--output-format", "json"}, jsonOutput, 0)

	result, err = client.RunPrompt("JSON test", &RunOptions{Format: JSONOutput})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Result != "JSON response" {
		t.Errorf("Expected %q, got %q", "JSON response", result.Result)
	}

	if result.SessionID != "abc123" {
		t.Errorf("Expected session ID %q, got %q", "abc123", result.SessionID)
	}

	if result.CostUSD != 0.001 {
		t.Errorf("Expected cost %f, got %f", 0.001, result.CostUSD)
	}

	// Test error handling
	execCommand = mockExecCommand(t, []string{"-p", "Error test"}, "", 1)

	_, err = client.RunPrompt("Error test", &RunOptions{})

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func TestStreamPrompt(t *testing.T) {
	// For streaming test, we'll create a simple mock that sends predefined messages
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Create a temporary file for our mock script
	tempDir := t.TempDir()
	mockScript := filepath.Join(tempDir, "mock_stream.go")

	// Write a simple Go program that outputs JSON messages
	scriptContent := `package main
import (
	"fmt"
	"time"
)
func main() {
	fmt.Println(` + "`" + `{"type":"system","subtype":"init","session_id":"test-session","tools":["Bash"]}` + "`" + `)
	time.Sleep(100 * time.Millisecond)
	fmt.Println(` + "`" + `{"type":"assistant","message":{},"session_id":"test-session","result":"Hello there!"}` + "`" + `)
	time.Sleep(100 * time.Millisecond)
	fmt.Println(` + "`" + `{"type":"result","subtype":"success","cost_usd":0.002,"duration_ms":300,"duration_api_ms":250,"is_error":false,"num_turns":1,"result":"Final result","session_id":"test-session"}` + "`" + `)
}
`
	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	// Compile the mock script
	mockBinary := filepath.Join(tempDir, "mock_stream")
	compileCmd := exec.Command("go", "build", "-o", mockBinary, mockScript)
	if err := compileCmd.Run(); err != nil {
		t.Fatalf("Failed to compile mock script: %v", err)
	}

	// Replace execCommand with our mock
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command(mockBinary)
	}

	// Now test streaming
	client := &ClaudeClient{BinPath: "claude"}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	messageCh, errCh := client.StreamPrompt(ctx, "Test streaming", &RunOptions{Format: StreamJSONOutput})

	// Collect messages and errors
	var messages []Message
	var streamErr error

	// Handle possible errors
	go func() {
		for err := range errCh {
			streamErr = err
		}
	}()

	// Collect messages
	for msg := range messageCh {
		messages = append(messages, msg)
	}

	// Check for streaming errors
	if streamErr != nil {
		t.Fatalf("Streaming error: %v", streamErr)
	}

	// Verify we got the expected messages
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Check first message (init)
	if messages[0].Type != "system" || messages[0].Subtype != "init" {
		t.Errorf("First message should be system/init, got %s/%s", messages[0].Type, messages[0].Subtype)
	}

	// Check last message (result)
	if messages[2].Type != "result" || messages[2].CostUSD != 0.002 {
		t.Errorf("Last message should be result with cost 0.002, got %s with cost %f",
			messages[2].Type, messages[2].CostUSD)
	}
}

func TestRunFromStdin(t *testing.T) {
	origExecCommand := exec.Command
	defer func() { exec.Command = origExecCommand }()

	// Test with text input from stdin
	exec.Command = mockCommand(t, []string{"-p"}, "Analyzed your input", 0)

	client := &ClaudeClient{BinPath: "claude"}
	stdin := bytes.NewBufferString("Code to analyze")

	result, err := client.RunFromStdin(stdin, "", &RunOptions{Format: TextOutput})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Result != "Analyzed your input" {
		t.Errorf("Expected %q, got %q", "Analyzed your input", result.Result)
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		opts     *RunOptions
		expected []string
	}{
		{
			name:     "Basic prompt",
			prompt:   "Hello",
			opts:     &RunOptions{},
			expected: []string{"-p", "Hello"},
		},
		{
			name:   "All options",
			prompt: "Complete test",
			opts: &RunOptions{
				Format:          JSONOutput,
				SystemPrompt:    "Custom system prompt",
				AppendPrompt:    "Additional instructions",
				MCPConfigPath:   "/path/to/mcp.json",
				AllowedTools:    []string{"tool1", "tool2"},
				DisallowedTools: []string{"bad1", "bad2"},
				PermissionTool:  "permit_tool",
				ResumeID:        "session123",
				MaxTurns:        5,
				Verbose:         true,
			},
			expected: []string{
				"-p", "Complete test",
				"--output-format", "json",
				"--system-prompt", "Custom system prompt",
				"--append-system-prompt", "Additional instructions",
				"--mcp-config", "/path/to/mcp.json",
				"--allowedTools", "tool1,tool2",
				"--disallowedTools", "bad1,bad2",
				"--permission-prompt-tool", "permit_tool",
				"--resume", "session123",
				"--max-turns", "5",
				"--verbose",
			},
		},
		{
			name:   "Continue session",
			prompt: "Continue",
			opts: &RunOptions{
				Continue: true,
			},
			expected: []string{"-p", "Continue", "--continue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildArgs(tt.prompt, tt.opts)

			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d arguments, got %d", len(tt.expected), len(args))
				t.Logf("Expected: %v", tt.expected)
				t.Logf("Got: %v", args)
				return
			}

			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("Expected arg[%d] to be %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("claude-bin-path")

	if client.BinPath != "claude-bin-path" {
		t.Errorf("Expected BinPath to be %q, got %q", "claude-bin-path", client.BinPath)
	}

	if client.DefaultOptions == nil {
		t.Error("DefaultOptions should not be nil")
	}

	if client.DefaultOptions.Format != TextOutput {
		t.Errorf("Expected default format to be %q, got %q", TextOutput, client.DefaultOptions.Format)
	}
}
