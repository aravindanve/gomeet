package client

import (
	"context"
	"fmt"

	"github.com/aravindanve/gomeet-server/src/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type MongoClientDeps interface {
	config.MongoConfigProvider
}

type MongoClientProvider interface {
	MongoClient() *mongo.Client
	MongoDatabase() *mongo.Database
}

func NewMongoClient(ctx context.Context, ds MongoClientDeps) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(ds.MongoConfig().ConnectionURI))
	if err != nil {
		panic(fmt.Sprintf("creating mongo client failed with %s", err.Error()))
	}
	return client
}

func GetMongoDatabaseDefault(client *mongo.Client, ds MongoClientDeps) *mongo.Database {
	cs, err := connstring.ParseAndValidate(ds.MongoConfig().ConnectionURI)
	if err != nil {
		panic(fmt.Sprintf("parsing mongo connection uri failed with %s", err.Error()))
	}
	if cs.Database != "" {
		return client.Database(cs.Database)
	} else {
		return client.Database("default")
	}
}
