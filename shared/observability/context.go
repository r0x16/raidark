package observability

import (
	"context"
	"sync/atomic"
)

// Echo context keys (read by handlers via echo.Context.Get) and Go context keys
// (used by FromContext-style helpers) are kept in sync so the same value is
// reachable through either path.
const (
	// EchoTraceIDKey is the echo.Context key under which W3CTrace stores the trace_id.
	EchoTraceIDKey = "trace_id"
	// EchoSpanIDKey is the echo.Context key under which W3CTrace stores the span_id.
	EchoSpanIDKey = "span_id"
	// EchoTraceFlagsKey is the echo.Context key under which W3CTrace stores the
	// trace flags (the third W3C component, two hex chars).
	EchoTraceFlagsKey = "trace_flags"
	// EchoTraceStateKey is the echo.Context key under which W3CTrace stores the
	// raw incoming tracestate header value.
	EchoTraceStateKey = "trace_state"
)

// Distinct unexported types prevent accidental collisions with other packages
// that store values in the same context.Context.
type (
	traceIDCtxKey    struct{}
	spanIDCtxKey     struct{}
	traceFlagsCtxKey struct{}
	traceStateCtxKey struct{}
	serviceCtxKey    struct{}
	eventIDCtxKey    struct{}
)

// defaultServiceName is the fallback service name returned by GetServiceName
// when the context does not carry one. It is set once at boot via
// SetDefaultServiceName so library code can emit "service" in logs without
// each call site having to thread the name through context.
var defaultServiceName atomic.Value // string

// SetDefaultServiceName registers the process-wide service name used by
// GetServiceName when the request context has not been augmented with one.
// Safe to call concurrently; the most recent call wins.
func SetDefaultServiceName(name string) {
	defaultServiceName.Store(name)
}

// GetDefaultServiceName returns the process-wide service name registered with
// SetDefaultServiceName, or "" if none has been configured yet.
func GetDefaultServiceName() string {
	if v, ok := defaultServiceName.Load().(string); ok {
		return v
	}
	return ""
}

// WithTraceID returns ctx augmented with traceID. The value is also reachable
// from echo.Context via EchoTraceIDKey when set by the W3CTrace middleware.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtxKey{}, traceID)
}

// GetTraceID returns the trace ID stored in ctx, or "" if absent.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(traceIDCtxKey{}).(string)
	return v
}

// WithSpanID returns ctx augmented with spanID.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDCtxKey{}, spanID)
}

// GetSpanID returns the span ID stored in ctx, or "" if absent.
func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(spanIDCtxKey{}).(string)
	return v
}

// WithTraceFlags returns ctx augmented with the W3C trace flags hex byte.
func WithTraceFlags(ctx context.Context, flags string) context.Context {
	return context.WithValue(ctx, traceFlagsCtxKey{}, flags)
}

// GetTraceFlags returns the trace flags stored in ctx, or "" if absent.
func GetTraceFlags(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(traceFlagsCtxKey{}).(string)
	return v
}

// WithTraceState returns ctx augmented with the raw tracestate header value.
func WithTraceState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, traceStateCtxKey{}, state)
}

// GetTraceState returns the tracestate stored in ctx, or "" if absent.
func GetTraceState(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(traceStateCtxKey{}).(string)
	return v
}

// WithServiceName returns ctx augmented with the service name. Per-context
// values take precedence over SetDefaultServiceName.
func WithServiceName(ctx context.Context, service string) context.Context {
	return context.WithValue(ctx, serviceCtxKey{}, service)
}

// GetServiceName returns the service name stored in ctx, falling back to the
// process-wide default registered with SetDefaultServiceName.
func GetServiceName(ctx context.Context) string {
	if ctx != nil {
		if v, ok := ctx.Value(serviceCtxKey{}).(string); ok && v != "" {
			return v
		}
	}
	return GetDefaultServiceName()
}

// WithEventID returns ctx augmented with the current event_id. Used by event
// consumers and publishers so subsequent log lines can be correlated to the
// envelope being processed.
func WithEventID(ctx context.Context, eventID string) context.Context {
	return context.WithValue(ctx, eventIDCtxKey{}, eventID)
}

// GetEventID returns the event_id stored in ctx, or "" if not in an event flow.
func GetEventID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(eventIDCtxKey{}).(string)
	return v
}
