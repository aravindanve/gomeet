package resource

import (
	"net/http"

	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/util"
	"github.com/gorilla/mux"
)

type SessionDeps interface {
	UserCollectionProvider
}

type Session struct {
	User *User `json:"user"`
}

type SessionController struct {
	SessionDeps
}

func NewSessionController(ds SessionDeps) *SessionController {
	return &SessionController{SessionDeps: ds}
}

func (c *SessionController) SessionRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// decode auth token
	auth, err := middleware.GetAuthToken(r)
	if err != nil {
		util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if auth != nil {
		// find one user by id
		user, err := c.UserCollection().FindOneByID(r.Context(), ResourceID(auth.UserID))
		if err != nil {
			util.WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if user == nil {
			util.WriteJSONError(w, http.StatusNotFound, "User not found")
			return
		}

		util.WriteJSONResponse(w, http.StatusOK, &Session{User: user})
	} else {
		util.WriteJSONResponse(w, http.StatusOK, &Session{User: nil})
	}
}

func RegisterSessionRoutes(r *mux.Router, ds SessionDeps) *mux.Router {
	c := NewSessionController(ds)

	r.HandleFunc("/session", c.SessionRetrieveHandler).Methods(http.MethodGet)

	return r
}
