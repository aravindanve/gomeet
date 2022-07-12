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
	authRefreshTokenTTL      = 90 * 24 * time.Hour
	AuthRefreshTokenCountMax = 50
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

func newAuth(userID ResourceID) (*Auth, error) {
	// create refresh token
	buf := make([]byte, 128)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(buf)
	refreshTokenExpiresAt := time.Now().Add(authRefreshTokenTTL)

	// create auth
	return &Auth{
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
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

func newAuthWithAccessToken(cf config.AuthConfig, auth *Auth) (*AuthWithAccessToken, error) {
	// create access token
	token, err := jwt.NewBuilder().
		Issuer(cf.Issuer).
		Expiration(time.Now().Add(cf.TTL)).
		Claim("id", auth.ID).
		Claim("userId", auth.UserID).
		Build()

	if err != nil {
		return nil, err
	}

	// sign access token
	signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
	if err != nil {
		return nil, err
	}

	return &AuthWithAccessToken{
		Auth:                 *auth,
		Scheme:               authSchemeBearer,
		AccessToken:          string(signed),
		AccessTokenExpiresAt: token.Expiration(),
	}, nil
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
		// index for expire
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "refreshTokenExpiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		})
		// index for gc sort
		if err == nil {
			_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{
					{Key: "userId", Value: 1},
					{Key: "refreshTokenExpiresAt", Value: 1},
					{Key: "_id", Value: 1},
				},
			})
		}

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
		{Key: "refreshTokenExpiresAt", Value: bson.D{{Key: "$gte", Value: time.Now()}}},
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
) error {
	if auth.ID == "" {
		now := time.Now()
		auth.CreatedAt = now
		auth.UpdatedAt = now

		r, err := c.collection.InsertOne(ctx, auth)
		if err != nil {
			return err
		}
		auth.ID = ResourceIDFromObjectID(r.InsertedID.(primitive.ObjectID))
		return nil
	} else {
		_id, err := auth.ID.ObjectID()
		if err != nil {
			return err
		}

		auth.UpdatedAt = time.Now()

		_, err = c.collection.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: _id},
		}, bson.D{
			{Key: "$set", Value: auth},
		})
		return err
	}
}

// keeps latest auths defined by max count
func (c *AuthCollection) gc(
	ctx context.Context, userID ResourceID, countMax int,
) error {
	cur, err := c.collection.Find(ctx, bson.D{
		{Key: "userId", Value: userID},
	}, options.Find().
		SetSort(bson.D{
			{Key: "refreshTokenExpiresAt", Value: -1},
			{Key: "_id", Value: -1},
		}).
		SetSkip(int64(countMax)).
		SetLimit(1),
	)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	var docs []Auth
	err = cur.All(ctx, &docs)
	if err != nil {
		return err
	}

	// return if no docs matched
	if len(docs) <= 0 {
		return nil
	}

	doc := docs[0]

	// delete all docs before the found doc
	_, err = c.collection.DeleteMany(ctx, bson.D{
		{Key: "userId", Value: userID},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "_id", Value: doc.ID}},
			bson.D{{Key: "refreshTokenExpiresAt", Value: bson.D{
				{Key: "$lt", Value: doc.RefreshTokenExpiresAt},
			}}},
			bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "refreshTokenExpiresAt", Value: bson.D{
					{Key: "$eq", Value: doc.RefreshTokenExpiresAt},
				}}},
				bson.D{{Key: "_id", Value: bson.D{
					{Key: "$lt", Value: doc.ID},
				}}},
			}}},
		}},
	})

	return err
}

type AuthController struct {
	AuthDeps
}

func NewAuthController(ds AuthDeps) *AuthController {
	return &AuthController{AuthDeps: ds}
}

type AuthCreateBody struct {
	GoogleIdToken string `json:"googleIdToken"`
}

func (c *AuthController) AuthCreateHandler(w http.ResponseWriter, r *http.Request) {
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
	err = c.UserCollection().Save(r.Context(), user)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// create auth
	auth, err := newAuth(user.ID)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// save auth
	err = c.AuthCollection().Save(r.Context(), auth)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// create response
	res, err := newAuthWithAccessToken(c.AuthConfig(), auth)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, res)

	// run gc
	go func(userID ResourceID) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c.AuthCollection().gc(ctx, userID, AuthRefreshTokenCountMax)
	}(auth.UserID)
}

type AuthRefreshBody struct {
	RefreshToken string `json:"refreshToken"`
}

func (c *AuthController) AuthRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// get auth id
	authID := mux.Vars(r)["authId"]
	if authID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing authId in request path")
		return
	}

	// decode body
	b := &AuthRefreshBody{}
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if b.RefreshToken == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing refresh token in request body")
		return
	}

	// find one by id
	authCurr, err := c.AuthCollection().FindOneByIDAndRefreshToken(r.Context(), ResourceID(authID), b.RefreshToken)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if authCurr == nil {
		util.WriteJSONError(w, http.StatusUnauthorized, "Authorization invalid or expired")
		return
	}

	// create auth next
	authNext, err := newAuth(authCurr.UserID)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// update auth next
	authNext.ID = authCurr.ID

	// save auth next
	err = c.AuthCollection().Save(r.Context(), authNext)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// create response
	res, err := newAuthWithAccessToken(c.AuthConfig(), authNext)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, res)
}

func RegisterAuthRoutes(r *mux.Router, ds AuthDeps) *mux.Router {
	c := NewAuthController(ds)

	r.HandleFunc("/auth", c.AuthCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/auth/{authId}/refresh", c.AuthRefreshHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
