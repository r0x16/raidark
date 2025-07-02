package domain

import "github.com/r0x16/Raidark/shared/domain/model/db"

type DatabaseProvider interface {
	Connect() error
	Close() error
	GetDataStore() *db.DataStore
}
