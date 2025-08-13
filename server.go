package main

import (
	"net/http"
	"os"

	h "github.com/anish-sahoo/image-storage-api/internal/handlers"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal().Msg("JWT_SECRET not set in environment")
		return
	}
	h.JWTSecret = []byte(secret)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	dbErr := h.InitDB("image_storage.db")
	if dbErr != nil {
		log.Fatal().Err(dbErr).Msg("Failed to initialize database")
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.Handle("/login", h.LoggingMiddleware(http.HandlerFunc(h.LoginHandler)))
	mux.Handle("/logout", h.LoggingMiddleware(http.HandlerFunc(h.LogoutHandler)))
	mux.Handle("/auth-check", h.LoggingMiddleware(http.HandlerFunc(h.AuthCheckHandler)))
	mux.Handle("/images", h.LoggingMiddleware(h.AuthMiddleware(http.HandlerFunc(h.ImagesHandler))))
	mux.Handle("/api/images", h.LoggingMiddleware(h.AuthMiddleware(http.HandlerFunc(h.ImagesAPIHandler))))

	log.Info().Msg("Server running on http://localhost:8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
