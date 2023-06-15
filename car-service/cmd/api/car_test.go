package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCarHandler(t *testing.T) {
	app := &application{}

	body := []byte(`{
		"brand": "Tesla",
		"description": "Model S",
		"color": "red",
		"year": 2022,
		"price": 50000,
		"user_id": 123
	}`)
	req, err := http.NewRequest("POST", "/create", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	app.createCarHandler(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	body = []byte(`{ invalid JSON`)
	req, err = http.NewRequest("POST", "/create", bytes.NewBuffer(body))
	assert.NoError(t, err)

	rr = httptest.NewRecorder()

	app.createCarHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestShowCarHandler(t *testing.T) {
	app := &application{}

	req, err := http.NewRequest("GET", "/cars/1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	app.showCarHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	req, err = http.NewRequest("GET", "/cars/999", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()

	app.showCarHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
