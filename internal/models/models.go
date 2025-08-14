package models

import (
	"time"
)

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}

type File struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	Filetype      string    `db:"file_type"`
	Location      string    `db:"location"`
	OwnerId       int       `db:"owner_id"`
	FileSizeBytes int64     `db:"file_size"`
	CreatedAt     time.Time `db:"created_at"`
	Tag           string    `db:"tag"`
}

type PhotoResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"` // public API URL
	FileType  string `json:"fileType"`
	SizeBytes int64  `json:"sizeBytes"`
	CreatedAt string `json:"createdAt"`
}

type ListPhotosResponse struct {
	Photos       []PhotoResponse `json:"photos"`
	Total        int             `json:"total"`
	Limit        int             `json:"limit"`
	Offset       int             `json:"offset"`
	NextPage     string          `json:"nextPage,omitempty"`
	PreviousPage string          `json:"previousPage,omitempty"`
}
