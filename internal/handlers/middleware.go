package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("remote", r.RemoteAddr).
			Msg("Request")
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("jwt")
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login.html")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			f, ferr := os.Open("web/login.html")
			if ferr != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			defer f.Close()
			w.Header().Set("Content-Type", "text/html")
			io.Copy(w, f)
			return
		}
		next.ServeHTTP(w, r)
	})
}
