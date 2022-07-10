package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AuthMiddlewareDeps interface {
	config.AuthConfigProvider
}

type AuthToken struct {
	jwt.Token
	AuthClaims
}

type AuthClaims struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type AuthContext interface {
	Token() (*AuthToken, error)
}

type authContext struct {
	parsed bool
	config config.AuthConfig
	token  *AuthToken
	err    error
	raw    string
}

func NewAuthContext(c config.AuthConfig, rawToken string) AuthContext {
	return &authContext{
		parsed: false,
		config: c,
		token:  nil,
		err:    nil,
		raw:    rawToken,
	}
}

func (a *authContext) Token() (*AuthToken, error) {
	switch {
	case !a.parsed:
		a.parsed = true
		v := a.raw
		if v == "" {
			break
		}
		s := strings.Split(v, " ")
		if len(s) < 2 || s[0] != "Bearer" {
			break
		}
		signed := s[1]

		// parse and verify jwt
		token, err := jwt.Parse(
			[]byte(signed),
			jwt.WithKey(a.config.Algorithm, a.config.Secret),
		)
		if err != nil {
			a.err = err
			break
		}
		if token.Issuer() != a.config.Issuer {
			a.err = fmt.Errorf("invalid issuer claim found in authorization")
			break
		}

		// extract custom claims
		var claims AuthClaims
		b, err := json.Marshal(token)
		if err != nil {
			a.err = fmt.Errorf("unable to encode id token to json")
			break
		}
		err = json.Unmarshal(b, &claims)
		if err != nil {
			a.err = fmt.Errorf("unable to decode id token from json")
			break
		}

		a.token = &AuthToken{
			Token:      token,
			AuthClaims: claims,
		}
	}
	if a.token != nil {
		tokenCopy := *a.token
		return &tokenCopy, a.err
	}
	return nil, a.err
}

type authContextKey struct{}

func GetAuthToken(r *http.Request) (*AuthToken, error) {
	v := r.Context().Value(authContextKey{})
	if a, ok := v.(*authContext); ok {
		return a.Token()
	}
	panic("auth middleware not initialized!")
}

func AuthMiddleware(ds AuthMiddlewareDeps) mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := NewAuthContext(ds.AuthConfig(), r.Header.Get("authorization"))
			n := r.WithContext(context.WithValue(r.Context(), authContextKey{}, a))

			handler.ServeHTTP(w, n)
		})
	}
}
