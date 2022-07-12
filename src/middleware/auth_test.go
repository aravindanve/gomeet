package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/net/context"
)

func MockAuthMiddleware(t *testing.T) {
	t.Parallel()
	ds := config.NewAuthConfigProvider()
	cf := ds.AuthConfig()

	// create the token
	token := jwt.New()
	token.Set(jwt.IssuerKey, cf.Issuer)
	token.Set("id", "some-id")
	token.Set("userId", "some-user-id")

	// sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
	if err != nil {
		t.Errorf("failed to generate signed token: %s\n", err)
		return
	}

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("authorization", "Bearer "+string(signed))

	// create recorder
	w := httptest.NewRecorder()

	// create middleware
	m := AuthMiddleware(ds)
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetAuthToken(r)
		if err != nil {
			t.Errorf("error gettting auth token: %s", err.Error())
			return
		}
		if token == nil {
			t.Error("expected auth token got nil")
			return
		}
		if token.ID != "some-id" {
			t.Errorf("expected id in payload to be %v got %v\n", "some-id", token.ID)
			return
		}
		if token.UserID != "some-user-id" {
			t.Errorf("expected id in payload to be %v got %v\n", "some-user-id", token.UserID)
			return
		}
	})).ServeHTTP(w, r)
}

func MockAuthMiddlewareBadToken(t *testing.T) {
	t.Parallel()
	ds := config.NewAuthConfigProvider()
	cf := ds.AuthConfig()

	// create the token
	token := jwt.New()
	token.Set(jwt.IssuerKey, cf.Issuer)
	token.Set("id", "some-id")
	token.Set("userId", "some-user-id")

	// sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, append(cf.Secret, 'a')))
	if err != nil {
		t.Errorf("failed to generate signed token: %s\n", err)
		return
	}

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("authorization", "Bearer "+string(signed))

	// create recorder
	w := httptest.NewRecorder()

	// create middleware
	m := AuthMiddleware(ds)
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := GetAuthToken(r)
		if err == nil || err.Error() != "could not verify message using any of the signatures or keys" {
			t.Errorf("expected token verify error got: %#v", err)
			return
		}
	})).ServeHTTP(w, r)
}

func TestGetAuthTokenParsed(t *testing.T) {
	t.Parallel()
	ds := config.NewAuthConfigProvider()
	cf := ds.AuthConfig()

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), authContextKey{}, &authContext{
		parsed: true,
		config: cf,
		token: &AuthToken{
			Token: jwt.New(),
			AuthClaims: AuthClaims{
				ID:     "some-id",
				UserID: "some-user-id",
			},
		},
		err: nil,
		raw: "",
	}))

	token, err := GetAuthToken(r)
	if err != nil {
		t.Errorf("error gettting auth token: %s", err.Error())
		return
	}
	if token == nil {
		t.Error("expected auth token got nil")
		return
	}
	if token.ID != "some-id" {
		t.Errorf("expected id in payload to be %v got %v\n", "some-id", token.ID)
		return
	}
	if token.UserID != "some-user-id" {
		t.Errorf("expected id in payload to be %v got %v\n", "some-user-id", token.UserID)
		return
	}
}

func TestGetAuthTokenNotInitialized(t *testing.T) {
	t.Parallel()

	// create request
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	defer func() {
		if r := recover(); r != "auth middleware not initialized!" {
			t.Errorf("expected panic 'auth middleware not initialized!' got %s", r)
			return
		}
	}()

	GetAuthToken(r)
	t.Error("expected panic 'auth middleware not initalized'")
}
