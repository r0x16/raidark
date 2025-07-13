package driver

import (
	"errors"

	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	driverdatastore "github.com/r0x16/Raidark/shared/datastore/driver"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type DatastoreProviderFactory struct {
	env domenv.EnvProvider
}

func (f *DatastoreProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

/*
* Register the datastore provider to the provider hub

* Supported databases: (default: sqlite)
* - Postgres
* - MySQL
* - SQLite

* Supported execution engines: (default: gorm)
* - Gorm
 */
func (f *DatastoreProviderFactory) Register(hub *domain.ProviderHub) error {
	dbtype := f.env.GetString("DATASTORE_TYPE", "sqlite")
	provider, err := f.getProvider(dbtype)

	if err != nil {
		return err
	}

	domain.Register(hub, provider)

	return nil
}

/*
*

	Get the datastore provider based on the database type
*/
func (f *DatastoreProviderFactory) getProvider(dbtype string) (domdatastore.DatabaseProvider, error) {
	switch dbtype {
	case "postgres":
		return f.providesPostgres()
	case "mysql":
		return f.providesMysql()
	case "sqlite":
		return f.providesSqlite()
	}
	return nil, errors.New("invalid database type: " + dbtype)
}

/*
*

	Get the postgres provider
*/
func (f *DatastoreProviderFactory) providesPostgres() (domdatastore.DatabaseProvider, error) {
	connection := driverdatastore.NewGormPostgresDatabaseProvider(f.env)
	err := connection.Connect()

	if err != nil {
		return nil, err
	}

	return connection, nil
}

/*
*

	Get the mysql provider
*/
func (f *DatastoreProviderFactory) providesMysql() (domdatastore.DatabaseProvider, error) {
	connection := driverdatastore.NewGormMysqlDatabaseProvider(f.env)
	err := connection.Connect()

	if err != nil {
		return nil, err
	}

	return connection, nil
}

/*
*

	Get the sqlite provider
*/
func (f *DatastoreProviderFactory) providesSqlite() (domdatastore.DatabaseProvider, error) {
	connection := driverdatastore.NewGormSqliteDatabaseProvider(f.env)
	err := connection.Connect()

	if err != nil {
		return nil, err
	}

	return connection, nil
}
