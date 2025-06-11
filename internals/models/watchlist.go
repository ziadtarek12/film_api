package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/lib/pq"
)

type Watchlist struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	FilmID    int64      `json:"film_id"`
	Film      *Film      `json:"film,omitempty"`
	AddedAt   time.Time  `json:"added_at"`
	Notes     string     `json:"notes"`
	Priority  int        `json:"priority"`
	Watched   bool       `json:"watched"`
	WatchedAt *time.Time `json:"watched_at,omitempty"`
	Rating    *int       `json:"rating,omitempty"`
	Version   int        `json:"version"`
}

type WatchlistModel struct {
	DB *sql.DB
}

func ValidateWatchlistEntry(v *validator.Validator, entry *Watchlist) {
	v.Check(entry.FilmID > 0, "film_id", "must be provided and greater than 0")
	v.Check(entry.Priority >= 1 && entry.Priority <= 10, "priority", "must be between 1 and 10")
	v.Check(len(entry.Notes) <= 1000, "notes", "must not be more than 1000 characters long")

	if entry.Rating != nil {
		v.Check(*entry.Rating >= 1 && *entry.Rating <= 10, "rating", "must be between 1 and 10")
	}

	if entry.Watched && entry.WatchedAt == nil {
		entry.WatchedAt = &time.Time{}
		*entry.WatchedAt = time.Now()
	}
}

func (m WatchlistModel) Insert(entry *Watchlist) error {
	query := `
		INSERT INTO watchlist (user_id, film_id, notes, priority, watched, watched_at, rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, added_at, version
	`

	args := []any{
		entry.UserID,
		entry.FilmID,
		entry.Notes,
		entry.Priority,
		entry.Watched,
		entry.WatchedAt,
		entry.Rating,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&entry.ID, &entry.AddedAt, &entry.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "watchlist_user_film_unique"`:
			return ErrDuplicateWatchlistEntry
		default:
			return err
		}
	}

	return nil
}

func (m WatchlistModel) Get(userID, entryID int64) (*Watchlist, error) {
	if entryID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT w.id, w.user_id, w.film_id, w.added_at, w.notes, w.priority, 
			   w.watched, w.watched_at, w.rating, w.version,
			   f.title, f.year, f.runtime, f.rating as film_rating, f.description, f.image, f.version as film_version,
			   (SELECT array_agg(g.name) FROM film_genres fg JOIN genres g ON fg.genre_id = g.id WHERE fg.film_id = f.id) AS genres,
			   (SELECT array_agg(a.name) FROM film_actors fa JOIN actors a ON fa.actor_id = a.id WHERE fa.film_id = f.id) AS actors,
			   (SELECT array_agg(d.name) FROM film_directors fd JOIN directors d ON fd.director_id = d.id WHERE fd.film_id = f.id) AS directors
		FROM watchlist w
		INNER JOIN films f ON w.film_id = f.id
		WHERE w.id = $1 AND w.user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var entry Watchlist
	var film Film
	var genres, actors, directors []string

	err := m.DB.QueryRowContext(ctx, query, entryID, userID).Scan(
		&entry.ID,
		&entry.UserID,
		&entry.FilmID,
		&entry.AddedAt,
		&entry.Notes,
		&entry.Priority,
		&entry.Watched,
		&entry.WatchedAt,
		&entry.Rating,
		&entry.Version,
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Set film data
	film.ID = entry.FilmID
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
	entry.Film = &film

	return &entry, nil
}

func (m WatchlistModel) GetAll(userID int64, watched *bool, priority int, filters Filters) ([]*Watchlist, Metadata, error) {
	whereClause := "w.user_id = $1"
	args := []any{userID}
	argCount := 1

	if watched != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND w.watched = $%d", argCount)
		args = append(args, *watched)
	}

	if priority > 0 {
		argCount++
		whereClause += fmt.Sprintf(" AND w.priority = $%d", argCount)
		args = append(args, priority)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(),
			   w.id, w.user_id, w.film_id, w.added_at, w.notes, w.priority, 
			   w.watched, w.watched_at, w.rating, w.version,
			   f.title, f.year, f.runtime, f.rating as film_rating, f.description, f.image, f.version as film_version,
			   (SELECT array_agg(g.name) FROM film_genres fg JOIN genres g ON fg.genre_id = g.id WHERE fg.film_id = f.id) AS genres,
			   (SELECT array_agg(a.name) FROM film_actors fa JOIN actors a ON fa.actor_id = a.id WHERE fa.film_id = f.id) AS actors,
			   (SELECT array_agg(d.name) FROM film_directors fd JOIN directors d ON fd.director_id = d.id WHERE fd.film_id = f.id) AS directors
		FROM watchlist w
		INNER JOIN films f ON w.film_id = f.id
		WHERE %s
		ORDER BY %s w.added_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, filters.sortColumn(), argCount+1, argCount+2)

	args = append(args, filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	watchlist := []*Watchlist{}
	totalRecords := 0

	for rows.Next() {
		var entry Watchlist
		var film Film
		var genres, actors, directors []string

		err := rows.Scan(
			&totalRecords,
			&entry.ID,
			&entry.UserID,
			&entry.FilmID,
			&entry.AddedAt,
			&entry.Notes,
			&entry.Priority,
			&entry.Watched,
			&entry.WatchedAt,
			&entry.Rating,
			&entry.Version,
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
			return nil, Metadata{}, err
		}

		// Set film data
		film.ID = entry.FilmID
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
		entry.Film = &film

		watchlist = append(watchlist, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return watchlist, metadata, nil
}

func (m WatchlistModel) Update(entry *Watchlist) error {
	query := `
		UPDATE watchlist
		SET notes = $1, priority = $2, watched = $3, watched_at = $4, rating = $5, version = version + 1
		WHERE id = $6 AND user_id = $7 AND version = $8
		RETURNING version
	`

	args := []any{
		entry.Notes,
		entry.Priority,
		entry.Watched,
		entry.WatchedAt,
		entry.Rating,
		entry.ID,
		entry.UserID,
		entry.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&entry.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m WatchlistModel) Delete(userID, entryID int64) error {
	if entryID < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM watchlist
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, entryID, userID)
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

func (m WatchlistModel) CheckExists(userID, filmID int64) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM watchlist WHERE user_id = $1 AND film_id = $2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	err := m.DB.QueryRowContext(ctx, query, userID, filmID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

var ErrDuplicateWatchlistEntry = errors.New("film already exists in watchlist")
