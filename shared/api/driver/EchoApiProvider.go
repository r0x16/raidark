package driver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/api/rest"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/observability"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoApiProvider struct {
	modules []domain.ApiModule
	port    string

	Server *echo.Echo

	// Providers
	Log     domlogger.LogProvider
	Env     domenv.EnvProvider
	Metrics obsdomain.MetricsProvider
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
	// MetricsProvider is optional: services that don't register a
	// MetricsProviderFactory in main.go simply skip the HTTPMetrics
	// middleware. The /metrics route itself is registered by the
	// EchoMetricsModule which performs the same nil check.
	if domprovider.Exists[obsdomain.MetricsProvider](hub) {
		provider.Metrics = domprovider.Get[obsdomain.MetricsProvider](hub)
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

	return nil
}

// configureCORS mounts the Echo CORS middleware only when CORS_ALLOW_ORIGINS is explicitly
// set in the environment. Omitting the variable is a deliberate opt-out: Raidark will not
// default to a wildcard policy because wildcard CORS plus credentials is insecure and
// services behind a BFF typically need no CORS at all.
//
// Empty entries in the comma-separated list are silently dropped; if every entry is empty
// the middleware is not mounted and the boot log records cors=disabled.
func (e *EchoApiProvider) configureCORS() {
	if !e.Env.IsSet("CORS_ALLOW_ORIGINS") {
		e.Log.Info("Bootstrap: CORS middleware not mounted", map[string]any{
			"cors": "disabled",
		})
		return
	}

	rawOrigins := e.Env.GetSlice("CORS_ALLOW_ORIGINS", nil)
	allowOrigins := make([]string, 0, len(rawOrigins))
	for _, o := range rawOrigins {
		if o != "" {
			allowOrigins = append(allowOrigins, o)
		}
	}

	if len(allowOrigins) == 0 {
		e.Log.Info("Bootstrap: CORS middleware not mounted (CORS_ALLOW_ORIGINS is empty after filtering)", map[string]any{
			"cors": "disabled",
		})
		return
	}

	allowHeaders := e.Env.GetSlice("CORS_ALLOW_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"})
	allowMethods := e.Env.GetSlice("CORS_ALLOW_METHODS", []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead})
	allowCredentials := e.Env.GetBool("CORS_ALLOW_CREDENTIALS", false)
	maxAge := e.Env.GetInt("CORS_MAX_AGE", 0)

	corsConfig := middleware.CORSConfig{
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins:     allowOrigins,
		AllowHeaders:     allowHeaders,
		AllowMethods:     allowMethods,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	}

	e.Server.Use(middleware.CORSWithConfig(corsConfig))

	e.Log.Info("Bootstrap: CORS middleware configured", map[string]any{
		"cors":              strings.Join(allowOrigins, ", "),
		"allow_headers":     strings.Join(allowHeaders, ", "),
		"allow_methods":     strings.Join(allowMethods, ", "),
		"allow_credentials": allowCredentials,
		"max_age":           maxAge,
	})
}

// configureCSRF mounts the Echo CSRF middleware only when CSRF_ENABLED=true. The default
// is disabled because services behind a BFF that already enforces CSRF should not add a
// second, contradictory protection layer. When disabled, the /csrf-token route is also
// not registered (see EchoMainModule).
func (e *EchoApiProvider) configureCSRF() {
	csrfEnabled := e.Env.GetBool("CSRF_ENABLED", false)

	if !csrfEnabled {
		e.Log.Info("Bootstrap: CSRF middleware not mounted", map[string]any{
			"csrf": "disabled",
		})
		return
	}

	tokenLength := e.Env.GetInt("CSRF_TOKEN_LENGTH", 32)
	cookieName := e.Env.GetString("CSRF_COOKIE_NAME", "_csrf")
	// CSRF_COOKIE_SECURE should be true in production (HTTPS). Defaults to false for local dev.
	cookieSecure := e.Env.GetBool("CSRF_COOKIE_SECURE", false)
	tokenLookup := e.Env.GetString("CSRF_TOKEN_LOOKUP", "cookie:_csrf")
	cookieMaxAge := e.Env.GetInt("CSRF_COOKIE_MAX_AGE", 86400)

	csrfConfig := middleware.CSRFConfig{
		Skipper:        middleware.DefaultSkipper,
		TokenLength:    uint8(tokenLength),
		TokenLookup:    tokenLookup,
		ContextKey:     "csrf",
		CookieName:     cookieName,
		CookieSecure:   cookieSecure,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		CookieMaxAge:   cookieMaxAge,
	}

	e.Server.Use(middleware.CSRFWithConfig(csrfConfig))

	e.Log.Info("Bootstrap: CSRF middleware configured", map[string]any{
		"csrf":           "enabled",
		"token_length":   tokenLength,
		"cookie_name":    cookieName,
		"token_lookup":   tokenLookup,
		"cookie_max_age": cookieMaxAge,
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
