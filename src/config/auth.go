package config

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
)

const (
	AuthAccessTokenTTL    = 24 * time.Hour
	AuthAccessTokenIssuer = "https://github.com/aravindanve/gomeet-server"
)

type AuthConfig struct {
	Algorithm jwa.SignatureAlgorithm
	Secret    []byte
	Issuer    string
	TTL       time.Duration
}

type AuthConfigProvider interface {
	AuthConfig() AuthConfig
}

type authConfigProvider struct {
	authConfig AuthConfig
}

func (p *authConfigProvider) AuthConfig() AuthConfig {
	return p.authConfig
}

func NewAuthConfigProvider() AuthConfigProvider {
	return &authConfigProvider{
		authConfig: AuthConfig{
			Algorithm: jwa.HS512,
			Secret:    GetenvStringBase64("AUTH_SECRET"),
			Issuer:    AuthAccessTokenIssuer,
			TTL:       AuthAccessTokenTTL,
		},
	}
}
