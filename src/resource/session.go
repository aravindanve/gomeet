package resource

import (
	"net/http"

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
	// TODO
}

func RegisterSessionRoutes(r *mux.Router, ds SessionDeps) *mux.Router {
	c := NewSessionController(ds)

	r.HandleFunc("/session", c.SessionRetrieveHandler).Methods(http.MethodOptions, http.MethodPost)

	return r
}
