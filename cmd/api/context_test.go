package main

import (
	"net/http"
	"testing"

	"filmapi.zeyadtarek.net/internals/models"
)

// TestContextSetUser tests the contextSetUser function
func TestContextSetUser(t *testing.T) {
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

	// Check that the user was set in the context
	if req.Context().Value(userContextKey) != user {
		t.Errorf("contextSetUser() did not set user in context")
	}
}

// TestContextGetUser tests the contextGetUser function
func TestContextGetUser(t *testing.T) {
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
	if gotUser != user {
		t.Errorf("contextGetUser() got user = %v, want %v", gotUser, user)
	}
}

// TestContextGetUserPanic tests that contextGetUser panics when user is not in context
func TestContextGetUserPanic(t *testing.T) {
	// Create a new application instance
	app := &application{}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Define a function that calls contextGetUser
	f := func() {
		app.contextGetUser(req)
	}

	// Check that the function panics
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("contextGetUser() did not panic")
		}
	}()

	f()
}
