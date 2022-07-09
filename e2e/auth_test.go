package main_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type mockAuthDeps struct {
	resource.AuthDeps
}

func newMockAuthDeps(ctx context.Context) resource.AuthDeps {
	return &mockAuthDeps{AuthDeps: provider.NewProvider(ctx)}
}

func (m *mockAuthDeps) GoogleOAuth2Client() client.GoogleOAuth2Client {
	return &mockGoogleOAuth2Client{}
}

type mockGoogleOAuth2Client struct{}

func (m *mockGoogleOAuth2Client) VerifyIDToken(ctx context.Context, signed string) (*client.GoogleOAuth2Token, error) {
	return &client.GoogleOAuth2Token{
		Token: jwt.New(),
		GoogleOAuth2Claims: client.GoogleOAuth2Claims{
			Email:         "user@example.com",
			EmailVerified: true,
			Picture:       nil,
		},
	}, nil
}

func TestAuthCreate(t *testing.T) {
	t.Parallel()
	defer PanicGuard(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d := newMockAuthDeps(ctx)
	r := resource.RegisterAuthRoutes(mux.NewRouter(), d)
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
	var m map[string]string
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
	if v := m["scheme"]; v == "" {
		t.Errorf("expected scheme in response got %#v", v)
		return
	}
	if v := m["accessToken"]; v == "" {
		t.Errorf("expected accessToken in response got %#v", v)
		return
	}
	if v := m["refreshToken"]; v == "" {
		t.Errorf("expected refreshToken in response got %#v", v)
		return
	}
}
