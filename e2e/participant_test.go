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

	"github.com/aravindanve/livemeet-server/src/client"
	"github.com/aravindanve/livemeet-server/src/config"
	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/provider"
	"github.com/aravindanve/livemeet-server/src/resource"
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
		Status:    resource.ParticipantStatus_Waiting,
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

	meeting := newMockMeeting(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

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
	var m resource.ParticipantWithRoomTokens
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID == "" {
		t.Errorf("expected id in response got %#v", m.ID)
		return
	}
	if m.MeetingID == "" {
		t.Errorf("expected meetingId in response got %#v", m.MeetingID)
		return
	}
	if m.Name != "Mock User" {
		t.Errorf(`expected name to be %q got %q`, "Mock User", m.Name)
		return
	}
	if m.ImageURL == nil {
		t.Errorf("expected imageUrl in response got %#v", m.ImageURL)
		return
	}
	if m.Status != resource.ParticipantStatus_Admitted {
		t.Errorf(`expected status to be %q got %q`, resource.ParticipantStatus_Admitted, m.Status)
		return
	}
	if len(m.RoomTokens) != 2 {
		t.Errorf("expected roomTokens to have 2 items got %#v", len(m.RoomTokens))
		return
	}
	if m.RoomTokens[0].RoomName == "" {
		t.Errorf("expected roomName in room token 0 got %#v", m.RoomTokens[0].RoomName)
		return
	}
	if m.RoomTokens[0].RoomType != resource.RoomType_Conference {
		t.Errorf(`expected roomType in room token 0 to be %q got %q`, resource.RoomType_Conference, m.RoomTokens[0].RoomType)
		return
	}
	if m.RoomTokens[0].AccessToken == "" {
		t.Errorf("expected accessToken in room toke 0 got %#v", m.RoomTokens[0].AccessToken)
		return
	}
	if m.RoomTokens[1].RoomName == "" {
		t.Errorf("expected roomName in room token 1 got %#v", m.RoomTokens[0].RoomName)
		return
	}
	if m.RoomTokens[1].RoomType != resource.RoomType_Waiting {
		t.Errorf(`expected roomType in room token 1 to be %q got %q`, resource.RoomType_Waiting, m.RoomTokens[0].RoomType)
		return
	}
	if m.RoomTokens[1].AccessToken == "" {
		t.Errorf("expected accessToken in room token 1 got %#v", m.RoomTokens[0].AccessToken)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq != nil {
		t.Errorf("expected livekit send data to be nil got %#v", p.livekitClient.sendDataReq)
	}
}

func TestParticipantCreateNoAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	meeting := newMockMeeting(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

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
	var m resource.ParticipantWithRoomTokens
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID == "" {
		t.Errorf("expected id in response got %#v", m.ID)
		return
	}
	if m.MeetingID == "" {
		t.Errorf("expected meetingId in response got %#v", m.MeetingID)
		return
	}
	if m.Name != "My Name" {
		t.Errorf(`expected name to be %q got %q`, "Mock User", m.Name)
		return
	}
	if m.ImageURL != nil {
		t.Errorf("expected imageUrl to be nil got %#v", m.ImageURL)
		return
	}
	if m.Status != resource.ParticipantStatus_Waiting {
		t.Errorf(`expected status to be %q got %q`, resource.ParticipantStatus_Waiting, m.Status)
		return
	}
	if len(m.RoomTokens) != 1 {
		t.Errorf("expected roomTokens to have 1 items got %#v", len(m.RoomTokens))
		return
	}
	if m.RoomTokens[0].RoomName == "" {
		t.Errorf("expected roomName in room token 0 got %#v", m.RoomTokens[0].RoomName)
		return
	}
	if m.RoomTokens[0].RoomType != resource.RoomType_Waiting {
		t.Errorf(`expected roomType in room token 0 to be %q got %q`, resource.RoomType_Waiting, m.RoomTokens[0].RoomType)
		return
	}
	if m.RoomTokens[0].AccessToken == "" {
		t.Errorf("expected accessToken in room toke 0 got %#v", m.RoomTokens[0].AccessToken)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq != nil {
		t.Errorf("expected livekit send data to be nil got %#v", p.livekitClient.sendDataReq)
	}
}

func TestParticipantRetrieve(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := provider.NewProvider(ctx)
	defer p.Release(ctx)

	meeting, participant := newMockMeetingAndParticipant(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
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
	if m.Status != resource.ParticipantStatus_Waiting {
		t.Errorf(`expected status to be %q got %q`, resource.ParticipantStatus_Waiting, m.Status)
		return
	}

	// test resource exists
	doc, err := p.ParticipantCollection().FindOneByID(ctx, participant.ID)
	if err != nil {
		t.Errorf("unexpected error retrieving data %s", err.Error())
	}
	if doc == nil {
		t.Errorf("expected doc in mongodb got %#v", doc)
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

	meeting, participant := newMockMeetingAndParticipant(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
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

	meeting, participant := newMockMeetingAndParticipant(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

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
	var m resource.Participant
	err := json.NewDecoder(w.Result().Body).Decode(&m)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}
	if m.ID == "" {
		t.Errorf("expected id in response got %#v", m.ID)
		return
	}
	if m.MeetingID == "" {
		t.Errorf("expected meetingId in response got %#v", m.MeetingID)
		return
	}
	if m.Name != "Mock User" {
		t.Errorf(`expected name to be %q got %q`, "Mock User", m.Name)
		return
	}
	if m.ImageURL == nil {
		t.Errorf("expected imageUrl in response got %#v", m.ImageURL)
		return
	}
	if m.Status != resource.ParticipantStatus_Denied {
		t.Errorf(`expected status to be %q got %q`, resource.ParticipantStatus_Denied, m.Status)
		return
	}

	// test livekit send data
	if p.livekitClient.sendDataReq == nil {
		t.Errorf("expected livekit send data to be set got %#v", p.livekitClient.sendDataReq)
	}
}

func TestParticipantUpdateBadAuth(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockParticipantProvider(ctx)
	defer p.Release(ctx)

	meeting, participant := newMockMeetingAndParticipant(ctx)

	r := resource.RegisterParticipantRoutes(mux.NewRouter(), p)
	r.Use(middleware.AuthMiddleware(p))

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
