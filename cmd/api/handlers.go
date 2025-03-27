package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	films, metadata ,err := app.models.Films.GetAll(input.Title, input.Genres, input.Actors, input.Directors, input.Filters)
	if err != nil{
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"films": films, "metadata": metadata}, nil)
	if err != nil{
		app.serverErrorResponse(w, r, err)
		return
	}

}
