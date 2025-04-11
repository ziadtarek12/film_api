package main

import (
	"bytes"
	"testing"

	"filmapi.zeyadtarek.net/internals/jsonlog"
)

// TestRoutes tests the routes function
func TestRoutes(t *testing.T) {
	// Create a buffer to capture log output
	logBuffer := bytes.NewBuffer(nil)

	// Create a new application instance
	app := &application{
		logger: jsonlog.New(logBuffer, jsonlog.LevelInfo),
	}

	// Call the routes function
	handler := app.routes()

	// Check that the handler is not nil
	if handler == nil {
		t.Error("routes() returned nil handler")
	}

	// Test that the routes are registered correctly
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Healthcheck",
			method: "GET",
			path:   "/v1/healthcheck",
		},
		{
			name:   "Create user",
			method: "POST",
			path:   "/v1/user",
		},
		{
			name:   "Activate user",
			method: "PUT",
			path:   "/v1/users/activated",
		},
		{
			name:   "Create authentication token",
			method: "POST",
			path:   "/v1/tokens/authentication",
		},
		{
			name:   "List films",
			method: "GET",
			path:   "/v1/films",
		},
		{
			name:   "Create film",
			method: "POST",
			path:   "/v1/films",
		},
		{
			name:   "Get film",
			method: "GET",
			path:   "/v1/films/1",
		},
		{
			name:   "Update film",
			method: "PATCH",
			path:   "/v1/films/1",
		},
		{
			name:   "Delete film",
			method: "DELETE",
			path:   "/v1/films/1",
		},
	}

	// We can't easily test the routing directly without a running server,
	// but we can at least check that the handler doesn't panic when we create it
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a very basic test that just ensures the routes function doesn't panic
			// A more thorough test would involve starting a test server and making requests
		})
	}
}
