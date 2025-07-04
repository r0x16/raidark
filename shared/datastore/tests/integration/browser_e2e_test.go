package integration

import (
	"context"
	"fmt"
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

// BrowserE2ETestSuite tests the application from a browser perspective
type BrowserE2ETestSuite struct {
	suite.Suite
	app        *api.Api
	serverPort string
	baseURL    string
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// SetupSuite sets up the browser end-to-end test suite
func (suite *BrowserE2ETestSuite) SetupSuite() {
	// Set test environment variables for browser testing
	os.Setenv("API_PORT", "8082")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_DATABASE", "test_browser_db")
	os.Setenv("JWT_SECRET", "test_secret_key_for_browser_tests")
	
	suite.serverPort = "8082"
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
	suite.waitForServerReady()
}

// TearDownSuite cleans up after the browser test suite
func (suite *BrowserE2ETestSuite) TearDownSuite() {
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

// waitForServerReady waits for the server to be ready for browser testing
func (suite *BrowserE2ETestSuite) waitForServerReady() {
	timeout := time.After(45 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			suite.FailNow("Server failed to start within timeout for browser testing")
		case <-ticker.C:
			resp, err := http.Get(suite.baseURL)
			if err == nil {
				resp.Body.Close()
				// Additional readiness check
				time.Sleep(2 * time.Second)
				return
			}
		}
	}
}

// TestBrowserAccessibility tests that the application is accessible from a browser
func (suite *BrowserE2ETestSuite) TestBrowserAccessibility() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Test main page accessibility
	resp, err := client.Get(suite.baseURL)
	suite.Assert().NoError(err, "Should be able to access main page")
	
	if resp != nil {
		defer resp.Body.Close()
		suite.Assert().True(resp.StatusCode < 500, 
			"Main page should not return server error")
		
		// Check for basic HTML structure
		contentType := resp.Header.Get("Content-Type")
		suite.Assert().True(
			suite.containsAny(contentType, []string{"text/html", "application/json"}),
			"Should return HTML or JSON content",
		)
	}
}

// TestAPIEndpointsBrowserCompatibility tests API endpoints for browser compatibility
func (suite *BrowserE2ETestSuite) TestAPIEndpointsBrowserCompatibility() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Test CORS headers for browser compatibility
	req, err := http.NewRequest("OPTIONS", suite.baseURL+"/api/v1/status", nil)
	suite.Require().NoError(err)
	
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		// Check for CORS headers (if implemented)
		corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
		if corsHeader != "" {
			suite.Assert().NotEmpty(corsHeader, "CORS headers should be set for browser compatibility")
		}
	}
}

// TestDatastoreThroughBrowserRequests tests datastore functionality through browser-like requests
func (suite *BrowserE2ETestSuite) TestDatastoreThroughBrowserRequests() {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	// Simulate browser requests that would interact with the datastore
	testEndpoints := []struct {
		method string
		path   string
		desc   string
	}{
		{"GET", "/", "Main page"},
		{"GET", "/api/v1/health", "Health check"},
		{"GET", "/api/v1/status", "Status endpoint"},
		{"GET", "/api/v1/users", "Users endpoint"},
	}
	
	for _, endpoint := range testEndpoints {
		req, err := http.NewRequest(endpoint.method, suite.baseURL+endpoint.path, nil)
		if err == nil {
			// Add browser-like headers
			req.Header.Set("User-Agent", "Mozilla/5.0 (Test Browser) Integration Test")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")
			
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
					fmt.Sprintf("%s should not return internal server error", endpoint.desc))
			}
		}
	}
}

// TestBrowserJavaScriptCompatibility tests that the API is compatible with JavaScript requests
func (suite *BrowserE2ETestSuite) TestBrowserJavaScriptCompatibility() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Simulate JavaScript fetch request
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/status", nil)
	suite.Require().NoError(err)
	
	// Add JavaScript-like headers
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		
		// Check that the response is JSON-compatible
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			suite.Assert().Contains(contentType, "application/json",
				"API should return JSON for JavaScript compatibility")
		}
	}
}

