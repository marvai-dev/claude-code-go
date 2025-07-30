package claude

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/marvai-dev/claude-code-go/pkg/claude/buffer"
)

func TestQuery_PythonSDKAlignment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	skipIfNoClaudeCLI(t)
	client := newTestClient(t)
	ctx := context.Background()

	// Test basic Query method with QueryOptions
	opts := QueryOptions{
		MaxTurns:       1,
		SystemPrompt:   "Test system prompt",
		AllowedTools:   []string{"Read", "Write"},
		PermissionMode: "ask",
		Format:         JSONOutput,
	}

	messageCh, err := client.Query(ctx, "Test prompt", opts)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if messageCh == nil {
		t.Fatal("Message channel is nil")
	}

	// Should receive at least one message
	select {
	case msg := <-messageCh:
		if msg.Type == "" {
			t.Error("Message type is empty")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestQuerySync_PythonSDKAlignment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	skipIfNoClaudeCLI(t)
	client := newTestClient(t)
	ctx := context.Background()

	// Test synchronous Query method
	opts := QueryOptions{
		MaxTurns:     1,
		SystemPrompt: "Test system prompt",
		Format:       JSONOutput,
	}

	result, err := client.QuerySync(ctx, "Test prompt", opts)
	if err != nil {
		t.Fatalf("QuerySync failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}
}

func TestQueryOptionsToRunOptions(t *testing.T) {
	client := &ClaudeClient{BinPath: "claude"}
	
	queryOpts := QueryOptions{
		MaxTurns:       3,
		SystemPrompt:   "Test prompt",
		WorkingDir:     "/test/dir",
		AllowedTools:   []string{"Read", "Write"},
		PermissionMode: "acceptEdits",
		Format:         JSONOutput,
		Model:          "claude-3-5-sonnet-20241022",
		ResumeID:       "test-session",
		Continue:       true,
		BufferConfig: &buffer.Config{
			MaxStdoutSize: 1024 * 1024,
		},
	}

	runOpts := client.queryOptionsToRunOptions(queryOpts)

	// Test field mappings
	if queryOpts.MaxTurns != runOpts.MaxTurns {
		t.Errorf("MaxTurns mismatch: got %d, want %d", runOpts.MaxTurns, queryOpts.MaxTurns)
	}
	if queryOpts.SystemPrompt != runOpts.SystemPrompt {
		t.Errorf("SystemPrompt mismatch: got %s, want %s", runOpts.SystemPrompt, queryOpts.SystemPrompt)
	}
	if queryOpts.Format != runOpts.Format {
		t.Errorf("Format mismatch: got %s, want %s", runOpts.Format, queryOpts.Format)
	}
	if queryOpts.Model != runOpts.Model {
		t.Errorf("Model mismatch: got %s, want %s", runOpts.Model, queryOpts.Model)
	}

	// Test permission mode mapping
	if !contains(runOpts.AllowedTools, "Write") {
		t.Error("AllowedTools should contain Write")
	}
	if !contains(runOpts.AllowedTools, "Edit") {
		t.Error("AllowedTools should contain Edit")
	}
	if !contains(runOpts.AllowedTools, "MultiEdit") {
		t.Error("AllowedTools should contain MultiEdit")
	}
}

func TestQueryOptionsPermissionModes(t *testing.T) {
	client := &ClaudeClient{BinPath: "claude"}

	tests := []struct {
		name           string
		permissionMode string
		expectedResult func(*RunOptions) bool
	}{
		{
			name:           "acceptEdits mode",
			permissionMode: "acceptEdits",
			expectedResult: func(opts *RunOptions) bool {
				return contains(opts.AllowedTools, "Write") &&
					contains(opts.AllowedTools, "Edit") &&
					contains(opts.AllowedTools, "MultiEdit")
			},
		},
		{
			name:           "rejectAll mode",
			permissionMode: "rejectAll",
			expectedResult: func(opts *RunOptions) bool {
				return contains(opts.DisallowedTools, "*")
			},
		},
		{
			name:           "ask mode (default)",
			permissionMode: "ask",
			expectedResult: func(opts *RunOptions) bool {
				// Should not add any special tools
				return len(opts.DisallowedTools) == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryOpts := QueryOptions{
				PermissionMode: tt.permissionMode,
			}

			runOpts := client.queryOptionsToRunOptions(queryOpts)
			if !tt.expectedResult(runOpts) {
				t.Errorf("Permission mode %s test failed", tt.permissionMode)
			}
		})
	}
}

func TestMessage_ContentFieldPopulation(t *testing.T) {
	// Test that Content field is populated from Message field for backward compatibility
	msg := Message{
		Type:    "user",
		Message: []byte(`{"content": "Hello world"}`),
	}

	// This would normally happen in StreamPrompt, but we're testing the logic
	if msg.Content == "" && len(msg.Message) > 0 {
		var messageContent struct {
			Content string `json:"content"`
			Text    string `json:"text"`
		}
		if err := json.Unmarshal(msg.Message, &messageContent); err == nil {
			if messageContent.Content != "" {
				msg.Content = messageContent.Content
			} else if messageContent.Text != "" {
				msg.Content = messageContent.Text
			}
		}
	}

	if msg.Content != "Hello world" {
		t.Errorf("Content mismatch: got %s, want Hello world", msg.Content)
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Local test helper functions to avoid import cycle with test/utils

// skipIfNoClaudeCLI skips the test if Claude Code CLI is not available
func skipIfNoClaudeCLI(t *testing.T) {
	if os.Getenv("USE_MOCK_SERVER") == "1" {
		return // Mock server tests don't need Claude CLI
	}
	
	claudePath := getTestClaudePath(t)
	if _, err := exec.LookPath(claudePath); err != nil {
		t.Skipf("Skipping test: Claude Code CLI not found at '%s'. Install Claude Code CLI or use mock server with USE_MOCK_SERVER=1", claudePath)
	}
	
	// Test if Claude CLI is working (it will handle auth automatically)
	cmd := exec.Command(claudePath, "--help")
	if err := cmd.Run(); err != nil {
		t.Skip("Skipping test: Claude Code CLI not working. Please ensure it's properly installed.")
	}
}

// newTestClient creates a Claude client for testing
func newTestClient(t *testing.T) *ClaudeClient {
	return NewClient(getTestClaudePath(t))
}

// getTestClaudePath returns the path to Claude CLI for testing
func getTestClaudePath(t *testing.T) string {
	if path := os.Getenv("CLAUDE_CODE_PATH"); path != "" {
		return path
	}
	
	// Check if mock server is preferred
	if os.Getenv("USE_MOCK_SERVER") == "1" {
		return createMockClaudeScript(t)
	}
	
	return "claude" // Default Claude Code CLI
}

// createMockClaudeScript creates a shell script that returns valid JSON
func createMockClaudeScript(t *testing.T) string {
	tempDir := t.TempDir()
	mockScript := filepath.Join(tempDir, "mock-claude.sh")
	
	content := `#!/bin/bash
# Mock Claude CLI that returns appropriate JSON response based on format
if [[ "$*" == *"--format json"* ]]; then
  # For regular JSON format (QuerySync)
  echo '{"result": "Mock response", "sessionId": "test-session", "costUSD": 0.001, "durationMS": 100}'
else
  # For streaming format or default (Query)
  echo '{"type": "system", "subtype": "init", "content": "Initializing"}'
  echo '{"type": "assistant", "content": "Mock streaming response"}'
  echo '{"type": "result", "result": "Mock response", "sessionId": "test-session", "costUSD": 0.001, "durationMS": 100}'
fi
`
	
	if err := os.WriteFile(mockScript, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}
	
	return mockScript
}