package resource

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ParticipantTTL = 30 * time.Minute
)

const (
	ParticipantStatusWaiting  ParticipantStatus = "waiting"
	ParticipantStatusAdmitted ParticipantStatus = "admitted"
	ParticipantStatusDenied   ParticipantStatus = "denied"
)

type ParticipantStatus string

type ParticipantDeps interface {
	ParticipantCollectionProvider
}

type Participant struct {
	ID        ResourceID        `json:"id" bson:"_id,omitempty"`
	MeetingID ResourceID        `json:"meetingId" bson:"meetingId"`
	Name      string            `json:"name" bson:"name"`
	ImageURL  *string           `json:"imageUrl" bson:"imageUrl"`
	Status    ParticipantStatus `json:"status" bson:"status"`
	CreatedAt time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt" bson:"updatedAt"`
	ExpiresAt time.Time         `json:"expiresAt" bson:"expiresAt"`
}

type ParticipantWithToken struct {
	Participant
	Token          string    `json:"Token"`
	TokenExpiresAt time.Time `json:"TokenExpiresAt"`
}

type ParticipantMetadataPayload struct {
	Name     string  `json:"name"`
	ImageURL *string `json:"imageUrl"`
}

type ParticipantCollectionProvider interface {
	ParticipantCollection() *ParticipantCollection
}

type ParticipantCollection struct {
	collection *mongo.Collection
}

func NewParticipantCollection(ctx context.Context, db *mongo.Database) *ParticipantCollection {
	collection := db.Collection("auth")

	// TODO
	// // create indexes
	// go func() {
	// 	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
	// 		Keys:    bson.D{{Key: "providerResourceId", Value: 1}},
	// 		Options: options.Index().SetUnique(true),
	// 	})

	// 	if err != nil {
	// 		msg := fmt.Sprintf("error creating mongo indexes: %s", err.Error())
	// 		if os.Getenv("APP_ENV") == "testing" {
	// 			log.Println(msg) // do not panic in tests
	// 		} else {
	// 			panic(msg)
	// 		}
	// 	}
	// }()

	return &ParticipantCollection{collection: collection}
}

type ParticipantController struct {
	ParticipantDeps
}

type ParticipantCreateBody struct {
	Name string `json:"name"`
}

func NewParticipantController(ds ParticipantDeps) *ParticipantController {
	return &ParticipantController{ParticipantDeps: ds}
}

func (c *ParticipantController) ParticipantCreateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

type ParticipantUpdateBody struct {
	Status ParticipantStatus `json:"status"`
}

func (c *ParticipantController) ParticipantUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *ParticipantController) ParticipantRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func RegisterParticipantRoutes(r *mux.Router, ds ParticipantDeps) *mux.Router {
	c := NewParticipantController(ds)

	r.HandleFunc("/participants", c.ParticipantCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/participants/{participantId}", c.ParticipantRetrieveHandler).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/participants/{participantId}", c.ParticipantUpdateHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
