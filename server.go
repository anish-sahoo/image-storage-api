package main

import (
	"net/http"
	"os"

	h "github.com/anish-sahoo/image-storage-api/internal/handlers"
	"github.com/gorilla/mux"
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

	router := mux.NewRouter()

	router.Handle("/login", h.LoggingMiddleware(http.HandlerFunc(h.LoginHandler)))
	router.Handle("/logout", h.LoggingMiddleware(http.HandlerFunc(h.LogoutHandler)))
	router.Handle("/auth-check", h.LoggingMiddleware(http.HandlerFunc(h.AuthCheckHandler)))
	router.Path("/images/{id}/delete").
		Handler(h.LoggingMiddleware(http.HandlerFunc(h.FileDeleteHandler)))
	router.Handle("/images", h.LoggingMiddleware(h.AuthMiddleware(http.HandlerFunc(h.ImagesHandler))))
	router.Handle("/api/images", h.LoggingMiddleware(http.HandlerFunc(h.AllImagesAPIHandler)))

	router.Path("/api/images/{id}/download").
		Handler(h.LoggingMiddleware(http.HandlerFunc(h.ImageDownloadAPIHandler)))

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("web")))

	log.Info().Msg("Server running on http://localhost:8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
