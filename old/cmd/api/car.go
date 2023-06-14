package main

import (
	"armageddon/internal/data"
	"armageddon/internal/models"
	"armageddon/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createCarHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Brand       string `json:"brand"`
		Description string `json:"description"`
		Color       string `json:"color,omitempty"`
		Year        int32  `json:"year,omitempty"`
		Price       int32  `json:"price"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	car := &models.Car{
		Brand:       input.Brand,
		Description: input.Description,
		Color:       input.Color,
		Year:        input.Year,
		Price:       input.Price,
		OwnerID:     user.ID,
	}

	v := validator.New()

	if models.ValidateCar(v, car); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Car.Insert(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/car/%d", car.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"car": car}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	car, err := app.models.Car.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": car}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	car, err := app.models.Car.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Brand       *string `json:"brand"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		Year        *int32  `json:"year"`
		Price       *int32  `json:"price"`
		IsUsed      *bool   `json:"is_used"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Brand != nil {
		car.Brand = *input.Brand
	}

	if input.Description != nil {
		car.Description = *input.Description
	}

	if input.Color != nil {
		car.Color = *input.Color
	}

	if input.Year != nil {
		car.Year = *input.Year
	}

	if input.Price != nil {
		car.Price = *input.Price
	}

	if input.IsUsed != nil {
		car.IsUsed = *input.IsUsed
	}

	v := validator.New()
	if models.ValidateCar(v, car); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Car.Update(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": car}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	car, err := app.models.Car.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user := app.contextGetUser(r)
	if user.Roles != "MODERATOR" {
		if car.OwnerID != user.ID {
			app.wrongCarResponse(w, r)
		}
		return
	}

	err = app.models.Car.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "car successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCarHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Brand string
		Color string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Brand = app.readString(qs, "brand", "")
	input.Color = app.readString(qs, "color", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "brand", "year", "price", "is_used", "-id", "-brand", "-year", "-price", "-is_used"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	car, metadata, err := app.models.Car.GetAll(input.Brand, input.Color, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": car, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) rentCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	car, err := app.models.Car.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if car.IsUsed != false {
		app.carOccupiedResponse(w, r)
		return
	}

	user := app.contextGetUser(r)
	err = app.models.Car.InsertToRent(car, user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	car.IsUsed = true
	err = app.models.Car.Update(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": car}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) returnRentedCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	car, err := app.models.Car.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if car.IsUsed != true {
		app.carNotUsedResponse(w, r)
		return
	}

	user := app.contextGetUser(r)
	err = app.models.Car.DeleteFromRent(car, user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	car.IsUsed = false
	err = app.models.Car.Update(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": car}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
