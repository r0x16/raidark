// Package observability verifies the canonical Prometheus collectors and Echo
// middleware used by Raidark services.
package observability

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMetrics_RecordsRouteStatusAndDuration(t *testing.T) {
	metrics := NewMetrics()
	e := echo.New()
	e.Use(HTTPMetrics(metrics))
	e.GET("/orders/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusAccepted)
	})

	request := httptest.NewRequest(http.MethodGet, "/orders/123", nil)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("202", "/orders/:id")))

	histogram := collectHistogramForLabel(t, metrics.HTTPRequestDurationMs, "endpoint", "/orders/:id")
	assert.Equal(t, uint64(1), histogram.GetSampleCount())
	assert.GreaterOrEqual(t, bucketCountAtOrBelow(histogram, 5000), uint64(1))
}

func TestHTTPMetrics_RecordsUnknownEndpointForUnmatchedRoutes(t *testing.T) {
	metrics := NewMetrics()
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/unmatched", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	handler := HTTPMetrics(metrics)(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	require.NoError(t, handler(context))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("204", "unknown")))
}

func TestHTTPMetrics_RecordsStatus500WhenHandlerReturnsErrorBeforeWriting(t *testing.T) {
	metrics := NewMetrics()
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/broken", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	handler := HTTPMetrics(metrics)(func(c echo.Context) error {
		return errors.New("boom")
	})

	require.EqualError(t, handler(context), "boom")
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("500", "unknown")))
}

func TestEventMetrics_RecordCountersHistogramsAndGauge(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordEventPublished("raidark.user.created", "success")
	metrics.RecordEventConsumed("raidark.user.created", "projection", "failure")
	metrics.RecordEventRedelivery("raidark.user.created", "projection")
	metrics.ObserveEventProcessing("raidark.user.created", "projection", 25)
	metrics.SetOutboxPending(7)

	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.EventsPublishedTotal.WithLabelValues("raidark.user.created", "success")))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.EventsConsumedTotal.WithLabelValues("raidark.user.created", "projection", "failure")))
	assert.Equal(t, 1.0, testutil.ToFloat64(metrics.EventsRedeliveriesTotal.WithLabelValues("raidark.user.created", "projection")))
	assert.Equal(t, 7.0, testutil.ToFloat64(metrics.OutboxPending))

	histogram := collectHistogramForLabel(t, metrics.EventProcessingDurationMs, "consumer", "projection")
	assert.Equal(t, uint64(1), histogram.GetSampleCount())
	assert.GreaterOrEqual(t, bucketCountAtOrBelow(histogram, 25), uint64(1))
}

func collectHistogramForLabel(t *testing.T, collector prometheus.Collector, labelName, labelValue string) *dto.Histogram {
	t.Helper()

	metricCh := make(chan prometheus.Metric)
	go func() {
		collector.Collect(metricCh)
		close(metricCh)
	}()

	for metric := range metricCh {
		dtoMetric := &dto.Metric{}
		require.NoError(t, metric.Write(dtoMetric))
		for _, label := range dtoMetric.GetLabel() {
			if label.GetName() == labelName && label.GetValue() == labelValue {
				return dtoMetric.GetHistogram()
			}
		}
	}
	t.Fatalf("histogram with label %s=%s was not collected", labelName, labelValue)
	return nil
}

func bucketCountAtOrBelow(histogram *dto.Histogram, upperBound float64) uint64 {
	for _, bucket := range histogram.GetBucket() {
		if bucket.GetUpperBound() == upperBound {
			return bucket.GetCumulativeCount()
		}
	}
	return 0
}
