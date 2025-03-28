package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func (app *application) serve() error{
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
        Addr:         ":"+ strconv.Itoa(app.config.port),
        Handler:      app.routes(),
        ErrorLog: log.New(app.logger, "", 0),
        // TLSConfig:    tlsConfig,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  time.Minute,
    }

	shutdownError := make(chan error)

	go func ()  {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)	
		s := <-quit
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()
    app.logger.PrintInfo("starting server", map[string]string{
        "addr": srv.Addr,
        "env": app.config.env,
    })

    // err = srv.ListenAndServeTLS(cert.CertPath, cert.KeyPath)
    err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed){
		return err
	}

	err = <-shutdownError
	if err != nil{
		return err
	}

	return nil
}