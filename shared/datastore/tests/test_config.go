package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfig holds configuration for datastore tests
type TestConfig struct {
	DatabaseDriver   string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	APIPort          string
	JWTSecret        string
	LogLevel         string
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseDriver:   "sqlite",
		DatabaseHost:     "localhost",
		DatabasePort:     "3306",
		DatabaseUser:     "test",
		DatabasePassword: "test",
		DatabaseName:     "test_datastore",
		APIPort:          "8080",
		JWTSecret:        "test_secret_key",
		LogLevel:         "error",
	}
}

// SetupTestEnvironment sets up environment variables for testing
func (tc *TestConfig) SetupTestEnvironment() {
	os.Setenv("DB_DRIVER", tc.DatabaseDriver)
	os.Setenv("DB_HOST", tc.DatabaseHost)
	os.Setenv("DB_PORT", tc.DatabasePort)
	os.Setenv("DB_USER", tc.DatabaseUser)
	os.Setenv("DB_PASSWORD", tc.DatabasePassword)
	os.Setenv("DB_DATABASE", tc.DatabaseName)
	os.Setenv("API_PORT", tc.APIPort)
	os.Setenv("JWT_SECRET", tc.JWTSecret)
	os.Setenv("LOG_LEVEL", tc.LogLevel)
}

// CleanupTestEnvironment cleans up environment variables after testing
func (tc *TestConfig) CleanupTestEnvironment() {
	envVars := []string{
		"DB_DRIVER", "DB_HOST", "DB_PORT", "DB_USER",
		"DB_PASSWORD", "DB_DATABASE", "API_PORT", "JWT_SECRET", "LOG_LEVEL",
	}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

// TestHelper provides common testing utilities
type TestHelper struct {
	t      *testing.T
	config *TestConfig
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:      t,
		config: DefaultTestConfig(),
	}
}

// NewTestHelperWithConfig creates a new test helper with custom config
func NewTestHelperWithConfig(t *testing.T, config *TestConfig) *TestHelper {
	return &TestHelper{
		t:      t,
		config: config,
	}
}

// SetupTest sets up a test with the helper's configuration
func (th *TestHelper) SetupTest() {
	th.config.SetupTestEnvironment()
}

// TeardownTest cleans up after a test
func (th *TestHelper) TeardownTest() {
	th.config.CleanupTestEnvironment()
}

// AssertNoError asserts that an error is nil
func (th *TestHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
	assert.NoError(th.t, err, msgAndArgs...)
}

// AssertError asserts that an error is not nil
func (th *TestHelper) AssertError(err error, msgAndArgs ...interface{}) {
	assert.Error(th.t, err, msgAndArgs...)
}

// AssertEqual asserts that two values are equal
func (th *TestHelper) AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(th.t, expected, actual, msgAndArgs...)
}

// AssertNotNil asserts that a value is not nil
func (th *TestHelper) AssertNotNil(object interface{}, msgAndArgs ...interface{}) {
	assert.NotNil(th.t, object, msgAndArgs...)
}

// AssertNil asserts that a value is nil
func (th *TestHelper) AssertNil(object interface{}, msgAndArgs ...interface{}) {
	assert.Nil(th.t, object, msgAndArgs...)
}

// GetConfig returns the test configuration
func (th *TestHelper) GetConfig() *TestConfig {
	return th.config
}

// TestCategories defines the different test categories
type TestCategories struct {
	Unit        bool
	Internal    bool
	Integration bool
	Browser     bool
}

// ShouldRunCategory checks if a test category should run based on environment variables
func (tc *TestCategories) ShouldRunCategory(category string) bool {
	switch category {
	case "unit":
		return tc.Unit || os.Getenv("TEST_UNIT") == "true" || os.Getenv("TEST_ALL") == "true"
	case "internal":
		return tc.Internal || os.Getenv("TEST_INTERNAL") == "true" || os.Getenv("TEST_ALL") == "true"
	case "integration":
		return tc.Integration || os.Getenv("TEST_INTEGRATION") == "true" || os.Getenv("TEST_ALL") == "true"
	case "browser":
		return tc.Browser || os.Getenv("TEST_BROWSER") == "true" || os.Getenv("TEST_ALL") == "true"
	default:
		return false
	}
}

// DefaultTestCategories returns default test categories (all enabled)
func DefaultTestCategories() *TestCategories {
	return &TestCategories{
		Unit:        true,
		Internal:    true,
		Integration: true,
		Browser:     true,
	}
}

// UnitTestsOnly returns test categories with only unit tests enabled
func UnitTestsOnly() *TestCategories {
	return &TestCategories{
		Unit:        true,
		Internal:    false,
		Integration: false,
		Browser:     false,
	}
}

// SkipIfNotCategory skips a test if the category is not enabled
func SkipIfNotCategory(t *testing.T, category string) {
	categories := DefaultTestCategories()
	if !categories.ShouldRunCategory(category) {
		t.Skipf("Skipping %s tests (category not enabled)", category)
	}
}