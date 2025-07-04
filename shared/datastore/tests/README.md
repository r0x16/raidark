# Datastore Module Test Suite

This comprehensive test suite covers all aspects of the datastore module functionality, organized into three main categories as requested.

## Test Categories

### 1. Unit Tests (`./unit/`)
Basic unit tests that verify individual components can be instantiated and function correctly in isolation.

**Coverage:**
- DataStore instantiation and basic operations
- DatabaseProvider interface implementations
- Connection string generation for MySQL and PostgreSQL
- Error handling with invalid parameters
- Mock database operations

**Files:**
- `datastore_basic_test.go` - Core DataStore functionality
- `database_provider_test.go` - Provider implementations and DSN generation

### 2. Internal Tests (`./internal/`)
Tests focused on how components communicate with each other and their internal behavior.

**Coverage:**
- Communication between DataStore and DatabaseProvider
- Provider lifecycle management
- Shared datastore instances across multiple providers
- CRUD operations through provider interfaces
- Interface compliance verification
- Environment variable interactions
- Concurrent access patterns

**Files:**
- `component_communication_test.go` - Inter-component communication and behavior

### 3. Integration Tests (`./integration/`)
Full application tests that start the server and verify functionality through HTTP endpoints and browser simulation.

**Coverage:**
- Application startup and datastore integration
- API endpoint functionality with datastore
- Authenticated endpoints using datastore
- Error handling and graceful degradation
- Concurrent request handling
- Browser compatibility and accessibility
- Mobile device compatibility
- Performance under load
- Session management
- End-to-end browser simulation

**Files:**
- `application_integration_test.go` - Full application integration testing
- `browser_e2e_test.go` - Browser-focused end-to-end testing

## Quick Start

### Prerequisites
- Go 1.19 or higher
- Make (optional, for using Makefile commands)

### Installation
```bash
# Install test dependencies
make deps
# or manually:
go mod download
```

### Running Tests

#### Run All Tests
```bash
make test
# or
go test -v ./...
```

#### Run Specific Categories
```bash
# Unit tests only
make unit

# Internal communication tests only
make internal

# Integration tests only
make integration

# Browser E2E tests only
make browser
```

#### Environment-based Test Selection
```bash
# Run only unit tests
TEST_UNIT=true go test -v ./...

# Run unit and internal tests
TEST_UNIT=true TEST_INTERNAL=true go test -v ./...

# Run all tests
TEST_ALL=true go test -v ./...
```

## Test Configuration

### Environment Variables
Tests can be configured using environment variables:

```bash
# Test category selection
TEST_UNIT=true          # Enable unit tests
TEST_INTERNAL=true      # Enable internal tests  
TEST_INTEGRATION=true   # Enable integration tests
TEST_BROWSER=true       # Enable browser tests
TEST_ALL=true           # Enable all test categories

# Test behavior
TEST_VERBOSE=true       # Enable verbose output
TEST_TIMEOUT=30s        # Set test timeout
BROWSER_TESTS=true      # Enable browser tests in CI

# Database configuration for integration tests
DB_HOST=localhost
DB_PORT=3306
DB_USER=test
DB_PASSWORD=test
DB_DATABASE=test_db

# API configuration
API_PORT=8080
JWT_SECRET=test_secret_key
LOG_LEVEL=error
```

### Test Helper Usage
```go
package mytest

import (
    "testing"
    "github.com/r0x16/Raidark/shared/datastore/tests"
)

func TestExample(t *testing.T) {
    helper := tests.NewTestHelper(t)
    helper.SetupTest()
    defer helper.TeardownTest()
    
    // Your test code here
    helper.AssertNotNil(someObject)
}
```

## Advanced Usage

### Coverage Reports
```bash
make coverage
# Generates coverage.html with detailed coverage report
```

### Race Detection
```bash
make race
# Runs tests with Go's race detector
```

### CI/CD Integration
```bash
make ci
# Runs tests suitable for CI environments (excludes browser tests by default)
```

### Performance Testing
```bash
make bench
# Runs benchmark tests
```

### Code Quality
```bash
make lint
# Runs go vet and go fmt on test code
```

## Test Structure

### Unit Tests
- ✅ DataStore instantiation with valid/invalid parameters
- ✅ BaseModel structure verification
- ✅ DatabaseProvider interface compliance
- ✅ Connection string generation (MySQL/PostgreSQL)
- ✅ Error handling with mock databases
- ✅ Provider behavior before/after connection

### Internal Tests
- ✅ DataStore-Provider integration
- ✅ Provider lifecycle management
- ✅ Shared datastore scenarios
- ✅ CRUD operations through providers
- ✅ Interface compliance verification
- ✅ Environment variable handling
- ✅ Concurrent access patterns

### Integration Tests
- ✅ Full application startup
- ✅ HTTP API integration with datastore
- ✅ Authentication flow integration
- ✅ Error handling and recovery
- ✅ Concurrent request processing
- ✅ Browser compatibility testing
- ✅ Mobile device simulation
- ✅ Performance under load
- ✅ Session management
- ✅ End-to-end workflows

## Troubleshooting

### Common Issues

1. **Tests timeout**: Increase timeout with `TEST_TIMEOUT=60s`
2. **Database connection errors**: Check environment variables and database availability
3. **Browser tests fail in CI**: Set `BROWSER_TESTS=true` to enable
4. **Port conflicts**: Modify `API_PORT` in test configuration

### Debug Mode
```bash
# Enable verbose logging
LOG_LEVEL=debug make test

# Run specific test with verbose output
go test -v -run TestSpecificTest ./unit/
```

### Test Isolation
Each test category runs in isolation:
- Unit tests use in-memory SQLite databases
- Integration tests use separate ports and database names
- Environment variables are cleaned up after each test

## Contributing

When adding new tests:

1. **Unit tests**: Add to appropriate file in `./unit/`
2. **Internal tests**: Add to `./internal/component_communication_test.go`
3. **Integration tests**: Add to appropriate file in `./integration/`
4. **Update documentation**: Update this README with new test coverage
5. **Follow naming conventions**: Use descriptive test names with `Test` prefix

### Test Naming Convention
```go
// Unit tests
func TestDataStoreInstantiation(t *testing.T) { ... }
func TestMysqlConnectionDsnGeneration(t *testing.T) { ... }

// Internal tests  
func TestDataStoreProviderIntegration(t *testing.T) { ... }
func TestProviderLifecycleManagement(t *testing.T) { ... }

// Integration tests
func TestApplicationStartup(t *testing.T) { ... }
func TestBrowserAccessibility(t *testing.T) { ... }
```

## Test Coverage Goals

- **Unit Tests**: 90%+ coverage of individual functions
- **Internal Tests**: 85%+ coverage of component interactions
- **Integration Tests**: 80%+ coverage of end-to-end workflows

Current coverage can be viewed by running `make coverage` and opening `coverage.html`.