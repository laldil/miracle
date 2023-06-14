package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

func (app *application) createCarHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
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

	data := struct {
		Brand       string `json:"brand"`
		Description string `json:"description"`
		Color       string `json:"color,omitempty"`
		Year        int32  `json:"year,omitempty"`
		Price       int32  `json:"price"`
		UserID      int64  `json:"user_id"`
	}{
		Brand:       input.Brand,
		Description: input.Description,
		Color:       input.Color,
		Year:        input.Year,
		Price:       input.Price,
		UserID:      user.ID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request, err := http.NewRequest("POST", "http://localhost:4000/car", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": data}, nil)
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

	request, err := http.NewRequest(http.MethodGet, "http://localhost:4000/car/"+strconv.Itoa(int(id)), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": result}, nil)
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

	request, err := http.NewRequest(http.MethodDelete, "http://localhost:4000/car/"+strconv.Itoa(int(id)), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "car successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCarHandler(w http.ResponseWriter, r *http.Request) {
	request, err := http.NewRequest(http.MethodGet, "http://localhost:4000/cars", nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": result}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) rentCarHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64 `json:"user_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	data := struct {
		UserID int64 `json:"user_id"`
	}{
		UserID: input.UserID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request, err := http.NewRequest(http.MethodPost, "http://localhost:4000/car/"+strconv.Itoa(int(id))+"/rent", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": result}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) returnRentedCarHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64 `json:"user_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	data := struct {
		UserID int64 `json:"user_id"`
	}{
		UserID: input.UserID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request, err := http.NewRequest(http.MethodPut, "http://localhost:4000/car/"+strconv.Itoa(int(id))+"/return", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	defer response.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"car": result}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
