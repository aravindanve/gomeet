package client

import (
	"context"
	"testing"

	"github.com/aravindanve/gomeet-server/src/config"
)

type mockMongoClientDeps struct {
	mongoConfig config.MongoConfig
}

func (p *mockMongoClientDeps) MongoConfig() config.MongoConfig {
	return p.mongoConfig
}

func NewMongoClientDeps(connectionURI string) MongoClientDeps {
	return &mockMongoClientDeps{
		mongoConfig: config.MongoConfig{
			ConnectionURI: connectionURI,
		},
	}
}

func TestNewMongoClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := NewMongoClientDeps("mongodb://host")
	c := NewMongoClient(ctx, p)
	defer c.Disconnect(ctx)
}

func TestGetMongoDatabaseDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pr := NewMongoClientDeps("mongodb://username:password@host:27017/database?authSource=admin")
	cl := NewMongoClient(ctx, pr)
	defer cl.Disconnect(ctx)

	db := GetMongoDatabaseDefault(cl, pr)

	a := "database"
	b := db.Name()
	if b != a {
		t.Errorf("expected database to be %#v got %#v", a, b)
		return
	}
}

func TestGetMongoDatabaseDefaultNoDatabase(t *testing.T) {
	t.Parallel()
	pr := NewMongoClientDeps("mongodb://username:password@host:27017/?authSource=admin")
	cl := NewMongoClient(context.Background(), pr)
	db := GetMongoDatabaseDefault(cl, pr)

	a := "default"
	b := db.Name()
	if b != a {
		t.Errorf("expected database to be %#v got %#v", a, b)
		return
	}
}
