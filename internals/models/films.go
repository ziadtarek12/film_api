package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

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
	Driver neo4j.DriverWithContext
}

func NewFilmModel(driver neo4j.DriverWithContext) FilmModel {
	return FilmModel{
		Driver: driver,
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
	if id < 0 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (f:Film)
			WHERE id(f) = $id
			OPTIONAL MATCH (f)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (f)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (f)-[:HAS_DIRECTOR]->(d:Director)
			RETURN f, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors
		`
		result, err := tx.Run(ctx, query, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		return record, nil
	})

	if err != nil {
		if err.Error() == "Result contains no more records" {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	record := result.(*neo4j.Record)
	node, _ := record.Get("f")
	filmNode := node.(neo4j.Node)

	film := &Film{
		ID:                  filmNode.Id,
		IMDbID:              getStringProp(filmNode, "imdb_id"),
		Title:               getStringProp(filmNode, "title"),
		OriginalTitle:       getStringProp(filmNode, "original_title"),
		Year:                int32(getIntProp(filmNode, "year")),
		ReleaseDate:         getStringProp(filmNode, "release_date"),
		Runtime:             Runtime(getIntProp(filmNode, "runtime")),
		RuntimeSeconds:      int32(getIntProp(filmNode, "runtime_seconds")),
		Rating:              float32(getFloatProp(filmNode, "rating")),
		VoteCount:           float32(getFloatProp(filmNode, "vote_count")),
		Description:         getStringProp(filmNode, "description"),
		PlotSummary:         getStringProp(filmNode, "plot_summary"),
		Certificate:         getStringProp(filmNode, "certificate"),
		ProductionStatus:    getStringProp(filmNode, "production_status"),
		MetacriticScore:     int32(getIntProp(filmNode, "metacritic_score")),
		TrailerID:           getStringProp(filmNode, "trailer_id"),
		WatchCategories:     getStringProp(filmNode, "watch_categories"),
		WatchProviders:      getStringProp(filmNode, "watch_providers"),
		PrimaryImageURL:     getStringProp(filmNode, "primary_image_url"),
		PrimaryImageCaption: getStringProp(filmNode, "primary_image_caption"),
		Img:                 getStringProp(filmNode, "image"),
		Version:             int32(getIntProp(filmNode, "version")),
	}

	genresRaw, _ := record.Get("genres")
	for _, g := range genresRaw.([]any) {
		film.Genres = append(film.Genres, Genre{Name: g.(string)})
	}

	actorsRaw, _ := record.Get("actors")
	for _, a := range actorsRaw.([]any) {
		film.Actors = append(film.Actors, Actor{Name: a.(string)})
	}

	directorsRaw, _ := record.Get("directors")
	for _, d := range directorsRaw.([]any) {
		film.Directors = append(film.Directors, Director{Name: d.(string)})
	}

	return film, nil
}

func (model FilmModel) Insert(film *Film) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			CREATE (f:Film {
				imdb_id: $imdb_id,
				title: $title,
				original_title: $original_title,
				year: $year,
				release_date: $release_date,
				runtime: $runtime,
				runtime_seconds: $runtime_seconds,
				rating: $rating,
				vote_count: $vote_count,
				description: $description,
				plot_summary: $plot_summary,
				certificate: $certificate,
				production_status: $production_status,
				metacritic_score: $metacritic_score,
				trailer_id: $trailer_id,
				watch_categories: $watch_categories,
				watch_providers: $watch_providers,
				primary_image_url: $primary_image_url,
				primary_image_caption: $primary_image_caption,
				image: $image,
				version: 1
			})
			WITH f
			FOREACH (genre_name IN $genres |
				MERGE (g:Genre {name: genre_name})
				MERGE (f)-[:HAS_GENRE]->(g)
			)
			WITH f
			FOREACH (actor_name IN $actors |
				MERGE (a:Actor {name: actor_name})
				MERGE (f)-[:HAS_ACTOR]->(a)
			)
			WITH f
			FOREACH (director_name IN $directors |
				MERGE (d:Director {name: director_name})
				MERGE (f)-[:HAS_DIRECTOR]->(d)
			)
			RETURN id(f)
		`

		genres := make([]string, len(film.Genres))
		for i, g := range film.Genres {
			genres[i] = g.Name
		}

		actors := make([]string, len(film.Actors))
		for i, a := range film.Actors {
			actors[i] = a.Name
		}

		directors := make([]string, len(film.Directors))
		for i, d := range film.Directors {
			directors[i] = d.Name
		}

		params := map[string]any{
			"imdb_id":               film.IMDbID,
			"title":                 film.Title,
			"original_title":        film.OriginalTitle,
			"year":                  film.Year,
			"release_date":          film.ReleaseDate,
			"runtime":               int(film.Runtime),
			"runtime_seconds":       film.RuntimeSeconds,
			"rating":                film.Rating,
			"vote_count":            film.VoteCount,
			"description":           film.Description,
			"plot_summary":          film.PlotSummary,
			"certificate":           film.Certificate,
			"production_status":     film.ProductionStatus,
			"metacritic_score":      film.MetacriticScore,
			"trailer_id":            film.TrailerID,
			"watch_categories":      film.WatchCategories,
			"watch_providers":       film.WatchProviders,
			"primary_image_url":     film.PrimaryImageURL,
			"primary_image_caption": film.PrimaryImageCaption,
			"image":                 film.Img,
			"genres":                genres,
			"actors":                actors,
			"directors":             directors,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		record, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}

		film.ID = record.Values[0].(int64)
		return nil, nil
	})

	return err
}

