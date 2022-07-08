package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AuthorizationDeps interface {
	config.AuthConfigProvider
}

type AuthorizationToken struct {
	jwt.Token
	AuthorizationClaims
}

type AuthorizationClaims struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

func GetAuthorization(r *http.Request, ds AuthorizationDeps) (*AuthorizationToken, error) {
	a := r.Header.Get("authorization")
	if a == "" {
		return nil, nil
	}
	s := strings.Split(a, " ")
	if len(s) < 2 || s[0] != "Bearer" {
		return nil, nil
	}
	signed := s[1]

	// parse and verify jwt
	cf := ds.AuthConfig()
	token, err := jwt.Parse(
		[]byte(signed),
		jwt.WithKey(cf.Algorithm, cf.Secret),
	)
	if err != nil {
		return nil, err
	}
	if token.Issuer() != cf.Issuer {
		return nil, fmt.Errorf("invalid issuer claim found in authorization")
	}

	// extract custom claims
	var claims AuthorizationClaims
	b, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("unable to encode id token to json")
	}
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, fmt.Errorf("unable to decode id token from json")
	}

	return &AuthorizationToken{
		Token:               token,
		AuthorizationClaims: claims,
	}, nil
}
