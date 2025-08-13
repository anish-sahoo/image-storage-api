package main

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var jwtSecret = []byte("supersecretkey")

func generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	username, ok := claims["username"].(string)
	if !ok {
		return "", err
	}
	return username, nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("web")))

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			if username != "" && password != "" {
				token, err := generateJWT(username)
				if err == nil {
					http.SetCookie(w, &http.Cookie{
						Name:     "jwt",
						Value:    token,
						Path:     "/",
						HttpOnly: true,
						MaxAge:   3600,
					})
					w.Header().Set("Content-Type", "text/html")
					w.Write([]byte(`<div id='login-area'>Logged in!</div>`))
					return
				}
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div class='file-item'>Example file.txt</div>`))
		case http.MethodPost:
			cookie, err := r.Cookie("jwt")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			username, err := validateJWT(cookie.Value)
			if err != nil || username == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<div>File uploaded!</div>`))
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Info().Msg("Server running on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
