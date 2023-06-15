package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRegisterUserHandler(t *testing.T) {
	client := &http.Client{}

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

	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))

	if err != nil {
		t.Fatal(err)
	}
	res, _ := client.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Print(string(body))
}
