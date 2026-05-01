package echo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
