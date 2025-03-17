package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"GET /v1/healthcheck": app.healthCheckHandler,
		"GET /v1/film/{id}":   app.getFilmHandler,
		"POST /v1/films":      app.createFilmHandler,
	}
	for url, handler := range handlers {
		router.Handle(url, http.HandlerFunc(handler))
	}

	return router
}
