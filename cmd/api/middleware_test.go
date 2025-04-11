package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"filmapi.zeyadtarek.net/internals/jsonlog"
	"filmapi.zeyadtarek.net/internals/models"
)

// TestRecoverPanic tests the recoverPanic middleware
func TestRecoverPanic(t *testing.T) {
	// Create a new application instance with mock dependencies
	app := &application{
		logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
	}

	// Create a handler that will panic
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap the handler with the recoverPanic middleware
	handler := app.recoverPanic(nextHandler)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// TestRequireAuthenticatedUser tests the requireAuthenticatedUser middleware
func TestRequireAuthenticatedUser(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		wantStatusCode int
	}{
		{
			name: "Authenticated user",
			user: &models.User{
				ID:    1,
				Name:  "Test User",
				Email: "test@example.com",
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Anonymous user",
			user:           models.AnonymousUser,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new application instance with mock dependencies
			app := &application{
				logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
			}

			// Create a handler that will return 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with the requireAuthenticatedUser middleware
			handler := app.requireAuthenticatedUser(nextHandler)

			// Create a new HTTP request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set the user in the request context
			req = app.contextSetUser(req, tt.user)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler function
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatusCode)
			}
		})
	}
}

// TestRequireActivatedUser tests the requireActivatedUser middleware
func TestRequireActivatedUser(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		wantStatusCode int
	}{
		{
			name: "Activated user",
			user: &models.User{
				ID:        1,
				Name:      "Test User",
				Email:     "test@example.com",
				Activated: true,
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Inactive user",
			user: &models.User{
				ID:        1,
				Name:      "Test User",
				Email:     "test@example.com",
				Activated: false,
			},
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "Anonymous user",
			user:           models.AnonymousUser,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new application instance with mock dependencies
			app := &application{
				logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
			}

			// Create a handler that will return 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with the requireActivatedUser middleware
			handler := app.requireActivatedUser(nextHandler)

			// Create a new HTTP request
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Set the user in the request context
			req = app.contextSetUser(req, tt.user)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler function
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatusCode)
			}
		})
	}
}

// TestRequirePermission tests the requirePermission middleware
func TestRequirePermission(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex mocking of models.PermissionModel")
}

// TestContextSetUser and TestContextGetUser test the context functions
func TestContextFunctions(t *testing.T) {
	// Create a new application instance
	app := &application{}

	// Create a user
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set the user in the request context
	req = app.contextSetUser(req, user)

	// Get the user from the request context
	gotUser := app.contextGetUser(req)

	// Check that the user is the same
	if gotUser.ID != user.ID {
		t.Errorf("contextGetUser() got user ID = %v, want %v", gotUser.ID, user.ID)
	}
	if gotUser.Name != user.Name {
		t.Errorf("contextGetUser() got user Name = %v, want %v", gotUser.Name, user.Name)
	}
	if gotUser.Email != user.Email {
		t.Errorf("contextGetUser() got user Email = %v, want %v", gotUser.Email, user.Email)
	}
}

// TestChainMiddleware tests the chainMiddleware function
func TestChainMiddleware(t *testing.T) {
	// Create a new application instance
	app := &application{}

	// Create a handler that will return 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware functions that add headers to the response
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	// Chain the middleware
	chainedHandler := app.chainMiddleware(handler, middleware1, middleware2)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function
	chainedHandler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the middleware headers were added
	if rr.Header().Get("X-Middleware-1") != "true" {
		t.Errorf("middleware1 header not set")
	}
	if rr.Header().Get("X-Middleware-2") != "true" {
		t.Errorf("middleware2 header not set")
	}
}
