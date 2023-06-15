package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterUserHandler(t *testing.T) {
	app := &application{}

	input := struct {
		Name     string `json:"name"`
		Surname  string `json:"surname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "Aldik",
		Surname:  "Aldiyarov",
		Email:    "ldi@gmail.com",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(input)

	request, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	recorder := httptest.NewRecorder()

	app.registerUserHandler(recorder, request)
}
