package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	data := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    input.Email,
		Password: input.Password,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	request, err := http.NewRequest("POST", "http://localhost:4001/tokens/authentication", bytes.NewBuffer(jsonData))
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

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": result}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
