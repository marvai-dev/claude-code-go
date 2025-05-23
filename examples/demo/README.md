# Claude Code Go SDK Demo

A minimal REPL demo that showcases the Claude Code Go SDK capabilities.

## What this demo does

1. Verifies that the `claude` CLI is installed and accessible
2. Uses the SDK's `RunWithSystemPrompt()` method to start a conversation
3. Maintains session continuity using `ResumeConversation()`
4. Implements a simple stdin/stdout REPL with Go standard library only

## Quick Start

```bash
# From the project root
./scripts/demo.sh
```

## Expected Output

```
✔️  Found claude CLI: /usr/local/bin/claude
Starting demo conversation...
Claude: [Initial response about Python/crypto engineering approach]

>>> Okay, go ahead and write the script.
Claude: [Continues the conversation in the same context]

>>> [Press Enter to exit]
Demo completed!
```

## Architecture

- **No external dependencies**: Uses only Go standard library
- **Session management**: Demonstrates multi-turn conversations
- **Error handling**: Graceful error handling and reporting
- **CLI validation**: Ensures Claude CLI is available before starting

This demo proves the SDK successfully wraps the Claude CLI and maintains conversation state across multiple interactions.