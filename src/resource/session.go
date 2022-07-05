package resource

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Session struct {
	User *User `json:"user"`
}

type SessionController struct{}

func (c *SessionController) SessionRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func NewSessionController() *SessionController {
	return &SessionController{}
}

func NewSessionRouter() *mux.Router {
	r := mux.NewRouter()
	c := NewSessionController()

	r.HandleFunc("/session", c.SessionRetrieveHandler).Methods(http.MethodOptions, http.MethodPost)

	return r
}
