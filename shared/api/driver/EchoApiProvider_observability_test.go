// Package driver_test verifies observability wiring in the shared Echo API
// provider.
package driver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/testutil"
	apidriver "github.com/r0x16/Raidark/shared/api/driver"
	envdomain "github.com/r0x16/Raidark/shared/env/domain"
	logdomain "github.com/r0x16/Raidark/shared/logger/domain"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	obsdriver "github.com/r0x16/Raidark/shared/observability/driver"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoApiProviderSetup_RecordsHTTPMetricsWhenProviderExists(t *testing.T) {
	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, testEnvProvider{})
	providerdomain.Register[logdomain.LogProvider](hub, testLogProvider{})
	metricsProvider := obsdriver.NewPrometheusMetricsProvider("/metrics")
	providerdomain.Register[obsdomain.MetricsProvider](hub, metricsProvider)
	provider := apidriver.NewEchoApiProvider("8080", hub)

	require.NoError(t, provider.Setup())
	provider.Server.GET("/widgets/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusAccepted)
	})

	request := httptest.NewRequest(http.MethodGet, "/widgets/123", nil)
	recorder := httptest.NewRecorder()
	provider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, 1.0, testutil.ToFloat64(metricsProvider.Metrics().HTTPRequestsTotal.WithLabelValues("202", "/widgets/:id")))
}
