package controller

import (
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
)

type SeederController struct {
	LogProvider      domlogger.LogProvider
	DatabaseProvider domdatastore.DatabaseProvider
	Modules          []apidomain.ApiModule
}

func (c *SeederController) SeedAction() error {
	seedDataError := c.seedData()
	if seedDataError != nil {
		c.LogProvider.Error("Error seeding data", map[string]any{"error": seedDataError})
		return seedDataError
	}
	c.LogProvider.Info("Data seeded successfully", map[string]any{"modules": c.Modules})

	return nil
}

func (c *SeederController) seedData() error {
	db := c.DatabaseProvider.GetDataStore().Exec
	tx := db.Begin()

	// Extract seed data from all modules
	seedDataSets := c.extractSeedData(c.Modules)
	c.LogProvider.Info("Seeding data", map[string]any{"dataSets": len(seedDataSets)})

	// Insert each seed data set
	for _, seedDataSet := range seedDataSets {
		if err := tx.Create(seedDataSet).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (c *SeederController) extractSeedData(modules []apidomain.ApiModule) []any {
	seedData := []any{}
	for _, module := range modules {
		seedData = append(seedData, module.GetSeedData()...)
	}
	return seedData
}

// Example of seeding
/* func (c *SeederController) seedTimeUnit(tx *gorm.DB) error {
	timeUnits := seed.TimeUnitList
	numericTimeUnits := seed.TimeUnitNumericList

	repo := repository.NewGormTimeUnitRepository(tx)

	for key, value := range timeUnits {
		err := repo.Store(&model.TimeUnit{
			Code:        key,
			NumericCode: numericTimeUnits[key],
			Name:        value,
		})

		if err != nil {
			return err
		}
	}
	return nil
} */
