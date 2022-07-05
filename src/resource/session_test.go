package resource

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func newSessionAndJSON() (Session, []byte) {
	var m = Session{User: nil}
	var j = []byte(`{"user":null}`)

	return m, j
}

func TestSessionMarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newSessionAndJSON()

	value, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func TestSessionUnmarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newSessionAndJSON()

	var value Session
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if m != value {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}

func newSessionWithUserAndJSON() (Session, []byte) {
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var m = Session{
		User: &User{
			ID:                 "some-id",
			Name:               "Aravindan",
			ImageURL:           nil,
			Provider:           UserProviderGoogle,
			ProviderResourceID: "google-id",
			CreatedAt:          t,
			UpdatedAt:          t,
		},
	}

	var j = []byte(`{"user":{"id":"some-id","name":"Aravindan","imageUrl":null,` +
		`"provider":"google","providerResourceId":"google-id","createdAt":"2022-01-01T00:00:00Z",` +
		`"updatedAt":"2022-01-01T00:00:00Z"}}`)

	return m, j
}

func TestSessionWithUserMarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newSessionWithUserAndJSON()

	value, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func TestSessionWithUserUnmarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newSessionWithUserAndJSON()

	var value Session
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if *m.User != *value.User {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}
