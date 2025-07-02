package controller

import (
	"github.com/r0x16/Raidark/api/drivers"
)

type SeederController struct {
	*drivers.ApplicationBundle
}

func (c *SeederController) SeedAction() error {

	db := c.Database.GetDataStore().Exec
	tx := db.Begin()

	// TODO: Example of seeding
	/* err = c.seedTimeUnit(tx)
	if err != nil {
		tx.Rollback()
		return err
	} */

	tx.Commit()

	return nil
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
