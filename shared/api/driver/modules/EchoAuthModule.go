package modules

import (
	"github.com/r0x16/Raidark/shared/api/domain"
	modelauth "github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/driver/controller"
)

type EchoAuthModule struct {
	*EchoModule
}

var _ domain.ApiModule = &EchoAuthModule{}

// Name implements domain.ApiModule.
func (e *EchoAuthModule) Name() string {
	return "Auth"
}

// Setup implements domain.ApiModule.
func (e *EchoAuthModule) Setup() error {

	e.Group.POST("/exchange", e.ActionInjection(controller.ExchangeAction))
	e.Group.POST("/refresh", e.ActionInjection(controller.RefreshAction))
	e.Group.POST("/logout", e.ActionInjection(controller.LogoutAction))

	return nil
}

func (e *EchoAuthModule) GetModel() []any {
	return []any{
		&modelauth.AuthSession{},
	}
}
