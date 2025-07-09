package domain

type DomainEventsProvider interface {
	Collect()
	Publish(event DomainEvent) error
	Subscribe(handler EventListener) error
	Dispatch(event DomainEvent) error
	Close() error
}
