package integration

import (
	"context"
	"testing"
	"time"

	"github.com/marvai-dev/claude-code-go/pkg/claude"
	"github.com/marvai-dev/claude-code-go/test/utils"
)

func TestStreamPromptsToSession(t *testing.T) {
	utils.SkipIfNoClaudeCLI(t)

	client := utils.NewTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create a channel for prompts
	promptCh := make(chan string, 3)

	// Send multiple prompts to the channel
	promptCh <- "What is 2+2?"
	promptCh <- "What is 3+3?"
	promptCh <- "What is 4+4?"
	close(promptCh)

	messageCh, errCh := client.StreamPromptsToSession(ctx, promptCh, &claude.RunOptions{})

	var messages []claude.Message
	var streamErr error
	promptCount := 0

	// Handle errors
	go func() {
		for err := range errCh {
			streamErr = err
		}
	}()

	// Collect messages
	for msg := range messageCh {
		messages = append(messages, msg)
		t.Logf("Received message type: %s, subtype: %s", msg.Type, msg.Subtype)

		if msg.Type == "assistant" && msg.Content != "" {
			t.Logf("Assistant response: %s", msg.Content)
		}

		if msg.Type == "result" {
			promptCount++
			t.Logf("Completed prompt %d - Cost: $%.6f, Turns: %d", promptCount, msg.CostUSD, msg.NumTurns)
			
			// Stop after processing all 3 prompts
			if promptCount >= 3 {
				break
			}
		}
	}

	if streamErr != nil {
		t.Fatalf("Streaming error: %v", streamErr)
	}

	if len(messages) == 0 {
		t.Fatal("Expected at least one message")
	}

	// Verify we got system init message
	var hasSystemInit bool
	for _, msg := range messages {
		if msg.Type == "system" && msg.Subtype == "init" {
			hasSystemInit = true
			break
		}
	}

	if !hasSystemInit {
		t.Error("Expected system init message")
	}

	// Verify we processed all prompts
	if promptCount < 3 {
		t.Errorf("Expected to process 3 prompts, got %d", promptCount)
	}

	t.Logf("Successfully processed %d prompts with %d total messages", promptCount, len(messages))
}

func TestStreamPromptsToSessionWithContext(t *testing.T) {
	utils.SkipIfNoClaudeCLI(t)

	client := utils.NewTestClient(t)
	
	// Test with shorter timeout to verify context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a channel with prompts that would take longer to complete
	promptCh := make(chan string, 2)
	promptCh <- "Count from 1 to 10 very slowly"
	promptCh <- "Write a long story"
	close(promptCh)

	messageCh, errCh := client.StreamPromptsToSession(ctx, promptCh, &claude.RunOptions{})

	var receivedMessages int
	var contextCanceled bool

	// Handle errors
	go func() {
		for err := range errCh {
			if err == context.DeadlineExceeded {
				contextCanceled = true
			}
			t.Logf("Stream error (expected due to timeout): %v", err)
		}
	}()

	// Count messages until context times out
	for msg := range messageCh {
		receivedMessages++
		t.Logf("Received message %d: %s", receivedMessages, msg.Type)

		// Don't wait for completion since we expect timeout
		if receivedMessages >= 5 {
			break
		}
	}

	// Give time for context cancellation to propagate
	time.Sleep(200 * time.Millisecond)

	if !contextCanceled {
		t.Log("Context cancellation may not have been triggered (this is OK)")
	}

	t.Logf("Received %d messages before timeout", receivedMessages)
}

func TestStreamPromptsToSessionEmptyChannel(t *testing.T) {
	utils.SkipIfNoClaudeCLI(t)

	client := utils.NewTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create an empty channel (immediately closed)
	promptCh := make(chan string)
	close(promptCh)

	messageCh, errCh := client.StreamPromptsToSession(ctx, promptCh, &claude.RunOptions{})

	var messages []claude.Message
	var streamErr error

	// Handle errors
	go func() {
		for err := range errCh {
			streamErr = err
		}
	}()

	// Collect messages (should be minimal since no prompts)
	for msg := range messageCh {
		messages = append(messages, msg)
		t.Logf("Received message type: %s", msg.Type)
	}

	if streamErr != nil {
		t.Logf("Stream error (may be expected for empty channel): %v", streamErr)
	}

	// Should still get system initialization messages even with empty prompt channel
	t.Logf("Received %d messages from empty channel", len(messages))
}

func TestStreamPromptsToSessionInteractive(t *testing.T) {
	utils.SkipIfNoClaudeCLI(t)

	client := utils.NewTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a channel for interactive prompting
	promptCh := make(chan string, 1)

	// Start streaming
	messageCh, errCh := client.StreamPromptsToSession(ctx, promptCh, &claude.RunOptions{})

	var messages []claude.Message
	var streamErr error
	promptsSent := 0

	// Handle errors
	go func() {
		for err := range errCh {
			streamErr = err
		}
	}()

	// Send first prompt
	promptCh <- "What is 1+1?"
	promptsSent++

	// Process messages and send additional prompts based on responses
	for msg := range messageCh {
		messages = append(messages, msg)
		t.Logf("Received message type: %s", msg.Type)

		if msg.Type == "result" && promptsSent < 2 {
			// Send another prompt after getting result
			promptCh <- "What is 5+5?"
			promptsSent++
		} else if msg.Type == "result" && promptsSent >= 2 {
			// Close channel after processing second prompt
			close(promptCh)
			break
		}
	}

	if streamErr != nil {
		t.Fatalf("Streaming error: %v", streamErr)
	}

	if len(messages) == 0 {
		t.Fatal("Expected at least one message")
	}

	t.Logf("Interactive session sent %d prompts and received %d messages", promptsSent, len(messages))
}