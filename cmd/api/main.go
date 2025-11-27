package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"filmapi.zeyadtarek.net/internals/jsonlog"
	"filmapi.zeyadtarek.net/internals/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type config struct {
	port int
	env  string
	db   struct {
		uri          string
		username     string
		password     string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	jwt struct {
		secret          string
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
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
	fmt.Println("Starting application...")
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")

	// Neo4j configuration
	flag.StringVar(&cfg.db.uri, "db-uri", os.Getenv("NEO4J_URI"), "Neo4j URI")
	flag.StringVar(&cfg.db.username, "db-username", os.Getenv("NEO4J_USERNAME"), "Neo4j Username")
	flag.StringVar(&cfg.db.password, "db-password", os.Getenv("NEO4J_PASSWORD"), "Neo4j Password")

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

	// JWT configuration
	cfg.jwt.secret = os.Getenv("JWT_SECRET")
	cfg.jwt.accessTokenTTL = 1 * time.Hour
	cfg.jwt.refreshTokenTTL = 7 * 24 * time.Hour

	fmt.Println(cfg.db.uri)
	fmt.Println(cfg.db.username)
	fmt.Println(cfg.db.password)
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	app := &application{
		config: cfg,
		logger: logger,
	}

	driver, err := openDB(cfg)
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}
	app.logger.PrintInfo("database connection established", nil)
	defer driver.Close(context.Background())

	app.models = models.New(driver)

	// Check if the database has less than 100 films
	if err := populateFilmsIfNeeded(app); err != nil {
		app.logger.PrintFatal(err, nil)
	}

	err = app.serve()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.db.uri, neo4j.BasicAuth(cfg.db.username, cfg.db.password, ""))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}

	return driver, nil
}

