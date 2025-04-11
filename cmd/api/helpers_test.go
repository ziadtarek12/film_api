package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"filmapi.zeyadtarek.net/internals/jsonlog"
)

// TestWriteJSON tests the writeJSON function
func TestWriteJSON(t *testing.T) {
	// Create a new application instance
	app := &application{}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create test data
	data := map[string]string{
		"message": "test message",
	}

	// Create custom headers
	headers := http.Header{
		"X-Custom-Header": []string{"test value"},
	}

	// Call the writeJSON function
	err := app.writeJSON(rr, http.StatusOK, data, headers)
	if err != nil {
		t.Fatalf("writeJSON() returned error: %v", err)
	}

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("writeJSON() returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("writeJSON() returned wrong Content-Type: got %v want %v", contentType, "application/json")
	}

	// Check the custom header
	customHeader := rr.Header().Get("X-Custom-Header")
	if customHeader != "test value" {
		t.Errorf("writeJSON() did not set custom header: got %v want %v", customHeader, "test value")
	}

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if message, ok := response["message"]; !ok || message != "test message" {
		t.Errorf("writeJSON() returned wrong body: got %v want %v", response, data)
	}
}

// TestReadJSON tests the readJSON function
func TestReadJSON(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantData  map[string]string
		wantError bool
	}{
		{
			name:     "Valid JSON",
			body:     `{"message": "test message"}`,
			wantData: map[string]string{"message": "test message"},
		},
		{
			name:      "Invalid JSON",
			body:      `{"message": "test message"`,
			wantError: true,
		},
		{
			name:      "Empty body",
			body:      "",
			wantError: true,
		},
		{
			name:      "Too large body",
			body:      string(make([]byte, 1024*1024+1)), // 1MB + 1 byte
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new application instance
			app := &application{
				logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
			}

			// Create a new HTTP request with the test body
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Set the Content-Type header
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Create a variable to hold the parsed data
			var data map[string]string

			// Call the readJSON function
			err = app.readJSON(rr, req, &data)

			// Check if an error was expected
			if (err != nil) != tt.wantError {
				t.Errorf("readJSON() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// If no error was expected, check the parsed data
			if !tt.wantError {
				if data["message"] != tt.wantData["message"] {
					t.Errorf("readJSON() parsed data = %v, want %v", data, tt.wantData)
				}
			}
		})
	}
}

// TestReadString tests the readString function
func TestReadString(t *testing.T) {
	tests := []struct {
		name         string
		queryString  url.Values
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "Key exists",
			queryString:  url.Values{"name": []string{"test"}},
			key:          "name",
			defaultValue: "default",
			want:         "test",
		},
		{
			name:         "Key doesn't exist",
			queryString:  url.Values{},
			key:          "name",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "Key exists but empty",
			queryString:  url.Values{"name": []string{""}},
			key:          "name",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new application instance
			app := &application{}

			// Call the readString function
			got := app.readString(tt.queryString, tt.key, tt.defaultValue)

			// Check the result
			if got != tt.want {
				t.Errorf("readString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestReadCSV tests the readCSV function
func TestReadCSV(t *testing.T) {
	// Skip this test for now as it requires more complex mocking
	t.Skip("Skipping test that requires complex mocking of readCSV")
}

// TestReadInt tests the readInt function
func TestReadInt(t *testing.T) {
	// Skip this test for now as it requires validator
	t.Skip("Skipping test that requires validator")
}

// TestErrorResponse tests the errorResponse function
func TestErrorResponse(t *testing.T) {
	// Create a new application instance
	app := &application{
		logger: jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo),
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the errorResponse function
	app.errorResponse(rr, req, http.StatusBadRequest, "test error")

	// Check the status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("errorResponse() returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Check the Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("errorResponse() returned wrong Content-Type: got %v want %v", contentType, "application/json")
	}

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if message, ok := response["error"]; !ok || message != "test error" {
		t.Errorf("errorResponse() returned wrong body: got %v want %v", response, map[string]string{"error": "test error"})
	}
}

// TestServerErrorResponse tests the serverErrorResponse function
func TestServerErrorResponse(t *testing.T) {
	// Create a buffer to capture log output
	logBuffer := bytes.NewBuffer(nil)

	// Create a new application instance
	app := &application{
		logger: jsonlog.New(logBuffer, jsonlog.LevelInfo),
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Call the serverErrorResponse function
	app.serverErrorResponse(rr, req, errors.New("test error"))

	// Check the status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("serverErrorResponse() returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("serverErrorResponse() returned wrong Content-Type: got %v want %v", contentType, "application/json")
	}

	// Check the response body
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedMessage := "The server encountred a problem and could not process your request"
	if message, ok := response["error"]; !ok || message != expectedMessage {
		t.Errorf("serverErrorResponse() returned wrong body: got %v want %v", response, map[string]string{"error": expectedMessage})
	}

	// Check that the error was logged
	logOutput := logBuffer.String()
	if logOutput == "" {
		t.Errorf("serverErrorResponse() did not log the error")
	}
}
