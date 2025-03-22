package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
)
type Film struct {
	ID          int64            `json:"id"`
	Title       string           `json:"title"`
	Year        int32            `json:"year"`
	Runtime     Runtime          `json:"runtime"`
	Genres      []Genre          `json:"genres"`
	Directors   []Director       `json:"directors"`
	Actors      []Actor          `json:"actors"`
	Rating      float32          `json:"rating"`
	Description string           `json:"description"`
	Img         string           `json:"image"`
	Version     int32            `json:"version"` // Add version field
}


type FilmModel struct {
	DB        *sql.DB
	Genres    GenreModel
	Actors    ActorModel
	Directors DirectorModel
}
func NewFilmModel(db *sql.DB) FilmModel {
	return FilmModel{
		DB:        db,
		Genres:    GenreModel{DB: db},
		Actors:    ActorModel{DB: db},
		Directors: DirectorModel{DB: db},
	}
}

func ValidateFilm(v *validator.Validator, film *Film) {
	v.Check(film.Title != "", "title", "must be provided")
	v.Check(len(film.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(film.Year != 0, "year", "must be provided")
	v.Check(film.Year >= 1888, "year", "must be greater than 1888")
	v.Check(film.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(film.Runtime != 0, "runtime", "must be provided")
	v.Check(film.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(film.Genres != nil, "genres", "must be provided")
	v.Check(len(film.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(film.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(film.Genres), "genres", "must not contain duplicate values")
	v.Check(validator.MatchesURL(film.Img), "image", "Must be an URL")
}

func (f Film) MarshalJSON() ([]byte, error) {
	var runtime string

	if f.Runtime != 0 {
		runtime = fmt.Sprintf("%d mins", f.Runtime)
	}

	type FilmALias Film

	aux := struct {
		FilmALias
		Runtime string `json:"runtime"`
	}{
		FilmALias(f),
		runtime,
	}

	return json.Marshal(aux)
}

func (model FilmModel) Get(id int64) (*Film, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query1 := `
		SELECT f.id, f.title, f.year, f.runtime, f.rating, f.description, f.image
		FROM films AS f
		WHERE f.id = $1
	`
	query2 := `
		SELECT g.id, g.name
		FROM film_genres
		INNER JOIN genres AS g ON g.id = film_genres.genre_id
		INNER JOIN films AS f ON f.id = film_genres.film_id
		WHERE f.id = $1
	`

	query3 := `
		SELECT a.id, a.name
		FROM film_actors
		INNER JOIN actors AS a ON a.id = film_actors.actor_id
		INNER JOIN films AS f ON f.id = film_actors.film_id
		WHERE f.id = $1
	`

	query4 := `
		SELECT d.id, d.name
		FROM film_directors
		INNER JOIN directors AS d ON d.id = film_directors.director_id
		INNER JOIN films AS f ON f.id = film_directors.film_id
		WHERE f.id = $1
	`

	tx, err := model.DB.Begin()
	defer tx.Rollback()
	if err != nil {
		return nil, err
	}

	var film Film
	var genres []Genre
	var actors []Actor
	var directors []Director

	err = tx.QueryRow(query1, id).Scan(&film.ID, &film.Title, &film.Year, &film.Runtime, &film.Rating, &film.Description, &film.Img)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, ErrRecordNotFound
		}else{
			return nil, err
		}
	}

	rows, err := tx.Query(query2, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var genre Genre
		err = rows.Scan(&genre.ID, &genre.Name)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	film.Genres = genres

	rows, err = tx.Query(query3, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var actor Actor
		err = rows.Scan(&actor.ID, &actor.Name)
		if err != nil {
			return nil, err
		}
		actors = append(actors, actor)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	film.Actors = actors

	rows, err = tx.Query(query4, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var director Director
		err = rows.Scan(&director.ID, &director.Name)
		if err != nil {
			return nil, err
		}
		directors = append(directors, director)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	film.Directors = directors
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func (model FilmModel) Insert(film *Film) error {
	// Start a transaction
	tx, err := model.DB.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	// Insert the film first
	query := `
		INSERT INTO films (title, year, runtime, rating, description, image, version) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	args := []any{film.Title, film.Year, film.Runtime, film.Rating, film.Description, film.Img, 1}
	err = tx.QueryRowContext(ctx, query, args...).Scan(&film.ID)
	if err != nil {
		return err
	}

	// Process directors with GetOrCreate and link them
	for i := range film.Directors {
		director, err := model.Directors.GetOrCreate(tx, film.Directors[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Directors[i] = *director
		err = director.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}

	// Process actors with GetOrCreate and link them
	for i := range film.Actors {
		actor, err := model.Actors.GetOrCreate(tx, film.Actors[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Actors[i] = *actor
		err = actor.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}

	// Process genres with GetOrCreate and link them
	for i := range film.Genres {
		genre, err := model.Genres.GetOrCreate(tx, film.Genres[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Genres[i] = *genre
		err = genre.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (model FilmModel) Update(film *Film) error {
	query1 := `
		UPDATE films
		SET title = $1, runtime = $2, year = $3, rating = $4, description = $5, image = $6, version = $8
		WHERE id = $7 AND version = $8
		RETURNING version
	`

	args := []any{
		film.Title, 
		film.Runtime, 
		film.Year,
		film.Rating,
		film.Description,
		film.Img,
		film.ID,
		film.Version,
	}	

	tx, err := model.DB.Begin()
	if err != nil{
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err = tx.QueryRowContext(ctx,query1, args...).Scan(&film.Version)
	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return ErrEditConflict
		}

		return err
	}
	
	for i := range film.Directors {
		director, err := model.Directors.GetOrCreate(tx, film.Directors[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Directors[i] = *director
		err = director.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}

	// Process actors with GetOrCreate and link them
	for i := range film.Actors {
		actor, err := model.Actors.GetOrCreate(tx, film.Actors[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Actors[i] = *actor
		err = actor.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}

	// Process genres with GetOrCreate and link them
	for i := range film.Genres {
		genre, err := model.Genres.GetOrCreate(tx, film.Genres[i].Name, ctx)
		if err != nil {
			return err
		}
		film.Genres[i] = *genre
		err = genre.LinkToFilm(tx, film, ctx)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil{
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		
		default:
			return err
		}

	}
	
	return nil
}

func (model FilmModel) Delete(id int64) error {
	if id < 1{
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM films
		WHERE id = $1
	`
	
	result, err := model.DB.Exec(query, id)
	if err != nil{
		return err
	}
	rowsNum, err := result.RowsAffected()
	if err != nil{
		return err
	}
	if rowsNum == 0{
		return ErrRecordNotFound
	}

	return nil
}


func (model FilmModel) Lock(id int64) error{
	query := `
		SELECT id FROM films
		WHERE id = $1
		FOR UPDATE
	`

	_, err := model.DB.Exec(query, id)

	return err
}

func (model FilmModel) GetAll(title string, genres []string, filters Filters) ([]*Film, error){
	query := `
		SELECT 
		f.id , f.title, f.year, f.runtime, f.rating, f.description, f.image, f.version,
		COALESCE(array_agg(DISTINCT g.name) FILTER (WHERE g.name IS NOT NULL), '{}') AS genres,
		COALESCE(array_agg(DISTINCT a.name) FILTER (WHERE a.name IS NOT NULL), '{}') AS actors,
		COALESCE(array_agg(DISTINCT d.name) FILTER (WHERE d.name IS NOT NULL), '{}') AS directors
		FROM 
			films f
		LEFT JOIN 
			film_genres fg ON f.id = fg.film_id
		LEFT JOIN 
			genres g ON fg.genre_id = g.id
		LEFT JOIN 
			film_actors fa ON f.id = fa.film_id
		LEFT JOIN 
			actors a ON fa.actor_id = a.id
		LEFT JOIN 
			film_directors fd ON f.id = fd.film_id
		LEFT JOIN 
			directors d ON fd.director_id = d.id
		GROUP BY 
			f.id;
		`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := model.DB.QueryContext(ctx, query)
	if err != nil{
		return nil, err
	}
	defer rows.Close()

	type input struct{
		ID          int64            `json:"id"`
		Title       string           `json:"title"`
		Year        int32            `json:"year"`
		Runtime     Runtime          `json:"runtime"`
		Genres      []string          `json:"genres"`
		Directors   []string       	`json:"directors"`
		Actors      []string          `json:"actors"`
		Rating      float32          `json:"rating"`
		Description string           `json:"description"`
		Img         string           `json:"image"`
		Version     int32            `json:"version"` // Add version field
	}
	films := []*Film{}

	for rows.Next(){
		var filmInput input
		var film Film
		err := rows.Scan(
			&filmInput.ID,
			&filmInput.Title,
			&filmInput.Year,
			&filmInput.Runtime,
			&filmInput.Rating,
			&filmInput.Description,
			&filmInput.Img,
			&filmInput.Version,
			&filmInput.Genres,
			&filmInput.Actors,
			&filmInput.Directors,

		)
		if err != nil{
			return nil, err
		}

		film.ID = filmInput.ID
		film.Title = filmInput.Title
		film.Year = filmInput.Year
		film.Runtime = filmInput.Runtime
		film.Rating = filmInput.Rating
		film.Description = filmInput.Description
		film.Img = filmInput.Img
		film.Version = filmInput.Version
		
		genres := make([]Genre, len(filmInput.Genres))
		for i, genre := range filmInput.Genres{
			genres[i] = Genre{Name: genre}
		}
		film.Genres = genres

		directors := make([]Director, len(filmInput.Directors))
		for i, director := range filmInput.Directors{
			directors[i] = Director{Name: director}
		}
		film.Directors = directors

		actors := make([]Actor, len(filmInput.Actors))
		for i, actor := range filmInput.Actors{
			actors[i] = Actor{Name: actor}
		}
		film.Actors = actors

		films = append(films, &film)
	}

	if err = rows.Err(); err != nil{
		return nil, err
	}

	return films, nil
}