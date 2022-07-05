package resource

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MeetingTTL = 365 * 24 * time.Hour
)

type Meeting struct {
	ID        ResourceID `json:"id" bson:"_id,omitempty"`
	UserID    ResourceID `json:"userId" bson:"userId"`
	Code      string     `json:"code" bson:"code"`
	CreatedAt time.Time  `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt" bson:"updatedAt"`
	ExpiresAt time.Time  `json:"expiresAt" bson:"expiresAt"`
}

type MeetingController struct {
	collection *mongo.Collection
}

func (c *MeetingController) MeetingCreateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *MeetingController) MeetingSearchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *MeetingController) MeetingRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func NewMeetingController() *MeetingController {
	return &MeetingController{}
}

func NewMeetingRouter() *mux.Router {
	r := mux.NewRouter()
	c := NewMeetingController()

	r.Headers("content-type", "application/json")
	r.HandleFunc("/meetings", c.MeetingCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/meetings", c.MeetingSearchHandler).Methods(http.MethodOptions, http.MethodGet).Queries("code", "{code}")
	r.HandleFunc("/meetings/{meetingId}", c.MeetingRetrieveHandler).Methods(http.MethodOptions, http.MethodGet)

	return r
}
