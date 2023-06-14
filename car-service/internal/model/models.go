package model

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Car CarModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Car: CarModel{DB: db},
	}
}
