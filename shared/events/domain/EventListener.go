package domain

import (
	"context"

	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// EventListener is the interface that wraps the Handle method.
// It is used to handle events.
// The Handle method is called when an event is published.
// The IsAsync method is used to determine if the event handler is asynchronous.
// If the event handler is asynchronous, the Handle method is called in a separate goroutine.
// If the event handler is synchronous, the Handle method is called in the same goroutine.
type EventListener interface {
	EventName() string
	Handle(context.Context, DomainEvent, *domprovider.ProviderHub) error
	IsAsync() bool
}

// SyncEventListener is a synchronous event handler.
// It is used to handle events synchronously.
// The Handle method is called in the same goroutine.
type SyncEventListener struct{}

func (h *SyncEventListener) IsAsync() bool {
	return false
}

// AsyncEventListener is an asynchronous event handler.
// It is used to handle events asynchronously.
// The Handle method is called in a separate goroutine.
type AsyncEventListener struct{}

func (h *AsyncEventListener) IsAsync() bool {
	return true
}
