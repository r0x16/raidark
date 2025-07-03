package domain

type ProviderFactory interface {
	Init(*ProviderHub)
	Register(*ProviderHub) error
}
