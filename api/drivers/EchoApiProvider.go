package drivers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/r0x16/Raidark/api/domain"
)

type EchoApiProvider struct {
	modules []domain.ApiModule
	port    string

	Server *echo.Echo
	Bundle *ApplicationBundle
}

var _ domain.ApiProvider = &EchoApiProvider{}

func NewEchoApiProvider(port string, bundle *ApplicationBundle) *EchoApiProvider {
	return &EchoApiProvider{
		modules: []domain.ApiModule{},
		port:    port,
		Server:  echo.New(),
		Bundle:  bundle,
	}
}

// Boot implements domain.ApiProvider.
func (e *EchoApiProvider) Setup() error {
	e.Server.Use(middleware.Recover())

	// TODO: Define correct Headers and Origins CORS settings
	e.Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowHeaders: []string{"*"},
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	e.Server.Pre(middleware.RemoveTrailingSlash())

	return nil
}

// Run implements domain.ApiProvider.
func (e *EchoApiProvider) Run() error {
	if _, err := strconv.Atoi(e.port); err != nil {
		e.Bundle.Log.Critical("Invalid port number, must to be a number", map[string]any{
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
