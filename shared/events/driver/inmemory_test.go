package driver

import (
	"context"
	"testing"
	"time"

	domain "github.com/r0x16/Raidark/shared/events/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type testEvent struct{}

func (e *testEvent) Name() string          { return "test" }
func (e *testEvent) OccurredAt() time.Time { return time.Now() }

type testListener struct{ domain.AsyncEventListener }

func (l *testListener) EventName() string { return "test" }
func (l *testListener) Handle(ctx context.Context, ev domain.DomainEvent, hub *domprovider.ProviderHub) error {
	return nil
}

type dummyLogger struct{ domlogger.LogProvider }

func (dummyLogger) Debug(string, map[string]any)    {}
func (dummyLogger) Info(string, map[string]any)     {}
func (dummyLogger) Warning(string, map[string]any)  {}
func (dummyLogger) Error(string, map[string]any)    {}
func (dummyLogger) Critical(string, map[string]any) {}
func (dummyLogger) SetLogLevel(domlogger.LogLevel)  {}

func TestPublishAndDispatch(t *testing.T) {
	hub := &domprovider.ProviderHub{}
	domprovider.Register[domlogger.LogProvider](hub, dummyLogger{})
	p := NewInMemoryDomainEventsProvider(10, 1, hub)
	listener := &testListener{}
	p.Subscribe(listener)
	p.Publish(&testEvent{})
	p.Close()
}
