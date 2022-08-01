package resource

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UserProvider_Google UserProvider = "google"
)

type UserProvider string

type User struct {
	ID                 ResourceID   `json:"id" bson:"_id,omitempty"`
	Name               string       `json:"name" bson:"name"`
	ImageURL           *string      `json:"imageUrl" bson:"imageUrl"`
	Provider           UserProvider `json:"provider" bson:"provider"`
	ProviderResourceID string       `json:"providerResourceId" bson:"providerResourceId"`
	CreatedAt          time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt" bson:"updatedAt"`
}

type UserCollectionProvider interface {
	UserCollection() *UserCollection
}

type UserCollection struct {
	collection *mongo.Collection
}

func NewUserCollection(ctx context.Context, db *mongo.Database) *UserCollection {
	collection := db.Collection("user")

	// create indexes
	go func() {
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "providerResourceId", Value: 1}},
			Options: options.Index().SetUnique(true),
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

	return &UserCollection{collection: collection}
}

func (c *UserCollection) FindOneByID(
	ctx context.Context, id ResourceID,
) (*User, error) {
	_id, err := id.ObjectID()
	if err != nil {
		return nil, err
	}

	var user User
	err = c.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: _id},
	}).Decode(&user)

	if err != nil && err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCollection) FindOneByProviderResourceID(
	ctx context.Context, provider UserProvider, providerResourceID string,
) (*User, error) {
	var user User
	err := c.collection.FindOne(ctx, bson.D{
		{Key: "provider", Value: provider},
		{Key: "providerResourceId", Value: providerResourceID},
	}).Decode(&user)

	if err != nil && err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCollection) Save(
	ctx context.Context, user *User,
) error {
	if user.ID == "" {
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		r, err := c.collection.InsertOne(ctx, user)
		if err != nil {
			return err
		}
		user.ID = ResourceIDFromObjectID(r.InsertedID.(primitive.ObjectID))
		return nil
	} else {
		_id, err := user.ID.ObjectID()
		if err != nil {
			return err
		}

		user.UpdatedAt = time.Now()

		_, err = c.collection.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: _id},
		}, bson.D{
			{Key: "$set", Value: user},
		})
		return err
	}
}
