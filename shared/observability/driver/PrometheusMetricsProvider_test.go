// Package driver verifies concrete observability adapters.
package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetricsProvider_ExposesPrivateRegistryAtConfiguredPath(t *testing.T) {
	provider := NewPrometheusMetricsProvider("/internal/metrics")
	var _ obsdomain.MetricsProvider = provider
	provider.Metrics().SetOutboxPending(11)

	request := httptest.NewRequest(http.MethodGet, "/internal/metrics", nil)
	recorder := httptest.NewRecorder()
	provider.Handler().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "/internal/metrics", provider.Path())
	assert.Contains(t, recorder.Body.String(), "outbox_pending_gauge")
	assert.Contains(t, recorder.Body.String(), " 11")
}
