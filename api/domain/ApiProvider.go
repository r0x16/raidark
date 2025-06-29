package domain

type ApiProvider interface {
	Setup() error
	Register(module ApiModule)
	ProvidesModules() []ApiModule
	Run() error
}
