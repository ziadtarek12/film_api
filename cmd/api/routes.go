package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	// Create main router
	mainRouter := http.NewServeMux()

	// Create public router
	publicRouter := http.NewServeMux()
	publicHandlers := map[string]func(http.ResponseWriter, *http.Request){
		"GET /healthcheck":            app.healthCheckHandler,
		"POST /user":                  app.createUserHandler,
		"PUT /users/activated":        app.activateUserHandler,
		"POST /tokens/authentication": app.createAuthenticationTokenHandler,
	}

	// Create films router
	filmsRouter := http.NewServeMux()
	filmsHandlers := map[string]func(http.ResponseWriter, *http.Request){
		"GET /{id}":    app.getFilmHandler,
		"POST /":       app.createFilmHandler,
		"PATCH /{id}":  app.updateFilmHandler,
		"DELETE /{id}": app.deleteFilmHandler,
		"GET /": app.ListFilmsHandler,
	}

	// Add handlers to their respective routers
	for url, handler := range publicHandlers {
		publicRouter.Handle(url, app.chainMiddleware(http.HandlerFunc(handler), app.recoverPanic, app.rateLimit, app.authenticate))
	}

	for url, handler := range filmsHandlers {
		filmsRouter.Handle(url, app.chainMiddleware(http.HandlerFunc(handler), app.recoverPanic, app.rateLimit, app.authenticate, app.requireActivatedUser))
	}

	// Mount the routers with their respective prefixes
	mainRouter.Handle("/v1/", http.StripPrefix("/v1", publicRouter))

	mainRouter.Handle("/v1/films/", http.StripPrefix("/v1/films", filmsRouter))
	mainRouter.Handle("/v1/films", app.ensureTrailingSlash(http.StripPrefix("/v1/films", filmsRouter)))
	return mainRouter
}
