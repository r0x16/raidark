package driver

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r0x16/Raidark/shared/api/domain"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoApiProvider struct {
	modules []domain.ApiModule
	port    string

	Server *echo.Echo

	// Providers
	Log domlogger.LogProvider
	Env domenv.EnvProvider
}

var _ domain.ApiProvider = &EchoApiProvider{}

func NewEchoApiProvider(port string, hub *domprovider.ProviderHub) *EchoApiProvider {
	return &EchoApiProvider{
		modules: []domain.ApiModule{},
		port:    port,
		Server:  echo.New(),
		Log:     domprovider.Get[domlogger.LogProvider](hub),
		Env:     domprovider.Get[domenv.EnvProvider](hub),
	}
}

// Setup implements domain.ApiProvider.
func (e *EchoApiProvider) Setup() error {
	e.Server.Use(middleware.Recover())

	// Configure CORS middleware with environment variables
	e.configureCORS()

	// Configure CSRF middleware with environment variables
	e.configureCSRF()

	e.Server.Pre(middleware.RemoveTrailingSlash())

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
