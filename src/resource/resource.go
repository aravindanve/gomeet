package resource

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceID string

func NewResourceID() ResourceID {
	return ResourceIDFromObjectID(primitive.NewObjectID())
}

func ResourceIDFromObjectID(o primitive.ObjectID) ResourceID {
	return ResourceID(o.Hex())
}

func (e ResourceID) ObjectID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(e))
}

func (e ResourceID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	v, err := e.ObjectID()
	if err != nil {
		return 0, nil, err
	}
	return bson.MarshalValue(v)
}

func (o *ResourceID) UnmarshalBSONValue(t bsontype.Type, v []byte) error {
	if t == bsontype.ObjectID {
		*o = ResourceID((bson.RawValue{Type: t, Value: v}).ObjectID().Hex())
		return nil
	}
	return fmt.Errorf("unexpected bsontype %s", t)
}
