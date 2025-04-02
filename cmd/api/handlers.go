package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"filmapi.zeyadtarek.net/internals/models"
	"filmapi.zeyadtarek.net/internals/validator"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	env := map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     "1",
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) getFilmHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	film, err := app.models.Films.Get(int64(id))
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		} else {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	env := map[string]any{
		"film": *film,
	}
	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createFilmHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string         `json:"title"`
		Year        int32          `json:"year"`
		Runtime     models.Runtime `json:"runtime"`
		Genres      []string       `json:"genres"`
		Directors   []string       `json:"directors"`
		Actors      []string       `json:"actors"`
		Rating      float32        `json:"rating"`
		Description string         `json:"description"`
		Image       string         `json:"image"`
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

func (app *application) updateFilmHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the latest version of the film
	film, err := app.models.Films.Get(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Title       *string         `json:"title"`
		Year        *int32          `json:"year"`
		Runtime     *models.Runtime `json:"runtime"`
		Genres      *[]string       `json:"genres"`
		Directors   *[]string       `json:"directors"`
		Actors      *[]string       `json:"actors"`
		Rating      *float32        `json:"rating"`
		Description *string         `json:"description"`
		Img         *string         `json:"image"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Apply partial updates
	if input.Title != nil {
		film.Title = *input.Title
	}
	if input.Year != nil {
		film.Year = *input.Year
	}
	if input.Runtime != nil {
		film.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		film.Genres = make([]models.Genre, len(*input.Genres))
		for i, genre := range *input.Genres {
			film.Genres[i] = models.Genre{Name: genre}
		}
	}
	if input.Directors != nil {
		film.Directors = make([]models.Director, len(*input.Directors))
		for i, director := range *input.Directors {
			film.Directors[i] = models.Director{Name: director}
		}
	}
	if input.Actors != nil {
		film.Actors = make([]models.Actor, len(*input.Actors))
		for i, actor := range *input.Actors {
			film.Actors[i] = models.Actor{Name: actor}
		}
	}
	if input.Rating != nil {
		film.Rating = *input.Rating
	}
	if input.Description != nil {
		film.Description = *input.Description
	}
	if input.Img != nil {
		film.Img = *input.Img
	}

	// Retry the update
	err = app.models.Films.Update(film)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"film": film}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFilmHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Films.Delete(int64(id))
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		} else {
			app.serverErrorResponse(w, r, err)
			return
		}

	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"message": "movie deleted succesfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) ListFilmsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title     string
		Genres    []string
		Directors []string
		Actors    []string
		Filters   models.Filters
	}

	v := validator.New()
	queryString := r.URL.Query()
	input.Title = app.readString(queryString, "title", "")
	input.Actors = app.readCSV(queryString, "actors", []string{})
	input.Directors = app.readCSV(queryString, "directors", []string{})
	input.Genres = app.readCSV(queryString, "genres", []string{})
	input.Filters.Page = app.readInt(queryString, "page", 1, v)
	input.Filters.PageSize = app.readInt(queryString, "page_size", 20, v)
	input.Filters.SortValues = app.readCSV(queryString, "sort", []string{})
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "rating", "-id", "-title", "-year", "-runtime", "-rating"}

	if models.ValidateFilters(v, input.Filters); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	films, metadata, err := app.models.Films.GetAll(input.Title, input.Genres, input.Actors, input.Directors, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"films": films, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &models.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if models.ValidateUser(v, user); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.faliedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Generate activation token
	activationToken, err := app.models.Tokens.New(user.ID, 24*time.Hour, models.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Include both user and activation token in response
	response := map[string]any{
		"user": user,
		"activation_token": struct {
			Token  string    `json:"token"`
			Expiry time.Time `json:"expiry"`
		}{
			Token:  activationToken.Plaintext,
			Expiry: activationToken.Expiry,
		},
	}

	err = app.writeJSON(w, http.StatusCreated, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if models.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(models.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			v.AddError("token", "invalid or expired actviation token")
			app.faliedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	user.Activated = true
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(models.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	models.ValidateEmail(v, input.Email)
	models.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, models.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, map[string]any{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
