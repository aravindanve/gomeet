package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/aravindanve/gomeet-server/src/config"
)

const (
	googleOAuth2JWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
	googleOAuth2JWKSTTL = 1 * time.Hour
)

type GoogleOAuth2Token struct {
	jwt.Token
	GoogleOAuth2Claims
}

type GoogleOAuth2Claims struct {
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	GivenName     string  `json:"given_name"`
	FamilyName    string  `json:"family_name"`
	Picture       *string `json:"picture"`
}

type GoogleOAuth2ClientDeps interface {
	config.GoogleOAuth2ConfigProvider
}

type GoogleOAuth2ClientProvider interface {
	GoogleOAuth2Client() GoogleOAuth2Client
}

type GoogleOAuth2Client interface {
	VerifyIDToken(ctx context.Context, signed string) (*GoogleOAuth2Token, error)
}

type googleOAuth2Client struct {
	config      config.GoogleOAuth2Config
	cache       *ttlcache.Cache[string, jwk.Set]
	fetchClient *http.Client
	fetchURL    *string
}

func NewGoogleOAuth2Client(ds GoogleOAuth2ClientDeps) *googleOAuth2Client {
	return &googleOAuth2Client{
		config: ds.GoogleOAuth2Config(),
		cache: ttlcache.New(
			ttlcache.WithTTL[string, jwk.Set](googleOAuth2JWKSTTL),
			ttlcache.WithCapacity[string, jwk.Set](1),
		),
	}
}

func (s *googleOAuth2Client) SetFetchClient(client *http.Client) *googleOAuth2Client {
	s.fetchClient = client
	return s
}

func (s *googleOAuth2Client) SetFetchURL(url *string) *googleOAuth2Client {
	s.fetchURL = url
	return s
}

func (s *googleOAuth2Client) VerifyIDToken(ctx context.Context, signed string) (*GoogleOAuth2Token, error) {
	var keyset jwk.Set

	// set jwks fetch fetchOptions
	var fetchOptions []jwk.FetchOption
	if s.fetchClient != nil {
		fetchOptions = append(fetchOptions, jwk.WithHTTPClient(s.fetchClient))
	}

	// set jwks fetch url
	var fetchURL = googleOAuth2JWKSURL
	if s.fetchURL != nil {
		fetchURL = *s.fetchURL
	}

	// get jwks from cache or fetch
	item := s.cache.Get("google")
	if item != nil {
		keyset = item.Value()
	} else if set, err := jwk.Fetch(ctx, fetchURL, fetchOptions...); err == nil {
		keyset = set
		s.cache.Set("google", set, googleOAuth2JWKSTTL)
	} else {
		return nil, err
	}

	// parse and verify jwt
	token, err := jwt.Parse(
		[]byte(signed),
		jwt.WithKeySet(keyset, jws.WithInferAlgorithmFromKey(true)),
	)
	if err != nil {
		return nil, err
	}
	if token.Issuer() != "https://accounts.google.com" && token.Issuer() != "accounts.google.com" {
		return nil, fmt.Errorf("invalid issuer claim found in id_token")
	}
	if strings.Join(token.Audience(), ",") != s.config.ClientID {
		return nil, fmt.Errorf("invalid audience claim found in id_token")
	}

	// extract custom claims
	var claims GoogleOAuth2Claims
	b, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("unable to encode id token to json")
	}
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, fmt.Errorf("unable to decode id token from json")
	}

	return &GoogleOAuth2Token{
		Token:              token,
		GoogleOAuth2Claims: claims,
	}, nil
}
