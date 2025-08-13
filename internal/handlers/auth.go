package handlers

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func CheckCredentials(username, password string) bool {
	username = strings.TrimSpace(username)
	if len(username) == 0 || len(username) > 64 {
		return false
	}
	user, err := getUserByUsername(username)
	if err != nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
