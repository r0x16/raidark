// Package observability provides the cross-cutting telemetry primitives used by
// every service built on top of Raidark: structured logging with automatic
// correlation fields, Prometheus metrics, and W3C trace-context propagation.
//
// The package is intentionally framework-light. Public middlewares plug into
// Echo, but trace/log helpers are pure context.Context utilities and can be
// reused from any background worker, CLI command, or event consumer that does
// not run inside an HTTP request.
//
// Subpackages:
//
//   - log: context-aware LogProvider that auto-injects trace_id, span_id,
//     service and event_id from context.
//
// The package is the integration point for downstream observability stacks
// (OpenTelemetry, Loki, Tempo). The W3C trace-context plumbing is wire-format
// compatible so a future OTLP exporter can be slotted in without touching
// handlers or middlewares.
package observability
