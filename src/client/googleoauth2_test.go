package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aravindanve/livemeet-server/src/config"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func TestNewGoogleOAuth2Client(t *testing.T) {
	t.Parallel()
	p := config.NewGoogleOAuth2ConfigProvider()
	var _ = NewGoogleOAuth2Client(p)
}

// see https://github.com/lestrrat-go/jwx/blob/develop/v2/docs/01-jwt.md#parse-and-verify-a-jwt-with-a-key-set-matching-kid
func TestGoogleOAuth2ClientVerifyIDTokenNoClaims(t *testing.T) {
	t.Parallel()
	pr := config.NewGoogleOAuth2ConfigProvider()
	cl := NewGoogleOAuth2Client(pr)

	// generate rsa key
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate rsa private key: %s\n", err)
		return
	}

	// create jwk
	key, err := jwk.FromRaw(raw)
	if err != nil {
		t.Errorf("failed to create jwk.Key from rsa private key: %s\n", err)
		return
	}

	// assign kid and set alg
	jwk.AssignKeyID(key)
	key.Set(jwk.AlgorithmKey, jwa.RS256)

	// create jwks
	set := jwk.NewSet()
	set.AddKey(key)

	// set jwks to cache
	p, err := jwk.PublicSetOf(set)
	if err != nil {
		t.Errorf("failed to get public key set from set: %s\n", err)
		return
	}
	cl.setKeySet("google", p, 5*time.Minute)

	// create the token
	token := jwt.New()
	token.Set(jwt.IssuerKey, "https://accounts.google.com")
	token.Set(jwt.AudienceKey, pr.GoogleOAuth2Config().ClientID)

	// sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Errorf("failed to generate signed token: %s\n", err)
		return
	}

	// test verify id token
	gtoken, err := cl.VerifyIDToken(context.Background(), string(signed))
	if err != nil {
		t.Errorf("failed to verify signed token: %s\n", err)
		return
	}

	// ensure email verified is false
	if v := gtoken.EmailVerified; v != false {
		t.Errorf("expected email_verifed to be %#v got %#v", false, v)
		return
	}
}

// see https://github.com/lestrrat-go/jwx/blob/develop/v2/docs/01-jwt.md#parse-and-verify-a-jwt-with-a-key-set-matching-kid
func TestGoogleOAuth2ClientVerifyIDTokenWithClaims(t *testing.T) {
	t.Parallel()
	pr := config.NewGoogleOAuth2ConfigProvider()
	cl := NewGoogleOAuth2Client(pr)

	// generate rsa key
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate rsa private key: %s\n", err)
		return
	}

	// create jwk
	key, err := jwk.FromRaw(raw)
	if err != nil {
		t.Errorf("failed to create jwk.Key from rsa private key: %s\n", err)
		return
	}

	// assign kid and set alg
	jwk.AssignKeyID(key)
	key.Set(jwk.AlgorithmKey, jwa.RS256)

	// create jwks
	set := jwk.NewSet()
	set.AddKey(key)

	// set jwks to cache
	p, err := jwk.PublicSetOf(set)
	if err != nil {
		t.Errorf("failed to get public key set from set: %s\n", err)
		return
	}
	cl.setKeySet("google", p, 5*time.Minute)

	// create the token
	token := jwt.New()
	token.Set(jwt.IssuerKey, "https://accounts.google.com")
	token.Set(jwt.AudienceKey, pr.GoogleOAuth2Config().ClientID)
	token.Set("email", "user@example.com")
	token.Set("email_verified", true)
	token.Set("given_name", "Aravindan")
	token.Set("family_name", "Ve")
	token.Set("picture", "https://example.com/image.jpg")

	// sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Errorf("failed to generate signed token: %s\n", err)
		return
	}

	// test verify id token
	gtoken, err := cl.VerifyIDToken(context.Background(), string(signed))
	if err != nil {
		t.Errorf("failed to verify signed token: %s\n", err)
		return
	}

	// ensure token data is correctly encoded
	if v := gtoken.Email; v != "user@example.com" {
		t.Errorf("expected email to be %#v got %#v", "user@example.com", v)
		return
	}
	if v := gtoken.EmailVerified; v != true {
		t.Errorf("expected email_verified to be %#v got %#v", true, v)
		return
	}
	if v := gtoken.GivenName; v != "Aravindan" {
		t.Errorf("expected given_name to be %#v got %#v", "Aravindan", v)
		return
	}
	if v := gtoken.FamilyName; v != "Ve" {
		t.Errorf("expected family_name to be %#v got %#v", "Ve", v)
		return
	}
	if v := gtoken.Picture; v == nil || *v != "https://example.com/image.jpg" {
		t.Errorf("expected picture to be %#v got %#v", "https://example.com/image.jpg", v)
		return
	}
}

// see https://github.com/lestrrat-go/jwx/blob/develop/v2/docs/01-jwt.md#parse-and-verify-a-jwt-with-a-key-set-matching-kid
func TestGoogleOAuth2ClientVerifyIDTokenRemoteJWKS(t *testing.T) {
	t.Parallel()
	pr := config.NewGoogleOAuth2ConfigProvider()
	cl := NewGoogleOAuth2Client(pr)

	// generate rsa key
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate rsa private key: %s\n", err)
		return
	}

	// create jwk
	key, err := jwk.FromRaw(raw)
	if err != nil {
		t.Errorf("failed to create jwk.Key from rsa private key: %s\n", err)
		return
	}

	// assign kid and set alg
	jwk.AssignKeyID(key)
	key.Set(jwk.AlgorithmKey, jwa.RS256)

	// create jwks
	set := jwk.NewSet()
	set.AddKey(key)

	// create the token
	token := jwt.New()
	token.Set(jwt.IssuerKey, "https://accounts.google.com")
	token.Set(jwt.AudienceKey, pr.GoogleOAuth2Config().ClientID)
	token.Set("email", "user@example.com")
	token.Set("email_verified", true)
	token.Set("given_name", "Aravindan")
	token.Set("family_name", "Ve")
	token.Set("picture", "https://example.com/image.jpg")

	// sign the token
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	if err != nil {
		t.Errorf("failed to generate signed token: %s\n", err)
		return
	}

	// create test jwks server
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := jwk.PublicSetOf(set)
		j, _ := json.Marshal(p)

		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}))
	defer srv.Close()

	// set http test client and url
	cl.setFetchClient(srv.Client())
	cl.setFetchURL(&srv.URL)

	// test verify id token
	gtoken, err := cl.VerifyIDToken(context.Background(), string(signed))
	if err != nil {
		t.Errorf("failed to verify signed token: %s\n", err)
		return
	}

	// ensure token data is correctly encoded
	if v := gtoken.Email; v != "user@example.com" {
		t.Errorf("expected email to be %#v got %#v", "user@example.com", v)
		return
	}
	if v := gtoken.EmailVerified; v != true {
		t.Errorf("expected email_verified to be %#v got %#v", true, v)
		return
	}
	if v := gtoken.GivenName; v != "Aravindan" {
		t.Errorf("expected given_name to be %#v got %#v", "Aravindan", v)
		return
	}
	if v := gtoken.FamilyName; v != "Ve" {
		t.Errorf("expected family_name to be %#v got %#v", "Ve", v)
		return
	}
	if v := gtoken.Picture; v == nil || *v != "https://example.com/image.jpg" {
		t.Errorf("expected picture to be %#v got %#v", "https://example.com/image.jpg", v)
		return
	}
}
