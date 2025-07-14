package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	"gorm.io/gorm"
)

type GormRepository struct{}

func GetTransactionExec(tx domain.Transaction) *gorm.DB {
	gormtx := tx.(*GormTransaction)

	gormtx.TransactionMutex.Lock()
	defer gormtx.TransactionMutex.Unlock()

	if gormtx.IsInTransaction {
		return gormtx.tx
	}
	return gormtx.db
}
