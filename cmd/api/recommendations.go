package main

import (
"errors"
"net/http"

"filmapi.zeyadtarek.net/internals/models"
)

func (app *application) getRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	films, err := app.models.Films.GetRecommendations(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recommendations": films}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
