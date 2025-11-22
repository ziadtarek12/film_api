package models

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Genre struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (g Genre) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Name)
}

type GenreModel struct {
	DB *sql.DB
}

func (m GenreModel) GetOrCreate(tx *sql.Tx, name string, ctx context.Context) (*Genre, error) {
	genre := &Genre{}

	query := `
		WITH new_genre AS (
			INSERT INTO genres (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name FROM new_genre
		UNION ALL
		SELECT id, name FROM genres WHERE name = $1
		LIMIT 1
	`

	err := tx.QueryRowContext(ctx, query, name).Scan(&genre.ID, &genre.Name)
	if err != nil {
		return nil, err
	}

	return genre, nil
}

func (genre Genre) LinkToFilm(tx *sql.Tx, film *Film, ctx context.Context) error {
	query := `
		INSERT INTO film_genres (film_id, genre_id) VALUES ($1, $2)
		ON CONFLICT (film_id, genre_id) DO NOTHING
	`
	args := []any{film.ID, genre.ID}

	_, err := tx.ExecContext(ctx, query, args...)

	return err
}
