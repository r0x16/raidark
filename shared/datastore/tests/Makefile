# Datastore Module Test Suite
# This Makefile provides commands to run different categories of tests for the datastore module

.PHONY: all test unit internal integration browser clean help deps

# Default target
all: test

# Help target
help:
	@echo "Datastore Module Test Suite"
	@echo ""
	@echo "Available targets:"
	@echo "  all         - Run all tests (default)"
	@echo "  test        - Run all tests"
	@echo "  unit        - Run unit tests only"
	@echo "  internal    - Run internal communication tests only"
	@echo "  integration - Run integration tests only"
	@echo "  browser     - Run browser E2E tests only"
	@echo "  deps        - Install test dependencies"
	@echo "  clean       - Clean test artifacts"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Environment Variables:"
	@echo "  TEST_VERBOSE=true    - Enable verbose test output"
	@echo "  TEST_TIMEOUT=30s     - Set test timeout (default: 30s)"
	@echo "  BROWSER_TESTS=true   - Enable browser tests in CI"

# Install test dependencies
deps:
	@echo "Installing test dependencies..."
	go mod download
	go install github.com/stretchr/testify@latest
	go install gorm.io/driver/sqlite@latest
	go install gorm.io/gorm@latest

# Run all tests
test: deps
	@echo "Running all datastore tests..."
	TEST_ALL=true go test -v -timeout=${TEST_TIMEOUT:-30s} ./...

# Run unit tests only
unit: deps
	@echo "Running unit tests..."
	TEST_UNIT=true go test -v -timeout=${TEST_TIMEOUT:-30s} ./unit/...

# Run internal communication tests only
internal: deps
	@echo "Running internal communication tests..."
	TEST_INTERNAL=true go test -v -timeout=${TEST_TIMEOUT:-30s} ./internal/...

# Run integration tests only
integration: deps
	@echo "Running integration tests..."
	TEST_INTEGRATION=true go test -v -timeout=${TEST_TIMEOUT:-30s} ./integration/...

# Run browser E2E tests only
browser: deps
	@echo "Running browser E2E tests..."
	TEST_BROWSER=true BROWSER_TESTS=true go test -v -timeout=${TEST_TIMEOUT:-60s} ./integration/ -run TestBrowserE2E

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	go clean -testcache
	rm -f *.log
	rm -f test_*.db

# Run tests with coverage
coverage: deps
	@echo "Running tests with coverage..."
	TEST_ALL=true go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests in CI mode (no browser tests unless explicitly enabled)
ci: deps
	@echo "Running tests in CI mode..."
	TEST_UNIT=true TEST_INTERNAL=true TEST_INTEGRATION=true go test -v -timeout=${TEST_TIMEOUT:-30s} ./unit/... ./internal/... -run "^((?!Browser).)*$$" ./integration/...

# Run tests with race detection
race: deps
	@echo "Running tests with race detection..."
	TEST_ALL=true go test -race -v -timeout=${TEST_TIMEOUT:-30s} ./...

# Run benchmarks
bench: deps
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint the test code
lint:
	@echo "Linting test code..."
	go vet ./...
	go fmt ./...

# Run tests in short mode (skip long-running tests)
short: deps
	@echo "Running tests in short mode..."
	TEST_ALL=true go test -short -v ./...