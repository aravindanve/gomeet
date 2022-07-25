package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aravindanve/livemeet-server/src/provider"
	"github.com/aravindanve/livemeet-server/src/route"
	"github.com/gorilla/mux"
)

func TestCORS(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	r := route.RegisterRoutes(mux.NewRouter(), p)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/meetings", nil)
	req.Header.Set("origin", "https://example.com")

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusNotFound {
		t.Errorf("expected status to be %#v got %#v", http.StatusNotFound, s)
		return
	}

	// test headers
	if h := w.Result().Header.Get("access-control-allow-origin"); h != "https://example.com" {
		t.Errorf(`expected access-control-allow-origin to be "https://example.com" got %q`, h)
		return
	}
}
