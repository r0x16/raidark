// Package observability verifies the context helpers that carry trace,
// service, and event metadata outside HTTP-specific middleware.
package observability

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextGetters_ReturnEmptyForNilContext(t *testing.T) {
	assert.Empty(t, GetTraceID(nil))
	assert.Empty(t, GetSpanID(nil))
	assert.Empty(t, GetTraceFlags(nil))
	assert.Empty(t, GetTraceState(nil))
	assert.Empty(t, GetEventID(nil))
}

func TestDefaultServiceName_ReturnsEmptyBeforeConfiguration(t *testing.T) {
	resetDefaultServiceName(t)

	assert.Empty(t, GetDefaultServiceName())
}

func TestServiceName_UsesContextBeforeDefault(t *testing.T) {
	SetDefaultServiceName("fallback-service")

	assert.Equal(t, "fallback-service", GetDefaultServiceName())
	assert.Equal(t, "fallback-service", GetServiceName(context.Background()))
	assert.Equal(t, "handler-service", GetServiceName(WithServiceName(context.Background(), "handler-service")))
}

func TestEventID_RoundTripThroughContext(t *testing.T) {
	ctx := WithEventID(context.Background(), "evt-123")

	assert.Equal(t, "evt-123", GetEventID(ctx))
}

func resetDefaultServiceName(t *testing.T) {
	t.Helper()

	previous := GetDefaultServiceName()
	defaultServiceName = atomic.Value{}
	t.Cleanup(func() {
		defaultServiceName = atomic.Value{}
		if previous != "" {
			SetDefaultServiceName(previous)
		}
	})
}
