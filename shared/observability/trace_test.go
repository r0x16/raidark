// Package observability verifies Raidark's trace-context propagation contract
// across HTTP middleware and transport-agnostic header carriers.
package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTraceID = "11111111111111111111111111111111"
	testSpanID  = "2222222222222222"
)

func TestW3CTrace_ValidIncomingTraceparentPropagatesTraceFields(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/work", nil)
	request.Header.Set(TraceParentHeader, "00-"+testTraceID+"-"+testSpanID+"-01")
	request.Header.Set(TraceStateHeader, "vendor=value")
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	var capturedSpan string
	handler := W3CTrace()(func(c echo.Context) error {
		capturedSpan = GetSpanID(c.Request().Context())
		assert.Equal(t, testTraceID, GetTraceID(c.Request().Context()))
		assert.Equal(t, capturedSpan, c.Get(ContextSpanIDKey))
		assert.Equal(t, testTraceID, c.Get(ContextTraceIDKey))
		assert.Equal(t, "01", c.Get(ContextTraceFlagsKey))
		assert.Equal(t, "vendor=value", c.Get(ContextTraceStateKey))
		return c.NoContent(http.StatusNoContent)
	})

	require.NoError(t, handler(context))

	require.Len(t, capturedSpan, spanIDLen)
	assert.NotEqual(t, testSpanID, capturedSpan)
	responseTrace, err := parseTraceParent(recorder.Header().Get(TraceParentHeader), recorder.Header().Get(TraceStateHeader))
	require.NoError(t, err)
	assert.Equal(t, testTraceID, responseTrace.TraceID)
	assert.Equal(t, capturedSpan, responseTrace.SpanID)
	assert.Equal(t, "vendor=value", responseTrace.State)
}

func TestW3CTrace_UsesCorrelationIDWhenTraceparentIsMissing(t *testing.T) {
	const correlationID = "018f6b7a-2c3d-7abc-8def-123456789abc"
	const expectedTraceID = "018f6b7a2c3d7abc8def123456789abc"

	recorder, capturedTraceID := runTraceMiddleware(t, map[string]string{
		correlationIDHeader: correlationID,
	})

	assert.Equal(t, expectedTraceID, capturedTraceID)
	assert.Contains(t, recorder.Header().Get(TraceParentHeader), expectedTraceID)
}

func TestW3CTrace_GeneratesTraceForMissingOrMalformedTraceparent(t *testing.T) {
	tests := map[string]map[string]string{
		"missing": {},
		"malformed": {
			TraceParentHeader: "not-a-traceparent",
		},
	}

	for name, headers := range tests {
		t.Run(name, func(t *testing.T) {
			recorder, capturedTraceID := runTraceMiddleware(t, headers)

			require.Len(t, capturedTraceID, traceIDLen)
			assert.True(t, isLowerHex(capturedTraceID))
			trace, err := parseTraceParent(recorder.Header().Get(TraceParentHeader), "")
			require.NoError(t, err)
			assert.Equal(t, capturedTraceID, trace.TraceID)
			assert.Len(t, trace.SpanID, spanIDLen)
		})
	}
}

func TestInjectAndExtractTrace_RoundTripThroughMapCarrier(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, testTraceID)
	ctx = WithSpanID(ctx, testSpanID)
	ctx = WithTraceFlags(ctx, "00")
	ctx = WithTraceState(ctx, "vendor=value")
	carrier := MapCarrier{}

	InjectTrace(ctx, carrier)

	assert.Equal(t, "00-"+testTraceID+"-"+testSpanID+"-00", carrier.Get(TraceParentHeader))
	assert.Equal(t, "vendor=value", carrier.Get(TraceStateHeader))

	extracted := ExtractTrace(context.Background(), carrier)
	assert.Equal(t, testTraceID, GetTraceID(extracted))
	assert.Equal(t, testSpanID, GetSpanID(extracted))
	assert.Equal(t, "00", GetTraceFlags(extracted))
	assert.Equal(t, "vendor=value", GetTraceState(extracted))
}

func TestInjectTrace_WithoutTraceIDIsNoop(t *testing.T) {
	carrier := MapCarrier{}

	InjectTrace(context.Background(), carrier)

	assert.Empty(t, carrier)
}

func TestInjectTrace_WithoutSpanIDIsNoop(t *testing.T) {
	carrier := MapCarrier{}

	InjectTrace(WithTraceID(context.Background(), testTraceID), carrier)

	assert.Empty(t, carrier)
}

