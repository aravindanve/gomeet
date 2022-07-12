package provider

import (
	"context"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

type Provider interface {
	config.Config
	client.MongoClientProvider
	client.GoogleOAuth2ClientProvider
	client.LiveKitClientProvider
	resource.UserCollectionProvider
	resource.AuthCollectionProvider
	resource.MeetingCollectionProvider
	resource.ParticipantCollectionProvider
	Release(ctx context.Context)
}

type provider struct {
	config.Config
	mongoClient           *mongo.Client
	mongoDatabase         *mongo.Database
	googleOAuth2Client    client.GoogleOAuth2Client
	livekitClient         client.LiveKitClient
	authCollection        *resource.AuthCollection
	userCollection        *resource.UserCollection
	meetingCollection     *resource.MeetingCollection
	participantCollection *resource.ParticipantCollection
}

func NewProvider(ctx context.Context) Provider {
	cf := config.NewConfig()

	mongoClient := client.NewMongoClient(ctx, cf)
	mongoDatabase := client.GetMongoDatabaseDefault(mongoClient, cf)

	return &provider{
		Config:                cf,
		mongoClient:           mongoClient,
		mongoDatabase:         mongoDatabase,
		googleOAuth2Client:    client.NewGoogleOAuth2Client(cf),
		livekitClient:         client.NewLiveKitClient(cf),
		authCollection:        resource.NewAuthCollection(ctx, mongoDatabase),
		userCollection:        resource.NewUserCollection(ctx, mongoDatabase),
		meetingCollection:     resource.NewMeetingCollection(ctx, mongoDatabase),
		participantCollection: resource.NewParticipantCollection(ctx, mongoDatabase),
	}
}

func (p *provider) Release(ctx context.Context) {
	p.mongoClient.Disconnect(ctx)
}

func (p *provider) MongoClient() *mongo.Client {
	return p.mongoClient
}

func (p *provider) MongoDatabase() *mongo.Database {
	return p.mongoDatabase
}

func (p *provider) GoogleOAuth2Client() client.GoogleOAuth2Client {
	return p.googleOAuth2Client
}

func (p *provider) LiveKitClient() client.LiveKitClient {
	return p.livekitClient
}

func (p *provider) AuthCollection() *resource.AuthCollection {
	return p.authCollection
}

func (p *provider) UserCollection() *resource.UserCollection {
	return p.userCollection
}

func (p *provider) MeetingCollection() *resource.MeetingCollection {
	return p.meetingCollection
}

func (p *provider) ParticipantCollection() *resource.ParticipantCollection {
	return p.participantCollection
}
