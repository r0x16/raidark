package driver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/api/rest"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/observability"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoApiProvider struct {
	modules []domain.ApiModule
	port    string

	Server *echo.Echo

	// Providers
	Log     domlogger.LogProvider
	Env     domenv.EnvProvider
	Metrics observability.MetricsProvider
}

var _ domain.ApiProvider = &EchoApiProvider{}

func NewEchoApiProvider(port string, hub *domprovider.ProviderHub) *EchoApiProvider {
	provider := &EchoApiProvider{
		modules: []domain.ApiModule{},
		port:    port,
		Server:  echo.New(),
		Log:     domprovider.Get[domlogger.LogProvider](hub),
		Env:     domprovider.Get[domenv.EnvProvider](hub),
	}
	// MetricsProvider is optional: services that pre-date RDK-003 may not
	// register one, and we still want the API to come up. Probe the hub
	// instead of unconditionally fetching.
	if domprovider.Exists[observability.MetricsProvider](hub) {
		provider.Metrics = domprovider.Get[observability.MetricsProvider](hub)
	}
	return provider
}

// Setup implements domain.ApiProvider.
func (e *EchoApiProvider) Setup() error {
	// Replace Echo's default error handler with the Raidark REST envelope handler.
	// All unhandled errors from handlers will be converted to {"error": {...}} JSON.
	e.Server.HTTPErrorHandler = rest.EchoErrorHandler

	e.Server.Use(middleware.Recover())

	// CorrelationID must be registered before CORS so that every request —
	// including OPTIONS preflight — carries a trace ID from the earliest point.
	e.Server.Use(rest.CorrelationID())

	// W3CTrace runs after CorrelationID so it can promote a UUIDv7
	// correlation-ID into a W3C trace_id when the caller hasn't sent
	// traceparent. Order matters: tracing depends on the correlation-id
	// header already being on the request.
	e.Server.Use(observability.W3CTrace())

	// HTTPMetrics is registered before CORS so it observes every response
	// the server emits, including CORS preflight rejections — those are
	// genuine traffic and oncall wants them in the dashboards.
	if e.Metrics != nil {
		e.Server.Use(observability.HTTPMetrics(e.Metrics.Metrics()))
	}

	// Configure CORS middleware with environment variables
	e.configureCORS()

	// Configure CSRF middleware with environment variables
	e.configureCSRF()

	e.Server.Pre(middleware.RemoveTrailingSlash())

	// /metrics is mounted last in Setup so it sits after every middleware
	// has been wired up. promhttp ignores middlewares anyway (the handler
	// is registered directly on the router) but mounting late keeps the
	// route definition next to the rest of the operational surface area.
	if e.Metrics != nil && e.Metrics.Enabled() {
		e.Metrics.MountScrapeEndpoint(e.Server, e.Env.GetString("METRICS_PATH", "/metrics"))
	}

	return nil
}

// configureCORS sets up CORS middleware using environment variables
func (e *EchoApiProvider) configureCORS() {
	allowOrigins := e.Env.GetSlice("CORS_ALLOW_ORIGINS", []string{"*"})
	allowHeaders := e.Env.GetSlice("CORS_ALLOW_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"})
	allowMethods := e.Env.GetSlice("CORS_ALLOW_METHODS", []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead})
	allowCredentials := e.Env.GetBool("CORS_ALLOW_CREDENTIALS", false)

	corsConfig := middleware.CORSConfig{
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins:     allowOrigins,
		AllowHeaders:     allowHeaders,
		AllowMethods:     allowMethods,
		AllowCredentials: allowCredentials,
	}

	e.Server.Use(middleware.CORSWithConfig(corsConfig))

	e.Log.Info("CORS middleware configured", map[string]any{
		"allow_origins":     allowOrigins,
		"allow_headers":     allowHeaders,
		"allow_methods":     allowMethods,
		"allow_credentials": allowCredentials,
	})
}

// configureCSRF sets up CSRF middleware using environment variables
func (e *EchoApiProvider) configureCSRF() {
	csrfEnabled := e.Env.GetBool("CSRF_ENABLED", false)

	if !csrfEnabled {
		e.Log.Info("CSRF middleware disabled by configuration", nil)
		return
	}

	tokenLength := e.Env.GetInt("CSRF_TOKEN_LENGTH", 32)
	cookieName := e.Env.GetString("CSRF_COOKIE_NAME", "_csrf")
	cookieSecure := e.Env.GetBool("CSRF_COOKIE_SECURE", false)
	tokenLookup := e.Env.GetString("CSRF_TOKEN_LOOKUP", "cookie:_csrf")
	cookieMaxAge := e.Env.GetInt("CSRF_COOKIE_MAX_AGE", 86400)

	csrfConfig := middleware.CSRFConfig{
		Skipper:        middleware.DefaultSkipper,
		TokenLength:    uint8(tokenLength),
		TokenLookup:    tokenLookup,
		ContextKey:     "csrf",
		CookieName:     cookieName,
		CookieSecure:   cookieSecure, // Set to true in production with HTTPS
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		CookieMaxAge:   cookieMaxAge,
	}

	e.Server.Use(middleware.CSRFWithConfig(csrfConfig))

	e.Log.Info("CSRF middleware configured", map[string]any{
		"token_length": tokenLength,
		"cookie_name":  cookieName,
		"enabled":      true,
	})
}

// Run implements domain.ApiProvider.
func (e *EchoApiProvider) Run() error {
	if _, err := strconv.Atoi(e.port); err != nil {
		e.Log.Critical("Invalid port number, must to be a number", map[string]any{
			"port":  e.port,
			"error": err.Error(),
		})
		return err
	}

	// Print registered routes
	fmt.Println("\nRegistered routes:")
	for _, r := range e.Server.Routes() {
		if r.Method == echo.RouteNotFound {
			continue
		}
		fmt.Printf("\n%-6s %s", r.Method, r.Path)
	}

	return e.Server.Start(":" + e.port)
}

// ProvidesModules implements domain.ApiProvider.
func (e *EchoApiProvider) ProvidesModules() []domain.ApiModule {
	return e.modules
}

// Register implements domain.ApiProvider.
func (e *EchoApiProvider) Register(module domain.ApiModule) {
	e.modules = append(e.modules, module)
}
