package models

type User struct {
	ID           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}

type File struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Filetype string `db:"filetype"`
	Location string `db:"location"`
}
