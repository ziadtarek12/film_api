package main

import (
	// "crypto/tls"
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	// "filmapi.zeyadtarek.net/internals/cert"
	"filmapi.zeyadtarek.net/internals/models"
	_ "github.com/lib/pq"
)

type config struct {
    port int
    env  string
    db struct{
        dsn string
        maxOpenConns int
        maxIdleConns int
        maxIdleTime  string
    }
}

type application struct {
    infoLogger  *log.Logger
    errorLogger *log.Logger
    config      config
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
    flag.Parse()
   

    infologger := log.New(os.Stdout, "Info: ", log.Ldate|log.Ltime|log.Lshortfile)
    errorLogger := log.New(os.Stderr, "Error: ", log.Ldate|log.Llongfile|log.Ltime)

    
    app := &application{
        infoLogger:  infologger,
        errorLogger: errorLogger,
        config:      cfg,
    }

    db, err := openDB(cfg)
    if err != nil{
        app.errorLogger.Fatal(err)
    }
    defer db.Close()

    app.models = models.New(db)

    // tlsConfig := &tls.Config{
    //     CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
    //     MinVersion:       tls.VersionTLS12,
    //     MaxVersion:       tls.VersionTLS12,
    //     CipherSuites: []uint16{
    //         tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
    //         tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
    //         tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
    //         tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
    //         tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
    //         tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    //     },
    // }

    // _, err := os.Stat(cert.CertPath)
    // if err != nil {
    //     if os.IsNotExist(err) {
    //         err = cert.GenerateTLSCertAndKey(cert.CertPath, cert.KeyPath)
    //         if err != nil {
    //             app.errorLogger.Fatal(err)
    //         }
    //     } else {
    //         app.errorLogger.Fatal(err)
    //     }
    // }

    srv := http.Server{
        Addr:         ":"+ strconv.Itoa(cfg.port),
        Handler:      app.routes(),
        ErrorLog:     app.errorLogger,
        // TLSConfig:    tlsConfig,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  time.Minute,
    }

    app.infoLogger.Printf("Starting listening on port %d", app.config.port)
    // err = srv.ListenAndServeTLS(cert.CertPath, cert.KeyPath)
    err = srv.ListenAndServe()
    app.errorLogger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error){
    db, err := sql.Open("postgres", cfg.db.dsn)
    if err != nil{
        return nil, err
    }

    db.SetMaxOpenConns(cfg.db.maxOpenConns)
    db.SetMaxIdleConns(cfg.db.maxIdleConns)
    
    duration, err := time.ParseDuration(cfg.db.maxIdleTime)
    if err != nil{
        return nil, err
    }

    db.SetConnMaxIdleTime(duration)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = db.PingContext(ctx)
    if err != nil{
        return nil, err
    }

    return db, nil
}