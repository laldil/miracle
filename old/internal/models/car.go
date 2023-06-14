package models

import (
	"armageddon/internal/data"
	"armageddon/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CarModel struct {
	DB *sql.DB
}

type Car struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Brand       string    `json:"brand"`
	Description string    `json:"description"`
	Color       string    `json:"color,omitempty"`
	Year        int32     `json:"year,omitempty"`
	Price       int32     `json:"price"`
	IsUsed      bool      `json:"is_used"`
	OwnerID     int64     `json:"owner_id"`
}

func (m CarModel) Insert(car *Car) error {
	query := `
		INSERT INTO car (brand, description, color, year, price, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, is_used`

	args := []any{car.Brand, car.Description, car.Color, car.Year, car.Price, car.OwnerID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&car.ID, &car.CreatedAt, &car.IsUsed)
}

func (m CarModel) InsertToRent(car *Car, user *User) error {
	query := `
		INSERT INTO rented_cars (user_id, car_id, price, taking_date)
		VALUES ($1, $2, $3, $4)
		RETURNING car_id`

	args := []any{user.ID, car.ID, car.Price, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&car.ID)
}

func (m CarModel) DeleteFromRent(car *Car, user *User) error {
	query := `
		DELETE FROM rented_cars
		WHERE car_id = $1 
		AND user_id = $2`

	args := []any{
		car.ID,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m CarModel) Get(id int64) (*Car, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT * FROM car WHERE id = $1`

	var car Car

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&car.ID,
		&car.CreatedAt,
		&car.Brand,
		&car.Description,
		&car.Color,
		&car.Year,
		&car.Price,
		&car.IsUsed,
		&car.OwnerID,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &car, nil
}

func (m CarModel) Update(car *Car) error {
	query := `
		UPDATE car
		SET brand = $1, description = $2, color = $3, year = $4, price = $5, is_used = $6, owner_id = $7
		WHERE id = $8
		RETURNING id`

	args := []any{
		car.Brand,
		car.Description,
		car.Color,
		car.Year,
		car.Price,
		car.IsUsed,
		car.OwnerID,
		car.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&car.ID)
}

func (m CarModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM car WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m CarModel) GetAll(brand string, color string, filters data.Filters) ([]*Car, data.Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), * FROM car
		WHERE (to_tsvector('simple', brand) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', color) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.SortColumn(), filters.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{brand, color, filters.Limit(), filters.Offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, data.Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	cars := []*Car{}

	for rows.Next() {
		var car Car

		err := rows.Scan(
			&totalRecords,
			&car.ID,
			&car.CreatedAt,
			&car.Brand,
			&car.Description,
			&car.Color,
			&car.Year,
			&car.Price,
			&car.IsUsed,
			&car.OwnerID,
		)
		if err != nil {
			return nil, data.Metadata{}, err
		}

		cars = append(cars, &car)
	}

	if err = rows.Err(); err != nil {
		return nil, data.Metadata{}, err
	}

	metadata := data.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return cars, metadata, nil
}

func ValidateCar(v *validator.Validator, car *Car) {
	v.Check(car.Brand != "", "brand", "must be provided")
	v.Check(len(car.Brand) <= 500, "brand", "must not be more than 500 bytes long")

	v.Check(car.Description != "", "brand", "must be provided")
	v.Check(car.Color != "", "color", "must be provided")

	v.Check(car.Year != 0, "year", "must be provided")
	v.Check(car.Year >= 1888, "year", "must be greater than 1888")
	v.Check(car.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(car.Price != 0, "price", "must be provided")
	v.Check(car.Price > 0, "price", "must be a positive integer")
}
