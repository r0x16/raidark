// Package observability provides the cross-cutting telemetry primitives used
// by every service built on top of Raidark: structured logging with
// automatic correlation fields, Prometheus metrics, and W3C trace-context
// propagation.
//
// The package is intentionally framework-light. Public middlewares plug
// into Echo, but trace/log helpers are pure context.Context utilities and
// can be reused from any background worker, CLI command, or event consumer
// that does not run inside an HTTP request.
package observability

import (
	"context"
	"sync/atomic"
)

// Context keys stored in echo.Context (and any other request-scoped
// container) when middlewares promote trace metadata into the per-request
// surface. The constants are framework-agnostic — they are plain string
// keys, intended to work for any web layer that exposes a get/set bag,
// not just Echo. Code that runs outside an HTTP server should read the
// values from the Go context via the GetTraceID / GetSpanID helpers
// instead.
const (
	// ContextTraceIDKey is the key under which W3CTrace stores the trace_id.
	ContextTraceIDKey = "trace_id"
	// ContextSpanIDKey is the key under which W3CTrace stores the span_id.
	ContextSpanIDKey = "span_id"
	// ContextTraceFlagsKey is the key under which W3CTrace stores the
	// trace flags (the third W3C component, two hex chars).
	ContextTraceFlagsKey = "trace_flags"
	// ContextTraceStateKey is the key under which W3CTrace stores the
	// raw incoming tracestate header value.
	ContextTraceStateKey = "trace_state"
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
// from echo.Context via ContextTraceIDKey when set by the W3CTrace middleware.
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
