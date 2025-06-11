package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"encoding/json"

	"filmapi.zeyadtarek.net/internals/jsonlog"
	"filmapi.zeyadtarek.net/internals/models"
	_ "github.com/lib/pq"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	cors struct {
		trustedOrigins []string
		enabled        bool
	}
}

type application struct {
	logger *jsonlog.Logger
	config config
	models models.Models
}

const version = "1.0.0"

var buildTime string

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FILMAPI_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum requests per second")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enabled rate limiter")

	// Use flag.Func to parse comma-separated origins
	flag.Func("cors-trusted-origin", "CORS trusted origins (comma-separated)", func(value string) error {
		cfg.cors.trustedOrigins = strings.Fields(value)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Println("Build time\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	app := &application{
		config: cfg,
		logger: logger,
	}

	db, err := openDB(cfg)
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}
	app.logger.PrintInfo("database connection established", nil)
	defer db.Close()

	app.models = models.New(db)

	// Check if the database has less than 9999 films
	if err := populateFilmsIfNeeded(app); err != nil {
		app.logger.PrintFatal(err, nil)
	}

	err = app.serve()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// populateFilmsIfNeeded checks the film count and populates the database if needed
func populateFilmsIfNeeded(app *application) error {

	count, err := app.models.Films.Count()
	if err != nil {
		return err
	}

	if count >= 9999 {
		app.logger.PrintInfo("Database already populated with 9999 films. No insertion needed.", nil)
		return nil
	}

	type FilmInput struct {
		ID          int64          `json:"id"`
		Title       string         `json:"title"`
		Year        int32          `json:"year"`
		Runtime     models.Runtime `json:"runtime"` // Will automatically handle "142 mins" format
		Rating      float32        `json:"rating"`
		Description string         `json:"description"`
		Img         string         `json:"image"`
		Version     int32          `json:"version"`
		Genres      []string       `json:"genres"`
		Directors   []string       `json:"directors"`
		Actors      []string       `json:"actors"`
	}
	// Read the JSON file
	data, err := os.ReadFile("./static/json/films.json")
	if err != nil {
		return err
	}

	// Parse the JSON data into the input struct
	var filmInputs []FilmInput
	if err := json.Unmarshal(data, &filmInputs); err != nil {
		return err
	}

	// Insert each film into the database
	for _, input := range filmInputs {
		film := &models.Film{
			Title:       input.Title,
			Year:        input.Year,
			Runtime:     input.Runtime, // Runtime will already be in the correct format
			Rating:      input.Rating,
			Description: input.Description,
			Img:         input.Img,
			Version:     1,
		}

		// Convert genres to models.Genre
		film.Genres = make([]models.Genre, len(input.Genres))
		for i, genreName := range input.Genres {
			film.Genres[i] = models.Genre{Name: genreName}
		}

		// Convert directors to models.Director
		film.Directors = make([]models.Director, len(input.Directors))
		for i, directorName := range input.Directors {
			film.Directors[i] = models.Director{Name: directorName}
		}

		// Convert actors to models.Actor
		film.Actors = make([]models.Actor, len(input.Actors))
		for i, actorName := range input.Actors {
			film.Actors[i] = models.Actor{Name: actorName}
		}

		// Insert the film into the database
		if err := app.models.Films.Insert(film); err != nil {
			app.logger.PrintError(fmt.Errorf("error inserting film %s: %v", input.Title, err), nil)
			continue
		}

		app.logger.PrintInfo(fmt.Sprintf("Inserted film: %s", film.Title), nil)
	}

	return nil
}
