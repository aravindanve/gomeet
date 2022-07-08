package resource

import (
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

type ParticipantCreateBody struct {
	Name string `json:"name"`
}

type ParticipantUpdateBody struct {
	Status ParticipantStatus `json:"status"`
}

type ParticipantController struct {
	collection *mongo.Collection
}

func (c *ParticipantController) ParticipantCreateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *ParticipantController) ParticipantRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (c *ParticipantController) ParticipantUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func NewParticipantController() *ParticipantController {
	return &ParticipantController{}
}

func RegisterParticipantRoutes(r *mux.Router) *mux.Router {
	c := NewParticipantController()

	r.HandleFunc("/participants", c.ParticipantCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/participants/{participantId}", c.ParticipantRetrieveHandler).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/participants/{participantId}", c.ParticipantUpdateHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
