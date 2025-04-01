package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/lib/pq"
)

type Film struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Year        int32      `json:"year"`
	Runtime     Runtime    `json:"runtime"`
	Genres      []Genre    `json:"genres"`
	Directors   []Director `json:"directors"`
	Actors      []Actor    `json:"actors"`
	Rating      float32    `json:"rating"`
	Description string     `json:"description"`
	Img         string     `json:"image"`
	Version     int32      `json:"version"` 
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

	query := `
		SELECT 
		f.id, f.title, f.year, f.runtime, f.rating, f.description, f.image, f.version,
		(SELECT array_agg(g.name) FROM film_genres fg JOIN genres g ON fg.genre_id = g.id WHERE fg.film_id = f.id) AS genres,
		(SELECT array_agg(a.name) FROM film_actors fa JOIN actors a ON fa.actor_id = a.id WHERE fa.film_id = f.id) AS actors,
		(SELECT array_agg(d.name) FROM film_directors fd JOIN directors d ON fd.director_id = d.id WHERE fd.film_id = f.id) AS directors
		FROM films f
		WHERE f.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var film Film
	var genres []string
	var actors []string
	var directors []string

	err := model.DB.QueryRowContext(ctx, query, id).Scan(
		&film.ID,
		&film.Title,
		&film.Year,
		&film.Runtime,
		&film.Rating,
		&film.Description,
		&film.Img,
		&film.Version,
		pq.Array(&genres),
		pq.Array(&actors),
		pq.Array(&directors),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	// Convert string arrays to respective types
	film.Genres = make([]Genre, len(genres))
	for i, genre := range genres {
		film.Genres[i] = Genre{Name: genre}
	}

	film.Actors = make([]Actor, len(actors))
	for i, actor := range actors {
		film.Actors[i] = Actor{Name: actor}
	}

	film.Directors = make([]Director, len(directors))
	for i, director := range directors {
		film.Directors[i] = Director{Name: director}
	}

	return &film, nil
}

func (model FilmModel) Insert(film *Film) error {
	tx, err := model.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Insert film
	query := `INSERT INTO films (title, year, runtime, rating, description, image, version) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = tx.QueryRowContext(ctx, query, film.Title, film.Year, film.Runtime, film.Rating, film.Description, film.Img, 1).Scan(&film.ID)
	if err != nil {
		return err
	}

	// Batch insert related entities
	if err := model.batchInsertRelations(tx, ctx, film); err != nil {
		return err
	}

	return tx.Commit()
}

