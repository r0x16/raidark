package observability

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// HTTPMetrics returns an Echo middleware that records every request into the
// Prometheus collectors held by m. The metrics emitted are:
//
//   - http_requests_total{status, endpoint}
//   - http_request_duration_ms{endpoint}
//
// "endpoint" uses the matched route pattern (Echo's c.Path()) rather than
// the raw URL path. Using the raw path would create a unique time series for
// every distinct value of any path parameter, which routinely exceeds the
// recommended 10k label-value cardinality and causes Prometheus to OOM. If
// the route did not match (404), the matched path is empty and we record
// "unknown" so the series is still scrape-able.
func HTTPMetrics(m *Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			// Status: prefer the response status set by the handler; if the
			// handler returned an error and Echo's error handler hasn't run
			// yet, fall back to 500 so we still record a sensible bucket.
			status := c.Response().Status
			if status == 0 {
				status = 500
			}

			endpoint := c.Path()
			if endpoint == "" {
				endpoint = "unknown"
			}

			statusLabel := strconv.Itoa(status)
			m.HTTPRequestsTotal.WithLabelValues(statusLabel, endpoint).Inc()
			m.HTTPRequestDurationMs.WithLabelValues(endpoint).Observe(
				float64(time.Since(start).Microseconds()) / 1000.0,
			)

			return err
		}
	}
}
