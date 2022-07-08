package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func TestGetAuthorization(t *testing.T) {
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

	// test get authorization
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("authorization", "Bearer "+string(signed))
	a, err := GetAuthorization(r, ds)
	if err != nil {
		t.Errorf("error when trying to get authorization: %s\n", err)
		return
	}
	if a == nil {
		t.Errorf("get nil when trying to get authorization: %s\n", err)
		return
	}
	if a.ID != "some-id" {
		t.Errorf("expected id in payload to be %v got %v\n", "some-id", a.ID)
		return
	}
	if a.UserID != "some-user-id" {
		t.Errorf("expected id in payload to be %v got %v\n", "some-user-id", a.UserID)
		return
	}
}
