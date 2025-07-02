package modules

import (
	"github.com/r0x16/Raidark/api/domain"
	"github.com/r0x16/Raidark/shared/auth/driver/controller"
)

type EchoAuthModule struct {
	EchoModule
}

var _ domain.ApiModule = &EchoAuthModule{}

// Name implements domain.ApiModule.
func (e *EchoAuthModule) Name() string {
	return "Auth"
}

// Setup implements domain.ApiModule.
func (e *EchoAuthModule) Setup() error {

	e.Group.POST("/exchange", e.Api.Bundle.ActionInjection(controller.ExchangeAction))
	e.Group.POST("/refresh", e.Api.Bundle.ActionInjection(controller.RefreshAction))
	e.Group.POST("/logout", e.Api.Bundle.ActionInjection(controller.LogoutAction))

	return nil
}
