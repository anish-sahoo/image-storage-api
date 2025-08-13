package main

import (
	"net/http"
	"os"

	"github.com/anish-sahoo/image-storage-api/internal/handlers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	mux := http.NewServeMux()
	mux.Handle("/images", handlers.Logger(http.HandlerFunc(handlers.ImagesHandler)))

	log.Info().Msg("Server running on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