func (model FilmModel) InsertBulk(films []*Film) error {
	batchSize := 1000

	for i := 0; i < len(films); i += batchSize {
		end := i + batchSize
		if end > len(films) {
			end = len(films)
		}

		batchFilms := films[i:end]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			query := `
				UNWIND $batch AS row
				CREATE (f:Film {
					imdb_id: row.imdb_id,
					title: row.title,
					original_title: row.original_title,
					year: row.year,
					release_date: row.release_date,
					runtime: row.runtime,
					runtime_seconds: row.runtime_seconds,
					rating: row.rating,
					vote_count: row.vote_count,
					description: row.description,
					plot_summary: row.plot_summary,
					certificate: row.certificate,
					production_status: row.production_status,
					metacritic_score: row.metacritic_score,
					trailer_id: row.trailer_id,
					watch_categories: row.watch_categories,
					watch_providers: row.watch_providers,
					primary_image_url: row.primary_image_url,
					primary_image_caption: row.primary_image_caption,
					image: row.image,
					version: 1
				})
				WITH f, row
				FOREACH (genre_name IN row.genres |
					MERGE (g:Genre {name: genre_name})
					MERGE (f)-[:HAS_GENRE]->(g)
				)
				FOREACH (actor_name IN row.actors |
					MERGE (a:Actor {name: actor_name})
					MERGE (f)-[:HAS_ACTOR]->(a)
				)
				FOREACH (director_name IN row.directors |
					MERGE (d:Director {name: director_name})
					MERGE (f)-[:HAS_DIRECTOR]->(d)
				)
				RETURN id(f)
			`

			batch := make([]map[string]any, len(batchFilms))
			for k, film := range batchFilms {
				genres := make([]string, len(film.Genres))
				for j, g := range film.Genres {
					genres[j] = g.Name
				}

				actors := make([]string, len(film.Actors))
				for j, a := range film.Actors {
					actors[j] = a.Name
				}

				directors := make([]string, len(film.Directors))
				for j, d := range film.Directors {
					directors[j] = d.Name
				}

				batch[k] = map[string]any{
					"imdb_id":               film.IMDbID,
					"title":                 film.Title,
					"original_title":        film.OriginalTitle,
					"year":                  film.Year,
					"release_date":          film.ReleaseDate,
					"runtime":               int(film.Runtime),
					"runtime_seconds":       film.RuntimeSeconds,
					"rating":                film.Rating,
					"vote_count":            film.VoteCount,
					"description":           film.Description,
					"plot_summary":          film.PlotSummary,
					"certificate":           film.Certificate,
					"production_status":     film.ProductionStatus,
					"metacritic_score":      film.MetacriticScore,
					"trailer_id":            film.TrailerID,
					"watch_categories":      film.WatchCategories,
					"watch_providers":       film.WatchProviders,
					"primary_image_url":     film.PrimaryImageURL,
					"primary_image_caption": film.PrimaryImageCaption,
					"image":                 film.Img,
					"genres":                genres,
					"actors":                actors,
					"directors":             directors,
				}
			}

			result, err := tx.Run(ctx, query, map[string]any{"batch": batch})
			if err != nil {
				return nil, err
			}

			var k int
			for result.Next(ctx) {
				if k < len(batchFilms) {
					batchFilms[k].ID = result.Record().Values[0].(int64)
					k++
				}
			}

			return nil, result.Err()
		})

		session.Close(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (model FilmModel) Update(film *Film) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Check version and update
		query := `
			MATCH (f:Film)
			WHERE id(f) = $id AND f.version = $version
			SET f.imdb_id = $imdb_id,
				f.title = $title,
				f.original_title = $original_title,
				f.year = $year,
				f.release_date = $release_date,
				f.runtime = $runtime,
				f.runtime_seconds = $runtime_seconds,
				f.rating = $rating,
				f.vote_count = $vote_count,
				f.description = $description,
				f.plot_summary = $plot_summary,
				f.certificate = $certificate,
				f.production_status = $production_status,
				f.metacritic_score = $metacritic_score,
				f.trailer_id = $trailer_id,
				f.watch_categories = $watch_categories,
				f.watch_providers = $watch_providers,
				f.primary_image_url = $primary_image_url,
				f.primary_image_caption = $primary_image_caption,
				f.image = $image,
				f.version = f.version + 1
			
			// Clear existing relationships
			WITH f
			OPTIONAL MATCH (f)-[r1:HAS_GENRE]->() DELETE r1
			WITH f
			OPTIONAL MATCH (f)-[r2:HAS_ACTOR]->() DELETE r2
			WITH f
			OPTIONAL MATCH (f)-[r3:HAS_DIRECTOR]->() DELETE r3

			// Re-create relationships
			WITH f
			FOREACH (genre_name IN $genres |
				MERGE (g:Genre {name: genre_name})
				MERGE (f)-[:HAS_GENRE]->(g)
			)
			WITH f
			FOREACH (actor_name IN $actors |
				MERGE (a:Actor {name: actor_name})
				MERGE (f)-[:HAS_ACTOR]->(a)
			)
			WITH f
			FOREACH (director_name IN $directors |
				MERGE (d:Director {name: director_name})
				MERGE (f)-[:HAS_DIRECTOR]->(d)
			)
			
			RETURN f.version
		`

		genres := make([]string, len(film.Genres))
		for i, g := range film.Genres {
			genres[i] = g.Name
		}

		actors := make([]string, len(film.Actors))
		for i, a := range film.Actors {
			actors[i] = a.Name
		}

		directors := make([]string, len(film.Directors))
		for i, d := range film.Directors {
			directors[i] = d.Name
		}

		params := map[string]any{
			"id":                    film.ID,
			"version":               film.Version,
			"imdb_id":               film.IMDbID,
			"title":                 film.Title,
			"original_title":        film.OriginalTitle,
			"year":                  film.Year,
			"release_date":          film.ReleaseDate,
			"runtime":               int(film.Runtime),
			"runtime_seconds":       film.RuntimeSeconds,
			"rating":                film.Rating,
			"vote_count":            film.VoteCount,
			"description":           film.Description,
			"plot_summary":          film.PlotSummary,
			"certificate":           film.Certificate,
			"production_status":     film.ProductionStatus,
			"metacritic_score":      film.MetacriticScore,
			"trailer_id":            film.TrailerID,
			"watch_categories":      film.WatchCategories,
			"watch_providers":       film.WatchProviders,
			"primary_image_url":     film.PrimaryImageURL,
			"primary_image_caption": film.PrimaryImageCaption,
			"image":                 film.Img,
			"genres":                genres,
			"actors":                actors,
			"directors":             directors,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			film.Version = int32(result.Record().Values[0].(int64))
			return nil, nil
		}

		return nil, ErrEditConflict
	})

	return err
}

