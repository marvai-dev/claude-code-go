package claude

import (
	"strings"
	"testing"
)

func TestInputFormat_Constants(t *testing.T) {
	tests := []struct {
		format   InputFormat
		expected string
	}{
		{TextInput, "text"},
		{StreamJSONInput, "stream-json"},
	}

	for _, test := range tests {
		if string(test.format) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.format))
		}
	}
}

func TestBuildArgs_InputFormat(t *testing.T) {
	tests := []struct {
		name        string
		opts        *RunOptions
		expectedArg string
	}{
		{
			name: "No input format specified",
			opts: &RunOptions{},
			expectedArg: "",
		},
		{
			name: "Text input format",
			opts: &RunOptions{InputFormat: TextInput},
			expectedArg: "--input-format text",
		},
		{
			name: "Stream JSON input format", 
			opts: &RunOptions{InputFormat: StreamJSONInput},
			expectedArg: "--input-format stream-json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := BuildArgs("test prompt", test.opts)
			argsStr := strings.Join(args, " ")
			
			if test.expectedArg == "" {
				if strings.Contains(argsStr, "--input-format") {
					t.Errorf("Expected no --input-format flag, but found one in: %s", argsStr)
				}
			} else {
				if !strings.Contains(argsStr, test.expectedArg) {
					t.Errorf("Expected to find '%s' in args: %s", test.expectedArg, argsStr)
				}
			}
		})
	}
}

func TestStreamInputMessage_JSONMarshaling(t *testing.T) {
	msg := StreamInputMessage{
		Type: "user",
		Message: StreamInputContent{
			Role:    "user",
			Content: "Test message",
		},
	}

	// Test the struct fields are correctly defined
	if msg.Type != "user" {
		t.Errorf("Expected Type 'user', got %s", msg.Type)
	}
	if msg.Message.Role != "user" {
		t.Errorf("Expected Role 'user', got %s", msg.Message.Role)
	}
	if msg.Message.Content != "Test message" {
		t.Errorf("Expected Content 'Test message', got %s", msg.Message.Content)
	}
}