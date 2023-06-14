package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/car/:id", app.showCarHandler)
	router.HandlerFunc(http.MethodGet, "/cars", app.listCarHandler)
	router.HandlerFunc(http.MethodPost, "/car", app.createCarHandler)
	router.HandlerFunc(http.MethodDelete, "/car/:id", app.deleteCarHandler)

	router.HandlerFunc(http.MethodPost, "/car/:id/rent", app.rentCarHandler)
	router.HandlerFunc(http.MethodPut, "/car/:id/return", app.returnRentedCarHandler)

	router.HandlerFunc(http.MethodPost, "/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodGet, "/users/:id", app.showUserHandler)
	router.HandlerFunc(http.MethodDelete, "/users/:id", app.deleteUserHandler)

	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationTokenHandler)

	return router
}
