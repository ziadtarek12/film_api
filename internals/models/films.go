package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
)

type Film struct{
	ID uint `json:"id"`
	Title string `json:"title"`
	Runtime Runtime  `json:"-"`
	Year 	int32   `json:"year"`
	Rating float32 `json:"rating"`
	Genres []Genre `json:"genres"`
	Actors []Actor `json:"actors"`
	Directors []Director `json:"directors"`
	Description string `json:"description"`
	Img	string `json:"image"`
}

type FilmModel struct{
	DB *sql.DB
	Genres GenreModel
	Actors ActorModel
	Directors DirectorModel
}

func NewFilmModel(db *sql.DB) FilmModel {
	return FilmModel{
		DB: db,
		Genres: GenreModel{DB: db},
		Actors: ActorModel{DB: db},
		Directors: DirectorModel{DB: db},
	}
}

func ValidateFilm(v *validator.Validator, film *Film){
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
}	

func (f Film) MarshalJSON() ([]byte, error){
	var runtime string

	if f.Runtime != 0{
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


func (model FilmModel) Get(id int64) (*Film, error){
	return nil, nil
}

func (model FilmModel) Insert(film *Film) error {
	// Start a transaction
	tx, err := model.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback if we encounter an error

	// Insert the film first
	query := `
		INSERT INTO films (title, year, runtime, rating, description, image) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	args := []any{film.Title, film.Year, film.Runtime, film.Rating, film.Description, film.Img}
	err = tx.QueryRow(query, args...).Scan(&film.ID)
	if err != nil {
		return err
	}

	// Process directors with GetOrCreate and link them
	for i := range film.Directors {
		director, err := model.Directors.GetOrCreate(tx, film.Directors[i].Name)
		if err != nil {
			return err
		}
		film.Directors[i] = *director
		err = director.LinkToFilm(tx, film)
		if err != nil {
			return err
		}
	}

	// Process actors with GetOrCreate and link them
	for i := range film.Actors {
		actor, err := model.Actors.GetOrCreate(tx, film.Actors[i].Name)
		if err != nil {
			return err
		}
		film.Actors[i] = *actor
		err = actor.LinkToFilm(tx, film)
		if err != nil {
			return err
		}
	}

	// Process genres with GetOrCreate and link them
	for i := range film.Genres {
		genre, err := model.Genres.GetOrCreate(tx, film.Genres[i].Name)
		if err != nil {
			return err
		}
		film.Genres[i] = *genre
		err = genre.LinkToFilm(tx, film)
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

func (model FilmModel) Update(film *Film) error{
	return nil
}

func (model FilmModel) Delete(id int64) error{
	return nil
}
