package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"time"

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

	flag.BoolVar(&cfg.cors.enabled, "cors-enabled", true, "Enable CORS")
	flag.Parse()

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
