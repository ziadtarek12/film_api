package models

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Director struct{
	ID uint `json:"id"`
	Name string `json:"name"`
}

func (d Director) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Name)
}



type DirectorModel struct{
	DB *sql.DB
}

func (m DirectorModel) GetOrCreate(tx *sql.Tx, name string, ctx context.Context) (*Director, error) {
	director := &Director{}
	
	query := `
		WITH new_director AS (
			INSERT INTO directors (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name FROM new_director
		UNION ALL
		SELECT id, name FROM directors WHERE name = $1
		LIMIT 1
	`
	
	err := tx.QueryRowContext(ctx, query, name).Scan(&director.ID, &director.Name)
	if err != nil {
		return nil, err
	}
	
	return director, nil
}

func (director Director) LinkToFilm(tx *sql.Tx, film *Film, ctx context.Context) error{
	query := `
		INSERT INTO film_directors (film_id, director_id) VALUES ($1, $2)
		ON CONFLICT (film_id, director_id) DO NOTHING
	`
	args := []any{film.ID, director.ID}

	_, err := tx.ExecContext(ctx, query, args...)

	return err
}