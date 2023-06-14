package main

import (
	"armageddon/internal/jsonlog"
	"armageddon/internal/models"
	"bytes"
	"encoding/json"
	"flag"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestRegisterUserHandler(t *testing.T) {
	var cfg config
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:7777@localhost/armageddon?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	app := application{
		config: cfg,
		logger: logger,
		models: models.NewModels(db),
	}

	type input struct {
		Name     string `json:"name"`
		Surname  string `json:"surname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	testTable := []struct {
		name      string
		userInput input
	}{
		{
			name: "Test 1",
			userInput: input{
				Name:     "testName",
				Surname:  "testSuernma",
				Email:    "test@test.test",
				Password: "123456789",
			},
		},
		{
			name: "Test 2",
			userInput: input{
				Name:     "testName1",
				Surname:  "testSuernma1",
				Email:    "test@b.b",
				Password: "123456789",
			},
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			marshal, err := json.Marshal(testTable.userInput)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(marshal))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(app.registerUserHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusAccepted {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusAccepted)
			}
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	var cfg config
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:7777@localhost/armageddon?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	app := application{
		config: cfg,
		logger: logger,
		models: models.NewModels(db),
	}

	router := httprouter.New()
	router.HandlerFunc(http.MethodDelete, "/users/:id", app.deleteUserHandler)

	testTable := []struct {
		name   string
		userID int
	}{
		{name: "Test 1", userID: 12},
		{name: "Test 2", userID: 13},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.userID)
			req, err := http.NewRequest(http.MethodDelete, "/users/"+id, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestShowUserHandler(t *testing.T) {
	var cfg config
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:7777@localhost/armageddon?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	app := application{
		config: cfg,
		logger: logger,
		models: models.NewModels(db),
	}

	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/users/:id", app.showUserHandler)

	testTable := []struct {
		name       string
		userID     int
		httpStatus int
	}{
		{
			name:       "Test 1",
			userID:     1,
			httpStatus: http.StatusOK,
		},
		{
			name:       "Test 2",
			userID:     2,
			httpStatus: http.StatusNotFound,
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.userID)
			req, err := http.NewRequest(http.MethodGet, "/users/"+id, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != testTable.httpStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, testTable.httpStatus)
			}

			//Check the response body is what we expect.
			//expected := `{"id":1,"first_name":"Krish","last_name":"Bhanushali","email_address":"krishsb2405@gmail.com","phone_number":"0987654321"}`
			//if rr.Body.String() != expected {
			//	t.Errorf("handler returned unexpected body: got %v want %v",
			//		rr.Body.String(), expected)
			//}
		})
	}
}
