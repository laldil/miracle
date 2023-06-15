package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCarHandler(t *testing.T) {
	app := &application{}

	input := struct {
		Brand       string `json:"brand"`
		Description string `json:"description"`
		Color       string `json:"color,omitempty"`
		Year        int32  `json:"year,omitempty"`
		Price       int32  `json:"price"`
		OwnerID     int64  `json:"owner_id"`
		IsUsed      bool   `json:"is_used"`
	}{
		Brand:       "Test Brand",
		Description: "Test Description",
		Color:       "Test Color",
		Year:        2023,
		Price:       50000,
		OwnerID:     5,
		IsUsed:      false,
	}
	jsonData, _ := json.Marshal(input)

	request, _ := http.NewRequest(http.MethodPost, "/cars", bytes.NewBuffer(jsonData))

	recorder := httptest.NewRecorder()

	app.createCarHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200; got %d", recorder.Code)
	}
	var response struct {
		Car struct {
			Brand       string `json:"brand"`
			Description string `json:"description"`
			Color       string `json:"color,omitempty"`
			Year        int32  `json:"year,omitempty"`
			Price       int32  `json:"price"`
			UserID      int64  `json:"user_id"`
			OwnerID     int64  `json:"owner_id"`
			IsUsed      bool   `json:"is_used"`
		} `json:"car"`
	}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}
	if response.Car.Brand != input.Brand {
		t.Errorf("expected brand %s; got %s", input.Brand, response.Car.Brand)
	}
	if response.Car.Description != input.Description {
		t.Errorf("expected description %s; got %s", input.Description, response.Car.Description)
	}
	if response.Car.Color != input.Color {
		t.Errorf("expected color %s; got %s", input.Color, response.Car.Color)
	}
	if response.Car.Year != input.Year {
		t.Errorf("expected year %d; got %d", input.Year, response.Car.Year)
	}
	if response.Car.Price != input.Price {
		t.Errorf("expected price %d; got %d", input.Price, response.Car.Price)
	}
}
