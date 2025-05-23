# Claude Code Go SDK - Makefile
# Alternative to Taskfile.yml for developers who prefer Make

# Variables
PROJECT_NAME := Claude Code Go SDK
BIN_DIR := ./bin
COVERAGE_DIR := ./coverage
GO_VERSION := 1.20

# Colors for output
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

# Default target
.DEFAULT_GOAL := help

# Phony targets (not files)
.PHONY: all build build-lib build-examples build-demo build-dangerous-example
.PHONY: test test-lib test-dangerous test-integration test-integration-real test-local coverage
.PHONY: demo run-dangerous check-go check-claude
.PHONY: clean help banner

##@ Build Targets

all: banner build test ## Build and test the SDK
	@echo "$(GREEN)✅ Build and test completed$(RESET)"

build: build-lib build-examples ## Build the SDK and all examples

build-lib: ## Build the core library
	@echo "$(BLUE)🔨 Building core library...$(RESET)"
	@go build ./pkg/claude
	@echo "$(GREEN)✅ Core library built successfully$(RESET)"

build-examples: build-demo build-dangerous-example ## Build all example programs
	@echo "$(BLUE)🔨 Building examples...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/basic-example ./examples/basic || echo "$(RED)❌ Basic example build failed$(RESET)"
	@go build -o $(BIN_DIR)/advanced-example ./examples/advanced || echo "$(RED)❌ Advanced example build failed$(RESET)"
	@go build -o $(BIN_DIR)/testing-example ./examples/testing || echo "$(RED)❌ Testing example build failed$(RESET)"
	@echo "$(GREEN)✅ Example builds completed$(RESET)"

build-demo: ## Build the interactive demo
	@echo "$(BLUE)🔨 Building demo...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@cd examples/demo && go mod tidy && go build -o ../../$(BIN_DIR)/demo ./cmd/demo
	@echo "$(GREEN)✅ Demo built successfully$(RESET)"

build-dangerous-example: ## Build dangerous usage example
	@echo "$(BLUE)🔨 Building dangerous example...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@cd examples/dangerous_usage && go mod tidy && go build -o ../../$(BIN_DIR)/dangerous-example .
	@echo "$(GREEN)✅ Dangerous example built successfully$(RESET)"

##@ Test Targets

test: test-lib ## Run unit tests

test-lib: ## Test the core library
	@echo "$(BLUE)🧪 Testing core library...$(RESET)"
	@go test -v ./pkg/claude || echo "$(RED)❌ Core library tests failed$(RESET)"
	@echo "$(GREEN)✅ Core library tests completed$(RESET)"

test-dangerous: ## Test dangerous package (security-sensitive features)
	@echo "$(YELLOW)🚨 Testing dangerous package...$(RESET)"
	@go test -v ./pkg/claude/dangerous || echo "$(RED)❌ Dangerous package tests failed$(RESET)"
	@echo "$(GREEN)✅ Dangerous package tests completed$(RESET)"

test-integration: ## Run integration tests with mock server
	@echo "$(BLUE)🔗 Running integration tests (mock server)...$(RESET)"
	@go test -v ./test/integration || echo "$(RED)❌ Integration tests failed$(RESET)"
	@echo "$(GREEN)✅ Integration tests completed$(RESET)"

test-integration-real: ## Run integration tests with real Claude CLI
	@echo "$(BLUE)🔗 Running integration tests (real Claude CLI)...$(RESET)"
	@CLAUDE_INTEGRATION_TEST=real go test -v ./test/integration || echo "$(RED)❌ Real integration tests failed$(RESET)"
	@echo "$(GREEN)✅ Real integration tests completed$(RESET)"

test-local: test test-integration ## Run all tests locally
	@echo "$(GREEN)✅ All local tests completed$(RESET)"

coverage: ## Generate test coverage report
	@echo "$(BLUE)📊 Generating coverage report...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./pkg/... || echo "$(RED)❌ Coverage generation failed$(RESET)"
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out || echo "$(RED)❌ Coverage summary failed$(RESET)"
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html || echo "$(RED)❌ HTML coverage report failed$(RESET)"
	@echo "$(GREEN)✅ Coverage generation completed$(RESET)"
	@echo "$(BLUE)📄 View HTML report at $(COVERAGE_DIR)/coverage.html$(RESET)"

##@ Demo and Examples

