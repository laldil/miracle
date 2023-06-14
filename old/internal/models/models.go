package models

import (
	"armageddon/internal/data"
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Car    CarModel
	Tokens data.TokenModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Car:    CarModel{DB: db},
		Tokens: data.TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
