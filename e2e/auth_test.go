package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aravindanve/livemeet-server/src/client"
	"github.com/aravindanve/livemeet-server/src/config"
	"github.com/aravindanve/livemeet-server/src/provider"
	"github.com/aravindanve/livemeet-server/src/resource"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mockAuthHeader *string
var mockAuthHeaderMut sync.Mutex

func getMockAuthHeader() string {
	mockAuthHeaderMut.Lock()
	if mockAuthHeader == nil {
		// create mock auth header
		cf := config.NewAuthConfigProvider().AuthConfig()
		user := getMockUser()
		token, err := jwt.NewBuilder().
			Issuer(cf.Issuer).
			Expiration(time.Now().Add(cf.TTL)).
			Claim("id", resource.ResourceIDFromObjectID(primitive.NewObjectID())).
			Claim("userId", user.ID).
			Build()

		if err != nil {
			panic(fmt.Sprintf("error creating jwt: %s", err.Error()))
		}

		// sign access token
		signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
		if err != nil {
			panic(fmt.Sprintf("error signing jwt: %s", err.Error()))
		}

		header := "Bearer " + string(signed)
		mockAuthHeader = &header
	}
	mockAuthHeaderMut.Unlock()
	return *mockAuthHeader
}

type mockAuthProvider struct {
	resource.AuthDeps
	Release       func(ctx context.Context)
	mongoDatabase *mongo.Database
}

func newMockAuthProvider(ctx context.Context) *mockAuthProvider {
	p := provider.NewProvider(ctx)
	return &mockAuthProvider{AuthDeps: p, Release: p.Release, mongoDatabase: p.MongoDatabase()}
}

func (m *mockAuthProvider) GoogleOAuth2Client() client.GoogleOAuth2Client {
	return newMockGoogleOAuth2Client()
}

type mockGoogleOAuth2Client struct {
	client.GoogleOAuth2Client
}

func newMockGoogleOAuth2Client() client.GoogleOAuth2Client {
	return &mockGoogleOAuth2Client{}
}

func (m *mockGoogleOAuth2Client) VerifyIDToken(ctx context.Context, signed string) (*client.GoogleOAuth2Token, error) {
	user := getMockUser()
	token := jwt.New()
	token.Set(jwt.SubjectKey, user.ProviderResourceID)

	return &client.GoogleOAuth2Token{
		Token: token,
		GoogleOAuth2Claims: client.GoogleOAuth2Claims{
			Email:         "user@example.com",
			EmailVerified: true,
			Picture:       nil,
		},
	}, nil
}

func TestAuthCreate(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockAuthProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterAuthRoutes(mux.NewRouter(), p)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"googleIdToken":"token"}`))

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.AuthWithAccessToken
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
	if m.Scheme == "" {
		t.Errorf("expected scheme in response got %#v", m.Scheme)
		return
	}
	if m.AccessToken == "" {
		t.Errorf("expected accessToken in response got %#v", m.AccessToken)
		return
	}
	if m.RefreshToken == "" {
		t.Errorf("expected refreshToken in response got %#v", m.RefreshToken)
		return
	}
}

func TestAuthRefresh(t *testing.T) {
	t.Parallel()
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockAuthProvider(ctx)
	defer p.Release(ctx)

	user := getMockUser()

	// create auth
	auth := &resource.Auth{
		UserID:                user.ID,
		RefreshToken:          "token",
		RefreshTokenExpiresAt: time.Now().Add(30 * time.Minute),
	}

	// save auth
	err := p.AuthCollection().Save(ctx, auth)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	r := resource.RegisterAuthRoutes(mux.NewRouter(), p)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/auth/"+string(auth.ID)+"/refresh", strings.NewReader(`{"refreshToken":"token"}`))

	// test route
	r.ServeHTTP(w, req)

	// test status code
	s := w.Result().StatusCode
	if s != http.StatusOK {
		t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
		return
	}

	// test response
	var m resource.AuthWithAccessToken
	err = json.NewDecoder(w.Result().Body).Decode(&m)
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
	if m.Scheme == "" {
		t.Errorf("expected scheme in response got %#v", m.Scheme)
		return
	}
	if m.AccessToken == "" {
		t.Errorf("expected accessToken in response got %#v", m.AccessToken)
		return
	}
	if m.RefreshToken == "" {
		t.Errorf("expected refreshToken in response got %#v", m.RefreshToken)
		return
	}
}

func TestAuthGC(t *testing.T) {
	// dont run in parallel
	defer panicGuard(t)
	ctx, cancel := newTestContext()
	defer cancel()
	p := newMockAuthProvider(ctx)
	defer p.Release(ctx)

	r := resource.RegisterAuthRoutes(mux.NewRouter(), p)

	resps := make([]resource.AuthWithAccessToken, 0)

	// create enough tokens to force gc
	for i := 0; i < resource.AuthRefreshTokenCountMax+1; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"googleIdToken":"token"}`))
		r.ServeHTTP(w, req)

		// test status code
		s := w.Result().StatusCode
		if s != http.StatusOK {
			t.Errorf("expected status to be %#v got %#v", http.StatusOK, s)
			return
		}

		var m resource.AuthWithAccessToken
		err := json.NewDecoder(w.Result().Body).Decode(&m)
		if err != nil {
			t.Errorf("expected error to be nil got %#v", err)
			return
		}

		resps = append(resps, m)
	}

	// wait for gc to complete
	time.Sleep(2 * time.Second)

	user := getMockUser()

	// test oldest auth
	cur, err := p.mongoDatabase.Collection("auth").Find(ctx, bson.D{
		{Key: "userId", Value: user.ID},
	}, options.Find().
		SetSort(bson.D{
			{Key: "refreshTokenExpiresAt", Value: 1},
			{Key: "_id", Value: 1},
		}).
		SetLimit(1))

	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	cur.Next(ctx)

	var oldest resource.Auth
	err = cur.Decode(&oldest)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	if oldest.ID != resps[1].ID {
		t.Errorf("expected oldest to be %v got %v", resps[1].ID, oldest.ID)
		return
	}

	cur.Close(ctx)

	// test newest auth
	cur, err = p.mongoDatabase.Collection("auth").Find(ctx, bson.D{
		{Key: "userId", Value: user.ID},
	}, options.Find().
		SetSort(bson.D{
			{Key: "refreshTokenExpiresAt", Value: -1},
			{Key: "_id", Value: -1},
		}).
		SetLimit(1))

	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	cur.Next(ctx)

	var newest resource.Auth
	err = cur.Decode(&newest)
	if err != nil {
		t.Errorf("expected error to be nil got %#v", err)
		return
	}

	if oldest.ID != resps[1].ID {
		t.Errorf("expected oldest to be %v got %v", resps[len(resps)-1].ID, oldest.ID)
		return
	}

	cur.Close(ctx)

	// test gced auth
	var gced resource.Auth
	err = p.mongoDatabase.Collection("auth").FindOne(ctx, bson.D{
		{Key: "_id", Value: resps[0].ID},
	}).Decode(&gced)

	if err != mongo.ErrNoDocuments {
		t.Errorf("expected error to be %#v got %#v, %#v", mongo.ErrNoDocuments, err, gced)
	}
}
