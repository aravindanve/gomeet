package main

import (
	"net/http"

	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.Handle("/session", resource.NewSessionRouter())
	r.Handle("/auth", resource.NewAuthRouter())
	r.Handle("/meetings", resource.NewMeetingRouter())
	r.Handle("/meetings/{meetingId}/participants", resource.NewParticipantRouter())
	r.Use(mux.CORSMethodMiddleware(r))

	http.ListenAndServe(":8080", r)
}
