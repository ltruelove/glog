package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	First string `json:"first"`
	Last  string `json:"last"`
	Pass  string `json:"pass"`
	Admin int    `json:"admin"`
}

func ValidateUser(email string, pass string) (bool, error) {

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var hashedPass string
	query := fmt.Sprintf("SELECT email, pass FROM users WHERE email='%s'", email)
	err = db.QueryRow(query).Scan(&email, &hashedPass)

	if err != nil {
		return false, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass)) == nil {
		return true, nil
	}

	return false, nil
}
