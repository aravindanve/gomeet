package resource

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aravindanve/gomeet-server/src/client"
	"github.com/aravindanve/gomeet-server/src/config"
	"github.com/aravindanve/gomeet-server/src/middleware"
	"github.com/aravindanve/gomeet-server/src/util"
	"github.com/gorilla/mux"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	participantTTL = 30 * time.Minute
)

const (
	ParticipantStatusWaiting  ParticipantStatus = "waiting"
	ParticipantStatusAdmitted ParticipantStatus = "admitted"
	ParticipantStatusDenied   ParticipantStatus = "denied"
)

type ParticipantStatus string

type ParticipantDeps interface {
	config.LiveKitConfigProvider
	client.LiveKitClientProvider
	UserCollectionProvider
	MeetingCollectionProvider
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

type ParticipantWithJoinToken struct {
	Participant
	JoinToken          *string    `json:"joinToken"`
	JoinTokenExpiresAt *time.Time `json:"joinTokenExpiresAt"`
}

type ParticipantMetadata struct {
	Name     string  `json:"name"`
	ImageURL *string `json:"imageUrl"`
}

func newParticipantWithJoinToken(cf config.LiveKitConfig, participant *Participant, room string, roomAdmin bool) (*ParticipantWithJoinToken, error) {
	if participant.Status == ParticipantStatusAdmitted {
		at := auth.NewAccessToken(cf.APIKey, cf.APISecret)
		cp := true
		grant := &auth.VideoGrant{
			Room:           room,
			RoomAdmin:      roomAdmin,
			RoomCreate:     roomAdmin,
			RoomJoin:       true,
			CanPublish:     &cp,
			CanPublishData: &cp,
			CanSubscribe:   &cp,
		}
		metadata, err := json.Marshal(ParticipantMetadata{
			Name:     participant.Name,
			ImageURL: participant.ImageURL,
		})
		if err != nil {
			return nil, err
		}

		at.AddGrant(grant).
			SetIdentity(string(participant.ID)).
			SetMetadata(string(metadata)).
			SetValidFor(cf.JoinTokenTTL)

		token, err := at.ToJWT()
		if err != nil {
			return nil, err
		}

		tokenExpiresAt := time.Now().Add(cf.JoinTokenTTL)

		return &ParticipantWithJoinToken{
			Participant:        *participant,
			JoinToken:          &token,
			JoinTokenExpiresAt: &tokenExpiresAt,
		}, nil
	} else {
		return &ParticipantWithJoinToken{
			Participant:        *participant,
			JoinToken:          nil,
			JoinTokenExpiresAt: nil,
		}, nil
	}
}

type ParticipantCollectionProvider interface {
	ParticipantCollection() *ParticipantCollection
}

type ParticipantCollection struct {
	collection *mongo.Collection
}

func NewParticipantCollection(ctx context.Context, db *mongo.Database) *ParticipantCollection {
	collection := db.Collection("participant")

	return &ParticipantCollection{collection: collection}
}

type ParticipantController struct {
	ParticipantDeps
}

func (c *ParticipantCollection) FindOneByID(
	ctx context.Context, id ResourceID,
) (*Participant, error) {
	_id, err := id.ObjectID()
	if err != nil {
		return nil, err
	}

	var participant Participant
	err = c.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: _id},
	}).Decode(&participant)

	if err != nil && err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &participant, nil
}

func (c *ParticipantCollection) DeleteOneByID(
	ctx context.Context, id ResourceID,
) error {
	_id, err := id.ObjectID()
	if err != nil {
		return err
	}

	_, err = c.collection.DeleteOne(ctx, bson.D{
		{Key: "_id", Value: _id},
	})
	return err
}

func (c *ParticipantCollection) Save(
	ctx context.Context, participant *Participant,
) error {
	if participant.ID == "" {
		now := time.Now()
		participant.CreatedAt = now
		participant.UpdatedAt = now

		r, err := c.collection.InsertOne(ctx, participant)
		if err != nil {
			return err
		}
		participant.ID = ResourceIDFromObjectID(r.InsertedID.(primitive.ObjectID))
		return nil
	} else {
		_id, err := participant.ID.ObjectID()
		if err != nil {
			return err
		}

		participant.UpdatedAt = time.Now()

		_, err = c.collection.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: _id},
		}, bson.D{
			{Key: "$set", Value: participant},
		})
		return err
	}
}

type ParticipantCreateBody struct {
	Name *string `json:"name"`
}

func NewParticipantController(ds ParticipantDeps) *ParticipantController {
	return &ParticipantController{ParticipantDeps: ds}
}