demo: build-demo check-go check-claude ## Run the interactive Claude Code Go SDK demo
	@echo "$(BLUE)🚀 Claude Code Go SDK Demo$(RESET)"
	@echo "$(BLUE)===========================$(RESET)"
	@echo ""
	@echo "$(BLUE)🎯 Starting interactive demo...$(RESET)"
	@echo "$(YELLOW)   Type your responses and press Enter$(RESET)"
	@echo "$(YELLOW)   Press Enter on empty line to exit$(RESET)"
	@echo ""
	@$(BIN_DIR)/demo

run-dangerous: build-dangerous-example check-dangerous ## Run dangerous features example (development only)
	@echo "$(YELLOW)🚨 Running Dangerous Features Example$(RESET)"
	@echo "$(YELLOW)=====================================$(RESET)"
	@echo "$(GREEN)✔️  Security requirements met$(RESET)"
	@echo ""
	@$(BIN_DIR)/dangerous-example

##@ Utility Targets

clean: ## Clean build artifacts and test cache
	@echo "$(BLUE)🧹 Cleaning build artifacts...$(RESET)"
	@rm -rf $(BIN_DIR) $(COVERAGE_DIR)
	@go clean -testcache
	@echo "$(GREEN)✅ Clean completed$(RESET)"

banner: ## Display project banner
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(BLUE)║                    $(PROJECT_NAME)                     ║$(RESET)"
	@echo "$(BLUE)║                      Build Pipeline                          ║$(RESET)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""

##@ Validation Targets

check-go: ## Check Go version requirements
	@if ! command -v go &> /dev/null; then \
		echo "$(RED)❌ Error: Go is not installed or not in PATH$(RESET)"; \
		exit 1; \
	fi
	@go_version=$$(go version | awk '{print $$3}' | sed 's/go//'); \
	major_version=$$(echo $$go_version | cut -d. -f1); \
	minor_version=$$(echo $$go_version | cut -d. -f2); \
	if [ $$major_version -lt 1 ] || [ $$major_version -eq 1 -a $$minor_version -lt 20 ]; then \
		echo "$(RED)❌ Error: Go ≥1.20 is required (found: $$go_version)$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✔️  Go version: $$(go version | awk '{print $$3}' | sed 's/go//')$(RESET)"

check-claude: ## Check Claude CLI availability
	@if ! claude_path=$$(command -v claude 2>/dev/null); then \
		echo "$(RED)❌ Error: claude CLI not found in PATH$(RESET)"; \
		echo "$(YELLOW)   Please install from: https://docs.anthropic.com/en/docs/claude-code/getting-started$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✔️  Found claude CLI: $$(command -v claude)$(RESET)"

check-dangerous: ## Check dangerous operation requirements
	@if [ "$(CLAUDE_ENABLE_DANGEROUS)" != "i-accept-all-risks" ]; then \
		echo "$(RED)❌ Error: CLAUDE_ENABLE_DANGEROUS must be set to 'i-accept-all-risks'$(RESET)"; \
		echo "$(YELLOW)   export CLAUDE_ENABLE_DANGEROUS=\"i-accept-all-risks\"$(RESET)"; \
		exit 1; \
	fi
	@if [ "$(NODE_ENV)" = "production" ] || [ "$(GO_ENV)" = "production" ] || [ "$(ENVIRONMENT)" = "production" ]; then \
		echo "$(RED)❌ Error: Cannot run dangerous example in production environment$(RESET)"; \
		exit 1; \
	fi

##@ Information

help: ## Display this help message
	@echo "$(BLUE)$(PROJECT_NAME) - Available Commands$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make $(BLUE)<target>$(RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  $(BLUE)make build$(RESET)                    Build the SDK and examples"
	@echo "  $(BLUE)make demo$(RESET)                     Run interactive demo"
	@echo "  $(BLUE)make test coverage$(RESET)            Run tests and generate coverage"
	@echo "  $(BLUE)CLAUDE_ENABLE_DANGEROUS=\"i-accept-all-risks\" make run-dangerous$(RESET)"
	@echo ""
	@echo "$(YELLOW)Alternative:$(RESET) You can also use $(BLUE)task <command>$(RESET) (see Taskfile.yml)"

version: ## Show version information
	@echo "$(BLUE)$(PROJECT_NAME)$(RESET)"
	@echo "Go version: $$(go version)"
	@echo "Make version: $$(make --version | head -n1)"
	@if command -v task &> /dev/null; then \
		echo "Task version: $$(task --version)"; \
	else \
		echo "Task: not installed"; \
	fi
	@if command -v claude &> /dev/null; then \
		echo "Claude CLI: available"; \
	else \
		echo "Claude CLI: not found"; \
	fi