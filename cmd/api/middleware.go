package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) chainMiddleware(mux http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	if len(middleware) == 0 {
		return mux
	}

	handler := middleware[len(middleware)-1](mux)

	// Chain the middleware in reverse order
	for i := len(middleware) - 2; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}

func (app *application) recoverPanic(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(){
			if err := recover(); err != nil{
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}



func (app *application) rateLimit(next http.Handler) http.Handler{

	if !app.config.limiter.enabled{
		return next
	}
	
	type client struct{
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)

	go func ()  {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3 * time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil{
			app.serverErrorResponse(w, r, err)
			return
		}
		mu.Lock()
		if _, found := clients[ip]; !found{
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}

		if !clients[ip].limiter.Allow(){
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}