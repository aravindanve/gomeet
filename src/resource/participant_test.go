package resource

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newParticipantAndJSON() (ParticipantWithRoomTokens, []byte) {
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var p = ParticipantWithRoomTokens{
		Participant: Participant{
			ID:        "some-id",
			MeetingID: "some-id",
			Name:      "Aravindan",
			ImageURL:  nil,
			Status:    ParticipantStatusWaiting,
			CreatedAt: t,
			UpdatedAt: t,
			ExpiresAt: t,
		},
		RoomTokens: []RoomToken{{
			RoomName:             "some-room",
			RoomType:             RoomTypeConference,
			AccessToken:          "some-token",
			AccessTokenExpiresAt: t,
		}},
	}

	var j = []byte(`{"id":"some-id","meetingId":"some-id","name":"Aravindan",` +
		`"imageUrl":null,"status":"waiting","createdAt":"2022-01-01T00:00:00Z",` +
		`"updatedAt":"2022-01-01T00:00:00Z","expiresAt":"2022-01-01T00:00:00Z",` +
		`"roomTokens":[{"roomName":"some-room","roomType":"conference","accessToken":"some-token",` +
		`"accessTokenExpiresAt":"2022-01-01T00:00:00Z"}]}`)

	return p, j
}

func TestParticipantMarshalJSON(t *testing.T) {
	t.Parallel()
	p, j := newParticipantAndJSON()

	value, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Error marshalling json: %#v", err)
	}
	if !bytes.Equal(j, value) {
		t.Fatalf("Unexpected marshalled json: %#v", string(value))
	}
}

func TestParticipantUnmarshalJSON(t *testing.T) {
	t.Parallel()
	p, j := newParticipantAndJSON()

	var value ParticipantWithRoomTokens
	err := json.Unmarshal([]byte(j), &value)
	if err != nil {
		t.Fatalf("Error unmarshalling json: %#v", err)
	}
	if !reflect.DeepEqual(p, value) {
		t.Fatalf("Unexpected unmarshalled json: %#v", value)
	}
}

func newParticipantAndBSON() (Participant, []byte) {
	var o = primitive.NewObjectID()
	var t, _ = time.Parse(time.RFC3339, "2022-01-01T00:00:00.000Z")
	var p = Participant{
		ID:        ResourceIDFromObjectID(o),
		MeetingID: ResourceIDFromObjectID(o),
		Name:      "Aravindan",
		ImageURL:  nil,
		Status:    ParticipantStatusWaiting,
		CreatedAt: t,
		UpdatedAt: t,
		ExpiresAt: t,
	}

	var d = primitive.NewDateTimeFromTime(t)
	var b, _ = bson.Marshal(bson.D{
		{Key: "_id", Value: o},
		{Key: "meetingId", Value: o},
		{Key: "name", Value: "Aravindan"},
		{Key: "imageUrl", Value: nil},
		{Key: "status", Value: "waiting"},
		{Key: "createdAt", Value: d},
		{Key: "updatedAt", Value: d},
		{Key: "expiresAt", Value: d},
	})

	return p, b
}

func TestParticipantMarshalBSON(t *testing.T) {
	t.Parallel()
	p, b := newParticipantAndBSON()

	value, err := bson.Marshal(p)
	if err != nil {
		t.Fatalf("Error marshalling bson: %#v", err)
	}
	if !bytes.Equal(b, value) {
		t.Fatalf("Unexpected marshalled bson: %#v", string(value))
	}
}

func TestParticipantUnmarshalBSON(t *testing.T) {
	t.Parallel()
	p, b := newParticipantAndBSON()

	var value Participant
	err := bson.Unmarshal(b, &value)
	if err != nil {
		t.Fatalf("Error unmarshalling bson: %#v", err)
	}
	if !reflect.DeepEqual(p, value) {
		t.Fatalf("Unexpected unmarshalled bson: %#v", value)
	}
}