// populateFilmsIfNeeded checks the film count and populates the database if needed
func populateFilmsIfNeeded(app *application) error {
	count, err := app.models.Films.Count()
	if err != nil {
		return err
	}

	if count >= 100 {
		app.logger.PrintInfo("Database already populated with films. No insertion needed.", nil)
		return nil
	}

	// Read the CSV file
	csvFile, err := os.Open("./filtered_films.csv")
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer csvFile.Close()

	// Create a CSV reader
	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Skip header row
	records = records[1:]

	var films []*models.Film

	// Insert each film into the database
	for i, record := range records {
		if len(record) < 30 {
			app.logger.PrintError(fmt.Errorf("record %d has insufficient fields", i+1), nil)
			continue
		}

		// Parse year
		var year int64
		if record[3] == "Not Provided" || record[3] == "" {
			app.logger.PrintInfo(fmt.Sprintf("Record %d: Year not provided, using default 1900", i+1), nil)
			year = 1900
		} else {
			var err error
			year, err = strconv.ParseInt(record[3], 10, 32)
			if err != nil {
				app.logger.PrintInfo(fmt.Sprintf("Record %d: invalid year '%s', using default 1900", i+1, record[3]), nil)
				year = 1900
			}
		}

		// Parse rating
		rating, err := strconv.ParseFloat(record[5], 32)
		if err != nil {
			app.logger.PrintError(fmt.Errorf("invalid rating in record %d: %v", i+1, err), nil)
			continue
		}

		// Parse vote count
		voteCount, err := strconv.ParseFloat(record[6], 32)
		if err != nil {
			voteCount = 0 // Default to 0 if parsing fails
		}

		// Parse runtime
		runtime, err := strconv.ParseInt(record[7], 10, 32)
		if err != nil {
			app.logger.PrintError(fmt.Errorf("invalid runtime in record %d: %v", i+1, err), nil)
			continue
		}

		// Parse runtime seconds
		runtimeSeconds, err := strconv.ParseInt(record[8], 10, 32)
		if err != nil {
			runtimeSeconds = runtime * 60 // Default calculation
		}

		// Parse metacritic score
		metacriticScore, err := strconv.ParseInt(record[16], 10, 32)
		if err != nil {
			metacriticScore = 0 // Default to 0 if parsing fails
		}

		film := &models.Film{
			IMDbID:              record[0],
			Title:               record[1],
			OriginalTitle:       record[2],
			Year:                int32(year),
			ReleaseDate:         record[4],
			Rating:              float32(rating),
			VoteCount:           float32(voteCount),
			Runtime:             models.Runtime(runtime),
			RuntimeSeconds:      int32(runtimeSeconds),
			Description:         record[9],
			PlotSummary:         record[10],
			Certificate:         record[11],
			ProductionStatus:    record[15],
			MetacriticScore:     int32(metacriticScore),
			TrailerID:           record[25],
			WatchCategories:     record[26],
			WatchProviders:      record[27],
			PrimaryImageURL:     record[28],
			PrimaryImageCaption: record[29],
			Img:                 record[28], // Use primary image URL as fallback
			Version:             1,
		}

		// Parse genres (combining multiple genre columns)
		var genreNames []string
		if record[12] != "" {
			genreNames = append(genreNames, strings.Split(strings.TrimSpace(record[12]), ",")...)
		}
		if record[13] != "" {
			genreNames = append(genreNames, strings.Split(strings.TrimSpace(record[13]), ",")...)
		}
		if record[14] != "" {
			genreNames = append(genreNames, strings.Split(strings.TrimSpace(record[14]), ",")...)
		}

		// Clean up genre names and create Genre slice
		var uniqueGenres []string
		genreMap := make(map[string]bool)
		for _, genre := range genreNames {
			cleanGenre := strings.TrimSpace(genre)
			if cleanGenre != "" && !genreMap[cleanGenre] {
				uniqueGenres = append(uniqueGenres, cleanGenre)
				genreMap[cleanGenre] = true
			}
		}

		film.Genres = make([]models.Genre, len(uniqueGenres))
		for i, genreName := range uniqueGenres {
			film.Genres[i] = models.Genre{Name: genreName}
		}

		// Parse actors (combining multiple actor columns)
		var actorNames []string
		if record[17] != "" {
			actorNames = append(actorNames, strings.Split(record[17], ",")...)
		}
		if record[18] != "" {
			actorNames = append(actorNames, strings.Split(record[18], ",")...)
		}
		if record[19] != "" {
			actorNames = append(actorNames, strings.Split(record[19], ",")...)
		}

		// Clean up actor names
		var uniqueActors []string
		actorMap := make(map[string]bool)
		for _, actor := range actorNames {
			cleanActor := strings.TrimSpace(actor)
			if cleanActor != "" && !actorMap[cleanActor] {
				uniqueActors = append(uniqueActors, cleanActor)
				actorMap[cleanActor] = true
			}
		}

		film.Actors = make([]models.Actor, len(uniqueActors))
		for i, actorName := range uniqueActors {
			film.Actors[i] = models.Actor{Name: actorName}
		}

		// Parse directors
		var directorNames []string
		if record[23] != "" {
			directorNames = strings.Split(record[23], ",")
		}

		// Clean up director names
		var uniqueDirectors []string
		directorMap := make(map[string]bool)
		for _, director := range directorNames {
			cleanDirector := strings.TrimSpace(director)
			if cleanDirector != "" && !directorMap[cleanDirector] {
				uniqueDirectors = append(uniqueDirectors, cleanDirector)
				directorMap[cleanDirector] = true
			}
		}

		film.Directors = make([]models.Director, len(uniqueDirectors))
		for i, directorName := range uniqueDirectors {
			film.Directors[i] = models.Director{Name: directorName}
		}

		// Ensure we have at least one genre (validation requirement)
		if len(film.Genres) == 0 {
			film.Genres = []models.Genre{{Name: "Unknown"}}
		}

		films = append(films, film)
	}

	if len(films) > 0 {
		if err := app.models.Films.InsertBulk(films); err != nil {
			return fmt.Errorf("failed to bulk insert films: %v", err)
		}
		app.logger.PrintInfo(fmt.Sprintf("Successfully bulk inserted %d films", len(films)), nil)
	}

	return nil
}
