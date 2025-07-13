package domain

type DatabaseProvider interface {
	Connect() error
	Close() error
	GetTransaction() Transaction

	// Deprecated: Use GetTransaction() instead
	GetDataStore() *DataStore
}
