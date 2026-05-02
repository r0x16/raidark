package observability

import "context"

// HeaderCarrier is the minimal contract observability needs to inject and
// extract W3C trace-context headers from any transport: HTTP, NATS, Kafka,
// or in-memory test maps. It is deliberately tiny so adapters can be written
// for new transports without pulling in this package's dependencies.
//
// Implementations MUST treat header names case-insensitively when reading
// (per the HTTP spec) and SHOULD preserve the case the writer used (W3C
// recommends lowercase, which this package always emits).
type HeaderCarrier interface {
	// Get returns the header value for key, or "" if unset.
	Get(key string) string
	// Set replaces the value for key.
	Set(key, value string)
}

// MapCarrier adapts a plain map[string]string to HeaderCarrier. It is the
// natural carrier for NATS message headers, Kafka headers represented as a
// map, and unit tests that don't want to spin up a full http.Header.
type MapCarrier map[string]string

// Get implements HeaderCarrier.
func (c MapCarrier) Get(key string) string { return c[key] }

// Set implements HeaderCarrier.
func (c MapCarrier) Set(key, value string) { c[key] = value }

// InjectTrace writes the trace-context fields stored in ctx into carrier
// under the canonical W3C header names. If ctx has no trace_id (i.e. the
// caller never went through a trace-aware middleware), the call is a no-op
// — we never invent a trace_id at injection time because that would create
// orphan traces that can't be joined to the originating request.
func InjectTrace(ctx context.Context, carrier HeaderCarrier) {
	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)
	if traceID == "" || spanID == "" {
		return
	}
	flags := GetTraceFlags(ctx)
	if flags == "" {
		flags = defaultTraceFlags
	}
	carrier.Set(TraceParentHeader, formatTraceParent(TraceContext{
		Version: supportedVersion,
		TraceID: traceID,
		SpanID:  spanID,
		Flags:   flags,
	}))
	if state := GetTraceState(ctx); state != "" {
		carrier.Set(TraceStateHeader, state)
	}
}

// ExtractTrace returns ctx augmented with whatever trace-context fields are
// readable from carrier. A malformed traceparent yields the original ctx
// untouched so callers can detect the absence (via GetTraceID(ctx) == "")
// and decide whether to mint a new trace.
func ExtractTrace(ctx context.Context, carrier HeaderCarrier) context.Context {
	tp := carrier.Get(TraceParentHeader)
	if tp == "" {
		return ctx
	}
	tc, err := parseTraceParent(tp, carrier.Get(TraceStateHeader))
	if err != nil {
		return ctx
	}
	ctx = WithTraceID(ctx, tc.TraceID)
	ctx = WithSpanID(ctx, tc.SpanID)
	ctx = WithTraceFlags(ctx, tc.Flags)
	if tc.State != "" {
		ctx = WithTraceState(ctx, tc.State)
	}
	return ctx
}
