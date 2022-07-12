package resource

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newUserAndJSON() (User, []byte) {
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var u = User{
		ID:                 "some-id",
		Name:               "Aravindan",
		ImageURL:           nil,
		Provider:           UserProviderGoogle,
		ProviderResourceID: "google-id",
		CreatedAt:          t,
		UpdatedAt:          t,
	}

	var j = []byte(`{"id":"some-id","name":"Aravindan","imageUrl":null,` +
		`"provider":"google","providerResourceId":"google-id","createdAt":"2022-01-01T00:00:00Z",` +
		`"updatedAt":"2022-01-01T00:00:00Z"}`)

	return u, j
}

func MockUserMarshalJSON(t *testing.T) {
	t.Parallel()
	u, j := newUserAndJSON()

	value, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func MockUserUnmarshalJSON(t *testing.T) {
	t.Parallel()
	u, j := newUserAndJSON()

	var value User
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if u != value {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}

func newUserAndBSON() (User, []byte) {
	var o = primitive.NewObjectID()
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var u = User{
		ID:                 ResourceIDFromObjectID(o),
		Name:               "Aravindan",
		ImageURL:           nil,
		Provider:           UserProviderGoogle,
		ProviderResourceID: "google-id",
		CreatedAt:          t,
		UpdatedAt:          t,
	}

	var d = primitive.NewDateTimeFromTime(t)
	var b, _ = bson.Marshal(bson.D{
		{Key: "_id", Value: o},
		{Key: "name", Value: "Aravindan"},
		{Key: "imageUrl", Value: nil},
		{Key: "provider", Value: UserProviderGoogle},
		{Key: "providerResourceId", Value: "google-id"},
		{Key: "createdAt", Value: d},
		{Key: "updatedAt", Value: d},
	})

	return u, b
}

func MockUserMarshalBSON(t *testing.T) {
	t.Parallel()
	u, b := newUserAndBSON()

	value, err := bson.Marshal(u)
	if err != nil {
		t.Fatalf("Error marshalling bson: %#v", err)
	}
	if !bytes.Equal(b, value) {
		t.Fatalf("Unexpected marshalled bson: %#v", string(value))
	}
}

func MockUserUnmarshalBSON(t *testing.T) {
	t.Parallel()
	u, b := newUserAndBSON()

	var value User
	err := bson.Unmarshal(b, &value)
	if err != nil {
		t.Fatalf("Error unmarshalling bson: %#v", err)
	}
	if u != value {
		t.Fatalf("Unexpected unmarshalled bson: %#v", value)
	}
}
