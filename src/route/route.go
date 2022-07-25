package route

import (
	"net/http"

	"github.com/aravindanve/livemeet-server/src/middleware"
	"github.com/aravindanve/livemeet-server/src/provider"
	"github.com/aravindanve/livemeet-server/src/resource"
	"github.com/aravindanve/livemeet-server/src/util"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, p provider.Provider) *mux.Router {
	// register routes
	resource.RegisterSessionRoutes(r, p)
	resource.RegisterAuthRoutes(r, p)
	resource.RegisterMeetingRoutes(r, p)
	resource.RegisterParticipantRoutes(r, p)

	// register middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuthMiddleware(p))

	// handle 404
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSONError(w, http.StatusNotFound, "Not found")
	})

	return r
}
