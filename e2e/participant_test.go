package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/middleware"
	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
)

func newMockMeetingAndParticipant(ctx context.Context) (resource.Meeting, resource.Participant) {
	// create mock participant
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	user := getMockUser()
	meeting := newMockMeeting(ctx)
	mockParticipant := &resource.Participant{
		MeetingID: meeting.ID,
		Name:      user.Name,
		ImageURL:  user.ImageURL,
		Status:    resource.ParticipantStatusWaiting,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err := p.ParticipantCollection().Save(ctx, mockParticipant)
	if err != nil {
		panic(fmt.Sprintf("error saving participant: %s", err.Error()))
	}
	return meeting, *mockParticipant
}

type mockParticipantProvider struct {
	config.AuthConfigProvider
	resource.ParticipantDeps
	livekitClient *mockLiveKitClient
	Release       func(ctx context.Context)
}

func newMockParticipantProvider(ctx context.Context) *mockParticipantProvider {
	p := provider.NewProvider(ctx)
	livekitClient := newMockLiveKitClient()

	return &mockParticipantProvider{
		AuthConfigProvider: p,
		ParticipantDeps:    p,
		Release:            p.Release,
		livekitClient:      livekitClient,
	}
}

func (m *mockParticipantProvider) LiveKitClient() client.LiveKitClient {
	return m.livekitClient
}

type mockLiveKitClient struct {
	sendDataReq *livekit.SendDataRequest
}

func newMockLiveKitClient() *mockLiveKitClient {
	return &mockLiveKitClient{}
}

func (m *mockLiveKitClient) SendData(ctx context.Context, req *livekit.SendDataRequest) (*livekit.SendDataResponse, error) {
	m.sendDataReq = req
	return &livekit.SendDataResponse{}, nil
}

func TestParticipantCreateWithAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	meeting := newMockMeeting(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/meetings/"+string(meeting.ID)+"/participants", nil)
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
	if v := m["meetingId"]; v == "" {
		t.Errorf("expected meetingId in response got %#v", v)
		return
	}
	if v := m["name"]; v != "Mock User" {
		t.Errorf(`expected name to be %q got %q`, "Mock User", v)
		return
	}
	if v := m["imageUrl"]; v == nil {
		t.Errorf("expected imageUrl in response got %#v", v)
		return
	}
	if v := m["status"]; v != "admitted" {
		t.Errorf(`expected status to be %q got %q`, "admitted", v)
		return
	}
	if v := m["joinToken"]; v == nil {
		t.Errorf("expected imageUrl in response got %#v", v)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq != nil {
		t.Errorf("expected livekit send data to be nil got %#v", err)
	}
}

func TestParticipantCreateNoAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	meeting := newMockMeeting(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/meetings/"+string(meeting.ID)+"/participants", strings.NewReader(`{"name":"My Name"}`))

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
	if v := m["meetingId"]; v == "" {
		t.Errorf("expected meetingId in response got %#v", v)
		return
	}
	if v := m["name"]; v != "My Name" {
		t.Errorf(`expected name to be %q got %q`, "My Name", v)
		return
	}
	if v := m["imageUrl"]; v != nil {
		t.Errorf("expected imageUrl to be nil got %#v", v)
		return
	}
	if v := m["status"]; v != "waiting" {
		t.Errorf(`expected status to be %q got %q`, "waiting", v)
		return
	}
	if v := m["joinToken"]; v != nil {
		t.Errorf("expected joinToken to be nil got %#v", v)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq == nil {
		t.Errorf("expected livekit send data to be set got %#v", nil)
	}
}

func TestParticipantRetrieve(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)

	meeting, participant := newMockMeetingAndParticipant(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/meetings/"+string(meeting.ID)+"/participants/"+string(participant.ID), nil)

	// create auth header
	cf := p.LiveKitConfig()
	at := auth.NewAccessToken(cf.APIKey, cf.APISecret)
	grant := &auth.VideoGrant{}

	at.AddGrant(grant).
		SetIdentity(string(participant.ID)).
		SetValidFor(2 * time.Minute)

	token, err := at.ToJWT()
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	// set auth header
	req.Header.Set("authorization", "Bearer "+token)

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.Participant
	err = json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID != participant.ID {
		t.Errorf("expected participant with ID %v got %v", participant.ID, m.ID)
		return
	}
}

func TestParticipantRetrieveBadAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)
	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)

	meeting, participant := newMockMeetingAndParticipant(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/meetings/"+string(meeting.ID)+"/participants/"+string(participant.ID), nil)

	// create auth header
	cf := p.LiveKitConfig()
	at := auth.NewAccessToken(cf.APIKey, cf.APISecret)
	grant := &auth.VideoGrant{}

	at.AddGrant(grant).
		SetIdentity(string(resource.NewResourceID())).
		SetValidFor(2 * time.Minute)

	token, err := at.ToJWT()
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	// set auth header
	req.Header.Set("authorization", "Bearer "+token)

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
	err = json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m["error"] != true {
		t.Errorf("expected error to be %v got %v", true, m["error"])
		return
	}
	msg := "The authorized identity does not match participant"
	if m["message"] != msg {
		t.Errorf("expected message to be %v got %v", msg, m["message"])
		return
	}
}

func TestParticipantUpdate(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	meeting, participant := newMockMeetingAndParticipant(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/meetings/"+string(meeting.ID)+"/participants/"+string(participant.ID), strings.NewReader(`{"status":"denied"}`))
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
	if v := m["meetingId"]; v == "" {
		t.Errorf("expected meetingId in response got %#v", v)
		return
	}
	if v := m["name"]; v != "Mock User" {
		t.Errorf(`expected name to be %q got %q`, "Mock User", v)
		return
	}
	if v := m["imageUrl"]; v == nil {
		t.Errorf("expected imageUrl in response got %#v", v)
		return
	}
	if v := m["status"]; v != "denied" {
		t.Errorf(`expected status to be %q got %q`, "denied", v)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq == nil {
		t.Errorf("expected livekit send data to be set got %#v", nil)
	}
}

func TestParticipantUpdateBadAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

	meeting, participant := newMockMeetingAndParticipant(ctx)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/meetings/"+string(meeting.ID)+"/participants/"+string(participant.ID), strings.NewReader(`{"status":"denied"}`))

	// create auth header
	cf := config.NewAuthConfigProvider().AuthConfig()
	token, err := jwt.NewBuilder().
		Issuer(cf.Issuer).
		Expiration(time.Now().Add(cf.TTL)).
		Claim("id", resource.NewResourceID()).
		Claim("userId", resource.NewResourceID()).
		Build()

	if err != nil {
		panic(fmt.Sprintf("error creating jwt: %s", err.Error()))
	}

	// sign access token
	signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
	if err != nil {
		panic(fmt.Sprintf("error signing jwt: %s", err.Error()))
	}

	// set auth header
	req.Header.Set("authorization", "Bearer "+string(signed))

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
	err = json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m["error"] != true {
		t.Errorf("expected error to be %v got %v", true, m["error"])
		return
	}
	msg := "Only meeting admins can update participants"
	if m["message"] != msg {
		t.Errorf(`expected message to be %q got %q`, msg, m["message"])
		return
	}
}
