package resource

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	AuthSchemeBearer AuthScheme = "Bearer"
)

type AuthScheme string

type Auth struct {
	ID                    ResourceID `json:"id" bson:"_id,omitempty"`
	UserID                ResourceID `json:"userId" bson:"userId"`
	RefreshToken          string     `json:"refreshToken" bson:"refreshToken"`
	RefreshTokenExpiresAt time.Time  `json:"refreshTokenExpiresAt" bson:"refreshTokenExpiresAt"`
	CreatedAt             time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt" bson:"updatedAt"`
}

type AuthWithAccessToken struct {
	Auth
	Scheme               AuthScheme `json:"scheme"`
	AccessToken          string     `json:"accessToken"`
	AccessTokenExpiresAt time.Time  `json:"accessTokenExpiresAt"`
}

type AuthAccessTokenPayload struct {
	Id     ResourceID `json:"id"`
	UserId ResourceID `json:"userId"`
}

type AuthCreateBody struct {
	GoogleCode string `json:"googleCode"`
}

type AuthRefreshBody struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthController struct {
	collection *mongo.Collection
}

func (c *AuthController) AuthCreateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *AuthController) AuthRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func NewAuthRouter() *mux.Router {
	r := mux.NewRouter()
	c := NewAuthController()

	r.HandleFunc("/auth", c.AuthCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/auth/{authId}", c.AuthRefreshHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
