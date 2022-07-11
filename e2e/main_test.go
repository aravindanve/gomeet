package main_test

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var pool *dockertest.Pool
var resources []*dockertest.Resource

func TestMain(m *testing.M) {
	code := 0

	// defer teardown
	defer teardownMain(&code)

	// setup
	setupMain()

	// run
	log.Println("running tests...")
	code = m.Run()
}

func setupMain() {
	log.Println("setting up tests...")

	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Panicf("could not connect to docker: %s", err)
	}

	// start mongo container
	resource, err := pool.Run("mongo", "latest", []string{
		"MONGO_INITDB_ROOT_USERNAME=mongo",
		"MONGO_INITDB_ROOT_PASSWORD=mongo",
	})
	if err != nil {
		log.Panicf("could not start mongo container: %s", err)
	}

	// add resource to resources
	resources = append(resources, resource)

	// format mongo connection uri
	uri := fmt.Sprintf(
		"mongodb://mongo:mongo@%s/mongo?authSource=admin",
		resource.GetHostPort("27017/tcp"),
	)

	// log.Printf("uri: %s\n", uri)

	// set mongo connection uri
	os.Setenv("MONGO_CONNECTION_URI", uri)

	// exponential backoff-retry until container ready to accept connections
	if err := pool.Retry(func() error {
		ctx := context.Background()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			return err
		}
		return client.Ping(ctx, readpref.Primary())
	}); err != nil {
		log.Panicf("could not connect to mongo container: %s", err)
	}
}

func teardownMain(code *int) {
	log.Println("tearing down tests...")

	if r := recover(); r != nil {
		log.Printf("recovered from panic: %s\n", r)
		debug.PrintStack()
	}

	if pool != nil {
		for _, resource := range resources {
			if err := pool.Purge(resource); err != nil {
				log.Panicf("could not purge resource: %s", err)
			}
		}
	}

	os.Exit(*code)
}

// helpers

func panicGuard(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("recovered from panic: %s\n", r)
		debug.PrintStack()
	}
}

var testUser *resource.User
var testUserMut sync.Mutex

func getTestUser() resource.User {
	testUserMut.Lock()
	if testUser == nil {
		// create test user
		ctx := context.Background()
		p := provider.NewProvider(ctx)
		defer p.Release(ctx)

		testUser = &resource.User{
			Name:               "Test User",
			Provider:           resource.UserProviderGoogle,
			ProviderResourceID: "some-google-id",
		}

		err := p.UserCollection().Save(ctx, testUser)
		if err != nil {
			panic(fmt.Sprintf("error saving user: %s", err.Error()))
		}
	}
	testUserMut.Unlock()
	return *testUser
}

var testAuthHeader *string
var testAuthHeaderMut sync.Mutex

func getTestAuthHeader() string {
	testAuthHeaderMut.Lock()
	if testAuthHeader == nil {
		// create test auth header
		cf := config.NewAuthConfigProvider().AuthConfig()
		user := getTestUser()
		token, err := jwt.NewBuilder().
			Issuer(cf.Issuer).
			Expiration(time.Now().Add(cf.TTL)).
			Claim("id", resource.ResourceIDFromObjectID(primitive.NewObjectID())).
			Claim("userId", user.ID).
			Build()

		if err != nil {
			panic(fmt.Sprintf("error creating jwt: %s", err.Error()))
		}

		// sign access token
		signed, err := jwt.Sign(token, jwt.WithKey(cf.Algorithm, cf.Secret))
		if err != nil {
			panic(fmt.Sprintf("error signing jwt: %s", err.Error()))
		}

		header := "Bearer " + string(signed)
		testAuthHeader = &header
	}
	testAuthHeaderMut.Unlock()
	return *testAuthHeader
}

var testMeeting *resource.Meeting
var testMeetingMut sync.Mutex

func getTestMeeting() resource.Meeting {
	testMeetingMut.Lock()
	if testMeeting == nil {
		// create test meeting
		ctx := context.Background()
		p := provider.NewProvider(ctx)
		defer p.Release(ctx)

		// create code
		buf := make([]byte, 7)
		_, err := rand.Read(buf)
		if err != nil {
			panic(fmt.Sprintf("error reading random bytes: %s", err.Error()))
		}

		code := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
		code = code[:4] + "-" + code[4:8] + "-" + code[8:]

		user := getTestUser()
		testMeeting = &resource.Meeting{
			UserID:    user.ID,
			Code:      code,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		err = p.MeetingCollection().Save(ctx, testMeeting)
		if err != nil {
			panic(fmt.Sprintf("error saving meeting: %s", err.Error()))
		}
	}
	testMeetingMut.Unlock()
	return *testMeeting
}
