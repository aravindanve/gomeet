package provider

import (
	"context"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

type provider struct {
	*config.Config
	mongoClient        *mongo.Client
	authCollection     *resource.AuthCollection
	userCollection     *resource.UserCollection
	googleOAuth2Client client.GoogleOAuth2Client
}

func NewProvider(ctx context.Context) *provider {
	cf := config.NewConfig()

	mongoClient := client.NewMongoClient(ctx, cf)
	mongoDB := client.GetMongoDatabaseDefault(mongoClient, cf)

	return &provider{
		Config:             cf,
		mongoClient:        mongoClient,
		authCollection:     resource.NewAuthCollection(ctx, mongoDB),
		userCollection:     resource.NewUserCollection(ctx, mongoDB),
		googleOAuth2Client: client.NewGoogleOAuth2Client(cf),
	}
}

func (p *provider) Release(ctx context.Context) {
	p.mongoClient.Disconnect(ctx)
}

func (p *provider) MongoClient() *mongo.Client {
	return p.mongoClient
}

func (p *provider) AuthCollection() *resource.AuthCollection {
	return p.authCollection
}

func (p *provider) UserCollection() *resource.UserCollection {
	return p.userCollection
}

func (p *provider) GoogleOAuth2Client() client.GoogleOAuth2Client {
	return p.googleOAuth2Client
}
