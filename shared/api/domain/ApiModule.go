package domain

type ApiModule interface {
	Name() string
	Setup() error
	GetModel() []any
	GetSeedData() []any
}
