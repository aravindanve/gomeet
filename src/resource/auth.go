package resource

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/util"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	authRefreshTokenTTL = 90 * 24 * time.Hour
)

const (
	authSchemeBearer AuthScheme = "Bearer"
)

type AuthScheme string

type AuthDeps interface {
	config.AuthConfigProvider
	client.GoogleOAuth2ClientProvider
	AuthCollectionProvider
	UserCollectionProvider
}

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
	ID     ResourceID `json:"id"`
	UserID ResourceID `json:"userId"`
}

type AuthCollectionProvider interface {
	AuthCollection() *AuthCollection
}

type AuthCollection struct {
	collection *mongo.Collection
}

func NewAuthCollection(ctx context.Context, db *mongo.Database) *AuthCollection {
	collection := db.Collection("auth")

	// create indexes
	go func() {
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "refreshTokenExpiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		})

		if err != nil {
			msg := fmt.Sprintf("error creating mongo indexes: %s", err.Error())
			if os.Getenv("APP_ENV") == "testing" {
				log.Println(msg) // do not panic in tests
			} else {
				panic(msg)
			}
		}
	}()

	return &AuthCollection{collection: collection}
}

func (c *AuthCollection) FindOneByIDAndRefreshToken(
	ctx context.Context, id ResourceID, refreshToken string,
) (*Auth, error) {
	_id, err := id.ObjectID()
	if err != nil {
		return nil, err
	}

	var auth Auth
	err = c.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: _id},
		{Key: "refreshToken", Value: refreshToken},
	}).Decode(&auth)

	if err != nil && err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &auth, nil
}

func (c *AuthCollection) DeleteOneByID(
	ctx context.Context, id ResourceID, refreshToken string,
) (*Auth, error) {
	_id, err := id.ObjectID()
	if err != nil {
		return nil, err
	}

	var auth Auth
	_, err = c.collection.DeleteOne(ctx, bson.D{
		{Key: "_id", Value: _id},
		{Key: "refreshToken", Value: refreshToken},
	})

	if err != nil {
		return nil, err
	}

	return &auth, nil
}

func (c *AuthCollection) Save(
	ctx context.Context, auth *Auth,
) (*Auth, error) {
	if auth.ID == "" {
		now := time.Now()
		auth.CreatedAt = now
		auth.UpdatedAt = now

		r, err := c.collection.InsertOne(ctx, auth)
		if err != nil {
			return nil, err
		}
		auth.ID = ResourceIDFromObjectID(r.InsertedID.(primitive.ObjectID))
		return auth, nil
	} else {
		_id, err := auth.ID.ObjectID()
		if err != nil {
			return nil, err
		}

		auth.UpdatedAt = time.Now()

		_, err = c.collection.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: _id},
		}, bson.D{
			{Key: "$set", Value: auth},
		})
		if err != nil {
			return nil, err
		}
		return auth, nil
	}
}

type authController struct {
	AuthDeps
}

func NewAuthController(ds AuthDeps) *authController {
	return &authController{AuthDeps: ds}
}

type AuthCreateBody struct {
	GoogleIdToken string `json:"googleIdToken"`
}

func (c *authController) AuthCreateHandler(w http.ResponseWriter, r *http.Request) {
	// decode body
	b := &AuthCreateBody{}

	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if b.GoogleIdToken == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing google id token in request body")
		return
	}

	// verify id token
	gtoken, err := c.GoogleOAuth2Client().VerifyIDToken(r.Context(), b.GoogleIdToken)
	if err != nil {
		util.WriteJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// ensure email verified
	if !gtoken.EmailVerified {
		util.WriteJSONError(w, http.StatusBadRequest, "Your email is not verified")
		return
	}

	// extract provider resource id
	provider, providerResourceID := UserProviderGoogle, gtoken.Subject()

	// find user
	user, err := c.UserCollection().FindOneByProviderResourceID(r.Context(), provider, providerResourceID)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// create user if not exists
	if user == nil {
		user = &User{
			Name:               "User",
			Provider:           UserProviderGoogle,
			ProviderResourceID: gtoken.Subject(),
		}
	}

	// update user name and image url
	if name := strings.Trim(gtoken.GivenName+" "+gtoken.FamilyName, " "); name != "" {
		user.Name = name
	}
	if imageURL := gtoken.Picture; imageURL != nil {
		user.ImageURL = imageURL
	}

	// save user
	c.UserCollection().Save(r.Context(), user)

	// create refresh token
	buf := make([]byte, 128)
	_, err = rand.Read(buf)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(buf)
	refreshTokenExpiresAt := time.Now().Add(authRefreshTokenTTL)

	// create auth
	auth := &Auth{
		UserID:                user.ID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}

	// save auth
	c.AuthCollection().Save(r.Context(), auth)

	// create access token
	cf := c.AuthConfig()
	token, err := jwt.NewBuilder().
		Issuer(cf.Issuer).
		Expiration(time.Now().Add(cf.TTL)).
		Claim("id", auth.ID).
		Claim("userId", user.ID).
		Build()

	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// sign access token
	signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// create response
	response := &AuthWithAccessToken{
		Auth:                 *auth,
		Scheme:               authSchemeBearer,
		AccessToken:          string(signed),
		AccessTokenExpiresAt: token.Expiration(),
	}

	util.WriteJSONResponse(w, http.StatusOK, response)
}

type AuthRefreshBody struct {
	RefreshToken string `json:"refreshToken"`
}

func (c *authController) AuthRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	// params := mux.Vars(r)
}

func RegisterAuthRoutes(r *mux.Router, ds AuthDeps) *mux.Router {
	c := NewAuthController(ds)

	r.HandleFunc("/auth", c.AuthCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/auth/{authId}", c.AuthRefreshHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
