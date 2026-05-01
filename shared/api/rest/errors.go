// Package rest provides standard HTTP response conventions for Raidark services:
// error envelopes, pagination, and correlation ID propagation.
// All public types and helpers in this package are designed to produce stable,
// client-parseable JSON shapes across every service built on top of Raidark.
package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RESTError is the canonical error payload returned by all Raidark REST endpoints.
// It wraps a namespaced code for machine parsing, a human-readable message,
// an optional details map for field-level validation context, and a trace_id
// that clients can use to correlate cross-service log entries.
type RESTError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// Error implements the error interface so RESTError can flow through standard Go error chains.
func (e *RESTError) Error() string {
	return e.Message
}

// Sentinel errors used as canonical error values across the application layer.
// Pass these (or errors that wrap them) to MapError to obtain the correct HTTP status.
var (
	// ErrNotFound signals that the requested resource does not exist.
	ErrNotFound = errors.New("rest: not found")

	// ErrConflict signals that the operation conflicts with the current resource state.
	ErrConflict = errors.New("rest: conflict")

	// ErrForbidden signals that the caller lacks permission for the operation.
	ErrForbidden = errors.New("rest: forbidden")

	// ErrValidation signals that the request payload failed validation.
	ErrValidation = errors.New("rest: validation failed")

	// ErrTransient signals a temporary failure that the caller may safely retry.
	ErrTransient = errors.New("rest: transient failure")

	// ErrPermanent signals a non-retryable server-side failure.
	ErrPermanent = errors.New("rest: permanent failure")
)

// RenderError serializes e as {"error": {...}} JSON and writes it to the response
// with the given HTTP status code. If e.TraceID is empty, it is populated from the
// correlation ID stored in c (set by the CorrelationID middleware).
func RenderError(c echo.Context, status int, e *RESTError) error {
	if e.TraceID == "" {
		e.TraceID = GetCorrelationID(c)
	}
	return c.JSON(status, map[string]any{"error": e})
}

// MapError translates a sentinel (or a wrapped sentinel) to the canonical HTTP status
// and a RESTError. Unknown errors always yield 500 with code "internal.unexpected" and
// a generic message — the original error is intentionally not exposed to the caller.
func MapError(err error) (int, *RESTError) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, &RESTError{
			Code:    "common.not_found",
			Message: "The requested resource was not found.",
		}
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, &RESTError{
			Code:    "common.conflict",
			Message: "The request conflicts with the current state of the resource.",
		}
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, &RESTError{
			Code:    "common.forbidden",
			Message: "You do not have permission to perform this action.",
		}
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest, &RESTError{
			Code:    "common.validation_failed",
			Message: "The request payload is invalid.",
		}
	case errors.Is(err, ErrTransient):
		return http.StatusServiceUnavailable, &RESTError{
			Code:    "common.transient_failure",
			Message: "A transient error occurred. Please retry.",
		}
	case errors.Is(err, ErrPermanent):
		return http.StatusInternalServerError, &RESTError{
			Code:    "common.permanent_failure",
			Message: "A permanent error occurred.",
		}
	default:
		return http.StatusInternalServerError, &RESTError{
			Code:    "internal.unexpected",
			Message: "An unexpected error occurred.",
		}
	}
}
