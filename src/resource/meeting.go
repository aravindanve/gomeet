package resource

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/util"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	meetingTTL = 365 * 24 * time.Hour
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
	collection := db.Collection("meeting")

	// create indexes
	go func() {
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "code", Value: 1}},
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

	return &MeetingCollection{collection: collection}
}

func (c *MeetingCollection) FindAnyByCode(
	ctx context.Context, code string,
) ([]*Meeting, error) {
	cur, err := c.collection.Find(ctx, bson.D{
		{Key: "code", Value: code},
	}, options.Find().SetLimit(1))
	if err != nil {
		return nil, err
	}

	meetings := make([]*Meeting, 0)
	err = cur.All(ctx, &meetings)
	if err != nil {
		return nil, err
	}

	return meetings, nil
}

func (c *MeetingCollection) FindOneByID(
	ctx context.Context, id ResourceID,
) (*Meeting, error) {
	_id, err := id.ObjectID()
	if err != nil {
		return nil, err
	}

	var meeting Meeting
	err = c.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: _id},
	}).Decode(&meeting)

	if err != nil && err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &meeting, nil
}

func (c *MeetingCollection) Save(
	ctx context.Context, meeting *Meeting,
) error {
	if meeting.ID == "" {
		now := time.Now()
		meeting.CreatedAt = now
		meeting.UpdatedAt = now

		r, err := c.collection.InsertOne(ctx, meeting)
		if err != nil {
			return err
		}
		meeting.ID = ResourceIDFromObjectID(r.InsertedID.(primitive.ObjectID))
		return nil
	} else {
		_id, err := meeting.ID.ObjectID()
		if err != nil {
			return err
		}

		meeting.UpdatedAt = time.Now()

		_, err = c.collection.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: _id},
		}, bson.D{
			{Key: "$set", Value: meeting},
		})
		return err
	}
}

type MeetingController struct {
	MeetingDeps
}

func NewMeetingController(ds MeetingDeps) *MeetingController {
	return &MeetingController{MeetingDeps: ds}
}

func (c *MeetingController) MeetingCreateHandler(w http.ResponseWriter, r *http.Request) {
	// decode auth token
	auth, err := middleware.GetAuthToken(r)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if auth == nil {
		util.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// create code
	buf := make([]byte, 7)
	_, err = rand.Read(buf)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	code := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf))
	code = code[:4] + "-" + code[4:8] + "-" + code[8:]

	// create meeting
	meeting := &Meeting{
		UserID:    ResourceID(auth.UserID),
		Code:      code,
		ExpiresAt: time.Now().Add(meetingTTL),
	}

	// save meeting
	err = c.MeetingCollection().Save(r.Context(), meeting)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, meeting)
}

func (c *MeetingController) MeetingSearchHandler(w http.ResponseWriter, r *http.Request) {
	// get meeting code
	code := r.URL.Query().Get("code")
	if code == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing code in request query")
		return
	}

	// find any by code
	meetings, err := c.MeetingCollection().FindAnyByCode(r.Context(), code)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := map[string]any{
		"meetings": meetings,
	}

	util.WriteJSONResponse(w, http.StatusOK, res)
}

func (c *MeetingController) MeetingRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// get meeting id
	meetingID := mux.Vars(r)["meetingId"]
	if meetingID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing meetingId in request path")
		return
	}

	// find one by id
	meeting, err := c.MeetingCollection().FindOneByID(r.Context(), ResourceID(meetingID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meeting == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Meeting not found")
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, meeting)
}

func RegisterMeetingRoutes(r *mux.Router, ds MeetingDeps) *mux.Router {
	c := NewMeetingController(ds)

	r.HandleFunc("/meetings", c.MeetingSearchHandler).Methods(http.MethodGet).Queries("code", "{code}")
	r.HandleFunc("/meetings", c.MeetingCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/meetings/{meetingId}", c.MeetingRetrieveHandler).Methods(http.MethodGet)

	return r
}
