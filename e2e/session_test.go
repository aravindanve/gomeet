package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aravindanve/gomeet-server/src/middleware"
	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/gorilla/mux"
)

func TestSessionRetrieveWithAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	user := getMockUser()

	r := resource.RegisterSessionRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	req.Header.Set("authorization", getMockAuthHeader())

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.Session
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.User.ID != user.ID {
		t.Errorf("expected session user with ID %v got %v", user.ID, m.User.ID)
		return
	}
}

func TestSessionRetrieveNotAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterSessionRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/session", nil)

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.Session
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.User != nil {
		t.Errorf("expected session user to be %v got %v", nil, m.User)
		return
	}
}
