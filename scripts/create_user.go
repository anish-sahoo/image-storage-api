package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run create_user.go <username> <password>")
		return
	}
	username := os.Args[1]
	password := os.Args[2]

	db, err := sqlx.Open("sqlite3", "image_storage.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, string(hash))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User created successfully!")
}