func TestInjectTrace_UsesDefaultFlagsAndOmitsEmptyTraceState(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, testTraceID)
	ctx = WithSpanID(ctx, testSpanID)
	carrier := MapCarrier{}

	InjectTrace(ctx, carrier)

	assert.Equal(t, "00-"+testTraceID+"-"+testSpanID+"-01", carrier.Get(TraceParentHeader))
	assert.Empty(t, carrier.Get(TraceStateHeader))
}

func TestExtractTrace_WithoutTraceparentLeavesContextUntouched(t *testing.T) {
	original := WithTraceID(context.Background(), "existing")

	extracted := ExtractTrace(original, MapCarrier{})

	assert.Equal(t, "existing", GetTraceID(extracted))
}

func TestExtractTrace_ValidTraceparentWithoutTracestate(t *testing.T) {
	extracted := ExtractTrace(context.Background(), MapCarrier{
		TraceParentHeader: "00-" + testTraceID + "-" + testSpanID + "-01",
	})

	assert.Equal(t, testTraceID, GetTraceID(extracted))
	assert.Equal(t, testSpanID, GetSpanID(extracted))
	assert.Equal(t, "01", GetTraceFlags(extracted))
	assert.Empty(t, GetTraceState(extracted))
}

func TestExtractTrace_MalformedTraceparentLeavesContextUntouched(t *testing.T) {
	original := WithTraceID(context.Background(), "existing")
	carrier := MapCarrier{TraceParentHeader: "00-invalid"}

	extracted := ExtractTrace(original, carrier)

	assert.Equal(t, "existing", GetTraceID(extracted))
}

func TestParseTraceParent_RejectsUnsupportedWireValues(t *testing.T) {
	tests := map[string]string{
		"bad version length":  "0-" + testTraceID + "-" + testSpanID + "-01",
		"bad trace id length": "00-abc-" + testSpanID + "-01",
		"bad span id length":  "00-" + testTraceID + "-abc-01",
		"bad flags length":    "00-" + testTraceID + "-" + testSpanID + "-0",
		"wrong field count":   "00-" + testTraceID + "-" + testSpanID,
		"unsupported version": "01-" + testTraceID + "-" + testSpanID + "-01",
		"uppercase trace id":  "00-" + strings.ToUpper("abcdefabcdefabcdefabcdefabcdefab") + "-" + testSpanID + "-01",
		"zero trace id":       "00-" + strings.Repeat("0", traceIDLen) + "-" + testSpanID + "-01",
		"zero span id":        "00-" + testTraceID + "-" + strings.Repeat("0", spanIDLen) + "-01",
		"bad flags":           "00-" + testTraceID + "-" + testSpanID + "-zz",
	}

	for name, traceparent := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parseTraceParent(traceparent, "")
			require.ErrorIs(t, err, errInvalidTraceParent)
		})
	}
}

func TestFormatTraceParent_DefaultsMissingVersionAndFlags(t *testing.T) {
	traceparent := formatTraceParent(TraceContext{
		TraceID: testTraceID,
		SpanID:  testSpanID,
	})

	assert.Equal(t, "00-"+testTraceID+"-"+testSpanID+"-01", traceparent)
}

func TestTraceIDFromCorrelation_NormalizesAndRejectsInvalidValues(t *testing.T) {
	tests := map[string]struct {
		correlationID string
		expected      string
		ok            bool
	}{
		"uuid with dashes": {
			correlationID: "018f6b7a-2c3d-7abc-8def-123456789abc",
			expected:      "018f6b7a2c3d7abc8def123456789abc",
			ok:            true,
		},
		"uppercase uuid": {
			correlationID: "018F6B7A-2C3D-7ABC-8DEF-123456789ABC",
			expected:      "018f6b7a2c3d7abc8def123456789abc",
			ok:            true,
		},
		"wrong length": {
			correlationID: "abc",
		},
		"not hex": {
			correlationID: "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
		},
		"zero id": {
			correlationID: strings.Repeat("0", traceIDLen),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, ok := traceIDFromCorrelation(tt.correlationID)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func runTraceMiddleware(t *testing.T, headers map[string]string) (*httptest.ResponseRecorder, string) {
	t.Helper()

	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/work", nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	var capturedTraceID string
	handler := W3CTrace()(func(c echo.Context) error {
		capturedTraceID = GetTraceID(c.Request().Context())
		return c.NoContent(http.StatusNoContent)
	})
	require.NoError(t, handler(context))

	return recorder, capturedTraceID
}
