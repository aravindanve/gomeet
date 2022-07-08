package resource

import (
	"bytes"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resource struct {
	ID ResourceID `bson:"_id"`
}

func newResourceAndBSON() (Resource, []byte) {
	var o = primitive.NewObjectID()
	var r = Resource{ID: ResourceIDFromObjectID(o)}
	var b, _ = bson.Marshal(bson.D{{Key: "_id", Value: o}})
	return r, b
}

func TestResourceIDMarshalBSON(t *testing.T) {
	t.Parallel()
	r, b := newResourceAndBSON()

	value, err := bson.Marshal(r)
	if err != nil {
		t.Fatalf("Error marshalling bson: %#v", err)
	}
	if !bytes.Equal(b, value) {
		t.Fatalf("Unexpected marshalled bson: %#v", string(value))
	}
}

func TestResourceIDUnmarshalBSON(t *testing.T) {
	t.Parallel()
	r, b := newResourceAndBSON()

	var value Resource
	err := bson.Unmarshal(b, &value)
	if err != nil {
		t.Fatalf("Error unmarshalling bson: %#v", err)
	}
	if r != value {
		t.Fatalf("Unexpected unmarshalled bson: %#v", value)
	}
}
