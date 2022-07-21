package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddlewareWithPreflightRequest(t *testing.T) {
	t.Parallel()

	// create request
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("origin", "https://example.com")

	// create recorder
	w := httptest.NewRecorder()

	// create middleware
	m := CORSMiddleware()
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})).ServeHTTP(w, r)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusNoContent {
		t.Errorf("expected status to be %#v got %#v", http.StatusNoContent, s)
		return
	}

	// test headers
	if h := w.Header().Get("access-control-allow-origin"); h != "https://example.com" {
		t.Errorf(`expected access-control-allow-origin to be "https://example.com" got %q`, h)
		return
	}
	if h := w.Header().Get("vary"); h != "Origin" {
		t.Errorf(`expected vary to be "Origin" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-credentials"); h != "true" {
		t.Errorf(`expected access-control-allow-credentials to be "true" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-methods"); h != "GET,HEAD,PUT,POST,DELETE,PATCH" {
		t.Errorf(`expected access-control-allow-methods to be "GET,HEAD,PUT,POST,DELETE,PATCH" got %q`, h)
		return
	}
}

func TestCORSMiddlewareWithCORSRequest(t *testing.T) {
	t.Parallel()

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("origin", "https://example.com")

	// create recorder
	w := httptest.NewRecorder()

	// create middleware
	m := CORSMiddleware()
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})).ServeHTTP(w, r)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusNotFound {
		t.Errorf("expected status to be %#v got %#v", http.StatusNotFound, s)
		return
	}

	// test headers
	if h := w.Header().Get("access-control-allow-origin"); h != "https://example.com" {
		t.Errorf(`expected access-control-allow-origin to be "https://example.com" got %q`, h)
		return
	}
	if h := w.Header().Get("vary"); h != "Origin" {
		t.Errorf(`expected vary to be "Origin" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-credentials"); h != "true" {
		t.Errorf(`expected access-control-allow-credentials to be "true" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-methods"); h != "" {
		t.Errorf(`expected access-control-allow-methods to be "" got %q`, h)
		return
	}
}

func TestCORSMiddlewareWithSameOriginRequest(t *testing.T) {
	t.Parallel()

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// create recorder
	w := httptest.NewRecorder()

	// create middleware
	m := CORSMiddleware()
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})).ServeHTTP(w, r)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusNotFound {
		t.Errorf("expected status to be %#v got %#v", http.StatusNotFound, s)
		return
	}

	// test headers
	if h := w.Header().Get("access-control-allow-origin"); h != "" {
		t.Errorf(`expected access-control-allow-origin to be "" got %q`, h)
		return
	}
	if h := w.Header().Get("vary"); h != "" {
		t.Errorf(`expected vary to be "" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-credentials"); h != "" {
		t.Errorf(`expected access-control-allow-credentials to be "" got %q`, h)
		return
	}
	if h := w.Header().Get("access-control-allow-methods"); h != "" {
		t.Errorf(`expected access-control-allow-methods to be "" got %q`, h)
		return
	}
}
