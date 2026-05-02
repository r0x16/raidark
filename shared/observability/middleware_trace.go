package observability

import "github.com/labstack/echo/v4"

// correlationIDHeader mirrors shared/api/rest.correlationIDHeader.
// We re-declare it here to keep observability free of an import cycle with
// shared/api/rest (which depends on shared/ids and the rest envelope, none of
// which observability needs). The header name is part of the public wire
// protocol of Raidark and changes are coordinated repo-wide.
const correlationIDHeader = "X-Correlation-ID"

// W3CTrace returns an Echo middleware that establishes a W3C trace-context
// for the request lifetime. Order of resolution for the trace_id:
//
//  1. A valid `traceparent` header on the incoming request — the caller is
//     joining an existing trace, we adopt their trace_id and parent span_id.
//  2. An `X-Correlation-ID` header that decodes to a 16-byte hex value
//     (UUIDv7 with dashes stripped qualifies). This bridges the legacy
//     correlation-ID protocol with W3C trace-context so services that haven't
//     adopted traceparent yet still get cross-service correlation for free.
//  3. A freshly generated trace_id from crypto/rand.
//
// The local span_id is always freshly generated — this middleware represents
// the entry into the local service, so it owns a new span regardless of
// whether the trace was inherited.
//
// Resolved values are exposed via:
//   - echo.Context keys EchoTraceIDKey / EchoSpanIDKey / EchoTraceFlagsKey /
//     EchoTraceStateKey for handlers that read echo.Context directly.
//   - The request's context.Context (replaced on c.Request via WithContext)
//     so log.FromContext, downstream goroutines, and DB calls can pick the
//     fields up without echo coupling.
//   - The response `traceparent` header so callers can pick up where the
//     server left off.
func W3CTrace() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			carrier := httpHeaderCarrier{header: req.Header}

			tc, ok := resolveTraceContext(carrier, req.Header.Get(correlationIDHeader))

			// Local span is always fresh: this is the entry point of the
			// local service and any incoming span_id describes the parent.
			tc.SpanID = newSpanID()
			if !ok && tc.Flags == "" {
				tc.Flags = defaultTraceFlags
			}

			ctx := req.Context()
			ctx = WithTraceID(ctx, tc.TraceID)
			ctx = WithSpanID(ctx, tc.SpanID)
			ctx = WithTraceFlags(ctx, tc.Flags)
			if tc.State != "" {
				ctx = WithTraceState(ctx, tc.State)
			}
			c.SetRequest(req.WithContext(ctx))

			c.Set(EchoTraceIDKey, tc.TraceID)
			c.Set(EchoSpanIDKey, tc.SpanID)
			c.Set(EchoTraceFlagsKey, tc.Flags)
			if tc.State != "" {
				c.Set(EchoTraceStateKey, tc.State)
			}

			// Echo the resolved trace-context back so clients can see what
			// the server is using. This is not strictly required by the
			// spec, but it is invaluable for debugging: a curl with -i
			// shows the trace_id without needing log access.
			c.Response().Header().Set(TraceParentHeader, formatTraceParent(tc))
			if tc.State != "" {
				c.Response().Header().Set(TraceStateHeader, tc.State)
			}

			return next(c)
		}
	}
}

// resolveTraceContext picks the trace_id from (in order) a valid incoming
// traceparent, the X-Correlation-ID fallback, or a freshly minted value.
// It returns the partial TraceContext (without local SpanID) and a boolean
// indicating whether the trace was inherited (true) or generated (false).
func resolveTraceContext(carrier HeaderCarrier, correlationID string) (TraceContext, bool) {
	if tp := carrier.Get(TraceParentHeader); tp != "" {
		if tc, err := parseTraceParent(tp, carrier.Get(TraceStateHeader)); err == nil {
			return tc, true
		}
	}
	if correlationID != "" {
		if id, ok := traceIDFromCorrelation(correlationID); ok {
			return TraceContext{
				Version: supportedVersion,
				TraceID: id,
				Flags:   defaultTraceFlags,
			}, false
		}
	}
	return TraceContext{
		Version: supportedVersion,
		TraceID: newTraceID(),
		Flags:   defaultTraceFlags,
	}, false
}

// httpHeaderCarrier adapts http.Header to HeaderCarrier. The standard library
// already canonicalizes header names internally, so callers can pass
// lowercase W3C names ("traceparent") and reach values stored under the
// canonical form ("Traceparent") without manual normalization.
type httpHeaderCarrier struct {
	header headerLike
}

// headerLike is the subset of http.Header consumed by httpHeaderCarrier.
// Spelled out explicitly so unit tests can fake it without depending on
// net/http internals.
type headerLike interface {
	Get(key string) string
	Set(key, value string)
}

func (c httpHeaderCarrier) Get(key string) string  { return c.header.Get(key) }
func (c httpHeaderCarrier) Set(key, value string)  { c.header.Set(key, value) }
