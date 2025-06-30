package modules

import (
	"github.com/r0x16/Raidark/api/auth/controller"
	"github.com/r0x16/Raidark/api/domain"
	"github.com/r0x16/Raidark/api/drivers"
)

type EchoAuthModule struct {
	Api *drivers.EchoApiProvider
}

var _ domain.ApiModule = &EchoAuthModule{}

// Name implements domain.ApiModule.
func (e *EchoAuthModule) Name() string {
	return "Auth"
}

// Setup implements domain.ApiModule.
func (e *EchoAuthModule) Setup() error {

	auth := e.Api.Server.Group("/auth")

	auth.POST("/exchange", e.Api.Bundle.ActionInjection(controller.ExchangeAction))
	auth.POST("/refresh", e.Api.Bundle.ActionInjection(controller.RefreshAction))
	auth.POST("/logout", e.Api.Bundle.ActionInjection(controller.LogoutAction))

	return nil
}
