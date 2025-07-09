package driver

import (
	"context"
	"sync"

	"github.com/r0x16/Raidark/shared/events/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type InMemoryDomainEventsProvider struct {
	queue           chan domain.DomainEvent
	subscribers     map[string][]domain.EventListener
	syncSubscribers map[string][]domain.EventListener
	mu              sync.RWMutex
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	workers         int
	hub             *domprovider.ProviderHub
	LogProvider     domlogger.LogProvider
}

var _ domain.DomainEventsProvider = &InMemoryDomainEventsProvider{}

func NewInMemoryDomainEventsProvider(bufferSize int, workers int, hub *domprovider.ProviderHub) *InMemoryDomainEventsProvider {
	ctx, cancel := context.WithCancel(context.Background())
	return &InMemoryDomainEventsProvider{
		queue:           make(chan domain.DomainEvent, bufferSize),
		subscribers:     make(map[string][]domain.EventListener),
		syncSubscribers: make(map[string][]domain.EventListener),
		mu:              sync.RWMutex{},
		wg:              sync.WaitGroup{},
		ctx:             ctx,
		cancel:          cancel,
		workers:         workers,
		hub:             hub,
		LogProvider:     domprovider.Get[domlogger.LogProvider](hub),
	}
}

func (p *InMemoryDomainEventsProvider) Collect() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case event := <-p.queue:
					p.Dispatch(event)
				case <-p.ctx.Done():
					p.LogProvider.Warning("collect worker stopped", map[string]any{"worker": i})
					return
				}
			}
		}()
	}
}

func (p *InMemoryDomainEventsProvider) Publish(event domain.DomainEvent) error {
	p.mu.RLock()
	handlers, ok := p.syncSubscribers[event.Name()]
	p.mu.RUnlock()
	if ok {
		for _, handler := range handlers {
			err := handler.Handle(context.Background(), event, p.hub)
			if err != nil {
				p.LogProvider.Error("error dispatching event for handler", map[string]any{
					"event":   event,
					"handler": handler,
					"error":   err,
				})
			}
		}
	}

	select {
	case p.queue <- event:
	default:
		go func(e domain.DomainEvent) {
			p.LogProvider.Warning("queue is full waiting for a slot", map[string]any{"event": e})
			p.queue <- e
		}(event)
	}
	return nil
}

func (p *InMemoryDomainEventsProvider) Subscribe(handler domain.EventListener) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	eventName := handler.EventName()
	p.LogProvider.Debug("subscribing to event", map[string]any{"event": eventName, "handler": handler})

	if handler.IsAsync() {
		p.subscribers[eventName] = append(p.subscribers[eventName], handler)
		return nil
	}

	p.syncSubscribers[eventName] = append(p.syncSubscribers[eventName], handler)
	return nil
}

func (p *InMemoryDomainEventsProvider) Dispatch(event domain.DomainEvent) error {
	p.mu.RLock()
	handlers, ok := p.subscribers[event.Name()]
	p.mu.RUnlock()
	if !ok {
		return nil
	}

	for _, handler := range handlers {
		go func(h domain.EventListener) {
			err := h.Handle(context.Background(), event, p.hub)
			if err != nil {
				p.LogProvider.Error("error dispatching event for handler", map[string]any{
					"event":   event,
					"handler": h,
					"error":   err,
				})
			}
		}(handler)
	}

	return nil
}

func (p *InMemoryDomainEventsProvider) Close() error {
	p.cancel()
	p.wg.Wait()
	return nil
}
