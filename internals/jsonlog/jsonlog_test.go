package jsonlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// TestLevelString tests the String method of Level
func TestLevelString(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "Info level",
			level: LevelInfo,
			want:  "INFO",
		},
		{
			name:  "Error level",
			level: LevelError,
			want:  "ERROR",
		},
		{
			name:  "Fatal level",
			level: LevelFatal,
			want:  "FATAL",
		},
		{
			name:  "Unknown level",
			level: Level(99),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	// Create a buffer to capture log output
	buffer := bytes.NewBuffer(nil)

	// Create a new logger
	logger := New(buffer, LevelInfo)

	// Check that the logger was created with the correct properties
	if logger.out != buffer {
		t.Errorf("New() logger.out = %v, want %v", logger.out, buffer)
	}
	if logger.minLevel != LevelInfo {
		t.Errorf("New() logger.minLevel = %v, want %v", logger.minLevel, LevelInfo)
	}
}

// TestPrintInfo tests the PrintInfo method
func TestPrintInfo(t *testing.T) {
	// Create a buffer to capture log output
	buffer := bytes.NewBuffer(nil)

	// Create a new logger
	logger := New(buffer, LevelInfo)

	// Call PrintInfo
	logger.PrintInfo("test message", map[string]string{"key": "value"})

	// Parse the log output
	var log map[string]interface{}
	err := json.Unmarshal(buffer.Bytes(), &log)
	if err != nil {
		t.Fatalf("Failed to unmarshal log output: %v", err)
	}

	// Check the log level
	if level, ok := log["level"]; !ok || level != "INFO" {
		t.Errorf("PrintInfo() log level = %v, want %v", level, "INFO")
	}

	// Check the log message
	if message, ok := log["message"]; !ok || message != "test message" {
		t.Errorf("PrintInfo() log message = %v, want %v", message, "test message")
	}

	// Check the properties
	if properties, ok := log["properties"].(map[string]interface{}); !ok {
		t.Errorf("PrintInfo() log properties not found")
	} else if value, ok := properties["key"]; !ok || value != "value" {
		t.Errorf("PrintInfo() log property = %v, want %v", value, "value")
	}

	// Check that there's no trace
	if _, ok := log["trace"]; ok {
		t.Errorf("PrintInfo() log should not have a trace")
	}
}

// TestPrintError tests the PrintError method
func TestPrintError(t *testing.T) {
	// Create a buffer to capture log output
	buffer := bytes.NewBuffer(nil)

	// Create a new logger
	logger := New(buffer, LevelInfo)

	// Call PrintError
	logger.PrintError(errors.New("test error"), map[string]string{"key": "value"})

	// Parse the log output
	var log map[string]interface{}
	err := json.Unmarshal(buffer.Bytes(), &log)
	if err != nil {
		t.Fatalf("Failed to unmarshal log output: %v", err)
	}

	// Check the log level
	if level, ok := log["level"]; !ok || level != "ERROR" {
		t.Errorf("PrintError() log level = %v, want %v", level, "ERROR")
	}

	// Check the log message
	if message, ok := log["message"]; !ok || message != "test error" {
		t.Errorf("PrintError() log message = %v, want %v", message, "test error")
	}

	// Check the properties
	if properties, ok := log["properties"].(map[string]interface{}); !ok {
		t.Errorf("PrintError() log properties not found")
	} else if value, ok := properties["key"]; !ok || value != "value" {
		t.Errorf("PrintError() log property = %v, want %v", value, "value")
	}

	// Check that there's a trace
	if trace, ok := log["trace"]; !ok || trace == "" {
		t.Errorf("PrintError() log should have a trace")
	}
}

// TestWrite tests the Write method
func TestWrite(t *testing.T) {
	// Create a buffer to capture log output
	buffer := bytes.NewBuffer(nil)

	// Create a new logger
	logger := New(buffer, LevelInfo)

	// Call Write
	message := []byte("test message")
	n, err := logger.Write(message)
	if err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	if n <= 0 {
		t.Errorf("Write() returned n = %v, want > 0", n)
	}

	// Parse the log output
	var log map[string]interface{}
	err = json.Unmarshal(buffer.Bytes(), &log)
	if err != nil {
		t.Fatalf("Failed to unmarshal log output: %v", err)
	}

	// Check the log level
	if level, ok := log["level"]; !ok || level != "ERROR" {
		t.Errorf("Write() log level = %v, want %v", level, "ERROR")
	}

	// Check the log message
	if message, ok := log["message"]; !ok || message != "test message" {
		t.Errorf("Write() log message = %v, want %v", message, "test message")
	}

	// Check that there's a trace
	if trace, ok := log["trace"]; !ok || trace == "" {
		t.Errorf("Write() log should have a trace")
	}
}

// TestLogLevelFiltering tests that logs below the minimum level are filtered out
func TestLogLevelFiltering(t *testing.T) {
	// Create a buffer to capture log output
	buffer := bytes.NewBuffer(nil)

	// Create a new logger with ERROR level
	logger := New(buffer, LevelError)

	// Call PrintInfo (which should be filtered out)
	logger.PrintInfo("test message", nil)

	// Check that nothing was logged
	if buffer.Len() > 0 {
		t.Errorf("PrintInfo() should not have logged anything, but got: %s", buffer.String())
	}

	// Call PrintError (which should be logged)
	logger.PrintError(errors.New("test error"), nil)

	// Check that the error was logged
	if buffer.Len() == 0 {
		t.Errorf("PrintError() should have logged something")
	}
	if !strings.Contains(buffer.String(), "ERROR") {
		t.Errorf("PrintError() log should contain ERROR level")
	}
}
