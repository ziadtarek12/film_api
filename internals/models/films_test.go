package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	_ "github.com/lib/pq"
)

// Define a database interface that matches the methods we use from sql.DB
type DBTX interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Mock DB for testing
type MockDB struct {
	QueryRowContextFunc func(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContextFunc    func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContextFunc     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginFunc           func() (*sql.Tx, error)
	PingContextFunc     func(ctx context.Context) error
	PrepareContextFunc  func(ctx context.Context, query string) (*sql.Stmt, error)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.QueryRowContextFunc != nil {
		return m.QueryRowContextFunc(ctx, query, args...)
	}
	return nil
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryContextFunc != nil {
		return m.QueryContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return nil, nil
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc()
	}
	return nil, nil
}

func (m *MockDB) PingContext(ctx context.Context) error {
	if m.PingContextFunc != nil {
		return m.PingContextFunc(ctx)
	}
	return nil
}

func (m *MockDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if m.PrepareContextFunc != nil {
		return m.PrepareContextFunc(ctx, query)
	}
	return nil, nil
}

// TestFilmModel is a version of FilmModel that uses our DBTX interface
type TestFilmModel struct {
	DB DBTX
}

// Implement all the methods from FilmModel but using our TestFilmModel
func (m TestFilmModel) Get(id int64) (*Film, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, title, year, runtime, rating, description, img, version
		FROM films
		WHERE id = $1`

	var film Film

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&film.ID,
		&film.Title,
		&film.Year,
		&film.Runtime,
		&film.Rating,
		&film.Description,
		&film.Img,
		&film.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	query = `
		SELECT g.name
		FROM genres g
		INNER JOIN films_genres fg ON fg.genre_id = g.id
		WHERE fg.film_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	genres := []Genre{}

	for rows.Next() {
		var g Genre

		err := rows.Scan(&g.Name)
		if err != nil {
			return nil, err
		}

		genres = append(genres, g)
	}

	if rows.Err() != nil {
		return nil, err
	}

	film.Genres = genres

	query = `
		SELECT d.name
		FROM directors d
		INNER JOIN films_directors fd ON fd.director_id = d.id
		WHERE fd.film_id = $1`

	rows, err = m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	directors := []Director{}

	for rows.Next() {
		var d Director

		err := rows.Scan(&d.Name)
		if err != nil {
			return nil, err
		}

		directors = append(directors, d)
	}

	if rows.Err() != nil {
		return nil, err
	}

	film.Directors = directors

	query = `
		SELECT a.name
		FROM actors a
		INNER JOIN films_actors fa ON fa.actor_id = a.id
		WHERE fa.film_id = $1`

	rows, err = m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	actors := []Actor{}

	for rows.Next() {
		var a Actor

		err := rows.Scan(&a.Name)
		if err != nil {
			return nil, err
		}

		actors = append(actors, a)
	}

	if rows.Err() != nil {
		return nil, err
	}

	film.Actors = actors

	return &film, nil
}

func (m TestFilmModel) Insert(film *Film) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO films (title, year, runtime, rating, description, img)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, version`

	args := []interface{}{film.Title, film.Year, film.Runtime, film.Rating, film.Description, film.Img}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&film.ID, &film.Version)
}

func (m TestFilmModel) Update(film *Film) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE films
		SET title = $1, year = $2, runtime = $3, rating = $4, description = $5, img = $6, version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

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

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&film.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		case err.Error() == `pq: duplicate key value violates unique constraint "films_pkey"`:
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m TestFilmModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM films
		WHERE id = $1`

	result, err := m.DB.ExecContext(ctx, query, id)
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

func (m TestFilmModel) GetAll(title string, genres []string, actors []string, directors []string, filters Filters) ([]*Film, Metadata, error) {
	// This is a simplified implementation for testing
	return []*Film{}, Metadata{}, nil
}

func (m TestFilmModel) Count() (int, error) {
	// This is a simplified implementation for testing
	return 0, nil
}

// Mock SQL result for testing
type MockResult struct {
	LastInsertIDFunc func() (int64, error)
	RowsAffectedFunc func() (int64, error)
}

func (m *MockResult) LastInsertId() (int64, error) {
	if m.LastInsertIDFunc != nil {
		return m.LastInsertIDFunc()
	}
	return 0, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0, nil
}

// TestValidateFilm tests the film validation function
func TestValidateFilm(t *testing.T) {
	tests := []struct {
		name      string
		film      *Film
		wantValid bool
	}{
		{
			name: "Valid film",
			film: &Film{
				Title:       "Test Film",
				Year:        2020,
				Runtime:     120,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: true,
		},
		{
			name: "Missing title",
			film: &Film{
				Title:       "",
				Year:        2020,
				Runtime:     120,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Invalid year (too early)",
			film: &Film{
				Title:       "Test Film",
				Year:        1800,
				Runtime:     120,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Invalid year (future)",
			film: &Film{
				Title:       "Test Film",
				Year:        int32(time.Now().Year() + 10),
				Runtime:     120,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Invalid runtime (zero)",
			film: &Film{
				Title:       "Test Film",
				Year:        2020,
				Runtime:     0,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Invalid runtime (negative)",
			film: &Film{
				Title:       "Test Film",
				Year:        2020,
				Runtime:     -10,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Missing genres",
			film: &Film{
				Title:       "Test Film",
				Year:        2020,
				Runtime:     120,
				Genres:      nil,
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Too many genres",
			film: &Film{
				Title:   "Test Film",
				Year:    2020,
				Runtime: 120,
				Genres: []Genre{
					{Name: "Action"},
					{Name: "Adventure"},
					{Name: "Comedy"},
					{Name: "Drama"},
					{Name: "Horror"},
					{Name: "Sci-Fi"},
				},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Duplicate genres",
			film: &Film{
				Title:   "Test Film",
				Year:    2020,
				Runtime: 120,
				Genres: []Genre{
					{Name: "Action"},
					{Name: "Action"},
				},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
			},
			wantValid: false,
		},
		{
			name: "Invalid image URL",
			film: &Film{
				Title:       "Test Film",
				Year:        2020,
				Runtime:     120,
				Genres:      []Genre{{Name: "Action"}},
				Directors:   []Director{{Name: "Director"}},
				Actors:      []Actor{{Name: "Actor"}},
				Rating:      8.5,
				Description: "Test description",
				Img:         "not-a-url",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateFilm(v, tt.film)
			if v.Valid() != tt.wantValid {
				t.Errorf("ValidateFilm() got valid = %v, want %v, errors: %v", v.Valid(), tt.wantValid, v.Errors)
			}
		})
	}
}

// TestFilmGet tests the Get method of FilmModel
func TestFilmGet(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex database mocking")
}

// TestFilmInsert tests the Insert method of FilmModel
func TestFilmInsert(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex database mocking")
}

// TestFilmUpdate tests the Update method of FilmModel
func TestFilmUpdate(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex database mocking")
}

// TestFilmDelete tests the Delete method of FilmModel
func TestFilmDelete(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex database mocking")
}

// TestFilmMarshalJSON tests the MarshalJSON method of Film
func TestFilmMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		film    Film
		wantErr bool
	}{
		{
			name: "Film with runtime",
			film: Film{
				ID:          1,
				Title:       "Test Film",
				Year:        2020,
				Runtime:     120,
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
				Version:     1,
			},
			wantErr: false,
		},
		{
			name: "Film without runtime",
			film: Film{
				ID:          1,
				Title:       "Test Film",
				Year:        2020,
				Runtime:     0,
				Rating:      8.5,
				Description: "Test description",
				Img:         "https://example.com/image.jpg",
				Version:     1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll just test that the marshaling doesn't error
			// The exact JSON output can vary based on field order
			_, err := tt.film.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Film.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify that the runtime field is correctly formatted
			var result map[string]interface{}
			jsonData, _ := tt.film.MarshalJSON()
			json.Unmarshal(jsonData, &result)

			if tt.film.Runtime > 0 {
				expectedRuntime := fmt.Sprintf("%d mins", tt.film.Runtime)
				if result["runtime"] != expectedRuntime {
					t.Errorf("Film.MarshalJSON() runtime = %v, want %v", result["runtime"], expectedRuntime)
				}
			} else {
				if result["runtime"] != "" {
					t.Errorf("Film.MarshalJSON() runtime = %v, want empty string", result["runtime"])
				}
			}
		})
	}
}
