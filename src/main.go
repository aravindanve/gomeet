package main

import (
	"context"
	"net/http"

	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/resource"
	"github.com/aravindanve/gomeet-server/src/util"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func main() {
	ctx := context.Background()

	// init provider
	p := provider.NewProvider(ctx)

	// init router
	r := mux.NewRouter()

	// register routes and middleware
	resource.RegisterSessionRoutes(r)
	resource.RegisterAuthRoutes(r, p)
	resource.RegisterMeetingRoutes(r)
	resource.RegisterParticipantRoutes(r)
	r.Use(mux.CORSMethodMiddleware(r))

	// set 404 and 405 handlers
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.WriteJSONError(w, http.StatusNotFound, "Not found")
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		util.WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	// init logger
	l := negroni.NewLogger()
	l.SetFormat(`{{.Status}} {{.Method}} {{.Request.RequestURI}} - {{.Duration}}`)

	// init negroni
	n := negroni.New(negroni.NewRecovery(), l)
	n.UseHandler(r)

	// listen
	addr := p.HttpConfig().Addr
	l.Printf("HTTP Server listening on %s\n", addr)
	http.ListenAndServe(addr, n)
}
