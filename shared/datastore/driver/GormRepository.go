package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	"gorm.io/gorm"
)

type GormRepository struct {
	dbProvider domain.DatabaseProvider
}

func (r *GormRepository) GetExec() *gorm.DB {
	gormProvider := r.dbProvider.(*GormPostgresDatabaseProvider)
	return gormProvider.db
}

func (r *GormRepository) GetTransactionExec(tx domain.Transaction) *gorm.DB {
	gormtx := tx.(*GormTransaction)

	gormtx.TransactionMutex.Lock()
	defer gormtx.TransactionMutex.Unlock()

	if gormtx.IsInTransaction {
		return gormtx.tx
	}
	return gormtx.db
}
