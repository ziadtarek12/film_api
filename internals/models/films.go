package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
)

// StringArray represents a slice of strings that can be stored as JSON in SQLite
type StringArray []string

// Scan implements the Scanner interface for database/sql
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*sa = StringArray{}
			return nil
		}
		return json.Unmarshal([]byte(v), sa)
	case []byte:
		if len(v) == 0 {
			*sa = StringArray{}
			return nil
		}
		return json.Unmarshal(v, sa)
	}

	return fmt.Errorf("cannot scan %T into StringArray", value)
}

// Value implements the Valuer interface for database/sql
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "[]", nil
	}
	return json.Marshal(sa)
}

type Film struct {
	ID                  int64      `json:"id"`
	IMDbID              string     `json:"imdb_id"`
	Title               string     `json:"title"`
	OriginalTitle       string     `json:"original_title"`
	Year                int32      `json:"year"`
	ReleaseDate         string     `json:"release_date"`
	Runtime             Runtime    `json:"runtime"`
	RuntimeSeconds      int32      `json:"runtime_seconds"`
	Genres              []Genre    `json:"genres"`
	Directors           []Director `json:"directors"`
	Actors              []Actor    `json:"actors"`
	Rating              float32    `json:"rating"`
	VoteCount           float32    `json:"vote_count"`
	Description         string     `json:"description"`
	PlotSummary         string     `json:"plot_summary"`
	Certificate         string     `json:"certificate"`
	ProductionStatus    string     `json:"production_status"`
	MetacriticScore     int32      `json:"metacritic_score"`
	TrailerID           string     `json:"trailer_id"`
	WatchCategories     string     `json:"watch_categories"`
	WatchProviders      string     `json:"watch_providers"`
	PrimaryImageURL     string     `json:"primary_image_url"`
	PrimaryImageCaption string     `json:"primary_image_caption"`
	Img                 string     `json:"image"` // Keep for backward compatibility
	Version             int32      `json:"version"`
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
	v.Check(film.Year <= int32(time.Now().Year()+10), "year", "must not be more than 10 years in the future")
	v.Check(film.Runtime != 0, "runtime", "must be provided")
	v.Check(film.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(film.Genres != nil, "genres", "must be provided")
	v.Check(len(film.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(film.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(film.Genres), "genres", "must not contain duplicate values")
	if film.Img != "" {
		v.Check(validator.MatchesURL(film.Img), "image", "Must be an URL")
	}
	if film.PrimaryImageURL != "" {
		v.Check(validator.MatchesURL(film.PrimaryImageURL), "primary_image_url", "Must be an URL")
	}
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
		f.id, f.imdb_id, f.title, f.original_title, f.year, f.release_date, f.runtime, 
		f.runtime_seconds, f.rating, f.vote_count, f.description, f.plot_summary, 
		f.certificate, f.production_status, f.metacritic_score, f.trailer_id, 
		f.watch_categories, f.watch_providers, f.primary_image_url, f.primary_image_caption, 
		f.image, f.version,
		(SELECT json_group_array(g.name) FROM film_genres fg JOIN genres g ON fg.genre_id = g.id WHERE fg.film_id = f.id) AS genres,
		(SELECT json_group_array(a.name) FROM film_actors fa JOIN actors a ON fa.actor_id = a.id WHERE fa.film_id = f.id) AS actors,
		(SELECT json_group_array(d.name) FROM film_directors fd JOIN directors d ON fd.director_id = d.id WHERE fd.film_id = f.id) AS directors
		FROM films f
		WHERE f.id = ?
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var film Film
	var genres StringArray
	var actors StringArray
	var directors StringArray

	err := model.DB.QueryRowContext(ctx, query, id).Scan(
		&film.ID,
		&film.IMDbID,
		&film.Title,
		&film.OriginalTitle,
		&film.Year,
		&film.ReleaseDate,
		&film.Runtime,
		&film.RuntimeSeconds,
		&film.Rating,
		&film.VoteCount,
		&film.Description,
		&film.PlotSummary,
		&film.Certificate,
		&film.ProductionStatus,
		&film.MetacriticScore,
		&film.TrailerID,
		&film.WatchCategories,
		&film.WatchProviders,
		&film.PrimaryImageURL,
		&film.PrimaryImageCaption,
		&film.Img,
		&film.Version,
		&genres,
		&actors,
		&directors,
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
	query := `INSERT INTO films (imdb_id, title, original_title, year, release_date, runtime, 
		runtime_seconds, rating, vote_count, description, plot_summary, certificate, 
		production_status, metacritic_score, trailer_id, watch_categories, watch_providers, 
		primary_image_url, primary_image_caption, image, version) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.ExecContext(ctx, query,
		film.IMDbID, film.Title, film.OriginalTitle, film.Year, film.ReleaseDate,
		film.Runtime, film.RuntimeSeconds, film.Rating, film.VoteCount, film.Description,
		film.PlotSummary, film.Certificate, film.ProductionStatus, film.MetacriticScore,
		film.TrailerID, film.WatchCategories, film.WatchProviders, film.PrimaryImageURL,
		film.PrimaryImageCaption, film.Img, 1)
	if err != nil {
		return err
	}

	filmID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	film.ID = filmID

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
		SET imdb_id = ?, title = ?, original_title = ?, year = ?, release_date = ?, 
		runtime = ?, runtime_seconds = ?, rating = ?, vote_count = ?, description = ?, 
		plot_summary = ?, certificate = ?, production_status = ?, metacritic_score = ?, 
		trailer_id = ?, watch_categories = ?, watch_providers = ?, primary_image_url = ?, 
		primary_image_caption = ?, image = ?, version = version + 1
		WHERE id = ? AND version = ?
	`

	result, err := tx.ExecContext(ctx, query,
		film.IMDbID, film.Title, film.OriginalTitle, film.Year, film.ReleaseDate,
		film.Runtime, film.RuntimeSeconds, film.Rating, film.VoteCount, film.Description,
		film.PlotSummary, film.Certificate, film.ProductionStatus, film.MetacriticScore,
		film.TrailerID, film.WatchCategories, film.WatchProviders, film.PrimaryImageURL,
		film.PrimaryImageCaption, film.Img, film.ID, film.Version)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrEditConflict
	}

	film.Version++

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
		DELETE FROM films WHERE id = ?
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

func (model FilmModel) GetAll(title string, genres []string, actors []string, directors []string, filters Filters) ([]*Film, Metadata, error) {
	// First get the total count
	countQuery := `
		SELECT COUNT(DISTINCT f.id)
		FROM films f
		LEFT JOIN film_genres fg ON f.id = fg.film_id
		LEFT JOIN genres g ON fg.genre_id = g.id
		LEFT JOIN film_actors fa ON f.id = fa.film_id
		LEFT JOIN actors a ON fa.actor_id = a.id
		LEFT JOIN film_directors fd ON f.id = fd.film_id
		LEFT JOIN directors d ON fd.director_id = d.id
		WHERE (? = '' OR f.title LIKE '%' || ? || '%')
	`

	args := []interface{}{title, title}

	if len(genres) > 0 {
		placeholders := make([]string, len(genres))
		for i, genre := range genres {
			placeholders[i] = "?"
			args = append(args, genre)
		}
		countQuery += fmt.Sprintf(" AND g.name IN (%s)", strings.Join(placeholders, ","))
	}

	if len(actors) > 0 {
		placeholders := make([]string, len(actors))
		for i, actor := range actors {
			placeholders[i] = "?"
			args = append(args, actor)
		}
		countQuery += fmt.Sprintf(" AND a.name IN (%s)", strings.Join(placeholders, ","))
	}

	if len(directors) > 0 {
		placeholders := make([]string, len(directors))
		for i, director := range directors {
			placeholders[i] = "?"
			args = append(args, director)
		}
		countQuery += fmt.Sprintf(" AND d.name IN (%s)", strings.Join(placeholders, ","))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var totalRecords int
	err := model.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalRecords)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Now get the actual films
	query := fmt.Sprintf(`
		SELECT DISTINCT f.id, f.imdb_id, f.title, f.original_title, f.year, f.release_date, 
		f.runtime, f.runtime_seconds, f.rating, f.vote_count, f.description, f.plot_summary, 
		f.certificate, f.production_status, f.metacritic_score, f.trailer_id, 
		f.watch_categories, f.watch_providers, f.primary_image_url, f.primary_image_caption, 
		f.image, f.version,
		(SELECT json_group_array(g2.name) FROM film_genres fg2 JOIN genres g2 ON fg2.genre_id = g2.id WHERE fg2.film_id = f.id) AS genres,
		(SELECT json_group_array(a2.name) FROM film_actors fa2 JOIN actors a2 ON fa2.actor_id = a2.id WHERE fa2.film_id = f.id) AS actors,
		(SELECT json_group_array(d2.name) FROM film_directors fd2 JOIN directors d2 ON fd2.director_id = d2.id WHERE fd2.film_id = f.id) AS directors
		FROM films f
		LEFT JOIN film_genres fg ON f.id = fg.film_id
		LEFT JOIN genres g ON fg.genre_id = g.id
		LEFT JOIN film_actors fa ON f.id = fa.film_id
		LEFT JOIN actors a ON fa.actor_id = a.id
		LEFT JOIN film_directors fd ON f.id = fd.film_id
		LEFT JOIN directors d ON fd.director_id = d.id
		WHERE (? = '' OR f.title LIKE '%%' || ? || '%%')
	`)

	args = []interface{}{title, title}

	if len(genres) > 0 {
		placeholders := make([]string, len(genres))
		for i, genre := range genres {
			placeholders[i] = "?"
			args = append(args, genre)
		}
		query += fmt.Sprintf(" AND g.name IN (%s)", strings.Join(placeholders, ","))
	}

	if len(actors) > 0 {
		placeholders := make([]string, len(actors))
		for i, actor := range actors {
			placeholders[i] = "?"
			args = append(args, actor)
		}
		query += fmt.Sprintf(" AND a.name IN (%s)", strings.Join(placeholders, ","))
	}

	if len(directors) > 0 {
		placeholders := make([]string, len(directors))
		for i, director := range directors {
			placeholders[i] = "?"
			args = append(args, director)
		}
		query += fmt.Sprintf(" AND d.name IN (%s)", strings.Join(placeholders, ","))
	}

	query += fmt.Sprintf(" ORDER BY %s f.id ASC LIMIT ? OFFSET ?", filters.sortColumn())
	args = append(args, filters.limit(), filters.offset())

	rows, err := model.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	films := []*Film{}
	for rows.Next() {
		var film Film
		var genres StringArray
		var actors StringArray
		var directors StringArray

		err := rows.Scan(
			&film.ID,
			&film.IMDbID,
			&film.Title,
			&film.OriginalTitle,
			&film.Year,
			&film.ReleaseDate,
			&film.Runtime,
			&film.RuntimeSeconds,
			&film.Rating,
			&film.VoteCount,
			&film.Description,
			&film.PlotSummary,
			&film.Certificate,
			&film.ProductionStatus,
			&film.MetacriticScore,
			&film.TrailerID,
			&film.WatchCategories,
			&film.WatchProviders,
			&film.PrimaryImageURL,
			&film.PrimaryImageCaption,
			&film.Img,
			&film.Version,
			&genres,
			&actors,
			&directors,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Convert string arrays to respective types
		film.Genres = make([]Genre, len(genres))
		for i, genre := range genres {
			film.Genres[i] = Genre{Name: genre}
		}

		film.Directors = make([]Director, len(directors))
		for i, director := range directors {
			film.Directors[i] = Director{Name: director}
		}

		film.Actors = make([]Actor, len(actors))
		for i, actor := range actors {
			film.Actors[i] = Actor{Name: actor}
		}

		films = append(films, &film)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return films, metadata, nil
}

func (model FilmModel) batchInsertRelations(tx *sql.Tx, ctx context.Context, film *Film) error {
	// Delete existing relations
	_, err := tx.ExecContext(ctx, "DELETE FROM film_directors WHERE film_id = ?", film.ID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM film_actors WHERE film_id = ?", film.ID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM film_genres WHERE film_id = ?", film.ID)
	if err != nil {
		return err
	}

	// Insert directors
	for _, director := range film.Directors {
		// Insert or get director
		var directorID int64
		err := tx.QueryRowContext(ctx, "INSERT OR IGNORE INTO directors (name) VALUES (?)", director.Name).Scan()
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.QueryRowContext(ctx, "SELECT id FROM directors WHERE name = ?", director.Name).Scan(&directorID)
		if err != nil {
			return err
		}

		// Link film and director
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO film_directors (film_id, director_id) VALUES (?, ?)", film.ID, directorID)
		if err != nil {
			return err
		}
	}

	// Insert actors
	for _, actor := range film.Actors {
		// Insert or get actor
		var actorID int64
		err := tx.QueryRowContext(ctx, "INSERT OR IGNORE INTO actors (name) VALUES (?)", actor.Name).Scan()
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.QueryRowContext(ctx, "SELECT id FROM actors WHERE name = ?", actor.Name).Scan(&actorID)
		if err != nil {
			return err
		}

		// Link film and actor
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO film_actors (film_id, actor_id) VALUES (?, ?)", film.ID, actorID)
		if err != nil {
			return err
		}
	}

	// Insert genres
	for _, genre := range film.Genres {
		// Insert or get genre
		var genreID int64
		err := tx.QueryRowContext(ctx, "INSERT OR IGNORE INTO genres (name) VALUES (?)", genre.Name).Scan()
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.QueryRowContext(ctx, "SELECT id FROM genres WHERE name = ?", genre.Name).Scan(&genreID)
		if err != nil {
			return err
		}

		// Link film and genre
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO film_genres (film_id, genre_id) VALUES (?, ?)", film.ID, genreID)
		if err != nil {
			return err
		}
	}

	return nil
}

// Count returns the number of films in the database.
func (m *FilmModel) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM films` // Adjust the table name as necessary
	err := m.DB.QueryRowContext(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
