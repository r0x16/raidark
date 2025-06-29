package services

import (
	"github.com/r0x16/Raidark/api/domain"
	"github.com/r0x16/Raidark/shared/domain/logger"
)

type ApiService struct {
	api     domain.ApiProvider
	modules []domain.ApiModule
	log     logger.LogProvider
}

func NewApiService(api domain.ApiProvider, log logger.LogProvider) *ApiService {
	modules := api.ProvidesModules()
	return &ApiService{
		api,
		modules,
		log,
	}
}

func (as *ApiService) Run() error {

	as.api.Setup()

	for _, module := range as.modules {
		err := module.Setup()
		if err != nil {
			as.log.Error("Cannot setup module", map[string]any{
				"name":   module.Name(),
				"module": module,
			})
			return err
		}
	}

	return as.api.Run()
}
