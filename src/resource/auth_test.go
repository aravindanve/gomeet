package resource

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newAuthAndJSON() (AuthWithAccessToken, []byte) {
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var a = AuthWithAccessToken{
		Auth: Auth{
			ID:                    "some-id",
			UserID:                "some-id",
			RefreshToken:          "some-token",
			RefreshTokenExpiresAt: t,
			CreatedAt:             t,
			UpdatedAt:             t,
		},
		Scheme:               AuthScheme_Bearer,
		AccessToken:          "some-token",
		AccessTokenExpiresAt: t,
	}

	var j = []byte(`{"id":"some-id","userId":"some-id","refreshToken":"some-token",` +
		`"refreshTokenExpiresAt":"2022-01-01T00:00:00Z","createdAt":"2022-01-01T00:00:00Z",` +
		`"updatedAt":"2022-01-01T00:00:00Z","scheme":"Bearer","accessToken":"some-token",` +
		`"accessTokenExpiresAt":"2022-01-01T00:00:00Z"}`)

	return a, j
}

func TestAuthMarshalJSON(t *testing.T) {
	t.Parallel()
	a, j := newAuthAndJSON()

	value, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func TestAuthUnmarshalJSON(t *testing.T) {
	t.Parallel()
	a, j := newAuthAndJSON()

	var value AuthWithAccessToken
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if a != value {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}

func newAuthAndBSON() (Auth, []byte) {
	var o = primitive.NewObjectID()
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var a = Auth{
		ID:                    ResourceIDFromObjectID(o),
		UserID:                ResourceIDFromObjectID(o),
		RefreshToken:          "some-token",
		RefreshTokenExpiresAt: t,
		CreatedAt:             t,
		UpdatedAt:             t,
	}

	var d = primitive.NewDateTimeFromTime(t)
	var b, _ = bson.Marshal(bson.D{
		{Key: "_id", Value: o},
		{Key: "userId", Value: o},
		{Key: "refreshToken", Value: "some-token"},
		{Key: "refreshTokenExpiresAt", Value: d},
		{Key: "createdAt", Value: d},
		{Key: "updatedAt", Value: d},
	})

	return a, b
}

func TestAuthMarshalBSON(t *testing.T) {
	t.Parallel()
	a, b := newAuthAndBSON()

	value, err := bson.Marshal(a)
	if err != nil {
		t.Fatalf("Error marshalling bson: %#v", err)
	}
	if !bytes.Equal(b, value) {
		t.Fatalf("Unexpected marshalled bson: %#v", string(value))
	}
}

func TestAuthUnmarshalBSON(t *testing.T) {
	t.Parallel()
	a, b := newAuthAndBSON()

	var value Auth
	err := bson.Unmarshal(b, &value)
	if err != nil {
		t.Fatalf("Error unmarshalling bson: %#v", err)
	}
	if a != value {
		t.Fatalf("Unexpected unmarshalled bson: %#v", value)
	}
}
