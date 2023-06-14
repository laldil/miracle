package main

import (
	"bytes"
	"car-service/internal/jsonlog"
	"car-service/internal/model"
	"encoding/json"
	"flag"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func getConfig() *application {
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

	app := &application{
		config: cfg,
		logger: logger,
		models: model.NewModels(db),
	}
	return app
}

func TestListCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/cars", app.listCarHandler)

	testTable := []struct {
		name             string
		httpType         string
		expectedHttpCode int
	}{
		{
			name:             "Test 1",
			httpType:         http.MethodGet,
			expectedHttpCode: http.StatusOK,
		},
		{
			name:             "Test 2",
			httpType:         http.MethodDelete,
			expectedHttpCode: http.StatusMethodNotAllowed,
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			req, err := http.NewRequest(testTable.httpType, "/cars", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != testTable.expectedHttpCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, testTable.expectedHttpCode)
			}
		})
	}
}

func TestShowCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/car/:id", app.showCarHandler)

	testTable := []struct {
		name       string
		carID      int
		httpStatus int
	}{
		{
			name:       "Test 1",
			carID:      7,
			httpStatus: http.StatusOK,
		},
		{
			name:       "Test 2",
			carID:      1,
			httpStatus: http.StatusNotFound,
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.carID)
			req, err := http.NewRequest(http.MethodGet, "/car/"+id, nil)
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
			t.Log(rr.Body.String())

		})
	}
}

func TestCreateCarHandler(t *testing.T) {
	app := getConfig()
	type input struct {
		Brand       string `json:"brand"`
		Description string `json:"description"`
		Color       string `json:"color,omitempty"`
		Year        int32  `json:"year,omitempty"`
		Price       int32  `json:"price"`
	}

	testTable := []struct {
		name      string
		userInput input
		token     string
	}{
		{
			name: "Test 1",
			userInput: input{
				Brand:       "Mercedes",
				Description: "E63",
				Color:       "White",
				Year:        2022,
				Price:       100000,
			},
			token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4",
		},
		{
			name: "Test 2",
			userInput: input{
				Brand:       "BMW",
				Description: "E63",
				Color:       "Black",
				Year:        2015,
				Price:       123432,
			},
			token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4",
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			marshal, err := json.Marshal(testTable.userInput)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodPost, "/car", bytes.NewReader(marshal))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			user, err := app.model.Users.GetForToken(data.ScopeAuthentication, testTable.token)
			if err != nil {
				t.Error(err)
			}
			req = app.contextSetUser(req, user)

			handler := http.HandlerFunc(app.createCarHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusCreated {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusCreated)
			}
		})
	}
}

func TestDeleteCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodDelete, "/car/:id", app.deleteCarHandler)

	testTable := []struct {
		name  string
		carID int
		token string
	}{
		{name: "Test 1", carID: 22, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
		{name: "Test 2", carID: 23, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.carID)
			req, err := http.NewRequest(http.MethodDelete, "/car/"+id, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			user, err := app.models.Users.GetForToken(data.ScopeAuthentication, testTable.token)
			if err != nil {
				t.Error(err)
			}
			req = app.contextSetUser(req, user)

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestRentCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/car/:id/rent", app.rentCarHandler)

	testTable := []struct {
		name  string
		carID int
		token string
	}{
		{name: "Test 1", carID: 22, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
		{name: "Test 2", carID: 23, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.carID)
			req, err := http.NewRequest(http.MethodPost, "/car/"+id+"/rent", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			user, err := app.models.Users.GetForToken(data.ScopeAuthentication, testTable.token)
			if err != nil {
				t.Error(err)
			}
			req = app.contextSetUser(req, user)

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestReturnRentedCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodPut, "/car/:id/return", app.returnRentedCarHandler)

	testTable := []struct {
		name  string
		carID int
		token string
	}{
		{name: "Test 1", carID: 22, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
		{name: "Test 2", carID: 23, token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4"},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			id := strconv.Itoa(testTable.carID)
			req, err := http.NewRequest(http.MethodPut, "/car/"+id+"/return", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			user, err := app.models.Users.GetForToken(data.ScopeAuthentication, testTable.token)
			if err != nil {
				t.Error(err)
			}
			req = app.contextSetUser(req, user)

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func TestUpdateCarHandler(t *testing.T) {
	app := getConfig()
	router := httprouter.New()
	router.HandlerFunc(http.MethodPatch, "/car/:id", app.updateCarHandler)

	type input struct {
		Brand       string `json:"brand"`
		Description string `json:"description"`
		Color       string `json:"color,omitempty"`
		Year        int32  `json:"year,omitempty"`
		Price       int32  `json:"price"`
	}

	testTable := []struct {
		name      string
		carID     int
		userInput input
		token     string
	}{
		{
			name:  "Test 1",
			carID: 22,
			userInput: input{
				Brand:       "volvo",
				Description: "ft",
				Color:       "grey",
				Year:        1999,
				Price:       1001,
			},
			token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4",
		},
		{
			name:  "Test 2",
			carID: 23,
			userInput: input{
				Brand:       "mazda",
				Description: "ft",
				Color:       "yellow",
				Year:        2011,
				Price:       1000,
			},
			token: "4R2Q3HMXNJKR43AGX2S7XU4ZS4",
		},
	}

	for _, testTable := range testTable {
		t.Run(testTable.name, func(t *testing.T) {
			marshal, err := json.Marshal(testTable.userInput)
			if err != nil {
				t.Error(err)
			}

			id := strconv.Itoa(testTable.carID)
			req, err := http.NewRequest(http.MethodPatch, "/car/"+id, bytes.NewReader(marshal))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			user, err := app.models.Users.GetForToken(data.ScopeAuthentication, testTable.token)
			if err != nil {
				t.Error(err)
			}
			req = app.contextSetUser(req, user)

			handler := http.Handler(router)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}
