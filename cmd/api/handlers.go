package main

import (
	"fmt"
	"net/http"
	"strconv"

	"filmapi.zeyadtarek.net/internals/models"
	"filmapi.zeyadtarek.net/internals/validator"
)



func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request){
	env := map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version": "1",
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil{
		app.errorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

}


func (app *application) getFilmHandler(w http.ResponseWriter, r *http.Request){
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil{
		app.errorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	film := models.Film{
		ID: uint(id),
	}

	env := map[string]any{
		"film": film,
	}
	app.writeJSON(w, http.StatusOK, env, nil)

}

func (app *application) createFilmHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string   `json:"title"`
		Year        int32    `json:"year"`
		Runtime     models.Runtime `json:"runtime"`
		Genres      []string `json:"genres"`
		Directors   []string `json:"directors"`
		Actors      []string `json:"actors"`
		Rating      float32  `json:"rating"`
		Description string   `json:"description"`
		Image       string   `json:"image"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create the film with the basic data
	film := &models.Film{
		Title:       input.Title,
		Year:        input.Year,
		Runtime:     input.Runtime,
		Rating:      input.Rating,
		Description: input.Description,
		Img:         input.Image,
	}

	// Initialize the slices with proper capacity
	film.Genres = make([]models.Genre, len(input.Genres))
	film.Directors = make([]models.Director, len(input.Directors))
	film.Actors = make([]models.Actor, len(input.Actors))

	// Convert string arrays to model types
	for i, name := range input.Genres {
		film.Genres[i] = models.Genre{Name: name}
	}

	for i, name := range input.Directors {
		film.Directors[i] = models.Director{Name: name}
	}

	for i, name := range input.Actors {
		film.Actors[i] = models.Actor{Name: name}
	}

	// Validate the film data
	v := validator.New()
	if models.ValidateFilm(v, film); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the film and its relationships
	err = app.models.Films.Insert(film)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/films/%d", film.ID))

	err = app.writeJSON(w, http.StatusCreated, map[string]any{"film": film}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}