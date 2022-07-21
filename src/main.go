package main

import (
	"context"
	"net/http"

	"github.com/aravindanve/gomeet-server/src/provider"
	"github.com/aravindanve/gomeet-server/src/route"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func main() {
	// init context
	ctx := context.Background()

	// init provider
	p := provider.NewProvider(ctx)

	// init router
	r := mux.NewRouter()

	// register routes and middleware
	route.RegisterRoutes(r, p)

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
