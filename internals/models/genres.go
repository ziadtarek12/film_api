package models

import "database/sql"

type Genre struct{
	ID uint `json:"id"`
	Name string `json:"name"`
}

type GenreModel struct {
	DB *sql.DB
}

func (m GenreModel) GetOrCreate(tx *sql.Tx, name string) (*Genre, error) {
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
	
	err := tx.QueryRow(query, name).Scan(&genre.ID, &genre.Name)
	if err != nil {
		return nil, err
	}
	
	return genre, nil
}

func (genre Genre) LinkToFilm(tx *sql.Tx, film *Film) error{
	query := `
		INSERT INTO film_genres (film_id, genre_id) VALUES ($1, $2)
		ON CONFLICT (film_id, genre_id) DO NOTHING
	`
	args := []any{film.ID, genre.ID}

	_, err := tx.Exec(query, args...)

	return err
}