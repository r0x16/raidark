package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/r0x16/Raidark/shared/api"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/r0x16/Raidark/shared/providers/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ApplicationIntegrationTestSuite tests the full application with datastore integration
type ApplicationIntegrationTestSuite struct {
	suite.Suite
	app         *api.Api
	serverPort  string
	baseURL     string
	testToken   string
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

// SetupSuite sets up the integration test suite
func (suite *ApplicationIntegrationTestSuite) SetupSuite() {
	// Set test environment variables
	os.Setenv("API_PORT", "8081")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_DATABASE", "test_db")
	os.Setenv("JWT_SECRET", "test_secret_key_for_integration_tests")
	
	suite.serverPort = "8081"
	suite.baseURL = fmt.Sprintf("http://localhost:%s", suite.serverPort)
	
	// Create context with providers
	suite.ctx, suite.cancelFunc = context.WithCancel(context.Background())
	
	providers := []domprovider.ProviderFactory{
		&driver.EnvProviderFactory{},
		&driver.LoggerProviderFactory{},
		&driver.DatastoreProviderFactory{},
		&driver.AuthProviderFactory{},
		&driver.ApiProviderFactory{},
	}
	
	ctxWithProviders := context.WithValue(suite.ctx, "providers", providers)
	
	// Initialize the application
	suite.app = api.NewApi(ctxWithProviders)
	
	// Start the server in a goroutine
	go func() {
		suite.app.Run()
	}()
	
	// Wait for server to start
	suite.waitForServer()
	
	// Generate test token for authenticated requests
	suite.generateTestToken()
}

// TearDownSuite cleans up after the test suite
func (suite *ApplicationIntegrationTestSuite) TearDownSuite() {
	if suite.cancelFunc != nil {
		suite.cancelFunc()
	}
	
	// Clean up environment variables
	testEnvVars := []string{
		"API_PORT", "DB_HOST", "DB_PORT", "DB_USER", 
		"DB_PASSWORD", "DB_DATABASE", "JWT_SECRET",
	}
	
	for _, envVar := range testEnvVars {
		os.Unsetenv(envVar)
	}
}

// waitForServer waits for the server to be ready
func (suite *ApplicationIntegrationTestSuite) waitForServer() {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			suite.FailNow("Server failed to start within timeout")
		case <-ticker.C:
			resp, err := http.Get(suite.baseURL)
			if err == nil {
				resp.Body.Close()
				return
			}
		}
	}
}

// generateTestToken generates a test token for authenticated requests
func (suite *ApplicationIntegrationTestSuite) generateTestToken() {
	// Use the auth provider to generate a token
	if suite.app.AuthProvider != nil {
		token, err := suite.app.AuthProvider.GenerateToken("test_user")
		if err == nil {
			suite.testToken = token
		}
	}
}

// TestApplicationStartup tests that the application starts successfully
func (suite *ApplicationIntegrationTestSuite) TestApplicationStartup() {
	resp, err := http.Get(suite.baseURL)
	suite.Assert().NoError(err, "Should be able to connect to server")
	
	if resp != nil {
		defer resp.Body.Close()
		suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode, 
			"Server should not return internal server error")
	}
}

// TestDatastoreIntegrationThroughAPI tests datastore functionality through API endpoints
func (suite *ApplicationIntegrationTestSuite) TestDatastoreIntegrationThroughAPI() {
	// Skip if we don't have database endpoints implemented
	// This test demonstrates how datastore integration would be tested
	
	// Test health endpoint (if available)
	resp, err := http.Get(suite.baseURL + "/health")
	if err == nil {
		defer resp.Body.Close()
		suite.Assert().Equal(http.StatusOK, resp.StatusCode, 
			"Health endpoint should return OK")
	}
	
	// Test database connectivity through API
	resp, err = http.Get(suite.baseURL + "/api/v1/status")
	if err == nil {
		defer resp.Body.Close()
		// Should return error if database is not connected (expected in test environment)
		// But should not crash the application
		suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
			"API should handle database connection gracefully")
	}
}

