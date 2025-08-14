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

func insertFile(file models.File) error {
	_, err := DB.NamedExec("INSERT INTO files (name, file_type, location, owner_id, file_size, tag) VALUES (:name, :file_type, :location, :owner_id, :file_size, :tag) ", file)
	return err
}

func getImages(tag string, limit int, offset int) ([]models.File, int, error) {
	var files []models.File
	var total int
	err := DB.Get(&total, "SELECT COUNT(*) FROM files WHERE tag = ?", tag)
	if err != nil {
		return nil, 0, err
	}
	err = DB.Select(&files, `
        SELECT * 
        FROM files 
        WHERE tag = ? 
        ORDER BY created_at DESC 
        LIMIT ? OFFSET ?
    `, tag, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

func getFile(id int) (models.File, error) {
	var file models.File
	err := DB.Get(&file, "SELECT * FROM files WHERE id = ? ORDER BY created_at DESC LIMIT 1", id)
	if err != nil {
		return models.File{}, err
	}
	return file, err
}

func getFilesByUser(username string) ([]models.File, error) {
	var files []models.File
	user, err := getUserByUsername(username)
	if err != nil {
		return nil, err
	}

	err = DB.Select(&files, "SELECT * FROM files WHERE owner_id = ? ORDER BY created_at DESC", user.ID)
	if err != nil {
		return nil, err
	}
	return files, err
}
