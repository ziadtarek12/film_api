package models

import (
	"database/sql"
)

type Director struct{
	ID uint `json:"id"`
	Name string `json:"name"`
}

type DirectorModel struct{
	DB *sql.DB
}

func (m DirectorModel) GetOrCreate(tx *sql.Tx, name string) (*Director, error) {
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
	
	err := tx.QueryRow(query, name).Scan(&director.ID, &director.Name)
	if err != nil {
		return nil, err
	}
	
	return director, nil
}

func (director Director) LinkToFilm(tx *sql.Tx, film *Film) error{
	query := `
		INSERT INTO film_directors (film_id, director_id) VALUES ($1, $2)
		ON CONFLICT (film_id, director_id) DO NOTHING
	`
	args := []any{film.ID, director.ID}

	_, err := tx.Exec(query, args...)

	return err
}