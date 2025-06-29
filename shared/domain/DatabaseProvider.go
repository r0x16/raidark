package domain

type DatabaseProvider interface {
	Connect() error
	Close() error
}
