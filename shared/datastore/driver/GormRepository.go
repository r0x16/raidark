package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domproviders "github.com/r0x16/Raidark/shared/providers/domain"
	"gorm.io/gorm"
)

type GormRepository struct {
	dbProvider  domain.DatabaseProvider
	envProvider domenv.EnvProvider
}

func NewGormRepository(hub *domproviders.ProviderHub) *GormRepository {
	return &GormRepository{
		dbProvider:  domproviders.Get[domain.DatabaseProvider](hub),
		envProvider: domproviders.Get[domenv.EnvProvider](hub),
	}
}

func (r *GormRepository) GetExec() *gorm.DB {
	switch r.envProvider.GetString("DATASTORE_TYPE", "sqlite") {
	case "postgres":
		gormProvider := r.dbProvider.(*GormPostgresDatabaseProvider)
		return gormProvider.db
	case "mysql":
		gormProvider := r.dbProvider.(*GormMysqlDatabaseProvider)
		return gormProvider.db
	case "sqlite":
		gormProvider := r.dbProvider.(*GormSqliteDatabaseProvider)
		return gormProvider.db
	default:
		panic("invalid datastore type")
	}
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
