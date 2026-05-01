// Package echo contiene helpers para construir contextos Echo en tests.
package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	echov4 "github.com/labstack/echo/v4"
)

// ContextOptions configura la request HTTP usada para construir el contexto.
type ContextOptions struct {
	Method  string
	Target  string
	Body    io.Reader
	Headers map[string]string
}

// Context agrupa todos los objetos útiles que participan en un test de handler
// Echo: instancia, request, recorder y contexto.
type Context struct {
	Echo     *echov4.Echo
	Request  *http.Request
	Recorder *httptest.ResponseRecorder
	Context  echov4.Context
}

// NewContext construye un contexto Echo respaldado por httptest, aplicando
// defaults seguros para method y target cuando el test no los especifica.
func NewContext(t testing.TB, opts ContextOptions) Context {
	t.Helper()

	method := opts.Method
	if method == "" {
		method = http.MethodGet
	}

	target := opts.Target
	if target == "" {
		target = "/"
	}

	req := httptest.NewRequest(method, target, opts.Body)
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	e := echov4.New()

	return Context{
		Echo:     e,
		Request:  req,
		Recorder: rec,
		Context:  e.NewContext(req, rec),
	}
}