func (model FilmModel) Update(film *Film) error {
	tx, err := model.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE films
		SET title = $1, year = $2, runtime = $3, rating = $4, description = $5, image = $6, version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version
	`

	args := []interface{}{
		film.Title,
		film.Year,
		film.Runtime,
		film.Rating,
		film.Description,
		film.Img,
		film.ID,
		film.Version,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&film.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}
		return err
	}

	if err := model.batchInsertRelations(tx, ctx, film); err != nil {
		return err
	}

	return tx.Commit()
}

func (model FilmModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM films WHERE id = $1
	`

	result, err := model.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (model FilmModel) GetAll(title string, genres []string, actors []string, directors []string,filters Filters) ([]*Film, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(),
		f.*,
		(SELECT array_agg(g.name) FROM film_genres fg 
		JOIN genres g ON fg.genre_id = g.id 
		WHERE fg.film_id = f.id) AS genres,
		(SELECT array_agg(a.name) FROM film_actors fa 
		JOIN actors a ON fa.actor_id = a.id 
		WHERE fa.film_id = f.id) AS actors,
		(SELECT array_agg(d.name) FROM film_directors fd 
		JOIN directors d ON fd.director_id = d.id 
		WHERE fd.film_id = f.id) AS directors
		FROM films f
		WHERE (to_tsvector('simple', f.title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND ( EXISTS (
		SELECT 1 FROM film_genres fg 
		JOIN genres g ON fg.genre_id = g.id 
		WHERE fg.film_id = f.id 
		AND g.name = ANY($2)
		) OR $2 = ARRAY[]::text[]
		)
		AND ( EXISTS ( SELECT 1 
		FROM film_actors fa 
		JOIN actors a ON fa.actor_id = a.id 
		WHERE fa.film_id = f.id 
		AND a.name = ANY($3)
		) OR $3 = ARRAY[]::text[]
		)
		AND (
		EXISTS (
		SELECT 1 
		FROM film_directors fd 
		JOIN directors d ON fd.director_id = d.id 
		WHERE fd.film_id = f.id 
		AND d.name = ANY($4)
		) OR $4 = ARRAY[]::text[]
		)
		ORDER BY %s id ASC
		LIMIT $5 OFFSET $6
		`, filters.sortColumn())
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := model.DB.QueryContext(ctx, query, title, pq.Array(genres), pq.Array(actors), pq.Array(directors), filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{},err
	}
	defer rows.Close()

	type input struct {
		ID          int64    `json:"id"`
		Title       string   `json:"title"`
		Year        int32    `json:"year"`
		Runtime     Runtime  `json:"runtime"`
		Genres      []string `json:"genres"`
		Directors   []string `json:"directors"`
		Actors      []string `json:"actors"`
		Rating      float32  `json:"rating"`
		Description string   `json:"description"`
		Img         string   `json:"image"`
		Version     int32    `json:"version"` 
	}
	films := []*Film{}
	totalRecords := 0
	for rows.Next() {
		var filmInput input
		var film Film
		err := rows.Scan(
			&totalRecords,
			&filmInput.ID,
			&filmInput.Title,
			&filmInput.Year,
			&filmInput.Runtime,
			&filmInput.Rating,
			&filmInput.Description,
			&filmInput.Img,
			&filmInput.Version,
			pq.Array(&filmInput.Genres),
			pq.Array(&filmInput.Actors),
			pq.Array(&filmInput.Directors),
		)
		if err != nil {
			return nil, Metadata{},err
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
		for i, genre := range filmInput.Genres {
			genres[i] = Genre{Name: genre}
		}
		film.Genres = genres

		directors := make([]Director, len(filmInput.Directors))
		for i, director := range filmInput.Directors {
			directors[i] = Director{Name: director}
		}
		film.Directors = directors

		actors := make([]Actor, len(filmInput.Actors))
		for i, actor := range filmInput.Actors {
			actors[i] = Actor{Name: actor}
		}
		film.Actors = actors

		films = append(films, &film)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{},err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return films, metadata, nil
}



func (model FilmModel) batchInsertRelations(tx *sql.Tx, ctx context.Context, film *Film) error {
	// Batch insert directors
	if len(film.Directors) > 0 {
		directorNames := make([]string, len(film.Directors))
		for i, d := range film.Directors {
			directorNames[i] = d.Name
		}

		query := `
			WITH inserted_directors AS (
				INSERT INTO directors (name)
				SELECT unnest($1::text[])
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id, name
			)
			INSERT INTO film_directors (film_id, director_id)
			SELECT $2, id FROM inserted_directors
			ON CONFLICT DO NOTHING
		`
		_, err := tx.ExecContext(ctx, query, pq.Array(directorNames), film.ID)
		if err != nil {
			return err
		}
	}

	// Batch insert actors
	if len(film.Actors) > 0 {
		actorNames := make([]string, len(film.Actors))
		for i, a := range film.Actors {
			actorNames[i] = a.Name
		}

		query := `
			WITH inserted_actors AS (
				INSERT INTO actors (name)
				SELECT unnest($1::text[])
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id, name
			)
			INSERT INTO film_actors (film_id, actor_id)
			SELECT $2, id FROM inserted_actors
			ON CONFLICT DO NOTHING
		`
		_, err := tx.ExecContext(ctx, query, pq.Array(actorNames), film.ID)
		if err != nil {
			return err
		}
	}

	// Batch insert genres
	if len(film.Genres) > 0 {
		genreNames := make([]string, len(film.Genres))
		for i, g := range film.Genres {
			genreNames[i] = g.Name
		}

		query := `
			WITH inserted_genres AS (
				INSERT INTO genres (name)
				SELECT unnest($1::text[])
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id, name
			)
			INSERT INTO film_genres (film_id, genre_id)
			SELECT $2, id FROM inserted_genres
			ON CONFLICT DO NOTHING
		`
		_, err := tx.ExecContext(ctx, query, pq.Array(genreNames), film.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
