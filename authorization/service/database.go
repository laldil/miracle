package main

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
)

const (
	dbConnectionString = "postgres://postgres:admin@localhost/armageddon?sslmode=disable"
)

func generateUniqueID(db *sql.DB) (int32, error) {
	var id int32
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM users").Scan(&id)
	if err != nil {
		log.Printf("Failed to generate unique ID: %v", err)
		return 0, err
	}
	return id, nil
}
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
