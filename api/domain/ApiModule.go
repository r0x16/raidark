package domain

type ApiModule interface {
	Name() string
	Setup() error
}
