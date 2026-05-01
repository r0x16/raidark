// Package echo valida los helpers de contexto Echo con smoke tests mínimos.
package echo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewContext_smoke verifica que NewContext preserve método, ruta y headers,
// y que entregue un contexto Echo listo para usar en handlers.
func TestNewContext_smoke(t *testing.T) {
	ctx := NewContext(t, ContextOptions{
		Method: http.MethodPost,
		Target: "/health",
		Headers: map[string]string{
			"X-Test": "raidark",
		},
	})

	require.NotNil(t, ctx.Echo)
	require.NotNil(t, ctx.Context)
	assert.Equal(t, http.MethodPost, ctx.Request.Method)
	assert.Equal(t, "/health", ctx.Request.URL.Path)
	assert.Equal(t, "raidark", ctx.Context.Request().Header.Get("X-Test"))
}
