package observability

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

// W3C trace-context wire-format constants. The shape is
// `<version>-<trace-id>-<parent-id>-<flags>` per
// https://www.w3.org/TR/trace-context/#traceparent-header.
const (
	// TraceParentHeader is the canonical request/response header name carrying
	// the W3C traceparent value.
	TraceParentHeader = "traceparent"
	// TraceStateHeader is the canonical header name carrying the W3C
	// tracestate value used for vendor-specific extensions.
	TraceStateHeader = "tracestate"

	traceVersionLen = 2
	traceIDLen      = 32
	spanIDLen       = 16
	traceFlagsLen   = 2

	// supportedVersion is the only version of W3C trace-context that this
	// implementation understands. Newer versions are rejected and a fresh
	// trace is generated to avoid silently propagating unsupported metadata.
	supportedVersion = "00"

	// defaultTraceFlags marks a trace as recorded ("01"). When we generate a
	// new trace we always mark it as sampled so downstream backends can
	// decide whether to actually persist it.
	defaultTraceFlags = "01"
)

// TraceContext holds a parsed W3C traceparent triple plus the raw incoming
// tracestate. It is the in-memory representation used by middlewares and
// propagation helpers.
type TraceContext struct {
	// Version is the W3C trace-context version byte in lowercase hex.
	// Currently only "00" is supported.
	Version string
	// TraceID is the 16-byte trace identifier rendered as 32 lowercase hex
	// characters. It MUST NOT be all zeros per the spec.
	TraceID string
	// SpanID is the 8-byte span identifier rendered as 16 lowercase hex
	// characters. For incoming traceparent values this is the parent span;
	// the middleware always generates a new SpanID for the local span.
	SpanID string
	// Flags is the trace-flags byte rendered as 2 lowercase hex characters
	// (e.g. "01" = sampled).
	Flags string
	// State is the raw tracestate value (may be empty). It is propagated
	// downstream verbatim and not interpreted.
	State string
}

// errInvalidTraceParent is returned by parseTraceParent when the input does
// not conform to the W3C trace-context spec. Callers handle the error by
// minting a fresh trace rather than failing the request — the spec says
// receivers MUST NOT propagate broken values, but they MAY continue.
var errInvalidTraceParent = errors.New("observability: invalid traceparent")

// parseTraceParent decodes a W3C traceparent header into a TraceContext.
// It enforces version "00", non-zero IDs, lowercase hex, and the canonical
// 4-field shape. Anything else is rejected. State is the raw tracestate
// header (passed through untouched).
func parseTraceParent(traceparent, tracestate string) (TraceContext, error) {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return TraceContext{}, errInvalidTraceParent
	}

	version, traceID, spanID, flags := parts[0], parts[1], parts[2], parts[3]

	if len(version) != traceVersionLen || !isLowerHex(version) {
		return TraceContext{}, errInvalidTraceParent
	}
	if version != supportedVersion {
		return TraceContext{}, errInvalidTraceParent
	}
	if len(traceID) != traceIDLen || !isLowerHex(traceID) || isAllZeroHex(traceID) {
		return TraceContext{}, errInvalidTraceParent
	}
	if len(spanID) != spanIDLen || !isLowerHex(spanID) || isAllZeroHex(spanID) {
		return TraceContext{}, errInvalidTraceParent
	}
	if len(flags) != traceFlagsLen || !isLowerHex(flags) {
		return TraceContext{}, errInvalidTraceParent
	}

	return TraceContext{
		Version: version,
		TraceID: traceID,
		SpanID:  spanID,
		Flags:   flags,
		State:   tracestate,
	}, nil
}

// formatTraceParent renders tc as a W3C traceparent header value.
// Spec: `<version>-<trace-id>-<span-id>-<flags>`, all lowercase hex.
func formatTraceParent(tc TraceContext) string {
	version := tc.Version
	if version == "" {
		version = supportedVersion
	}
	flags := tc.Flags
	if flags == "" {
		flags = defaultTraceFlags
	}
	return version + "-" + tc.TraceID + "-" + tc.SpanID + "-" + flags
}

// newTraceID returns a freshly generated 32-character lowercase hex trace_id
// using crypto/rand. Falls back to all-ones if the OS entropy source fails so
// the request can still be served — the spec disallows all-zero IDs but does
// not constrain other fixed patterns, and a deterministic fallback is
// preferable to refusing the request outright.
func newTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strings.Repeat("f", traceIDLen)
	}
	return hex.EncodeToString(buf[:])
}

// newSpanID returns a freshly generated 16-character lowercase hex span_id
// using crypto/rand. Same fallback rationale as newTraceID.
func newSpanID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strings.Repeat("f", spanIDLen)
	}
	return hex.EncodeToString(buf[:])
}

// traceIDFromCorrelation tries to derive a W3C trace_id from a value supplied
// in another header (typically X-Correlation-ID populated as a UUIDv7).
// Returns the normalized 32-char lowercase hex string and true on success;
// returns "", false otherwise so the caller can fall back to newTraceID.
//
// The conversion strips dashes (UUIDs use 8-4-4-4-12 form, traceparent uses
// no separators), lower-cases the result, and rejects anything that does not
// resolve to exactly 32 hex characters. Non-zero check is enforced because
// the spec disallows the zero trace_id.
func traceIDFromCorrelation(correlationID string) (string, bool) {
	candidate := strings.ToLower(strings.ReplaceAll(correlationID, "-", ""))
	if len(candidate) != traceIDLen || !isLowerHex(candidate) || isAllZeroHex(candidate) {
		return "", false
	}
	return candidate, true
}

// isLowerHex reports whether s contains only lowercase hex digits. The W3C
// spec requires lowercase; uppercase values must be rejected.
func isLowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// isAllZeroHex reports whether s is composed entirely of '0' characters.
// Used to reject the spec-disallowed all-zero trace_id and span_id.
func isAllZeroHex(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}
