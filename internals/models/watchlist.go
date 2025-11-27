package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"filmapi.zeyadtarek.net/internals/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
	Driver neo4j.DriverWithContext
}

func ValidateWatchlistEntry(v *validator.Validator, entry *Watchlist) {
	v.Check(entry.FilmID >= 0, "film_id", "must be provided")
	v.Check(entry.Priority >= 1 && entry.Priority <= 10, "priority", "must be between 1 and 10")
	v.Check(len(entry.Notes) <= 1000, "notes", "must not be more than 1000 characters long")

	if entry.Rating != nil {
		v.Check(*entry.Rating >= 1 && *entry.Rating <= 10, "rating", "must be between 1 and 10")
	}

	if entry.Watched && entry.WatchedAt == nil {
		entry.WatchedAt = &time.Time{}
		*entry.WatchedAt = time.Now().UTC()
	}
}

var ErrDuplicateWatchlistEntry = errors.New("film already exists in watchlist")

func (m WatchlistModel) Insert(entry *Watchlist) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Check for duplicate
		checkQuery := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)-[:FOR_FILM]->(f:Film)
			WHERE id(u) = $user_id AND id(f) = $film_id
			RETURN count(w)
		`
		result, err := tx.Run(ctx, checkQuery, map[string]any{
			"user_id": entry.UserID,
			"film_id": entry.FilmID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			count := result.Record().Values[0].(int64)
			if count > 0 {
				return nil, ErrDuplicateWatchlistEntry
			}
		}

		query := `
			MATCH (u:User), (f:Film)
			WHERE id(u) = $user_id AND id(f) = $film_id
			CREATE (w:Watchlist {
				notes: $notes,
				priority: $priority,
				watched: $watched,
				watched_at: datetime($watched_at),
				rating: $rating,
				added_at: datetime(),
				version: 1
			})
			MERGE (u)-[:HAS_WATCHLIST]->(w)
			MERGE (w)-[:FOR_FILM]->(f)
			RETURN id(w), w.added_at, w.version
		`

		var watchedAt any
		if entry.WatchedAt != nil {
			watchedAt = *entry.WatchedAt
		}

		params := map[string]any{
			"user_id":    entry.UserID,
			"film_id":    entry.FilmID,
			"notes":      entry.Notes,
			"priority":   entry.Priority,
			"watched":    entry.Watched,
			"watched_at": watchedAt,
			"rating":     entry.Rating,
		}

		result, err = tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			entry.ID = record.Values[0].(int64)
			entry.AddedAt = record.Values[1].(time.Time)
			entry.Version = int(record.Values[2].(int64))
			return nil, nil
		}

		return nil, errors.New("failed to insert watchlist entry")
	})

	return err
}

func (m WatchlistModel) Get(userID, entryID int64) (*Watchlist, error) {
	if entryID < 0 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)-[:FOR_FILM]->(f:Film)
			WHERE id(w) = $entry_id AND id(u) = $user_id
			OPTIONAL MATCH (f)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (f)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (f)-[:HAS_DIRECTOR]->(d:Director)
			RETURN w, f, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"entry_id": entryID,
			"user_id":  userID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record(), nil
		}

		return nil, ErrRecordNotFound
	})

	if err != nil {
		return nil, err
	}

	record := result.(*neo4j.Record)
	wNode := record.Values[0].(neo4j.Node)
	fNode := record.Values[1].(neo4j.Node)

	entry := &Watchlist{
		ID:       wNode.Id,
		UserID:   userID,
		FilmID:   fNode.Id,
		AddedAt:  wNode.Props["added_at"].(time.Time),
		Notes:    getStringProp(wNode, "notes"),
		Priority: int(getIntProp(wNode, "priority")),
		Watched:  wNode.Props["watched"].(bool),
		Version:  int(getIntProp(wNode, "version")),
	}

	if val, ok := wNode.Props["watched_at"]; ok && val != nil {
		t := val.(time.Time)
		entry.WatchedAt = &t
	}

	if val, ok := wNode.Props["rating"]; ok && val != nil {
		r := int(val.(int64))
		entry.Rating = &r
	}

	film := &Film{
		ID:                  fNode.Id,
		IMDbID:              getStringProp(fNode, "imdb_id"),
		Title:               getStringProp(fNode, "title"),
		OriginalTitle:       getStringProp(fNode, "original_title"),
		Year:                int32(getIntProp(fNode, "year")),
		ReleaseDate:         getStringProp(fNode, "release_date"),
		Runtime:             Runtime(getIntProp(fNode, "runtime")),
		RuntimeSeconds:      int32(getIntProp(fNode, "runtime_seconds")),
		Rating:              float32(getFloatProp(fNode, "rating")),
		VoteCount:           float32(getFloatProp(fNode, "vote_count")),
		Description:         getStringProp(fNode, "description"),
		PlotSummary:         getStringProp(fNode, "plot_summary"),
		Certificate:         getStringProp(fNode, "certificate"),
		ProductionStatus:    getStringProp(fNode, "production_status"),
		MetacriticScore:     int32(getIntProp(fNode, "metacritic_score")),
		TrailerID:           getStringProp(fNode, "trailer_id"),
		WatchCategories:     getStringProp(fNode, "watch_categories"),
		WatchProviders:      getStringProp(fNode, "watch_providers"),
		PrimaryImageURL:     getStringProp(fNode, "primary_image_url"),
		PrimaryImageCaption: getStringProp(fNode, "primary_image_caption"),
		Img:                 getStringProp(fNode, "image"),
		Version:             int32(getIntProp(fNode, "version")),
	}

	genresRaw := record.Values[2]
	for _, g := range genresRaw.([]any) {
		film.Genres = append(film.Genres, Genre{Name: g.(string)})
	}

	actorsRaw := record.Values[3]
	for _, a := range actorsRaw.([]any) {
		film.Actors = append(film.Actors, Actor{Name: a.(string)})
	}

	directorsRaw := record.Values[4]
	for _, d := range directorsRaw.([]any) {
		film.Directors = append(film.Directors, Director{Name: d.(string)})
	}

	entry.Film = film
	return entry, nil
}

func (m WatchlistModel) GetAll(userID int64, watched *bool, priority int, filters Filters) ([]*Watchlist, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		whereClause := "WHERE id(u) = $user_id"
		if watched != nil {
			whereClause += " AND w.watched = $watched"
		}
		if priority > 0 {
			whereClause += " AND w.priority = $priority"
		}

		countQuery := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)
			` + whereClause + `
			RETURN count(w)
		`

		params := map[string]any{
			"user_id":  userID,
			"watched":  watched,
			"priority": priority,
			"limit":    filters.limit(),
			"offset":   filters.offset(),
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
			sortStr = "added_at DESC"
		}
		sortStr = strings.TrimSuffix(sortStr, ",")
		sortStr = strings.ReplaceAll(sortStr, ",", ", w.")
		sortStr = "w." + sortStr

		dataQuery := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)-[:FOR_FILM]->(f:Film)
			` + whereClause + `
			OPTIONAL MATCH (f)-[:HAS_GENRE]->(g:Genre)
			OPTIONAL MATCH (f)-[:HAS_ACTOR]->(a:Actor)
			OPTIONAL MATCH (f)-[:HAS_DIRECTOR]->(d:Director)
			RETURN w, f, collect(DISTINCT g.name) as genres, collect(DISTINCT a.name) as actors, collect(DISTINCT d.name) as directors
			ORDER BY ` + sortStr + `
			SKIP $offset
			LIMIT $limit
		`

		dataResult, err := tx.Run(ctx, dataQuery, params)
		if err != nil {
			return nil, err
		}

		watchlist := []*Watchlist{}
		for dataResult.Next(ctx) {
			record := dataResult.Record()
			wNode := record.Values[0].(neo4j.Node)
			fNode := record.Values[1].(neo4j.Node)

			entry := &Watchlist{
				ID:       wNode.Id,
				UserID:   userID,
				FilmID:   fNode.Id,
				AddedAt:  wNode.Props["added_at"].(time.Time),
				Notes:    getStringProp(wNode, "notes"),
				Priority: int(getIntProp(wNode, "priority")),
				Watched:  wNode.Props["watched"].(bool),
				Version:  int(getIntProp(wNode, "version")),
			}

			if val, ok := wNode.Props["watched_at"]; ok && val != nil {
				t := val.(time.Time)
				entry.WatchedAt = &t
			}

			if val, ok := wNode.Props["rating"]; ok && val != nil {
				r := int(val.(int64))
				entry.Rating = &r
			}

			film := &Film{
				ID:                  fNode.Id,
				IMDbID:              getStringProp(fNode, "imdb_id"),
				Title:               getStringProp(fNode, "title"),
				OriginalTitle:       getStringProp(fNode, "original_title"),
				Year:                int32(getIntProp(fNode, "year")),
				ReleaseDate:         getStringProp(fNode, "release_date"),
				Runtime:             Runtime(getIntProp(fNode, "runtime")),
				RuntimeSeconds:      int32(getIntProp(fNode, "runtime_seconds")),
				Rating:              float32(getFloatProp(fNode, "rating")),
				VoteCount:           float32(getFloatProp(fNode, "vote_count")),
				Description:         getStringProp(fNode, "description"),
				PlotSummary:         getStringProp(fNode, "plot_summary"),
				Certificate:         getStringProp(fNode, "certificate"),
				ProductionStatus:    getStringProp(fNode, "production_status"),
				MetacriticScore:     int32(getIntProp(fNode, "metacritic_score")),
				TrailerID:           getStringProp(fNode, "trailer_id"),
				WatchCategories:     getStringProp(fNode, "watch_categories"),
				WatchProviders:      getStringProp(fNode, "watch_providers"),
				PrimaryImageURL:     getStringProp(fNode, "primary_image_url"),
				PrimaryImageCaption: getStringProp(fNode, "primary_image_caption"),
				Img:                 getStringProp(fNode, "image"),
				Version:             int32(getIntProp(fNode, "version")),
			}

			genresRaw := record.Values[2]
			for _, g := range genresRaw.([]any) {
				film.Genres = append(film.Genres, Genre{Name: g.(string)})
			}

			actorsRaw := record.Values[3]
			for _, a := range actorsRaw.([]any) {
				film.Actors = append(film.Actors, Actor{Name: a.(string)})
			}

			directorsRaw := record.Values[4]
			for _, d := range directorsRaw.([]any) {
				film.Directors = append(film.Directors, Director{Name: d.(string)})
			}

			entry.Film = film
			watchlist = append(watchlist, entry)
		}

		metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
		return struct {
			Watchlist []*Watchlist
			Metadata  Metadata
		}{watchlist, metadata}, nil
	})

	if err != nil {
		return nil, Metadata{}, err
	}

	res := result.(struct {
		Watchlist []*Watchlist
		Metadata  Metadata
	})

	return res.Watchlist, res.Metadata, nil
}

func (m WatchlistModel) Update(entry *Watchlist) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (w:Watchlist)
			WHERE id(w) = $id AND w.version = $version
			MATCH (u:User)-[:HAS_WATCHLIST]->(w)
			WHERE id(u) = $user_id
			SET w.notes = $notes,
				w.priority = $priority,
				w.watched = $watched,
				w.watched_at = datetime($watched_at),
				w.rating = $rating,
				w.version = w.version + 1
			RETURN w.version
		`

		var watchedAt any
		if entry.WatchedAt != nil {
			watchedAt = *entry.WatchedAt
		}

		params := map[string]any{
			"id":         entry.ID,
			"user_id":    entry.UserID,
			"version":    entry.Version,
			"notes":      entry.Notes,
			"priority":   entry.Priority,
			"watched":    entry.Watched,
			"watched_at": watchedAt,
			"rating":     entry.Rating,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			entry.Version = int(result.Record().Values[0].(int64))
			return nil, nil
		}

		return nil, ErrEditConflict
	})

	return err
}

func (m WatchlistModel) Delete(userID, entryID int64) error {
	if entryID < 0 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)
			WHERE id(w) = $entry_id AND id(u) = $user_id
			DETACH DELETE w
			RETURN count(w)
		`

		params := map[string]any{
			"entry_id": entryID,
			"user_id":  userID,
		}

		result, err := tx.Run(ctx, query, params)
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

func (m WatchlistModel) CheckExists(userID, filmID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session := m.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (u:User)-[:HAS_WATCHLIST]->(w:Watchlist)-[:FOR_FILM]->(f:Film)
			WHERE id(u) = $user_id AND id(f) = $film_id
			RETURN count(w) > 0
		`

		params := map[string]any{
			"user_id": userID,
			"film_id": filmID,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return false, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0].(bool), nil
		}

		return false, nil
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}
