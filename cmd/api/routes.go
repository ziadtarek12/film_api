package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	// Healthcheck
	router.Handle("GET /v1/healthcheck", http.HandlerFunc(app.healthCheckHandler))

	// User routes
	router.Handle("POST /v1/users", http.HandlerFunc(app.createUserHandler))
	router.Handle("PUT /v1/users/activate", http.HandlerFunc(app.activateUserHandler))
	router.Handle("POST /v1/tokens/authentication", http.HandlerFunc(app.createAuthenticationTokenHandler))

	// Films routes
	router.Handle("GET /v1/films", app.requirePermission("films:read", http.HandlerFunc(app.ListFilmsHandler)))
	router.Handle("POST /v1/films", app.requirePermission("films:write", http.HandlerFunc(app.createFilmHandler)))
	router.Handle("GET /v1/films/{id}", app.requirePermission("films:read", http.HandlerFunc(app.getFilmHandler)))
	router.Handle("PATCH /v1/films/{id}", app.requirePermission("films:write", http.HandlerFunc(app.updateFilmHandler)))
	router.Handle("DELETE /v1/films/{id}", app.requirePermission("films:write", http.HandlerFunc(app.deleteFilmHandler)))

	// Chain middleware
	return app.chainMiddleware(router, app.recoverPanic, app.rateLimit, app.authenticate)
}
