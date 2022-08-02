package resource

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aravindanve/livemeet-server/src/client"
	"github.com/aravindanve/livemeet-server/src/config"
	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/util"
	"github.com/gorilla/mux"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	participantTTL               = 30 * time.Minute
	participantWaitingRoomSuffix = "_waiting"
)

type ParticipantStatus string

const (
	ParticipantStatus_Waiting  ParticipantStatus = "waiting"
	ParticipantStatus_Admitted ParticipantStatus = "admitted"
	ParticipantStatus_Denied   ParticipantStatus = "denied"
)

type RoomType string

const (
	RoomType_Waiting    RoomType = "waiting"
	RoomType_Conference RoomType = "conference"
)

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

type ParticipantWithRoomTokens struct {
	Participant
	RoomTokens []RoomToken `json:"roomTokens"`
}

type RoomToken struct {
	RoomName             string    `json:"roomName"`
	RoomType             RoomType  `json:"roomType"`
	AccessToken          string    `json:"accessToken"`
	AccessTokenExpiresAt time.Time `json:"accessTokenExpiresAt"`
}

type ParticipantMetadata struct {
	Name     string  `json:"name"`
	ImageURL *string `json:"imageUrl"`
}

func newParticipantWithRoomTokens(cf config.LiveKitConfig, participant *Participant, room string, roomAdmin bool) (*ParticipantWithRoomTokens, error) {
	var roomTokens []RoomToken

	// issue conference room token to admin or admitted
	if roomAdmin || participant.Status == ParticipantStatus_Admitted {
		at := auth.NewAccessToken(cf.APIKey, cf.APISecret)
		tr := true
		grant := &auth.VideoGrant{
			Room:           room,
			RoomAdmin:      roomAdmin,
			RoomCreate:     true,
			RoomJoin:       true,
			CanPublish:     &tr,
			CanPublishData: &tr,
			CanSubscribe:   &tr,
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
			SetValidFor(cf.RoomTokenTTL)

		token, err := at.ToJWT()
		if err != nil {
			return nil, err
		}

		roomTokens = append(roomTokens, RoomToken{
			RoomName:             room,
			RoomType:             RoomType_Conference,
			AccessToken:          token,
			AccessTokenExpiresAt: time.Now().Add(cf.RoomTokenTTL),
		})
	}

	// issue waiting room token to admin or waiting
	if roomAdmin || participant.Status == ParticipantStatus_Waiting {
		waitingRoom := room + participantWaitingRoomSuffix

		at := auth.NewAccessToken(cf.APIKey, cf.APISecret)
		fa := false
		grant := &auth.VideoGrant{
			Room:           waitingRoom,
			RoomCreate:     true,
			RoomJoin:       true,
			CanPublish:     &fa,
			CanPublishData: &fa,
			CanSubscribe:   &fa,
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
			SetValidFor(cf.RoomTokenTTL)

		token, err := at.ToJWT()
		if err != nil {
			return nil, err
		}

		roomTokens = append(roomTokens, RoomToken{
			RoomName:             room,
			RoomType:             RoomType_Waiting,
			AccessToken:          token,
			AccessTokenExpiresAt: time.Now().Add(cf.RoomTokenTTL),
		})
	}

	return &ParticipantWithRoomTokens{
		Participant: *participant,
		RoomTokens:  roomTokens,
	}, nil
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
		}, options.Update().SetUpsert(true))
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
	var status ParticipantStatus = ParticipantStatus_Waiting
	if auth != nil && meeting.UserID == ResourceID(auth.UserID) {
		admin = true
		status = ParticipantStatus_Admitted
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

	// save participant
	if status == ParticipantStatus_Waiting {
		// save participant
		err = c.ParticipantCollection().Save(r.Context(), participant)
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// create response
	res, err := newParticipantWithRoomTokens(c.LiveKitConfig(), participant, meeting.Code, admin)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.WriteJSONResponse(w, http.StatusOK, res)
}

func (c *ParticipantController) ParticipantRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// decode room token
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
	res, err := newParticipantWithRoomTokens(c.LiveKitConfig(), participant, meeting.Code, false)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// delete participant if not waiting
	if participant.Status != ParticipantStatus_Waiting {
		err = c.ParticipantCollection().DeleteOneByID(r.Context(), participant.ID)
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	util.WriteJSONResponse(w, http.StatusOK, res)
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
	if b.Status != ParticipantStatus_Admitted && b.Status != ParticipantStatus_Denied {
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

	// notify waiting room about updated participant
	var _type string
	if participant.Status == ParticipantStatus_Admitted {
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
		Room: meeting.Code + participantWaitingRoomSuffix,
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

	r.HandleFunc("/meetings/{meetingId}/participants", c.ParticipantCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/meetings/{meetingId}/participants/{participantId}", c.ParticipantRetrieveHandler).Methods(http.MethodGet)
	r.HandleFunc("/meetings/{meetingId}/participants/{participantId}", c.ParticipantUpdateHandler).Methods(http.MethodPut)

	return r
}
