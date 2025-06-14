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

func (app *application) welcomeHandler(w http.ResponseWriter, r *http.Request) {
	welcomeText := `
██╗    ██╗███████╗██╗      ██████╗ ██████╗ ███╗   ███╗███████╗
██║    ██║██╔════╝██║     ██╔════╝██╔═══██╗████╗ ████║██╔════╝
██║ █╗ ██║█████╗  ██║     ██║     ██║   ██║██╔████╔██║█████╗  
██║███╗██║██╔══╝  ██║     ██║     ██║   ██║██║╚██╔╝██║██╔══╝  
╚███╔███╔╝███████╗███████╗╚██████╗╚██████╔╝██║ ╚═╝ ██║███████╗
 ╚══╝╚══╝ ╚══════╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝

                    🎬 FILM API 🎬
                    
Welcome to the Film API - Your Gateway to Cinema Data!

═══════════════════════════════════════════════════════════════

📚 API Documentation:

🔍 Films Endpoints:
   GET    /v1/films           - List all films (with advanced filtering!)
   POST   /v1/films           - Create a new film  
   GET    /v1/films/{id}      - Get film by ID
   PATCH  /v1/films/{id}      - Update film
   DELETE /v1/films/{id}      - Delete film

🎯 Films Filtering & Searching:
   • Title Search:
     /v1/films?title=godfather
     
   • Genre Filtering:
     /v1/films?genres=action,drama
     /v1/films?genres=sci-fi
     
   • Director Search:
     /v1/films?directors=nolan,scorsese
     /v1/films?directors=kubrick
     
   • Actor Search:
     /v1/films?actors=dicaprio,pacino
     /v1/films?actors=deniro
     
   • Combined Filters:
     /v1/films?title=dark&genres=action&directors=nolan
     
   • Pagination:
     /v1/films?page=2&page_size=10
     
   • Sorting:
     /v1/films?sort=title          (A-Z)
     /v1/films?sort=-rating        (highest rated first)
     /v1/films?sort=year,-rating   (newest first, then by rating)
     
   Available sort fields: id, title, year, runtime, rating
   Use '-' prefix for descending order (e.g., -year, -rating)

👤 User Endpoints:
   POST   /v1/users           - Register new user
   PUT    /v1/users/activate  - Activate user account
   POST   /v1/tokens/authentication - Login

📋 Watchlist Endpoints:
   GET    /v1/watchlist       - Get user's watchlist
   POST   /v1/watchlist       - Add film to watchlist
   GET    /v1/watchlist/{id}  - Get watchlist entry
   PATCH  /v1/watchlist/{id}  - Update watchlist entry
   DELETE /v1/watchlist/{id}  - Remove from watchlist

🎯 Watchlist Filtering:
   • By status:
     /v1/watchlist?watched=true    (watched films)
     /v1/watchlist?watched=false   (unwatched films)
     
   • By priority:
     /v1/watchlist?priority=5      (high priority films)
     
   • Sorting:
     /v1/watchlist?sort=priority   (by priority)
     /v1/watchlist?sort=-added_at  (newest additions first)

💡 System:
   GET    /v1/healthcheck     - API health status

═══════════════════════════════════════════════════════════════

🔐 Authentication: Most endpoints require authentication tokens
📄 Format: All responses are in JSON format
🌍 Version: v1
📊 Status: Online and Ready

💡 Pro Tips:
   • Combine multiple filters for precise searches
   • Use pagination for large result sets
   • Sort results to find what you need faster
   • Try partial title matches for flexible searching

Happy Coding! 🚀

═══════════════════════════════════════════════════════════════
`

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(welcomeText))
}

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

	err = app.models.Permissions.AddForUser(user.ID, "films:read")
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
			v.AddError("token", "invalid or expired actvation token")
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

	app.models.Permissions.AddForUser(user.ID, "films:read")
	app.models.Permissions.AddForUser(user.ID, "films:write")
	err = app.writeJSON(w, http.StatusCreated, map[string]any{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Watchlist handlers

func (app *application) addToWatchlistHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FilmID   int64  `json:"film_id"`
		Notes    string `json:"notes"`
		Priority int    `json:"priority"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)

	// Check if film exists
	_, err = app.models.Films.Get(input.FilmID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create watchlist entry
	entry := &models.Watchlist{
		UserID:   user.ID,
		FilmID:   input.FilmID,
		Notes:    input.Notes,
		Priority: input.Priority,
		Watched:  false,
	}

	// Set default priority if not provided
	if entry.Priority == 0 {
		entry.Priority = 5
	}

	v := validator.New()
	if models.ValidateWatchlistEntry(v, entry); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Watchlist.Insert(entry)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateWatchlistEntry):
			v.AddError("film_id", "film is already in your watchlist")
			app.faliedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Get the full entry with film details
	fullEntry, err := app.models.Watchlist.Get(user.ID, entry.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/watchlist/%d", entry.ID))

	err = app.writeJSON(w, http.StatusCreated, map[string]any{"watchlist_entry": fullEntry}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getWatchlistHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Watched  *bool
		Priority int
		Filters  models.Filters
	}

	v := validator.New()
	queryString := r.URL.Query()

	// Parse watched filter
	watchedStr := app.readString(queryString, "watched", "")
	if watchedStr != "" {
		switch watchedStr {
		case "true":
			watched := true
			input.Watched = &watched
		case "false":
			watched := false
			input.Watched = &watched
		default:
			v.AddError("watched", "must be 'true' or 'false'")
		}
	}

	input.Priority = app.readInt(queryString, "priority", 0, v)
	input.Filters.Page = app.readInt(queryString, "page", 1, v)
	input.Filters.PageSize = app.readInt(queryString, "page_size", 20, v)
	input.Filters.SortValues = app.readCSV(queryString, "sort", []string{})
	input.Filters.SortSafelist = []string{"id", "added_at", "priority", "watched", "-id", "-added_at", "-priority", "-watched"}

	if models.ValidateFilters(v, input.Filters); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	entries, metadata, err := app.models.Watchlist.GetAll(user.ID, input.Watched, input.Priority, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"watchlist": entries, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getWatchlistEntryHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	entry, err := app.models.Watchlist.Get(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"watchlist_entry": entry}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateWatchlistEntryHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	// Get the current entry
	entry, err := app.models.Watchlist.Get(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Notes    *string `json:"notes"`
		Priority *int    `json:"priority"`
		Watched  *bool   `json:"watched"`
		Rating   *int    `json:"rating"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Apply partial updates
	if input.Notes != nil {
		entry.Notes = *input.Notes
	}
	if input.Priority != nil {
		entry.Priority = *input.Priority
	}
	if input.Watched != nil {
		entry.Watched = *input.Watched
		if *input.Watched && entry.WatchedAt == nil {
			now := time.Now()
			entry.WatchedAt = &now
		} else if !*input.Watched {
			entry.WatchedAt = nil
			entry.Rating = nil // Clear rating if marking as unwatched
		}
	}
	if input.Rating != nil {
		entry.Rating = input.Rating
		// If rating is provided, mark as watched
		if !entry.Watched {
			entry.Watched = true
			if entry.WatchedAt == nil {
				now := time.Now()
				entry.WatchedAt = &now
			}
		}
	}

	v := validator.New()
	if models.ValidateWatchlistEntry(v, entry); !v.Valid() {
		app.faliedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Watchlist.Update(entry)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Get the updated entry with film details
	fullEntry, err := app.models.Watchlist.Get(user.ID, entry.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"watchlist_entry": fullEntry}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) removeFromWatchlistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)

	err = app.models.Watchlist.Delete(user.ID, id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, map[string]any{"message": "watchlist entry removed successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
