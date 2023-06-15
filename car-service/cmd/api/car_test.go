package main

import (
	"car-service/internal/jsonlog"
	"car-service/internal/model"
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
