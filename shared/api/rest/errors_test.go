// Package rest_test verifies Raidark's public REST error envelope contract
// from the point of view of services and HTTP clients.
package rest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const snapshotTraceID = "trace-rdk-002"

type errorEnvelopeSnapshot struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// TestRenderError_matchesSentinelSnapshots fixes the client-visible JSON shape
// for every canonical sentinel so accidental envelope drift is caught early.
func TestRenderError_matchesSentinelSnapshots(t *testing.T) {
	snapshots := loadErrorEnvelopeSnapshots(t)
	cases := map[string]error{
		"not_found":  rest.ErrNotFound,
		"conflict":   rest.ErrConflict,
		"forbidden":  rest.ErrForbidden,
		"validation": rest.ErrValidation,
		"transient":  rest.ErrTransient,
		"permanent":  rest.ErrPermanent,
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			snapshot, ok := snapshots[name]
			require.True(t, ok, "missing snapshot for %s", name)

			status, body := renderMappedError(t, input)

			assert.Equal(t, snapshot.Status, status)
			assert.JSONEq(t, string(snapshot.Body), body)
		})
	}
}

// TestMapError_supportsWrappedSentinels verifies application services can add
// context with error wrapping without losing the REST status mapping.
func TestMapError_supportsWrappedSentinels(t *testing.T) {
	status, restErr := rest.MapError(fmt.Errorf("lookup account: %w", rest.ErrNotFound))

	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "common.not_found", restErr.Code)
}

// TestMapError_unknownErrorIsGeneric prevents accidental leakage of internal
// error text to HTTP clients.
func TestMapError_unknownErrorIsGeneric(t *testing.T) {
	status, body := renderMappedError(t, errors.New("database password leaked in stack"))

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Contains(t, body, `"code":"internal.unexpected"`)
	assert.Contains(t, body, "An unexpected error occurred.")
	assert.NotContains(t, body, "database password")
}

// TestRenderError_detailsAreOptional locks the details field contract: it is
// emitted only when the caller provides validation context.
func TestRenderError_detailsAreOptional(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		status, body := renderDirectError(t, &rest.RESTError{
			Code:    "profile.invalid",
			Message: "Invalid profile payload.",
			Details: map[string]any{
				"display_name": "required",
			},
			TraceID: snapshotTraceID,
		})

		assert.Equal(t, http.StatusBadRequest, status)
		assert.JSONEq(t, `{
			"error": {
				"code": "profile.invalid",
				"message": "Invalid profile payload.",
				"details": { "display_name": "required" },
				"trace_id": "trace-rdk-002"
			}
		}`, body)
	})

	t.Run("without details", func(t *testing.T) {
		status, body := renderDirectError(t, &rest.RESTError{
			Code:    "profile.invalid",
			Message: "Invalid profile payload.",
			TraceID: snapshotTraceID,
		})

		assert.Equal(t, http.StatusBadRequest, status)
		assert.NotContains(t, body, "details")
		assert.JSONEq(t, `{
			"error": {
				"code": "profile.invalid",
				"message": "Invalid profile payload.",
				"trace_id": "trace-rdk-002"
			}
		}`, body)
	})
}

// TestRESTError_implementsError keeps RESTError usable in Go error chains.
func TestRESTError_implementsError(t *testing.T) {
	restErr := &rest.RESTError{Message: "human message"}

	assert.Equal(t, "human message", restErr.Error())
}

// TestEchoErrorHandler_rendersReturnedSentinels covers the global Echo bridge
// used when handlers return Raidark sentinels instead of rendering directly.
func TestEchoErrorHandler_rendersReturnedSentinels(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("X-Correlation-ID", snapshotTraceID)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	require.NoError(t, rest.CorrelationID()(func(c echo.Context) error {
		rest.EchoErrorHandler(rest.ErrForbidden, c)
		return nil
	})(context))

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "common.forbidden",
			"message": "You do not have permission to perform this action.",
			"trace_id": "trace-rdk-002"
		}
	}`, recorder.Body.String())
}

func loadErrorEnvelopeSnapshots(t *testing.T) map[string]errorEnvelopeSnapshot {
	t.Helper()

	data, err := os.ReadFile("testdata/error_envelope_snapshots.json")
	require.NoError(t, err)

	var snapshots map[string]errorEnvelopeSnapshot
	require.NoError(t, json.Unmarshal(data, &snapshots))

	return snapshots
}

func renderMappedError(t *testing.T, input error) (int, string) {
	t.Helper()

	status, restErr := rest.MapError(input)
	return renderWithCorrelationID(t, status, restErr)
}

func renderDirectError(t *testing.T, restErr *rest.RESTError) (int, string) {
	t.Helper()

	return renderWithCorrelationID(t, http.StatusBadRequest, restErr)
}

func renderWithCorrelationID(t *testing.T, status int, restErr *rest.RESTError) (int, string) {
	t.Helper()

	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	request.Header.Set("X-Correlation-ID", snapshotTraceID)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	handler := rest.CorrelationID()(func(c echo.Context) error {
		return rest.RenderError(c, status, restErr)
	})
	require.NoError(t, handler(context))

	return recorder.Code, recorder.Body.String()
}
