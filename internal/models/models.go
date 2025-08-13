package models

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}

type File struct {
	ID       int    `db:"id"`
	name     string `db:"name"`
	filetype string `db:"filetype"`
	location string `db:"location"`
}
