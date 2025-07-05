package controller

import (
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
)

type DbMigrationController struct {
	LogProvider      domlogger.LogProvider
	DatabaseProvider domdatastore.DatabaseProvider
	Modules          []apidomain.ApiModule
}

func (d *DbMigrationController) MigrateAction() error {
	modelsError := d.migrateModels()
	if modelsError != nil {
		d.LogProvider.Error("Error migrating models", map[string]any{"error": modelsError})
		return modelsError
	}
	d.LogProvider.Info("Models migrated successfully", map[string]any{"models": d.Modules})

	return nil
}

func (d *DbMigrationController) migrateModels() error {
	models := d.extractModels(d.Modules)
	d.LogProvider.Info("Migrating models", map[string]any{"models": models})
	return d.DatabaseProvider.GetDataStore().Exec.AutoMigrate(models...)
}

func (d *DbMigrationController) extractModels(modules []apidomain.ApiModule) []any {
	models := []any{}
	for _, module := range modules {
		models = append(models, module.GetModel()...)
	}
	return models
}
