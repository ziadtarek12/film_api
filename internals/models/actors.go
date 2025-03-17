package models

import "database/sql"

type Actor struct{
	ID uint `json:"id"`
	Name string `json:"name"`
}

type ActorModel struct {
	DB *sql.DB
}

func (m ActorModel) GetOrCreate(tx *sql.Tx, name string) (*Actor, error) {
	actor := &Actor{}
	
	query := `
		WITH new_actor AS (
			INSERT INTO actors (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id, name
		)
		SELECT id, name FROM new_actor
		UNION ALL
		SELECT id, name FROM actors WHERE name = $1
		LIMIT 1
	`
	
	err := tx.QueryRow(query, name).Scan(&actor.ID, &actor.Name)
	if err != nil {
		return nil, err
	}
	
	return actor, nil
}

func (actor Actor) LinkToFilm(tx *sql.Tx, film *Film) error{
	query := `
		INSERT INTO film_actors (film_id, actor_id) VALUES ($1, $2)
		ON CONFLICT (film_id, actor_id) DO NOTHING
	`
	args := []any{film.ID, actor.ID}

	_, err := tx.Exec(query, args...)

	return err
}