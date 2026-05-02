package modules

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/domain"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// EchoMetricsModule mounts the Prometheus scrape endpoint on the shared
// Echo server, following the standard Raidark module pattern (embed
// *EchoModule, register routes in Setup, declare empty model / seed /
// listener slices).
//
// Setup is a no-op when no MetricsProvider has been registered on the hub
// — services opt into metrics by adding driverprovider.MetricsProviderFactory
// to their main.go providers list, and only then does this module mount
// the route. Keeping the module always present (and conditional inside
// Setup) means raidark.go does not have to know whether metrics are on.
type EchoMetricsModule struct {
	*EchoModule
}

var _ domain.ApiModule = &EchoMetricsModule{}

// Name implements domain.ApiModule. Surfaces in the route listing logged
// at startup ("Setup module: Metrics").
func (e *EchoMetricsModule) Name() string {
	return "Metrics"
}

// Setup implements domain.ApiModule. The route is mounted at the absolute
// root of the Echo server (not nested under any /api group) because
// Prometheus scrapers are configured to hit `/metrics` directly.
func (e *EchoMetricsModule) Setup() error {
	if !domprovider.Exists[obsdomain.MetricsProvider](e.Hub) {
		return nil
	}
	provider := domprovider.Get[obsdomain.MetricsProvider](e.Hub)
	e.Group.GET(provider.Path(), echo.WrapHandler(provider.Handler()))
	return nil
}
