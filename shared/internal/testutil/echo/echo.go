package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	echov4 "github.com/labstack/echo/v4"
)

type ContextOptions struct {
	Method  string
	Target  string
	Body    io.Reader
	Headers map[string]string
}

type Context struct {
	Echo     *echov4.Echo
	Request  *http.Request
	Recorder *httptest.ResponseRecorder
	Context  echov4.Context
}

// NewContext builds an Echo context backed by httptest request/recorder values.
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