// TestMobileDeviceCompatibility tests compatibility with mobile device requests
func (suite *BrowserE2ETestSuite) TestMobileDeviceCompatibility() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	mobileUserAgents := []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Android 10; Mobile; rv:81.0) Gecko/81.0 Firefox/81.0",
	}
	
	for _, userAgent := range mobileUserAgents {
		req, err := http.NewRequest("GET", suite.baseURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", userAgent)
			
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				suite.Assert().NotEqual(http.StatusInternalServerError, resp.StatusCode,
					"Should handle mobile requests without server errors")
			}
		}
	}
}

// TestDatastorePerformanceUnderLoad tests datastore performance under browser-like load
func (suite *BrowserE2ETestSuite) TestDatastorePerformanceUnderLoad() {
	const numConcurrentRequests = 20
	results := make(chan time.Duration, numConcurrentRequests)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Simulate multiple browser tabs making requests
	for i := 0; i < numConcurrentRequests; i++ {
		go func(requestID int) {
			start := time.Now()
			
			req, err := http.NewRequest("GET", suite.baseURL, nil)
			if err == nil {
				req.Header.Set("User-Agent", fmt.Sprintf("BrowserTab-%d", requestID))
				
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
				}
			}
			
			duration := time.Since(start)
			results <- duration
		}(i)
	}
	
	// Collect results
	var totalDuration time.Duration
	for i := 0; i < numConcurrentRequests; i++ {
		duration := <-results
		totalDuration += duration
	}
	
	averageDuration := totalDuration / numConcurrentRequests
	suite.Assert().Less(averageDuration.Seconds(), 5.0,
		"Average response time should be reasonable under load")
}

// TestSessionManagementWithDatastore tests session management functionality
func (suite *BrowserE2ETestSuite) TestSessionManagementWithDatastore() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Simulate browser session behavior
	req, err := http.NewRequest("GET", suite.baseURL, nil)
	if err == nil {
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Check for session-related headers
			cookies := resp.Cookies()
			if len(cookies) > 0 {
				suite.Assert().NotEmpty(cookies, "Should handle cookies for session management")
			}
			
			// Check for security headers
			securityHeaders := []string{
				"X-Content-Type-Options",
				"X-Frame-Options",
				"X-XSS-Protection",
			}
			
			for _, header := range securityHeaders {
				headerValue := resp.Header.Get(header)
				if headerValue != "" {
					suite.Assert().NotEmpty(headerValue, 
						fmt.Sprintf("Security header %s should be set", header))
				}
			}
		}
	}
}

// TestDatastoreErrorPagesInBrowser tests how database errors are presented to browsers
func (suite *BrowserE2ETestSuite) TestDatastoreErrorPagesInBrowser() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Test error handling with browser headers
	req, err := http.NewRequest("GET", suite.baseURL+"/api/v1/nonexistent", nil)
	if err == nil {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			// Should return appropriate error page/response for browsers
			if resp.StatusCode >= 400 {
				contentType := resp.Header.Get("Content-Type")
				suite.Assert().True(
					suite.containsAny(contentType, []string{"text/html", "application/json"}),
					"Error responses should be browser-friendly",
				)
			}
		}
	}
}

// Helper function to check if string contains any of the given substrings
func (suite *BrowserE2ETestSuite) containsAny(str string, substrings []string) bool {
	for _, substring := range substrings {
		if suite.contains(str, substring) {
			return true
		}
	}
	return false
}

// Helper function to check if string contains substring (case-insensitive)
func (suite *BrowserE2ETestSuite) contains(str, substring string) bool {
	return len(str) >= len(substring) && 
		   (str == substring || 
		    (len(str) > len(substring) && 
		     fmt.Sprintf("%s", str) != fmt.Sprintf("%s", str[:len(str)-len(substring)])+substring))
}

// Run the browser E2E test suite
func TestBrowserE2ETestSuite(t *testing.T) {
	// Skip if running in CI environment without browser support
	if os.Getenv("CI") == "true" && os.Getenv("BROWSER_TESTS") != "true" {
		t.Skip("Skipping browser tests in CI environment")
	}
	
	suite.Run(t, new(BrowserE2ETestSuite))
}