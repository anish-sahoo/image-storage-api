package handlers

import (
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func CheckCredentials(username, password string) bool {
	username = strings.TrimSpace(username)
	if len(username) == 0 || len(username) > 64 {
		return false
	}
	user, err := getUserByUsername(username)
	if err != nil {
		log.Err(err).Msg("Error getUserByUsername")
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
