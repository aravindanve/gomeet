package main_test

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/provider"
	"github.com/aravindanve/livemeet-server/src/resource"
	"github.com/gorilla/mux"
)

func newMockMeeting(ctx context.Context) resource.Meeting {
	// create mock meeting
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	// create code
	buf := make([]byte, 7)
	_, err := rand.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("error reading random bytes: %s", err.Error()))
	}

	code := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
	code = code[:4] + "-" + code[4:8] + "-" + code[8:]

	user := getMockUser()
	mockMeeting := &resource.Meeting{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err = p.MeetingCollection().Save(ctx, mockMeeting)
	if err != nil {
		panic(fmt.Sprintf("error saving meeting: %s", err.Error()))
	}
	return *mockMeeting
}

func TestMeetingCreate(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/meetings", nil)
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
	var m resource.Meeting
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID == "" {
		t.Errorf("expected id in response got %#v", m.ID)
		return
	}
	if m.UserID == "" {
		t.Errorf("expected userId in response got %#v", m.UserID)
		return
	}
	if m.Code == "" {
		t.Errorf("expected code in response got %#v", m.Code)
		return
	}
}

func TestMeetingCreateNoAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
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
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	meeting := newMockMeeting(ctx)

	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)
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
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	meeting := newMockMeeting(ctx)

	r := resource.RegisterMeetingRoutes(mux.NewRouter(), p)
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
