package models

import (
	"database/sql"
	"errors"
	"user-service/internal/data"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Users  UserModel
	Tokens data.TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: data.TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
