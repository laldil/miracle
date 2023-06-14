package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	//router.HandlerFunc(http.MethodGet, "/car/:id", app.showCarHandler)
	//router.HandlerFunc(http.MethodGet, "/cars", app.listCarHandler)
	//router.HandlerFunc(http.MethodPost, "/car", app.requireActivatedUser(app.createCarHandler))
	//router.HandlerFunc(http.MethodPatch, "/car/:id", app.requireActivatedUser(app.updateCarHandler))
	//router.HandlerFunc(http.MethodDelete, "/car/:id", app.requireActivatedUser(app.deleteCarHandler))
	//
	//router.HandlerFunc(http.MethodPost, "/car/:id/rent", app.requireActivatedUser(app.rentCarHandler))
	//router.HandlerFunc(http.MethodPut, "/car/:id/return", app.requireActivatedUser(app.returnRentedCarHandler))

	router.HandlerFunc(http.MethodPost, "/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodGet, "/users/:id", app.showUserHandler)
	router.HandlerFunc(http.MethodDelete, "/users/:id", app.requireAdminRole(app.deleteUserHandler))
	router.HandlerFunc(http.MethodPatch, "/users/:id", app.requireAdminRole(app.setRoleHandler))

	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.authenticate(router))
}
