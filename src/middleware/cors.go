package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

func CORSMiddleware() mux.MiddlewareFunc {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if h := r.Header.Get("origin"); h != "" {
				w.Header().Set("vary", "Origin")
				w.Header().Set("access-control-allow-origin", h)
				w.Header().Set("access-control-allow-credentials", "true")

				if m := r.Method; m == http.MethodOptions {
					w.Header().Set("access-control-allow-methods", "GET,HEAD,PUT,POST,DELETE,PATCH")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			handler.ServeHTTP(w, r)
		})
	}
}
