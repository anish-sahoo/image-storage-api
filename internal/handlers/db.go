package handlers

import (
	"github.com/anish-sahoo/image-storage-api/internal/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sqlx.Open("sqlite3", dataSourceName)
	return err
}

func getUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := DB.Get(&user, "SELECT id, username, password_hash FROM users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func insertFile(file models.File) (bool, error) {
	return true, nil
}

func getFiles(user models.User, start int, end int) ([]models.File, error) {
	return nil, nil
}
