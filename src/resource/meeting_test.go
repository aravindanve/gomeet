package resource

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newMeetingAndJSON() (Meeting, []byte) {
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var m = Meeting{
		ID:        "some-id",
		UserID:    "some-id",
		Code:      "some-code",
		CreatedAt: t,
		UpdatedAt: t,
		ExpiresAt: t,
	}

	var j = []byte(`{"id":"some-id","userId":"some-id","code":"some-code",` +
		`"createdAt":"2022-01-01T00:00:00Z","updatedAt":"2022-01-01T00:00:00Z",` +
		`"expiresAt":"2022-01-01T00:00:00Z"}`)

	return m, j
}

func TestMeetingMarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newMeetingAndJSON()

	value, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func TestMeetingUnmarshalJSON(t *testing.T) {
	t.Parallel()

	m, j := newMeetingAndJSON()

	var value Meeting
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if m != value {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}

func newMeetingAndBSON() (Meeting, []byte) {
	var o = primitive.NewObjectID()
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var m = Meeting{
		ID:        ResourceIDFromObjectID(o),
		UserID:    ResourceIDFromObjectID(o),
		Code:      "some-code",
		CreatedAt: t,
		UpdatedAt: t,
		ExpiresAt: t,
	}

	var d = primitive.NewDateTimeFromTime(t)
	var b, _ = bson.Marshal(bson.D{
		{Key: "_id", Value: o},
		{Key: "userId", Value: o},
		{Key: "code", Value: "some-code"},
		{Key: "createdAt", Value: d},
		{Key: "updatedAt", Value: d},
		{Key: "expiresAt", Value: d},
	})

	return m, b
}

func TestMeetingMarshalBSON(t *testing.T) {
	t.Parallel()

	m, b := newMeetingAndBSON()

	value, err := bson.Marshal(m)
	if err != nil {
		t.Fatalf("Error marshalling bson: %#v", err)
	}
	if !bytes.Equal(b, value) {
		t.Fatalf("Unexpected marshalled bson: %#v", string(value))
	}
}

func TestMeetingUnmarshalBSON(t *testing.T) {
	t.Parallel()

	m, b := newMeetingAndBSON()

	var value Meeting
	err := bson.Unmarshal(b, &value)
	if err != nil {
		t.Fatalf("Error unmarshalling bson: %#v", err)
	}
	if m != value {
		t.Fatalf("Unexpected unmarshalled bson: %#v", value)
	}
}
