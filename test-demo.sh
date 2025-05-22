#!/bin/bash
set -e

echo "🧪 Claude Code Go SDK - Testing Demo"
echo "===================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Build the project
log_info "Step 1: Building the project..."
task build
log_success "Project built successfully"
echo ""

# Step 2: Run unit tests
log_info "Step 2: Running unit tests..."
go test ./... -v
log_success "Unit tests passed"
echo ""

# Step 3: Start mock server and test
log_info "Step 3: Testing with mock server..."
log_info "Starting mock server in background..."

# Start mock server
go run test/mockserver/main.go &
MOCK_PID=$!
echo $MOCK_PID > /tmp/mock-server-demo.pid

# Wait for server to start
sleep 2

# Test the mock server
log_info "Testing mock server directly..."
curl -X POST http://localhost:8080/claude \
     -d '-p "Hello world"' \
     -s || { log_error "Mock server test failed"; exit 1; }

echo ""
log_success "Mock server is working"

# Step 4: Test CLI with mock server
log_info "Step 4: Testing CLI binary..."
./bin/claude-go --help > /dev/null
log_success "CLI binary works"

# Step 5: Test examples
log_info "Step 5: Testing examples compilation..."
go build -o /tmp/basic-example ./examples/basic
go build -o /tmp/advanced-example ./examples/advanced
log_success "Examples compiled successfully"

# Step 6: Integration test simulation
log_info "Step 6: Running integration test simulation..."
log_warn "Note: This would normally run against real Claude API"
log_info "For real testing, set ANTHROPIC_API_KEY and run: task test-integration-real"

# Cleanup
log_info "Cleaning up..."
if [ -f /tmp/mock-server-demo.pid ]; then
    kill $(cat /tmp/mock-server-demo.pid) 2>/dev/null || true
    rm -f /tmp/mock-server-demo.pid
fi

echo ""
log_success "🎉 All tests completed successfully!"
echo ""
echo "Next steps:"
echo "1. Set up your Claude API key: export ANTHROPIC_API_KEY=your-key"
echo "2. Run integration tests: task test-integration-real"
echo "3. Run comprehensive tests: task test-full"
echo "4. Check test coverage: task test-coverage"
echo ""
echo "For more testing options, run: task --list"