func (model FilmModel) Delete(id int64) error {
	if id < 0 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (f:Film)
			WHERE id(f) = $id
			DETACH DELETE f
			RETURN count(f)
		`
		result, err := tx.Run(ctx, query, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			count := result.Record().Values[0].(int64)
			if count == 0 {
				return nil, ErrRecordNotFound
			}
			return nil, nil
		}

		return nil, ErrRecordNotFound
	})

	return err
}

func (model FilmModel) GetAll(title string, genres []string, actors []string, directors []string, filters Filters) ([]*Film, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Build dynamic query
		whereClause := "WHERE ($title = '' OR toLower(f.title) CONTAINS toLower($title))"

		if len(genres) > 0 {
			whereClause += " AND ANY(g IN genres WHERE g IN $genres)"
		}
		if len(actors) > 0 {
			whereClause += " AND ANY(a IN actors WHERE a IN $actors)"
		}
		if len(directors) > 0 {
			whereClause += " AND ANY(d IN directors WHERE d IN $directors)"
		}

		// Count query
		countQuery := `
			MATCH (f:Film)
			OPTIONAL MATCH (f)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (f)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (f)-[:HAS_DIRECTOR]->(d:Director)
			WITH f, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors
			` + whereClause + `
			RETURN count(f)
		`

		params := map[string]any{
			"title":     title,
			"genres":    genres,
			"actors":    actors,
			"directors": directors,
			"limit":     filters.limit(),
			"offset":    filters.offset(),
		}

		countResult, err := tx.Run(ctx, countQuery, params)
		if err != nil {
			return nil, err
		}

		totalRecords := 0
		if countResult.Next(ctx) {
			totalRecords = int(countResult.Record().Values[0].(int64))
		}

		// Prepare sort string
		sortStr := filters.sortColumn()
		if sortStr == "" {
			sortStr = "id ASC"
		}
		sortStr = strings.TrimSuffix(sortStr, ",")
		sortStr = strings.ReplaceAll(sortStr, ",", ", f.")
		sortStr = "f." + sortStr

		// Data query
		dataQuery := `
			MATCH (f:Film)
			OPTIONAL MATCH (f)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (f)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (f)-[:HAS_DIRECTOR]->(d:Director)
			WITH f, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors
			` + whereClause + `
			RETURN f, genres, actors, directors
			ORDER BY ` + sortStr + `
			SKIP $offset
			LIMIT $limit
		`

		dataResult, err := tx.Run(ctx, dataQuery, params)
		if err != nil {
			return nil, err
		}

		films := []*Film{}
		for dataResult.Next(ctx) {
			record := dataResult.Record()
			node, _ := record.Get("f")
			filmNode := node.(neo4j.Node)

			film := &Film{
				ID:                  filmNode.Id,
				IMDbID:              getStringProp(filmNode, "imdb_id"),
				Title:               getStringProp(filmNode, "title"),
				OriginalTitle:       getStringProp(filmNode, "original_title"),
				Year:                int32(getIntProp(filmNode, "year")),
				ReleaseDate:         getStringProp(filmNode, "release_date"),
				Runtime:             Runtime(getIntProp(filmNode, "runtime")),
				RuntimeSeconds:      int32(getIntProp(filmNode, "runtime_seconds")),
				Rating:              float32(getFloatProp(filmNode, "rating")),
				VoteCount:           float32(getFloatProp(filmNode, "vote_count")),
				Description:         getStringProp(filmNode, "description"),
				PlotSummary:         getStringProp(filmNode, "plot_summary"),
				Certificate:         getStringProp(filmNode, "certificate"),
				ProductionStatus:    getStringProp(filmNode, "production_status"),
				MetacriticScore:     int32(getIntProp(filmNode, "metacritic_score")),
				TrailerID:           getStringProp(filmNode, "trailer_id"),
				WatchCategories:     getStringProp(filmNode, "watch_categories"),
				WatchProviders:      getStringProp(filmNode, "watch_providers"),
				PrimaryImageURL:     getStringProp(filmNode, "primary_image_url"),
				PrimaryImageCaption: getStringProp(filmNode, "primary_image_caption"),
				Img:                 getStringProp(filmNode, "image"),
				Version:             int32(getIntProp(filmNode, "version")),
			}

			genresRaw, _ := record.Get("genres")
			for _, g := range genresRaw.([]any) {
				film.Genres = append(film.Genres, Genre{Name: g.(string)})
			}

			actorsRaw, _ := record.Get("actors")
			for _, a := range actorsRaw.([]any) {
				film.Actors = append(film.Actors, Actor{Name: a.(string)})
			}

			directorsRaw, _ := record.Get("directors")
			for _, d := range directorsRaw.([]any) {
				film.Directors = append(film.Directors, Director{Name: d.(string)})
			}

			films = append(films, film)
		}

		metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
		return struct {
			Films    []*Film
			Metadata Metadata
		}{films, metadata}, nil
	})

	if err != nil {
		return nil, Metadata{}, err
	}

	res := result.(struct {
		Films    []*Film
		Metadata Metadata
	})

	return res.Films, res.Metadata, nil
}

func (m *FilmModel) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (f:Film) RETURN count(f)`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return 0, err
		}

		if result.Next(ctx) {
			return int(result.Record().Values[0].(int64)), nil
		}
		return 0, nil
	})

	if err != nil {
		return 0, err
	}

	return result.(int), nil
}

