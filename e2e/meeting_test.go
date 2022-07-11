package main_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aravindanve/gomeet-server/src/middleware"
	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/gorilla/mux"
)

func TestMeetingCreate(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/meetings", nil)
	req.Header.Set("authorization", getTestAuthHeader())

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m map[string]any
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if v := m["id"]; v == "" {
		t.Errorf("expected id in response got %#v", v)
		return
	}
	if v := m["userId"]; v == "" {
		t.Errorf("expected userId in response got %#v", v)
		return
	}
	if v := m["code"]; v == "" {
		t.Errorf("expected code in response got %#v", v)
		return
	}
}

func TestMeetingCreateNoAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/meetings", nil)

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusUnauthorized {
		t.Errorf("expected status to be %#v got %#v", http.StatusUnauthorized, s)
		return
	}

	// test response
	var m map[string]any
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if v := m["error"]; v != true {
		t.Errorf("expected error to be %#v got %#v", true, v)
		return
	}
	if v := m["message"]; v != "Unauthorized" {
		t.Errorf("expected error to be %#v got %#v", "Unauthorized", v)
		return
	}
}

func TestMeetingSearch(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)

	meeting := getTestMeeting()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/meetings?code="+meeting.Code, nil)

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m map[string][]*resource.Meeting
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	meetings := m["meetings"]
	if len(meetings) != 1 {
		t.Errorf("expected meetings with one item in response got %#v", meetings)
		return
	}
	if meetings[0].ID != meeting.ID {
		t.Errorf("expected meeting with ID %v got %v", meeting.ID, meetings[0].ID)
		return
	}
}

func TestMeetingRetrieve(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)

	meeting := getTestMeeting()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/meetings/"+string(meeting.ID), nil)

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.Meeting
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID != meeting.ID {
		t.Errorf("expected meeting with ID %v got %v", meeting.ID, m.ID)
		return
	}
}
