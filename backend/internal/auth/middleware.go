package auth

import (
	"context"
	"log"
	"net/http"
)

type contextKey string

const UserContextKey = contextKey("user")

func BasicAuthMiddleware(us *UserStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		log.Println("request attempt by", username, password, "to", r.RequestURI)
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		valid, err := us.ValidateUser(r.Context(), username, password)
		if err != nil || !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
