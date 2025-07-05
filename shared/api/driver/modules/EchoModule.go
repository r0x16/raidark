package modules

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	driverapi "github.com/r0x16/Raidark/shared/api/driver"
	domauth "github.com/r0x16/Raidark/shared/auth/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoModule struct {
	Api   domapi.ApiProvider
	Group *echo.Group
	Hub   *domprovider.ProviderHub
	Log   domlogger.LogProvider
	Auth  domauth.AuthProvider
}

type ActionCallback func(echo.Context, *domprovider.ProviderHub) error

func NewEchoModule(groupPath string, hub *domprovider.ProviderHub) *EchoModule {
	api := domprovider.Get[domapi.ApiProvider](hub)
	echoServer := api.(*driverapi.EchoApiProvider).Server
	group := echoServer.Group(groupPath)
	return &EchoModule{
		Api:   api,
		Group: group,
		Hub:   hub,
		Log:   domprovider.Get[domlogger.LogProvider](hub),
		Auth:  domprovider.Get[domauth.AuthProvider](hub),
	}
}

func NewAuthenticatedEchoModule(groupPath string, hub *domprovider.ProviderHub) *EchoModule {
	module := NewEchoModule(groupPath, hub)
	module.Group.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
		Validator: func(key string, c echo.Context) (bool, error) {
			token, err := module.Auth.ParseToken(key)
			if err != nil {
				module.Log.Error("Error parsing token", map[string]any{"error": err})
				return false, err
			}
			c.Set("user", token)
			return true, nil
		},
	}))
	return module
}

func (e *EchoModule) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	if e.Hub == nil {
		panic("Hub is not set in EchoModule")
	}

	return func(c echo.Context) error {
		return callback(c, e.Hub)
	}
}

func (e *EchoModule) GetModel() []any {
	return []any{}
}

func (e *EchoModule) GetSeedData() []any {
	return []any{}
}