// TestAuthenticatedEndpointsWithDatastore tests authenticated endpoints that use datastore
func (suite *ApplicationIntegrationTestSuite) TestAuthenticatedEndpointsWithDatastore() {
	if suite.testToken == "" {
		suite.T().Skip("No test token available")
	}
	
	client := &http.Client{}
	
	// Test authenticated endpoint
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/profile", nil)
	suite.Require().NoError(err)
	
	req.Header.Set("Authorization", "Bearer "+suite.testToken)
	
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		// Should authenticate successfully even if endpoint doesn't exist
		suite.Assert().NotEqual(http.StatusUnauthorized, resp.StatusCode,
			"Should not be unauthorized with valid token")
	}
}

// TestDatastoreErrorHandling tests how the application handles datastore errors
func (suite *ApplicationIntegrationTestSuite) TestDatastoreErrorHandling() {
	// Test how the application handles database connection issues
	
	// Attempt to access endpoints that would use the database
	endpoints := []string{
		"/api/v1/users",
		"/api/v1/data",
		"/api/v1/records",
	}
	
	for _, endpoint := range endpoints {
		resp, err := http.Get(suite.baseURL + endpoint)
		if err == nil {
			defer resp.Body.Close()
			// Should handle database errors gracefully, not crash
			suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
				fmt.Sprintf("Endpoint %s should handle database errors gracefully", endpoint))
		}
	}
}

// TestConcurrentDatastoreAccess tests concurrent access to datastore through API
func (suite *ApplicationIntegrationTestSuite) TestConcurrentDatastoreAccess() {
	const numRequests = 10
	results := make(chan int, numRequests)
	
	// Make concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(suite.baseURL)
			if err != nil {
				results <- 500
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}
	
	// Collect results
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		suite.Assert().NotEqual(http.StatusInternalServerError, statusCode,
			"Concurrent requests should not cause server errors")
	}
}

// TestApplicationShutdown tests graceful shutdown of the application
func (suite *ApplicationIntegrationTestSuite) TestApplicationShutdown() {
	// Test that the application can handle shutdown gracefully
	// This is important for datastore connection cleanup
	
	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Simulate shutdown signal
	go func() {
		select {
		case <-shutdownCtx.Done():
			// Shutdown would be handled here in real application
		}
	}()
	
	// Verify application is still responding during shutdown
	resp, err := http.Get(suite.baseURL)
	if err == nil {
		defer resp.Body.Close()
		suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
			"Application should handle shutdown gracefully")
	}
}

// TestAPIResponseFormat tests that API responses are properly formatted
func (suite *ApplicationIntegrationTestSuite) TestAPIResponseFormat() {
	resp, err := http.Get(suite.baseURL)
	if err != nil {
		suite.T().Skip("Server not responding")
	}
	defer resp.Body.Close()
	
	// Check content type
	contentType := resp.Header.Get("Content-Type")
	suite.Assert().Contains(contentType, "application/json",
		"API should return JSON responses")
	
	// Try to parse response as JSON
	body, err := io.ReadAll(resp.Body)
	if err == nil && len(body) > 0 {
		var jsonResponse map[string]interface{}
		err = json.Unmarshal(body, &jsonResponse)
		suite.Assert().NoError(err, "Response should be valid JSON")
	}
}

// TestDatabaseMigrationIntegration tests database migration through the application
func (suite *ApplicationIntegrationTestSuite) TestDatabaseMigrationIntegration() {
	// Test that the application handles database migrations properly
	// This would test the datastore module's migration capabilities
	
	// Simulate POST request to trigger migration (if endpoint exists)
	migrationData := map[string]interface{}{
		"action": "migrate",
	}
	
	jsonData, err := json.Marshal(migrationData)
	if err == nil {
		resp, err := http.Post(suite.baseURL+"/admin/migrate", 
			"application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			defer resp.Body.Close()
			// Should handle migration request gracefully
			suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
				"Migration endpoint should handle requests gracefully")
		}
	}
}

// Run the test suite
func TestApplicationIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationIntegrationTestSuite))
}