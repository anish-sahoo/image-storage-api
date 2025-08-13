package handlers

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("remote", r.RemoteAddr).Str("method", r.Method).Str("url", r.URL.String()).Msg("request received")
		next.ServeHTTP(w, r)
	})
}