func (c *ParticipantController) ParticipantCreateHandler(w http.ResponseWriter, r *http.Request) {
	// get meeting id
	meetingID := mux.Vars(r)["meetingId"]
	if meetingID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing meetingId in request path")
		return
	}

	// find one meeting by id
	meeting, err := c.MeetingCollection().FindOneByID(r.Context(), ResourceID(meetingID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meeting == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Meeting not found")
		return
	}

	// decode auth token
	auth, err := middleware.GetAuthToken(r)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// get name and image
	var name string
	var imageURL *string
	if auth != nil {
		user, err := c.UserCollection().FindOneByID(r.Context(), ResourceID(auth.UserID))
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if user != nil {
			name = user.Name
			imageURL = user.ImageURL
		}
	} else {
		// decode body
		b := &ParticipantCreateBody{}
		if err := json.NewDecoder(r.Body).Decode(b); err != nil {
			util.WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		name = *b.Name
	}

	// get admin and status
	var admin bool
	var status ParticipantStatus = ParticipantStatusWaiting
	if auth != nil && meeting.UserID == ResourceID(auth.UserID) {
		admin = true
		status = ParticipantStatusAdmitted
	}

	// create participant
	now := time.Now()
	participant := &Participant{
		ID:        ResourceIDFromObjectID(primitive.NewObjectID()),
		MeetingID: meeting.ID,
		Name:      name,
		ImageURL:  imageURL,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(participantTTL),
	}

	// save participant and notify if waiting
	if status == ParticipantStatusWaiting {
		// save participant
		err = c.ParticipantCollection().Save(r.Context(), participant)
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// notify room about waiting participant
		data, err := util.EncodeLiveKitDataJSON(map[string]any{
			"type":     "participantWaiting",
			"id":       participant.ID,
			"name":     participant.Name,
			"imageUrl": participant.ImageURL,
		})
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_, err = c.LiveKitClient().SendData(r.Context(), &livekit.SendDataRequest{
			Room: meeting.Code,
			Data: data,
			Kind: livekit.DataPacket_RELIABLE,
		})
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// create response
	response, err := newParticipantWithJoinToken(c.LiveKitConfig(), participant, meeting.Code, admin)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, response)
}

func (c *ParticipantController) ParticipantRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// decode join token
	authHeader := r.Header.Get("authorization")
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) < 2 || authHeaderParts[0] != "Bearer" {
		util.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	authToken := authHeaderParts[1]
	authVerifier, err := auth.ParseAPIToken(authToken)
	if err != nil {
		util.WriteJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	authClaims, err := authVerifier.Verify(c.LiveKitConfig().APISecret)
	if err != nil {
		util.WriteJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// get participant id
	participantID := mux.Vars(r)["participantId"]
	if participantID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing participantId in request path")
		return
	}

	if authClaims.Identity != participantID {
		util.WriteJSONError(w, http.StatusUnauthorized, "The authorized identity does not match participant")
		return
	}

	// find one participant by id
	participant, err := c.ParticipantCollection().FindOneByID(r.Context(), ResourceID(participantID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if participant == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Participant not found")
		return
	}

	// get meeting id
	meetingID := mux.Vars(r)["meetingId"]
	if meetingID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing meetingId in request path")
		return
	}

	// find one meeting by id
	meeting, err := c.MeetingCollection().FindOneByID(r.Context(), ResourceID(meetingID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meeting == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Meeting not found")
		return
	}

	// create response
	response, err := newParticipantWithJoinToken(c.LiveKitConfig(), participant, meeting.Code, false)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// delete participant
	err = c.ParticipantCollection().DeleteOneByID(r.Context(), participant.ID)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, response)
}

type ParticipantUpdateBody struct {
	Status ParticipantStatus `json:"status"`
}

func (c *ParticipantController) ParticipantUpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	// get meeting id
	meetingID := mux.Vars(r)["meetingId"]
	if meetingID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing meetingId in request path")
		return
	}

	// find one meeting by id
	meeting, err := c.MeetingCollection().FindOneByID(r.Context(), ResourceID(meetingID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meeting == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Meeting not found")
		return
	}

	// get participant id
	participantID := mux.Vars(r)["participantId"]
	if participantID == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing participantId in request path")
		return
	}

	// find one participant by id
	participant, err := c.ParticipantCollection().FindOneByID(r.Context(), ResourceID(participantID))
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if participant == nil {
		util.WriteJSONError(w, http.StatusNotFound, "Participant not found")
		return
	}

	// ensure auth user is the meeting admin
	if auth.UserID != string(meeting.UserID) {
		util.WriteJSONError(w, http.StatusUnauthorized, "Only meeting admins can update participants")
		return
	}
	// decode body
	b := &ParticipantUpdateBody{}
	if err := json.NewDecoder(r.Body).Decode(b); err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if b.Status == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "Missing status in request body")
		return
	}
	if b.Status != ParticipantStatusAdmitted && b.Status != ParticipantStatusDenied {
		util.WriteJSONError(w, http.StatusBadRequest, "Unexpected status in request body")
		return
	}

	// update participant
	participant.Status = b.Status
	participant.ExpiresAt = time.Now().Add(participantTTL)

	// save participant
	err = c.ParticipantCollection().Save(r.Context(), participant)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// notify room about updated participant
	var _type string
	if participant.Status == ParticipantStatusAdmitted {
		_type = "participantAdmitted"
	} else {
		_type = "participantDenied"
	}

	data, err := util.EncodeLiveKitDataJSON(map[string]any{
		"type": _type,
		"id":   participant.ID,
	})
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = c.LiveKitClient().SendData(r.Context(), &livekit.SendDataRequest{
		Room: meeting.Code,
		Data: data,
		Kind: livekit.DataPacket_RELIABLE,
	})
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, participant)
}

func RegisterParticipantRoutes(r *mux.Router, ds ParticipantDeps) *mux.Router {
	c := NewParticipantController(ds)

	r.HandleFunc("/meetings/{meetingId}/participants", c.ParticipantCreateHandler).Methods(http.MethodOptions, http.MethodPost)
	r.HandleFunc("/meetings/{meetingId}/participants/{participantId}", c.ParticipantRetrieveHandler).Methods(http.MethodOptions, http.MethodGet)
	r.HandleFunc("/meetings/{meetingId}/participants/{participantId}", c.ParticipantUpdateHandler).Methods(http.MethodOptions, http.MethodPut)

	return r
}
