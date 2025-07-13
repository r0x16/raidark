package driver

import (
	"context"
	"testing"
	"time"

	"github.com/r0x16/Raidark/shared/events/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	logdriver "github.com/r0x16/Raidark/shared/logger/driver"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type testEvent struct{ time time.Time }

func (t testEvent) Name() string          { return "test" }
func (t testEvent) OccurredAt() time.Time { return t.time }

type testListener struct {
	domain.SyncEventListener
	called bool
}

func (l *testListener) EventName() string { return "test" }
func (l *testListener) Handle(ctx context.Context, e domain.DomainEvent, hub *domprovider.ProviderHub) error {
	l.called = true
	return nil
}

func TestPublishAndSubscribe(t *testing.T) {
	hub := &domprovider.ProviderHub{}
	domprovider.Register[domlogger.LogProvider](hub, logdriver.NewStdOutLogManager())
	p := NewInMemoryDomainEventsProvider(10, 1, hub)
	l := &testListener{}
	p.Subscribe(l)
	p.Collect()
	p.Publish(testEvent{time.Now()})
	time.Sleep(50 * time.Millisecond)
	p.Close()
	if !l.called {
		t.Fatal("listener not called")
	}
}