func (model FilmModel) GetRecommendations(userID int64, limit int) ([]*Film, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := model.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)-[:FOR_FILM]->(source:Film)
			WHERE id(u) = $user_id
			
			// Determine multiplier based on user preference (Rating > Priority > Default)
			WITH u, source, 
				CASE 
					WHEN w.rating IS NOT NULL THEN toFloat(w.rating)
					WHEN w.priority IS NOT NULL THEN toFloat(w.priority)
					ELSE 5.0
				END AS multiplier

			// Find films sharing relationship attributes (Genre/Actor/Director)
			MATCH (source)-->(shared)<--(rec:Film)
			WHERE NOT (u)-[:HAS_WATCHLIST]->(:Watchlist)-[:FOR_FILM]->(rec) AND rec <> source
			
			// Calculate relationship-based score
			WITH u, source, rec, multiplier,
				sum(
					CASE 
						WHEN 'Director' IN labels(shared) THEN 5 
						WHEN 'Actor' IN labels(shared) THEN 3 
						WHEN 'Genre' IN labels(shared) THEN 1 
						ELSE 0 
					END
				) AS rel_score
			
			// Add additional attribute matching scores
			WITH rec, multiplier, rel_score,
				CASE WHEN source.certificate = rec.certificate THEN 2 ELSE 0 END AS cert_score,
				CASE WHEN abs(source.year - rec.year) <= 5 THEN 2 ELSE 0 END AS year_score,
				CASE WHEN abs(source.runtime - rec.runtime) <= 20 THEN 1 ELSE 0 END AS runtime_score
			
			// Calculate total weighted score
			WITH rec, (rel_score + cert_score + year_score + runtime_score) * multiplier AS score
			ORDER BY score DESC
			LIMIT $limit

			// Fetch details for result
			OPTIONAL MATCH (rec)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (rec)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (rec)-[:HAS_DIRECTOR]->(d:Director)
			RETURN rec, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors, score
		`

		params := map[string]any{
			"user_id": userID,
			"limit":   limit,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		films := []*Film{}
		for result.Next(ctx) {
			record := result.Record()
			node := record.Values[0].(neo4j.Node)
			filmNode := node

			film := &Film{
				ID:                  filmNode.Id,
				IMDbID:              getStringProp(filmNode, "imdb_id"),
				Title:               getStringProp(filmNode, "title"),
				OriginalTitle:       getStringProp(filmNode, "original_title"),
				Year:                int32(getIntProp(filmNode, "year")),
				ReleaseDate:         getStringProp(filmNode, "release_date"),
				Runtime:             Runtime(getIntProp(filmNode, "runtime")),
				RuntimeSeconds:      int32(getIntProp(filmNode, "runtime_seconds")),
				Rating:              float32(getFloatProp(filmNode, "rating")),
				VoteCount:           float32(getFloatProp(filmNode, "vote_count")),
				Description:         getStringProp(filmNode, "description"),
				PlotSummary:         getStringProp(filmNode, "plot_summary"),
				Certificate:         getStringProp(filmNode, "certificate"),
				ProductionStatus:    getStringProp(filmNode, "production_status"),
				MetacriticScore:     int32(getIntProp(filmNode, "metacritic_score")),
				TrailerID:           getStringProp(filmNode, "trailer_id"),
				WatchCategories:     getStringProp(filmNode, "watch_categories"),
				WatchProviders:      getStringProp(filmNode, "watch_providers"),
				PrimaryImageURL:     getStringProp(filmNode, "primary_image_url"),
				PrimaryImageCaption: getStringProp(filmNode, "primary_image_caption"),
				Img:                 getStringProp(filmNode, "image"),
				Version:             int32(getIntProp(filmNode, "version")),
			}

			genresRaw := record.Values[1]
			for _, g := range genresRaw.([]any) {
				film.Genres = append(film.Genres, Genre{Name: g.(string)})
			}

			actorsRaw := record.Values[2]
			for _, a := range actorsRaw.([]any) {
				film.Actors = append(film.Actors, Actor{Name: a.(string)})
			}

			directorsRaw := record.Values[3]
			for _, d := range directorsRaw.([]any) {
				film.Directors = append(film.Directors, Director{Name: d.(string)})
			}

			films = append(films, film)
		}

		return films, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*Film), nil
}

// Helper functions to safely get properties
func getStringProp(node neo4j.Node, key string) string {
	if val, ok := node.Props[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}

func getIntProp(node neo4j.Node, key string) int64 {
	if val, ok := node.Props[key]; ok && val != nil {
		return val.(int64)
	}
	return 0
}

func getFloatProp(node neo4j.Node, key string) float64 {
	if val, ok := node.Props[key]; ok && val != nil {
		return val.(float64)
	}
	return 0.0
}
