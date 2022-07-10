package resource

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MeetingTTL = 365 * 24 * time.Hour
)

type MeetingDeps interface {
	MeetingCollectionProvider
}

type Meeting struct {
	ID        ResourceID `json:"id" bson:"_id,omitempty"`
	UserID    ResourceID `json:"userId" bson:"userId"`
	Code      string     `json:"code" bson:"code"`
	CreatedAt time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt" bson:"updatedAt"`
	ExpiresAt time.Time  `json:"expiresAt" bson:"expiresAt"`
}

type MeetingCollectionProvider interface {
	MeetingCollection() *MeetingCollection
}

type MeetingCollection struct {
	collection *mongo.Collection
}

func NewMeetingCollection(ctx context.Context, db *mongo.Database) *MeetingCollection {
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

	return &MeetingCollection{collection: collection}
}

type MeetingController struct {
	MeetingDeps
}

func NewMeetingController(ds MeetingDeps) *MeetingController {
	return &MeetingController{MeetingDeps: ds}
}

func (c *MeetingController) MeetingCreateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	// // decode authorization
	// a, err := util.GetAuthorization(r, c)
	// if err != nil {
	// 	util.WriteJSONError(w, http.StatusBadRequest, err.Error())
	// 	return
	// }
	// if a == nil {
	// 	util.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
	// 	return
	// }
}

func (c *MeetingController) MeetingSearchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *MeetingController) MeetingRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func RegisterMeetingRoutes(r *mux.Router, ds MeetingDeps) *mux.Router {
	c := NewMeetingController(ds)

	r.HandleFunc("/meetings", c.MeetingSearchHandler).Methods(http.MethodOptions, http.MethodGet).Queries("code", "{code}")
	r.HandleFunc("/meetings", c.MeetingCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/meetings/{meetingId}", c.MeetingRetrieveHandler).Methods(http.MethodOptions, http.MethodGet)

	return r
}
