package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	// Welcome page
	router.Handle("GET /", http.HandlerFunc(app.welcomeHandler))

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

	// Watchlist routes (require authentication)
	router.Handle("GET /v1/watchlist", app.requireActivatedUser(http.HandlerFunc(app.getWatchlistHandler)))
	router.Handle("POST /v1/watchlist", app.requireActivatedUser(http.HandlerFunc(app.addToWatchlistHandler)))
	router.Handle("GET /v1/watchlist/{id}", app.requireActivatedUser(http.HandlerFunc(app.getWatchlistEntryHandler)))
	router.Handle("PATCH /v1/watchlist/{id}", app.requireActivatedUser(http.HandlerFunc(app.updateWatchlistEntryHandler)))
	router.Handle("DELETE /v1/watchlist/{id}", app.requireActivatedUser(http.HandlerFunc(app.removeFromWatchlistHandler)))

	// Chain middleware
	return app.chainMiddleware(router, app.recoverPanic, app.rateLimit, app.authenticate, app.enableCORS)
}
