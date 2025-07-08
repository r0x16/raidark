package domain

type DomainEventsProvider interface {
	Collect()
	Publish(event DomainEvent) error
	Subscribe(eventName string, handler EventHandler) error
	Dispatch(event DomainEvent) error
	Close() error
}
