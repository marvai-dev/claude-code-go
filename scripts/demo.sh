#!/bin/bash
set -e

echo "🚀 Claude Code Go SDK Demo"
echo "=========================="

# Check Go version
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
major_version=$(echo $go_version | cut -d. -f1)
minor_version=$(echo $go_version | cut -d. -f2)

if [[ $major_version -lt 1 ]] || [[ $major_version -eq 1 && $minor_version -lt 20 ]]; then
    echo "❌ Error: Go ≥1.20 is required (found: $go_version)"
    exit 1
fi

echo "✔️  Go version: $go_version"

# Check for claude CLI
if ! claude_path=$(command -v claude 2>/dev/null); then
    echo "❌ Error: claude CLI not found in PATH"
    echo "   Please install from: https://docs.anthropic.com/en/docs/claude-code/getting-started"
    exit 1
fi

echo "✔️  Found claude CLI: $claude_path"

# Create bin directory
mkdir -p bin

# Build the demo
echo "🔨 Building demo..."
cd examples/demo
go build -o ../../bin/demo ./cmd/demo
cd ../..

echo "✔️  Demo built successfully"
echo ""
echo "🎯 Starting interactive demo..."
echo "   Type your responses and press Enter"
echo "   Press Enter on empty line to exit"
echo ""

# Run the demo
./bin/demo